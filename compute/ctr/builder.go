package ctr

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/content"
	"github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/core/images/archive"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	"github.com/containerd/containerd/v2/pkg/rootfs"
	"github.com/containerd/errdefs"
	"github.com/opencontainers/go-digest"
	imgspecs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/compute/provider"
	"github.com/cruciblehq/crux/crex"
)

// containerd-backed [provider.ImageBuilder].
type builder struct {
	client  *containerd.Client
	mu      sync.Mutex
	created []string // image refs created by this builder, cleaned up on Close
}

// A single file entry collected from an image layer during extraction.
type layerEntry struct {
	hdr  tar.Header
	data []byte
}

// Connects to the containerd socket at socketPath and returns a new [provider.ImageBuilder].
//
// The caller must call Close when done to release the connection and remove
// any intermediate images created during the build.
func NewImageBuilder(socketPath string) (provider.ImageBuilder, error) {
	c, err := containerd.New(socketPath)
	if err != nil {
		return nil, crex.Wrap(ErrConnect, err)
	}
	return &builder{client: c}, nil
}

// Loads an OCI image archive into the builder's image store.
//
// Returns the image reference for use with other methods.
func (b *builder) Import(ctx context.Context, r io.Reader) (string, error) {
	ctx = namespaces.WithNamespace(ctx, crucibleNamespace)
	imgs, err := b.client.Import(ctx, r)
	if err != nil {
		return "", crex.Wrap(ErrImport, err)
	}
	if len(imgs) == 0 {
		return "", crex.Wrap(ErrImport, ErrNoImages)
	}
	ref := imgs[0].Name

	// Unpack the imported image so that the layer chain exists in the snapshot
	// store before any subsequent Copy or Run steps. Without this, Run calls
	// Unpack on a derived image whose parent snapshots do not yet exist,
	// causing "parent snapshot does not exist" errors. Unpack is idempotent.
	img, err := b.client.GetImage(ctx, ref)
	if err != nil {
		return "", crex.Wrapf(ErrImport, "get imported image: %w", err)
	}
	if err := img.Unpack(ctx, ""); err != nil {
		return "", crex.Wrapf(ErrImport, "unpack imported image: %w", err)
	}

	b.trackImage(ref)
	return ref, nil
}

// Runs a shell command in a build container derived from imageRef, commits
// the resulting filesystem diff as a new layer, and returns the updated image
// reference.
func (b *builder) Run(ctx context.Context, imageRef string, cfg *provider.RunConfig) (string, error) {
	ctx = namespaces.WithNamespace(ctx, crucibleNamespace)

	img, err := b.client.GetImage(ctx, imageRef)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "get image %q: %w", imageRef, err)
	}

	// containerd.WithNewSnapshot calls snapshotter.Prepare with the image's
	// chain ID as parent but does not unpack the image layers first. Unpack
	// ensures the layer chain exists in the snapshot store before the
	// container is created. Unpack is idempotent and a no-op when the layers
	// are already present.
	if err := img.Unpack(ctx, "overlayfs"); err != nil {
		return "", crex.Wrapf(ErrBuild, "unpack image %q: %w", imageRef, err)
	}

	stepID := newStepID()
	snapshotID := stepID + "-snapshot"
	containerID := stepID + "-container"

	platform, err := imagePlatform(ctx, img, b.client.ContentStore())
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "get image platform: %w", err)
	}

	container, err := b.client.NewContainer(ctx, containerID,
		containerd.WithRuntime(runcV2Runtime, nil),
		containerd.WithImage(img),
		containerd.WithSnapshotter("overlayfs"),
		containerd.WithNewSnapshot(snapshotID, img),
		containerd.WithNewSpec(
			oci.WithDefaultSpecForPlatform(platform),
			oci.WithImageConfig(img),
			withSecurityPolicy(cfg.Security),
			withBuildStep(cfg),
		),
	)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "create container: %w", err)
	}

	// Remove the snapshot after the function returns; the container metadata
	// is removed separately (no WithSnapshotCleanup so we control timing).
	info, err := container.Info(ctx)
	if err != nil {
		container.Delete(ctx)
		return "", crex.Wrapf(ErrBuild, "container info: %w", err)
	}
	snapshotterName := info.Snapshotter
	defer func() {
		b.client.SnapshotService(snapshotterName).Remove(ctx, snapshotID)
		container.Delete(ctx)
	}()

	if err := b.runTask(ctx, container, cfg.Stdout, cfg.Stderr); err != nil {
		return "", err
	}

	newRef, err := b.commitSnapshotDiff(ctx, img, snapshotID, snapshotterName, "crucible run "+cfg.Command)
	if err != nil {
		return "", err
	}
	b.trackImage(newRef)
	return newRef, nil
}

// Commits the diff of snapshotID as a new image layer appended to img.
//
// Creates a temporary lease to protect the new blob from GC until the
// manifest's GC labels are written. Returns the new image reference.
func (b *builder) commitSnapshotDiff(ctx context.Context, img containerd.Image, snapshotID, snapshotterName, createdBy string) (string, error) {
	leaseCtx, doneLease, err := b.client.WithLease(ctx)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "create lease: %w", err)
	}
	defer doneLease(ctx)

	snapshotter := b.client.SnapshotService(snapshotterName)
	layerDesc, err := rootfs.CreateDiff(leaseCtx, snapshotID, snapshotter, b.client.DiffService())
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "create diff: %w", err)
	}

	diffID, err := b.uncompressedDigest(leaseCtx, layerDesc)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "compute diff ID: %w", err)
	}

	return b.appendLayerToImage(leaseCtx, img, layerDesc, diffID, createdBy)
}

// Appends the uncompressed tar archive r as a new layer on top of imageRef.
//
// The layer is compressed and stored in the builder's content store.
// Returns the updated image reference.
func (b *builder) Copy(ctx context.Context, imageRef string, r io.Reader) (string, error) {
	ctx = namespaces.WithNamespace(ctx, crucibleNamespace)

	img, err := b.client.GetImage(ctx, imageRef)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "get image %q: %w", imageRef, err)
	}

	// Protect new content from GC until the manifest's GC labels are written.
	leaseCtx, doneLease, err := b.client.WithLease(ctx)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "create lease: %w", err)
	}
	defer doneLease(ctx)

	layerDesc, diffID, err := b.compressAndStoreLayer(leaseCtx, r)
	if err != nil {
		return "", err
	}

	newRef, err := b.appendLayerToImage(leaseCtx, img, layerDesc, diffID, "crucible copy")
	if err != nil {
		return "", err
	}
	b.trackImage(newRef)
	return newRef, nil
}

// Compresses r as a gzip layer, writes the blob to the content store under
// ctx, and returns the layer descriptor and uncompressed diff ID.
//
// ctx must carry a containerd lease to protect the blob from GC until the
// caller writes a manifest that references it.
func (b *builder) compressAndStoreLayer(ctx context.Context, r io.Reader) (ocispec.Descriptor, digest.Digest, error) {
	var compressedBuf bytes.Buffer
	uncompressedDigester := digest.SHA256.Digester()
	compressedDigester := digest.SHA256.Digester()

	gw := gzip.NewWriter(io.MultiWriter(&compressedBuf, compressedDigester.Hash()))
	teeIn := io.TeeReader(r, uncompressedDigester.Hash())
	if _, err := io.Copy(gw, teeIn); err != nil {
		return ocispec.Descriptor{}, "", crex.Wrapf(ErrBuild, "compress layer: %w", err)
	}
	if err := gw.Close(); err != nil {
		return ocispec.Descriptor{}, "", crex.Wrapf(ErrBuild, "finalize compressed layer: %w", err)
	}

	compressedDgst := compressedDigester.Digest()
	diffID := uncompressedDigester.Digest()

	layerDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageLayerGzip,
		Digest:    compressedDgst,
		Size:      int64(compressedBuf.Len()),
	}

	// Use a unique ingest ref to prevent resumption of a stale ingest from a
	// prior run. Omit the digest from the descriptor so that the server's
	// pre-flight check — which queries BoltDB and returns AlreadyExists when a
	// stale metadata record exists for a digest whose blob file was deleted by
	// GC — is bypassed on both the STAT and COMMIT gRPC messages. The blob is
	// still committed under the correct digest (sha256 of the actual data
	// written), and the BoltDB bucket is upserted so any stale record is
	// overwritten. AlreadyExists on Commit means another writer raced us to the
	// same blob and succeeded first, which is fine.
	ref := "layer-" + compressedDgst.Encoded() + "-" + newStepID()
	cw, err := content.OpenWriter(ctx, b.client.ContentStore(),
		content.WithRef(ref),
		content.WithDescriptor(ocispec.Descriptor{
			MediaType: ocispec.MediaTypeImageLayerGzip,
			Size:      layerDesc.Size,
		}),
	)
	if err != nil {
		if !errdefs.IsAlreadyExists(err) {
			return ocispec.Descriptor{}, "", crex.Wrapf(ErrBuild, "open layer writer: %w", err)
		}
		// Blob already committed — nothing more to write.
	} else {
		defer cw.Close()
		if _, err := io.Copy(cw, &compressedBuf); err != nil {
			return ocispec.Descriptor{}, "", crex.Wrapf(ErrBuild, "write layer: %w", err)
		}
		if err := cw.Commit(ctx, layerDesc.Size, ""); err != nil {
			if !errdefs.IsAlreadyExists(err) {
				return ocispec.Descriptor{}, "", crex.Wrapf(ErrBuild, "commit layer: %w", err)
			}
		}
	}

	return layerDesc, diffID, nil
}

// Applies configuration updates to imageRef without adding a new layer.
//
// Only non-zero fields in cfg are applied. Returns the updated image
// reference.
func (b *builder) Configure(ctx context.Context, imageRef string, cfg *provider.ConfigUpdate) (string, error) {
	ctx = namespaces.WithNamespace(ctx, crucibleNamespace)

	img, err := b.client.GetImage(ctx, imageRef)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "get image %q: %w", imageRef, err)
	}

	cs := b.client.ContentStore()
	mfst, err := images.Manifest(ctx, cs, img.Target(), nil)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "get manifest: %w", err)
	}

	configData, err := content.ReadBlob(ctx, cs, mfst.Config)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "read config: %w", err)
	}

	var imgCfg ocispec.Image
	if err := json.Unmarshal(configData, &imgCfg); err != nil {
		return "", crex.Wrapf(ErrBuild, "unmarshal config: %w", err)
	}

	applyConfigUpdate(&imgCfg, cfg)

	imgCfg.History = append(imgCfg.History, ocispec.History{
		CreatedBy:  "crucible config",
		EmptyLayer: true,
	})

	newRef, err := b.writeNewImage(ctx, mfst, imgCfg, nil)
	if err != nil {
		return "", err
	}
	b.trackImage(newRef)
	return newRef, nil
}

// Extracts srcPath from imageRef as an uncompressed tar stream.
//
// Spawns a goroutine that reads layers from the content store and writes the
// result to a pipe; the caller receives the read end. The caller is
// responsible for closing the returned ReadCloser.
func (b *builder) Extract(ctx context.Context, imageRef, srcPath string) (io.ReadCloser, error) {
	ctx = namespaces.WithNamespace(ctx, crucibleNamespace)

	img, err := b.client.GetImage(ctx, imageRef)
	if err != nil {
		return nil, crex.Wrapf(ErrBuild, "get image %q: %w", imageRef, err)
	}

	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(b.streamExtract(ctx, img, srcPath, pw))
	}()
	return pr, nil
}

// Reads image layers in order, collects entries whose paths match srcPath with
// last-writer-wins semantics (later layers override earlier ones), applies
// whiteout deletions, then writes the survivors as an uncompressed tar to w.
//
// Returns an error when srcPath is not found in any layer.
func (b *builder) streamExtract(ctx context.Context, img containerd.Image, srcPath string, w io.Writer) error {
	cs := b.client.ContentStore()

	mfst, err := images.Manifest(ctx, cs, img.Target(), nil)
	if err != nil {
		return crex.Wrapf(ErrBuild, "get manifest: %w", err)
	}

	srcClean := strings.Trim(srcPath, "/")
	found, err := collectLayerEntries(ctx, cs, mfst.Layers, srcClean)
	if err != nil {
		return err
	}

	if len(found) == 0 {
		return crex.Wrapf(ErrBuild, "path %q not found in image", srcPath)
	}

	return writeExtractedTar(w, found)
}

// Iterates layers in order, collecting all entries under srcClean into found
// with last-writer-wins semantics and applying whiteout deletions.
func collectLayerEntries(ctx context.Context, cs content.Store, layers []ocispec.Descriptor, srcClean string) (map[string]layerEntry, error) {
	found := make(map[string]layerEntry)
	for _, layer := range layers {
		if err := readLayerIntoFound(ctx, cs, layer, srcClean, found); err != nil {
			return nil, err
		}
	}
	return found, nil
}

// Reads a single layer's tar entries into found, applying whiteout deletions
// and skipping entries that do not match srcClean.
func readLayerIntoFound(ctx context.Context, cs content.Store, layer ocispec.Descriptor, srcClean string, found map[string]layerEntry) error {
	ra, err := cs.ReaderAt(ctx, layer)
	if err != nil {
		return crex.Wrapf(ErrBuild, "read layer %s: %w", layer.Digest, err)
	}
	defer ra.Close()

	var r io.Reader = content.NewReader(ra)
	switch layer.MediaType {
	case ocispec.MediaTypeImageLayerGzip, images.MediaTypeDockerSchema2LayerGzip:
		gr, err := gzip.NewReader(r)
		if err != nil {
			return crex.Wrapf(ErrBuild, "decompress layer: %w", err)
		}
		defer gr.Close()
		r = gr
	}

	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return crex.Wrapf(ErrBuild, "read tar entry: %w", err)
		}

		name := strings.TrimPrefix(strings.TrimPrefix(hdr.Name, "./"), "/")
		base := filepath.Base(name)

		if strings.HasPrefix(base, ".wh.") {
			applyWhiteout(found, name, base)
			if _, err := io.Copy(io.Discard, tr); err != nil {
				return err
			}
			continue
		}

		if name != srcClean && !strings.HasPrefix(name, srcClean+"/") {
			if _, err := io.Copy(io.Discard, tr); err != nil {
				return err
			}
			continue
		}

		data, err := io.ReadAll(tr)
		if err != nil {
			return crex.Wrapf(ErrBuild, "read entry %q: %w", name, err)
		}
		found[name] = layerEntry{hdr: *hdr, data: data}
	}
	return nil
}

// Applies a whiteout tar entry to found.
//
// An opaque whiteout (.wh..wh..opq) deletes all entries under the same
// directory. A regular whiteout (.wh.<name>) deletes the named sibling entry.
func applyWhiteout(found map[string]layerEntry, name, base string) {
	if base == ".wh..wh..opq" {
		dir := filepath.Dir(name)
		if dir == "." {
			dir = ""
		}
		for k := range found {
			if dir == "" || k == dir || strings.HasPrefix(k, dir+"/") {
				delete(found, k)
			}
		}
	} else {
		whTarget := filepath.Join(filepath.Dir(name), strings.TrimPrefix(base, ".wh."))
		delete(found, whTarget)
	}
}

// Writes the entries in found to w as an uncompressed tar stream.
func writeExtractedTar(w io.Writer, found map[string]layerEntry) error {
	tw := tar.NewWriter(w)
	for name, e := range found {
		e.hdr.Name = name
		e.hdr.Size = int64(len(e.data))
		if err := tw.WriteHeader(&e.hdr); err != nil {
			return crex.Wrapf(ErrBuild, "write header for %q: %w", name, err)
		}
		if _, err := tw.Write(e.data); err != nil {
			return crex.Wrapf(ErrBuild, "write data for %q: %w", name, err)
		}
	}
	return tw.Close()
}

// Exports the image as an OCI tar archive to w.
func (b *builder) Export(ctx context.Context, imageRef string, w io.Writer) error {
	ctx = namespaces.WithNamespace(ctx, crucibleNamespace)
	return b.client.Export(ctx, w, archive.WithImage(b.client.ImageService(), imageRef))
}

// Releases the containerd connection and removes intermediate images.
func (b *builder) Close() error {
	b.mu.Lock()
	created := b.created
	b.created = nil
	b.mu.Unlock()

	ctx := namespaces.WithNamespace(context.Background(), crucibleNamespace)
	for _, ref := range created {
		b.client.ImageService().Delete(ctx, ref)
	}
	return b.client.Close()
}

// Starts a task in container, waits for it to exit, and returns an error for
// non-zero exit codes.
//
// stdout and stderr receive the task's output streams. When either is nil,
// the corresponding process stream is used as a fallback.
func (b *builder) runTask(ctx context.Context, container containerd.Container, stdout, stderr io.Writer) error {
	// FIFOs do not work across the macOS/Lima VM kernel boundary via virtiofs.
	// Use a log file in the virtiofs-accessible cache dir instead.
	logDir, _ := os.UserCacheDir()
	logDir = filepath.Join(logDir, "crux", "logs")
	os.MkdirAll(logDir, 0o700)
	logPath := filepath.Join(logDir, container.ID()+".log")
	if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600); err == nil {
		f.Close()
	}
	defer os.Remove(logPath)

	task, err := container.NewTask(ctx, cio.LogFile(logPath))
	if err != nil {
		return crex.Wrapf(ErrBuild, "create task: %w", err)
	}

	exitCh, err := task.Wait(ctx)
	if err != nil {
		task.Delete(ctx)
		return crex.Wrapf(ErrBuild, "wait: %w", err)
	}

	if err := task.Start(ctx); err != nil {
		task.Delete(ctx)
		return crex.Wrapf(ErrBuild, "start task: %w", err)
	}

	var taskErr error
	select {
	case status := <-exitCh:
		task.Delete(ctx)
		if err := status.Error(); err != nil {
			taskErr = crex.Wrapf(ErrBuild, "task exit: %w", err)
		} else if code := status.ExitCode(); code != 0 {
			taskErr = crex.Wrapf(ErrBuild, "command exited with code %d", code)
		}
	case <-ctx.Done():
		task.Kill(ctx, syscall.SIGTERM)
		task.Delete(ctx)
		taskErr = ctx.Err()
	}

	// Forward log file output to the callers' writers (combined stdout+stderr).
	if stdout != nil || stderr != nil {
		if data, err := os.ReadFile(logPath); err == nil && len(data) > 0 {
			w := stdout
			if w == nil {
				w = stderr
			}
			w.Write(data)
		}
	}

	return taskErr
}

// Computes the digest of the uncompressed content behind desc.
//
// Reads the compressed blob from the content store, decompresses it, and
// returns the sha256 digest of the resulting stream.
func (b *builder) uncompressedDigest(ctx context.Context, desc ocispec.Descriptor) (digest.Digest, error) {
	ra, err := b.client.ContentStore().ReaderAt(ctx, desc)
	if err != nil {
		return "", err
	}
	defer ra.Close()

	var r io.Reader = content.NewReader(ra)

	switch desc.MediaType {
	case ocispec.MediaTypeImageLayerGzip, images.MediaTypeDockerSchema2LayerGzip:
		gr, err := gzip.NewReader(r)
		if err != nil {
			return "", err
		}
		defer gr.Close()
		r = gr
	}

	digester := digest.SHA256.Digester()
	if _, err := io.Copy(digester.Hash(), r); err != nil {
		return "", err
	}
	return digester.Digest(), nil
}

// Creates a new image by appending a layer to img's manifest.
//
// The new config includes diffID in its DiffIDs and a history entry with
// createdBy. Returns the new image reference.
func (b *builder) appendLayerToImage(ctx context.Context, img containerd.Image, layerDesc ocispec.Descriptor, diffID digest.Digest, createdBy string) (string, error) {
	cs := b.client.ContentStore()

	mfst, err := images.Manifest(ctx, cs, img.Target(), nil)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "get manifest: %w", err)
	}

	configData, err := content.ReadBlob(ctx, cs, mfst.Config)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "read config: %w", err)
	}

	var cfg ocispec.Image
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return "", crex.Wrapf(ErrBuild, "unmarshal config: %w", err)
	}

	cfg.RootFS.DiffIDs = append(cfg.RootFS.DiffIDs, diffID)
	cfg.History = append(cfg.History, ocispec.History{
		CreatedBy: createdBy,
	})

	return b.writeNewImage(ctx, mfst, cfg, &layerDesc)
}

// Writes a new image to the content store based on the parent manifest and
// updated config. When extraLayer is non-nil it is appended to the manifest's
// layer list. Returns the new image reference.
func (b *builder) writeNewImage(ctx context.Context, parentMfst ocispec.Manifest, cfg ocispec.Image, extraLayer *ocispec.Descriptor) (string, error) {
	mfstDesc, err := b.writeImageManifest(ctx, parentMfst, cfg, extraLayer)
	if err != nil {
		return "", err
	}
	return b.registerImageRef(ctx, mfstDesc)
}

// Serialises cfg and layers into config and manifest blobs in the content
// store. Sets GC reference labels on the manifest so the garbage collector
// can trace reachability from the manifest to its config and layer blobs.
// Returns the manifest descriptor.
func (b *builder) writeImageManifest(ctx context.Context, parentMfst ocispec.Manifest, cfg ocispec.Image, extraLayer *ocispec.Descriptor) (ocispec.Descriptor, error) {
	cs := b.client.ContentStore()

	// Write new config blob.
	newConfigData, err := json.Marshal(cfg)
	if err != nil {
		return ocispec.Descriptor{}, crex.Wrapf(ErrBuild, "marshal config: %w", err)
	}
	newConfigDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    digest.FromBytes(newConfigData),
		Size:      int64(len(newConfigData)),
	}
	if err := content.WriteBlob(ctx, cs,
		"config-"+newConfigDesc.Digest.Encoded(),
		bytes.NewReader(newConfigData), newConfigDesc,
	); err != nil {
		return ocispec.Descriptor{}, crex.Wrapf(ErrBuild, "write config: %w", err)
	}

	// Build new layer list.
	layers := append([]ocispec.Descriptor{}, parentMfst.Layers...)
	if extraLayer != nil {
		layers = append(layers, *extraLayer)
	}

	// Build GC reference labels for the manifest so the garbage collector
	// knows which config and layer blobs are reachable through it. Without
	// these labels the GC may delete layer blobs before Unpack runs.
	gcLabels := map[string]string{
		"containerd.io/gc.ref.content.config": newConfigDesc.Digest.String(),
	}
	for i, layer := range layers {
		gcLabels[fmt.Sprintf("containerd.io/gc.ref.content.l.%d", i)] = layer.Digest.String()
	}

	// Write new manifest blob.
	newMfst := ocispec.Manifest{
		Versioned: imgspecs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageManifest,
		Config:    newConfigDesc,
		Layers:    layers,
	}
	newMfstData, err := json.Marshal(newMfst)
	if err != nil {
		return ocispec.Descriptor{}, crex.Wrapf(ErrBuild, "marshal manifest: %w", err)
	}
	newMfstDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(newMfstData),
		Size:      int64(len(newMfstData)),
	}

	if err := content.WriteBlob(ctx, cs,
		"manifest-"+newMfstDesc.Digest.Encoded(),
		bytes.NewReader(newMfstData), newMfstDesc,
		content.WithLabels(gcLabels),
	); err != nil {
		return ocispec.Descriptor{}, crex.Wrapf(ErrBuild, "write manifest: %w", err)
	}

	// Ensure GC labels are present even when the manifest blob already
	// existed (WriteBlob returns nil without setting labels on AlreadyExists).
	if _, err := cs.Update(ctx, content.Info{
		Digest: newMfstDesc.Digest,
		Labels: gcLabels,
	}, "labels"); err != nil {
		return ocispec.Descriptor{}, crex.Wrapf(ErrBuild, "update manifest gc labels: %w", err)
	}

	return newMfstDesc, nil
}

// Upserts the image record in the image service under a name derived from
// mfstDesc. If a record for this digest already exists (e.g. from a prior
// build with identical content), it is updated in place rather than failing.
// Returns the image reference.
func (b *builder) registerImageRef(ctx context.Context, mfstDesc ocispec.Descriptor) (string, error) {
	ref := fmt.Sprintf("crucible/build:%s", mfstDesc.Digest.Encoded()[:16])
	img := images.Image{Name: ref, Target: mfstDesc}
	if _, err := b.client.ImageService().Create(ctx, img); err != nil {
		if !errdefs.IsAlreadyExists(err) {
			return "", crex.Wrapf(ErrBuild, "register image: %w", err)
		}
		if _, err := b.client.ImageService().Update(ctx, img); err != nil {
			return "", crex.Wrapf(ErrBuild, "register image: %w", err)
		}
	}
	return ref, nil
}

// Applies a declarative [provider.ConfigUpdate] to cfg in place.
func applyConfigUpdate(cfg *ocispec.Image, u *provider.ConfigUpdate) {
	for k, v := range u.AddEnv {
		cfg.Config.Env = mergeEnv(cfg.Config.Env, []string{k + "=" + v})
	}
	if u.SetWorkDir != "" {
		cfg.Config.WorkingDir = u.SetWorkDir
	}
	if u.SetUser != "" {
		cfg.Config.User = u.SetUser
	}
	if u.SetEntrypoint != nil {
		cfg.Config.Entrypoint = u.SetEntrypoint
	}
	if u.SetCmd != nil {
		cfg.Config.Cmd = u.SetCmd
	}
}

// Adds ref to the list of images to be cleaned up on Close.
func (b *builder) trackImage(ref string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.created = append(b.created, ref)
}

// Returns a random hex identifier for use as a step or snapshot ID.
func newStepID() string {
	var buf [8]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

// Returns an [oci.SpecOpts] that applies the command, environment, working
// directory, and user from a build step to the container spec.
func withBuildStep(cfg *provider.RunConfig) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		shell := "/bin/sh"
		if cfg.Shell != "" {
			shell = cfg.Shell
		}
		s.Process.Args = []string{shell, "-c", cfg.Command}

		if len(cfg.Env) > 0 {
			s.Process.Env = mergeEnv(s.Process.Env, cfg.Env)
		}
		if cfg.WorkDir != "" {
			s.Process.Cwd = cfg.WorkDir
		}
		if cfg.User != "" {
			parts := strings.SplitN(cfg.User, ":", 2)
			uid, err := strconv.ParseUint(parts[0], 10, 32)
			if err == nil {
				s.Process.User.UID = uint32(uid)
			}
			if len(parts) == 2 {
				gid, err := strconv.ParseUint(parts[1], 10, 32)
				if err == nil {
					s.Process.User.GID = uint32(gid)
				}
			}
		}
		return nil
	}
}

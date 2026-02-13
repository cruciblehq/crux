//go:build linux

package runtime

import (
	_ "embed"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/cruciblehq/crux/kit/archive"
	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/paths"
)

const (

	// Binary name for the containerd daemon.
	containerdBin = "containerd"

	// Directory name used for containerd data and state.
	containerdDir = "containerd"

	// Configuration filename written to the containerd data directory.
	containerdConfigFile = "config.toml"
)

// Extracts containerd binaries from a gzipped tar archive.
//
// All entries are extracted into destDir preserving the original directory
// structure and executable permissions. Returns [ErrContainerdNotFound] if
// the containerd binary is not present in the archive.
func extractContainerd(r io.Reader, destDir string) error {
	if err := archive.ExtractFromReader(r, destDir, archive.Gzip); err != nil {
		return crex.Wrap(ErrContainerdDownload, err)
	}

	if _, err := os.Stat(filepath.Join(destDir, "bin", containerdBin)); err != nil {
		return ErrContainerdNotFound
	}
	return nil
}

//go:embed templates/containerd.toml.tmpl
var configTemplateSource string

// Containerd TOML configuration template.
var configTemplate = template.Must(template.New("containerd").Parse(configTemplateSource))

// Values injected into the containerd config template.
type configData struct {
	Root    string // Persistent data directory.
	State   string // Runtime state directory.
	Address string // gRPC socket path.
}

// Writes a containerd configuration file to disk.
//
// The path to the generated file is returned. Directories are derived from
// the [paths] package:
//
//	root:    paths.Data()/containerd
//	state:   paths.Runtime()/containerd
//	address: paths.Runtime()/containerd.sock
func generateConfig() (string, error) {
	dataDir := paths.Data()
	runtimeDir := paths.Runtime()

	data := configData{
		Root:    filepath.Join(dataDir, containerdDir),
		State:   filepath.Join(runtimeDir, containerdDir),
		Address: filepath.Join(runtimeDir, containerdSock),
	}

	configDir := filepath.Join(dataDir, containerdDir)
	if err := os.MkdirAll(configDir, paths.DefaultDirMode); err != nil {
		return "", crex.Wrap(ErrContainerdConfig, err)
	}

	configPath := filepath.Join(configDir, containerdConfigFile)
	f, err := os.Create(configPath)
	if err != nil {
		return "", crex.Wrap(ErrContainerdConfig, err)
	}
	defer f.Close()

	if err := configTemplate.Execute(f, data); err != nil {
		return "", crex.Wrap(ErrContainerdConfig, err)
	}

	return configPath, nil
}

//go:build darwin

package vm

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/paths"
)

const (

	// Lima version to use for the crux VM.
	limaVersion = "2.0.3"

	// Download URL template for Lima releases.
	// Placeholders: version, version, OS, arch.
	limaDownloadURL = "https://github.com/lima-vm/lima/releases/download/v%s/lima-%s-%s-%s.tar.gz"

	// Binary name for the Lima CLI.
	limactlBin = "limactl"

	// Go GOARCH values.
	goarchARM64 = "arm64"
	goarchAMD64 = "amd64"

	// Architecture identifiers used in Lima YAML configuration.
	limaArchARM64 = "aarch64"
	limaArchAMD64 = "x86_64"

	// Architecture identifiers used in Darwin release asset filenames.
	downloadArchARM64 = "arm64"
	downloadArchAMD64 = "x86_64"
)

// Returns the Lima architecture identifier for the YAML config.
func limaArch() string {
	switch runtime.GOARCH {
	case goarchARM64:
		return limaArchARM64
	case goarchAMD64:
		return limaArchAMD64
	default:
		return limaArchAMD64
	}
}

// Returns the architecture identifier for Darwin release asset URLs.
func downloadArch() string {
	switch runtime.GOARCH {
	case goarchARM64:
		return downloadArchARM64
	case goarchAMD64:
		return downloadArchAMD64
	default:
		return downloadArchAMD64
	}
}

// Returns the path to the vendored limactl binary.
//
// The binary is stored in the crux data directory so it persists across
// sessions and does not require system-wide installation.
func limactlPath() string {
	return filepath.Join(limaDir(), "bin", limactlBin)
}

// Returns the root directory where Lima is extracted.
func limaDir() string {
	return filepath.Join(paths.Data(), "lima")
}

// Ensures the limactl binary is available, downloading it if necessary.
//
// Returns the absolute path to the limactl binary. If the binary does not
// exist at the expected location, the full Lima distribution is downloaded
// from GitHub releases and extracted.
func ensureLima() (string, error) {
	bin := limactlPath()
	if _, err := os.Stat(bin); err == nil {
		return bin, nil
	}

	if err := downloadLima(limaDir()); err != nil {
		return "", err
	}
	return bin, nil
}

// Downloads and extracts Lima from GitHub releases.
func downloadLima(dest string) error {
	url := fmt.Sprintf(limaDownloadURL, limaVersion, limaVersion, "Darwin", downloadArch())

	resp, err := http.Get(url)
	if err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return crex.Wrap(ErrLimaDownload, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url))
	}

	return extractLima(resp.Body, dest)
}

// Extracts the Lima distribution from a gzipped tar archive.
//
// Extracts all regular files from the archive into the destination directory,
// preserving the archive's internal structure. This includes the limactl binary
// and supporting files like guest agents that Lima requires at runtime.
func extractLima(r io.Reader, dest string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return crex.Wrap(ErrLimaDownload, err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		target := filepath.Join(dest, filepath.Clean(header.Name))

		if err := os.MkdirAll(filepath.Dir(target), paths.DefaultDirMode); err != nil {
			return crex.Wrap(ErrLimaDownload, err)
		}

		mode := os.FileMode(header.Mode)
		f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			return crex.Wrap(ErrLimaDownload, err)
		}

		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return crex.Wrap(ErrLimaDownload, err)
		}
		f.Close()
	}

	// Verify limactl was extracted.
	limactlDest := filepath.Join(dest, "bin", limactlBin)
	if _, err := os.Stat(limactlDest); err != nil {
		return crex.Wrap(ErrLimaDownload, fmt.Errorf("limactl not found in archive"))
	}

	return nil
}

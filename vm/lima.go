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
	// Placeholders: version, os, arch (aarch64/x86_64).
	limaDownloadURL = "https://github.com/lima-vm/lima/releases/download/v%s/lima-%s-%s-%s.tar.gz"

	// Binary name for the Lima CLI.
	limactlBin = "limactl"

	// Go GOARCH values.
	goarchARM64 = "arm64"
	goarchAMD64 = "amd64"

	// Lima architecture identifiers.
	limaArchARM64 = "aarch64"
	limaArchAMD64 = "x86_64"
)

// Returns the Lima architecture identifier for the current host.
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

// Returns the path to the vendored limactl binary.
//
// The binary is stored in the crux data directory so it persists across
// sessions and does not require system-wide installation.
func limactlPath() string {
	return filepath.Join(paths.Data(), "bin", limactlBin)
}

// Ensures the limactl binary is available, downloading it if necessary.
//
// Returns the absolute path to the limactl binary. If the binary does not
// exist at the expected location, it is downloaded from the Lima GitHub
// releases and extracted.
func ensureLima() (string, error) {
	bin := limactlPath()
	if _, err := os.Stat(bin); err == nil {
		return bin, nil
	}

	if err := downloadLima(bin); err != nil {
		return "", err
	}
	return bin, nil
}

// Downloads and extracts the limactl binary from GitHub releases.
func downloadLima(dest string) error {
	url := fmt.Sprintf(limaDownloadURL, limaVersion, limaVersion, runtime.GOOS, limaArch())

	resp, err := http.Get(url)
	if err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return crex.Wrap(ErrLimaDownload, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url))
	}

	return extractLimactl(resp.Body, dest)
}

// Extracts the limactl binary from a gzipped tar archive.
//
// Scans the archive for the bin/limactl entry and writes it to dest. All other
// entries are skipped. The destination directory is created if needed.
func extractLimactl(r io.Reader, dest string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return crex.Wrap(ErrLimaDownload, fmt.Errorf("limactl not found in archive"))
		}
		if err != nil {
			return crex.Wrap(ErrLimaDownload, err)
		}

		if filepath.Base(header.Name) != limactlBin || header.Typeflag != tar.TypeReg {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(dest), paths.DefaultDirMode); err != nil {
			return crex.Wrap(ErrLimaDownload, err)
		}

		f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, paths.DefaultDirMode)
		if err != nil {
			return crex.Wrap(ErrLimaDownload, err)
		}
		defer f.Close()

		if _, err := io.Copy(f, tr); err != nil {
			return crex.Wrap(ErrLimaDownload, err)
		}
		return nil
	}
}

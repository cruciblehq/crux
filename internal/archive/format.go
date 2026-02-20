package archive

import (
	"compress/gzip"
	"io"
	"strings"

	"github.com/klauspost/compress/zstd"
)

// Supported archive compression formats.
//
// Each format corresponds to a tar archive compressed with a specific
// algorithm. The format is inferred from the file extension by [Create] and
// [Extract], or supplied explicitly to [ExtractFromReader].
type Format int

const (

	// Zstandard compression (.tar.zst).
	Zstd Format = iota

	// Gzip compression (.tar.gz, .tgz).
	Gzip

	// File extension for Zstandard-compressed tar archives.
	extZstd = ".tar.zst"

	// File extension for Gzip-compressed tar archives.
	extGzip = ".tar.gz"

	// Alternate file extension for Gzip-compressed tar archives.
	extTgz = ".tgz"
)

// String returns the canonical file extension for the format.
func (f Format) String() string {
	switch f {
	case Zstd:
		return extZstd
	case Gzip:
		return extGzip
	default:
		return ""
	}
}

// Detects the archive format from a filename.
//
// Returns [ErrUnsupportedFormat] if the extension is not recognised.
func detect(name string) (Format, error) {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, extZstd):
		return Zstd, nil
	case strings.HasSuffix(lower, extGzip):
		return Gzip, nil
	case strings.HasSuffix(lower, extTgz):
		return Gzip, nil
	default:
		return 0, ErrUnsupportedFormat
	}
}

// Returns a write-closer that compresses data with the given format.
func newCompressWriter(w io.Writer, f Format) (io.WriteCloser, error) {
	switch f {
	case Zstd:
		return zstd.NewWriter(w)
	case Gzip:
		return gzip.NewWriter(w), nil
	default:
		return nil, ErrUnsupportedFormat
	}
}

// Returns a read-closer that decompresses data with the given format.
func newDecompressReader(r io.Reader, f Format) (io.ReadCloser, error) {
	switch f {
	case Zstd:
		zr, err := zstd.NewReader(r)
		if err != nil {
			return nil, err
		}
		return zr.IOReadCloser(), nil
	case Gzip:
		return gzip.NewReader(r)
	default:
		return nil, ErrUnsupportedFormat
	}
}

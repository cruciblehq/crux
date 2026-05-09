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
	Zstd Format = iota // Zstandard compression (.tar.zst).
	Gzip               // Gzip compression (.tar.gz, .tgz).
	Tar                // Plain tar, no compression (.tar).
)

const (
	extZstd = ".tar.zst" // File extension for Zstandard-compressed tar archives.
	extGzip = ".tar.gz"  // File extension for Gzip-compressed tar archives.
	extTgz  = ".tgz"     // Alternate file extension for Gzip-compressed tar archives.
	extTar  = ".tar"     // File extension for plain tar archives.
)

// Returns the canonical file extension for the format.
func (f Format) String() string {
	switch f {
	case Zstd:
		return extZstd
	case Gzip:
		return extGzip
	case Tar:
		return extTar
	default:
		return ""
	}
}

// Detects the archive format from a filename.
//
// The check is case-insensitive and looks only at the end of the filename.
// This means that ".TAR.GZ" will be detected as Gzip even though it has an
// extra ".TAR" segment. Returns [ErrUnsupportedFormat] if the extension is
// not recognised.
func detect(name string) (Format, error) {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, extZstd):
		return Zstd, nil
	case strings.HasSuffix(lower, extGzip):
		return Gzip, nil
	case strings.HasSuffix(lower, extTgz):
		return Gzip, nil
	case strings.HasSuffix(lower, extTar):
		return Tar, nil
	default:
		return 0, ErrUnsupportedFormat
	}
}

// Wraps a writer with a no-op Close method.
type nopWriteCloser struct {
	io.Writer
}

// No-op Close method.
func (nopWriteCloser) Close() error {
	return nil
}

// Returns a write-closer that compresses data with the given format.
//
// The caller must call Close on the returned writer to flush all data and
// release resources. For the Tar format, the returned writer's Close method
// is a no-op and does not need to be called. The returned writer will write
// to the provided writer and compress the data according to the specified format.
func newCompressWriter(w io.Writer, f Format) (io.WriteCloser, error) {
	switch f {
	case Zstd:
		return zstd.NewWriter(w)
	case Gzip:
		return gzip.NewWriter(w), nil
	case Tar:
		return nopWriteCloser{w}, nil
	default:
		return nil, ErrUnsupportedFormat
	}
}

// Returns a read-closer that decompresses data with the given format.
//
// The caller must call Close on the returned reader to release resources. For
// the Tar format, the returned reader's Close method is a no-op and does not
// need to be called. The returned reader will read from the provided reader
// and decompress the data according to the specified format.
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
	case Tar:
		return io.NopCloser(r), nil
	default:
		return nil, ErrUnsupportedFormat
	}
}

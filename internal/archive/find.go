package archive

import (
	"archive/tar"
	"io"

	"github.com/cruciblehq/crex"
)

// Reads a named file from a tar archive.
//
// Scans the tar reader sequentially until filename is found or the archive is
// exhausted. Returns the file contents and nil error on success, (nil, nil) if
// the file is not present, or (nil, error) if a read error occurs. The tar
// reader is advanced past the matched entry and cannot be rewound.
func Find(tr *tar.Reader, filename string) ([]byte, error) {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil, nil
		}
		if err != nil {
			return nil, crex.Wrap(ErrReadFailed, err)
		}

		if header.Name == filename {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, crex.Wrap(ErrReadFailed, err)
			}
			return data, nil
		}
	}
}

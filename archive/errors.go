package archive

import "github.com/cruciblehq/crux/crex"

var (
	ErrCreateFailed        = crex.New("archive creation failed")
	ErrExtractFailed       = crex.New("archive extraction failed")
	ErrReadFailed          = crex.New("archive read failed")
	ErrNotFound            = crex.New("file not found in archive")
	ErrInvalidPath         = crex.New("invalid archive path")
	ErrUnsupportedFileType = crex.New("unsupported file type")
	ErrUnsupportedFormat   = crex.New("unsupported archive format")
)

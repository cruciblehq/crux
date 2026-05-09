package archive

import "errors"

var (
	ErrCreateFailed        = errors.New("archive creation failed")
	ErrExtractFailed       = errors.New("archive extraction failed")
	ErrReadFailed          = errors.New("archive read failed")
	ErrNotFound            = errors.New("file not found in archive")
	ErrInvalidPath         = errors.New("invalid archive path")
	ErrUnsupportedFileType = errors.New("unsupported file type")
	ErrUnsupportedFormat   = errors.New("unsupported archive format")
)

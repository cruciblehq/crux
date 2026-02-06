package archive

import "errors"

var (
	ErrCreateFailed        = errors.New("archive creation failed")
	ErrExtractFailed       = errors.New("extraction failed")
	ErrInvalidPath         = errors.New("invalid path")
	ErrUnsupportedFileType = errors.New("unsupported file type")
)

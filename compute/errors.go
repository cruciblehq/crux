package compute

import "errors"

var (
	ErrUnknownProvider = errors.New("unknown provider")
	ErrConnect         = errors.New("containerd connection failed")
	ErrImport          = errors.New("image import failed")
	ErrNoImages        = errors.New("no images in archive")
	ErrContainer       = errors.New("container operation failed")
)

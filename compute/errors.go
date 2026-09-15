package compute

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrUnknownProvider = crex.New("unknown provider")
	ErrConnect         = crex.New("containerd connection failed")
	ErrImport          = crex.New("image import failed")
	ErrNoImages        = crex.New("no images in archive")
	ErrContainer       = crex.New("container operation failed")
	ErrStorage         = crex.New("storage operation failed")
	ErrNetwork         = crex.New("network operation failed")
)

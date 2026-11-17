package machines

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrNotFound      = crex.New("machine record not found")
	ErrInvalidRecord = crex.New("invalid machine record")
)

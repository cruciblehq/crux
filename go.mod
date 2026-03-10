module github.com/cruciblehq/crux

go 1.25.1

require (
	github.com/adrg/xdg v0.5.3
	github.com/alecthomas/kong v1.13.0
	github.com/cruciblehq/crex v0.1.0
	github.com/cruciblehq/spec v0.3.6
	github.com/evanw/esbuild v0.27.2
	github.com/fsnotify/fsnotify v1.9.0
	github.com/klauspost/compress v1.18.4
	golang.org/x/sys v0.40.0
)

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/cruciblehq/spec => ../spec

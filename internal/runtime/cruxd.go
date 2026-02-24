//go:build darwin || linux

package runtime

const (

	// Pinned cruxd version. This must match a published release at
	// github.com/cruciblehq/cruxd so that crux always downloads a
	// known-compatible daemon binary.
	cruxdVersion = "0.1.3"

	// Download URL template for the cruxd release archive. The single
	// placeholder is the Linux architecture (e.g. "amd64", "aarch64").
	cruxdDownloadURL = "https://github.com/cruciblehq/cruxd/releases/download/v" + cruxdVersion + "/cruxd-linux-%s.tar.gz"
)

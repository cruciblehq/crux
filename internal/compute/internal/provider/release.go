package provider

import "fmt"

// Maps Go GOARCH values to the architecture identifiers used in cruxd
// release asset filenames.
var architectures = map[string]string{
	"amd64": "amd64",
	"arm64": "aarch64",
}

// Returns the download URL for the cruxd release archive targeting the given
// version and Go architecture (e.g. "arm64", "amd64").
func CruxdDownloadURL(version, goarch string) string {
	arch := architectures[goarch]
	if arch == "" {
		arch = goarch
	}
	return fmt.Sprintf("https://github.com/cruciblehq/cruxd/releases/download/v%s/cruxd-linux-%s.tar.gz", version, arch)
}

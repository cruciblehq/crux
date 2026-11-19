//go:build linux

package lima

import (
	_ "embed"
	"text/template"
)

// Source of the Lima YAML configuration template for Linux.
//
//go:embed templates/lima_linux.yaml.tmpl
var limaConfigSource string

// Parses the Linux template into [limaConfigTemplate].
func init() {
	limaConfigTemplate = template.Must(template.New("lima").Parse(limaConfigSource))
}

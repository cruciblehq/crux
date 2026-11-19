//go:build darwin

package lima

import (
	_ "embed"
	"text/template"
)

// Source of the Lima YAML configuration template for macOS.
//
//go:embed templates/lima_darwin.yaml.tmpl
var limaConfigSource string

// Parses the macOS template into [limaConfigTemplate].
func init() {
	limaConfigTemplate = template.Must(template.New("lima").Parse(limaConfigSource))
}

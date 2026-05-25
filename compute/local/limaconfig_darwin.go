//go:build darwin

package local

import (
	_ "embed"
	"text/template"
)

//go:embed templates/lima.yaml.tmpl
var limaConfigSource string

func init() {
	limaConfigTemplate = template.Must(template.New("lima").Parse(limaConfigSource))
}

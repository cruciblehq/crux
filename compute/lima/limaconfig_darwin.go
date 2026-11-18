//go:build darwin

package lima

import (
	_ "embed"
	"text/template"
)

//go:embed templates/lima_darwin.yaml.tmpl
var limaConfigSource string

func init() {
	limaConfigTemplate = template.Must(template.New("lima").Parse(limaConfigSource))
}

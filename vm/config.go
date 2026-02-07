//go:build darwin

package vm

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/paths"
)

const (

	// Lima instance name used for the crux VM.
	limaInstanceName = "crux"

	// Configuration file name written to paths.VM().
	limaConfigFile = "lima.yaml"

	// Default number of virtual CPUs allocated to the VM.
	defaultCPUs = 2

	// Default memory in GiB allocated to the VM.
	defaultMemoryGiB = 2

	// Default disk size in GiB allocated to the VM.
	defaultDiskGiB = 10
)

//go:embed templates/lima.yaml.tmpl
var configTemplateSource string

// Lima YAML configuration template.
//
// Uses Virtualization.framework (vz) on macOS with virtiofs mounts. The
// provisioning script installs containerd and CNI plugins on first boot.
var configTemplate = template.Must(template.New("lima").Parse(configTemplateSource))

// Holds the values injected into the Lima YAML template.
type configData struct {
	Arch    string // Lima architecture identifier (e.g. "aarch64", "x86_64").
	CPUs    int    // Number of virtual CPUs.
	Memory  string // Memory allocation with unit suffix (e.g. "2GiB").
	Disk    string // Disk size with unit suffix (e.g. "10GiB").
	DataDir string // Host directory mounted into the VM at /mnt/crux.
}

// Generates the Lima YAML configuration for the crux VM.
//
// The configuration targets the host's native architecture and uses sensible
// defaults for CPU, memory, and disk allocation.
func generateConfig() (string, error) {
	data := configData{
		Arch:    limaArch(),
		CPUs:    defaultCPUs,
		Memory:  fmt.Sprintf("%dGiB", defaultMemoryGiB),
		Disk:    fmt.Sprintf("%dGiB", defaultDiskGiB),
		DataDir: paths.Data(),
	}

	configDir := paths.VM()
	if err := os.MkdirAll(configDir, paths.DefaultDirMode); err != nil {
		return "", crex.Wrap(ErrVMConfig, err)
	}

	configPath := filepath.Join(configDir, limaConfigFile)
	f, err := os.Create(configPath)
	if err != nil {
		return "", crex.Wrap(ErrVMConfig, err)
	}
	defer f.Close()

	if err := configTemplate.Execute(f, data); err != nil {
		return "", crex.Wrap(ErrVMConfig, err)
	}

	return configPath, nil
}

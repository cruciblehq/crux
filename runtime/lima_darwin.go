//go:build darwin

package runtime

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"text/template"

	"github.com/cruciblehq/crux/kit/archive"
	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/paths"
)

const (

	// Lima version to use for the crux VM.
	limaVersion = "2.0.3"

	// Lima instance name used for the crux VM.
	limaInstanceName = "crux"

	// Configuration file name written to paths.VM().
	limaConfigFile = "lima.yaml"

	// Status string returned by limactl when the VM is running.
	limaStatusRunning = "Running"

	// Status string returned by limactl when the VM is stopped.
	limaStatusStopped = "Stopped"

	// Default number of virtual CPUs allocated to the VM.
	defaultCPUs = 2

	// Default memory in GiB allocated to the VM.
	defaultMemoryGiB = 2

	// Default disk size in GiB allocated to the VM.
	defaultDiskGiB = 10

	// Download URL template for Lima releases.
	// Placeholders: version, version, OS, arch.
	limaDownloadURL = "https://github.com/lima-vm/lima/releases/download/v%s/lima-%s-%s-%s.tar.gz"

	// Binary name for the Lima CLI.
	limactlBin = "limactl"

	// Go GOARCH values.
	goarchARM64 = "arm64"
	goarchAMD64 = "amd64"

	// Architecture identifiers used in Lima YAML configuration.
	limaArchARM64 = "aarch64"
	limaArchAMD64 = "x86_64"

	// Architecture identifiers used in Darwin release asset filenames.
	downloadArchARM64 = "arm64"
	downloadArchAMD64 = "x86_64"
)

// Lima architecture identifier for the YAML config.
func limaArch() string {
	switch goruntime.GOARCH {
	case goarchARM64:
		return limaArchARM64
	case goarchAMD64:
		return limaArchAMD64
	default:
		return limaArchAMD64
	}
}

// Architecture identifier for Darwin release asset URLs.
func downloadArch() string {
	switch goruntime.GOARCH {
	case goarchARM64:
		return downloadArchARM64
	case goarchAMD64:
		return downloadArchAMD64
	default:
		return downloadArchAMD64
	}
}

// Path to the vendored limactl binary.
//
// The binary is stored in the crux data directory so it persists across
// sessions and does not require system-wide installation.
func limactlPath() string {
	return filepath.Join(limaDir(), "bin", limactlBin)
}

// Root directory where Lima is extracted.
func limaDir() string {
	return filepath.Join(paths.Data(), "lima")
}

// Ensures the runtime binary is available, downloading it if necessary.
//
// The absolute path to the limactl binary is returned. If the binary does
// not exist at the expected location, the full Lima distribution is
// downloaded from GitHub releases and extracted.
func ensureBinary() (string, error) {
	bin := limactlPath()
	if _, err := os.Stat(bin); err == nil {
		return bin, nil
	}

	if err := downloadLima(limaDir()); err != nil {
		return "", err
	}
	return bin, nil
}

// Downloads and extracts Lima from GitHub releases.
func downloadLima(dest string) error {
	url := fmt.Sprintf(limaDownloadURL, limaVersion, limaVersion, "Darwin", downloadArch())

	resp, err := http.Get(url)
	if err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return crex.Wrap(ErrLimaDownload, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url))
	}

	return extractLima(resp.Body, dest)
}

// Extracts the Lima distribution from a gzipped tar archive.
//
// All entries are extracted into the destination directory preserving the
// archive's internal structure and executable permissions. This includes
// the limactl binary and supporting files like guest agents.
func extractLima(r io.Reader, dest string) error {
	if err := archive.ExtractFromReader(r, dest, archive.Gzip); err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}

	if _, err := os.Stat(filepath.Join(dest, "bin", limactlBin)); err != nil {
		return crex.Wrap(ErrLimaDownload, fmt.Errorf("limactl not found in archive"))
	}
	return nil
}

//go:embed templates/lima.yaml.tmpl
var configTemplateSource string

// Lima YAML configuration template.
//
// Uses Virtualization.framework (vz) on macOS with virtiofs mounts. The
// provisioning script installs containerd and CNI plugins on first boot.
var configTemplate = template.Must(template.New("lima").Parse(configTemplateSource))

// Values injected into the Lima YAML template.
type configData struct {
	Arch           string // Lima architecture identifier (e.g. "aarch64", "x86_64").
	CPUs           int    // Number of virtual CPUs.
	Memory         string // Memory allocation with unit suffix (e.g. "2GiB").
	Disk           string // Disk size with unit suffix (e.g. "10GiB").
	ContainerdSock string // Host socket path for forwarding the guest containerd socket.
}

// Generates the Lima YAML configuration for the crux VM.
//
// The configuration targets the host's native architecture and uses sensible
// defaults for CPU, memory, and disk allocation.
func generateConfig() (string, error) {
	data := configData{
		Arch:           limaArch(),
		CPUs:           defaultCPUs,
		Memory:         fmt.Sprintf("%dGiB", defaultMemoryGiB),
		Disk:           fmt.Sprintf("%dGiB", defaultDiskGiB),
		ContainerdSock: containerdForwardedSocket(),
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

// Handle to the crux Lima instance.
//
// A Lima instance named "crux" is managed by shelling out to a vendored
// limactl binary. All operations parse limactl's output.
type lima struct {
	limactl string // Absolute path to the limactl binary.
}

// Creates a Lima handle, ensuring the binary is available.
//
// Limactl is downloaded on first use if it is not already present in the
// crux data directory. The VM is not started.
func newLima() (*lima, error) {
	bin, err := ensureBinary()
	if err != nil {
		return nil, err
	}
	return &lima{limactl: bin}, nil
}

// Creates and starts the VM, or starts an existing stopped VM.
//
// On first call a Lima configuration is generated, the VM instance is
// created and booted. Blocks until the VM passes its readiness probes
// (containerd socket available). Returns [ErrVMAlreadyRunning] if the VM
// is already running.
func (l *lima) start() error {
	status, err := l.status()
	if err != nil {
		return err
	}

	switch status {
	case StateRunning:
		return ErrVMAlreadyRunning

	case StateStopped:
		if err := l.run("start", "--tty=false", limaInstanceName); err != nil {
			return crex.Wrap(ErrVMStart, err)
		}
		return nil

	case StateNotCreated:
		configPath, err := generateConfig()
		if err != nil {
			return err
		}
		if err := l.run("start", "--tty=false", "--name="+limaInstanceName, configPath); err != nil {
			return crex.Wrap(ErrVMStart, err)
		}
		return nil
	}

	return nil
}

// Gracefully shuts down the VM.
//
// An ACPI shutdown signal is sent and the call blocks until the VM stops.
// Returns [ErrVMNotRunning] if the VM is not currently running.
func (l *lima) stop() error {
	status, err := l.status()
	if err != nil {
		return err
	}
	if status != StateRunning {
		return ErrVMNotRunning
	}

	if err := l.run("stop", limaInstanceName); err != nil {
		return crex.Wrap(ErrVMStop, err)
	}
	return nil
}

// Deletes the VM and its disk images.
//
// Deletion is forced without confirmation. The VM is stopped first if it
// is running. After this call the status becomes [StatusNotCreated].
func (l *lima) destroy() error {
	status, err := l.status()
	if err != nil {
		return err
	}
	if status == StateNotCreated {
		return ErrVMNotCreated
	}

	if err := l.run("delete", "--force", limaInstanceName); err != nil {
		return crex.Wrap(ErrVMStop, err)
	}
	return nil
}

// Queries the current state of the VM.
//
// Limactl is called to determine whether the VM exists and whether it is
// running or stopped.
func (l *lima) status() (State, error) {
	var stdout bytes.Buffer
	cmd := exec.Command(l.limactl, "list", "--format={{.Status}}", limaInstanceName)
	cmd.Stdout = &stdout
	cmd.Env = l.env()

	if err := cmd.Run(); err != nil {
		return StateNotCreated, nil
	}

	output := strings.TrimSpace(stdout.String())
	switch output {
	case limaStatusRunning:
		return StateRunning, nil
	case limaStatusStopped:
		return StateStopped, nil
	default:
		return StateNotCreated, nil
	}
}

// Runs a command inside the VM and captures its output.
//
// Blocks until the command completes. The command runs as the default Lima
// user inside the guest.
func (l *lima) exec(command string, args ...string) (*ExecResult, error) {
	status, err := l.status()
	if err != nil {
		return nil, err
	}
	if status != StateRunning {
		return nil, ErrVMNotRunning
	}

	shellArgs := append([]string{"shell", limaInstanceName, command}, args...)

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(l.limactl, shellArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = l.env()

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, crex.Wrap(ErrVMExec, err)
		}
	}

	return &ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

// Runs a limactl subcommand.
//
// A [*CommandError] is returned on failure.
func (l *lima) run(args ...string) error {
	cmd := exec.Command(l.limactl, args...)
	cmd.Env = l.env()
	output, err := cmd.CombinedOutput()
	if err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return &commandError{
			subcommand: args[0],
			exitCode:   exitCode,
			output:     strings.TrimSpace(string(output)),
		}
	}
	return nil
}

// Environment for limactl commands.
//
// LIMA_HOME is set to the crux VM directory so Lima stores its instance
// data alongside other crux state rather than in ~/.lima. PATH and HOME
// are preserved from the current process so that limactl can find system
// tools and resolve user directories.
func (l *lima) env() []string {
	env := []string{"LIMA_HOME=" + paths.VM()}

	appendIfSet := func(key string) {
		if val := os.Getenv(key); val != "" {
			env = append(env, key+"="+val)
		}
	}

	appendIfSet("PATH")
	appendIfSet("HOME")
	appendIfSet("USER")
	appendIfSet("TMPDIR")

	return env
}

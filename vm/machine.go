//go:build darwin

package vm

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/paths"
)

const (

	// Status string returned by limactl when the VM is running.
	limaStatusRunning = "Running"

	// Status string returned by limactl when the VM is stopped.
	limaStatusStopped = "Stopped"
)

// Handle to the crux virtual machine.
//
// Manages a Lima instance named "crux" by shelling out to a vendored
// limactl binary. All operations parse limactl's output.
type Machine struct {
	limactl string // Absolute path to the limactl binary.
}

// Returns a Machine handle, ensuring Lima is available.
//
// Downloads limactl on first use if it is not already present in the crux
// data directory. Does not start the VM.
func NewMachine() (*Machine, error) {
	bin, err := ensureLima()
	if err != nil {
		return nil, err
	}
	return &Machine{limactl: bin}, nil
}

// Creates and starts the VM if it does not already exist, or starts an
// existing stopped VM.
//
// On first call, generates a Lima configuration, creates the VM instance,
// and boots it. Blocks until the VM passes its readiness probes (containerd
// socket available). Subsequent calls on an already-running VM return
// [ErrVMAlreadyRunning].
func (m *Machine) Start() error {
	status, err := m.Status()
	if err != nil {
		return err
	}

	switch status {
	case StatusRunning:
		return ErrVMAlreadyRunning

	case StatusStopped:
		if err := m.run("start", "--tty=false", limaInstanceName); err != nil {
			return crex.Wrap(ErrVMStart, err)
		}
		return nil

	case StatusNotCreated:
		configPath, err := generateConfig()
		if err != nil {
			return err
		}
		if err := m.run("start", "--tty=false", "--name="+limaInstanceName, configPath); err != nil {
			return crex.Wrap(ErrVMStart, err)
		}
		return nil
	}

	return nil
}

// Gracefully shuts down the VM.
//
// Sends an ACPI shutdown signal and waits for the VM to stop. Returns
// [ErrVMNotRunning] if the VM is not currently running.
func (m *Machine) Stop() error {
	status, err := m.Status()
	if err != nil {
		return err
	}
	if status != StatusRunning {
		return ErrVMNotRunning
	}

	if err := m.run("stop", limaInstanceName); err != nil {
		return crex.Wrap(ErrVMStop, err)
	}
	return nil
}

// Deletes the VM and its disk images.
//
// Forces deletion without confirmation. The VM is stopped first if it is
// running. After this call, [Status] returns [StatusNotCreated].
func (m *Machine) Destroy() error {
	status, err := m.Status()
	if err != nil {
		return err
	}
	if status == StatusNotCreated {
		return ErrVMNotCreated
	}

	if err := m.run("delete", "--force", limaInstanceName); err != nil {
		return crex.Wrap(ErrVMStop, err)
	}
	return nil
}

// Returns the current state of the VM.
//
// Queries limactl to determine whether the VM exists and whether it is
// running or stopped.
func (m *Machine) Status() (Status, error) {
	var stdout bytes.Buffer
	cmd := exec.Command(m.limactl, "list", "--format={{.Status}}", limaInstanceName)
	cmd.Stdout = &stdout
	cmd.Env = m.env()

	if err := cmd.Run(); err != nil {
		// limactl list exits non-zero if instance doesn't exist
		return StatusNotCreated, nil
	}

	output := strings.TrimSpace(stdout.String())
	switch output {
	case limaStatusRunning:
		return StatusRunning, nil
	case limaStatusStopped:
		return StatusStopped, nil
	default:
		return StatusNotCreated, nil
	}
}

// Executes a command inside the VM and returns its output.
//
// Blocks until the command completes. The command runs as the default Lima
// user inside the guest.
func (m *Machine) Exec(command string, args ...string) (*ExecResult, error) {
	status, err := m.Status()
	if err != nil {
		return nil, err
	}
	if status != StatusRunning {
		return nil, ErrVMNotRunning
	}

	shellArgs := append([]string{"shell", limaInstanceName, command}, args...)

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(m.limactl, shellArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = m.env()

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

// Runs a limactl subcommand, returning a [*CommandError] on failure.
func (m *Machine) run(args ...string) error {
	cmd := exec.Command(m.limactl, args...)
	cmd.Env = m.env()
	output, err := cmd.CombinedOutput()
	if err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return &CommandError{
			Subcommand: args[0],
			ExitCode:   exitCode,
			Output:     strings.TrimSpace(string(output)),
		}
	}
	return nil
}

// Builds the environment for limactl commands.
//
// Sets LIMA_HOME to the crux VM directory so Lima stores its instance
// data alongside other crux state rather than in ~/.lima. Preserves PATH
// and HOME from the current process so that limactl can find system tools
// and resolve user directories.
func (m *Machine) env() []string {
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

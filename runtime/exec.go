package runtime

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/cruciblehq/crux/kit/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Options for executing a command inside a container.
//
// When Env is set it replaces the process environment entirely. When Workdir
// is set it overrides the process working directory. A zero-value ExecOptions
// inherits everything from the container's OCI spec.
type ExecOptions struct {
	Env     []string // Environment as KEY=VAL pairs, replaces default if set.
	Workdir string   // Working directory override.
}

// Output captured from a command executed inside a container.
type ExecResult struct {
	Stdout   string // Standard output from the command.
	Stderr   string // Standard error from the command.
	ExitCode int    // Process exit code (0 = success).
}

// Sequence counter for generating unique exec process identifiers.
var execSeq uint64

// Runs a command inside a container and captures its output.
//
// This is the core exec primitive. It opens a containerd client, loads the
// container and its task, builds an OCI process spec from the container's
// configuration, applies any option overrides, and runs the process.
func containerExec(ctx context.Context, registry, id string, opts ExecOptions, command string, args ...string) (*ExecResult, error) {
	client, err := newContainerdClient(registry)
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}
	defer client.Close()

	ctr, err := client.LoadContainer(ctx, id)
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	task, err := ctr.Task(ctx, nil)
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	pspec, err := execSpec(ctx, ctr, opts, command, args...)
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	return runExec(ctx, task, pspec)
}

// Builds an OCI process spec for an exec.
//
// The container's existing process configuration is cloned and the arguments
// are replaced with the provided command. Environment and working directory
// are then overridden from opts if set. Terminal mode is always disabled.
func execSpec(ctx context.Context, ctr containerd.Container, opts ExecOptions, command string, args ...string) (*specs.Process, error) {
	spec, err := ctr.Spec(ctx)
	if err != nil {
		return nil, err
	}

	pspec := *spec.Process
	pspec.Terminal = false
	pspec.Args = append([]string{command}, args...)

	if len(opts.Env) > 0 {
		pspec.Env = opts.Env
	}
	if opts.Workdir != "" {
		pspec.Cwd = opts.Workdir
	}

	return &pspec, nil
}

// Runs a process spec inside a task and collects its output.
//
// A uniquely identified exec process is created, started, and awaited. Stdout
// and stderr are captured in memory. The process is deleted after completion
// regardless of outcome.
func runExec(ctx context.Context, task containerd.Task, pspec *specs.Process) (*ExecResult, error) {
	execID := fmt.Sprintf("exec-%d", atomic.AddUint64(&execSeq, 1))

	var stdout, stderr bytes.Buffer
	process, err := task.Exec(ctx, execID, pspec, cio.NewCreator(
		cio.WithStreams(nil, &stdout, &stderr),
	))
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	statusC, err := process.Wait(ctx)
	if err != nil {
		process.Delete(ctx)
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	if err := process.Start(ctx); err != nil {
		process.Delete(ctx)
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	exitStatus := <-statusC
	process.Delete(ctx)

	code, _, err := exitStatus.Result()
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	return &ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: int(code),
	}, nil
}

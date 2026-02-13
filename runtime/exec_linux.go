//go:build linux

package runtime

import (
	"bytes"
	"context"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/cruciblehq/crux/kit/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Runs a command inside a container and captures its output.
//
// On Linux the containerd client and shim share the same kernel, so the
// standard Task.Exec API works directly with in-process FIFO IO.
func containerExec(ctx context.Context, client *containerd.Client, _, id string, opts ExecOptions, command string, args ...string) (*ExecResult, error) {
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
// are replaced with the provided command. Environment variables from opts are
// merged on top of the container's existing env (preserving PATH etc.).
// Working directory is overridden from opts if set. Terminal mode is always
// disabled.
func execSpec(ctx context.Context, ctr containerd.Container, opts ExecOptions, command string, args ...string) (*specs.Process, error) {
	spec, err := ctr.Spec(ctx)
	if err != nil {
		return nil, err
	}

	pspec := *spec.Process
	pspec.Terminal = false
	pspec.Args = append([]string{command}, args...)

	if len(opts.Env) > 0 {
		pspec.Env = mergeEnv(pspec.Env, opts.Env)
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
	var stdout, stderr bytes.Buffer
	process, err := task.Exec(ctx, nextExecID(), pspec, cio.NewCreator(
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

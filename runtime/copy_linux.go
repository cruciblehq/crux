//go:build linux

package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/cruciblehq/crux/kit/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Copies a tar stream into a running container.
//
// On Linux the containerd client and shim share the same kernel, so the
// standard Task.Exec API works directly with in-process FIFO IO. The tar
// stream is piped into the container process's stdin.
func containerCopy(ctx context.Context, client *containerd.Client, _, id string, r io.Reader, destDir string) error {
	ctr, err := client.LoadContainer(ctx, id)
	if err != nil {
		return crex.Wrap(ErrContainerCopy, err)
	}

	task, err := ctr.Task(ctx, nil)
	if err != nil {
		return crex.Wrap(ErrContainerCopy, err)
	}

	spec, err := ctr.Spec(ctx)
	if err != nil {
		return crex.Wrap(ErrContainerCopy, err)
	}

	pspec := *spec.Process
	pspec.Terminal = false
	pspec.Args = []string{"tar", "xf", "-", "-C", destDir}

	if err := execWithStdin(ctx, task, &pspec, r); err != nil {
		return crex.Wrap(ErrContainerCopy, err)
	}

	return nil
}

// Runs a process inside a container task, piping r into its stdin, and
// waits for it to exit. Returns an error if the process exits non-zero.
func execWithStdin(ctx context.Context, task containerd.Task, pspec *specs.Process, r io.Reader) error {
	var stderr bytes.Buffer
	process, err := task.Exec(ctx, nextExecID(), pspec, cio.NewCreator(
		cio.WithStreams(r, io.Discard, &stderr),
	))
	if err != nil {
		return err
	}

	statusC, err := process.Wait(ctx)
	if err != nil {
		process.Delete(ctx)
		return err
	}

	if err := process.Start(ctx); err != nil {
		process.Delete(ctx)
		return err
	}

	exitStatus := <-statusC
	process.Delete(ctx)

	code, _, err := exitStatus.Result()
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("process exited with code %d: %s", code, stderr.String())
	}

	return nil
}

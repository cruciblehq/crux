package cli

import (
	"context"
	"errors"
	"os"

	"github.com/cruciblehq/crux/compute"
	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/utils-go/crex"
)

// Represents the 'crux local logs' command.
type LocalLogsCmd struct {
	Follow bool   `short:"f" help:"Keep printing new output until interrupted."`
	Unit   string `arg:"" help:"Unit whose output to print."`
}

// Prints the output of a unit running in the local environment.
func (c *LocalLogsCmd) Run(ctx context.Context) error {
	const description = "cannot read unit logs"

	p, err := compute.ProviderFor(compute.Local)
	if err != nil {
		return err
	}

	m, err := p.Connect(ctx, compute.DefaultLocalMachineID())
	if err != nil {
		return localError(err, description, "Run 'crux local start' to start the local environment first.")
	}
	defer m.Close()

	err = m.Units().Logs(ctx, speccompute.UnitID(c.Unit), c.Follow, os.Stdout)
	var opErr *speccompute.Error
	if errors.As(err, &opErr) && opErr.Code == speccompute.ErrorCodeNotFound {
		return crex.UserErrorf(description, "unit %q does not exist", c.Unit).Reclassify(err)
	}
	return err
}

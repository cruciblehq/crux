package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux vm status' command.
type VmStatusCmd struct{}

// Executes the VM status command.
func (c *VmStatusCmd) Run(ctx context.Context) error {
	status, err := runtime.GetStatus()
	if err != nil {
		return err
	}

	fmt.Println(status)
	return nil
}

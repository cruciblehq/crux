package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/internal/runtime"
)

// Represents the 'crux runtime status' command.
type RuntimeStatusCmd struct{}

// Shows the current state of the container runtime environment.
func (c *RuntimeStatusCmd) Run(ctx context.Context) error {
	status, err := runtime.Status()
	if err != nil {
		return err
	}

	fmt.Println(status)
	return nil
}

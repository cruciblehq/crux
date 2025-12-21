package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/internal"
)

// Represents the 'version' command.
type VersionCmd struct{}

// Runs the version command, printing the current version of Crux.
func (c *VersionCmd) Run(ctx context.Context) error {
	fmt.Println(internal.VersionString())
	return nil
}

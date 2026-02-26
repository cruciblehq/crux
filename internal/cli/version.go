package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/internal"
)

// Represents the 'crux version' command.
type VersionCmd struct{}

// Executes the version command.
func (c *VersionCmd) Run(ctx context.Context) error {
	fmt.Println(internal.VersionString())
	return nil
}

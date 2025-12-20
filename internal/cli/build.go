package cli

import (
	"context"

	"github.com/cruciblehq/crux/pkg/build"
	"github.com/cruciblehq/crux/pkg/manifest"
)

// Represents the 'crux build' command
type BuildCmd struct {
	Watch bool `short:"w" help:"Watch for changes and rebuild automatically."`
}

// Executes the build command
func (c *BuildCmd) Run() error {

	// Load manifest options
	man, err := manifest.Read()
	if err != nil {
		return err
	}

	// Always build first
	if err := build.Build(context.Background(), *man); err != nil {
		// In watch mode, log error but continue watching
		if c.Watch {
			// Error already logged by build
		} else {
			return err
		}
	}

	// Watch mode
	// if c.Watch {
	// 	var mux sync.RWMutex
	// 	if err := watch.WatchResource(man, &mux); err != nil {
	// 		return err
	// 	}
	// 	select {} // Block forever
	// }

	return nil
}

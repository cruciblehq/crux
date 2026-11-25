package cli

import (
	"context"
	"errors"
	"strings"

	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
)

// Represents the 'crux local add' command.
type LocalAddCmd struct{}

// Registers the service in the current directory with the local blueprint.
//
// Reads the manifest from the directory set by -C (defaults to ".") and
// adds a service entry to the local blueprint that names that directory
// by its absolute path. The local ID is derived from the last component
// of the resource name (e.g. "myns/my-api" becomes "my-api").
func (c *LocalAddCmd) Run(ctx context.Context) error {
	man, err := readProjectManifest()
	if err != nil {
		return err
	}

	ref := manifest.Ref{
		ID:   resourceID(man.Resource.Name),
		Path: RootCmd.Context,
	}

	if err := modifyLocalBlueprint(ctx, func(bp *manifest.Blueprint) error {
		return bp.AddService(ref)
	}); err != nil {
		if errors.Is(err, manifest.ErrServiceExists) {
			return crex.UserError("service already added", ref.ID).
				Recovery("This service is already registered in the local blueprint.").
				Cause(err).
				Err()
		}
		return err
	}

	return nil
}

// Returns the last component of the path of a resource name (e.g., "my-api"
// from "myns/my-api").
func resourceID(name string) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

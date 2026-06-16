package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
)

// Represents the 'crux local add' command.
type LocalAddCmd struct{}

// Registers the service in the current directory with the local blueprint.
//
// Reads the manifest from the directory set by -C (defaults to ".") and adds a
// service entry to the local blueprint. The local ID is derived from the last
// component of the resource name (e.g. "myns/my-api" → "my-api").
func (c *LocalAddCmd) Run(ctx context.Context) error {
	man, err := manifest.ReadAt(RootCmd.Context)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return crex.UserError("no manifest found", "no crucible.yaml was found in the current directory").
				Recovery("Run this command from a directory containing a crucible.yaml.").
				Cause(err).
				Err()
		}
		return crex.UserError("invalid manifest", "the crucible.yaml could not be read or parsed").
			Recovery("Check that crucible.yaml is present and valid.").
			Cause(err).
			Err()
	}

	name := man.Resource.Name
	ref := manifest.Ref{
		ID:  resourceID(name),
		Ref: name,
	}
	if man.Resource.Version != "" {
		ref.Ref = fmt.Sprintf("%s %s", name, man.Resource.Version)
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

// Returns the last path component of a resource name.
//
// For example, "myns/my-api" → "my-api".
func resourceID(name string) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

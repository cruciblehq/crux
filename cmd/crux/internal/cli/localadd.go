package cli

import (
	"context"
	"fmt"
	"strings"

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
		return err
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

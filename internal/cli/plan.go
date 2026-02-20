package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/blueprint"
	"github.com/cruciblehq/crux/internal/config"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/paths"
	spec "github.com/cruciblehq/spec/blueprint"
)

const (

	// Subdirectory under dist for plan outputs.
	planOutputSubdir = "plans"

	// Timestamp format for plan filenames.
	timestampFormat = "20060102-150405"
)

// Represents the 'crux plan' command.
type PlanCmd struct {
	Blueprint string `arg:"" help:"Path to blueprint file"`
	State     string `optional:"" help:"Path to existing state file for incremental planning"`
	Registry  string `help:"Registry URL for resolving references (default: http://hub.cruciblehq.xyz:8080)."`
	Provider  string `help:"Provider profile name (empty = default)"`
}

// Executes the plan command.
func (c *PlanCmd) Run(ctx context.Context) error {
	registryURL := c.Registry
	if registryURL == "" {
		registryURL = internal.DefaultRegistryURL
	}

	// Load provider configuration
	provider, err := config.GetOrDefaultProvider(c.Provider)
	if err != nil {
		return err
	}

	slog.Info("generating deployment plan...", "blueprint", c.Blueprint, "state", c.State)

	bpData, err := os.ReadFile(c.Blueprint)
	if err != nil {
		return err
	}

	bp, err := spec.Decode(bpData)
	if err != nil {
		return err
	}

	p, err := blueprint.Execute(ctx, bp, blueprint.ExecuteOptions{
		State:            c.State,
		Registry:         registryURL,
		Provider:         provider.Type,
		DefaultNamespace: internal.DefaultNamespace,
	})
	if err != nil {
		return err
	}

	// Determine output path
	output, err := determinePlanOutputPath(c.Blueprint)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(output, data, paths.DefaultFileMode); err != nil {
		return err
	}

	slog.Info("deployment plan generated successfully", "output", output)

	return nil
}

// Determines the output path for the plan file.
func determinePlanOutputPath(blueprintPath string) (string, error) {
	timestamp := time.Now().Format(timestampFormat)
	dir := filepath.Dir(blueprintPath)
	plansDir := filepath.Join(paths.DistDir(dir), planOutputSubdir)
	if err := os.MkdirAll(plansDir, paths.DefaultDirMode); err != nil {
		return "", crex.Wrap(ErrFileSystem, err)
	}
	return filepath.Join(plansDir, fmt.Sprintf("plan-%s.json", timestamp)), nil
}

package build

import (
	"maps"

	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/runtime"
)

// Default shell used for run steps when no shell modifier has been set.
const defaultShell = "/bin/sh"

// Tracks accumulated modifiers during step execution.
//
// State flows linearly through the step list. Standalone modifiers update
// the state permanently via [recipeState.apply]. Operations read the
// effective values for a single step via [recipeState.resolve] without
// modifying the persistent state.
type recipeState struct {
	shell   string
	env     map[string]string
	workdir string
}

// Creates a new recipeState with default values.
func newRecipeState() *recipeState {
	return &recipeState{
		shell: defaultShell,
		env:   make(map[string]string),
	}
}

// Persists modifier fields from a step into the state.
//
// Called for standalone modifier steps and platform groups. The state is
// mutated permanently, affecting all subsequent steps.
func (s *recipeState) apply(step manifest.Step) {
	if step.Shell != "" {
		s.shell = step.Shell
	}
	if step.Workdir != "" {
		s.workdir = step.Workdir
	}
	for k, v := range step.Env {
		s.env[k] = v
	}
}

// Computes the effective shell, environment, and working directory for a
// single operation without modifying the persistent state.
//
// Step-level modifiers override the corresponding state values for this
// operation only. The returned ExecOptions and shell are ready to use.
func (s *recipeState) resolve(step manifest.Step) (string, runtime.ExecOptions) {
	shell := s.shell
	if step.Shell != "" {
		shell = step.Shell
	}

	workdir := s.workdir
	if step.Workdir != "" {
		workdir = step.Workdir
	}

	return shell, runtime.ExecOptions{
		Env:     s.environWith(step.Env),
		Workdir: workdir,
	}
}

// Formats the persistent environment merged with step-scoped overrides.
//
// Entries in extra take precedence over entries already in the state. Neither
// the persistent state nor extra are modified.
func (s *recipeState) environWith(extra map[string]string) []string {
	merged := make(map[string]string, len(s.env)+len(extra))
	maps.Copy(merged, s.env)
	maps.Copy(merged, extra)
	env := make([]string, 0, len(merged))
	for k, v := range merged {
		env = append(env, k+"="+v)
	}
	return env
}

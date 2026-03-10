package compute

import "github.com/cruciblehq/crux/internal/compute/internal/provider"

const (
	StateNotProvisioned = provider.StateNotProvisioned // Instance has not been provisioned.
	StateRunning        = provider.StateRunning        // Instance is running and reachable.
	StateStopped        = provider.StateStopped        // Instance exists but is not running.
)

// Interface for compute backend implementations.
type Backend = provider.Backend

// Interface for communicating with a cruxd instance.
type Client = provider.Client

// Lifecycle state of a cruxd instance.
type State = provider.State

// Output captured from a command executed on the instance's host.
type ExecResult = provider.ExecResult

// Returns a new [ExecResult].
var NewExecResult = provider.NewExecResult

package compute

import "github.com/cruciblehq/crux/internal/compute/internal/provider"

// Interface for compute backend implementations.
type Backend = provider.Backend

// Interface for communicating with a cruxd instance.
type Client = provider.Client

// Config holds parameters for provisioning a cruxd instance.
type Config = provider.Config

// Lifecycle state of a cruxd instance.
type State = provider.State

const (
	StateNotProvisioned = provider.StateNotProvisioned // Instance has not been provisioned.
	StateRunning        = provider.StateRunning        // Instance is running and reachable.
	StateStopped        = provider.StateStopped        // Instance exists but is not running.
)

// Output captured from a command executed on the instance's host.
type ExecResult = provider.ExecResult

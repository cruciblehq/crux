// Package compute provisions and manages infrastructure for Crucible.
//
// The package maintains a registry of provider backends, initialised lazily.
// Callers select a backend via [BackendFor] with a [Provider] constant, then
// access the resource manager they need through [Backend.Compute],
// [Backend.Storage], or [Backend.Network].
//
//	b, err := compute.BackendFor(compute.Local)
//	c := b.Compute()
//
// Provider implementations live in sub-packages (e.g. compute/local) and are
// adapted to the compute model by shims that live in this package.
//
// Compute methods are synchronous: they block until the underlying instance
// reaches the expected target state. If it does not converge, the provider
// reverts any partial changes and returns an error. Context cancellation is
// the mechanism for aborting a long-running call.
//
// Provisioning returns a provider-assigned identifier for the new instance.
// All subsequent operations use that identifier. If provisioning fails, the
// provider tears down any partial state automatically.
//
//	id, err := c.Provision(ctx, imageID, opts)
//
//	err = c.Stop(ctx, id)
//	err = c.Start(ctx, id)
//	err = c.Deprovision(ctx, id)
//
// On macOS the local backend provisions a lightweight VM on first use that
// runs containerd. On Linux containerd runs natively as a system service.
//
// [Compute.Status] returns the current [State] of an instance:
// [StateNotProvisioned], [StateStopped], or [StateRunning].
package compute

// Package compute provisions and manages compute hosts for Crucible.
//
// The package maintains a registry of compute backends, initialised lazily.
// Callers select a backend via [BackendFor] with a [Provider] constant, then
// interact with it through the [Backend] interface.
//
//	b, err := compute.BackendFor(compute.Local)
//
// Provider implementations live in sub-packages (e.g. compute/local) and are
// adapted to the compute model by shims that live in this package.
//
// Backend methods are synchronous: they block until the underlying host reaches
// the expected target state. If it does not converge, the provider reverts any
// partial changes and returns an error. Context cancellation is the mechanism
// for aborting a long-running call.
//
// Provisioning assigns a name to the new host. If provisioning fails, the
// provider tears down any partial state automatically.
//
//	err = b.Provision(ctx, imageID, "local", nil)
//
//	err = b.Stop(ctx, "local")
//	err = b.Start(ctx, "local")
//	err = b.Deprovision(ctx, "local")
//
// On macOS the local backend provisions a lightweight VM on first use that
// runs containerd. On Linux containerd runs natively as a system service.
//
// [Backend.Status] returns the current [State] of a named host:
// [StateNotProvisioned], [StateStopped], or [StateRunning].
package compute

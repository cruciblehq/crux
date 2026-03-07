// Package compute provisions cruxd runtime instances for Crucible.
//
// The package maintains a registry of compute backends, initialised lazily.
// Callers select a backend via [BackendFor] with a [Provider] constant,
// then interact with it through the [Backend] interface.
//
//	b, err := compute.BackendFor(compute.Local)
//
// Provider implementations live in sub-packages (e.g. compute/internal/local)
// and implement the [provider.Backend] interface defined in compute/internal/provider.
//
// Lifecycle methods are synchronous: they block until the underlying runtime
// reaches the expected target state. If it does not converge, the provider
// reverts any partial changes and returns an error. Context cancellation is
// the mechanism for aborting a long-running call.
//
// Provisioning creates and starts the instance. If it fails, the provider
// tears down any partial state automatically.
//
//	cfg := &compute.Config{Name: "local", Version: "0.1.3"}
//	err := b.Provision(ctx, cfg)
//
//	err = b.Stop(ctx, "local")
//	err = b.Start(ctx, "local")
//	err = b.Deprovision(ctx, "local")
//
// On macOS the local backend provisions a lightweight VM on first use and runs
// cruxd inside it. On Linux no VM is needed because cruxd runs natively.
//
// The backend's Status method returns the current [State]:
// [StateNotProvisioned], [StateStopped], or [StateRunning].
//
//	state, err := b.Status(ctx, "local")
//	if state == compute.StateRunning {
//	    // instance is reachable
//	}
package compute

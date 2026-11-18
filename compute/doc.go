// Package compute provisions machines and talks to cruxd on them.
//
// A [Provider] creates and runs machines on one kind of infrastructure.
// Callers look one up by name with [ProviderFor]:
//
//	p, err := compute.ProviderFor(compute.Local)
//
// Provider methods block until the machine reaches the expected state. If it
// does not, the provider reverts partial changes and returns an error.
// Cancelling the context aborts a long call.
//
// Provision returns the identifier that every later call uses:
//
//	id, err := p.Provision(ctx, imageID, opts)
//	err = p.Stop(ctx, id)
//	err = p.Start(ctx, id)
//	err = p.Deprovision(ctx, id)
//
// [Provider.Status] returns the [State] of a machine: [StateNotProvisioned],
// [StateStopped], or [StateRunning]. [Provider.Connect] opens a session with
// cruxd on a running machine.
//
// The local provider runs one Lima VM through the [lima] package.
package compute

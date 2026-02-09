package cli

// Manages the container runtime environment.
type RuntimeCmd struct {
	Start   *RuntimeStartCmd   `cmd:"" help:"Provision and start the runtime."`
	Stop    *RuntimeStopCmd    `cmd:"" help:"Stop the runtime."`
	Status  *RuntimeStatusCmd  `cmd:"" help:"Show runtime status."`
	Destroy *RuntimeDestroyCmd `cmd:"" help:"Destroy the runtime and all its data."`
	Exec    *RuntimeExecCmd    `cmd:"" help:"Run a command inside the runtime."`
}

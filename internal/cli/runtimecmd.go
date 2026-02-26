package cli

// Manages the container runtime environment.
type RuntimeCmd struct {
	Start   *RuntimeStartCmd   `cmd:"" help:"Provision and start the runtime."`
	Stop    *RuntimeStopCmd    `cmd:"" help:"Stop the runtime."`
	Restart *RuntimeRestartCmd `cmd:"" help:"Stop and restart the runtime."`
	Reset   *RuntimeResetCmd   `cmd:"" help:"Destroy and recreate the runtime from scratch."`
	Destroy *RuntimeDestroyCmd `cmd:"" help:"Destroy the runtime and all its data."`
	Status  *RuntimeStatusCmd  `cmd:"" help:"Show runtime status."`
	Exec    *RuntimeExecCmd    `cmd:"" help:"Run a command inside the runtime."`
}

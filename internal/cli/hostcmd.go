package cli

// Manages the Crucible host environment.
type HostCmd struct {
	Start   *HostStartCmd   `cmd:"" help:"Provision and start the host."`
	Stop    *HostStopCmd    `cmd:"" help:"Stop the host."`
	Restart *HostRestartCmd `cmd:"" help:"Stop and restart the host."`
	Reset   *HostResetCmd   `cmd:"" help:"Destroy and recreate the host from scratch."`
	Destroy *HostDestroyCmd `cmd:"" help:"Destroy the host and all its data."`
	Status  *HostStatusCmd  `cmd:"" help:"Show host status."`
	Exec    *HostExecCmd    `cmd:"" help:"Run a command inside the host."`
}

package cli

// Represents the 'crux server' command group.
type ServerCmd struct {
	Start  ServerStartCmd  `cmd:"" help:"Start the development server."`
	Stop   ServerStopCmd   `cmd:"" help:"Stop the development server."`
	Status ServerStatusCmd `cmd:"" help:"Show server status."`
	Logs   ServerLogsCmd   `cmd:"" help:"Show server logs."`
}

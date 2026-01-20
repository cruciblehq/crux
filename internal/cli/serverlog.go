package cli

// Represents the 'crux server logs' command.
type ServerLogsCmd struct {
	Watch bool `short:"w" help:"Watch for new logs."`
}

func (c *ServerLogsCmd) Run() error {
	// TODO: Implement server logs
	return nil
}

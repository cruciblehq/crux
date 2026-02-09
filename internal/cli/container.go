package cli

// Manages containers in the container runtime.
type ContainerCmd struct {
	Stop    *ContainerStopCmd    `cmd:"" help:"Stop a running container."`
	Destroy *ContainerDestroyCmd `cmd:"" help:"Remove a container and its snapshot."`
	Status  *ContainerStatusCmd  `cmd:"" help:"Show the state of a container."`
	Exec    *ContainerExecCmd    `cmd:"" help:"Execute a command inside a container."`
	Update  *ContainerUpdateCmd  `cmd:"" help:"Re-import an image and restart a container."`
}

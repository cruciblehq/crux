package runtime

// ContainerState represents the lifecycle state of a container.
type ContainerState string

const (
	ContainerRunning    ContainerState = "running"     // Task is active.
	ContainerStopped    ContainerState = "stopped"     // Container exists but has no running task.
	ContainerNotCreated ContainerState = "not created" // Container does not exist.
)

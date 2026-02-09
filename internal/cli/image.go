package cli

// Manages OCI images in the container runtime.
type ImageCmd struct {
	Import  *ImageImportCmd  `cmd:"" help:"Import an image tarball into the runtime."`
	Destroy *ImageDestroyCmd `cmd:"" help:"Remove an image and all its containers."`
	Start   *ImageStartCmd   `cmd:"" help:"Start a container from an image."`
}

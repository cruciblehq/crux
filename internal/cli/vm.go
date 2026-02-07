package cli

// Manages the local development VM.
type VmCmd struct {
	Start   *VmStartCmd   `cmd:"" help:"Create and start the VM."`
	Stop    *VmStopCmd    `cmd:"" help:"Stop the VM."`
	Status  *VmStatusCmd  `cmd:"" help:"Show VM status."`
	Destroy *VmDestroyCmd `cmd:"" help:"Delete the VM and its disk images."`
}

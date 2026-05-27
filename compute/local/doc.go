// Package local implements the local compute backend.
//
// On macOS and Linux, the local provider manages a Lima virtual machine that
// runs containerd. On macOS the VM uses Apple's Virtualization framework
// (vmType: vz); on Linux it uses QEMU (vmType: qemu). On unsupported platforms
// every method returns [ErrUnsupportedPlatform].
//
// [NewBackend] returns a [provider.Backend] whose lifecycle methods are
// synchronous: they block until the operation completes. Lima is invoked
// inline via limactl; the VM is always named after [limaInstanceName].
//
// A typical sequence provisions an instance, interacts with it, and tears
// it down when done:
//
//	b := local.NewBackend()
//	err := b.Provision(ctx, "my-instance", imageID, nil)
//
//	state, _ := b.Status(ctx, "my-instance") // provider.StateRunning
//
//	exitCode, err := b.Exec(ctx, "my-instance", os.Stdout, os.Stderr, "uname", "-a")
//
//	b.Stop(ctx, "my-instance")
//	b.Deprovision(ctx, "my-instance")
//
// [provider.Backend.Start] resumes a previously provisioned instance
// without re-downloading dependencies:
//
//	b.Start(ctx, "my-instance")
//
// Note: Lima (limactl) is used for VM lifecycle management, but it requires a
// ~30 MB download on first use and the use of SSH with ControlMaster for host-
// guest communication, which introduces timing-sensitive connection setup and
// requires the Lima user to be in the "staff" group on macOS. In the future,
// crux should consider implementing a custom VM management layer for the local
// provider, possibly using Apple's Virtualization.framework directly via its Go
// bindings (github.com/Code-Hex/vz) on macOS and QEMU/KVM on Linux, using vsock
// for host-to-guest communication instead of SSH. vsock crosses the VM boundary
// at the kernel level with no daemon, no ControlMaster timing constraints, and
// no group membership issues. The containerd socket would be proxied over vsock,
// and the image would ship pre-configured so no runtime provisioning scripts
// would be needed, avoiding several race conditions. Furthermore, this is well
// aligned with the planned "machine" resource type, which would define VM images
// declaratively and make the hypervisor layer an implementation detail of each
// compute provider. It would also enable deriving the exact VM configuration
// from affordances, which currently isn't possible.
package local

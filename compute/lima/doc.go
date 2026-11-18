// Package lima runs the local machine as a Lima virtual machine.
//
// On macOS the VM uses Apple's Virtualization framework (vmType: vz). On
// Linux it uses QEMU (vmType: qemu). On other platforms every operation
// returns [ErrUnsupportedPlatform]. There is one VM, named after
// [limaInstanceName], and every operation blocks until it completes.
//
//	err := lima.Provision(ctx, imagePath, enrollment)
//	state, err := lima.Status(ctx)
//	code, err := lima.Exec(ctx, os.Stdout, os.Stderr, "uname", "-a")
//	err = lima.Stop(ctx)
//	err = lima.Start(ctx)
//	err = lima.Deprovision(ctx)
//
// imagePath is a qcow2 disk image on this computer. The compute package pulls
// it from the Crucible registry before provisioning.
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
package lima

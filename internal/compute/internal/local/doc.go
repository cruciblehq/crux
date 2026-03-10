// Package local implements the local compute backend.
//
// On macOS, the local provider manages a Lima virtual machine that runs cruxd.
// On Linux, it manages a native cruxd process. On unsupported platforms every
// method returns [ErrUnsupportedPlatform].
//
// [NewBackend] returns a [provider.Backend] whose lifecycle methods are
// synchronous: they block until the underlying process exits. On macOS
// this is achieved by running limactl inline; on Linux, process exit is
// detected via pidfd.
//
// Image and container operations (import, start, exec, etc.) are handled
// by cruxd; this package only manages the underlying VM or cruxd process.
//
// A typical sequence provisions an instance, communicates with it through
// a client, and tears it down when done:
//
//	b := local.NewBackend()
//	err := b.Provision(ctx, "my-instance", source)
//
//	state, _ := b.Status(ctx, "my-instance") // provider.StateRunning
//
//	client, _ := b.Client(ctx, "my-instance")
//	result, _ := client.Build(ctx, req)
//
//	b.Stop(ctx, "my-instance")
//	b.Deprovision(ctx, "my-instance")
//
// [provider.Backend.Start] resumes a previously provisioned instance
// without re-downloading dependencies:
//
//	b.Start(ctx, "my-instance")
//
// [provider.Backend.Exec] runs a command inside the host environment
// and returns its output:
//
//	result, err := b.Exec(ctx, "my-instance", "uname", "-a")
//	fmt.Println(result.Stdout)
package local

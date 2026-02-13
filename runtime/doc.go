// Package runtime manages the container runtime environment for Crucible.
//
// The package exposes a platform-agnostic API over containerd. On macOS,
// containers run inside a Lima virtual machine whose containerd socket is
// forwarded to the host. On Linux, containers run natively via a vendored
// containerd. Platform-specific details are handled internally; calling any
// function on an unsupported platform returns [ErrUnsupportedPlatform].
//
// There are two levels of abstraction. The package-level functions ([Start],
// [Stop], [Status], [Destroy], [Exec]) manage the runtime itself (the VM
// on macOS, and the containerd process on Linux). [Image] and [Container]
// manage individual OCI images and their containers within that runtime.
//
// Container exec uses containerd's Task.Exec API with FIFO-based IO. On
// Linux the client and shim share the same kernel, so FIFOs work directly.
// On macOS the shim runs inside the Lima VM while the Go client runs on
// the host; FIFOs do not work across the VM boundary because pipe buffers
// are kernel-local objects. To work around this, macOS exec routes through
// limactl to invoke ctr inside the guest, where the shim, FIFOs, and
// client all share the same kernel. See [containerExec] for details.
//
// Starting and stopping the runtime:
//
//	if err := runtime.Start(); err != nil {
//		log.Fatal(err)
//	}
//	defer runtime.Stop()
//
// Querying the runtime status:
//
//	status, err := runtime.Status()
//	fmt.Println(status) // "running", "stopped", or "not created"
//
// Importing an OCI image and starting a container:
//
//	img := runtime.NewImage(id, version)
//	if err := img.Import(ctx, "build/image.tar"); err != nil {
//		log.Fatal(err)
//	}
//	c, err := img.Start(ctx, "my-service")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer c.Stop(ctx)
//
// Executing a command inside a container:
//
//	result, err := c.Exec(ctx, "uname", "-a")
//	fmt.Println(result.Stdout)
//
// Updating a running container with a new image:
//
//	if err := img.Update(ctx, ctr, "build/image.tar"); err != nil {
//		log.Fatal(err)
//	}
//
// Destroying an image and all its containers:
//
//	if err := img.Destroy(ctx); err != nil {
//		log.Fatal(err)
//	}
package runtime

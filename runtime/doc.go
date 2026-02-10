// Package runtime manages the container runtime environment for Crucible.
//
// On macOS, containers run inside a Lima virtual machine with containerd.
// The containerd socket is forwarded from the guest to the host via Lima's
// portForwards, allowing the Go client to connect directly. On Linux,
// containers run natively via a vendored containerd. The package exposes a
// platform-agnostic API; platform-specific details are handled internally.
// Calling any function on an unsupported platform returns [ErrUnsupportedPlatform].
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
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(status) // "running", "stopped", or "not created"
//
// Executing a command inside the runtime:
//
//	result, err := runtime.Exec("uname", "-a")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.Stdout)
//
// Tearing down the runtime and its resources:
//
//	if err := runtime.Destroy(); err != nil {
//		log.Fatal(err)
//	}
//
// Importing an OCI image and managing containers:
//
//	opts, err := reference.NewIdentifierOptions(registryURL, namespace)
//	if err != nil {
//		log.Fatal(err)
//	}
//	id, err := reference.ParseIdentifier(ref, resource.TypeService, opts)
//	if err != nil {
//		log.Fatal(err)
//	}
//	img := runtime.NewImage(id, version)
//	if err := img.Import(ctx, "build/image.tar"); err != nil {
//		log.Fatal(err)
//	}
//
//	c, err := img.Start(ctx, "my-service")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer c.Stop(ctx)
//
// Constructing a handle for an existing container:
//
//	ctr := runtime.NewContainer(id.Registry(), "my-service")
//	status, err := ctr.Status(ctx)
//
// Updating a container with a new image:
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

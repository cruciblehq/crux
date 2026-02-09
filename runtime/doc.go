// Package runtime manages the container runtime environment for Crucible.
//
// On macOS, containers run inside a Lima virtual machine with containerd.
// On Linux, containers run natively via a vendored containerd. The package
// exposes a platform-agnostic API; platform-specific details are handled
// internally. Calling any function on an unsupported platform returns
// [ErrUnsupportedPlatform].
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
//	status, err := runtime.GetStatus()
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
// Importing an OCI image into the runtime:
//
//	err := runtime.ImportImage(m.Resource.Ref, m.Resource.Type, m.Resource.Version, "build/image.tar")
//	if err != nil {
//		log.Fatal(err)
//	}
package runtime

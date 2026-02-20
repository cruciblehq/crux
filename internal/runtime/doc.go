// Package runtime manages the container runtime environment lifecycle.
//
// On macOS, the runtime is a Lima virtual machine that hosts containerd.
// The package-level functions ([Start], [Stop], [Status], [Destroy], [Exec])
// manage the VM lifecycle. On unsupported platforms every function returns
// [ErrUnsupportedPlatform].
//
// Image and container operations (import, start, exec, etc.) are handled
// by the cruxd daemon; this package only manages the underlying VM.
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
package runtime

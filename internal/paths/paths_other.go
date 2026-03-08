//go:build !darwin && !linux

package paths

// Path to the cruxd Unix socket for an instance.
//
// Not supported on this platform.
func CruxdSocket(_ string) string {
	panic("cruxd socket path is not available on this platform")
}

// Path to the cruxd PID file for an instance.
//
// Not supported on this platform.
func CruxdPIDFile(_ string) string {
	panic("cruxd PID file path is not available on this platform")
}

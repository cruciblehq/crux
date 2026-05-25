//go:build darwin

package compute

import (
	"encoding/binary"
	"runtime"

	"golang.org/x/sys/unix"
)

// Returns the number of logical CPUs and total physical memory in bytes on the host machine.
func HostResources() (cpus int, memoryBytes uint64) {
	cpus = runtime.NumCPU()
	raw, err := unix.SysctlRaw("hw.memsize")
	if err == nil && len(raw) == 8 {
		memoryBytes = binary.LittleEndian.Uint64(raw)
	}
	return
}

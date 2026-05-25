//go:build linux

package compute

import (
	"runtime"
	"syscall"
)

// Returns the number of logical CPUs and total physical memory in bytes on the host machine.
func HostResources() (cpus int, memoryBytes uint64) {
	cpus = runtime.NumCPU()
	var si syscall.Sysinfo_t
	if err := syscall.Sysinfo(&si); err == nil {
		memoryBytes = si.Totalram * uint64(si.Unit)
	}
	return
}

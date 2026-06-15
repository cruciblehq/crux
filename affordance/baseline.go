package affordance

import (
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Constructs the deny-all OCI runtime spec baseline.
//
// Every field that the model can grant against is initialised to its most
// restrictive value so that subsystems are only ever loosening the spec.
// Fields that are the executor's responsibility (Root.Path, Linux.CgroupsPath,
// container identity) remain at their zero value.
func newOCIBaseline() *specs.Spec {
	umask := uint32(0o777)
	memZero := int64(0)
	cpuQuotaMin := int64(1000)
	cpuPeriod := uint64(100000)
	pidUnlimited := int64(-1)
	blkWeight := uint16(10)

	return &specs.Spec{
		Version: specs.Version,
		Root:    &specs.Root{},
		Process: &specs.Process{
			User: specs.User{
				UID:   65534,
				GID:   65534,
				Umask: &umask,
			},
			Cwd:             "/",
			NoNewPrivileges: true,
			Rlimits:         zeroRlimits(),
			Capabilities:    &specs.LinuxCapabilities{},
		},
		Linux: &specs.Linux{
			Namespaces: []specs.LinuxNamespace{
				{Type: specs.PIDNamespace},
				{Type: specs.IPCNamespace},
				{Type: specs.UTSNamespace},
				{Type: specs.MountNamespace},
				{Type: specs.NetworkNamespace},
				{Type: specs.CgroupNamespace},
			},
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{
					{Allow: false, Access: "rwm"},
				},
				Memory: &specs.LinuxMemory{
					Limit: &memZero,
					Swap:  &memZero,
				},
				CPU: &specs.LinuxCPU{
					Quota:  &cpuQuotaMin,
					Period: &cpuPeriod,
				},
				Pids: &specs.LinuxPids{
					Limit: pidUnlimited,
				},
				BlockIO: &specs.LinuxBlockIO{
					Weight: &blkWeight,
				},
			},
			Seccomp: denyAllSeccomp(),
			MaskedPaths: []string{
				"/proc/acpi",
				"/proc/asound",
				"/proc/kcore",
				"/proc/keys",
				"/proc/latency_stats",
				"/proc/timer_list",
				"/proc/timer_stats",
				"/proc/sched_debug",
				"/proc/scsi",
				"/sys/firmware",
			},
			ReadonlyPaths: []string{
				"/proc/bus",
				"/proc/fs",
				"/proc/irq",
				"/proc/sys",
				"/proc/sysrq-trigger",
			},
		},
		Mounts: []specs.Mount{},
	}
}

// Canonical RLIMIT_* types in the order the baseline emits them.
//
// The set mirrors the kernel's standard limits. Each entry is initialised
// with soft=hard=0 so that any unspecified limit denies all use of that
// resource.
var rlimitBaselineTypes = []string{
	"RLIMIT_AS",
	"RLIMIT_CORE",
	"RLIMIT_CPU",
	"RLIMIT_DATA",
	"RLIMIT_FSIZE",
	"RLIMIT_LOCKS",
	"RLIMIT_MEMLOCK",
	"RLIMIT_MSGQUEUE",
	"RLIMIT_NICE",
	"RLIMIT_NOFILE",
	"RLIMIT_NPROC",
	"RLIMIT_RSS",
	"RLIMIT_RTPRIO",
	"RLIMIT_RTTIME",
	"RLIMIT_SIGPENDING",
	"RLIMIT_STACK",
}

// Returns every POSIX rlimit set to zero (soft=hard=0).
func zeroRlimits() []specs.POSIXRlimit {
	out := make([]specs.POSIXRlimit, len(rlimitBaselineTypes))
	for i, t := range rlimitBaselineTypes {
		out[i] = specs.POSIXRlimit{Type: t, Soft: 0, Hard: 0}
	}
	return out
}

// Returns the deny-all seccomp baseline.
//
// Default action is errno; only exit_group is allowed so processes can
// terminate cleanly. Subsystem grants append further allow entries.
func denyAllSeccomp() *specs.LinuxSeccomp {
	return &specs.LinuxSeccomp{
		DefaultAction: specs.ActErrno,
		Syscalls: []specs.LinuxSyscall{
			{
				Names:  []string{"exit_group"},
				Action: specs.ActAllow,
			},
		},
	}
}

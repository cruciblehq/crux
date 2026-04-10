package runtime

import (
	"context"
	"path/filepath"

	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Returns the OCI spec options that build the absolute-zero security baseline.
//
// The baseline constructs the entire OCI spec from scratch — no containerd
// defaults are used. The resulting container is maximally isolated with zero
// granted resources: no mounts, no capabilities, no environment, no cgroup
// limits, no rlimits, deny-all seccomp (only exit_group permitted), six
// namespaces (pid, ipc, uts, mount, network, cgroup), device deny-all,
// read-only rootfs, nobody user, and no-new-privileges.
//
// hugepageSizes lists the OCI pagesize strings (e.g., "2MB", "1GB") for the
// selected kernel's base page configuration. Each size gets a zero-limit entry
// in the cgroup hugepage controller. The caller derives these from the kernel
// selection — they are a compile-time constant of the kernel, not a runtime
// discovery.
//
// This container cannot run a meaningful workload. Affordances compose on
// top to grant the specific capabilities a workload requires: syscalls,
// memory, CPU, PIDs, mounts, environment variables, rlimits, and so on.
func constrictionOpts(hugepageSizes []string) []oci.SpecOpts {
	return []oci.SpecOpts{withBaselineSpec(hugepageSizes)}
}

// Constructs the absolute-zero OCI spec from scratch.
//
// Every field is explicitly set. Nothing is inherited from containerd's
// [oci.WithDefaultSpecForPlatform] or any other default population function.
// The spec is the fixed security invariant for every Crucible container:
//
//   - Rootfs: read-only, path set by containerd snapshot.
//   - Process: nobody user (65534:65534), cwd /, no-new-privileges,
//     no capabilities (all five sets nil), no environment, zero rlimits
//     (all resource limits set to 0), umask 0777 (deny-all).
//   - Namespaces: pid, ipc, uts, mount, network, cgroup, time.
//   - Seccomp: deny all syscalls except exit_group.
//   - Devices: deny all (rwm).
//   - Cgroup resources: zero memory, zero swap, minimum CPU (1ms/100ms),
//     zero PIDs, minimum block I/O weight, zero hugepages (derived from
//     kernel selection). Defense in depth — even if seccomp is bypassed,
//     no resources available.
//   - Mounts: none — no /proc, /dev, /sys, /run, /dev/shm, /dev/pts.
//   - Masked/readonly paths: none — irrelevant without mounts.
//
// User namespace is excluded because it requires UID/GID mapping
// infrastructure. Can be added via affordances when the infrastructure exists.
func withBaselineSpec(hugepageSizes []string) oci.SpecOpts {
	return func(ctx context.Context, _ oci.Client, c *containers.Container, s *specs.Spec) error {
		ns, err := namespaces.NamespaceRequired(ctx)
		if err != nil {
			return err
		}

		umask := uint32(0777)
		memZero := int64(0)
		cpuQuotaMin := int64(1000)
		cpuPeriod := uint64(100000)
		pidZero := int64(0)
		blkWeight := uint16(10)

		*s = specs.Spec{
			Version: specs.Version,
			Root: &specs.Root{
				Path:     "rootfs",
				Readonly: true,
			},
			Process: &specs.Process{
				User: specs.User{
					UID:   65534,
					GID:   65534,
					Umask: &umask,
				},
				Cwd:             "/",
				NoNewPrivileges: true,
				Rlimits:         zeroRlimits(),
			},
			Linux: &specs.Linux{
				CgroupsPath: filepath.Join("/", ns, c.ID),
				Namespaces: []specs.LinuxNamespace{
					{Type: specs.PIDNamespace},
					{Type: specs.IPCNamespace},
					{Type: specs.UTSNamespace},
					{Type: specs.MountNamespace},
					{Type: specs.NetworkNamespace},
					{Type: specs.CgroupNamespace},
					{Type: specs.TimeNamespace},
				},
				Seccomp: exitOnlySeccomp(),
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
						Limit: &pidZero,
					},
					BlockIO: &specs.LinuxBlockIO{
						Weight: &blkWeight,
					},
					HugepageLimits: zeroHugepageLimits(hugepageSizes),
				},
			},
			Mounts: []specs.Mount{},
		}
		return nil
	}
}

// Converts hugepage size strings into zero-limit entries for the cgroup
// hugepage controller. Each size (e.g., "2MB", "1GB") gets Limit: 0.
func zeroHugepageLimits(sizes []string) []specs.LinuxHugepageLimit {
	if len(sizes) == 0 {
		return nil
	}
	limits := make([]specs.LinuxHugepageLimit, len(sizes))
	for i, s := range sizes {
		limits[i] = specs.LinuxHugepageLimit{Pagesize: s, Limit: 0}
	}
	return limits
}

// Returns all POSIX resource limits set to zero (soft=0, hard=0).
func zeroRlimits() []specs.POSIXRlimit {
	types := []string{
		"RLIMIT_AS",
		"RLIMIT_CORE",
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
	limits := make([]specs.POSIXRlimit, len(types))
	for i, t := range types {
		limits[i] = specs.POSIXRlimit{Type: t, Soft: 0, Hard: 0}
	}
	return limits
}

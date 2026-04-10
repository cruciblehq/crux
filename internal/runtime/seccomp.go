package runtime

import (
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Returns a seccomp profile that denies all syscalls except exit_group.
//
// This is the absolute-zero seccomp baseline. The container process can
// terminate cleanly but cannot perform any other operation. Affordances
// layer additional syscall groups on top of this profile.
func exitOnlySeccomp() *specs.LinuxSeccomp {
	return &specs.LinuxSeccomp{
		DefaultAction: specs.ActErrno,
		Syscalls: []specs.LinuxSyscall{
			{Names: []string{"exit_group"}, Action: specs.ActAllow},
		},
	}
}

// Combined bitmask of all CLONE_NEW* namespace flags. Used to deny clone calls
// that attempt to create new namespaces inside the container.
//
//	CLONE_NEWTIME    0x00000080
//	CLONE_NEWNS      0x00020000
//	CLONE_NEWCGROUP  0x02000000
//	CLONE_NEWUTS     0x04000000
//	CLONE_NEWIPC     0x08000000
//	CLONE_NEWUSER    0x10000000
//	CLONE_NEWPID     0x20000000
//	CLONE_NEWNET     0x40000000
const cloneNamespaceMask = 0x7E020080

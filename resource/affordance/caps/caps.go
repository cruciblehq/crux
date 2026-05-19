package caps

import (
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// Linux capability that can be assigned to a process or file.
//
// Stored in lowercase short form without the CAP_ prefix (e.g. "net_admin").
// Use [NormalizeCap] when crossing into the OCI spec or kernel interfaces.
type Cap string

const (
	Chown             Cap = "chown"              // Change file ownership.
	DacOverride       Cap = "dac_override"       // Bypass file permission checks.
	DacReadSearch     Cap = "dac_read_search"    // Bypass file and directory permission checks.
	Fowner            Cap = "fowner"             // Bypass permission checks on operations that require the file owner ID.
	Fsetid            Cap = "fsetid"             // Don't clear set-user-ID and set-group-ID bits when creating files.
	Kill              Cap = "kill"               // Bypass permission checks for sending signals.
	Setgid            Cap = "setgid"             // Make arbitrary manipulations of process GIDs and supplementary GID list.
	Setuid            Cap = "setuid"             // Make arbitrary manipulations of process UIDs.
	Setpcap           Cap = "setpcap"            // Modify process capabilities.
	LinuxImmutable    Cap = "linux_immutable"    // Set the FS_APPEND_FL and FS_IMMUTABLE_FL i-node flags.
	NetBindService    Cap = "net_bind_service"   // Bind a socket to Internet domain privileged ports (port numbers less than 1024).
	NetBroadcast      Cap = "net_broadcast"      // Make socket broadcasts, and listen to multicast.
	NetAdmin          Cap = "net_admin"          // Perform various network-related operations.
	NetRaw            Cap = "net_raw"            // Use RAW and PACKET sockets.
	IpcLock           Cap = "ipc_lock"           // Lock memory (mlock(2), mlockall(2), mmap(2) with MAP_LOCKED).
	IpcOwner          Cap = "ipc_owner"          // Bypass permission checks for operations on System V IPC objects.
	SysModule         Cap = "sys_module"         // Load and unload kernel modules.
	SysRawio          Cap = "sys_rawio"          // Perform I/O port operations (ioperm(2) and iopl(2)).
	SysChroot         Cap = "sys_chroot"         // Use chroot(2).
	SysPtrace         Cap = "sys_ptrace"         // Trace arbitrary processes using ptrace(2).
	SysPacct          Cap = "sys_pacct"          // Configure process accounting.
	SysAdmin          Cap = "sys_admin"          // Perform a range of system administration operations.
	SysBoot           Cap = "sys_boot"           // Use reboot(2) and kexec_load(2), reboot the system, or enable Ctrl-Alt-Del.
	SysNice           Cap = "sys_nice"           // Change the nice value of arbitrary processes.
	SysResource       Cap = "sys_resource"       // Set resource limits on arbitrary processes.
	SysTime           Cap = "sys_time"           // Set system clock (settimeofday(2), stime(2), adjtimex(2)), and set real-time (RTC) alarms.
	SysTtyConfig      Cap = "sys_tty_config"     // Use vhangup(2), and employ various privileged ioctl(2) operations on virtual terminals.
	Mknod             Cap = "mknod"              // Create special files using mknod(2).
	Lease             Cap = "lease"              // Establish leases on arbitrary files (see fcntl(2) F_SETLEASE command).
	AuditWrite        Cap = "audit_write"        // Write records to kernel auditing log.
	AuditControl      Cap = "audit_control"      // Enable and disable kernel auditing, change auditing filter rules, and read audit logs via netlink(7).
	Setfcap           Cap = "setfcap"            // Set file capabilities.
	MacOverride       Cap = "mac_override"       // Override Mandatory Access Control (MAC) policy.
	MacAdmin          Cap = "mac_admin"          // Perform MAC administration operations.
	Syslog            Cap = "syslog"             // Perform privileged syslog(2) operations such as setting the console log level.
	WakeAlarm         Cap = "wake_alarm"         // Trigger something that will wake up the system from a suspended state.
	BlockSuspend      Cap = "block_suspend"      // Employ features that can block system suspend.
	AuditRead         Cap = "audit_read"         // Read the kernel's audit log via netlink(7).
	Perfmon           Cap = "perfmon"            // Use performance monitoring counters.
	Bpf               Cap = "bpf"               // Employ various BPF features that require CAP_BPF.
	CheckpointRestore Cap = "checkpoint_restore" // Use CRIU to checkpoint and restore processes.
)

// Parses a capability name, returning an error for unknown values.
//
// The name should be lowercase without the CAP_ prefix, matching the format
// used in .w source files and in the Grant.Name field.
func Parse(s string) (Cap, error) {
	c := Cap(s)
	switch c {
	case Chown, DacOverride, DacReadSearch, Fowner, Fsetid,
		Kill, Setgid, Setuid, Setpcap, LinuxImmutable,
		NetBindService, NetBroadcast, NetAdmin, NetRaw,
		IpcLock, IpcOwner, SysModule, SysRawio, SysChroot,
		SysPtrace, SysPacct, SysAdmin, SysBoot, SysNice,
		SysResource, SysTime, SysTtyConfig, Mknod, Lease,
		AuditWrite, AuditControl, Setfcap, MacOverride,
		MacAdmin, Syslog, WakeAlarm, BlockSuspend,
		AuditRead, Perfmon, Bpf, CheckpointRestore:
		return c, nil
	default:
		return "", crex.Wrapf(ErrUnknownCap, "unknown capability %q", s)
	}
}

// Returns the kernel-style fully qualified name for a capability.
//
// The catalog stores capabilities in their lowercase short form (e.g.
// "net_admin") because that is the form accepted in source. The OCI runtime
// spec and the kernel both expect the uppercase CAP_-prefixed form (e.g.
// "CAP_NET_ADMIN"), so this helper bridges the two representations whenever
// a name crosses into the spec.
func Normalize(c Cap) string {
	return "CAP_" + strings.ToUpper(string(c))
}

package seccomp

// Set of x86_64 syscall names recognized by the kernel 6.18 ABI.
//
// Used to reject grants that name a syscall the kernel does not expose,
// ensuring the seccomp baseline cannot leak permission for a typo. The
// list is derived from the official syscall table for Linux 6.18, filtered
// to include only entries with the "common" or "64" ABI, which are the ones
// exposed to 64-bit user-space processes. The table is available at
// arch/x86/entry/syscalls/syscall_64.tbl.
var syscalls = map[string]struct{}{
	"read":                    {}, // Reads bytes from a file descriptor into a buffer.
	"write":                   {}, // Writes bytes from a buffer to a file descriptor.
	"open":                    {}, // Opens or creates a file by path and returns a descriptor.
	"close":                   {}, // Closes a file descriptor.
	"stat":                    {}, // Returns file status by path, following symlinks.
	"fstat":                   {}, // Returns file status for an open descriptor.
	"lstat":                   {}, // Returns file status by path without following symlinks.
	"poll":                    {}, // Waits for events on a set of file descriptors.
	"lseek":                   {}, // Repositions the offset of an open descriptor.
	"mmap":                    {}, // Maps files or anonymous pages into the process address space.
	"mprotect":                {}, // Changes protection on a region of mapped memory.
	"munmap":                  {}, // Unmaps a region of memory.
	"brk":                     {}, // Adjusts the program break to grow or shrink the heap.
	"rt_sigaction":            {}, // Installs an action for a signal.
	"rt_sigprocmask":          {}, // Examines or changes the blocked signal mask.
	"rt_sigreturn":            {}, // Returns from a signal handler and restores context.
	"ioctl":                   {}, // Performs device-specific control operations on a descriptor.
	"pread64":                 {}, // Reads from a descriptor at an explicit offset.
	"pwrite64":                {}, // Writes to a descriptor at an explicit offset.
	"readv":                   {}, // Reads into multiple buffers in one call.
	"writev":                  {}, // Writes from multiple buffers in one call.
	"access":                  {}, // Checks accessibility of a file by real UID and GID.
	"pipe":                    {}, // Creates a unidirectional pipe pair.
	"select":                  {}, // Waits for activity on sets of file descriptors with a timeout.
	"sched_yield":             {}, // Yields the processor to other runnable threads.
	"mremap":                  {}, // Remaps an existing memory mapping to a new size or address.
	"msync":                   {}, // Flushes changes in a memory mapping back to storage.
	"mincore":                 {}, // Reports which pages of a mapping are resident in memory.
	"madvise":                 {}, // Advises the kernel about expected use of a memory range.
	"shmget":                  {}, // Allocates or looks up a System V shared memory segment.
	"shmat":                   {}, // Attaches a System V shared memory segment to the address space.
	"shmctl":                  {}, // Controls a System V shared memory segment.
	"dup":                     {}, // Duplicates a file descriptor.
	"dup2":                    {}, // Duplicates a file descriptor onto a specified target descriptor.
	"pause":                   {}, // Suspends the calling thread until a signal is delivered.
	"nanosleep":               {}, // Suspends execution for a high-resolution interval.
	"getitimer":               {}, // Gets the current interval timer for the process.
	"alarm":                   {}, // Schedules a SIGALRM after a number of seconds.
	"setitimer":               {}, // Sets an interval timer for the process.
	"getpid":                  {}, // Returns the calling process ID.
	"sendfile":                {}, // Transfers data between two descriptors inside the kernel.
	"socket":                  {}, // Creates a communication endpoint.
	"connect":                 {}, // Initiates a connection on a socket.
	"accept":                  {}, // Accepts a connection on a listening socket.
	"sendto":                  {}, // Sends a message on a socket, possibly with a destination address.
	"recvfrom":                {}, // Receives a message from a socket and reports the source address.
	"sendmsg":                 {}, // Sends a message described by a msghdr on a socket.
	"recvmsg":                 {}, // Receives a message described by a msghdr on a socket.
	"shutdown":                {}, // Shuts down part of a full-duplex socket connection.
	"bind":                    {}, // Assigns a local address to a socket.
	"listen":                  {}, // Marks a socket as accepting connections.
	"getsockname":             {}, // Returns the local address of a socket.
	"getpeername":             {}, // Returns the remote address of a connected socket.
	"socketpair":              {}, // Creates a pair of connected sockets.
	"setsockopt":              {}, // Sets an option on a socket.
	"getsockopt":              {}, // Gets an option from a socket.
	"clone":                   {}, // Creates a new task with selectable shared resources.
	"fork":                    {}, // Creates a new process by duplicating the caller.
	"vfork":                   {}, // Creates a child that shares the parent address space until exec.
	"execve":                  {}, // Executes a program in the calling process image.
	"exit":                    {}, // Terminates the calling thread.
	"wait4":                   {}, // Waits for a child process to change state and reports usage.
	"kill":                    {}, // Sends a signal to a process or process group.
	"uname":                   {}, // Returns kernel and system identification.
	"semget":                  {}, // Allocates or looks up a System V semaphore set.
	"semop":                   {}, // Performs operations on a System V semaphore set.
	"semctl":                  {}, // Controls a System V semaphore set.
	"shmdt":                   {}, // Detaches a System V shared memory segment.
	"msgget":                  {}, // Allocates or looks up a System V message queue.
	"msgsnd":                  {}, // Sends a message to a System V message queue.
	"msgrcv":                  {}, // Receives a message from a System V message queue.
	"msgctl":                  {}, // Controls a System V message queue.
	"fcntl":                   {}, // Performs file-descriptor control operations.
	"flock":                   {}, // Applies or removes an advisory lock on an open file.
	"fsync":                   {}, // Flushes file data and metadata to disk.
	"fdatasync":               {}, // Flushes file data to disk without metadata.
	"truncate":                {}, // Truncates a file by path to a specified length.
	"ftruncate":               {}, // Truncates an open file to a specified length.
	"getdents":                {}, // Reads directory entries from an open directory.
	"getcwd":                  {}, // Returns the current working directory path.
	"chdir":                   {}, // Changes the current working directory by path.
	"fchdir":                  {}, // Changes the current working directory by descriptor.
	"rename":                  {}, // Renames or moves a file.
	"mkdir":                   {}, // Creates a directory.
	"rmdir":                   {}, // Removes an empty directory.
	"creat":                   {}, // Creates or truncates a file and opens it for writing.
	"link":                    {}, // Creates a hard link to a file.
	"unlink":                  {}, // Removes a name from the filesystem.
	"symlink":                 {}, // Creates a symbolic link.
	"readlink":                {}, // Reads the target of a symbolic link.
	"chmod":                   {}, // Changes the mode bits of a file by path.
	"fchmod":                  {}, // Changes the mode bits of an open file.
	"chown":                   {}, // Changes the owner and group of a file by path.
	"fchown":                  {}, // Changes the owner and group of an open file.
	"lchown":                  {}, // Changes the owner and group of a symbolic link.
	"umask":                   {}, // Sets the file-mode creation mask.
	"gettimeofday":            {}, // Returns the current wall-clock time.
	"getrlimit":               {}, // Gets a resource limit for the calling process.
	"getrusage":               {}, // Returns resource usage statistics.
	"sysinfo":                 {}, // Returns system memory and load information.
	"times":                   {}, // Returns process and children CPU time accounting.
	"ptrace":                  {}, // Traces or controls another process for debugging.
	"getuid":                  {}, // Returns the real user ID of the caller.
	"syslog":                  {}, // Reads from or controls the kernel ring buffer.
	"getgid":                  {}, // Returns the real group ID of the caller.
	"setuid":                  {}, // Sets the effective and possibly real and saved user ID.
	"setgid":                  {}, // Sets the effective and possibly real and saved group ID.
	"geteuid":                 {}, // Returns the effective user ID of the caller.
	"getegid":                 {}, // Returns the effective group ID of the caller.
	"setpgid":                 {}, // Sets the process group ID of a process.
	"getppid":                 {}, // Returns the parent process ID.
	"getpgrp":                 {}, // Returns the process group ID of the caller.
	"setsid":                  {}, // Creates a new session and becomes the session leader.
	"setreuid":                {}, // Sets the real and effective user IDs.
	"setregid":                {}, // Sets the real and effective group IDs.
	"getgroups":               {}, // Returns the supplementary group list.
	"setgroups":               {}, // Sets the supplementary group list.
	"setresuid":               {}, // Sets real, effective, and saved user IDs.
	"getresuid":               {}, // Returns real, effective, and saved user IDs.
	"setresgid":               {}, // Sets real, effective, and saved group IDs.
	"getresgid":               {}, // Returns real, effective, and saved group IDs.
	"getpgid":                 {}, // Returns the process group ID of a process.
	"setfsuid":                {}, // Sets the filesystem user ID of the caller.
	"setfsgid":                {}, // Sets the filesystem group ID of the caller.
	"getsid":                  {}, // Returns the session ID of a process.
	"capget":                  {}, // Reads the capability sets of a thread.
	"capset":                  {}, // Sets the capability sets of a thread.
	"rt_sigpending":           {}, // Returns the set of pending signals.
	"rt_sigtimedwait":         {}, // Waits for a signal in a set with a timeout.
	"rt_sigqueueinfo":         {}, // Sends a signal with queued data to a process.
	"rt_sigsuspend":           {}, // Replaces the signal mask and suspends until delivery.
	"sigaltstack":             {}, // Sets or queries the alternate signal stack.
	"utime":                   {}, // Sets file access and modification times by path.
	"mknod":                   {}, // Creates a regular, device, or FIFO file node.
	"personality":             {}, // Sets the process execution domain.
	"ustat":                   {}, // Returns filesystem statistics by device number.
	"statfs":                  {}, // Returns filesystem statistics by path.
	"fstatfs":                 {}, // Returns filesystem statistics for an open descriptor.
	"sysfs":                   {}, // Returns filesystem type information by index or name.
	"getpriority":             {}, // Returns the scheduling priority of a process or group.
	"setpriority":             {}, // Sets the scheduling priority of a process or group.
	"sched_setparam":          {}, // Sets scheduling parameters of a thread.
	"sched_getparam":          {}, // Returns scheduling parameters of a thread.
	"sched_setscheduler":      {}, // Sets the scheduling policy and parameters of a thread.
	"sched_getscheduler":      {}, // Returns the scheduling policy of a thread.
	"sched_get_priority_max":  {}, // Returns the maximum priority for a scheduling policy.
	"sched_get_priority_min":  {}, // Returns the minimum priority for a scheduling policy.
	"sched_rr_get_interval":   {}, // Returns the round-robin time slice for a thread.
	"mlock":                   {}, // Locks pages in physical memory.
	"munlock":                 {}, // Unlocks pages in physical memory.
	"mlockall":                {}, // Locks all current and future pages in memory.
	"munlockall":              {}, // Unlocks all locked pages.
	"vhangup":                 {}, // Hangs up the current terminal.
	"modify_ldt":              {}, // Reads or writes the per-process local descriptor table.
	"pivot_root":              {}, // Changes the root and put_old directories.
	"_sysctl":                 {}, // Reads or writes kernel parameters via the legacy sysctl interface (deprecated).
	"prctl":                   {}, // Performs operations on the calling process or thread.
	"arch_prctl":              {}, // Performs architecture-specific thread state operations.
	"adjtimex":                {}, // Tunes the kernel clock.
	"setrlimit":               {}, // Sets a resource limit for the calling process.
	"chroot":                  {}, // Changes the root directory of the calling process.
	"sync":                    {}, // Flushes filesystem caches to disk.
	"acct":                    {}, // Enables or disables BSD-style process accounting.
	"settimeofday":            {}, // Sets the wall-clock time and timezone.
	"mount":                   {}, // Mounts a filesystem.
	"umount2":                 {}, // Unmounts a filesystem with flags.
	"swapon":                  {}, // Enables a swap file or device.
	"swapoff":                 {}, // Disables a swap file or device.
	"reboot":                  {}, // Reboots, halts, or powers off the system.
	"sethostname":             {}, // Sets the system hostname.
	"setdomainname":           {}, // Sets the system NIS domain name.
	"iopl":                    {}, // Sets the I/O privilege level of the calling thread.
	"ioperm":                  {}, // Sets I/O port access permissions.
	"init_module":             {}, // Loads a kernel module from a buffer.
	"delete_module":           {}, // Unloads a kernel module by name.
	"quotactl":                {}, // Controls disk quotas.
	"gettid":                  {}, // Returns the kernel thread ID of the caller.
	"readahead":               {}, // Initiates readahead on a file.
	"setxattr":                {}, // Sets an extended attribute by path.
	"lsetxattr":               {}, // Sets an extended attribute on a symbolic link.
	"fsetxattr":               {}, // Sets an extended attribute on an open file.
	"getxattr":                {}, // Reads an extended attribute by path.
	"lgetxattr":               {}, // Reads an extended attribute from a symbolic link.
	"fgetxattr":               {}, // Reads an extended attribute from an open file.
	"listxattr":               {}, // Lists extended attributes by path.
	"llistxattr":              {}, // Lists extended attributes on a symbolic link.
	"flistxattr":              {}, // Lists extended attributes on an open file.
	"removexattr":             {}, // Removes an extended attribute by path.
	"lremovexattr":            {}, // Removes an extended attribute from a symbolic link.
	"fremovexattr":            {}, // Removes an extended attribute from an open file.
	"tkill":                   {}, // Sends a signal to a thread by TID.
	"time":                    {}, // Returns the wall-clock time in seconds since the epoch.
	"futex":                   {}, // Performs fast userspace synchronization operations.
	"sched_setaffinity":       {}, // Sets the CPU affinity mask of a thread.
	"sched_getaffinity":       {}, // Returns the CPU affinity mask of a thread.
	"io_setup":                {}, // Creates an asynchronous I/O context.
	"io_destroy":              {}, // Destroys an asynchronous I/O context.
	"io_getevents":            {}, // Reads completed events from an AIO context.
	"io_submit":               {}, // Submits asynchronous I/O requests.
	"io_cancel":               {}, // Cancels an outstanding asynchronous I/O request.
	"epoll_create":            {}, // Creates an epoll instance.
	"remap_file_pages":        {}, // Remaps pages of a mapping (deprecated).
	"getdents64":              {}, // Reads directory entries with 64-bit inode fields.
	"set_tid_address":         {}, // Sets the address used by clear_child_tid on thread exit.
	"restart_syscall":         {}, // Restarts a syscall after an interrupted wait.
	"semtimedop":              {}, // Performs operations on a semaphore set with a timeout.
	"fadvise64":               {}, // Advises the kernel about expected file access patterns.
	"timer_create":            {}, // Creates a POSIX per-process timer.
	"timer_settime":           {}, // Arms or disarms a POSIX timer.
	"timer_gettime":           {}, // Returns the remaining time on a POSIX timer.
	"timer_getoverrun":        {}, // Returns the overrun count for a POSIX timer.
	"timer_delete":            {}, // Deletes a POSIX timer.
	"clock_settime":           {}, // Sets the value of a clock.
	"clock_gettime":           {}, // Returns the value of a clock.
	"clock_getres":            {}, // Returns the resolution of a clock.
	"clock_nanosleep":         {}, // Sleeps until a clock reaches a specified time.
	"exit_group":              {}, // Terminates all threads in the calling process.
	"epoll_wait":              {}, // Waits for events on an epoll instance.
	"epoll_ctl":               {}, // Adds, modifies, or removes targets on an epoll instance.
	"tgkill":                  {}, // Sends a signal to a specific thread in a thread group.
	"utimes":                  {}, // Sets file access and modification times by path.
	"mbind":                   {}, // Sets the NUMA memory policy of a memory range.
	"set_mempolicy":           {}, // Sets the default NUMA memory policy of the caller.
	"get_mempolicy":           {}, // Returns the NUMA memory policy of the caller or address.
	"mq_open":                 {}, // Opens or creates a POSIX message queue.
	"mq_unlink":               {}, // Removes a POSIX message queue.
	"mq_timedsend":            {}, // Sends a message to a POSIX message queue with a timeout.
	"mq_timedreceive":         {}, // Receives a message from a POSIX message queue with a timeout.
	"mq_notify":               {}, // Registers for notification of POSIX message queue events.
	"mq_getsetattr":           {}, // Gets or sets POSIX message queue attributes.
	"kexec_load":              {}, // Loads a new kernel for kexec.
	"waitid":                  {}, // Waits for a child to change state with extended options.
	"add_key":                 {}, // Adds a key to a kernel keyring.
	"request_key":             {}, // Requests a key from the kernel keyring subsystem.
	"keyctl":                  {}, // Manipulates kernel keys and keyrings.
	"ioprio_set":              {}, // Sets the I/O priority of a thread or process.
	"ioprio_get":              {}, // Returns the I/O priority of a thread or process.
	"inotify_init":            {}, // Creates an inotify instance.
	"inotify_add_watch":       {}, // Adds a watch to an inotify instance.
	"inotify_rm_watch":        {}, // Removes a watch from an inotify instance.
	"migrate_pages":           {}, // Moves pages of a process between NUMA nodes.
	"openat":                  {}, // Opens a file relative to a directory descriptor.
	"mkdirat":                 {}, // Creates a directory relative to a directory descriptor.
	"mknodat":                 {}, // Creates a file node relative to a directory descriptor.
	"fchownat":                {}, // Changes ownership relative to a directory descriptor.
	"futimesat":               {}, // Sets file times relative to a directory descriptor.
	"newfstatat":              {}, // Returns file status relative to a directory descriptor.
	"unlinkat":                {}, // Removes a name relative to a directory descriptor.
	"renameat":                {}, // Renames a file relative to directory descriptors.
	"linkat":                  {}, // Creates a hard link relative to directory descriptors.
	"symlinkat":               {}, // Creates a symbolic link relative to a directory descriptor.
	"readlinkat":              {}, // Reads a symbolic link target relative to a directory descriptor.
	"fchmodat":                {}, // Changes mode bits relative to a directory descriptor.
	"faccessat":               {}, // Checks file accessibility relative to a directory descriptor.
	"pselect6":                {}, // Waits on descriptor sets with a signal mask and timeout.
	"ppoll":                   {}, // Waits for events on descriptors with a signal mask and timeout.
	"unshare":                 {}, // Unshares parts of the calling process execution context.
	"set_robust_list":         {}, // Sets the head of the robust futex list.
	"get_robust_list":         {}, // Returns the head of the robust futex list of a thread.
	"splice":                  {}, // Moves data between two descriptors via a kernel pipe.
	"tee":                     {}, // Duplicates pipe contents from one pipe to another.
	"sync_file_range":         {}, // Flushes a byte range of a file to disk.
	"vmsplice":                {}, // Splices user pages into or out of a pipe.
	"move_pages":              {}, // Moves specified pages of a process between NUMA nodes.
	"utimensat":               {}, // Sets file timestamps with nanosecond precision.
	"epoll_pwait":             {}, // Waits on an epoll instance with a signal mask.
	"signalfd":                {}, // Creates a descriptor that delivers signals.
	"timerfd_create":          {}, // Creates a descriptor for timer events.
	"eventfd":                 {}, // Creates an event-counter descriptor.
	"fallocate":               {}, // Preallocates or punches holes in a file.
	"timerfd_settime":         {}, // Arms or disarms a timerfd.
	"timerfd_gettime":         {}, // Returns the remaining time on a timerfd.
	"accept4":                 {}, // Accepts a connection with flags such as nonblock and cloexec.
	"signalfd4":               {}, // Creates a signalfd descriptor with flags.
	"eventfd2":                {}, // Creates an eventfd descriptor with flags.
	"epoll_create1":           {}, // Creates an epoll instance with flags.
	"dup3":                    {}, // Duplicates a file descriptor with flags.
	"pipe2":                   {}, // Creates a pipe pair with flags.
	"inotify_init1":           {}, // Creates an inotify instance with flags.
	"preadv":                  {}, // Reads into multiple buffers at an explicit offset.
	"pwritev":                 {}, // Writes from multiple buffers at an explicit offset.
	"rt_tgsigqueueinfo":       {}, // Sends a signal with queued data to a thread.
	"perf_event_open":         {}, // Opens a performance event counter or sampler.
	"recvmmsg":                {}, // Receives multiple messages on a socket in one call.
	"fanotify_init":           {}, // Creates a fanotify notification group.
	"fanotify_mark":           {}, // Adds, removes, or modifies a fanotify mark.
	"prlimit64":               {}, // Gets or sets a resource limit of another process.
	"name_to_handle_at":       {}, // Returns an opaque handle for a path.
	"open_by_handle_at":       {}, // Opens a file referred to by a handle.
	"clock_adjtime":           {}, // Tunes a POSIX clock.
	"syncfs":                  {}, // Flushes a filesystem associated with a descriptor.
	"sendmmsg":                {}, // Sends multiple messages on a socket in one call.
	"setns":                   {}, // Reassociates a thread with a namespace by descriptor.
	"getcpu":                  {}, // Returns the CPU and NUMA node of the caller.
	"process_vm_readv":        {}, // Reads memory from another process into local buffers.
	"process_vm_writev":       {}, // Writes from local buffers into another process.
	"kcmp":                    {}, // Compares two processes for kernel resource sharing.
	"finit_module":            {}, // Loads a kernel module from an open descriptor.
	"sched_setattr":           {}, // Sets extended scheduling attributes of a thread.
	"sched_getattr":           {}, // Returns extended scheduling attributes of a thread.
	"renameat2":               {}, // Renames a file with extended flags.
	"seccomp":                 {}, // Configures seccomp filtering for the caller.
	"getrandom":               {}, // Returns cryptographically secure random bytes.
	"memfd_create":            {}, // Creates an anonymous file in a memfd.
	"kexec_file_load":         {}, // Loads a new kernel from a file descriptor.
	"bpf":                     {}, // Loads, queries, or interacts with BPF programs and maps.
	"execveat":                {}, // Executes a program by descriptor with optional flags.
	"userfaultfd":             {}, // Creates a userfault file descriptor.
	"membarrier":              {}, // Issues memory barriers across threads.
	"mlock2":                  {}, // Locks pages in memory with extended flags.
	"copy_file_range":         {}, // Copies a range of bytes between two files in kernel.
	"preadv2":                 {}, // Reads into multiple buffers at an offset with flags.
	"pwritev2":                {}, // Writes from multiple buffers at an offset with flags.
	"pkey_mprotect":           {}, // Sets protection and a protection key on a memory range.
	"pkey_alloc":              {}, // Allocates a memory protection key.
	"pkey_free":               {}, // Frees a memory protection key.
	"statx":                   {}, // Returns extended file status with selectable fields.
	"io_pgetevents":           {}, // Reads AIO events with a signal mask and timeout.
	"rseq":                    {}, // Registers a restartable sequences area for the thread.
	"uretprobe":               {}, // Triggers a uretprobe trap (kernel internal).
	"uprobe":                  {}, // Triggers a uprobe trap (kernel internal).
	"pidfd_send_signal":       {}, // Sends a signal to a process referenced by a pidfd.
	"io_uring_setup":          {}, // Creates an io_uring instance.
	"io_uring_enter":          {}, // Submits and reaps io_uring requests.
	"io_uring_register":       {}, // Registers resources with an io_uring instance.
	"open_tree":               {}, // Returns a detached descriptor for a mount tree.
	"move_mount":              {}, // Moves a mount tree to a new location.
	"fsopen":                  {}, // Creates a filesystem context for new-style mounting.
	"fsconfig":                {}, // Configures a filesystem context.
	"fsmount":                 {}, // Materializes a mount from a filesystem context.
	"fspick":                  {}, // Picks an existing superblock into a filesystem context.
	"pidfd_open":              {}, // Returns a pidfd for a process by PID.
	"clone3":                  {}, // Creates a new task with extended clone arguments.
	"close_range":             {}, // Closes a range of file descriptors.
	"openat2":                 {}, // Opens a file relative to a directory with extended controls.
	"pidfd_getfd":             {}, // Returns a duplicate of a remote process file descriptor.
	"faccessat2":              {}, // Checks file accessibility with flags relative to a directory.
	"process_madvise":         {}, // Applies madvise to a memory range of another process.
	"epoll_pwait2":            {}, // Waits on an epoll instance with a nanosecond timeout.
	"mount_setattr":           {}, // Changes mount attributes on an existing mount.
	"quotactl_fd":             {}, // Performs quotactl operations against a descriptor.
	"landlock_create_ruleset": {}, // Creates a landlock ruleset for unprivileged sandboxing.
	"landlock_add_rule":       {}, // Adds a rule to a landlock ruleset.
	"landlock_restrict_self":  {}, // Enforces a landlock ruleset on the caller.
	"memfd_secret":            {}, // Creates a memfd backed by secret kernel memory.
	"process_mrelease":        {}, // Releases the memory of a dying process.
	"futex_waitv":             {}, // Waits on a vector of futexes.
	"set_mempolicy_home_node": {}, // Sets the preferred NUMA home node for a memory range.
	"cachestat":               {}, // Returns page cache state for a range of a file.
	"fchmodat2":               {}, // Changes mode bits relative to a directory with flags.
	"map_shadow_stack":        {}, // Allocates a shadow stack for the calling thread.
	"futex_wake":              {}, // Wakes waiters on a futex.
	"futex_wait":              {}, // Waits on a futex with extended arguments.
	"futex_requeue":           {}, // Requeues waiters between two futexes.
	"statmount":               {}, // Returns extended attributes of a mount.
	"listmount":               {}, // Lists mounts in the current mount namespace.
	"lsm_get_self_attr":       {}, // Reads an LSM attribute of the caller.
	"lsm_set_self_attr":       {}, // Sets an LSM attribute of the caller.
	"lsm_list_modules":        {}, // Lists active LSM modules.
	"mseal":                   {}, // Seals a memory range against further protection changes.
	"setxattrat":              {}, // Sets an extended attribute relative to a directory.
	"getxattrat":              {}, // Reads an extended attribute relative to a directory.
	"listxattrat":             {}, // Lists extended attributes relative to a directory.
	"removexattrat":           {}, // Removes an extended attribute relative to a directory.
	"open_tree_attr":          {}, // Opens a mount tree and sets attributes in one call.
	"file_getattr":            {}, // Returns extended file attributes for a descriptor.
	"file_setattr":            {}, // Sets extended file attributes on a descriptor.
}

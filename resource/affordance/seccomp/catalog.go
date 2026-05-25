package seccomp

// Syscall names that carry curated sub-filters.
const (
	sysIoctl = "ioctl" // Multiplexed device control syscall, sub-filtered by request value at arg 1.
	sysFcntl = "fcntl" // File descriptor control syscall, sub-filtered by cmd value at arg 1.
	sysPrctl = "prctl" // Process control syscall, sub-filtered by option value at arg 0.
)

// fcntl(2) cmd constants.
const (
	fDupfd        = 0    // Duplicates an existing file descriptor (F_DUPFD).
	fGetfd        = 1    // Reads the close-on-exec flag for a file descriptor (F_GETFD).
	fSetfd        = 2    // Writes the close-on-exec flag for a file descriptor (F_SETFD).
	fGetfl        = 3    // Reads the file status flags (F_GETFL).
	fSetfl        = 4    // Writes the file status flags (F_SETFL).
	fGetlk        = 5    // Queries an existing record lock (F_GETLK).
	fSetlk        = 6    // Acquires or releases a record lock without blocking (F_SETLK).
	fSetlkw       = 7    // Acquires or releases a record lock, blocking until available (F_SETLKW).
	fSetown       = 8    // Sets the process or process group that receives SIGIO/SIGURG (F_SETOWN).
	fGetown       = 9    // Reads the process or process group that receives SIGIO/SIGURG (F_GETOWN).
	fSetsig       = 10   // Sets the signal delivered on asynchronous I/O readiness (F_SETSIG).
	fGetsig       = 11   // Reads the signal delivered on asynchronous I/O readiness (F_GETSIG).
	fOfdGetlk     = 36   // Queries an open file description lock (F_OFD_GETLK).
	fOfdSetlk     = 37   // Acquires or releases an open file description lock without blocking (F_OFD_SETLK).
	fOfdSetlkw    = 38   // Acquires or releases an open file description lock, blocking (F_OFD_SETLKW).
	fSetlease     = 1024 // Acquires or releases a file lease (F_SETLEASE).
	fGetlease     = 1025 // Reads the current file lease type (F_GETLEASE).
	fDupfdCloexec = 1030 // Duplicates a file descriptor with the close-on-exec flag set (F_DUPFD_CLOEXEC).
	fSetpipeSz    = 1031 // Resizes a pipe's kernel buffer (F_SETPIPE_SZ).
	fGetpipeSz    = 1032 // Reads a pipe's kernel buffer size (F_GETPIPE_SZ).
	fAddSeals     = 1033 // Adds seals to a memfd to restrict future operations (F_ADD_SEALS).
	fGetSeals     = 1034 // Reads the active seals on a memfd (F_GET_SEALS).
)

// ioctl(2) request constants.
const (
	ioctlTcgets         = 0x5401     // Reads terminal attributes from a tty (TCGETS).
	ioctlTcsets         = 0x5402     // Writes terminal attributes to a tty (TCSETS).
	ioctlTiocsctty      = 0x540e     // Makes a tty the controlling terminal of the calling session (TIOCSCTTY).
	ioctlTiocgpgrp      = 0x540f     // Reads the foreground process group ID of a tty (TIOCGPGRP).
	ioctlTiocspgrp      = 0x5410     // Sets the foreground process group ID of a tty (TIOCSPGRP).
	ioctlTiocgwinsz     = 0x5413     // Reads the window size (rows/cols) of a tty (TIOCGWINSZ).
	ioctlTiocswinsz     = 0x5414     // Writes the window size (rows/cols) of a tty (TIOCSWINSZ).
	ioctlFionread       = 0x541b     // Reports the number of bytes available to read on a fd (FIONREAD).
	ioctlFionbio        = 0x5421     // Toggles non-blocking I/O on a fd (FIONBIO).
	ioctlTiocnotty      = 0x5422     // Detaches the calling session from its controlling tty (TIOCNOTTY).
	ioctlFionclex       = 0x5450     // Clears the close-on-exec flag for a fd (FIONCLEX).
	ioctlFioclex        = 0x5451     // Sets the close-on-exec flag for a fd (FIOCLEX).
	ioctlSiocgifconf    = 0x8912     // Lists configured network interfaces (SIOCGIFCONF).
	ioctlSiocgifflags   = 0x8913     // Reads the active flags of a network interface (SIOCGIFFLAGS).
	ioctlSiocgifaddr    = 0x8915     // Reads the IPv4 address of a network interface (SIOCGIFADDR).
	ioctlSiocgifnetmask = 0x891b     // Reads the IPv4 netmask of a network interface (SIOCGIFNETMASK).
	ioctlSiocgifhwaddr  = 0x8927     // Reads the link-layer address of a network interface (SIOCGIFHWADDR).
	ioctlSiocgifindex   = 0x8933     // Reads the index of a named network interface (SIOCGIFINDEX).
	ioctlBlkflsbuf      = 0x1261     // Flushes a block device's buffer cache (BLKFLSBUF).
	ioctlBlksszget      = 0x1268     // Reads a block device's logical sector size (BLKSSZGET).
	ioctlBlkdiscard     = 0x1277     // Discards a range of sectors on a block device (BLKDISCARD).
	ioctlBlkpbszget     = 0x127b     // Reads a block device's physical sector size (BLKPBSZGET).
	ioctlBlkgetsize64   = 0x80081272 // Reads a block device's total size in bytes (BLKGETSIZE64).
	ioctlFsGetflags     = 0x80086601 // Reads filesystem-specific inode flags (FS_IOC_GETFLAGS).
	ioctlFsSetflags     = 0x40086602 // Writes filesystem-specific inode flags (FS_IOC_SETFLAGS).
	ioctlFsGetversion   = 0x80087601 // Reads an inode's generation number (FS_IOC_GETVERSION).
	ioctlFsSetversion   = 0x40087602 // Writes an inode's generation number (FS_IOC_SETVERSION).
	ioctlFibmap         = 0x01       // Translates a file logical block to a physical block on disk (FIBMAP).
	ioctlFiemap         = 0xc020660b // Reads a file's extent map from the filesystem (FIEMAP).
)

// prctl(2) option constants.
const (
	prSetPdeathsig      = 1          // Sets the signal delivered when the parent process dies (PR_SET_PDEATHSIG).
	prGetPdeathsig      = 2          // Reads the signal delivered when the parent process dies (PR_GET_PDEATHSIG).
	prGetDumpable       = 3          // Reads the calling process's core-dump flag (PR_GET_DUMPABLE).
	prSetDumpable       = 4          // Writes the calling process's core-dump flag (PR_SET_DUMPABLE).
	prGetKeepcaps       = 7          // Reads whether capabilities are preserved across execve (PR_GET_KEEPCAPS).
	prSetKeepcaps       = 8          // Writes whether capabilities are preserved across execve (PR_SET_KEEPCAPS).
	prSetName           = 15         // Sets the calling thread's name (PR_SET_NAME).
	prGetName           = 16         // Reads the calling thread's name (PR_GET_NAME).
	prGetSeccomp        = 21         // Reads the calling thread's seccomp mode (PR_GET_SECCOMP).
	prSetSeccomp        = 22         // Switches the calling thread into seccomp filter mode (PR_SET_SECCOMP).
	prCapbsetRead       = 23         // Tests whether a capability is in the bounding set (PR_CAPBSET_READ).
	prCapbsetDrop       = 24         // Drops a capability from the bounding set (PR_CAPBSET_DROP).
	prSetTimerslack     = 29         // Sets the kernel's timer slack value for the thread (PR_SET_TIMERSLACK).
	prGetTimerslack     = 30         // Reads the kernel's timer slack value for the thread (PR_GET_TIMERSLACK).
	prSetMm             = 35         // Updates one of the calling process's mm_struct addresses (PR_SET_MM).
	prSetChildSubreaper = 36         // Marks the calling process as a child subreaper (PR_SET_CHILD_SUBREAPER).
	prGetChildSubreaper = 37         // Reads whether the calling process is a child subreaper (PR_GET_CHILD_SUBREAPER).
	prSetNoNewPrivs     = 38         // Locks the no-new-privileges bit on the calling thread (PR_SET_NO_NEW_PRIVS).
	prGetNoNewPrivs     = 39         // Reads the no-new-privileges bit (PR_GET_NO_NEW_PRIVS).
	prGetSpecCtrl       = 52         // Reads the speculation-mitigation state for a feature (PR_GET_SPECULATION_CTRL).
	prSetSpecCtrl       = 53         // Writes the speculation-mitigation state for a feature (PR_SET_SPECULATION_CTRL).
	prSetSud            = 59         // Configures Syscall User Dispatch for the calling thread (PR_SET_SYSCALL_USER_DISPATCH).
	prSetMdwe           = 65         // Enables the memory-deny-write-execute policy on the calling process (PR_SET_MDWE).
	prGetMdwe           = 66         // Reads the memory-deny-write-execute policy state (PR_GET_MDWE).
	prSetVma            = 0x53564d41 // Sets a name on an anonymous VMA (PR_SET_VMA, ARG2 = PR_SET_VMA_ANON_NAME).
)

// Maps ioctl sub-filter names to the allowed request values for argument position 1.
var ioctlSubs = map[string][]uint64{
	"tty":     {ioctlTcgets, ioctlTcsets, ioctlTiocsctty, ioctlTiocgpgrp, ioctlTiocspgrp, ioctlTiocgwinsz, ioctlTiocswinsz, ioctlTiocnotty},
	"fio":     {ioctlFionread, ioctlFionbio, ioctlFioclex, ioctlFionclex},
	"net":     {ioctlSiocgifaddr, ioctlSiocgifflags, ioctlSiocgifnetmask, ioctlSiocgifhwaddr, ioctlSiocgifindex, ioctlSiocgifconf},
	"blk":     {ioctlBlkgetsize64, ioctlBlksszget, ioctlBlkpbszget, ioctlBlkdiscard, ioctlBlkflsbuf},
	"fsflags": {ioctlFsGetflags, ioctlFsSetflags, ioctlFsGetversion, ioctlFsSetversion},
	"fiemap":  {ioctlFibmap, ioctlFiemap},
}

// Maps fcntl sub-filter names to the allowed cmd values for argument position 1.
var fcntlSubs = map[string][]uint64{
	"flags":  {fGetfd, fSetfd, fGetfl, fSetfl},
	"dup":    {fDupfd, fDupfdCloexec},
	"lock":   {fGetlk, fSetlk, fSetlkw, fOfdGetlk, fOfdSetlk, fOfdSetlkw},
	"signal": {fSetown, fGetown, fSetsig, fGetsig},
	"lease":  {fSetlease, fGetlease, fAddSeals, fGetSeals},
	"pipe":   {fSetpipeSz, fGetpipeSz},
}

// Maps prctl sub-filter names to the allowed option values for argument position 0.
var prctlSubs = map[string][]uint64{
	"name":      {prSetName, prGetName},
	"dump":      {prSetDumpable, prGetDumpable},
	"nnp":       {prSetNoNewPrivs, prGetNoNewPrivs},
	"seccomp":   {prSetSeccomp, prGetSeccomp},
	"caps":      {prCapbsetRead, prCapbsetDrop, prSetKeepcaps, prGetKeepcaps},
	"lifecycle": {prSetPdeathsig, prGetPdeathsig, prSetChildSubreaper, prGetChildSubreaper},
	"timer":     {prSetTimerslack, prGetTimerslack},
	"mm":        {prSetMm},
	"spec":      {prGetSpecCtrl, prSetSpecCtrl},
	"mdwe":      {prSetMdwe, prGetMdwe},
	"vma":       {prSetVma},
	"sud":       {prSetSud},
}

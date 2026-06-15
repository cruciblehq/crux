// Package mount implements the .mount AGL subsystem.
//
// A .mount grant adds a kernel VFS filesystem mount to the container's OCI
// runtime spec. Only in-kernel filesystem types are accepted; bind mounts from
// host paths are not supported through this subsystem.
//
// Grant syntax:
//
//	.mount <type> <destination> [size=<quantity>] [mode=<octal>]
//
// type is the filesystem type: tmpfs, proc, sysfs, devpts, mqueue, or cgroup2.
// destination is the absolute path inside the container where the filesystem
// is mounted. size and mode are optional keyword arguments accepted only for
// tmpfs mounts; they set the size limit and permission mode respectively.
package mount

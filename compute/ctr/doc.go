// Package ctr provides a containerd-backed container runtime client.
//
// Callers use [New] to connect to a containerd socket and receive a
// [provider.ContainerRuntime] that can import OCI image archives and run
// containers with a caller-supplied security spec.
package ctr

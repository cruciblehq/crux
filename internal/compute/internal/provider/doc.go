// Package provider defines the shared types for the compute layer.
//
// This package is a leaf dependency that both the compute package and its
// provider sub-packages (e.g. compute/local) import. It owns the [Backend]
// interface and the types needed to compose it.
package provider

// Package oci provides building, reading, and manipulation of OCI container
// images.
//
// The package supports single-platform and multi-platform image creation via
// [Builder] and [MultiPlatformBuilder], reading and validating existing images
// via [ReadIndex], and low-level operations like digest computation and layer
// inspection. All image I/O uses the OCI image layout format.
package oci

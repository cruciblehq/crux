// Package pack provides functionality for packaging built Crucible resources
// into distributable archives.
//
// The package creates zstd-compressed tar archives containing the resource
// manifest (crucible.yaml) and build artifacts from the dist/ directory.
// Archives are ready for deployment or distribution to a Crucible Hub.
// packaging process validates that the resource structure matches its type
// before creating the archive.
//
// Example usage:
//
//	if err := pack.Pack(ctx, pack.PackOptions{
//	    Manifestfile:  "crucible.yaml",
//	    Dist:          "dist",
//	    PackageOutput: "package.tar.zst",
//	}); err != nil {
//	    return err
//	}
//
// The resulting package.tar.zst file contains the complete resource and can
// be deployed using the deploy command or pushed to a Hub registry.
package pack

// Package build compiles Crucible resources into deployable artifacts.
//
// The package reads a resource manifest, selects the appropriate builder
// for the resource type (widget, service, or runtime), and produces build
// artifacts in the specified output directory. Widget builds run esbuild
// locally to bundle JavaScript, while service and runtime builds delegate
// to the cruxd daemon for container-based image creation. Each builder
// implements the Builder interface and handles the specifics of its
// resource type.
//
// Building a resource from its manifest:
//
//	result, err := build.Build(ctx, build.Options{
//	    Manifest: "crucible.yaml",
//	    Output:   "build",
//	    Registry: "http://hub.cruciblehq.xyz:8080",
//	})
//	if err != nil {
//	    return err
//	}
//	fmt.Println("artifacts written to", result.Output)
package build

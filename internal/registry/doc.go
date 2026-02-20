// Package registry provides an HTTP client for the Crucible artifact registry.
//
// [Client] implements the [Registry] interface from [github.com/cruciblehq/spec/registry]
// as an HTTP client against the Hub API, using vendor-specific media types
// (application/vnd.crucible.{name}.v0) in Content-Type and Accept headers.
//
// [ResolveVersion] maps a [reference.Reference] to a concrete version by
// resolving either a channel to its pointed-to version or a semver constraint
// to the highest matching version.
//
// Creating a client and reading a namespace:
//
//	client := registry.NewClient("https://hub.example.com", nil)
//	ns, err := client.ReadNamespace(ctx, "my-namespace")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Resolving a version reference:
//
//	ver, err := registry.ResolveVersion(ctx, client, ref)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(ver.String, *ver.Digest)
package registry

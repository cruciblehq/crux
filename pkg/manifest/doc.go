// Package packager creates distributable archives from Crucible resources.
//
// A manifest describes a Crucible resource and its configuration. Manifests are
// YAML files located at .cruciblerc/manifest.yaml within a resource directory.
// Use [Read] to load and parse a manifest:
//
//	m, err := manifest.Read("/path/to/resource")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// The manifest's [Resource.Type] field determines which concrete type is stored
// in [Manifest.Config]. Use a type assertion to access type-specific fields:
//
//	switch cfg := m.Config.(type) {
//	case *manifest.Widget:
//	    fmt.Println(cfg.Build.Main)
//	case *manifest.Service:
//	    fmt.Println(cfg.Image.Ref)
//	}
//
// Manifest structures use "crux" struct tags for field mapping, decoupling Go
// field names from YAML keys. If the manifest specifies an unrecognized resource
// type, [Read] returns an [UnknownResourceTypeError].
package manifest

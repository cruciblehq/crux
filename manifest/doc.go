// Package manifest defines the structure and parsing logic for resource manifests.
//
// A manifest describes a Crucible resource and its configuration. Manifests are
// YAML files named crucible.yaml at the root of a resource directory.
// Use [Read] to load and parse a manifest:
//
//	m, err := manifest.Read("/path/to/resource/crucible.yaml")
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
//	    fmt.Println(cfg.Build.Main)
//	}
//
// Manifest structures use "field" struct tags for field mapping, decoupling Go
// field names from YAML keys.
package manifest

// Package template provides utilities for discovering and instantiating project
// templates from a filesystem.
//
// Templates are organized as subdirectories within a root directory, each
// containing a metadata file (typically YAML) describing the template. The
// package supports listing available templates, reading their metadata, and
// instantiating them by rendering files and directories with variable
// substitution using Go's text/template engine.
//
// The main purpose of this package is to facilitate the bootstrapping of new
// Crucible resources from predefined templates, allowing users to quickly set
// up projects with standard structures and configurations. It can be used for
// other effects, but the terminology is oriented around Crucible's resource
// model (e.g., [TemplateOptions.Resource]).
//
// Key features:
//
//   - Template discovery: List all available templates in a filesystem, reading
//     their metadata and reporting errors for corrupt or missing templates.
//   - Template instantiation: Render templates into an output directory,
//     substituting variables and handling file/directory creation with
//     configurable permissions.
//   - Metadata parsing: Supports custom decoders for template metadata (e.g.,
//     YAML, JSON).
//   - Safe processing: Templates are rendered in a temporary directory before
//     being moved to the final location to avoid partial output on error.
//
// Types:
//
//   - Template[T]: Represents a discovered template, including its name,
//     metadata (of type T), and any error encountered.
//   - TemplateOptions: Options for instantiating a template (name, output
//     location, permissions, data).
//   - Decoder: Callback for parsing template metadata.
//
// Functions:
//
//   - List: Lists all templates in a filesystem, parsing their metadata.
//   - Create: Instantiates a template into an output directory, rendering files
//     and directories.
//   - buildFromTemplate: Renders a single template file with provided data.
//   - processTemplateEntry: Processes a single file or directory entry from a
//     template.
//   - processTemplate: Walks a template directory and processes all entries.
//
// Usage example:
//
//	templates, err := template.List[TemplateInfo](fs, ".template.yaml", yamlDecoder)
//	if err != nil {
//	    // handle error
//	}
//	for _, t := range templates {
//	    fmt.Println(t.Name, t.Metadata.Description)
//	}
//
// To instantiate a template:
//
//	opts := template.TemplateOptions{
//	    Name:     "widget",
//	    Metafile: ".template.yaml",
//	    Location: "./output",
//	    FileMode: 0644,
//	    DirMode:  0755,
//	    Data:     map[string]any{"ProjectName": "MyWidget"},
//	}
//	err := template.Create[TemplateInfo](fs, &opts)
//	if err != nil {
//	    // handle error
//	}
package template

package template

import (
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

// Describes a single template and its metadata.
//
// Name is both the name of the template and the directory that contains it.
// Each template directory must contain exactly one metadata file (e.g.,
// ".template.yaml"). The path to that file is stored in Metafile, and its
// decoded contents are stored in Metadata. The concrete type of Metadata
// depends on the decoder used when templates are loaded.
type Template[T any] struct {
	Name     string // Template name
	Metafile string // Location of the template metadata file
	Metadata T      // Parsed metadata (type depends on the decoder used)
}

// Contains options for creating a template from a given filesystem.
//
// The instantiation of a template is called a "resource", and its name is
// specified by Resource and its location within the filesystem by Location.
// Template is the name of the template to use, and Metafile is the location
// of the template metadata file. The permissions for created files and
// directories are specified by FileMode and DirMode, respectively. Finally,
// Data contains the data to be used when rendering templates.
type TemplateOptions struct {
	Resource string      // Name of the output resource
	Template string      // Name of the template to use
	Metafile string      // Location of the template metadata file
	Location string      // Output location where the template will be instantiated
	FileMode os.FileMode // Permission bits for created files
	DirMode  os.FileMode // Permission bits for created directories
	Data     any         // Data to be used when rendering templates
}

// Callback that decodes data into the provided variable, used to parse
// metadata files in various formats.
type Decoder[T any] func(data []byte, info *T) error

// Returns a template.FuncMap with custom functions to be used in templates.
//
// These functions can be used to provide additional functionality when
// rendering templates. It defines the following functions:
//
//   - "default": Returns a default value if the provided value is nil, an empty
//     string, or an empty slice/array. Otherwise, it returns the provided value.
//   - "json": Converts a value to its JSON string representation.
//   - "slice": Creates a slice from the provided arguments.
func templateFuncs() template.FuncMap {
	return template.FuncMap{

		// "default" returns [defaultVal] if [val] is nil, an empty string,
		// or an empty slice/array. Otherwise, it returns [val].
		//
		// Usage in template:
		//   {{ .SomeField | default "defaultValue" }}
		"default": func(defaultVal, val interface{}) interface{} {

			// Return default on nil
			if val == nil {
				return defaultVal
			}

			// Return default on empty string
			if str, ok := val.(string); ok && str == "" {
				return defaultVal
			}

			// Use reflect for any slice type
			v := reflect.ValueOf(val)
			if v.Kind() == reflect.Slice && v.Len() == 0 {
				return defaultVal
			}

			return val
		},

		// "json" converts [v] to its JSON string representation.
		//
		// Usage in template:
		//   {{ .SomeField | json }}
		"json": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},

		// "slice" creates a slice from the provided arguments.
		//
		// Usage in template:
		//   {{ slice "a" "b" "c" }}
		"slice": func(args ...interface{}) []interface{} {
			return args
		},
	}
}

// Processes the given data as a Go template, substituting variables according
// to template.Execute.
//
// Returns the processed data or an error if template parsing or execution fails.
func executeTemplate(templateFile string, outPath string, templateContents []byte, data any) ([]byte, error) {

	// Parse data as a template
	tmpl, err := template.New(outPath).
		Funcs(templateFuncs()).
		Parse(string(templateContents))

	if err != nil {
		return nil, &TemplateError{
			Path: templateFile,
			Err:  err,
		}
	}

	// Execute the template with the provided variables
	var builder strings.Builder

	if err := tmpl.Execute(&builder, data); err != nil {
		return nil, errorFromExecuteError(err, templateFile)
	}

	return []byte(builder.String()), nil
}

// Processes a single entry (file or directory) from the template filesystem.
//
// If the entry is a directory, it creates the corresponding directory in the
// output location. If it's a file, it reads the file, processes it as a
// template if it has a .tmpl extension, and writes it to the output location.
// It returns an error if any operation fails.
func instantiateTemplateEntry(ifs fs.FS, entryPath string, entry fs.DirEntry, outPath string, options *TemplateOptions) error {

	if entry.IsDir() {

		// Create the output directory
		if err := os.MkdirAll(outPath, options.DirMode); err != nil {
			return &OutputError{
				Path: outPath,
				Err:  err,
			}
		}

		return nil
	}

	// If it's a file, read it and write it to the output location
	templateContents, err := fs.ReadFile(ifs, entryPath)

	if err != nil {

		// This could be anything, since it depends on the filesystem
		// implementation. Since we can't be specific, we use a generic
		// template read error. Callers should know more about the filesystem
		// and be able to identify the specific error from [TemplateError.Err].
		return &TemplateError{
			Path: entryPath,
			Err:  err,
		}
	}

	// Process the file as a template if it has a .tmpl extension
	if strings.HasSuffix(entryPath, ".tmpl") {
		outPath = strings.TrimSuffix(outPath, ".tmpl")
		templateContents, err = executeTemplate(entryPath, outPath, templateContents, options.Data)

		if err != nil {
			return err
		}
	}

	// Write the file to the output location (either processed or as is).
	if err := os.WriteFile(outPath, templateContents, options.FileMode); err != nil {

		// According to the docs, "a failure mid-operation can leave the file
		// in a partially written state", so we try to remove it. We ignore
		// this error since we're working on a temporary directory that will
		// be removed later anyway, and the original error is more important.
		_ = os.Remove(outPath)

		// No need to handle specific errors here, since we're working on a
		// temporary directory that we just created, so permission denied,
		// invalid path, or already exists should not happen. Errors here
		// are more likely to be I/O errors (disk full, quota exceeded, etc).
		return &OutputError{
			Path: outPath,
			Err:  err,
		}
	}

	return nil
}

// Walks the template directory in the given filesystem ifs, processing each
// entry (file or directory) and writing the results to the output location
// specified in options.Location.
func walkTemplate(ifs fs.FS, options *TemplateOptions) error {

	// Walk the template directory
	err := fs.WalkDir(ifs, options.Template, func(entryPath string, entry fs.DirEntry, err error) error {

		if err != nil {
			return &TemplateError{
				Path: entryPath,
				Err:  err,
			}
		}

		// The relative path of the current entry inside the template
		// directory. This is used to recreate the directory structure
		// in the output directory. This is already protected against path
		// traversal attacks, since [fs.WalkDir] won't walk outside the
		// root directory.
		relPath, err := filepath.Rel(options.Template, entryPath)
		if err != nil {
			return &TemplateError{
				Path: entryPath,
				Err:  err,
			}
		}

		// Skip the metafile.
		if options.Metafile != "" {

			//  If there's an error, then it's not the metafile, so it's safe
			// to ignore.
			if relMetafile, _ := filepath.Rel(relPath, options.Metafile); relMetafile == "." {
				return nil
			}
		}

		// The path where the entry will be created
		outPath := filepath.Join(options.Location, relPath)

		// Process the entry
		if err := instantiateTemplateEntry(ifs, entryPath, entry, outPath, options); err != nil {
			return err
		}

		return nil
	})

	return err
}

// Processes the template located in the given filesystem [ifs], passing the
// data in Data to the template engine.
//
// All files and directories are created with the permissions in
// TemplateOptions.FileMode and TemplateOptions.DirMode, with the exception of
// TemplateOptions.Metafile, which is skipped during processing.
//
// If the output location TemplateOptions.Location already exists, Create fails
// without making any changes. Otherwise, it creates the output location and all
// files and directories inside it, mimicking the structure of the template
// directory. The template is first processed in a temporary directory, and only
// moved to the final location if all operations succeed.
//
// Create does not protect against symlinks when writing files and directories
// to the output location. This is deliberate, to allow advanced use cases. The
// caller must ensure that the output location is safe to use. When reading the
// template from the filesystem, Create is protected against path traversal
// only within the given filesystem ifs, so ifs should not encompass any
// important files or directories.
//
// Template injection is also possible, so the caller must ensure that the
// template source is trusted and any TemplateOptions.Data passed to the
// template engine is properly sanitized.
func Create[T any](ifs fs.FS, options *TemplateOptions) error {

	// Clean and prepare output path
	outPath := filepath.Join(options.Location, options.Resource)
	outPath = filepath.Clean(outPath)

	// If the output path exists (file or directory), we don't even try.
	if _, err := os.Lstat(outPath); err == nil {
		return &OutputError{
			Path: outPath,
			Err:  fs.ErrExist,
		}
	}

	// If the template directory doesn't exist, we don't even try.
	if _, err := fs.Lstat(ifs, options.Template); err != nil {
		return &TemplateError{
			Path: options.Template,
			Err:  err,
		}
	}

	// Try creating the output directory ahead of template processing. There's
	// no point in continuing if this fails.
	if err := os.MkdirAll(options.Location, options.DirMode); err != nil {
		return &OutputError{
			Path: options.Location,
			Err:  err,
		}
	}

	// Process the template in a temporary directory first, to avoid leaving
	// a half-baked output in case of errors.
	tmpDir, err := os.MkdirTemp("", "template-*")
	if err != nil {
		return &OutputError{
			Path: options.Location,
			Err:  err,
		}
	}

	// We use a copy of the options, but with the temporary directory
	// as the output location.
	tmpOptions := TemplateOptions{
		Resource: options.Resource,
		Template: options.Template,
		Metafile: options.Metafile,
		Location: tmpDir, // Replace with temporary directory
		FileMode: options.FileMode,
		DirMode:  options.DirMode,
		Data:     options.Data,
	}

	// Process the template into the temporary directory
	err = walkTemplate(ifs, &tmpOptions)

	if err == nil {

		// Move the temporary directory to the final destination. If tmpDir and
		// outPath are on different filesystems, an actual copy occurs (which
		// is not atomic).
		err = os.Rename(tmpDir, outPath)

		if err != nil {

			// Try removing the output directory, in case it was partially created.
			// If we fail because it doesn't exist, all's good; we just warn and
			// continue with the root error.
			if rmErr := os.RemoveAll(outPath); rmErr != nil && !os.IsNotExist(rmErr) {
				slog.Warn("output directory will persist in the filesystem with an incomplete template installation",
					"outPath", outPath,
				)
			}

			// We can know why [os.Rename] failed, but not which file caused
			// the error. We assume it's the output path, since the temporary
			// directory was just created and should be writable. If something
			// happened to the temporary directory in the meantime, this will
			// create confusion for some users, but that should be a rare case.
			err = &OutputError{
				Path: outPath,
				Err:  err,
			}
		}
	}

	// Always remove the temporary directory, regardless of success or failure.
	// If we failed to remove it, we warn but continue with the root error.
	if os.RemoveAll(tmpDir) != nil {
		slog.Warn("failed to remove temporary directory used for template processing",
			"tempDir", tmpDir,
		)
	}

	return err
}

// Lists all templates available in the given filesystem ifs.
//
// It looks for subdirectories in the root of the filesystem, and expects each
// template to be in its own subdirectory, with the metadata file metafile
// inside it. The metadata file is parsed using the given decoder, and the
// parsed data is stored in the Template.Metadata field. If any error is
// encountered while processing a template, Template.Error is set and
// Template.Metadata remains its zero value. If the metafile is an empty string,
// no metadata reading is performed, and the decoder is never called. The
// function returns a slice of all templates found, including those with errors.
func List[T any](ifs fs.FS, metafile string, decoder Decoder[T]) ([]*Template[T], error) {

	// Aggregate results
	results := []*Template[T]{}

	// List entries in FS
	entries, err := fs.ReadDir(ifs, ".")

	if err != nil {
		return nil, &TemplateError{
			Path: ".",
			Err:  err,
		}
	}

	for _, e := range entries {

		if e.IsDir() {

			// Each template is expected to be in its own subdirectory, with
			// the metadata file inside it.
			template := &Template[T]{
				Name:     e.Name(),
				Metafile: metafile,
			}

			results = append(results, template)

			// Skip metadata reading if the metafile is empty
			if template.Metafile == "" {
				continue
			}

			// Set the metafile path
			template.Metafile = filepath.Join(e.Name(), metafile)
			template.Metafile = filepath.Clean(template.Metafile)

			// Read the template metadata and decode its contents
			if data, err := fs.ReadFile(ifs, template.Metafile); err != nil {

				// Here we just flag I/O errors. The caller decides what to do.
				return nil, &TemplateError{
					Path: template.Name,
					Err:  err,
				}

			} else if err := decoder(data, &template.Metadata); err != nil {

				// The error depends on the decoder used (YAML, JSON, etc).
				// Since this error is defined by the caller, we just propagate.
				return nil, err
			}
		}
	}

	return results, nil
}

// Processes the template file located at path in the given filesystem ifs,
// passing the data in data to the template engine.
//
// The rendered output is written to the provided writer w. If any error is
// encountered during parsing or execution, a TemplateError is returned.
func Render[T any](ifs fs.FS, path string, w io.Writer, data T) error {

	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// Parse the template file
	tmpl, err := template.New(filepath.Base(path)).
		Funcs(templateFuncs()).
		ParseFS(ifs, path)
	if err != nil {
		return &TemplateError{
			Path: path,
			Err:  err,
		}
	}

	// Render the template
	if err := tmpl.Execute(w, data); err != nil {
		return errorFromExecuteError(err, path)
	}

	return nil
}

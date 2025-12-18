package types

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Encodes a value to a file.
//
// The content type is inferred from the file extension. The key parameter
// specifies the struct tag to use for field mapping. The v parameter is the
// value to be encoded. The path parameter is the file path to write to.
func EncodeFile(path string, key string, v any) error {
	ct, err := contentTypeFromExtension(path)
	if err != nil {
		return err
	}

	data, err := Encode(ct, key, v)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("%w: %v", ErrEncodingFailed, err)
	}

	return nil
}

// Decodes a file into the target.
//
// The content type is inferred from the file extension. The key parameter
// specifies the struct tag to use for field mapping. The target parameter is a
// pointer to the structure where the decoded data should be stored. The path
// parameter is the file path to read from.
func DecodeFile(path string, key string, target any) (ContentType, error) {
	ct, err := contentTypeFromExtension(path)
	if err != nil {
		return ContentTypeUnknown, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ContentTypeUnknown, fmt.Errorf("%w: %v", ErrDecodingFailed, err)
	}

	return ct, Decode(ct, key, target, data)
}

// Returns the content type for the file extension.
func contentTypeFromExtension(path string) (ContentType, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return ContentTypeJSON, nil
	case ".yaml", ".yml":
		return ContentTypeYAML, nil
	case ".toml":
		return ContentTypeTOML, nil
	default:
		return ContentTypeUnknown, fmt.Errorf("%w: unknown extension %q", ErrUnsupportedContentType, ext)
	}
}

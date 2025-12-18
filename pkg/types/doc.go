// Package types defines shared data structures and utilities.
//
// Content types provide serialization via [ContentType], supporting JSON,
// YAML, and TOML formats. Use [ParseContentType] to parse MIME type strings
// into [ContentType] values. The [Encode] and [Decode] functions work with
// byte slices. The [EncodeFile] and [DecodeFile] functions handle file I/O
// and infer the content type from the file extension.
//
// Examples:
//
//	type Config struct {
//	    Name    string `key:"name"`
//	    Version int    `key:"version"`
//	}
//
//	// Encode to bytes.
//	cfg := Config{Name: "app", Version: 1}
//	data, err := types.Encode(types.ContentTypeJSON, "key", cfg)
//
//	// Decode from bytes.
//	var decoded Config
//	err = types.Decode(types.ContentTypeJSON, "key", &decoded, data)
//
//	// Write to file (content type inferred from extension).
//	err = types.EncodeFile("config.yaml", "key", cfg)
//
//	// Read from file (content type inferred from extension).
//	ct, err := types.DecodeFile("config.yaml", "key", &decoded)
package types

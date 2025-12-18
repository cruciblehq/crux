package types

import (
	"fmt"
	"mime"
)

// Represents a content serialization type.
//
// ContentType defines the serialization format used for encoding and decoding
// data structures. Supported content types include [ContentTypeJSON],
// [ContentTypeYAML], and [ContentTypeTOML].
type ContentType int

const (
	ContentTypeUnknown ContentType = iota // Unknown
	ContentTypeJSON                       // application/json
	ContentTypeYAML                       // application/yaml
	ContentTypeTOML                       // application/toml
)

const (
	contentTypeApplicationJSON = "application/json"
	contentTypeApplicationYAML = "application/yaml"
	contentTypeApplicationTOML = "application/toml"

	contentTypeTextJSON = "text/json"
	contentTypeTextYAML = "text/yaml"
	contentTypeTextTOML = "text/toml"

	contentTypeApplicationXYAML = "application/x-yaml"
	contentTypeApplicationXTOML = "application/x-toml"
)

// Returns the MIME type for the format.
//
// This method returns the standard MIME type string associated with the
// content type, in the form "type/subtype".
func (c ContentType) MIMEType() string {
	switch c {
	case ContentTypeJSON:
		return contentTypeApplicationJSON
	case ContentTypeYAML:
		return contentTypeApplicationYAML
	case ContentTypeTOML:
		return contentTypeApplicationTOML
	default:
		return ""
	}
}

// Returns a human-readable name for the content type.
func (c ContentType) String() string {
	switch c {
	case ContentTypeJSON:
		return "json"
	case ContentTypeYAML:
		return "yaml"
	case ContentTypeTOML:
		return "toml"
	default:
		return "unknown"
	}
}

// Parses a media type string into a ContentType.
//
// The function recognizes standard media types for JSON, YAML, and TOML. If
// the provided media type is not supported, an error is returned.
func ParseContentType(mediaType string) (ContentType, error) {
	parsed, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return ContentTypeUnknown, fmt.Errorf("%w: %s", ErrInvalidContentType, mediaType)
	}

	switch parsed {
	case contentTypeApplicationJSON, contentTypeTextJSON:
		return ContentTypeJSON, nil
	case contentTypeApplicationYAML, contentTypeApplicationXYAML, contentTypeTextYAML:
		return ContentTypeYAML, nil
	case contentTypeApplicationTOML, contentTypeApplicationXTOML, contentTypeTextTOML:
		return ContentTypeTOML, nil
	default:
		return ContentTypeUnknown, fmt.Errorf("%w: %s", ErrUnsupportedContentType, parsed)
	}
}

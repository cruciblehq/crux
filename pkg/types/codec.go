package types

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"
)

// Encodes a value to the specified format.
//
// The contentType specifies the output format. The key parameter specifies the
// struct tag to use for field mapping. The v parameter is the value to be
// encoded. The function returns the encoded data as a byte slice or an error
// if encoding fails or if the content type is unsupported.
func Encode(contentType ContentType, key string, v any) ([]byte, error) {

	// Struct to map
	var raw map[string]any
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &raw,
		TagName: key,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncodingFailed, err)
	}

	// Decode
	if err := decoder.Decode(v); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncodingFailed, err)
	}

	// Map to the target format
	switch contentType {
	case ContentTypeJSON:
		return encodeJSON(raw)
	case ContentTypeYAML:
		return encodeYAML(raw)
	case ContentTypeTOML:
		return encodeTOML(raw)
	default:
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedContentType, contentType)
	}
}

func encodeJSON(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncodingFailed, err)
	}
	return data, nil
}

func encodeYAML(v any) ([]byte, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncodingFailed, err)
	}
	return data, nil
}

func encodeTOML(v any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(v); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncodingFailed, err)
	}
	return buf.Bytes(), nil
}

// Decodes data in the specified format into the target.
//
// The contentType specifies the format of the input data. The key parameter
// specifies the struct tag to use for field mapping. The target parameter is a
// pointer to the structure where the decoded data should be stored. The data
// parameter is the raw bytes to decode.
func Decode(contentType ContentType, key string, target any, data []byte) error {
	var raw map[string]any

	switch contentType {
	case ContentTypeJSON:
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("%w: %v", ErrDecodingFailed, err)
		}
	case ContentTypeYAML:
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("%w: %v", ErrDecodingFailed, err)
		}
	case ContentTypeTOML:
		if err := toml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("%w: %v", ErrDecodingFailed, err)
		}
	default:
		return fmt.Errorf("%w: %v", ErrUnsupportedContentType, contentType)
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  target,
		TagName: key,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDecodingFailed, err)
	}

	if err := decoder.Decode(raw); err != nil {
		return fmt.Errorf("%w: %v", ErrDecodingFailed, err)
	}

	return nil
}

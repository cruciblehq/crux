package oci

import (
	"github.com/cruciblehq/crux/kit/crex"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// Wraps a v1.Image.
//
// Image provides digest computation and layer inspection without exposing
// go-containerregistry types. Obtain an Image from [Index.LoadImage] to inspect
// a specific platform, or from [Builder.Image] after building.
type Image struct {
	img v1.Image // Underlying image.
}

// Contains metadata about an image layer.
//
// Returned by [Image.Layers] to describe each layer in an image without
// exposing the underlying go-containerregistry types or the layer content.
type LayerInfo struct {
	Digest    string // Content digest in "algorithm:hex" format.
	Size      int64  // Compressed size in bytes.
	MediaType string // IANA media type of the layer.
}

// Returns the content digest of the image manifest.
//
// The digest is computed over the serialized manifest and returned in
// "algorithm:hex" format. This value uniquely identifies the image content and
// can be used as an immutable reference in registry URLs or deployment plans.
func (i *Image) Digest() (string, error) {
	h, err := i.img.Digest()
	if err != nil {
		return "", crex.Wrap(ErrInvalidImage, err)
	}
	return h.String(), nil
}

// Returns metadata about each layer in the image.
//
// Each [LayerInfo] includes the layer's content digest, compressed size, and
// media type. Layers are returned in the order they appear in the image
// manifest, from base to topmost.
func (i *Image) Layers() ([]LayerInfo, error) {
	layers, err := i.img.Layers()
	if err != nil {
		return nil, crex.Wrap(ErrInvalidImage, err)
	}

	infos := make([]LayerInfo, 0, len(layers))
	for _, l := range layers {
		digest, err := l.Digest()
		if err != nil {
			return nil, crex.Wrap(ErrInvalidImage, err)
		}
		size, err := l.Size()
		if err != nil {
			return nil, crex.Wrap(ErrInvalidImage, err)
		}
		mt, err := l.MediaType()
		if err != nil {
			return nil, crex.Wrap(ErrInvalidImage, err)
		}
		infos = append(infos, LayerInfo{
			Digest:    digest.String(),
			Size:      size,
			MediaType: string(mt),
		})
	}
	return infos, nil
}

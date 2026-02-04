package build

import "errors"

var (
	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrInvalidResourceType = errors.New("invalid resource type")
	ErrInvalidPath         = errors.New("invalid path")
	ErrImageBuild          = errors.New("image build failed")
	ErrLayerCreate         = errors.New("layer creation failed")
	ErrLayoutWrite         = errors.New("layout write failed")
	ErrInvalidImage        = errors.New("invalid image")
)

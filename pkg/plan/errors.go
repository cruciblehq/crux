package plan

import "errors"

var (
	ErrChannelNotFound       = errors.New("channel not found")
	ErrNoMatchingVersion     = errors.New("no matching version found")
	ErrMissingDigest         = errors.New("version does not have an archive uploaded")
	ErrCannotListVersions    = errors.New("cannot list versions")
	ErrCannotReadVersion     = errors.New("cannot read version")
	ErrCannotCreateReference = errors.New("cannot create frozen reference")
	ErrInvalidConfiguration  = errors.New("invalid configuration")
)

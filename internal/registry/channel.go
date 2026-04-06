package registry

import "github.com/cruciblehq/crex"

// Mutable properties of a channel for creation or update.
//
// Used as the request body for channel creation and update operations. For
// update requests, the name field must match the URL path parameter or update
// context. The version field is a simple string reference to an existing version;
// changing this pointer updates where the channel points. Contains only user-
// modifiable fields. The media type is [MediaTypeChannelInfo].
type ChannelInfo struct {
	Name        string `codec:"name"`        // Channel name.
	Version     string `codec:"version"`     // Version this channel points to.
	Description string `codec:"description"` // Description.
}

// Validates the channel info.
//
// The name must conform to the registry naming rules (see [ValidateName]) and
// the version must be a valid semantic version (see [ValidateVersionString]).
func (info *ChannelInfo) Validate() error {
	if err := ValidateName(info.Name); err != nil {
		return crex.Wrap(ErrInvalidChannel, err)
	}
	if err := ValidateVersionString(info.Version); err != nil {
		return crex.Wrap(ErrInvalidChannel, err)
	}
	return nil
}

// Lightweight channel representation for listings.
//
// Provides channel metadata with a version string reference. Used in resource
// listings and channel lists to keep payloads compact. Includes read-only
// fields like timestamps.
type ChannelSummary struct {
	Name        string `codec:"name"`        // Channel name.
	Version     string `codec:"version"`     // Version this channel points to.
	Description string `codec:"description"` // Description.
	CreatedAt   int64  `codec:"createdAt"`   // When the channel was created.
	UpdatedAt   int64  `codec:"updatedAt"`   // When the channel was last updated.
}

// Validates the channel summary.
func (s *ChannelSummary) Validate() error {
	if err := ValidateName(s.Name); err != nil {
		return crex.Wrap(ErrInvalidChannel, err)
	}
	if err := ValidateVersionString(s.Version); err != nil {
		return crex.Wrap(ErrInvalidChannel, err)
	}
	if err := ValidateTimestamps(s.CreatedAt, s.UpdatedAt); err != nil {
		return crex.Wrap(ErrInvalidChannel, err)
	}
	return nil
}

// Mutable pointer to a version with complete version details.
//
// Provides a named reference that can be updated to point to different versions
// over time, primarily supporting QA/testing workflows. The embedded Version
// object provides full details about the currently targeted version, including
// archive availability and publication status. Channels enable dynamic version
// references during development but are discouraged for production use where
// explicit version references ensure reproducibility. Includes scoping
// information to identify the channel's location. The media type is
// [MediaTypeChannel].
type Channel struct {
	Namespace   string  `codec:"namespace"`   // Namespace this channel belongs to.
	Resource    string  `codec:"resource"`    // Resource this channel belongs to.
	Name        string  `codec:"name"`        // Channel name.
	Version     Version `codec:"version"`     // Full version object this channel points to.
	Description string  `codec:"description"` // Description.
	CreatedAt   int64   `codec:"createdAt"`   // When the channel was created.
	UpdatedAt   int64   `codec:"updatedAt"`   // When the channel was last updated.
}

// Validates the channel.
func (ch *Channel) Validate() error {
	if err := ValidateName(ch.Namespace); err != nil {
		return crex.Wrap(ErrInvalidChannel, err)
	}
	if err := ValidateName(ch.Resource); err != nil {
		return crex.Wrap(ErrInvalidChannel, err)
	}
	if err := ValidateName(ch.Name); err != nil {
		return crex.Wrap(ErrInvalidChannel, err)
	}
	if err := ValidateTimestamps(ch.CreatedAt, ch.UpdatedAt); err != nil {
		return crex.Wrap(ErrInvalidChannel, err)
	}
	if err := ch.Version.Validate(); err != nil {
		return crex.Wrap(ErrInvalidChannel, err)
	}
	return nil
}

// Collection of channels with their current version targets.
//
// The media type is [MediaTypeChannelList].
type ChannelList struct {

	// List of channels.
	Channels []ChannelSummary `codec:"channels"`
}

// Validates the channel list.
func (l *ChannelList) Validate() error {
	for i := range l.Channels {
		if err := l.Channels[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidChannel, err)
		}
	}
	return nil
}

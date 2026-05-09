package registry

import (
	"errors"
	"testing"
)

func TestChannelInfo_Validate_Valid(t *testing.T) {
	info := ChannelInfo{Name: "stable", Version: "1.0.0"}
	if err := info.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChannelInfo_Validate_Invalid_Name(t *testing.T) {
	info := ChannelInfo{Name: "", Version: "1.0.0"}
	err := info.Validate()
	if !errors.Is(err, ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

func TestChannelInfo_Validate_Invalid_Version(t *testing.T) {
	info := ChannelInfo{Name: "stable", Version: "bad"}
	err := info.Validate()
	if !errors.Is(err, ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

func TestChannelSummary_Validate_Valid(t *testing.T) {
	s := ChannelSummary{
		Name:      "stable",
		Version:   "1.0.0",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	if err := s.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChannelSummary_Validate_Invalid_Name(t *testing.T) {
	s := ChannelSummary{
		Name:      "",
		Version:   "1.0.0",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

func TestChannelSummary_Validate_Invalid_Version(t *testing.T) {
	s := ChannelSummary{
		Name:      "stable",
		Version:   "bad",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

func TestChannelSummary_Validate_Invalid_Timestamps(t *testing.T) {
	s := ChannelSummary{
		Name:      "stable",
		Version:   "1.0.0",
		CreatedAt: 0,
		UpdatedAt: 1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

// validChannelVersion returns a valid Version for use in channel tests.
func validChannelVersion() Version {
	return Version{
		Namespace: "my-namespace",
		Resource:  "my-resource",
		String:    "1.0.0",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
}

func TestChannel_Validate_Valid(t *testing.T) {
	ch := Channel{
		Namespace: "my-namespace",
		Resource:  "my-resource",
		Name:      "stable",
		Version:   validChannelVersion(),
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	if err := ch.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChannel_Validate_Invalid_Namespace(t *testing.T) {
	ch := Channel{
		Namespace: "",
		Resource:  "my-resource",
		Name:      "stable",
		Version:   validChannelVersion(),
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := ch.Validate()
	if !errors.Is(err, ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

func TestChannel_Validate_Invalid_Resource(t *testing.T) {
	ch := Channel{
		Namespace: "my-namespace",
		Resource:  "",
		Name:      "stable",
		Version:   validChannelVersion(),
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := ch.Validate()
	if !errors.Is(err, ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

func TestChannel_Validate_Invalid_Name(t *testing.T) {
	ch := Channel{
		Namespace: "my-namespace",
		Resource:  "my-resource",
		Name:      "",
		Version:   validChannelVersion(),
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := ch.Validate()
	if !errors.Is(err, ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

func TestChannel_Validate_Invalid_Timestamps(t *testing.T) {
	ch := Channel{
		Namespace: "my-namespace",
		Resource:  "my-resource",
		Name:      "stable",
		Version:   validChannelVersion(),
		CreatedAt: 2,
		UpdatedAt: 1,
	}
	err := ch.Validate()
	if !errors.Is(err, ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

func TestChannel_Validate_Invalid_NestedVersion(t *testing.T) {
	v := validChannelVersion()
	v.String = "bad"
	ch := Channel{
		Namespace: "my-namespace",
		Resource:  "my-resource",
		Name:      "stable",
		Version:   v,
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := ch.Validate()
	if !errors.Is(err, ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

func TestChannelList_Validate_Valid_Empty(t *testing.T) {
	l := ChannelList{}
	if err := l.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChannelList_Validate_Valid_NonEmpty(t *testing.T) {
	l := ChannelList{
		Channels: []ChannelSummary{
			{Name: "stable", Version: "1.0.0", CreatedAt: 1, UpdatedAt: 1},
		},
	}
	if err := l.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChannelList_Validate_Invalid_NestedChannel(t *testing.T) {
	l := ChannelList{
		Channels: []ChannelSummary{
			{Name: "", Version: "1.0.0", CreatedAt: 1, UpdatedAt: 1},
		},
	}
	err := l.Validate()
	if !errors.Is(err, ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

package registry

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateName_Empty(t *testing.T) {
	err := ValidateName("")
	if !errors.Is(err, ErrNameEmpty) {
		t.Errorf("expected ErrNameEmpty, got %v", err)
	}
}

func TestValidateName_TooLong(t *testing.T) {
	err := ValidateName(strings.Repeat("a", 64))
	if !errors.Is(err, ErrNameTooLong) {
		t.Errorf("expected ErrNameTooLong, got %v", err)
	}
}

func TestValidateName_Invalid_Uppercase(t *testing.T) {
	err := ValidateName("MyName")
	if !errors.Is(err, ErrNameInvalid) {
		t.Errorf("expected ErrNameInvalid, got %v", err)
	}
}

func TestValidateName_Invalid_Underscore(t *testing.T) {
	err := ValidateName("my_name")
	if !errors.Is(err, ErrNameInvalid) {
		t.Errorf("expected ErrNameInvalid, got %v", err)
	}
}

func TestValidateName_Invalid_StartsWithHyphen(t *testing.T) {
	err := ValidateName("-my-name")
	if !errors.Is(err, ErrNameInvalid) {
		t.Errorf("expected ErrNameInvalid, got %v", err)
	}
}

func TestValidateName_Invalid_EndsWithHyphen(t *testing.T) {
	err := ValidateName("my-name-")
	if !errors.Is(err, ErrNameInvalid) {
		t.Errorf("expected ErrNameInvalid, got %v", err)
	}
}

func TestValidateName_Valid_Single(t *testing.T) {
	if err := ValidateName("a"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateName_Valid_WithHyphens(t *testing.T) {
	if err := ValidateName("my-widget-123"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateName_Valid_MaxLength(t *testing.T) {
	if err := ValidateName(strings.Repeat("a", 63)); err != nil {
		t.Errorf("unexpected error for 63-char name: %v", err)
	}
}

func TestValidateVersionString_Valid(t *testing.T) {
	if err := ValidateVersionString("1.0.0"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateVersionString_Invalid_Empty(t *testing.T) {
	err := ValidateVersionString("")
	if !errors.Is(err, ErrVersionInvalid) {
		t.Errorf("expected ErrVersionInvalid, got %v", err)
	}
}

func TestValidateVersionString_Invalid_Format(t *testing.T) {
	err := ValidateVersionString("not-a-version")
	if !errors.Is(err, ErrVersionInvalid) {
		t.Errorf("expected ErrVersionInvalid, got %v", err)
	}
}

func TestValidateVersionString_Invalid_MajorMinorOnly(t *testing.T) {
	err := ValidateVersionString("1.0")
	if !errors.Is(err, ErrVersionInvalid) {
		t.Errorf("expected ErrVersionInvalid, got %v", err)
	}
}

func TestValidateTimestamps_Valid_Equal(t *testing.T) {
	if err := ValidateTimestamps(1, 1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateTimestamps_Valid_UpdatedAfterCreated(t *testing.T) {
	if err := ValidateTimestamps(1, 2); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateTimestamps_Invalid_ZeroCreated(t *testing.T) {
	err := ValidateTimestamps(0, 1)
	if !errors.Is(err, ErrTimestampInvalid) {
		t.Errorf("expected ErrTimestampInvalid, got %v", err)
	}
}

func TestValidateTimestamps_Invalid_ZeroUpdated(t *testing.T) {
	err := ValidateTimestamps(1, 0)
	if !errors.Is(err, ErrTimestampInvalid) {
		t.Errorf("expected ErrTimestampInvalid, got %v", err)
	}
}

func TestValidateTimestamps_Invalid_UpdatedBeforeCreated(t *testing.T) {
	err := ValidateTimestamps(2, 1)
	if !errors.Is(err, ErrTimestampOrder) {
		t.Errorf("expected ErrTimestampOrder, got %v", err)
	}
}

func TestValidateResourceType_Valid(t *testing.T) {
	if err := ValidateResourceType("widget"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateResourceType_Invalid_Empty(t *testing.T) {
	err := ValidateResourceType("")
	if !errors.Is(err, ErrTypeEmpty) {
		t.Errorf("expected ErrTypeEmpty, got %v", err)
	}
}

func TestValidateResourceType_Invalid_Whitespace(t *testing.T) {
	err := ValidateResourceType("   ")
	if !errors.Is(err, ErrTypeEmpty) {
		t.Errorf("expected ErrTypeEmpty, got %v", err)
	}
}

func TestValidateDigest_Valid(t *testing.T) {
	if err := ValidateDigest("sha256:abc123def456"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateDigest_Invalid_NoColon(t *testing.T) {
	err := ValidateDigest("sha256abc123")
	if !errors.Is(err, ErrDigestInvalid) {
		t.Errorf("expected ErrDigestInvalid, got %v", err)
	}
}

func TestValidateDigest_Invalid_UppercaseHex(t *testing.T) {
	err := ValidateDigest("sha256:ABCDEF")
	if !errors.Is(err, ErrDigestInvalid) {
		t.Errorf("expected ErrDigestInvalid, got %v", err)
	}
}

func TestValidateDigest_Invalid_Empty(t *testing.T) {
	err := ValidateDigest("")
	if !errors.Is(err, ErrDigestInvalid) {
		t.Errorf("expected ErrDigestInvalid, got %v", err)
	}
}

func TestValidateArchiveFields_AllNil(t *testing.T) {
	if err := ValidateArchiveFields(nil, nil, nil); err != nil {
		t.Errorf("unexpected error for all-nil fields: %v", err)
	}
}

func TestValidateArchiveFields_AllSet_Valid(t *testing.T) {
	archive := "https://example.com/archive.tar.zst"
	size := int64(1024)
	digest := "sha256:abc123"
	if err := ValidateArchiveFields(&archive, &size, &digest); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateArchiveFields_Incomplete_OnlyArchive(t *testing.T) {
	archive := "https://example.com/archive.tar.zst"
	err := ValidateArchiveFields(&archive, nil, nil)
	if !errors.Is(err, ErrArchiveIncomplete) {
		t.Errorf("expected ErrArchiveIncomplete, got %v", err)
	}
}

func TestValidateArchiveFields_Incomplete_OnlySize(t *testing.T) {
	size := int64(1024)
	err := ValidateArchiveFields(nil, &size, nil)
	if !errors.Is(err, ErrArchiveIncomplete) {
		t.Errorf("expected ErrArchiveIncomplete, got %v", err)
	}
}

func TestValidateArchiveFields_Incomplete_ArchiveAndSize(t *testing.T) {
	archive := "https://example.com/archive.tar.zst"
	size := int64(1024)
	err := ValidateArchiveFields(&archive, &size, nil)
	if !errors.Is(err, ErrArchiveIncomplete) {
		t.Errorf("expected ErrArchiveIncomplete, got %v", err)
	}
}

func TestValidateArchiveFields_EmptyArchive(t *testing.T) {
	archive := ""
	size := int64(1024)
	digest := "sha256:abc123"
	err := ValidateArchiveFields(&archive, &size, &digest)
	if !errors.Is(err, ErrArchiveEmpty) {
		t.Errorf("expected ErrArchiveEmpty, got %v", err)
	}
}

func TestValidateArchiveFields_NegativeSize(t *testing.T) {
	archive := "https://example.com/archive.tar.zst"
	size := int64(-1)
	digest := "sha256:abc123"
	err := ValidateArchiveFields(&archive, &size, &digest)
	if !errors.Is(err, ErrSizeInvalid) {
		t.Errorf("expected ErrSizeInvalid, got %v", err)
	}
}

func TestValidateArchiveFields_ZeroSize(t *testing.T) {
	archive := "https://example.com/archive.tar.zst"
	size := int64(0)
	digest := "sha256:abc123"
	err := ValidateArchiveFields(&archive, &size, &digest)
	if !errors.Is(err, ErrSizeInvalid) {
		t.Errorf("expected ErrSizeInvalid, got %v", err)
	}
}

func TestValidateArchiveFields_InvalidDigest(t *testing.T) {
	archive := "https://example.com/archive.tar.zst"
	size := int64(1024)
	digest := "invalid-digest"
	err := ValidateArchiveFields(&archive, &size, &digest)
	if !errors.Is(err, ErrDigestInvalid) {
		t.Errorf("expected ErrDigestInvalid, got %v", err)
	}
}

func TestValidateCount_Zero(t *testing.T) {
	if err := ValidateCount(0); err != nil {
		t.Errorf("unexpected error for zero: %v", err)
	}
}

func TestValidateCount_Positive(t *testing.T) {
	if err := ValidateCount(5); err != nil {
		t.Errorf("unexpected error for positive: %v", err)
	}
}

func TestValidateCount_Negative(t *testing.T) {
	err := ValidateCount(-1)
	if !errors.Is(err, ErrCountNegative) {
		t.Errorf("expected ErrCountNegative, got %v", err)
	}
}

func TestValidateNamespace_Valid(t *testing.T) {
	if err := ValidateNamespace("my-namespace"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateNamespace_Invalid(t *testing.T) {
	err := ValidateNamespace("")
	if !errors.Is(err, ErrNameEmpty) {
		t.Errorf("expected ErrNameEmpty, got %v", err)
	}
}

func TestValidateIdentifier_Valid(t *testing.T) {
	if err := ValidateIdentifier("my-namespace", "my-resource"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateIdentifier_Invalid_Namespace(t *testing.T) {
	err := ValidateIdentifier("", "my-resource")
	if !errors.Is(err, ErrNameEmpty) {
		t.Errorf("expected ErrNameEmpty, got %v", err)
	}
}

func TestValidateIdentifier_Invalid_Resource(t *testing.T) {
	err := ValidateIdentifier("my-namespace", "")
	if !errors.Is(err, ErrNameEmpty) {
		t.Errorf("expected ErrNameEmpty, got %v", err)
	}
}

func TestValidateReference_Valid(t *testing.T) {
	if err := ValidateReference("my-namespace", "my-resource", "1.0.0"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateReference_Invalid_Namespace(t *testing.T) {
	err := ValidateReference("", "my-resource", "1.0.0")
	if !errors.Is(err, ErrNameEmpty) {
		t.Errorf("expected ErrNameEmpty, got %v", err)
	}
}

func TestValidateReference_Invalid_Resource(t *testing.T) {
	err := ValidateReference("my-namespace", "", "1.0.0")
	if !errors.Is(err, ErrNameEmpty) {
		t.Errorf("expected ErrNameEmpty, got %v", err)
	}
}

func TestValidateReference_Invalid_Version(t *testing.T) {
	err := ValidateReference("my-namespace", "my-resource", "not-a-version")
	if !errors.Is(err, ErrVersionInvalid) {
		t.Errorf("expected ErrVersionInvalid, got %v", err)
	}
}

func TestValidateChannelReference_Valid(t *testing.T) {
	if err := ValidateChannelReference("my-namespace", "my-resource", "stable"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateChannelReference_Invalid_Channel(t *testing.T) {
	err := ValidateChannelReference("my-namespace", "my-resource", "")
	if !errors.Is(err, ErrNameEmpty) {
		t.Errorf("expected ErrNameEmpty, got %v", err)
	}
}

func TestValidateChannelInfo_Valid(t *testing.T) {
	info := ChannelInfo{Name: "stable", Version: "1.0.0"}
	if err := ValidateChannelInfo("my-namespace", "my-resource", info); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateChannelInfo_Invalid_ChannelName(t *testing.T) {
	info := ChannelInfo{Name: "", Version: "1.0.0"}
	err := ValidateChannelInfo("my-namespace", "my-resource", info)
	if !errors.Is(err, ErrNameEmpty) {
		t.Errorf("expected ErrNameEmpty, got %v", err)
	}
}

func TestValidateChannelInfo_Invalid_Version(t *testing.T) {
	info := ChannelInfo{Name: "stable", Version: "bad"}
	err := ValidateChannelInfo("my-namespace", "my-resource", info)
	if !errors.Is(err, ErrVersionInvalid) {
		t.Errorf("expected ErrVersionInvalid, got %v", err)
	}
}

package config

import (
	"testing"
)

func TestAWSProfileAuth_Validate_EmptyProfile(t *testing.T) {
	auth := &AWSProfileAuth{
		Profile: "",
	}

	err := auth.Validate()
	if err == nil {
		t.Error("Validate() with empty profile should return error")
	}
}

func TestAWSProfileAuth_Validate_NonExistentProfile(t *testing.T) {
	auth := &AWSProfileAuth{
		Profile: "non-existent-profile-12345",
	}

	err := auth.Validate()
	if err == nil {
		t.Error("Validate() with non-existent profile should return error")
	}
}

func TestAWSKeysAuth_Validate_EmptyCredentials(t *testing.T) {
	tests := []struct {
		name            string
		accessKeyID     string
		secretAccessKey string
	}{
		{"both empty", "", ""},
		{"access key empty", "", "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"secret key empty", "AKIAXXXXXXXXXXXXXXXX", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &AWSKeysAuth{
				AccessKeyID:     tt.accessKeyID,
				SecretAccessKey: tt.secretAccessKey,
			}

			err := auth.Validate()
			if err == nil {
				t.Error("Validate() with empty credentials should return error")
			}
		})
	}
}

func TestAWSKeysAuth_Validate_InvalidCredentials(t *testing.T) {
	auth := &AWSKeysAuth{
		AccessKeyID:     "AKIAXXXXXXXXXXXXXXXX",
		SecretAccessKey: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
	}

	// This should fail because the credentials are invalid (network call will fail)
	err := auth.Validate()
	if err == nil {
		t.Error("Validate() with invalid credentials should return error")
	}
}

// Note: We cannot easily test successful validation without real AWS credentials
// or mocking the AWS SDK, which is complex. The validation logic itself is tested
// by checking error cases. In a real environment with valid credentials, the
// validation would succeed.

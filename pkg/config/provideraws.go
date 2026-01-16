package config

import (
	"regexp"
	"strings"

	"github.com/cruciblehq/crux/pkg/crex"
)

const (
	// AWS Access Key ID length
	AWSAccessKeyIDLength = 20

	// AWS Secret Access Key length
	AWSSecretAccessKeyLength = 40
)

var (
	// AWS Access Key ID pattern (starts with AKIA for IAM users or ASIA for temporary credentials)
	awsAccessKeyIDPattern = regexp.MustCompile(`^(AKIA|ASIA)[A-Z0-9]{16}$`)

	// AWS Secret Access Key pattern (base64-like characters)
	awsSecretAccessKeyPattern = regexp.MustCompile(`^[A-Za-z0-9/+=]{40}$`)
)

// Specifies the authentication method for AWS providers.
type AuthMethod string

const (

	// Uses AWS CLI profiles from ~/.aws/credentials.
	AuthMethodProfile AuthMethod = "profile"

	// Uses explicit AWS access key ID and secret access key.
	AuthMethodKeys AuthMethod = "keys"
)

// AWS-specific provider configuration.
//
// AWS deployments require a region and authentication credentials. Two
// authentication methods are supported: profile-based and access key-based.
//
// Profile-based uses AWS CLI profiles configured in the user's home directory
// (~/.aws/credentials and ~/.aws/config). This method supports AWS SSO, MFA,
// and automatic credential rotation. Profiles must be created beforehand using
// the AWS CLI (e.g., `aws configure --profile <name>`).
//
// Access key-based uses explicit AWS access key ID and secret access key for
// authentication. This method is simpler but less secure, as credentials are
// stored in the crux configuration file. It does not support SSO, MFA, or role
// assumptions.
//
// The AuthMethod field specifies which authentication method to use ("profile"
// or "keys"). The Auth field holds the corresponding authentication details
// as either *[AWSProfileAuth] or *[AWSKeysAuth].
type AWSProvider struct {
	Region     string     `field:"region"`      // AWS region (e.g., us-east-1, eu-west-1)
	AuthMethod AuthMethod `field:"auth_method"` // Authentication method: AuthMethodProfile or AuthMethodKeys
	Auth       any        `field:"auth"`        // *[AWSProfileAuth] or *[AWSKeysAuth] depending on AuthMethod
}

// AWS profile-based authentication.
//
// Uses AWS CLI profile configuration from ~/.aws/credentials and ~/.aws/config.
// The profile must exist before it can be used. The AWS SDK will automatically
// refresh credentials when using profiles.
type AWSProfileAuth struct {
	Profile string `field:"profile"` // AWS profile name from ~/.aws/credentials
}

// AWS access key authentication.
//
// Uses explicit AWS access key ID and secret access key for authentication.
// These credentials are stored in the crux configuration file at
// ~/.config/crux/providers.yaml. This method is simpler than profile-based
// authentication but lacks support for AWS SSO, MFA, and automatic credential
// rotation. It is recommended to use profile-based authentication instead.
type AWSKeysAuth struct {
	AccessKeyID     string `field:"access_key_id"`     // AWS access key ID (format: AKIA...)
	SecretAccessKey string `field:"secret_access_key"` // AWS secret access key
}

// Validates AWS access key credentials format.
//
// Checks that both AccessKeyID and SecretAccessKey are provided and conform
// to AWS credential format requirements.
func (a *AWSKeysAuth) Validate() error {
	// Check for empty values
	if a.AccessKeyID == "" || a.SecretAccessKey == "" {
		return crex.UserError("invalid AWS credentials", "access key ID and secret access key cannot be empty").Err()
	}

	// Validate Access Key ID format
	if len(a.AccessKeyID) != AWSAccessKeyIDLength {
		return crex.UserError("invalid AWS access key ID", "access key ID must be exactly 20 characters long").
			Fallback("Verify your AWS access key ID and try again.").Err()
	}

	if !awsAccessKeyIDPattern.MatchString(strings.TrimSpace(a.AccessKeyID)) {
		return crex.UserError("invalid AWS access key ID", "access key ID format is incorrect (must start with AKIA or ASIA)").
			Fallback("Verify your AWS access key ID and try again.").Err()
	}

	// Validate Secret Access Key format
	if len(a.SecretAccessKey) != AWSSecretAccessKeyLength {
		return crex.UserError("invalid AWS secret access key", "secret access key must be exactly 40 characters long").
			Fallback("Verify your AWS secret access key and try again.").Err()
	}

	if !awsSecretAccessKeyPattern.MatchString(strings.TrimSpace(a.SecretAccessKey)) {
		return crex.UserError("invalid AWS secret access key", "secret access key contains invalid characters").
			Fallback("Verify your AWS secret access key and try again.").Err()
	}

	return nil
}

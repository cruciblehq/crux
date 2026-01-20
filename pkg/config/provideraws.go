package config

import (
	"context"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/cruciblehq/crux/pkg/crex"
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

// Validates that an AWS profile exists and can be loaded.
//
// Uses the AWS SDK to attempt loading the profile configuration. This properly
// handles all AWS credential sources including SSO, role assumptions, credential
// chains, and environment variables.
func (a *AWSProfileAuth) Validate() error {
	if a.Profile == "" {
		return crex.UserError("invalid AWS profile", "profile name cannot be empty").Err()
	}

	// Load the AWS config with the specified profile
	ctx := context.Background()
	_, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithSharedConfigProfile(a.Profile),
	)

	if err != nil {
		return crex.UserErrorf("invalid AWS profile configuration", "failed to load profile %q", a.Profile).
			Fallback("Ensure the profile exists and is properly configured with 'aws configure --profile " + a.Profile + "', or use access keys instead").
			Cause(err).
			Err()
	}

	return nil
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

// Validates AWS access key credentials.
//
// Uses AWS STS GetCallerIdentity to verify that the credentials are valid and
// have the necessary permissions. This is the recommended way to validate AWS
// credentials as it checks actual validity, not just format. This makes a
// network call to AWS and requires sts:GetCallerIdentity permissions.
func (a *AWSKeysAuth) Validate() error {

	// Create AWS config with the provided credentials
	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			a.AccessKeyID,
			a.SecretAccessKey,
			"", // session token (empty for long-term credentials)
		)),
	)
	if err != nil {
		return crex.UserError("invalid AWS credentials", "failed to configure AWS SDK").
			Cause(err).Err()
	}

	// Validate credentials by calling STS GetCallerIdentity
	stsClient := sts.NewFromConfig(cfg)
	_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return crex.UserError("invalid AWS credentials", "credentials failed AWS validation").
			Fallback("Verify your AWS access key ID and secret access key and try again.").
			Cause(err).
			Err()
	}

	return nil
}

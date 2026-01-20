package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cruciblehq/protocol/pkg/codec"
)

func TestProvidersConfig_AddProvider_FirstProviderBecomesDefault(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}

	provider1 := Provider{
		Type:   ProviderTypeAWS,
		Config: &AWSProvider{Region: "us-east-1"},
	}
	config.AddProvider("aws-prod", provider1)

	if config.Default != "aws-prod" {
		t.Errorf("Default = %v, want aws-prod", config.Default)
	}
	if len(config.Providers) != 1 {
		t.Errorf("len(Providers) = %d, want 1", len(config.Providers))
	}
	if config.Providers["aws-prod"].Name != "aws-prod" {
		t.Errorf("Provider name = %v, want aws-prod", config.Providers["aws-prod"].Name)
	}
}

func TestProvidersConfig_AddProvider_SecondProviderDoesNotChangeDefault(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}

	provider1 := Provider{
		Type:   ProviderTypeAWS,
		Config: &AWSProvider{Region: "us-east-1"},
	}
	config.AddProvider("aws-prod", provider1)

	provider2 := Provider{
		Type:   ProviderTypeLocal,
		Config: &LocalProvider{},
	}
	config.AddProvider("local-dev", provider2)

	if config.Default != "aws-prod" {
		t.Errorf("Default = %v, want aws-prod (unchanged)", config.Default)
	}
	if len(config.Providers) != 2 {
		t.Errorf("len(Providers) = %d, want 2", len(config.Providers))
	}
}

func TestProvidersConfig_AddProvider_DuplicateName(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}

	provider1 := Provider{
		Type:   ProviderTypeAWS,
		Config: &AWSProvider{Region: "us-east-1"},
	}
	err := config.AddProvider("aws-prod", provider1)
	if err != nil {
		t.Fatalf("AddProvider() error = %v, want nil", err)
	}

	provider2 := Provider{
		Type:   ProviderTypeAWS,
		Config: &AWSProvider{Region: "eu-west-1"},
	}
	err = config.AddProvider("aws-prod", provider2)
	if !errors.Is(err, ErrProviderAlreadyExists) {
		t.Errorf("AddProvider() error = %v, want ErrProviderAlreadyExists", err)
	}

	// Verify original provider was not replaced
	if len(config.Providers) != 1 {
		t.Errorf("len(Providers) = %d, want 1", len(config.Providers))
	}
	awsProvider := config.Providers["aws-prod"].Config.(*AWSProvider)
	if awsProvider.Region != "us-east-1" {
		t.Errorf("Provider region = %v, want us-east-1 (unchanged)", awsProvider.Region)
	}
}

func TestValidateProviderName(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		wantError bool
	}{
		{"valid alphanumeric", "aws123", false},
		{"valid with hyphen", "aws-prod", false},
		{"valid with underscore", "aws_dev", false},
		{"valid mixed", "aws-prod_1", false},
		{"empty", "", true},
		{"starts with hyphen", "-aws", true},
		{"starts with underscore", "_aws", true},
		{"contains space", "aws prod", true},
		{"contains at symbol", "user@aws", true},
		{"too long", "aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffff12345", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProviderName(tt.provider)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateProviderName(%q) error = %v, wantError %v", tt.provider, err, tt.wantError)
			}
		})
	}
}

func TestProvidersConfig_AddProvider_InvalidName(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}

	provider := Provider{
		Type:   ProviderTypeAWS,
		Config: &AWSProvider{Region: "us-east-1"},
	}
	err := config.AddProvider("-invalid", provider)
	if err == nil {
		t.Error("AddProvider() with invalid name should return error")
	}
}

func TestProvidersConfig_RemoveProvider_NonExistent(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}

	err := config.RemoveProvider("non-existent")
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("RemoveProvider() error = %v, want ErrProviderNotFound", err)
	}
	if config.Default != "" {
		t.Errorf("Default = %v, want empty string", config.Default)
	}
	if len(config.Providers) != 0 {
		t.Errorf("len(Providers) = %d, want 0", len(config.Providers))
	}
}

func TestProvidersConfig_RemoveProvider_Default(t *testing.T) {
	config := &ProvidersConfig{
		Default:   "aws-prod",
		Providers: make(map[string]Provider),
	}
	config.Providers["aws-prod"] = Provider{
		Type:   ProviderTypeAWS,
		Name:   "aws-prod",
		Config: &AWSProvider{},
	}

	err := config.RemoveProvider("aws-prod")
	if err != nil {
		t.Errorf("RemoveProvider() error = %v, want nil", err)
	}
	if config.Default != "" {
		t.Errorf("Default = %v, want empty string", config.Default)
	}
	if len(config.Providers) != 0 {
		t.Errorf("len(Providers) = %d, want 0", len(config.Providers))
	}
}

func TestProvidersConfig_RemoveProvider_NonDefault(t *testing.T) {
	config := &ProvidersConfig{
		Default:   "aws-prod",
		Providers: make(map[string]Provider),
	}
	config.Providers["aws-prod"] = Provider{Type: ProviderTypeAWS, Name: "aws-prod"}
	config.Providers["local-dev"] = Provider{Type: ProviderTypeLocal, Name: "local-dev"}

	err := config.RemoveProvider("local-dev")
	if err != nil {
		t.Errorf("RemoveProvider() error = %v, want nil", err)
	}
	if config.Default != "aws-prod" {
		t.Errorf("Default = %v, want aws-prod", config.Default)
	}
	if len(config.Providers) != 1 {
		t.Errorf("len(Providers) = %d, want 1", len(config.Providers))
	}
}

func TestProvidersConfig_GetProvider_Existing(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}
	config.Providers["aws-prod"] = Provider{
		Type:   ProviderTypeAWS,
		Name:   "aws-prod",
		Config: &AWSProvider{Region: "us-east-1"},
	}

	provider, err := config.GetProvider("aws-prod")
	if err != nil {
		t.Fatalf("GetProvider() error = %v, want nil", err)
	}
	if provider.Name != "aws-prod" {
		t.Errorf("Provider name = %v, want aws-prod", provider.Name)
	}
}

func TestProvidersConfig_GetProvider_NonExistent(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}

	_, err := config.GetProvider("non-existent")
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("GetProvider() error = %v, want ErrProviderNotFound", err)
	}
}

func TestProvidersConfig_GetDefault_ExistingDefault(t *testing.T) {
	config := &ProvidersConfig{
		Default:   "aws-prod",
		Providers: make(map[string]Provider),
	}
	config.Providers["aws-prod"] = Provider{
		Type: ProviderTypeAWS,
		Name: "aws-prod",
	}

	provider, err := config.GetDefault()
	if err != nil {
		t.Errorf("GetDefault() error = %v, want nil", err)
	}
	if provider.Name != "aws-prod" {
		t.Errorf("Provider name = %v, want aws-prod", provider.Name)
	}
}

func TestProvidersConfig_GetDefault_NoDefaultSet(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}

	_, err := config.GetDefault()
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("GetDefault() error = %v, want ErrProviderNotFound", err)
	}
}

func TestProvidersConfig_GetDefault_DefaultDoesNotExist(t *testing.T) {
	config := &ProvidersConfig{
		Default:   "non-existent",
		Providers: make(map[string]Provider),
	}

	_, err := config.GetDefault()
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("GetDefault() error = %v, want ErrProviderNotFound", err)
	}
}

func TestProvidersConfig_SetDefault_ExistingProvider(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}
	config.Providers["aws-prod"] = Provider{Type: ProviderTypeAWS, Name: "aws-prod"}
	config.Providers["local-dev"] = Provider{Type: ProviderTypeLocal, Name: "local-dev"}

	err := config.SetDefault("local-dev")
	if err != nil {
		t.Fatalf("SetDefault() error = %v, want nil", err)
	}
	if config.Default != "local-dev" {
		t.Errorf("Default = %v, want local-dev", config.Default)
	}
}

func TestProvidersConfig_SetDefault_NonExistent(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}

	err := config.SetDefault("non-existent")
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("SetDefault() error = %v, want ErrProviderNotFound", err)
	}
}

func TestProvidersConfig_GetOrDefault_ByName(t *testing.T) {
	config := &ProvidersConfig{
		Default:   "aws-prod",
		Providers: make(map[string]Provider),
	}
	config.Providers["aws-prod"] = Provider{Type: ProviderTypeAWS, Name: "aws-prod"}
	config.Providers["local-dev"] = Provider{Type: ProviderTypeLocal, Name: "local-dev"}

	provider, err := config.GetOrDefault("local-dev")
	if err != nil {
		t.Fatalf("GetOrDefault() error = %v, want nil", err)
	}
	if provider.Name != "local-dev" {
		t.Errorf("Provider name = %v, want local-dev", provider.Name)
	}
}

func TestProvidersConfig_GetOrDefault_EmptyNameReturnsDefault(t *testing.T) {
	config := &ProvidersConfig{
		Default:   "aws-prod",
		Providers: make(map[string]Provider),
	}
	config.Providers["aws-prod"] = Provider{Type: ProviderTypeAWS, Name: "aws-prod"}
	config.Providers["local-dev"] = Provider{Type: ProviderTypeLocal, Name: "local-dev"}

	provider, err := config.GetOrDefault("")
	if err != nil {
		t.Fatalf("GetOrDefault() error = %v, want nil", err)
	}
	if provider.Name != "aws-prod" {
		t.Errorf("Provider name = %v, want aws-prod", provider.Name)
	}
}

func TestProvidersConfig_GetOrDefault_NonExistent(t *testing.T) {
	config := &ProvidersConfig{
		Default:   "aws-prod",
		Providers: make(map[string]Provider),
	}
	config.Providers["aws-prod"] = Provider{Type: ProviderTypeAWS, Name: "aws-prod"}

	_, err := config.GetOrDefault("non-existent")
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("GetOrDefault() error = %v, want ErrProviderNotFound", err)
	}
}

func TestProvidersConfig_ListProviders_Empty(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}

	providers := config.ListProviders()
	if len(providers) != 0 {
		t.Errorf("len(ListProviders()) = %d, want 0", len(providers))
	}
}

func TestProvidersConfig_ListProviders_Multiple(t *testing.T) {
	config := &ProvidersConfig{
		Providers: make(map[string]Provider),
	}
	config.Providers["aws-prod"] = Provider{Type: ProviderTypeAWS, Name: "aws-prod"}
	config.Providers["local-dev"] = Provider{Type: ProviderTypeLocal, Name: "local-dev"}

	providers := config.ListProviders()
	if len(providers) != 2 {
		t.Errorf("len(ListProviders()) = %d, want 2", len(providers))
	}

	names := make(map[string]bool)
	for _, p := range providers {
		names[p.Name] = true
	}
	if !names["aws-prod"] || !names["local-dev"] {
		t.Errorf("ListProviders() missing expected providers")
	}
}

func TestProvidersConfig_SaveLoad(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test-providers.yaml")

	original := &ProvidersConfig{
		Default:   "local-dev",
		Providers: make(map[string]Provider),
	}
	original.Providers["local-dev"] = Provider{
		Type:   ProviderTypeLocal,
		Name:   "local-dev",
		Config: &LocalProvider{},
	}

	if err := os.MkdirAll(filepath.Dir(testPath), 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	if err := codec.EncodeFile(testPath, "field", original); err != nil {
		t.Fatalf("Failed to encode config: %v", err)
	}

	var loaded ProvidersConfig
	if _, err := codec.DecodeFile(testPath, "field", &loaded); err != nil {
		t.Fatalf("Failed to decode config: %v", err)
	}

	if loaded.Default != "local-dev" {
		t.Errorf("Default = %v, want local-dev", loaded.Default)
	}
	if len(loaded.Providers) != 1 {
		t.Errorf("len(Providers) = %d, want 1", len(loaded.Providers))
	}

	provider := loaded.Providers["local-dev"]
	if provider.Type != ProviderTypeLocal {
		t.Errorf("Provider type = %v, want %v", provider.Type, ProviderTypeLocal)
	}
	if provider.Name != "local-dev" {
		t.Errorf("Provider name = %v, want local-dev", provider.Name)
	}

	if provider.Config == nil {
		t.Error("Provider config is nil")
	}
}

package registry

import (
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid names
		{"valid simple", "hello", false},
		{"valid with numbers", "hello123", false},
		{"valid with hyphens", "hello-world", false},
		{"valid with numbers and hyphens", "test-123-app", false},
		{"valid 63 characters", "a12345678901234567890123456789012345678901234567890123456789012", false},
		{"valid single character", "a", false},
		{"valid two characters", "ab", false},

		// Invalid names
		{"empty string", "", true},
		{"too long", "a1234567890123456789012345678901234567890123456789012345678901234", true},
		{"starts with hyphen", "-hello", true},
		{"ends with hyphen", "hello-", true},
		{"uppercase letters", "Hello", true},
		{"uppercase in middle", "heLLo", true},
		{"special characters", "hello@world", true},
		{"underscore", "hello_world", true},
		{"dot", "hello.world", true},
		{"space", "hello world", true},
		{"starts with number followed by hyphen", "1-hello", false},
		{"only hyphen", "-", true},
		{"only number", "1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateVersionString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid versions
		{"simple version", "1.2.3", false},
		{"zero version", "0.0.0", false},
		{"large numbers", "100.200.300", false},
		{"with prerelease and numbers", "1.0.0-alpha.1", false},
		{"with build metadata", "1.0.0+20230101", false},
		{"with both", "1.0.0-beta.1+exp.sha.5114f85", false},

		// Invalid versions
		{"empty string", "", true},
		{"only major", "1", true},
		{"only major.minor", "1.2", true},
		{"with spaces", "1 .2.3", true},
		{"non-numeric", "a.b.c", true},
		{"negative numbers", "-1.2.3", true},
		{"trailing dot", "1.2.3.", true},
		{"leading dot", ".1.2.3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVersionString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVersionString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantErr   bool
	}{
		{"valid namespace", "my-namespace", false},
		{"invalid namespace", "My-Namespace", true},
		{"empty namespace", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNamespace(tt.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNamespace(%q) error = %v, wantErr %v", tt.namespace, err, tt.wantErr)
			}
		})
	}
}

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		resource  string
		wantErr   bool
	}{
		{"valid identifier", "my-namespace", "my-resource", false},
		{"invalid namespace", "My-Namespace", "my-resource", true},
		{"invalid resource", "my-namespace", "My-Resource", true},
		{"both invalid", "My-Namespace", "My-Resource", true},
		{"empty namespace", "", "my-resource", true},
		{"empty resource", "my-namespace", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIdentifier(tt.namespace, tt.resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIdentifier(%q, %q) error = %v, wantErr %v", tt.namespace, tt.resource, err, tt.wantErr)
			}
		})
	}
}

func TestValidateReference(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		resource  string
		version   string
		wantErr   bool
	}{
		{"valid reference", "my-namespace", "my-resource", "1.2.3", false},
		{"invalid namespace", "My-Namespace", "my-resource", "1.2.3", true},
		{"invalid resource", "my-namespace", "My-Resource", "1.2.3", true},
		{"all invalid", "My-Namespace", "My-Resource", "1.2.3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateReference(tt.namespace, tt.resource, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateReference(%q, %q, %q) error = %v, wantErr %v", tt.namespace, tt.resource, tt.version, err, tt.wantErr)
			}
		})
	}
}

func TestValidateChannelReference(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		resource  string
		channel   string
		wantErr   bool
	}{
		{"valid channel reference", "my-namespace", "my-resource", "stable", false},
		{"invalid namespace", "My-Namespace", "my-resource", "stable", true},
		{"invalid resource", "my-namespace", "My-Resource", "stable", true},
		{"invalid channel", "my-namespace", "my-resource", "Stable", true},
		{"all invalid", "My-Namespace", "My-Resource", "Stable", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChannelReference(tt.namespace, tt.resource, tt.channel)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateChannelReference(%q, %q, %q) error = %v, wantErr %v", tt.namespace, tt.resource, tt.channel, err, tt.wantErr)
			}
		})
	}
}

func TestValidateChannelInfo(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		resource  string
		info      ChannelInfo
		wantErr   bool
	}{
		{
			name:      "valid channel info",
			namespace: "my-namespace",
			resource:  "my-resource",
			info:      ChannelInfo{Name: "stable", Version: "1.2.3"},
			wantErr:   false,
		},
		{
			name:      "invalid namespace",
			namespace: "My-Namespace",
			resource:  "my-resource",
			info:      ChannelInfo{Name: "stable", Version: "1.2.3"},
			wantErr:   true,
		},
		{
			name:      "invalid resource",
			namespace: "my-namespace",
			resource:  "My-Resource",
			info:      ChannelInfo{Name: "stable", Version: "1.2.3"},
			wantErr:   true,
		},
		{
			name:      "invalid channel name",
			namespace: "my-namespace",
			resource:  "my-resource",
			info:      ChannelInfo{Name: "Stable", Version: "1.2.3"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChannelInfo(tt.namespace, tt.resource, tt.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateChannelInfo(%q, %q, %+v) error = %v, wantErr %v", tt.namespace, tt.resource, tt.info, err, tt.wantErr)
			}
		})
	}
}

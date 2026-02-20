// Package config provides configuration management for cloud providers.
//
// The package supports multiple cloud provider types (AWS, local) with
// type-safe configuration structures. Provider configurations are stored
// in the user's config directory and include authentication credentials
// and deployment settings. Provider names must start with an alphanumeric
// character and may contain alphanumeric characters, hyphens, or underscores.
//
// Loading and listing all configured providers:
//
//	cfg, err := config.LoadProviders()
//	if err != nil {
//	    return err
//	}
//	for _, p := range cfg.ListProviders() {
//	    fmt.Println(p.Name, p.Type)
//	}
//
// Adding a new provider and setting it as default:
//
//	cfg, err := config.LoadProviders()
//	if err != nil {
//	    return err
//	}
//	err = cfg.AddProvider("staging", config.Provider{
//	    Type:   config.ProviderTypeAWS,
//	    Name:   "staging",
//	    Config: &config.AWSProvider{Region: "us-west-2"},
//	})
//	if err != nil {
//	    return err
//	}
//	cfg.SetDefault("staging")
//	cfg.Save()
package config

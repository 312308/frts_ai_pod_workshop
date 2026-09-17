// Package config loads FOP integration settings for tcg-service.
//
// Source: BRD FR-024 ("Acquire OAuth2 JWT token from TCG per environment...
// partnerId = environment config (never hard-coded)... Secret in Key Vault
// per environment.").
package config

import (
	"fmt"
	"os"
)

// Config holds FOP client settings. Values come from environment variables
// at process start; production deployments back these with Azure Key Vault
// references injected as env vars by the platform — never hardcoded, never
// logged (BRD FR-024, rule 06-backend-golang.mdc).
type Config struct {
	AADTenantID     string
	AADClientID     string
	AADClientSecret string
	AADScope        string
	PartnerID       string
	FOPBaseURL      string

	// AADTokenURL overrides the derived Azure AD v2.0 token endpoint.
	// Optional; empty in production (derived from AADTenantID). Exists as a
	// test seam so unit tests can point token acquisition at an httptest
	// server instead of the real Azure AD endpoint.
	AADTokenURL string
}

// Load reads FOP client configuration from environment variables and fails
// fast if any required field is missing (TC-006: the client must never be
// constructed in a state that could reach FOP without partnerId).
func Load() (*Config, error) {
	cfg := &Config{
		AADTenantID:     os.Getenv("FOP_AAD_TENANT_ID"),
		AADClientID:     os.Getenv("FOP_AAD_CLIENT_ID"),
		AADClientSecret: os.Getenv("FOP_AAD_CLIENT_SECRET"),
		AADScope:        os.Getenv("FOP_AAD_SCOPE"),
		PartnerID:       os.Getenv("FOP_PARTNER_ID"),
		FOPBaseURL:      os.Getenv("FOP_BASE_URL"),
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate reports every missing required field in one error (required
// fields exclude AADTokenURL, which is a test-only override).
func (c *Config) Validate() error {
	var missing []string
	for _, f := range []struct {
		name, value string
	}{
		{"FOP_AAD_TENANT_ID", c.AADTenantID},
		{"FOP_AAD_CLIENT_ID", c.AADClientID},
		{"FOP_AAD_CLIENT_SECRET", c.AADClientSecret},
		{"FOP_AAD_SCOPE", c.AADScope},
		{"FOP_PARTNER_ID", c.PartnerID},
		{"FOP_BASE_URL", c.FOPBaseURL},
	} {
		if f.value == "" {
			missing = append(missing, f.name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("fop client config: missing required settings: %v", missing)
	}
	return nil
}

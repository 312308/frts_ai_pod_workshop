// Package fopclient provides an authenticated HTTP client for the FOP API.
//
// Source: BRD FR-024 ("Implement FOP client wrapper... Acquire OAuth2 JWT
// token from TCG per environment (client credentials flow, per-environment
// registered client). Inject partnerId on every API call."); API Guide PDF
// ("401/403... Missing or invalid OAUTH token... Managed Identity or Service
// Principal issue" — confirms Azure AD as the identity provider).
package fopclient

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"tcg-service/internal/config"
)

// aadTokenURL builds the Azure AD v2.0 token endpoint for the given tenant.
func aadTokenURL(tenantID string) string {
	return fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)
}

// tokenSource returns an oauth2.TokenSource that acquires a JWT via the
// OAuth2 Client Credentials grant. golang.org/x/oauth2 caches the token and
// re-acquires it automatically once expired (AC-005a, AC-005c) — no
// hand-rolled expiry/refresh logic needed here.
func tokenSource(ctx context.Context, cfg *config.Config) oauth2.TokenSource {
	tokenURL := cfg.AADTokenURL
	if tokenURL == "" {
		tokenURL = aadTokenURL(cfg.AADTenantID)
	}
	ccConfig := clientcredentials.Config{
		ClientID:     cfg.AADClientID,
		ClientSecret: cfg.AADClientSecret,
		TokenURL:     tokenURL,
		Scopes:       []string{cfg.AADScope},
		// Azure AD's v2.0 token endpoint expects client_id/client_secret in
		// the POST body, not HTTP Basic auth. Pin the style explicitly
		// rather than rely on AuthStyleAutoDetect's first-request guess.
		AuthStyle: oauth2.AuthStyleInParams,
	}
	return ccConfig.TokenSource(ctx)
}

// partnerIDTransport injects the partnerId query parameter on every
// outbound request.
//
// Source: API Guide PDF Rule 1 ("All API end points, GET or POST, require
// query parameter partnerId. If the Partner Id is missing or invalid, a
// HTTP Error code 406 will be returned"); BRD FR-024 ("config-driven, never
// hard-coded").
type partnerIDTransport struct {
	base      http.RoundTripper
	partnerID string
}

func (t *partnerIDTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqClone := req.Clone(req.Context())
	q := reqClone.URL.Query()
	q.Set("partnerId", t.partnerID)
	reqClone.URL.RawQuery = q.Encode()
	return t.base.RoundTrip(reqClone)
}

func newPartnerIDTransport(base http.RoundTripper, partnerID string) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &partnerIDTransport{base: base, partnerID: partnerID}
}

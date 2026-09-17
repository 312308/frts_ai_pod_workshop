package fopclient

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"tcg-service/internal/config"
)

const testSecret = "s3cr3t-sentinel-value-do-not-leak"

// mockAADServer returns an httptest.Server standing in for the Azure AD
// v2.0 token endpoint. tokenFn is invoked once per token request and
// returns the (status code, response body) to send.
func mockAADServer(t *testing.T, tokenFn func(callN int) (int, string)) *httptest.Server {
	t.Helper()
	var calls int32
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(atomic.AddInt32(&calls, 1))
		if err := r.ParseForm(); err != nil {
			t.Fatalf("mock AAD: parse form: %v", err)
		}
		if got := r.FormValue("grant_type"); got != "client_credentials" {
			t.Errorf("mock AAD: grant_type = %q, want client_credentials", got)
		}
		if got := r.FormValue("client_id"); got != "test-client" {
			t.Errorf("mock AAD: client_id = %q, want test-client", got)
		}
		if got := r.FormValue("client_secret"); got != testSecret {
			t.Errorf("mock AAD: client_secret = %q, want %q", got, testSecret)
		}
		status, body := tokenFn(n)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func baseTestConfig(tokenURL, fopURL string) *config.Config {
	return &config.Config{
		AADTenantID:     "test-tenant",
		AADClientID:     "test-client",
		AADClientSecret: testSecret,
		AADScope:        "api://fop/.default",
		PartnerID:       "ISM001",
		FOPBaseURL:      fopURL,
		AADTokenURL:     tokenURL,
	}
}

// TestFOPClient_TokenAcquired_TC_001 — JWT acquired via OAuth2 Client
// Credentials flow; outbound request carries the returned bearer token.
func TestFOPClient_TokenAcquired_TC_001(t *testing.T) {
	aad := mockAADServer(t, func(int) (int, string) {
		return http.StatusOK, `{"access_token":"jwt-abc123","token_type":"Bearer","expires_in":3600}`
	})
	defer aad.Close()

	var gotAuthHeader string
	fop := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer fop.Close()

	cfg := baseTestConfig(aad.URL, fop.URL)
	client, err := NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, fop.URL+"/api/v1/wholesale-products", nil)
	if _, err := client.Do(req); err != nil {
		t.Fatalf("Do: %v", err)
	}

	if gotAuthHeader != "Bearer jwt-abc123" {
		t.Errorf("Authorization header = %q, want %q", gotAuthHeader, "Bearer jwt-abc123")
	}
}

// TestFOPClient_PartnerIDInjected_TC_002 — partnerId query parameter is
// injected on every outbound call, sourced from config.
func TestFOPClient_PartnerIDInjected_TC_002(t *testing.T) {
	aad := mockAADServer(t, func(int) (int, string) {
		return http.StatusOK, `{"access_token":"jwt-abc123","token_type":"Bearer","expires_in":3600}`
	})
	defer aad.Close()

	var gotPartnerIDs []string
	fop := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPartnerIDs = append(gotPartnerIDs, r.URL.Query().Get("partnerId"))
		w.WriteHeader(http.StatusOK)
	}))
	defer fop.Close()

	cfg := baseTestConfig(aad.URL, fop.URL)
	client, err := NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	for _, path := range []string{"/api/v1/wholesale-products", "/api/v1/suppliers"} {
		req, _ := http.NewRequest(http.MethodGet, fop.URL+path, nil)
		if _, err := client.Do(req); err != nil {
			t.Fatalf("Do(%s): %v", path, err)
		}
	}

	if len(gotPartnerIDs) != 2 {
		t.Fatalf("got %d requests, want 2", len(gotPartnerIDs))
	}
	for i, got := range gotPartnerIDs {
		if got != "ISM001" {
			t.Errorf("request %d: partnerId = %q, want ISM001", i, got)
		}
	}
}

// TestFOPClient_TokenRefreshOnExpiry_TC_003 — a stale token triggers
// re-acquisition; the second call uses a newly issued token.
func TestFOPClient_TokenRefreshOnExpiry_TC_003(t *testing.T) {
	aad := mockAADServer(t, func(n int) (int, string) {
		if n == 1 {
			return http.StatusOK, `{"access_token":"jwt-first","token_type":"Bearer","expires_in":1}`
		}
		return http.StatusOK, `{"access_token":"jwt-second","token_type":"Bearer","expires_in":3600}`
	})
	defer aad.Close()

	var gotAuthHeaders []string
	fop := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeaders = append(gotAuthHeaders, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer fop.Close()

	cfg := baseTestConfig(aad.URL, fop.URL)
	client, err := NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	req1, _ := http.NewRequest(http.MethodGet, fop.URL+"/api/v1/wholesale-products", nil)
	if _, err := client.Do(req1); err != nil {
		t.Fatalf("Do (first): %v", err)
	}

	time.Sleep(1100 * time.Millisecond) // past the 1s expiry

	req2, _ := http.NewRequest(http.MethodGet, fop.URL+"/api/v1/wholesale-products", nil)
	if _, err := client.Do(req2); err != nil {
		t.Fatalf("Do (second): %v", err)
	}

	if len(gotAuthHeaders) != 2 {
		t.Fatalf("got %d requests, want 2", len(gotAuthHeaders))
	}
	if gotAuthHeaders[0] != "Bearer jwt-first" {
		t.Errorf("first request Authorization = %q, want Bearer jwt-first", gotAuthHeaders[0])
	}
	if gotAuthHeaders[1] != "Bearer jwt-second" {
		t.Errorf("second request Authorization = %q, want Bearer jwt-second (expected re-acquisition)", gotAuthHeaders[1])
	}
}

// TestFOPClient_TokenAcquisitionFailure_TC_004 — invalid credentials surface
// as an error from Do; no request reaches the FOP resource endpoint.
func TestFOPClient_TokenAcquisitionFailure_TC_004(t *testing.T) {
	aad := mockAADServer(t, func(int) (int, string) {
		return http.StatusUnauthorized, `{"error":"invalid_client","error_description":"invalid client credentials"}`
	})
	defer aad.Close()

	fopCalled := false
	fop := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fopCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer fop.Close()

	cfg := baseTestConfig(aad.URL, fop.URL)
	client, err := NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, fop.URL+"/api/v1/wholesale-products", nil)
	_, err = client.Do(req)
	if err == nil {
		t.Fatal("Do: got nil error, want token-acquisition failure")
	}
	if fopCalled {
		t.Error("FOP resource endpoint was called despite token acquisition failure")
	}

	// TC-005: the client secret must never appear in the surfaced error.
	if strings.Contains(err.Error(), testSecret) {
		t.Errorf("error message leaks client secret: %v", err)
	}
}

// TestFOPClient_NoSecretInErrors_TC_005 — duplicate-checks TC-004's error
// path specifically for secret leakage, independent of the failure
// assertion above (kept as its own test so a future refactor of TC-004
// cannot silently drop the secret-leak check).
func TestFOPClient_NoSecretInErrors_TC_005(t *testing.T) {
	aad := mockAADServer(t, func(int) (int, string) {
		return http.StatusUnauthorized, `{"error":"invalid_client"}`
	})
	defer aad.Close()

	cfg := baseTestConfig(aad.URL, "http://fop.invalid")
	client, err := NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "http://fop.invalid/api/v1/wholesale-products", nil)
	_, err = client.Do(req)
	if err == nil {
		t.Fatal("Do: got nil error, want token-acquisition failure")
	}
	if strings.Contains(err.Error(), testSecret) {
		t.Errorf("error message leaks client secret: %v", err)
	}
}

// TestFOPClient_MissingPartnerID_TC_006 — a config with no partnerId must
// fail at construction time, never reach FOP and risk a live 406.
func TestFOPClient_MissingPartnerID_TC_006(t *testing.T) {
	cfg := baseTestConfig("https://aad.invalid", "https://fop.invalid")
	cfg.PartnerID = ""

	_, err := NewClient(context.Background(), cfg)
	if err == nil {
		t.Fatal("NewClient: got nil error for missing PartnerID, want fail-fast error")
	}
	if !strings.Contains(err.Error(), "FOP_PARTNER_ID") {
		t.Errorf("error = %v, want it to name FOP_PARTNER_ID", err)
	}
}

// TestAADTokenURL — locks down the real Azure AD v2.0 token endpoint shape,
// since every other test overrides it via AADTokenURL and never exercises
// the production code path.
func TestAADTokenURL(t *testing.T) {
	got := aadTokenURL("my-tenant-id")
	want := "https://login.microsoftonline.com/my-tenant-id/oauth2/v2.0/token"
	if got != want {
		t.Errorf("aadTokenURL(%q) = %q, want %q", "my-tenant-id", got, want)
	}
}

func init() {
	// Guard against copy-paste drift: fail loudly at test-binary init if the
	// sentinel secret ever collides with a real-looking value, since TC-005
	// depends on this string being distinctive enough to detect a leak.
	if len(testSecret) < 10 {
		panic(fmt.Sprintf("testSecret too short for a meaningful leak check: %q", testSecret))
	}
}

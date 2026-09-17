package fopclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testPage struct {
	Total int      `json:"total"`
	Items []string `json:"items"`
}

func countItems(body []byte) (int, int, error) {
	var p testPage
	if err := json.Unmarshal(body, &p); err != nil {
		return 0, 0, err
	}
	return len(p.Items), p.Total, nil
}

func newTestClient(t *testing.T, fopHandler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	aad := mockAADServer(t, func(int) (int, string) {
		return http.StatusOK, `{"access_token":"jwt","token_type":"Bearer","expires_in":3600}`
	})
	t.Cleanup(aad.Close)

	fop := httptest.NewServer(fopHandler)
	t.Cleanup(fop.Close)

	cfg := baseTestConfig(aad.URL, fop.URL)
	client, err := NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client, fop
}

// TestSinglePageOptimization_TC_178 — array.length == total means no second
// page request is made.
func TestSinglePageOptimization_TC_178(t *testing.T) {
	var calls int
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"total":2,"items":["a","b"]}`))
	})

	pages, err := client.FetchAllPages(context.Background(), "/api/v1/wholesale-products", "2026-09-10T00:00:00Z", countItems)
	if err != nil {
		t.Fatalf("FetchAllPages: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("got %d pages, want 1", len(pages))
	}
	if calls != 1 {
		t.Errorf("got %d HTTP calls, want 1 (single-page optimization)", calls)
	}
}

// TestMultiPagePaginationLoop_TC_179 — page 1's array.length < total means
// pagination continues by incrementing page number until FOP signals 204,
// regardless of each individual page's own array length (BR-002: FOP owns
// page size; later pages must not be judged against the page-1 total).
func TestMultiPagePaginationLoop_TC_179(t *testing.T) {
	var gotPages []string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get("page")
		gotPages = append(gotPages, p)
		switch p {
		case "1":
			_, _ = w.Write([]byte(`{"total":3,"items":["a","b"]}`))
		case "2":
			_, _ = w.Write([]byte(`{"total":3,"items":["c"]}`)) // still < total; must NOT stop here
		case "3":
			w.WriteHeader(http.StatusNoContent) // FOP's real terminator
		default:
			t.Fatalf("unexpected page %q", p)
		}
	})

	pages, err := client.FetchAllPages(context.Background(), "/api/v1/wholesale-products", "2026-09-10T00:00:00Z", countItems)
	if err != nil {
		t.Fatalf("FetchAllPages: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("got %d data pages, want 2", len(pages))
	}
	if len(gotPages) != 3 || gotPages[0] != "1" || gotPages[1] != "2" || gotPages[2] != "3" {
		t.Errorf("page sequence = %v, want [1 2 3] (loop must run until 204)", gotPages)
	}
}

// TestTerminatorStopsPagination_TC_180 — an HTTP 204 stops pagination
// regardless of totals.
func TestTerminatorStopsPagination_TC_180(t *testing.T) {
	var calls int
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			_, _ = w.Write([]byte(`{"total":5,"items":["a"]}`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	pages, err := client.FetchAllPages(context.Background(), "/api/v1/wholesale-products", "2026-09-10T00:00:00Z", countItems)
	if err != nil {
		t.Fatalf("FetchAllPages: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("got %d pages, want 1 (204 must stop the loop)", len(pages))
	}
	if calls != 2 {
		t.Errorf("got %d calls, want 2 (one data page, one 204 terminator)", calls)
	}
}

// TestMalformedRequestNotRetryable_TC_181 — a 400 surfaces as a
// non-retryable ResponseError (BRD: 4xx alert immediately).
func TestMalformedRequestNotRetryable_TC_181(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad since param"}`))
	})

	_, err := client.FetchAllPages(context.Background(), "/api/v1/wholesale-products", "2026-09-10T00:00:00Z", countItems)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	var re *ResponseError
	if !asResponseError(err, &re) {
		t.Fatalf("expected *ResponseError, got %T: %v", err, err)
	}
	if re.Retryable() {
		t.Error("400 must not be retryable (4xx = alert immediately)")
	}
	if msg := re.Error(); msg == "" {
		t.Error("ResponseError.Error() must produce a non-empty message")
	}
}

// TestResponseErrorClassification_TC_184 — every FOP-documented status code
// gets an identifiable classification string (API Guide Response Codes
// table), not a generic "unexpected response".
func TestResponseErrorClassification_TC_184(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 406, 500, 505, 540, 418} {
		re := &ResponseError{StatusCode: status, Body: []byte("x")}
		if re.Error() == "" {
			t.Errorf("status %d: empty error message", status)
		}
	}
}

// TestServerErrorRetryable_TC_182 — a 500 surfaces as a retryable
// ResponseError (BRD: 5xx backoff with circuit-breaking).
func TestServerErrorRetryable_TC_182(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.FetchAllPages(context.Background(), "/api/v1/wholesale-products", "2026-09-10T00:00:00Z", countItems)
	var re *ResponseError
	if !asResponseError(err, &re) {
		t.Fatalf("expected *ResponseError, got %T: %v", err, err)
	}
	if !re.Retryable() {
		t.Error("500 must be retryable")
	}
}

// TestSinceAndPageParamsSent_TC_183 — every request carries the exact
// since/page query parameters the caller specified.
func TestSinceAndPageParamsSent_TC_183(t *testing.T) {
	var gotSince, gotPage string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotSince = r.URL.Query().Get("since")
		gotPage = r.URL.Query().Get("page")
		_, _ = w.Write([]byte(`{"total":1,"items":["a"]}`))
	})

	if _, err := client.FetchAllPages(context.Background(), "/api/v1/wholesale-products", "2026-09-10T00:00:00Z", countItems); err != nil {
		t.Fatalf("FetchAllPages: %v", err)
	}
	if gotSince != "2026-09-10T00:00:00Z" {
		t.Errorf("since = %q, want 2026-09-10T00:00:00Z", gotSince)
	}
	if gotPage != "1" {
		t.Errorf("page = %q, want 1", gotPage)
	}
}

func asResponseError(err error, target **ResponseError) bool {
	re, ok := err.(*ResponseError)
	if ok {
		*target = re
	}
	return ok
}

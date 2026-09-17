package fopclient

import (
	"fmt"
	"net/http"
)

// ResponseError wraps a non-2xx FOP response with the API Guide's own
// classification of what each code means (BRD FR-026; API Guide PDF
// Response Codes table).
type ResponseError struct {
	StatusCode int
	Body       []byte
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("fop: %s (status %d): %s", classify(e.StatusCode), e.StatusCode, e.Body)
}

// Retryable reports whether the caller should backoff-and-retry (5xx/540)
// rather than alert immediately as an operational/contract defect (4xx).
// Source: BRD FR-026 ("4xx alert immediately; 5xx/540 backoff with
// circuit-breaking"); actual backoff/circuit-breaker policy lives in the
// Temporal workflow (SP-06/SP-13/SP-20), not here.
func (e *ResponseError) Retryable() bool {
	switch e.StatusCode {
	case http.StatusInternalServerError, 505, 540:
		return true
	default:
		return false
	}
}

func classify(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "malformed request"
	case http.StatusUnauthorized, http.StatusForbidden:
		return "unauthorized (IP restriction, missing/invalid OAuth token, or managed identity/service principal issue)"
	case http.StatusNotFound:
		return "invalid URI"
	case 406:
		return "partnerId invalid or partner not allowed to use this function"
	case http.StatusInternalServerError:
		return "internal platform error"
	case 505:
		return "API version no longer supported"
	case 540:
		return "temporarily disabled (FOP configuration)"
	default:
		return "unexpected response"
	}
}

// responseCodeError returns a *ResponseError for any status FOP defines as
// non-success, or nil for 200/204 (BRD FR-026, AC-005f).
func responseCodeError(status int, body []byte) error {
	switch status {
	case http.StatusOK, http.StatusNoContent:
		return nil
	default:
		return &ResponseError{StatusCode: status, Body: body}
	}
}

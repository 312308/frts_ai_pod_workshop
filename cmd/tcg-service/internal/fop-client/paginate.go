package fopclient

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

// CountFunc extracts (arrayLen, total) from a page's raw JSON body, so
// FetchAllPages can decide whether to keep paginating without needing to
// know each entity's concrete Go type (BRD FR-025).
type CountFunc func(body []byte) (arrayLen, total int, err error)

// FetchAllPages implements the FOP delta+pagination contract.
//
// Per the API Guide PDF verbatim: the array-length-vs-total comparison is
// only ever made on page 1, purely as an optimization to skip a
// known-unnecessary extra call ("a length that is equal to the total
// property... an additional call... is not required"). For every other
// case ("a length that is LESS THAN the total property") the real
// termination signal for ALL subsequent pages is the HTTP 204 response —
// page size is FOP's to own and tune (BR-002), so later pages' individual
// lengths are not a reliable stopping signal and must not be compared
// page-by-page.
func (c *Client) FetchAllPages(ctx context.Context, path, since string, count CountFunc) ([][]byte, error) {
	var pages [][]byte
	page := 1
	for {
		body, status, err := c.fetchPage(ctx, path, since, page)
		if err != nil {
			return pages, err
		}
		if err := responseCodeError(status, body); err != nil {
			return pages, err
		}
		if status == http.StatusNoContent {
			break
		}
		pages = append(pages, body)

		if page == 1 {
			arrayLen, total, err := count(body)
			if err != nil {
				return pages, err
			}
			if arrayLen >= total {
				break // single-page optimization
			}
		}
		page++
	}
	return pages, nil
}

func (c *Client) fetchPage(ctx context.Context, path, since string, page int) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return nil, 0, err
	}
	q := req.URL.Query()
	q.Set("since", since)
	q.Set("page", strconv.Itoa(page))
	req.URL.RawQuery = q.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

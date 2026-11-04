package gogithub

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type secondaryRateLimitTransport struct {
	base http.RoundTripper
}

func (t secondaryRateLimitTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := t.base.RoundTrip(request)
	if err != nil || !needsSecondaryRateLimitHint(response) {
		return response, err
	}
	remaining, parseErr := strconv.ParseInt(response.Header.Get("X-RateLimit-Remaining"), 10, 64)
	if parseErr != nil || remaining <= 0 {
		return response, nil
	}

	const maxErrorBody = 64 << 10
	data, readErr := io.ReadAll(io.LimitReader(response.Body, maxErrorBody+1))
	response.Body = restoredResponseBody{
		Reader: io.MultiReader(bytes.NewReader(data), response.Body),
		Closer: response.Body,
	}
	if readErr != nil || len(data) > maxErrorBody {
		return response, nil
	}
	var failure struct {
		DocumentationURL string `json:"documentation_url"`
	}
	if json.Unmarshal(data, &failure) != nil ||
		(!strings.HasSuffix(failure.DocumentationURL, "secondary-rate-limits") &&
			!strings.HasSuffix(failure.DocumentationURL, "#abuse-rate-limits")) {
		return response, nil
	}
	// go-github otherwise treats an unexhausted primary reset as a secondary wait
	// and caches it. GitHub documents a one-minute minimum when Retry-After is absent.
	response.Header = response.Header.Clone()
	response.Header.Set("Retry-After", "60")
	return response, nil
}

func needsSecondaryRateLimitHint(response *http.Response) bool {
	return response != nil && response.StatusCode == http.StatusForbidden &&
		response.Header.Get("Retry-After") == "" && response.Body != nil
}

type restoredResponseBody struct {
	io.Reader
	io.Closer
}

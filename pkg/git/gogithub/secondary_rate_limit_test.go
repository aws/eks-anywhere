package gogithub

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v74/github"
)

type responseTransport func(*http.Request) (*http.Response, error)

func (t responseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return t(request)
}

func TestSecondaryLimitSDKCacheUsesDocumentedWait(t *testing.T) {
	for _, test := range []struct {
		name       string
		remaining  string
		retryAfter string
		wantDelay  time.Duration
		primary    bool
	}{
		{name: "unexhausted primary reset is not a secondary wait", remaining: "100", wantDelay: time.Minute},
		{name: "explicit server delay preserved", remaining: "100", retryAfter: "600", wantDelay: 10 * time.Minute},
		{name: "primary exhaustion preserved", remaining: "0", wantDelay: time.Hour, primary: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			base := responseTransport(func(request *http.Request) (*http.Response, error) {
				calls++
				return &http.Response{
					StatusCode: http.StatusForbidden, Request: request,
					Header: http.Header{
						"X-Ratelimit-Remaining": {test.remaining},
						"X-Ratelimit-Limit":     {"5000"},
						"X-Ratelimit-Reset":     {strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10)},
						"Retry-After":           {test.retryAfter},
					},
					Body: io.NopCloser(strings.NewReader(`{"message":"You have exceeded a secondary rate limit","documentation_url":"https://docs.github.com/rest/overview/resources-in-the-rest-api#secondary-rate-limits"}`)),
				}, nil
			})
			sdk := github.NewClient(&http.Client{Transport: secondaryRateLimitTransport{base: base}})
			for attempt := 0; attempt < 2; attempt++ {
				_, _, err := sdk.Users.Get(context.Background(), "")
				assertSDKRateLimitDelay(t, err, test.primary, test.wantDelay)
			}
			if calls != 1 {
				t.Fatalf("SDK rate cache did not hold second call: requests = %d", calls)
			}
		})
	}
}

func assertSDKRateLimitDelay(t *testing.T, err error, primaryLimit bool, wantDelay time.Duration) {
	t.Helper()
	var delay time.Duration
	if primaryLimit {
		var primary *github.RateLimitError
		if !errors.As(err, &primary) {
			t.Fatalf("error type = %T; want primary limit", err)
		}
		delay = time.Until(primary.Rate.Reset.Time)
	} else {
		var secondary *github.AbuseRateLimitError
		if !errors.As(err, &secondary) || secondary.RetryAfter == nil {
			t.Fatalf("error = %v; want typed secondary limit", err)
		}
		delay = *secondary.RetryAfter
	}
	if delay < wantDelay-2*time.Second || delay > wantDelay {
		t.Fatalf("rate limit delay = %v; want %v", delay, wantDelay)
	}
}

func TestSecondaryLimitTransportRetainsPermissionErrorAndBody(t *testing.T) {
	body := `{"message":"Resource not accessible by integration","documentation_url":"https://docs.github.com/rest/repos/repos"}`
	calls := 0
	base := responseTransport(func(request *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{
			StatusCode: http.StatusForbidden, Request: request,
			Header: http.Header{"X-Ratelimit-Remaining": {"100"}},
			Body:   io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	sdk := github.NewClient(&http.Client{Transport: secondaryRateLimitTransport{base: base}})
	for range 2 {
		_, response, err := sdk.Users.Get(context.Background(), "")
		var permission *github.ErrorResponse
		if !errors.As(err, &permission) || permission.Message != "Resource not accessible by integration" {
			t.Fatalf("permission error changed: %v", err)
		}
		if response.Header.Get("Retry-After") != "" {
			t.Fatal("permission error gained a retry delay")
		}
	}
	if calls != 2 {
		t.Fatalf("permission error was cached as rate limit: requests = %d", calls)
	}
}

func TestSecondaryLimitTransportPreservesOversizedBody(t *testing.T) {
	body := strings.Repeat("x", (64<<10)+100)
	base := responseTransport(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusForbidden, Request: request,
			Header: http.Header{"X-Ratelimit-Remaining": {"100"}},
			Body:   io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	request, err := http.NewRequest(http.MethodGet, "https://example.test/user", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := (secondaryRateLimitTransport{base: base}).RoundTrip(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	retained, err := io.ReadAll(response.Body)
	if err != nil || string(retained) != body || response.Header.Get("Retry-After") != "" {
		t.Fatalf("oversized body changed: bytes=%d err=%v", len(retained), err)
	}
}

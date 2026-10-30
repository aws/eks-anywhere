package githubretry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v74/github"
)

func response(status int, headers http.Header) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     headers,
		Request:    &http.Request{Method: http.MethodPost, URL: &url.URL{Scheme: "https", Host: "api.github.com", Path: "/user/repos"}},
	}
}

func TestRateLimitDelay(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	retryAfter := 90 * time.Second
	tests := []struct {
		name    string
		err     error
		attempt int
		delay   time.Duration
		limited bool
	}{
		{
			name: "wrapped secondary limit honors Retry-After",
			err: fmt.Errorf("create repository: %w", &github.AbuseRateLimitError{
				Response: response(403, http.Header{"Retry-After": {"120"}}),
			}),
			attempt: 1, delay: 2 * time.Minute, limited: true,
		},
		{
			name: "secondary client retry duration",
			err: &github.AbuseRateLimitError{
				Response: response(403, nil), RetryAfter: &retryAfter,
			},
			attempt: 1, delay: retryAfter, limited: true,
		},
		{
			name: "primary rate reset",
			err: &github.RateLimitError{
				Response: response(403, nil),
				Rate:     github.Rate{Reset: github.Timestamp{Time: now.Add(10 * time.Minute)}},
			},
			attempt: 1, delay: 10 * time.Minute, limited: true,
		},
		{
			name: "429 with Retry-After date",
			err: &github.ErrorResponse{Response: response(429, http.Header{
				"Retry-After": {now.Add(3 * time.Minute).Format(http.TimeFormat)},
			})},
			attempt: 1, delay: 3 * time.Minute, limited: true,
		},
		{
			name: "exhausted primary quota headers",
			err: &github.ErrorResponse{Response: response(429, http.Header{
				"X-Ratelimit-Remaining": {"0"},
				"X-Ratelimit-Reset":     {fmt.Sprint(now.Add(5 * time.Minute).Unix())},
			})},
			attempt: 1, delay: 5 * time.Minute, limited: true,
		},
		{
			name: "quota reset later than Retry-After",
			err: &github.ErrorResponse{Response: response(429, http.Header{
				"Retry-After":           {"30"},
				"X-Ratelimit-Remaining": {"0"},
				"X-Ratelimit-Reset":     {fmt.Sprint(now.Add(5 * time.Minute).Unix())},
			})},
			attempt: 1, delay: 5 * time.Minute, limited: true,
		},
		{
			name: "typed quota reset later than Retry-After",
			err: &github.RateLimitError{
				Response: response(403, http.Header{"Retry-After": {"30"}}),
				Rate:     github.Rate{Reset: github.Timestamp{Time: now.Add(10 * time.Minute)}},
			},
			attempt: 1, delay: 10 * time.Minute, limited: true,
		},
		{
			name:    "Flux secondary rate limit without headers",
			err:     errors.New("executing flux bootstrap github: POST https://api.github.com/user/repos: 403 You have exceeded a secondary rate limit"),
			attempt: 1, delay: time.Minute, limited: true,
		},
		{
			name:    "Flux 429 without headers backs off exponentially",
			err:     errors.New("POST https://api.github.com/user/repos: 429 Too Many Requests"),
			attempt: 3, delay: 4 * time.Minute, limited: true,
		},
		{
			name: "invalid Retry-After uses safe fallback",
			err: &github.AbuseRateLimitError{
				Response: response(403, http.Header{"Retry-After": {"invalid"}}),
			},
			attempt: 2, delay: 2 * time.Minute, limited: true,
		},
		{
			name:    "ordinary permission error is not a rate limit",
			err:     &github.ErrorResponse{Response: response(403, nil), Message: "Resource not accessible by integration"},
			attempt: 1,
		},
		{
			name:    "unrelated error containing 429",
			err:     errors.New("failed to open file 429"),
			attempt: 1,
		},
		{
			name:    "API error without response",
			err:     &github.ErrorResponse{Message: "Forbidden"},
			attempt: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay, limited := rateLimitDelay(tt.err, tt.attempt, now)
			if delay != tt.delay || limited != tt.limited {
				t.Fatalf("got (%v, %v), want (%v, %v)", delay, limited, tt.delay, tt.limited)
			}
		})
	}
}

func TestRetryRateLimitThenSuccess(t *testing.T) {
	calls := 0
	var delays []time.Duration
	err := retry(context.Background(), 0, func() error {
		calls++
		if calls <= 2 {
			return &github.AbuseRateLimitError{Response: response(403, nil)}
		}
		return nil
	}, func(_ context.Context, delay time.Duration) error {
		delays = append(delays, delay)
		return nil
	})
	if err != nil || calls != 3 || len(delays) != 2 {
		t.Fatalf("err=%v calls=%d delays=%v", err, calls, delays)
	}
	for i, delay := range delays {
		minimum := time.Minute * time.Duration(1<<i)
		if delay < minimum || delay >= minimum+30*time.Second {
			t.Fatalf("retry %d delay %v outside expected range", i, delay)
		}
	}
}

func TestRetryStopsAfterFiveAttempts(t *testing.T) {
	calls, sleeps := 0, 0
	want := &github.AbuseRateLimitError{Response: response(403, nil)}
	err := retry(context.Background(), 0, func() error {
		calls++
		return want
	}, func(context.Context, time.Duration) error {
		sleeps++
		return nil
	})
	if err != want || calls != 5 || sleeps != 4 {
		t.Fatalf("err=%v calls=%d sleeps=%d", err, calls, sleeps)
	}
}

func TestRetryDoesNotRepeatUnrelatedAPIError(t *testing.T) {
	calls := 0
	want := &github.ErrorResponse{Response: response(403, nil), Message: "Forbidden"}
	err := Retry(context.Background(), 0, func() error {
		calls++
		return want
	})
	if err != want || calls != 1 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

func TestRetryPreservesOrdinaryBackoff(t *testing.T) {
	calls := 0
	err := retry(context.Background(), 5*time.Second, func() error {
		calls++
		if calls < 3 {
			return errors.New("temporary bootstrap failure")
		}
		return nil
	}, func(_ context.Context, delay time.Duration) error {
		if delay != 5*time.Second {
			t.Fatalf("ordinary backoff=%v", delay)
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

func TestRetryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Retry(ctx, 0, func() error {
		t.Fatal("operation executed after cancellation")
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestRetryCancellationWhileWaiting(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	calls := 0
	err := Retry(ctx, 0, func() error {
		calls++
		return &github.AbuseRateLimitError{Response: response(403, nil)}
	})
	if !errors.Is(err, context.DeadlineExceeded) || calls != 1 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

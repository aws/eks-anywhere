package githubretry

import (
	"context"
	"errors"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const maxAttempts = 5

// Retry honors GitHub rate limits. Other errors are retried only if backoff is positive.
func Retry(ctx context.Context, backoff time.Duration, operation func() error) error {
	return retry(ctx, backoff, operation, wait)
}

func retry(ctx context.Context, backoff time.Duration, operation func() error, sleep func(context.Context, time.Duration) error) error {
	for attempt := 1; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := operation()
		if err == nil || attempt == maxAttempts {
			return err
		}

		delay, limited := rateLimitDelay(err, attempt, time.Now())
		if limited {
			// Spread retries from independent runners without shortening GitHub's requested wait.
			delay += time.Duration(rand.Int64N(int64(30 * time.Second)))
			logger.Info("GitHub rate limited the request; waiting before retry", "wait", delay, "attempt", attempt)
		} else {
			if backoff <= 0 {
				return err
			}
			delay = backoff
		}
		if err := sleep(ctx, delay); err != nil {
			return err
		}
	}
}

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func rateLimitDelay(err error, attempt int, now time.Time) (time.Duration, bool) {
	fallback := time.Minute * time.Duration(1<<(attempt-1))
	var primary *github.RateLimitError
	if errors.As(err, &primary) {
		delay, ok := headerDelay(primary.Response, now)
		if reset := primary.Rate.Reset.Sub(now); reset > delay {
			return reset, true
		}
		if ok {
			return delay, true
		}
		return fallback, true
	}
	var secondary *github.AbuseRateLimitError
	if errors.As(err, &secondary) {
		if delay, ok := headerDelay(secondary.Response, now); ok {
			return delay, true
		}
		if secondary.RetryAfter != nil && *secondary.RetryAfter > 0 {
			return *secondary.RetryAfter, true
		}
		return fallback, true
	}
	var response *github.ErrorResponse
	if errors.As(err, &response) {
		if response.Response == nil {
			if isRateLimitMessage(response.Message) {
				return fallback, true
			}
			return 0, false
		}
		status := response.Response.StatusCode
		if status == http.StatusTooManyRequests ||
			(status == http.StatusForbidden && isRateLimitMessage(response.Message)) {
			if delay, ok := headerDelay(response.Response, now); ok {
				return delay, true
			}
			return fallback, true
		}
		return 0, false
	}
	// Flux bootstrap is a subprocess, so its GitHub API errors arrive as text.
	if isRateLimitMessage(err.Error()) {
		return fallback, true
	}
	return 0, false
}

func isRateLimitMessage(message string) bool {
	message = strings.ToLower(message)
	return strings.Contains(message, "secondary rate limit") ||
		strings.Contains(message, "api rate limit exceeded") ||
		(strings.Contains(message, "api.github.com") &&
			(strings.Contains(message, ": 429 ") || strings.Contains(message, "429 too many requests")))
}

func headerDelay(response *http.Response, now time.Time) (time.Duration, bool) {
	if response == nil {
		return 0, false
	}
	var delay time.Duration
	if value := response.Header.Get("Retry-After"); value != "" {
		if seconds, err := strconv.ParseInt(value, 10, 32); err == nil && seconds > 0 {
			delay = time.Duration(seconds) * time.Second
		}
		if date, err := http.ParseTime(value); err == nil && date.After(now) {
			delay = date.Sub(now)
		}
	}
	if response.Header.Get("X-RateLimit-Remaining") == "0" {
		if seconds, err := strconv.ParseInt(response.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil {
			if reset := time.Unix(seconds, 0).Sub(now); reset > delay {
				delay = reset
			}
		}
	}
	return delay, delay > 0
}

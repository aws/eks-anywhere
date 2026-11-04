package githubretry

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// RetryGitPush retries a GitHub server failure while preserving permanent push errors.
func RetryGitPush(ctx context.Context, repositoryURL string, operation func() error) error {
	return retryGitPush(ctx, repositoryURL, operation, wait)
}

func retryGitPush(ctx context.Context, repositoryURL string, operation func() error, sleep func(context.Context, time.Duration) error) error {
	remote, err := url.Parse(repositoryURL)
	githubRemote := err == nil && strings.EqualFold(remote.Hostname(), "github.com")
	for attempt := 1; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := operation()
		// A failed response can hide a successful push; the retry then finds no changes.
		if gitPushSucceeded(err, attempt > 1) {
			return nil
		}
		if !githubRemote || attempt == maxAttempts || !isTransientGitPush(err) {
			return err
		}
		if err := sleep(ctx, time.Second*time.Duration(1<<(attempt-1))); err != nil {
			return err
		}
	}
}

func gitPushSucceeded(err error, retried bool) bool {
	return err == nil || retried && errors.Is(err, gogit.NoErrAlreadyUpToDate)
}

func isTransientGitPush(err error) bool {
	var unexpected *plumbing.UnexpectedError
	if errors.As(err, &unexpected) {
		err = unexpected.Err
	}
	var response *githttp.Err
	if errors.As(err, &response) && response.Response != nil {
		switch response.Response.StatusCode {
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		}
		return false
	}
	// go-git returns receive-pack command status as an untyped error.
	message := err.Error()
	return strings.HasPrefix(message, "command error on refs/") &&
		strings.HasSuffix(message, ": Internal Server Error")
}

package githubretry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

func TestGitPushReceivePackServerFailureThenSuccess(t *testing.T) {
	for _, next := range []error{nil, gogit.NoErrAlreadyUpToDate} {
		calls := 0
		var delays []time.Duration
		server := (&packp.CommandStatus{ReferenceName: "refs/heads/main", Status: "Internal Server Error"}).Error()
		err := retryGitPush(context.Background(), "https://github.com/testuser/testrepo.git", func() error {
			calls++
			if calls == 1 {
				return server
			}
			return next
		}, func(_ context.Context, delay time.Duration) error {
			delays = append(delays, delay)
			return nil
		})
		if err != nil || calls != 2 || len(delays) != 1 || delays[0] != time.Second {
			t.Fatalf("err=%v calls=%d delays=%v", err, calls, delays)
		}
	}
}

func TestGitPushTransientHTTPBounded(t *testing.T) {
	for _, status := range []int{500, 502, 503, 504} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			server := plumbing.NewUnexpectedError(&githttp.Err{Response: &http.Response{StatusCode: status}})
			calls := 0
			var total time.Duration
			err := retryGitPush(context.Background(), "https://github.com/testuser/testrepo.git", func() error {
				calls++
				return server
			}, func(_ context.Context, delay time.Duration) error {
				total += delay
				return nil
			})
			if err != server || calls != 5 || total != 15*time.Second {
				t.Fatalf("err=%v calls=%d delay=%v", err, calls, total)
			}
		})
	}
}

func TestGitPushPermanentFailureAndOtherRemotesRemainTerminal(t *testing.T) {
	for _, test := range []struct {
		remote string
		err    error
	}{
		{"https://github.com/testuser/testrepo.git", (&packp.CommandStatus{ReferenceName: "refs/heads/main", Status: "non-fast-forward"}).Error()},
		{"https://github.com/testuser/testrepo.git", errors.New("Internal Server Error in local hook")},
		{"https://github.com/testuser/testrepo.git", gogit.NoErrAlreadyUpToDate},
		{"https://github.com/testuser/testrepo.git", plumbing.NewUnexpectedError(&githttp.Err{Response: &http.Response{StatusCode: 403}})},
		{"https://git.example.test/testrepo.git", (&packp.CommandStatus{ReferenceName: "refs/heads/main", Status: "Internal Server Error"}).Error()},
		{"https://github.com.example.test/testrepo.git", (&packp.CommandStatus{ReferenceName: "refs/heads/main", Status: "Internal Server Error"}).Error()},
	} {
		calls := 0
		err := retryGitPush(context.Background(), test.remote, func() error {
			calls++
			return test.err
		}, func(context.Context, time.Duration) error {
			t.Fatal("terminal push slept")
			return nil
		})
		if err != test.err || calls != 1 {
			t.Fatalf("remote=%s calls=%d", test.remote, calls)
		}
	}
}

func TestGitPushCancellationStopsRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := retryGitPush(ctx, "https://github.com/testuser/testrepo.git", func() error {
		calls++
		return (&packp.CommandStatus{ReferenceName: "refs/heads/main", Status: "Internal Server Error"}).Error()
	}, func(context.Context, time.Duration) error {
		cancel()
		return ctx.Err()
	})
	if err != context.Canceled || calls != 1 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

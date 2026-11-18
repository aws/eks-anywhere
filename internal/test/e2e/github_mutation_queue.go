package e2e

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"k8s.io/utils/clock"
)

// Pace initial setup writes while independent provider drivers share the CI account.
const githubSetupMutationInterval = 5 * time.Second

var githubSetupMutations = newGithubMutationQueue(githubSetupMutationInterval, clock.RealClock{})

type githubMutationQueue struct {
	turn     chan struct{}
	clock    clock.Clock
	interval time.Duration
	next     time.Time
}

func newGithubMutationQueue(interval time.Duration, c clock.Clock) *githubMutationQueue {
	return &githubMutationQueue{turn: make(chan struct{}, 1), clock: c, interval: interval}
}

func (q *githubMutationQueue) acquire(ctx context.Context) (func(), error) {
	select {
	case q.turn <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if err := q.wait(ctx); err != nil {
		<-q.turn
		return nil, err
	}
	return func() {
		q.next = q.clock.Now().Add(q.interval)
		<-q.turn
	}, nil
}

func (q *githubMutationQueue) wait(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	delay := q.next.Sub(q.clock.Now())
	if delay <= 0 {
		return nil
	}
	timer := q.clock.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C():
		return ctx.Err()
	}
}

func githubSetupHTTPClient(ctx context.Context, queue *githubMutationQueue) *http.Client {
	client := http.DefaultClient
	if configured, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok && configured != nil {
		client = configured
	}
	copy := *client
	base := copy.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	copy.Transport = githubSetupMutationTransport{base: base, queue: queue}
	return &copy
}

type githubSetupMutationTransport struct {
	base  http.RoundTripper
	queue *githubMutationQueue
}

func (t githubSetupMutationTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if !isGithubSetupMutation(request) {
		return t.base.RoundTrip(request)
	}
	release, err := t.queue.acquire(request.Context())
	if err != nil {
		if request.Body != nil {
			_ = request.Body.Close()
		}
		return nil, err
	}
	response, err := t.base.RoundTrip(request)
	if err != nil || response == nil || response.Body == nil {
		release()
		return response, err
	}
	// The SDK replaces error bodies after reading them, so EOF must release ownership too.
	response.Body = &githubMutationResponseBody{ReadCloser: response.Body, release: release}
	return response, nil
}

func isGithubSetupMutation(request *http.Request) bool {
	if request.Method != http.MethodPost || !strings.EqualFold(request.URL.Hostname(), "api.github.com") {
		return false
	}
	path := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	switch len(path) {
	case 2:
		return path[0] == "user" && path[1] == "repos"
	case 3:
		return path[0] == "orgs" && path[2] == "repos"
	case 4:
		return path[0] == "repos" && path[3] == "keys"
	default:
		return false
	}
}

type githubMutationResponseBody struct {
	io.ReadCloser
	release func()
	once    sync.Once
}

func (b *githubMutationResponseBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err != nil {
		b.once.Do(b.release)
	}
	return n, err
}

func (b *githubMutationResponseBody) Close() error {
	err := b.ReadCloser.Close()
	b.once.Do(b.release)
	return err
}

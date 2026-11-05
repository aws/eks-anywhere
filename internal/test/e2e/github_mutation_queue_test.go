package e2e

import (
	"context"
	"errors"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

	goGithub "github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
	clocktesting "k8s.io/utils/clock/testing"

	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/git/gogithub"
)

type setupRoundTripper func(*http.Request) (*http.Response, error)

type setupRequestBody struct {
	io.Reader
	closed chan struct{}
}

func (b *setupRequestBody) Close() error {
	close(b.closed)
	return nil
}

func (f setupRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func setupResponse(request *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}},
		Body: io.NopCloser(strings.NewReader(body)), Request: request,
	}
}

func pacedSetupClient(ctx context.Context, base http.RoundTripper, queue *githubMutationQueue) *http.Client {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: base, Timeout: time.Minute})
	return githubSetupHTTPClient(ctx, queue)
}

func setupRequest(t *testing.T, ctx context.Context, method, path string) *http.Request {
	t.Helper()
	request, err := http.NewRequestWithContext(ctx, method, "https://api.github.com"+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func awaitSetupEvent(t *testing.T, events <-chan time.Time) time.Time {
	t.Helper()
	select {
	case event := <-events:
		return event
	case <-time.After(time.Second):
		t.Fatal("expected a setup mutation admission")
		return time.Time{}
	}
}

func awaitSetupTimer(t *testing.T, c *clocktesting.FakeClock, events <-chan time.Time) {
	t.Helper()
	timeout := time.After(time.Second)
	for !c.HasWaiters() {
		select {
		case <-events:
			t.Fatal("setup mutation bypassed the queue's pause")
		case <-timeout:
			t.Fatal("expected a fake-clock pacing timer")
		default:
			runtime.Gosched()
		}
	}
}

func TestGithubSetupMutationsShareQueueAcrossClients(t *testing.T) {
	c := clocktesting.NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	queue := newGithubMutationQueue(5*time.Second, c)
	events := make(chan time.Time, 4)
	results := make(chan error, 4)
	base := setupRoundTripper(func(request *http.Request) (*http.Response, error) {
		events <- c.Now()
		return setupResponse(request, http.StatusCreated, "{}"), nil
	})
	for _, path := range []string{"/user/repos", "/repos/example/first/keys", "/orgs/example/repos", "/repos/example/second/keys"} {
		client := pacedSetupClient(context.Background(), base, queue)
		go func() {
			response, err := client.Do(setupRequest(t, context.Background(), http.MethodPost, path))
			if err == nil {
				err = response.Body.Close()
			}
			results <- err
		}()
	}
	start := awaitSetupEvent(t, events)
	for i := 1; i < 4; i++ {
		awaitSetupTimer(t, c, events)
		c.Step(5*time.Second - time.Nanosecond)
		select {
		case <-events:
			t.Fatal("setup mutation started before the full pause")
		default:
		}
		c.Step(time.Nanosecond)
		got := awaitSetupEvent(t, events)
		if want := start.Add(time.Duration(i) * 5 * time.Second); !got.Equal(want) {
			t.Fatalf("admission=%v, want=%v", got, want)
		}
	}
	for range 4 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
}

func TestGithubSetupQueueLeavesOtherRequestsUnpaced(t *testing.T) {
	queue := newGithubMutationQueue(5*time.Second, clocktesting.NewFakeClock(time.Now()))
	base := setupRoundTripper(func(request *http.Request) (*http.Response, error) {
		return setupResponse(request, http.StatusOK, "{}"), nil
	})
	client := pacedSetupClient(context.Background(), base, queue)
	response, err := client.Do(setupRequest(t, context.Background(), http.MethodPost, "/user/repos"))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	for _, tc := range []struct{ method, url string }{
		{http.MethodGet, "https://api.github.com/repos/example/repo"},
		{http.MethodPost, "https://api.github.com/repos/example/repo/issues"},
		{http.MethodPost, "https://example.invalid/user/repos"},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		request, requestErr := http.NewRequestWithContext(ctx, tc.method, tc.url, nil)
		if requestErr != nil {
			t.Fatal(requestErr)
		}
		unpaced, doErr := client.Do(request)
		cancel()
		if doErr != nil {
			t.Fatalf("%s %s was delayed by the setup queue: %v", tc.method, tc.url, doErr)
		}
		unpaced.Body.Close()
	}
}

func TestGithubSetupQueueCancellationDoesNotConsumeNextSlot(t *testing.T) {
	c := clocktesting.NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	queue := newGithubMutationQueue(5*time.Second, c)
	events := make(chan time.Time, 3)
	base := setupRoundTripper(func(request *http.Request) (*http.Response, error) {
		events <- c.Now()
		return setupResponse(request, http.StatusCreated, "{}"), nil
	})
	client := pacedSetupClient(context.Background(), base, queue)
	response, err := client.Do(setupRequest(t, context.Background(), http.MethodPost, "/user/repos"))
	if err != nil {
		t.Fatal(err)
	}
	start := awaitSetupEvent(t, events)
	response.Body.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancelled := make(chan error, 1)
	go func() {
		r, doErr := client.Do(setupRequest(t, ctx, http.MethodPost, "/user/repos"))
		if doErr == nil {
			r.Body.Close()
		}
		cancelled <- doErr
	}()
	awaitSetupTimer(t, c, events)
	cancel()
	if err := <-cancelled; !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled queue wait returned %v", err)
	}
	completed := make(chan error, 1)
	go func() {
		r, doErr := client.Do(setupRequest(t, context.Background(), http.MethodPost, "/repos/example/repo/keys"))
		if doErr == nil {
			r.Body.Close()
		}
		completed <- doErr
	}()
	awaitSetupTimer(t, c, events)
	c.Step(5 * time.Second)
	if admitted := awaitSetupEvent(t, events); !admitted.Equal(start.Add(5 * time.Second)) {
		t.Fatal("canceled wait consumed a mutation slot")
	}
	if err := <-completed; err != nil {
		t.Fatal(err)
	}
}

func TestGithubSetupQueueClosesCanceledRequestBody(t *testing.T) {
	c := clocktesting.NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	queue := newGithubMutationQueue(5*time.Second, c)
	events := make(chan time.Time, 2)
	base := setupRoundTripper(func(request *http.Request) (*http.Response, error) {
		events <- c.Now()
		return setupResponse(request, http.StatusCreated, "{}"), nil
	})
	transport := githubSetupMutationTransport{base: base, queue: queue}
	first, err := transport.RoundTrip(setupRequest(t, context.Background(), http.MethodPost, "/user/repos"))
	if err != nil {
		t.Fatal(err)
	}
	awaitSetupEvent(t, events)
	first.Body.Close()
	ctx, cancel := context.WithCancel(context.Background())
	body := &setupRequestBody{Reader: strings.NewReader("{}"), closed: make(chan struct{})}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/user/repos", body)
	if err != nil {
		t.Fatal(err)
	}
	finished := make(chan error, 1)
	go func() {
		response, doErr := transport.RoundTrip(request)
		if response != nil {
			response.Body.Close()
		}
		finished <- doErr
	}()
	awaitSetupTimer(t, c, events)
	cancel()
	if err := <-finished; !errors.Is(err, context.Canceled) {
		t.Fatalf("queued request returned %v", err)
	}
	select {
	case <-body.closed:
	default:
		t.Fatal("canceled queue admission leaked the request body")
	}
	select {
	case <-events:
		t.Fatal("canceled request reached the underlying transport")
	default:
	}
}

func TestGithubSetupQueueSerializesUntilResponseCloses(t *testing.T) {
	c := clocktesting.NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	queue := newGithubMutationQueue(5*time.Second, c)
	events := make(chan time.Time, 2)
	base := setupRoundTripper(func(request *http.Request) (*http.Response, error) {
		events <- c.Now()
		return setupResponse(request, http.StatusCreated, "{}"), nil
	})
	client := pacedSetupClient(context.Background(), base, queue)
	first, err := client.Do(setupRequest(t, context.Background(), http.MethodPost, "/user/repos"))
	if err != nil {
		t.Fatal(err)
	}
	defer first.Body.Close()
	awaitSetupEvent(t, events)
	c.Step(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan error, 1)
	go func() {
		response, doErr := client.Do(setupRequest(t, ctx, http.MethodPost, "/user/repos"))
		if doErr == nil {
			response.Body.Close()
		}
		finished <- doErr
	}()
	cancel()
	if err := <-finished; !errors.Is(err, context.Canceled) {
		t.Fatalf("waiting for another response to close ignored cancellation: %v", err)
	}
	select {
	case <-events:
		t.Fatal("another setup mutation started while the first response remained open")
	default:
	}
}

func TestGithubSetupQueueCoversRealSDKAndOAuthTransport(t *testing.T) {
	c := clocktesting.NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	queue := newGithubMutationQueue(5*time.Second, c)
	events := make(chan time.Time, 2)
	base := setupRoundTripper(func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Authorization") != "Bearer example-token" {
			t.Error("the setup transport lost OAuth authentication")
		}
		if request.Method == http.MethodGet {
			return setupResponse(request, http.StatusNotFound, `{"message":"Not Found"}`), nil
		}
		events <- c.Now()
		if request.URL.Path == "/user/repos" {
			return setupResponse(request, http.StatusCreated, `{"name":"repo","owner":{"login":"example"},"clone_url":"https://example.invalid/repo.git"}`), nil
		}
		return setupResponse(request, http.StatusCreated, `{"id":1}`), nil
	})
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, pacedSetupClient(context.Background(), base, queue))
	client := gogithub.New(ctx, gogithub.Options{Auth: git.TokenAuth{Token: "example-token"}})
	_, err := client.CreateRepo(ctx, git.CreateRepoOpts{Name: "repo", Owner: "example", Personal: true})
	if err != nil {
		t.Fatal(err)
	}
	awaitSetupEvent(t, events)
	finished := make(chan error, 1)
	go func() {
		finished <- client.AddDeployKeyToRepo(ctx, git.AddDeployKeyOpts{Owner: "example", Repository: "repo", Key: "example-key"})
	}()
	awaitSetupTimer(t, c, events)
	c.Step(5 * time.Second)
	awaitSetupEvent(t, events)
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
}

func TestGithubSetupQueueReleasesSDKErrorsAfterBodyRead(t *testing.T) {
	c := clocktesting.NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	queue := newGithubMutationQueue(5*time.Second, c)
	events := make(chan time.Time, 2)
	base := setupRoundTripper(func(request *http.Request) (*http.Response, error) {
		events <- c.Now()
		if request.URL.Path == "/user/repos" {
			return setupResponse(request, http.StatusForbidden, `{"message":"Forbidden"}`), nil
		}
		return setupResponse(request, http.StatusCreated, `{"id":1}`), nil
	})
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, pacedSetupClient(context.Background(), base, queue))
	client := gogithub.New(ctx, gogithub.Options{Auth: git.TokenAuth{Token: "example-token"}})
	name := "example"
	_, _, err := client.Client.CreateRepo(ctx, "", &goGithub.Repository{Name: &name})
	if err == nil {
		t.Fatal("expected the SDK to reject the forbidden response")
	}
	awaitSetupEvent(t, events)
	finished := make(chan error, 1)
	go func() {
		finished <- client.AddDeployKeyToRepo(ctx, git.AddDeployKeyOpts{Owner: "example", Repository: "repo", Key: "example-key"})
	}()
	awaitSetupTimer(t, c, events)
	c.Step(5 * time.Second)
	awaitSetupEvent(t, events)
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
}

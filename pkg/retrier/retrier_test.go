package retrier_test

import (
	"errors"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/retrier"
)

func TestNewWithMaxRetriesExhausted(t *testing.T) {
	wantRetries := 10

	r := retrier.NewWithMaxRetries(wantRetries, 0)
	gotRetries := 0
	fn := func() error {
		gotRetries += 1
		return errors.New("")
	}

	err := r.Retry(fn)
	if err == nil {
		t.Fatal("Retrier.Retry() error = nil, want not nil")
	}

	if gotRetries != wantRetries {
		t.Fatalf("Wrong number of retries, got %d, want %d", gotRetries, wantRetries)
	}
}

func TestNewWithMaxRetriesSuccessAfterRetries(t *testing.T) {
	wantRetries := 5

	r := retrier.NewWithMaxRetries(wantRetries, 0)
	gotRetries := 0
	fn := func() error {
		gotRetries += 1
		if wantRetries == gotRetries {
			return nil
		}
		return errors.New("")
	}

	err := r.Retry(fn)
	if err != nil {
		t.Fatalf("Retrier.Retry() error = %v, want nil", err)
	}

	if gotRetries != wantRetries {
		t.Fatalf("Wrong number of retries, got %d, want %d", gotRetries, wantRetries)
	}
}

func TestNewWithNoTimeout(t *testing.T) {
	r := retrier.NewWithNoTimeout()
	fn := func() error {
		return nil
	}

	err := r.Retry(fn)
	if err != nil {
		t.Fatalf("Retrier.Retry() error = %v, want nil", err)
	}
}

func TestRetry(t *testing.T) {
	wantRetries := 5
	gotRetries := 0
	fn := func() error {
		gotRetries += 1
		if wantRetries == gotRetries {
			return nil
		}
		return errors.New("")
	}

	err := retrier.Retry(wantRetries, 0, fn)
	if err != nil {
		t.Fatalf("Retry() error = %v, want nil", err)
	}

	if gotRetries != wantRetries {
		t.Fatalf("Wrong number of retries, got %d, want %d", gotRetries, wantRetries)
	}
}

func TestNewDefaultFinishByFn(t *testing.T) {
	wantRetries := 5

	r := retrier.New(10 * time.Second)
	gotRetries := 0
	fn := func() error {
		gotRetries += 1
		if wantRetries == gotRetries {
			return nil
		}
		return errors.New("")
	}

	err := r.Retry(fn)
	if err != nil {
		t.Fatalf("Retrier.Retry() error = %v, want nil", err)
	}

	if gotRetries != wantRetries {
		t.Fatalf("Wrong number of retries, got %d, want %d", gotRetries, wantRetries)
	}
}

func TestNewDefaultFinishByTimeout(t *testing.T) {
	wantRetries := 100

	r := retrier.New(1 * time.Microsecond)
	gotRetries := 0
	fn := func() error {
		gotRetries += 1
		time.Sleep(2 * time.Microsecond)
		if wantRetries == gotRetries {
			return nil
		}
		return errors.New("")
	}

	err := r.Retry(fn)
	if err == nil {
		t.Fatal("Retrier.Retry() error = nil, want not nil")
	}

	if gotRetries == wantRetries {
		t.Fatalf("Retries shouldn't have got to wantRetries, got and want %d", gotRetries)
	}
}

func TestNewWithRetryPolicyFinishByTimeout(t *testing.T) {
	wantRetries := 100

	retryPolicy := func(totalRetries int, _ error) (bool, time.Duration) {
		return true, (2 * time.Microsecond)
	}

	gotRetries := 0
	fn := func() error {
		gotRetries += 1
		time.Sleep(2 * time.Microsecond)
		if wantRetries == gotRetries {
			return nil
		}
		return errors.New("")
	}

	r := retrier.New(1*time.Microsecond, retrier.WithRetryPolicy(retryPolicy))
	if err := r.Retry(fn); err == nil {
		t.Fatalf("Retrier.Retry() error = nil, want not nil. Got retries = %d", gotRetries)
	}

	if gotRetries == wantRetries {
		t.Fatalf("Retries shouldn't have got to wantRetries, got and want %d", gotRetries)
	}
}

func TestNewWithRetryPolicyFinishByPolicy(t *testing.T) {
	wantRetries := 5

	retryPolicy := func(totalRetries int, _ error) (bool, time.Duration) {
		if totalRetries == wantRetries {
			return false, 0
		}
		return true, 0
	}

	gotRetries := 0
	fn := func() error {
		gotRetries += 1
		return errors.New("")
	}

	r := retrier.New(1*time.Second, retrier.WithRetryPolicy(retryPolicy))
	if err := r.Retry(fn); err == nil {
		t.Fatal("Retrier.Retry() error = nil, want not nil")
	}

	if gotRetries != wantRetries {
		t.Fatalf("Wrong number of retries, got %d, want %d", gotRetries, wantRetries)
	}
}

func TestRetrierWithNilReceiver(t *testing.T) {
	var retrier *retrier.Retrier = nil // This seems improbable, but happens in some other unit tests.

	expectedError := errors.New("my expected error")
	retryable := func() error {
		return expectedError
	}

	err := retrier.Retry(retryable)
	if err == nil || err.Error() != expectedError.Error() {
		t.Errorf("Retrier didn't correctly handle nil receiver")
	}
}

func TestBackOffPolicy(t *testing.T) {
	g := NewWithT(t)
	backOff := time.Second
	p := retrier.BackOffPolicy(backOff)

	retry, gotBackOff := p(10, errors.New(""))
	g.Expect(retry).To(BeTrue())
	g.Expect(gotBackOff).To(Equal(backOff))
}

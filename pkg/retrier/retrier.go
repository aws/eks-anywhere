package retrier

import (
	"math"
	"time"

	"github.com/aws/eks-anywhere/pkg/logger"
)

type Retrier struct {
	retryPolicy   RetryPolicy
	timeout       time.Duration
	backoffFactor *float32
}

type (
	// RetryPolicy allows to customize the retrying logic. The boolean retry indicates if a new retry
	// should be performed and the wait duration indicates the wait time before the next retry.
	RetryPolicy func(totalRetries int, err error) (retry bool, wait time.Duration)
	RetrierOpt  func(*Retrier)
)

// New creates a new retrier with a global timeout (max time allowed for the whole execution)
// The default retry policy is to always retry with no wait time in between retries.
func New(timeout time.Duration, opts ...RetrierOpt) *Retrier {
	r := &Retrier{
		timeout:     timeout,
		retryPolicy: zeroWaitPolicy,
	}
	for _, o := range opts {
		o(r)
	}

	return r
}

// NewWithMaxRetries creates a new retrier with no global timeout and a max retries policy.
func NewWithMaxRetries(maxRetries int, backOffPeriod time.Duration) *Retrier {
	// this value is roughly 292 years, so in practice there is no timeout
	return New(time.Duration(math.MaxInt64), WithMaxRetries(maxRetries, backOffPeriod))
}

// NewWithNoTimeout creates a new retrier with no global timeout and infinite retries.
func NewWithNoTimeout() *Retrier {
	return New(time.Duration(math.MaxInt64))
}

// WithMaxRetries sets a retry policy that will retry up to maxRetries times
// with a wait time between retries of backOffPeriod.
func WithMaxRetries(maxRetries int, backOffPeriod time.Duration) RetrierOpt {
	return func(r *Retrier) {
		r.retryPolicy = maxRetriesPolicy(maxRetries, backOffPeriod)
	}
}

func WithBackoffFactor(factor float32) RetrierOpt {
	return func(r *Retrier) {
		r.backoffFactor = &factor
	}
}

func WithRetryPolicy(policy RetryPolicy) RetrierOpt {
	return func(r *Retrier) {
		r.retryPolicy = policy
	}
}

// Retry runs the fn function until it either successful completes (not error),
// the set timeout reached or the retry policy aborts the execution.
func (r *Retrier) Retry(fn func() error) error {
	// While it seems aberrant to call a method with a nil receiver, several unit tests actually do.  With a previous
	// version of this module (which didn't attempt to dereference the receiver until after the wrapped function failed)
	// these passed.  Changes below, to log the receiver struct's key params changed that breaking the unit tests.
	// The below conditional block restores the original behavior, enabling these tests to again pass.
	if r == nil {
		return fn()
	}

	start := time.Now()
	retries := 0
	var err error
	logger.V(5).Info("Retrier:", "timeout", r.timeout, "backoffFactor", r.backoffFactor)
	for retry := true; retry; retry = time.Since(start) < r.timeout {
		err = fn()
		retries += 1
		if err == nil {
			logger.V(5).Info("Retry execution successful", "retries", retries, "duration", time.Since(start))
			return nil
		}
		logger.V(5).Info("Error happened during retry", "error", err, "retries", retries)

		retry, wait := r.retryPolicy(retries, err)
		if !retry {
			logger.V(5).Info("Execution aborted by retry policy")
			return err
		}
		if r.backoffFactor != nil {
			wait = time.Duration(float32(wait) * (*r.backoffFactor * float32(retries)))
		}

		// If there's not enough time left for the policy-proposed wait, there's no value in waiting that duration
		// before quitting at the bottom of the loop.  Just do it now.
		retrierTimeoutTime := start.Add(r.timeout)
		policyTimeoutTime := time.Now().Add(wait)
		if retrierTimeoutTime.Before(policyTimeoutTime) {
			break
		}

		logger.V(5).Info("Sleeping before next retry", "time", wait)
		time.Sleep(wait)
	}

	logger.V(5).Info("Timeout reached. Returning error", "retries", retries, "duration", time.Since(start), "error", err)

	return err
}

// Retry runs fn with a MaxRetriesPolicy.
func Retry(maxRetries int, backOffPeriod time.Duration, fn func() error) error {
	r := NewWithMaxRetries(maxRetries, backOffPeriod)
	return r.Retry(fn)
}

// BackOffPolicy retries until top level timeout is reached, waiting a
// backoff period in between retries.
func BackOffPolicy(backoff time.Duration) RetryPolicy {
	return func(totalRetries int, _ error) (retry bool, wait time.Duration) {
		return true, backoff
	}
}

func zeroWaitPolicy(_ int, _ error) (retry bool, wait time.Duration) {
	return true, 0
}

func maxRetriesPolicy(maxRetries int, backOffPeriod time.Duration) RetryPolicy {
	return func(totalRetries int, _ error) (retry bool, wait time.Duration) {
		return totalRetries < maxRetries, backOffPeriod
	}
}

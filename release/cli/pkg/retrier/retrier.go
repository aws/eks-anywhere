// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package retrier

import (
	"fmt"
	"time"
)

type Retrier struct {
	retryPolicy RetryPolicy
	timeout     time.Duration
}

type (
	// RetryPolicy allows to customize the retrying logic. The boolean retry indicates if a new retry
	// should be performed and the wait duration indicates the wait time before the next retry.
	RetryPolicy func(totalRetries int, err error) (retry bool, wait time.Duration)
	RetrierOpt  func(*Retrier)
)

// NewRetrier creates a new retrier with a global timeout (max time allowed for the whole execution)
// The default retry policy is to always retry with no wait time in between retries.
func NewRetrier(timeout time.Duration, opts ...RetrierOpt) *Retrier {
	r := &Retrier{
		timeout:     timeout,
		retryPolicy: zeroWaitPolicy,
	}
	for _, opt := range opts {
		opt(r)
	}

	return r
}

func WithRetryPolicy(policy RetryPolicy) RetrierOpt {
	return func(r *Retrier) {
		r.retryPolicy = policy
	}
}

// Retry runs the fn function until it either successful completes (not error),
// the set timeout reached or the retry policy aborts the execution.
func (r *Retrier) Retry(fn func() error) error {
	start := time.Now()
	retries := 0
	var err error
	for retry := true; retry; retry = time.Since(start) < r.timeout {
		err = fn()
		retries += 1
		if err == nil {
			fmt.Printf("Retry execution successful with %d retries in duration %v\n", retries, time.Since(start))
			return nil
		}
		fmt.Printf("Error happened during retry after %d retries: %v\n", retries, err)

		retry, wait := r.retryPolicy(retries, err)
		if !retry {
			fmt.Println("Execution aborted by retry policy")
			return err
		}

		fmt.Printf("Sleeping before next retry: duration - %v\n", wait)
		time.Sleep(wait)
	}

	fmt.Printf("Timeout reached after %d retries in duration %v. Returning error: %v\n", retries, time.Since(start), err)

	return err
}

func zeroWaitPolicy(_ int, _ error) (retry bool, wait time.Duration) {
	return true, 0
}

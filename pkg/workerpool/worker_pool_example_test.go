package workerpool_test

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/workerpool"
)

type Validation func(context.Context) error

// -----------------------------
// Example 1 - Worker Pool

type RunnerWithPool struct {
	validation []Validation
}

func (r RunnerWithPool) RunValidations(ctx context.Context) []error {
	var aggr []error
	errCh := make(chan error)

	pool := workerpool.Spawn(workerpool.Buffer(len(r.validation)))
	defer pool.Stop()

	for _, v := range r.validation {
		pool.Run(func() {
			errCh <- v(ctx)
		})
	}

	for err := range errCh {
		aggr = append(aggr, err)
		if len(aggr) == len(r.validation) {
			close(errCh)
		}
	}

	return aggr
}

// -----------------------------
// Example 2 - Builtin Pool

type RunnerWithBuiltinPool struct {
	validation []Validation
}

func (r RunnerWithBuiltinPool) RunValidations(ctx context.Context) []error {
	var aggr []error
	errCh := make(chan error)

	work := make(chan Validation, len(r.validation))
	defer close(work)

	for i := 0; i < 10; i++ {
		go func() {
			for w := range work {
				w(ctx)
			}
		}()
	}

	for _, v := range r.validation {
		go func(v Validation) {
			errCh <- v(ctx)
		}(v)
	}

	for err := range errCh {
		aggr = append(aggr, err)
		if len(aggr) == len(r.validation) {
			close(errCh)
		}
	}

	return aggr
}

package defaulting

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/errors"
)

// Default is the logic for a default for a type O. It should return a value of O
// whether it updates it or not. When there is an error, return the zero value of O
// and the error.
type Default[O any] func(ctx context.Context, obj O) (O, error)

// Runner allows to compose and run validations/defaults.
type Runner[O any] struct {
	defaults []Default[O]
}

// NewRunner constructs a new Runner.
func NewRunner[O any]() *Runner[O] {
	return &Runner[O]{}
}

// Register adds defaults to the Runner.
func (r *Runner[O]) Register(defaults ...Default[O]) {
	r.defaults = append(r.defaults, defaults...)
}

// RunAll runs all defaults sequentially and returns the updated O. When there are errors,
// it returns the zero value of O and the aggregated errors.
func (r *Runner[O]) RunAll(ctx context.Context, obj O) (O, errors.Aggregate) {
	var allErr []error
	updatedObj := obj

	for _, d := range r.defaults {
		if newObj, err := d(ctx, updatedObj); err != nil {
			allErr = append(allErr, flatten(err)...)
		} else {
			updatedObj = newObj
		}
	}

	if len(allErr) != 0 {
		return *new(O), errors.NewAggregate(allErr)
	}

	return updatedObj, nil
}

// flatten unfolds and flattens errors inside a errors.Aggregate. If err is not
// a errors.Aggregate, it just returns a slice with one single error.
func flatten(err error) []error {
	if agg, ok := err.(errors.Aggregate); ok {
		return errors.Flatten(agg).Errors()
	}

	return []error{err}
}

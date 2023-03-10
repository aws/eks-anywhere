package validation

import (
	"context"
	"reflect"
	"runtime"
	"sync"

	"github.com/aws/eks-anywhere/pkg/errors"
)

// Validatable is anything that can be validated.
type Validatable[O any] interface {
	DeepCopy() O
}

// Validation is the logic for a validation of a type O.
type Validation[O Validatable[O]] func(ctx context.Context, obj O) error

// Runner allows to compose and run validations.
type Runner[O Validatable[O]] struct {
	validations []Validation[O]
	config      *RunnerConfig
}

// RunnerConfig contains the configuration for a Runner.
type RunnerConfig struct {
	maxJobs int
}

// RunnerOpt allows to configure a Runner with optional parameters.
type RunnerOpt func(*RunnerConfig)

// WithMaxJobs sets the maximun number of concurrent routines the runner will use.
func WithMaxJobs(m int) RunnerOpt {
	return func(c *RunnerConfig) {
		c.maxJobs = m
	}
}

// NewRunner constructs a new Runner.
func NewRunner[O Validatable[O]](opts ...RunnerOpt) *Runner[O] {
	r := &Runner[O]{
		config: &RunnerConfig{
			maxJobs: runtime.GOMAXPROCS(0),
		},
	}

	for _, opt := range opts {
		opt(r.config)
	}

	return r
}

// Register adds validations to the Runner.
func (r *Runner[O]) Register(validations ...Validation[O]) {
	r.validations = append(r.validations, validations...)
}

// RunAll runs all validations concurrently and waits until they all finish,
// aggregating the errors if present. obj must not be modified. If it is, this
// indicates a programming error and the method will panic.
func (r *Runner[O]) RunAll(ctx context.Context, obj O) errors.Aggregate {
	copyObj := obj.DeepCopy()
	var allErr []error
	for err := range r.run(ctx, obj) {
		allErr = append(allErr, err)
	}

	if !reflect.DeepEqual(obj, copyObj) {
		panic("validations must not modify the object under validation")
	}

	return errors.NewAggregate(allErr)
}

func (r *Runner[O]) run(ctx context.Context, obj O) <-chan error {
	results := make(chan error)
	validations := make(chan Validation[O])
	var wg sync.WaitGroup
	numWorkers := r.config.maxJobs
	if numWorkers > len(r.validations) {
		numWorkers = len(r.validations)
	}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			for validate := range validations {
				if err := validate(ctx, obj); err != nil {
					for _, err := range flatten(err) {
						results <- err
					}
				}
			}
			wg.Done()
		}()
	}

	go func() {
		for _, v := range r.validations {
			validations <- v
		}
		close(validations)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

// Sequentially composes a set of validations into one which will run them sequentially and in order.
func Sequentially[O Validatable[O]](validations ...Validation[O]) Validation[O] {
	return func(ctx context.Context, obj O) error {
		var allErr []error
		for _, h := range validations {
			if err := h(ctx, obj); err != nil {
				allErr = append(allErr, flatten(err)...)
			}
		}

		return errors.NewAggregate(allErr)
	}
}

// flatten unfolds and flattens errors inside a errors.Aggregate. If err is not
// a errors.Aggregate, it just returns a slice with one single error.
func flatten(err error) []error {
	if agg, ok := err.(errors.Aggregate); ok {
		return errors.Flatten(agg).Errors()
	}

	return []error{err}
}

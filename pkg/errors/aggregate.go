package errors

import "k8s.io/apimachinery/pkg/util/errors"

// Aggregate represents an object that contains multiple errors, but does not necessarily have singular semantic meaning.
// The aggregate can be used with `errors.Is()` to check for the occurrence of a specific error type.
// Errors.As() is not supported, because the caller presumably cares about a specific error of potentially multiple that match the given type.
type Aggregate errors.Aggregate

// NewAggregate converts a slice of errors into an Aggregate interface, which
// is itself an implementation of the error interface.  If the slice is empty,
// this returns nil.
// It will check if any of the element of input error list is nil, to avoid
// nil pointer panic when call Error().
func NewAggregate(errList []error) Aggregate {
	return errors.NewAggregate(errList)
}

// Flatten takes an Aggregate, which may hold other Aggregates in arbitrary
// nesting, and flattens them all into a single Aggregate, recursively.
func Flatten(agg Aggregate) Aggregate {
	return errors.Flatten(agg)
}

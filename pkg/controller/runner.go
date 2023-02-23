package controller

import (
	"context"

	"github.com/go-logr/logr"
)

// Phase represents a generic reconciliation phase for a cluster spec.
type Phase[O any] func(ctx context.Context, log logr.Logger, obj O) (Result, error)

// PhaseRunner allows to execute Phases in order.
type PhaseRunner[O any] struct {
	phases []Phase[O]
}

// NewPhaseRunner creates a new PhaseRunner without any Phases.
func NewPhaseRunner[O any]() PhaseRunner[O] {
	return PhaseRunner[O]{}
}

// Register adds a phase to the runnner.
func (r PhaseRunner[O]) Register(phases ...Phase[O]) PhaseRunner[O] {
	r.phases = append(r.phases, phases...)
	return r
}

// Run will execute phases in the order they were registered until a phase
// returns an error or a Result that requests to an interruption.
func (r PhaseRunner[O]) Run(ctx context.Context, log logr.Logger, obj O) (Result, error) {
	for _, p := range r.phases {
		if r, err := p(ctx, log, obj); r.Return() {
			return r, nil
		} else if err != nil {
			return Result{}, err
		}
	}

	return Result{}, nil
}

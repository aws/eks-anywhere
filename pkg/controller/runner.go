package controller

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

// Phase represents a generic reconciliation phase for a cluster spec.
type Phase func(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (Result, error)

// PhaseRunner allows to execute Phases in order.
type PhaseRunner struct {
	phases []Phase
}

// NewPhaseRunner creates a new PhaseRunner without any Phases.
func NewPhaseRunner() PhaseRunner {
	return PhaseRunner{}
}

// Register adds a phase to the runnner.
func (r PhaseRunner) Register(phases ...Phase) PhaseRunner {
	r.phases = append(r.phases, phases...)
	return r
}

// Run will execute phases in the order they were registered until a phase
// returns an error or a Result that requests to an interruption.
func (r PhaseRunner) Run(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (Result, error) {
	for _, p := range r.phases {
		if r, err := p(ctx, log, clusterSpec); r.Return() {
			return r, nil
		} else if err != nil {
			return Result{}, err
		}
	}

	return Result{}, nil
}

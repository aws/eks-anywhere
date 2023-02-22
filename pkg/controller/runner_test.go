package controller_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
)

func TestPhaseRunnerRunError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	phase1 := newPhase()
	phase3 := newPhase()
	r := controller.NewPhaseRunner[*cluster.Spec]().Register(
		phase1.run,
		phaseReturnError,
		phase3.run,
	)

	_, err := r.Run(ctx, test.NewNullLogger(), &cluster.Spec{})
	g.Expect(err).To(HaveOccurred())
	g.Expect(phase1.executed).To(BeTrue())
	g.Expect(phase3.executed).To(BeFalse())
}

func TestPhaseRunnerRunRequeue(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	phase1 := newPhase()
	phase3 := newPhase()
	r := controller.NewPhaseRunner[*cluster.Spec]().Register(
		phase1.run,
		phaseReturnRequeue,
		phase3.run,
	)

	result, err := r.Run(ctx, test.NewNullLogger(), &cluster.Spec{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase1.executed).To(BeTrue())
	g.Expect(phase3.executed).To(BeFalse())
	g.Expect(result.ToCtrlResult().RequeueAfter).To(Equal(1 * time.Second))
}

func TestPhaseRunnerRunAllPhasesFinished(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	phase1 := newPhase()
	phase2 := newPhase()
	phase3 := newPhase()
	r := controller.NewPhaseRunner[*cluster.Spec]().Register(
		phase1.run,
		phase2.run,
		phase3.run,
	)

	result, err := r.Run(ctx, test.NewNullLogger(), &cluster.Spec{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase1.executed).To(BeTrue())
	g.Expect(phase2.executed).To(BeTrue())
	g.Expect(phase3.executed).To(BeTrue())
	g.Expect(result.Result).To(BeNil())
}

func newPhase() *phase {
	return &phase{}
}

type phase struct {
	executed bool
}

func (p *phase) run(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	p.executed = true
	return controller.Result{}, nil
}

func phaseReturnError(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	return controller.Result{}, errors.New("running phase")
}

func phaseReturnRequeue(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	return controller.ResultWithRequeue(1 * time.Second), nil
}

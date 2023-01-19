package validation_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	eksaerrors "github.com/aws/eks-anywhere/pkg/errors"
	"github.com/aws/eks-anywhere/pkg/validation"
)

func TestRunnerRunAllSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	r := validation.NewRunner[*apiCluster]()
	r.Register(
		func(ctx context.Context, cluster *apiCluster) error {
			if cluster.controlPlaneCount == 0 {
				return errors.New("controlPlaneCount can't be 0")
			}

			return nil
		},
		func(ctx context.Context, cluster *apiCluster) error {
			if cluster.bundlesName == "" {
				return errors.New("bundlesName can't be empty")
			}

			return nil
		},
	)

	cluster := &apiCluster{
		controlPlaneCount: 3,
		bundlesName:       "bundles-1",
	}

	g.Expect(r.RunAll(ctx, cluster)).To(Succeed())
}

func TestRunnerRunAllSequentially(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	r := validation.NewRunner[*apiCluster](validation.WithMaxJobs(10))

	callCounter := 0

	r.Register(
		func(ctx context.Context, cluster *apiCluster) error {
			if cluster.bundlesName == "" {
				return errors.New("bundlesName can't be empty")
			}

			return nil
		},
		validation.Sequentially(
			func(ctx context.Context, _ *apiCluster) error {
				g.Expect(callCounter).To(Equal(0))
				callCounter++
				return errors.New("invalid 1")
			},
			func(ctx context.Context, _ *apiCluster) error {
				g.Expect(callCounter).To(Equal(1))
				callCounter++
				return errors.New("invalid 2")
			},
		),
	)

	cluster := &apiCluster{
		controlPlaneCount: 0,
		bundlesName:       "bundles-1",
	}
	err := r.RunAll(ctx, cluster)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("invalid 1")))
	g.Expect(err).To(MatchError(ContainSubstring("invalid 2")))
	g.Expect(callCounter).To(Equal(2))
}

func TestRunnerRunAllAggregatedError(t *testing.T) {
	e1 := errors.New("first error")
	e2 := errors.New("second error")
	e3 := errors.New("third error")

	g := NewWithT(t)
	ctx := context.Background()
	r := validation.NewRunner[*apiCluster]()
	r.Register(
		func(ctx context.Context, _ *apiCluster) error {
			return eksaerrors.NewAggregate([]error{e1, e2})
		},
		func(ctx context.Context, _ *apiCluster) error {
			return e3
		},
	)

	cluster := &apiCluster{}

	err := r.RunAll(ctx, cluster)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Errors()).To(ConsistOf(e1, e2, e3))
}

func TestRunnerRunAllPanicAfterModifyingObject(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	r := validation.NewRunner[*apiCluster]()
	r.Register(
		func(ctx context.Context, _ *apiCluster) error {
			return nil
		},
		func(ctx context.Context, cluster *apiCluster) error {
			cluster.controlPlaneCount = 5
			return nil
		},
	)

	cluster := &apiCluster{}
	run := func() {
		_ = r.RunAll(ctx, cluster)
	}
	g.Expect(run).To(PanicWith("validations must not modify the object under validation"))
}

type apiCluster struct {
	controlPlaneCount int
	bundlesName       string
}

func (a *apiCluster) DeepCopy() *apiCluster {
	copy := *a
	return &copy
}

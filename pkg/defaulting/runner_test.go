package defaulting_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/defaulting"
	eksaerrors "github.com/aws/eks-anywhere/pkg/errors"
)

func TestRunnerRunAll(t *testing.T) {
	g := NewWithT(t)
	r := defaulting.NewRunner[apiCluster]()

	r.Register(
		func(ctx context.Context, cluster apiCluster) (apiCluster, error) {
			if cluster.bundlesName == "" {
				cluster.bundlesName = "bundles-1"
			}
			return cluster, nil
		},
		func(ctx context.Context, cluster apiCluster) (apiCluster, error) {
			if cluster.controlPlaneCount == 0 {
				cluster.controlPlaneCount = 3
			}
			return cluster, nil
		},
	)

	ctx := context.Background()
	cluster := apiCluster{}
	newCluster, err := r.RunAll(ctx, cluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(newCluster.bundlesName).To(Equal("bundles-1"))
	g.Expect(newCluster.controlPlaneCount).To(Equal(3))
}

func TestRunnerRunAllError(t *testing.T) {
	g := NewWithT(t)
	e1 := errors.New("first error")
	e2 := errors.New("second error")
	e3 := errors.New("third error")
	r := defaulting.NewRunner[apiCluster]()

	r.Register(
		func(ctx context.Context, cluster apiCluster) (apiCluster, error) {
			return apiCluster{}, eksaerrors.NewAggregate([]error{e1, e2})
		},
		func(ctx context.Context, cluster apiCluster) (apiCluster, error) {
			return cluster, e3
		},
	)

	ctx := context.Background()
	cluster := apiCluster{}
	g.Expect(r.RunAll(ctx, cluster)).Error().To(And(
		MatchError(ContainSubstring("first error")),
		MatchError(ContainSubstring("second error")),
		MatchError(ContainSubstring("third error")),
	))
}

type apiCluster struct {
	controlPlaneCount int
	bundlesName       string
}

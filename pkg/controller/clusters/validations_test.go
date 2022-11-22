package clusters_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestCleanupStatusAfterValidate(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Status.FailureMessage = ptr.String("invalid cluster")
	})

	g.Expect(
		clusters.CleanupStatusAfterValidate(context.Background(), test.NewNullLogger(), spec),
	).To(Equal(controller.Result{}))
	g.Expect(spec.Cluster.Status.FailureMessage).To(BeNil())
}

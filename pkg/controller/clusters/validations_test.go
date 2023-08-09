package clusters_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func TestCleanupStatusAfterValidate(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.SetFailure(anywherev1.FailureReasonType("InvalidCluster"), "invalid cluster")
	})

	g.Expect(
		clusters.CleanupStatusAfterValidate(context.Background(), test.NewNullLogger(), spec),
	).To(Equal(controller.Result{}))
	g.Expect(spec.Cluster.Status.FailureMessage).To(BeNil())
	g.Expect(spec.Cluster.Status.FailureReason).To(BeNil())
}

func TestValidateManagementClusterNameSuccess(t *testing.T) {
	tt := newClusterValidatorTest(t)

	objs := []runtime.Object{tt.cluster, tt.managementCluster}
	cb := fakeClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	validator := clusters.NewClusterValidator(cl)
	tt.Expect(validator.ValidateManagementClusterName(context.Background(), tt.logger, tt.cluster)).To(BeNil())
}

func TestValidateManagementClusterNameDifferentNamespaceSuccess(t *testing.T) {
	tt := newClusterValidatorTest(t)
	tt.managementCluster.Namespace = "different-namespace"

	objs := []runtime.Object{tt.cluster, tt.managementCluster}
	cb := fakeClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	validator := clusters.NewClusterValidator(cl)
	tt.Expect(validator.ValidateManagementClusterName(context.Background(), tt.logger, tt.cluster)).To(BeNil())
}

func TestValidateManagementClusterNameMissing(t *testing.T) {
	tt := newClusterValidatorTest(t)

	tt.cluster.Spec.ManagementCluster.Name = "missing"
	objs := []runtime.Object{tt.cluster, tt.managementCluster}
	cb := fakeClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	validator := clusters.NewClusterValidator(cl)
	err := validator.ValidateManagementClusterName(context.Background(), tt.logger, tt.cluster)
	tt.Expect(err.Error()).
		To(Equal("unable to retrieve management cluster missing: clusters.anywhere.eks.amazonaws.com \"missing\" not found"))
}

func TestValidateManagementClusterNameMultiple(t *testing.T) {
	tt := newClusterValidatorTest(t)

	tt.managementCluster.Namespace = "different-namespace-1"
	managementCluster2 := tt.managementCluster.DeepCopy()
	managementCluster2.Namespace = "different-namespace-2"
	objs := []runtime.Object{tt.cluster, tt.managementCluster, managementCluster2}
	cb := fakeClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	validator := clusters.NewClusterValidator(cl)
	err := validator.ValidateManagementClusterName(context.Background(), tt.logger, tt.cluster)
	tt.Expect(err.Error()).To(Equal("found multiple clusters with the name my-management-cluster"))
}

func TestValidateManagementClusterNameInvalid(t *testing.T) {
	tt := newClusterValidatorTest(t)

	tt.managementCluster.SetManagedBy("differentCluster")
	objs := []runtime.Object{tt.cluster, tt.managementCluster}
	cb := fakeClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	validator := clusters.NewClusterValidator(cl)
	tt.Expect(validator.ValidateManagementClusterName(context.Background(), tt.logger, tt.cluster)).
		To(MatchError(errors.New("my-management-cluster is not a valid management cluster")))
}

type clusterValidatorTest struct {
	*WithT
	logger            logr.Logger
	cluster           *anywherev1.Cluster
	managementCluster *anywherev1.Cluster
}

func newClusterValidatorTest(t *testing.T) *clusterValidatorTest {
	logger := test.NewNullLogger()
	managementCluster, cluster := createClustersForTest()
	return &clusterValidatorTest{
		WithT:             NewWithT(t),
		logger:            logger,
		cluster:           cluster,
		managementCluster: managementCluster,
	}
}

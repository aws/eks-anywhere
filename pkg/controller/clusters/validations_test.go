package clusters_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

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
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	validator := clusters.NewClusterValidator(cl)
	tt.Expect(validator.ValidateManagementClusterName(context.Background(), tt.logger, tt.cluster)).To(BeNil())
}

func TestValidateManagementClusterNameMissing(t *testing.T) {
	tt := newClusterValidatorTest(t)

	tt.cluster.Spec.ManagementCluster.Name = "missing"
	objs := []runtime.Object{tt.cluster, tt.managementCluster}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	validator := clusters.NewClusterValidator(cl)
	tt.Expect(validator.ValidateManagementClusterName(context.Background(), tt.logger, tt.cluster)).
		To(MatchError(errors.New("unable to retrieve management cluster missing: clusters.anywhere.eks.amazonaws.com \"missing\" not found")))
}

func TestValidateManagementClusterNameInvalid(t *testing.T) {
	tt := newClusterValidatorTest(t)

	tt.managementCluster.SetManagedBy("differentCluster")
	objs := []runtime.Object{tt.cluster, tt.managementCluster}
	cb := fake.NewClientBuilder()
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
	version := anywherev1.EksaVersion("v0.0.0")
	logger := test.NewNullLogger()
	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "my-namespace",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "my-namespace",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
	}
	cluster.SetManagedBy("my-management-cluster")
	return &clusterValidatorTest{
		WithT:             NewWithT(t),
		logger:            logger,
		cluster:           cluster,
		managementCluster: managementCluster,
	}
}

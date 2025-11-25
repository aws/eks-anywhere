package clusters_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/controller/clusters/mocks"
)

func TestValidateControlPlaneIPSuccess(t *testing.T) {
	tt := newIPValidatorTest(t)
	client := fake.NewClientBuilder().Build()
	ipUniquenessValidator := mocks.NewMockIPUniquenessValidator(gomock.NewController(t))
	ipUniquenessValidator.EXPECT().ValidateControlPlaneIPUniqueness(tt.testCluster).Return(nil)
	ipValidator := clusters.NewIPValidator(ipUniquenessValidator, client)
	tt.Expect(ipValidator.ValidateControlPlaneIP(context.Background(), tt.logger, tt.spec)).To(Equal(controller.Result{}))
	tt.Expect(tt.spec.Cluster.Status.FailureMessage).To(BeNil())
	tt.Expect(tt.spec.Cluster.Status.FailureReason).To(BeNil())
}

func TestValidateControlPlaneIPUnavailable(t *testing.T) {
	tt := newIPValidatorTest(t)
	client := fake.NewClientBuilder().Build()
	ipUniquenessValidator := mocks.NewMockIPUniquenessValidator(gomock.NewController(t))
	ipUniquenessValidator.EXPECT().ValidateControlPlaneIPUniqueness(tt.testCluster).Return(errors.New("already in use"))
	ipValidator := clusters.NewIPValidator(ipUniquenessValidator, client)
	tt.Expect(ipValidator.ValidateControlPlaneIP(context.Background(), tt.logger, tt.spec)).To(Equal(controller.ResultWithReturn()))
	tt.Expect(tt.spec.Cluster.Status.FailureMessage).To(HaveValue(ContainSubstring("already in use")))
	tt.Expect(tt.spec.Cluster.Status.FailureReason).To(HaveValue(Equal(anywherev1.UnavailableControlPlaneIPReason)))
}

func TestValidateControlPlaneIPCapiClusterExists(t *testing.T) {
	tt := newIPValidatorTest(t)
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = "test-cluster"
	})
	client := fake.NewClientBuilder().WithObjects(capiCluster).Build()
	ipUniquenessValidator := mocks.NewMockIPUniquenessValidator(gomock.NewController(t))
	ipValidator := clusters.NewIPValidator(ipUniquenessValidator, client)
	tt.Expect(ipValidator.ValidateControlPlaneIP(context.Background(), tt.logger, tt.spec)).To(Equal(controller.Result{}))
	tt.Expect(tt.spec.Cluster.Status.FailureMessage).To(BeNil())
	tt.Expect(tt.spec.Cluster.Status.FailureReason).To(BeNil())
}

type ipValidatorTest struct {
	t testing.TB
	*WithT
	logger      logr.Logger
	spec        *cluster.Spec
	testCluster *anywherev1.Cluster
}

func newIPValidatorTest(t testing.TB) *ipValidatorTest {
	logger := test.NewNullLogger()
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &anywherev1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: constants.EksaSystemNamespace,
			},
			Spec: anywherev1.ClusterSpec{
				ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
					Endpoint: &anywherev1.Endpoint{
						Host: "test-ip",
					},
				},
			},
		}
	})
	testCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.EksaSystemNamespace,
			Name:      "test-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Endpoint: &anywherev1.Endpoint{
					Host: "test-ip",
				},
			},
		},
	}

	tt := &ipValidatorTest{
		t:           t,
		WithT:       NewWithT(t),
		logger:      logger,
		spec:        spec,
		testCluster: testCluster,
	}
	return tt
}

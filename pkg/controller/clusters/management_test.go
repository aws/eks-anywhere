package clusters_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func TestFetchManagementEksaClusterSuccess(t *testing.T) {
	tt := newFetchManagementTest(t)

	objs := []runtime.Object{tt.cluster, tt.managementCluster}
	cb := fakeClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	managementCluster, err := clusters.FetchManagementEksaCluster(context.Background(), cl, tt.cluster)
	tt.Expect(err).To(BeNil())
	tt.Expect(managementCluster).To(Equal(tt.managementCluster))
}

func TestFetchManagementEksaClusterDifferentNamespaceSuccess(t *testing.T) {
	tt := newFetchManagementTest(t)
	tt.managementCluster.Namespace = "different-namespace"

	objs := []runtime.Object{tt.cluster, tt.managementCluster}
	cb := fakeClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	managementCluster, err := clusters.FetchManagementEksaCluster(context.Background(), cl, tt.cluster)
	tt.Expect(err).To(BeNil())
	tt.Expect(managementCluster).To(Equal(tt.managementCluster))
}

func TestFetchManagementEksaClusterMissing(t *testing.T) {
	tt := newFetchManagementTest(t)

	tt.cluster.Spec.ManagementCluster.Name = "missing"
	objs := []runtime.Object{tt.cluster, tt.managementCluster}
	cb := fakeClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	managementCluster, err := clusters.FetchManagementEksaCluster(context.Background(), cl, tt.cluster)
	tt.Expect(err.Error()).
		To(Equal("unable to retrieve management cluster missing: clusters.anywhere.eks.amazonaws.com \"missing\" not found"))
	tt.Expect(managementCluster).To(BeNil())
}

func TestFetchManagementEksaClusterMultiple(t *testing.T) {
	tt := newFetchManagementTest(t)

	tt.managementCluster.Namespace = "different-namespace-1"
	managementCluster2 := tt.managementCluster.DeepCopy()
	managementCluster2.Namespace = "different-namespace-2"
	objs := []runtime.Object{tt.cluster, tt.managementCluster, managementCluster2}
	cb := fakeClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	managementCluster, err := clusters.FetchManagementEksaCluster(context.Background(), cl, tt.cluster)
	tt.Expect(err.Error()).To(Equal("found multiple clusters with the name my-management-cluster"))
	tt.Expect(managementCluster).To(BeNil())
}

type fetchManagementTest struct {
	*WithT
	cluster           *anywherev1.Cluster
	managementCluster *anywherev1.Cluster
}

func newFetchManagementTest(t *testing.T) *fetchManagementTest {
	managementCluster, cluster := createClustersForTest()
	cluster.SetManagedBy("my-management-cluster")
	return &fetchManagementTest{
		WithT:             NewWithT(t),
		cluster:           cluster,
		managementCluster: managementCluster,
	}
}

func createClustersForTest() (managementCluster *anywherev1.Cluster, workloadCluster *anywherev1.Cluster) {
	managementCluster = &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "my-namespace",
		},
	}

	workloadCluster = &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "my-namespace",
		},
	}
	workloadCluster.SetManagedBy("my-management-cluster")
	return managementCluster, workloadCluster
}

func fakeClientBuilder() *fake.ClientBuilder {
	return fake.NewClientBuilder().WithIndex(&anywherev1.Cluster{}, "metadata.name", clientutil.ClusterNameIndexer)
}

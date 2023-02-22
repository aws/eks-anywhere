package cluster_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestParseConfigMissingCloudstackDatacenter(t *testing.T) {
	g := NewWithT(t)
	got, err := cluster.ParseConfigFromFile("testdata/cluster_cloudstack_missing_datacenter.yaml")

	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(got.CloudStackDatacenter).To(BeNil())
}

func TestDefaultConfigClientBuilderBuildCloudStackClusterSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	b := cluster.NewDefaultConfigClientBuilder()
	cluster := cloudStackCluster()
	datacenter := &anywherev1.CloudStackDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.CloudStackDatacenterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: "default",
		},
	}

	client := test.NewFakeKubeClient(datacenter)
	config, err := b.Build(ctx, client, cluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.Cluster).To(Equal(cluster))
	g.Expect(config.CloudStackDatacenter).To(Equal(datacenter))
}

func TestDefaultConfigClientBuilderBuildCloudStackClusterFailure(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	b := cluster.NewDefaultConfigClientBuilder()
	cluster := cloudStackCluster()
	client := test.NewFakeKubeClient()
	_, err := b.Build(ctx, client, cluster)
	g.Expect(err).To(MatchError(ContainSubstring("cloudstackdatacenterconfigs.anywhere.eks.amazonaws.com \"datacenter\" not found")))
}

func cloudStackCluster() *anywherev1.Cluster {
	return &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			DatacenterRef: anywherev1.Ref{
				Kind: anywherev1.CloudStackDatacenterKind,
				Name: "datacenter",
			},
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Count: 1,
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Name: "md-0",
				},
			},
		},
	}
}

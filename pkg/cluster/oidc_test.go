package cluster_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/cluster/mocks"
)

func TestDefaultConfigClientBuilderOIDC(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	b := cluster.NewDefaultConfigClientBuilder()
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			IdentityProviderRefs: []anywherev1.Ref{
				{
					Kind: anywherev1.OIDCConfigKind,
					Name: "my-oidc",
				},
			},
		},
	}
	oidcConfig := &anywherev1.OIDCConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-oidc",
			Namespace: "default",
		},
	}

	client.EXPECT().Get(ctx, "my-oidc", "default", &anywherev1.OIDCConfig{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			c := obj.(*anywherev1.OIDCConfig)
			c.ObjectMeta = oidcConfig.ObjectMeta
			return nil
		},
	)

	config, err := b.Build(ctx, client, cluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.Cluster).To(Equal(cluster))
	g.Expect(len(config.OIDCConfigs)).To(Equal(1))
	g.Expect(config.OIDCConfigs["my-oidc"]).To(Equal(oidcConfig))
}

func TestConfigManagerValidateOIDCConfigSuccess(t *testing.T) {
	g := NewWithT(t)
	c := clusterConfigFromFile(t, "testdata/docker_cluster_oidc_awsiam_flux.yaml")
	m, err := cluster.NewDefaultConfigManager()
	g.Expect(err).To(BeNil())

	err = m.Validate(c)
	g.Expect(err).To(Succeed())
}

func TestConfigManagerValidateOIDCConfigMultipleErrors(t *testing.T) {
	g := NewWithT(t)
	c := clusterConfigFromFile(t, "testdata/docker_cluster_oidc_awsiam_flux.yaml")
	c.OIDCConfigs["eksa-unit-test"] = &anywherev1.OIDCConfig{
		Spec: anywherev1.OIDCConfigSpec{
			ClientId: "",
		},
	}
	m, err := cluster.NewDefaultConfigManager()
	g.Expect(err).To(BeNil())

	err = m.Validate(c)
	g.Expect(err).To(MatchError(ContainSubstring("clientId is required")))
}

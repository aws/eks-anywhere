package helm_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/helm"
)

func TestHelmFactoryGetClientForClusterSuccess(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: c.Name,
		}
	})

	client := test.NewFakeKubeClient()
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithKubeClient(client).
		WithHelmFactory(
			helm.WithRegistryMirror(nil),
			helm.WithEnv(map[string]string{}),
			helm.WithInsecure(),
		).
		Build(context.Background())

	g.Expect(err).To(BeNil())
	g.Expect(deps.HelmFactory).ToNot(BeNil())

	helm, err := deps.HelmFactory.GetClientForCluster(ctx, cluster)

	g.Expect(err).To(BeNil())
	g.Expect(helm).NotTo(BeNil())
}

func TestHelmFactoryGetClientForClusterErrorManagmentClusterNotFound(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "management-cluster",
		}
	})

	client := test.NewFakeKubeClient(cluster)

	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithKubeClient(client).
		WithHelmFactory().
		Build(context.Background())

	g.Expect(err).To(BeNil())

	helm, err := deps.HelmFactory.GetClientForCluster(ctx, cluster)

	g.Expect(helm).To(BeNil())
	g.Expect(err).To(MatchError(ContainSubstring("\"management-cluster\" not found")))
}

func TestHelmFactoryGetClientForClusterRegistryMirrorErrorGettingSecret(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "test-cluster",
		}
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Authenticate: true,
			Endpoint:     "1.2.3.4",
			Port:         "65536",
		}
	})

	client := test.NewFakeKubeClientAlwaysError()
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithKubeClient(client).
		WithHelmFactory().
		Build(context.Background())

	g.Expect(err).To(BeNil())

	_, err = deps.HelmFactory.GetClientForCluster(ctx, cluster)
	g.Expect(err).To(MatchError(ContainSubstring("fetching registry auth secret: no kind is registered for the type v1.Secret")))
}

func TestHelmFactoryGetClientForClusterSuccessRegistryMirrorSecretCredentials(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "test-cluster",
		}
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Authenticate: true,
			Endpoint:     "1.2.3.4",
			Port:         "65536",
		}
	})

	registryAuthSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "registry-credentials",
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{
			"username": []byte("username"),
			"password": []byte("password"),
		},
	}

	client := test.NewFakeKubeClient(cluster, registryAuthSecret)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithKubeClient(client).
		WithHelmFactory().
		Build(context.Background())

	g.Expect(err).To(BeNil())

	helmClient, err := deps.HelmFactory.GetClientForCluster(ctx, cluster)
	g.Expect(err).To(BeNil())
	g.Expect(helmClient).ToNot(BeNil())
}

func TestHelmFactoryGetClientForClusterRegistryMirrorErrorNoRegistryCredentials(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	managmentCluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "management-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "management-cluster",
		}
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Authenticate: true,
			Endpoint:     "1.2.3.4",
			Port:         "65536",
		}
	})

	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "management-cluster",
		}
	})

	client := test.NewFakeKubeClient(managmentCluster, cluster)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithKubeClient(client).
		WithHelmFactory().
		Build(context.Background())

	g.Expect(err).To(BeNil())
	_, err = deps.HelmFactory.GetClientForCluster(ctx, cluster)
	g.Expect(err).To(MatchError(ContainSubstring("please set REGISTRY_USERNAME")))
}

func TestHelmFactoryGetClientForClusterSuccessRegistryMirrorEnvCredendialss(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	managmentCluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "management-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "management-cluster",
		}
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Authenticate: true,
			Endpoint:     "1.2.3.4",
			Port:         "65536",
		}
	})

	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "management-cluster",
		}
	})

	client := test.NewFakeKubeClient(managmentCluster, cluster)
	deps, err := dependencies.NewFactory().
		WithLocalExecutables().
		WithKubeClient(client).
		WithHelmFactory().
		Build(context.Background())

	g.Expect(err).To(BeNil())

	t.Setenv("REGISTRY_USERNAME", "username")
	t.Setenv("REGISTRY_PASSWORD", "password")

	helmClient, err := deps.HelmFactory.GetClientForCluster(ctx, cluster)
	g.Expect(err).To(BeNil())
	g.Expect(helmClient).ToNot(BeNil())
}

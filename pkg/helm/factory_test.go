package helm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/helm"
	helmmocks "github.com/aws/eks-anywhere/pkg/helm/mocks"
)

type helmFactoryTest struct {
	*WithT
	ctx     context.Context
	builder *helmmocks.MockExecutableBuilder
	helm    *helmmocks.MockExecuteableClient
}

func newHelmFactoryTest(t *testing.T) *helmFactoryTest {
	ctrl := gomock.NewController(t)
	builder := helmmocks.NewMockExecutableBuilder(ctrl)
	helm := helmmocks.NewMockExecuteableClient(ctrl)
	return &helmFactoryTest{
		WithT:   NewWithT(t),
		ctx:     context.Background(),
		builder: builder,
		helm:    helm,
	}
}

func TestHelmFactoryGetClientForClusterSuccess(t *testing.T) {
	tt := newHelmFactoryTest(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: c.Name,
		}
	})

	client := test.NewFakeKubeClient(cluster)
	helmFactory := helm.NewClientFactory(client, tt.builder)

	tt.builder.EXPECT().BuildHelmExecutable(gomock.Any()).Return(tt.helm)

	helm, err := helmFactory.GetClientForCluster(tt.ctx, cluster)

	tt.Expect(err).To(BeNil())
	tt.Expect(helm).NotTo(BeNil())
}

func TestHelmFactoryGetClientForClusterErrorManagmentClusterNotFound(t *testing.T) {
	tt := newHelmFactoryTest(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "management-cluster",
		}
	})

	client := test.NewFakeKubeClient(cluster)
	helmFactory := helm.NewClientFactory(client, tt.builder)

	helm, err := helmFactory.GetClientForCluster(tt.ctx, cluster)

	tt.Expect(helm).To(BeNil())
	tt.Expect(err).To(MatchError(ContainSubstring("\"management-cluster\" not found")))
}

func TestHelmFactoryGetClientForClusterAuthenticatedRegistryMirrorErrorGettingSecret(t *testing.T) {
	tt := newHelmFactoryTest(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "test-cluster",
		}
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Authenticate: true,
			Endpoint:     "1.2.3.4",
			Port:         "5000",
		}
	})

	client := test.NewFakeKubeClientAlwaysError()
	helmFactory := helm.NewClientFactory(client, tt.builder)

	_, err := helmFactory.GetClientForCluster(tt.ctx, cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("fetching registry auth secret: no kind is registered for the type v1.Secret")))
}

func TestHelmFactoryGetClientForClusterSuccessAuthenticatedRegistryMirror(t *testing.T) {
	tt := newHelmFactoryTest(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "test-cluster",
		}
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Authenticate: true,
			Endpoint:     "1.2.3.4",
			Port:         "5000",
		}
	})

	rUsername := "username"
	rPassword := "password"
	registryAuthSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "registry-credentials",
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{
			"username": []byte(rUsername),
			"password": []byte(rPassword),
		},
	}

	client := test.NewFakeKubeClient(cluster, registryAuthSecret)
	helmFactory := helm.NewClientFactory(client, tt.builder)

	tt.builder.EXPECT().BuildHelmExecutable(gomock.Any()).Return(tt.helm)
	tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(cluster), rUsername, rPassword).Return(nil)

	helmClient, err := helmFactory.GetClientForCluster(tt.ctx, cluster)

	tt.Expect(err).To(BeNil())
	tt.Expect(helmClient).ToNot(BeNil())
}

func TestHelmFactoryGetClientForClusterErrorLoginRegistry(t *testing.T) {
	tt := newHelmFactoryTest(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "test-cluster",
		}
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Authenticate: true,
			Endpoint:     "1.2.3.4",
			Port:         "5000",
		}
	})

	rUsername := "username"
	rPassword := "password"
	registryAuthSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "registry-credentials",
			Namespace: cluster.Namespace,
		},
		Data: map[string][]byte{
			"username": []byte(rUsername),
			"password": []byte(rPassword),
		},
	}

	client := test.NewFakeKubeClient(cluster, registryAuthSecret)
	helmFactory := helm.NewClientFactory(client, tt.builder)

	tt.builder.EXPECT().BuildHelmExecutable(gomock.Any()).Return(tt.helm)
	tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(cluster), rUsername, rPassword).Return(errors.New("login registry error"))

	_, err := helmFactory.GetClientForCluster(tt.ctx, cluster)

	tt.Expect(err).To(MatchError(ContainSubstring("login registry error")))
}

func TestHelmFactoryGetClientForClusterRegistryMirrorErrorNoRegistryCredentials(t *testing.T) {
	tt := newHelmFactoryTest(t)
	managmentCluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "management-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "management-cluster",
		}
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Authenticate: true,
			Endpoint:     "1.2.3.4",
			Port:         "5000",
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

	helmFactory := helm.NewClientFactory(client, tt.builder)
	_, err := helmFactory.GetClientForCluster(tt.ctx, cluster)

	tt.Expect(err).To(MatchError(ContainSubstring("please set REGISTRY_USERNAME")))
}

func TestHelmFactoryGetClientForClusterSuccessRegistryMirrorEnvCredendialss(t *testing.T) {
	tt := newHelmFactoryTest(t)
	managmentCluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "management-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "management-cluster",
		}
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Authenticate: true,
			Endpoint:     "1.2.3.4",
			Port:         "5000",
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

	helmFactory := helm.NewClientFactory(client, tt.builder)
	rUsername := "username"
	rPassword := "password"

	t.Setenv("REGISTRY_USERNAME", rUsername)
	t.Setenv("REGISTRY_PASSWORD", rPassword)

	tt.builder.EXPECT().BuildHelmExecutable(gomock.Any()).Return(tt.helm)
	tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(managmentCluster), rUsername, rPassword).Return(nil)

	helmClient, err := helmFactory.GetClientForCluster(tt.ctx, cluster)
	tt.Expect(err).To(BeNil())
	tt.Expect(helmClient).ToNot(BeNil())
}

package helm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/helm"
	helmmocks "github.com/aws/eks-anywhere/pkg/helm/mocks"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

type helmEnvClientFactoryTest struct {
	*WithT
	ctx     context.Context
	builder *helmmocks.MockClientBuilder
	helm    *helmmocks.MockClient
}

func newHelmEnvClientFactoryTest(t *testing.T) *helmEnvClientFactoryTest {
	ctrl := gomock.NewController(t)
	builder := helmmocks.NewMockClientBuilder(ctrl)
	helm := helmmocks.NewMockClient(ctrl)
	return &helmEnvClientFactoryTest{
		WithT:   NewWithT(t),
		ctx:     context.Background(),
		builder: builder,
		helm:    helm,
	}
}

func TestHelmEnvClientFactoryGetSuccessNoRegistryMirror(t *testing.T) {
	tt := newHelmEnvClientFactoryTest(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: c.Name,
		}
	})

	helmFactory := helm.NewEnvClientFactory(tt.builder)

	tt.builder.EXPECT().BuildHelm(gomock.Any()).Return(tt.helm)

	err := helmFactory.Init(tt.ctx, registrymirror.FromCluster(cluster))
	tt.Expect(err).To(BeNil())

	helm, _ := helmFactory.Get(tt.ctx, cluster)
	tt.Expect(helm).NotTo(BeNil())
}

func TestHelmEnvClientFactoryGetSuccessNoAuthRegistryMirror(t *testing.T) {
	tt := newHelmEnvClientFactoryTest(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: c.Name,
		}
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Endpoint: "1.2.3.4",
			Port:     "5000",
		}
	})

	helmFactory := helm.NewEnvClientFactory(tt.builder)

	tt.builder.EXPECT().BuildHelm(gomock.Any()).Return(tt.helm)

	err := helmFactory.Init(tt.ctx, registrymirror.FromCluster(cluster))
	tt.Expect(err).To(BeNil())

	helm, _ := helmFactory.Get(tt.ctx, cluster)
	tt.Expect(helm).NotTo(BeNil())
}

func TestHelmEnvClientFactoryGetErrorMissingRegistryCredentials(t *testing.T) {
	tt := newHelmEnvClientFactoryTest(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
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

	helmFactory := helm.NewEnvClientFactory(tt.builder)

	tt.builder.EXPECT().BuildHelm(gomock.Any()).Return(tt.helm)

	err := helmFactory.Init(tt.ctx, registrymirror.FromCluster(cluster))
	tt.Expect(err).To(MatchError(ContainSubstring("please set REGISTRY_USERNAME")))

	helm, _ := helmFactory.Get(tt.ctx, cluster)
	tt.Expect(helm).To(BeNil())
}

func TestHelmEnvClientFactoryGetErrorRegistryLogin(t *testing.T) {
	tt := newHelmEnvClientFactoryTest(t)
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

	t.Setenv("REGISTRY_USERNAME", rUsername)
	t.Setenv("REGISTRY_PASSWORD", rPassword)

	helmFactory := helm.NewEnvClientFactory(tt.builder)

	tt.builder.EXPECT().BuildHelm(gomock.Any()).Return(tt.helm)
	tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(cluster), rUsername, rPassword).Return(errors.New("login registry error"))

	err := helmFactory.Init(tt.ctx, registrymirror.FromCluster(cluster))
	tt.Expect(err).To(MatchError(ContainSubstring("login registry error")))

	helm, _ := helmFactory.Get(tt.ctx, cluster)
	tt.Expect(helm).To(BeNil())
}

func TestHelmEnvClientFactoryGetSuccessAuthenticatedRegistryMirror(t *testing.T) {
	tt := newHelmEnvClientFactoryTest(t)
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

	t.Setenv("REGISTRY_USERNAME", rUsername)
	t.Setenv("REGISTRY_PASSWORD", rPassword)

	helmFactory := helm.NewEnvClientFactory(tt.builder)

	tt.builder.EXPECT().BuildHelm(gomock.Any()).Return(tt.helm)
	tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(cluster), rUsername, rPassword).Return(nil)

	err := helmFactory.Init(tt.ctx, registrymirror.FromCluster(cluster))
	tt.Expect(err).To(BeNil())

	helm, _ := helmFactory.Get(tt.ctx, cluster)
	tt.Expect(helm).ToNot(BeNil())
}

func TestHelmEnvClientFactoryGetAlreadyInitialized(t *testing.T) {
	tt := newHelmEnvClientFactoryTest(t)
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

	t.Setenv("REGISTRY_USERNAME", rUsername)
	t.Setenv("REGISTRY_PASSWORD", rPassword)

	helmFactory := helm.NewEnvClientFactory(tt.builder)

	tt.builder.EXPECT().BuildHelm(gomock.Any()).Return(tt.helm)
	tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(cluster), rUsername, rPassword).Return(nil)

	err := helmFactory.Init(tt.ctx, registrymirror.FromCluster(cluster))
	tt.Expect(err).To(BeNil())

	newCluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "new-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "new-cluster",
		}
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Authenticate: true,
			Endpoint:     "1.2.3.4",
			Port:         "5000",
		}
	})

	err = helmFactory.Init(tt.ctx, registrymirror.FromCluster(newCluster))
	tt.Expect(err).To(BeNil())
}

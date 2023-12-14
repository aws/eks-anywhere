package helm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

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
	helm    *helmmocks.MockClient
}

func newHelmFactoryTest(t *testing.T) *helmFactoryTest {
	ctrl := gomock.NewController(t)
	builder := helmmocks.NewMockExecutableBuilder(ctrl)
	helm := helmmocks.NewMockClient(ctrl)
	return &helmFactoryTest{
		WithT:   NewWithT(t),
		ctx:     context.Background(),
		builder: builder,
		helm:    helm,
	}
}

func TestHelmFactoryGetSuccess(t *testing.T) {
	tt := newHelmFactoryTest(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: c.Name,
		}
	})

	client := fake.NewClientBuilder().WithRuntimeObjects(cluster).Build()
	helmFactory := helm.NewClientForClusterFactory(client, tt.builder)

	tt.builder.EXPECT().BuildHelmExecutable(gomock.Any()).Return(tt.helm)

	helm, err := helmFactory.Get(tt.ctx, cluster)

	tt.Expect(err).To(BeNil())
	tt.Expect(helm).NotTo(BeNil())
}

func TestHelmFactoryGetErrorManagmentClusterNotFound(t *testing.T) {
	tt := newHelmFactoryTest(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: "management-cluster",
		}
	})

	client := fake.NewClientBuilder().WithRuntimeObjects(cluster).Build()
	helmFactory := helm.NewClientForClusterFactory(client, tt.builder)

	helm, err := helmFactory.Get(tt.ctx, cluster)

	tt.Expect(helm).To(BeNil())
	tt.Expect(err).To(MatchError(ContainSubstring("unable to retrieve management cluster")))
}

func TestHelmFactoryGetAuthenticatedRegistryMirrorErrorGettingSecret(t *testing.T) {
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

	client := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()
	helmFactory := helm.NewClientForClusterFactory(client, tt.builder)

	_, err := helmFactory.Get(tt.ctx, cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("fetching registry auth secret: no kind is registered for the type v1.Secret")))
}

func TestHelmFactoryGetSuccessAuthenticatedRegistryMirrorClientReused(t *testing.T) {
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

	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, registryAuthSecret).Build()
	helmFactory := helm.NewClientForClusterFactory(client, tt.builder)

	tt.builder.EXPECT().BuildHelmExecutable(gomock.Any()).Return(tt.helm)
	tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(cluster), rUsername, rPassword).Return(nil)
	helmClient, err := helmFactory.Get(tt.ctx, cluster)

	tt.Expect(err).To(BeNil())
	tt.Expect(helmClient).ToNot(BeNil())

	tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(cluster), rUsername, rPassword).Return(nil)
	newHelmClient, err := helmFactory.Get(tt.ctx, cluster)

	tt.Expect(err).To(BeNil())
	tt.Expect(newHelmClient).ToNot(BeNil())

}

func TestHelmFactoryGetSuccessAuthenticatedRegistryMirrorClientRebuilt(t *testing.T) {
	tt := newHelmFactoryTest(t)

	testCases := map[string]struct {
		oldRegistryMirror *anywherev1.RegistryMirrorConfiguration
		newRegistryMirror *anywherev1.RegistryMirrorConfiguration
	}{
		"SuccessEndpointChanged": {
			oldRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate: true,
				Endpoint:     "1.2.3.4",
				Port:         "5000",
			},
			newRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate: true,
				Endpoint:     "2.3.4.5",
				Port:         "5000",
			},
		},
		"SuccessAuthChanged": {
			oldRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate: true,
				Endpoint:     "1.2.3.4",
				Port:         "5000",
			},
			newRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate: false,
				Endpoint:     "1.2.3.4",
				Port:         "5000",
			},
		},
		"SuccessPortChanged": {
			oldRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate: true,
				Endpoint:     "1.2.3.4",
				Port:         "5000",
			},
			newRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate: true,
				Endpoint:     "1.2.3.4",
				Port:         "5050",
			},
		},
		"SuccessCACertChanged": {
			oldRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate:  true,
				Endpoint:      "1.2.3.4",
				Port:          "5000",
				CACertContent: "",
			},
			newRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate:  true,
				Endpoint:      "1.2.3.4",
				Port:          "5050",
				CACertContent: "test cert",
			},
		},
		"SuccessOCINamespacesChanged": {
			oldRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate:  true,
				Endpoint:      "1.2.3.4",
				Port:          "5000",
				CACertContent: "test cert",
				OCINamespaces: []anywherev1.OCINamespace{
					{
						Registry:  "1.2.3.4",
						Namespace: "publice.ecr.aws",
					},
				},
			},
			newRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate:  true,
				Endpoint:      "1.2.3.4",
				Port:          "5050",
				CACertContent: "test cert",
				OCINamespaces: []anywherev1.OCINamespace{
					{
						Registry:  "1.2.3.4",
						Namespace: "test.namespace.net",
					},
				},
			},
		},
		"SuccessInsecureSkipVerifyChanged": {
			oldRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate:       true,
				Endpoint:           "1.2.3.4",
				Port:               "5000",
				InsecureSkipVerify: false,
			},
			newRegistryMirror: &anywherev1.RegistryMirrorConfiguration{
				Authenticate:       true,
				Endpoint:           "1.2.3.4",
				Port:               "5050",
				InsecureSkipVerify: true,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			cluster := test.Cluster(func(c *v1alpha1.Cluster) {
				c.Name = "test-cluster"
				c.Namespace = constants.EksaSystemNamespace
				c.Spec.ManagementCluster = anywherev1.ManagementCluster{
					Name: "test-cluster",
				}
				c.Spec.RegistryMirrorConfiguration = testCase.oldRegistryMirror
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

			client := fake.NewClientBuilder().WithRuntimeObjects(cluster, registryAuthSecret).Build()
			helmFactory := helm.NewClientForClusterFactory(client, tt.builder)

			tt.builder.EXPECT().BuildHelmExecutable(gomock.Any()).Return(tt.helm)
			if cluster.RegistryAuth() {
				tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(cluster), rUsername, rPassword).Return(nil)
			}

			helmClient, err := helmFactory.Get(tt.ctx, cluster)

			tt.Expect(err).To(BeNil())
			tt.Expect(helmClient).ToNot(BeNil())

			cluster.Spec.RegistryMirrorConfiguration = testCase.newRegistryMirror

			tt.builder.EXPECT().BuildHelmExecutable(gomock.Any()).Return(tt.helm)
			if cluster.RegistryAuth() {
				tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(cluster), rUsername, rPassword).Return(nil)
			}

			newHelmClient, err := helmFactory.Get(tt.ctx, cluster)

			tt.Expect(err).To(BeNil())
			tt.Expect(newHelmClient).ToNot(BeNil())
		})
	}
}

func TestHelmFactoryGetSuccessAuthenticatedRegistryMirror(t *testing.T) {
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

	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, registryAuthSecret).Build()
	helmFactory := helm.NewClientForClusterFactory(client, tt.builder)

	tt.builder.EXPECT().BuildHelmExecutable(gomock.Any()).Return(tt.helm)
	tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(cluster), rUsername, rPassword).Return(nil)

	helmClient, err := helmFactory.Get(tt.ctx, cluster)

	tt.Expect(err).To(BeNil())
	tt.Expect(helmClient).ToNot(BeNil())
}

func TestHelmFactoryGetErrorLoginRegistry(t *testing.T) {
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

	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, registryAuthSecret).Build()
	helmFactory := helm.NewClientForClusterFactory(client, tt.builder)

	tt.builder.EXPECT().BuildHelmExecutable(gomock.Any()).Return(tt.helm)
	tt.helm.EXPECT().RegistryLogin(tt.ctx, test.RegistryMirrorEndpoint(cluster), rUsername, rPassword).Return(errors.New("login registry error"))

	_, err := helmFactory.Get(tt.ctx, cluster)

	tt.Expect(err).To(MatchError(ContainSubstring("login registry error")))
}

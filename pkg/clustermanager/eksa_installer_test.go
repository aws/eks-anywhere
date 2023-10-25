package clustermanager_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type installerTest struct {
	*WithT
	ctx         context.Context
	log         logr.Logger
	client      *mocks.MockKubernetesClient
	currentSpec *cluster.Spec
	newSpec     *cluster.Spec
	installer   *clustermanager.EKSAInstaller
	cluster     *types.Cluster
}

func newInstallerTest(t *testing.T, opts ...clustermanager.EKSAInstallerOpt) *installerTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)
	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Eksa.Version = "v0.1.0"
		s.Cluster = &anywherev1.Cluster{
			Spec: anywherev1.ClusterSpec{
				ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
					Endpoint: &anywherev1.Endpoint{
						Host: "1.2.3.4",
					},
				},
				KubernetesVersion: "1.19",
			},
		}
	})

	return &installerTest{
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		log:         test.NewNullLogger(),
		client:      client,
		installer:   clustermanager.NewEKSAInstaller(client, files.NewReader(), opts...),
		currentSpec: currentSpec,
		newSpec:     currentSpec.DeepCopy(),
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "k.kubeconfig",
		},
	}
}

func TestEKSAInstallerInstallSuccessWithRealManifest(t *testing.T) {
	tt := newInstallerTest(t)
	tt.newSpec.VersionsBundles["1.19"].Eksa.Components.URI = "../../config/manifest/eksa-components.yaml"
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any()).Times(34) // there are 34 objects in the manifest
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m0s", "Available", "eksa-controller-manager", "eksa-system")

	tt.Expect(tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newSpec)).To(Succeed())
}

func TestEKSAInstallerInstallSuccessWithTestManifest(t *testing.T) {
	tt := newInstallerTest(t)
	tt.newSpec.VersionsBundles["1.19"].Eksa.Components.URI = "testdata/eksa_components.yaml"
	tt.newSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "1.2.3.4"
	tt.newSpec.Cluster.Spec.DatacenterRef.Kind = anywherev1.VSphereDatacenterKind
	tt.newSpec.Cluster.Spec.ProxyConfiguration = &anywherev1.ProxyConfiguration{
		HttpProxy:  "proxy",
		HttpsProxy: "proxy",
		NoProxy:    []string{"no-proxy", "no-proxy-2"},
	}

	wantDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-controller-manager",
			Namespace: "eksa-system",
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Args: []string{
								"--leader-elect",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "HTTPS_PROXY",
									Value: "proxy",
								},
								{
									Name:  "HTTP_PROXY",
									Value: "proxy",
								},
								{
									Name:  "NO_PROXY",
									Value: "no-proxy,no-proxy-2,1.2.3.4",
								},
							},
						},
					},
				},
			},
		},
	}

	wantNamespace := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "Namespace",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"name": "eksa-system",
			},
		},
	}

	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, wantDeployment)
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, wantNamespace)
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m0s", "Available", "eksa-controller-manager", "eksa-system")

	tt.Expect(tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newSpec)).To(Succeed())
}

func TestEKSAInstallerInstallSuccessWithNoTimeout(t *testing.T) {
	tt := newInstallerTest(t, clustermanager.WithEKSAInstallerNoTimeouts())
	tt.newSpec.VersionsBundles["1.19"].Eksa.Components.URI = "../../config/manifest/eksa-components.yaml"
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any()).Times(34) // there are 34 objects in the manifest
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, maxTime.String(), "Available", "eksa-controller-manager", "eksa-system")

	tt.Expect(tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newSpec)).To(Succeed())
}

func TestInstallerUpgradeNoSelfManaged(t *testing.T) {
	tt := newInstallerTest(t)
	tt.newSpec.Cluster.SetManagedBy("management-cluster")

	tt.Expect(tt.installer.Upgrade(tt.ctx, tt.log, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestInstallerUpgradeNoChanges(t *testing.T) {
	tt := newInstallerTest(t)

	tt.Expect(tt.installer.Upgrade(tt.ctx, tt.log, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestInstallerUpgradeSuccess(t *testing.T) {
	tt := newInstallerTest(t)

	tt.newSpec.VersionsBundles["1.19"].Eksa.Version = "v0.2.0"
	tt.newSpec.VersionsBundles["1.19"].Eksa.Components = v1alpha1.Manifest{
		URI: "testdata/eksa_components.yaml",
	}

	wantDiff := &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{
			{
				ComponentName: "EKS-A",
				NewVersion:    "v0.2.0",
				OldVersion:    "v0.1.0",
			},
		},
	}

	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&unstructured.Unstructured{}))
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m0s", "Available", "eksa-controller-manager", "eksa-system")
	tt.Expect(tt.installer.Upgrade(tt.ctx, tt.log, tt.cluster, tt.currentSpec, tt.newSpec)).To(Equal(wantDiff))
}

func TestInstallerUpgradeInstallError(t *testing.T) {
	tt := newInstallerTest(t)

	tt.newSpec.VersionsBundles["1.19"].Eksa.Version = "v0.2.0"

	// components file not set so this should return an error in failing to load manifest
	_, err := tt.installer.Upgrade(tt.ctx, tt.log, tt.cluster, tt.currentSpec, tt.newSpec)
	tt.Expect(err).NotTo(BeNil())
}

func TestSetManagerFlags(t *testing.T) {
	tests := []struct {
		name           string
		deployment     *appsv1.Deployment
		spec           *cluster.Spec
		featureEnvVars []string
		want           *appsv1.Deployment
	}{
		{
			name:       "no flags",
			deployment: deployment(),
			spec:       test.NewClusterSpec(),
			want:       deployment(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features.ClearCache()
			for _, e := range tt.featureEnvVars {
				t.Setenv(e, "true")
			}
			g := NewWithT(t)
			clustermanager.SetManagerFlags(tt.deployment, tt.spec)
			g.Expect(tt.deployment).To(Equal(tt.want))
		})
	}
}

func TestSetManagerEnvVars(t *testing.T) {
	tests := []struct {
		name       string
		deployment *appsv1.Deployment
		spec       *cluster.Spec
		want       *appsv1.Deployment
	}{
		{
			name:       "no env vars",
			deployment: deployment(),
			spec:       test.NewClusterSpec(),
			want:       deployment(),
		},
		{
			name:       "proxy env vars",
			deployment: deployment(),
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = &anywherev1.Cluster{
					Spec: anywherev1.ClusterSpec{
						ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
							Endpoint: &anywherev1.Endpoint{
								Host: "1.2.3.4",
							},
						},
						ProxyConfiguration: &anywherev1.ProxyConfiguration{
							HttpProxy:  "proxy",
							HttpsProxy: "proxy",
							NoProxy:    []string{"no-proxy", "no-proxy-2"},
						},
					},
				}
			}),
			want: deployment(func(d *appsv1.Deployment) {
				d.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
					{
						Name:  "HTTPS_PROXY",
						Value: "proxy",
					},
					{
						Name:  "HTTP_PROXY",
						Value: "proxy",
					},
					{
						Name:  "NO_PROXY",
						Value: "no-proxy,no-proxy-2,1.2.3.4",
					},
				}
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			clustermanager.SetManagerEnvVars(tt.deployment, tt.spec)
			g.Expect(tt.deployment).To(Equal(tt.want))
		})
	}
}

type deploymentOpt func(*appsv1.Deployment)

func deployment(opts ...deploymentOpt) *appsv1.Deployment {
	d := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{},
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

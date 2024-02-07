package clustermanager_test

import (
	"context"
	"errors"
	"os"
	"strings"
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
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type installerTest struct {
	*WithT
	ctx                         context.Context
	log                         logr.Logger
	client                      *mocks.MockKubernetesClient
	currentManagementComponents *cluster.ManagementComponents
	newManagementComponents     *cluster.ManagementComponents
	currentSpec                 *cluster.Spec
	newSpec                     *cluster.Spec
	installer                   *clustermanager.EKSAInstaller
	cluster                     *types.Cluster
}

func newInstallerTest(t *testing.T, opts ...clustermanager.EKSAInstallerOpt) *installerTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)
	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
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
		s.Bundles.Spec.VersionsBundles[0].Eksa.Version = "v0.1.0"
	})

	return &installerTest{
		WithT:                       NewWithT(t),
		ctx:                         context.Background(),
		log:                         test.NewNullLogger(),
		client:                      client,
		installer:                   clustermanager.NewEKSAInstaller(client, files.NewReader(), opts...),
		currentManagementComponents: cluster.ManagementComponentsFromBundles(currentSpec.Bundles),
		newManagementComponents:     cluster.ManagementComponentsFromBundles(currentSpec.Bundles),
		currentSpec:                 currentSpec,
		newSpec:                     currentSpec.DeepCopy(),
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "k.kubeconfig",
		},
	}
}

func TestEKSAInstallerInstallSuccessWithRealManifest(t *testing.T) {
	tt := newInstallerTest(t)
	tt.newManagementComponents.Eksa.Components.URI = "../../config/manifest/eksa-components.yaml"
	file, err := os.ReadFile("../../config/manifest/eksa-components.yaml")
	if err != nil {
		t.Fatalf("could not read eksa-components")
	}
	manifest := string(file)
	expectedObjectCount := strings.Count(manifest, "\n---\n") + 1
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any()).Times(expectedObjectCount)
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m0s", "Available", "eksa-controller-manager", "eksa-system")
	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Any()).Times(2)
	tt.client.EXPECT().GetConfigMap(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any(), gomock.Any()).Return(nil, errors.New("NotFound"))

	tt.Expect(tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newManagementComponents, tt.newSpec)).To(Succeed())
}

func TestEKSAInstallerInstallFailComponentsDeployment(t *testing.T) {
	tt := newInstallerTest(t)
	tt.newManagementComponents.Eksa.Components.URI = "../../config/manifest/eksa-components.yaml"
	file, err := os.ReadFile("../../config/manifest/eksa-components.yaml")
	if err != nil {
		t.Fatalf("could not read eksa-components")
	}

	manifest := string(file)
	expectedObjectCount := strings.Count(manifest, "\n---\n")
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any()).Times(expectedObjectCount)
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m0s", "Available", "eksa-controller-manager", "eksa-system").Return(errors.New("test"))

	err = tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newManagementComponents, tt.newSpec)
	tt.Expect(err.Error()).To(ContainSubstring("waiting for eksa-controller-manager"))
}

func TestEKSAInstallerInstallFailComponents(t *testing.T) {
	tt := newInstallerTest(t)
	tt.newManagementComponents.Eksa.Components.URI = "../../config/manifest/eksa-components.yaml"

	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{})).Return(errors.New("test"))

	err := tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newManagementComponents, tt.newSpec)
	tt.Expect(err.Error()).To(ContainSubstring("applying eksa components"))
}

func TestEKSAInstallerInstallFailBundles(t *testing.T) {
	tt := newInstallerTest(t)
	tt.newManagementComponents.Eksa.Components.URI = "../../config/manifest/eksa-components.yaml"
	file, err := os.ReadFile("../../config/manifest/eksa-components.yaml")
	if err != nil {
		t.Fatalf("could not read eksa-components")
	}

	tt.newSpec.Bundles = &v1alpha1.Bundles{}
	tt.newSpec.EKSARelease = &v1alpha1.EKSARelease{}
	manifest := string(file)
	expectedObjectCount := strings.Count(manifest, "\n---\n")
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any()).Times(expectedObjectCount)
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m0s", "Available", "eksa-controller-manager", "eksa-system")
	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Any()).Return(errors.New("test"))

	err = tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newManagementComponents, tt.newSpec)
	tt.Expect(err.Error()).To(ContainSubstring("applying bundle spec"))
}

func TestEKSAInstallerInstallFailEKSARelease(t *testing.T) {
	tt := newInstallerTest(t)
	tt.newManagementComponents.Eksa.Components.URI = "../../config/manifest/eksa-components.yaml"
	file, err := os.ReadFile("../../config/manifest/eksa-components.yaml")
	if err != nil {
		t.Fatalf("could not read eksa-components")
	}

	tt.newSpec.Bundles = &v1alpha1.Bundles{}
	tt.newSpec.EKSARelease = &v1alpha1.EKSARelease{}
	manifest := string(file)
	expectedObjectCount := strings.Count(manifest, "\n---\n") + 1
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any()).Times(expectedObjectCount)
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m0s", "Available", "eksa-controller-manager", "eksa-system")
	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Any())
	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Any()).Return(errors.New("test"))
	tt.client.EXPECT().GetConfigMap(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any(), gomock.Any()).Return(nil, errors.New("NotFound"))

	err = tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newManagementComponents, tt.newSpec)
	tt.Expect(err.Error()).To(ContainSubstring("applying EKSA release spec"))
}

func TestEKSAInstallerInstallSuccessWithTestManifest(t *testing.T) {
	tt := newInstallerTest(t)
	tt.newManagementComponents.Eksa.Components.URI = "testdata/eksa_components.yaml"
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

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.UpgraderConfigMapName,
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string]string{
			"v1.28": "test-image",
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
	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Any()).Times(2)
	tt.client.EXPECT().GetConfigMap(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any(), gomock.Any()).Return(configMap, nil)
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, configMap)

	tt.Expect(tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newManagementComponents, tt.newSpec)).To(Succeed())
}

func TestEKSAInstallerInstallSuccessWithNoTimeout(t *testing.T) {
	tt := newInstallerTest(t, clustermanager.WithEKSAInstallerNoTimeouts())
	newManagementComponents := cluster.ManagementComponentsFromBundles(tt.newSpec.Bundles)
	newManagementComponents.Eksa.Components.URI = "../../config/manifest/eksa-components.yaml"
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	file, err := os.ReadFile("../../config/manifest/eksa-components.yaml")
	if err != nil {
		t.Fatalf("could not read eksa-components")
	}
	manifest := string(file)
	expectedObjectCount := strings.Count(manifest, "\n---\n") + 1
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any()).Times(expectedObjectCount)
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, maxTime.String(), "Available", "eksa-controller-manager", "eksa-system")
	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Any()).Times(2)
	tt.client.EXPECT().GetConfigMap(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any(), gomock.Any()).Return(nil, errors.New("NotFound"))

	tt.Expect(tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, newManagementComponents, tt.newSpec)).To(Succeed())
}

func TestInstallerUpgradeNoSelfManaged(t *testing.T) {
	tt := newInstallerTest(t)
	tt.newSpec.Cluster.SetManagedBy("management-cluster")

	tt.Expect(tt.installer.Upgrade(tt.ctx, tt.log, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.newSpec)).To(BeNil())
}

func TestInstallerUpgradeNoChanges(t *testing.T) {
	tt := newInstallerTest(t)

	tt.Expect(tt.installer.Upgrade(tt.ctx, tt.log, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.newSpec)).To(BeNil())
}

func TestInstallerUpgradeSuccess(t *testing.T) {
	tt := newInstallerTest(t)

	tt.newManagementComponents.Eksa.Version = "v0.2.0"
	tt.newManagementComponents.Eksa.Components = v1alpha1.Manifest{
		URI: "testdata/eksa_components.yaml",
	}

	wantDiff := &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{
			{
				ComponentName: "EKS-A Management",
				NewVersion:    "v0.2.0",
				OldVersion:    "v0.1.0",
			},
		},
	}

	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&unstructured.Unstructured{}))
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m0s", "Available", "eksa-controller-manager", "eksa-system")
	tt.Expect(tt.installer.Upgrade(tt.ctx, tt.log, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.newSpec)).To(Equal(wantDiff))
}

func TestInstallerUpgradeInstallError(t *testing.T) {
	tt := newInstallerTest(t)

	tt.newManagementComponents.Eksa.Version = "v0.2.0"

	// components file not set so this should return an error in failing to load manifest
	_, err := tt.installer.Upgrade(tt.ctx, tt.log, tt.cluster, tt.currentManagementComponents, tt.newManagementComponents, tt.newSpec)
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

func TestSetManagerEnvVarsVSphereInPlaceUpgrade(t *testing.T) {
	g := NewWithT(t)
	features.ClearCache()
	t.Setenv(features.VSphereInPlaceEnvVar, "true")

	deploy := deployment()
	spec := test.NewClusterSpec()
	want := deployment(func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
			{
				Name:  "VSPHERE_IN_PLACE_UPGRADE",
				Value: "true",
			},
		}
	})

	clustermanager.SetManagerEnvVars(deploy, spec)
	g.Expect(deploy).To(Equal(want))
}

func TestEKSAInstallerNewUpgraderConfigMap(t *testing.T) {
	tt := newInstallerTest(t)

	tt.newManagementComponents.Eksa.Components.URI = "../../config/manifest/eksa-components.yaml"
	file, err := os.ReadFile("../../config/manifest/eksa-components.yaml")
	if err != nil {
		t.Fatalf("could not read eksa-components")
	}

	tt.newSpec.Bundles = &v1alpha1.Bundles{}
	tt.newSpec.EKSARelease = &v1alpha1.EKSARelease{}
	manifest := string(file)
	expectedObjectCount := strings.Count(manifest, "\n---\n") + 1
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any()).Times(expectedObjectCount)
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m0s", "Available", "eksa-controller-manager", "eksa-system")
	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Any())
	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Any())
	tt.client.EXPECT().GetConfigMap(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any(), gomock.Any()).Return(nil, errors.New("NotFound"))

	tt.Expect(tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newManagementComponents, tt.newSpec)).To(Succeed())
}

func TestEKSAInstallerNewUpgraderConfigMapFailure(t *testing.T) {
	tt := newInstallerTest(t)

	tt.newManagementComponents.Eksa.Components.URI = "../../config/manifest/eksa-components.yaml"
	file, err := os.ReadFile("../../config/manifest/eksa-components.yaml")
	if err != nil {
		t.Fatalf("could not read eksa-components")
	}

	tt.newSpec.Bundles = &v1alpha1.Bundles{}
	tt.newSpec.EKSARelease = &v1alpha1.EKSARelease{}
	manifest := string(file)
	expectedObjectCount := strings.Count(manifest, "\n---\n")
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any()).Times(expectedObjectCount)
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m0s", "Available", "eksa-controller-manager", "eksa-system")
	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Any())
	tt.client.EXPECT().GetConfigMap(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any(), gomock.Any()).Return(nil, errors.New(""))
	err = tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newManagementComponents, tt.newSpec)

	tt.Expect(err.Error()).To(ContainSubstring("getting upgrader images from bundle"))
}

func TestEKSAInstallerFailureApplyUpgraderConfigMap(t *testing.T) {
	tt := newInstallerTest(t)

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.UpgraderConfigMapName,
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string]string{
			"v1.28": "test-image",
		},
	}

	tt.newManagementComponents.Eksa.Components.URI = "../../config/manifest/eksa-components.yaml"
	file, err := os.ReadFile("../../config/manifest/eksa-components.yaml")
	if err != nil {
		t.Fatalf("could not read eksa-components")
	}

	manifest := string(file)
	expectedObjectCount := strings.Count(manifest, "\n---\n")
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.AssignableToTypeOf(&appsv1.Deployment{}))
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any()).Times(expectedObjectCount)
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m0s", "Available", "eksa-controller-manager", "eksa-system")
	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, gomock.Any())
	tt.client.EXPECT().GetConfigMap(tt.ctx, tt.cluster.KubeconfigFile, gomock.Any(), gomock.Any()).Return(configMap, nil)
	tt.client.EXPECT().Apply(tt.ctx, tt.cluster.KubeconfigFile, configMap).Return(errors.New(""))

	err = tt.installer.Install(tt.ctx, test.NewNullLogger(), tt.cluster, tt.newManagementComponents, tt.newSpec)
	tt.Expect(err.Error()).To(ContainSubstring("applying upgrader images config map"))
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

package stack_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	filewritermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack/mocks"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	overridesFileName = "tinkerbell-chart-overrides.yaml"
	boots             = "boots"
	testIP            = "1.2.3.4"

	helmChartPath    = "public.ecr.aws/eks-anywhere/tinkerbell/tinkerbell-chart"
	helmChartName    = "tinkerbell-chart"
	helmChartVersion = "0.1.0"
)

var helmChartURI = fmt.Sprintf("%s:%s", helmChartPath, helmChartVersion)

func getTinkBundle() releasev1alpha1.TinkerbellBundle {
	return releasev1alpha1.TinkerbellBundle{
		TinkerbellStack: releasev1alpha1.TinkerbellStackBundle{
			Tink: releasev1alpha1.TinkBundle{
				TinkController: releasev1alpha1.Image{URI: "public.ecr.aws/eks-anywhere/tink-controller:latest"},
				TinkServer:     releasev1alpha1.Image{URI: "public.ecr.aws/eks-anywhere/tink-server:latest"},
				TinkWorker:     releasev1alpha1.Image{URI: "public.ecr.aws/eks-anywhere/tink-worker:latest"},
			},
			Boots: releasev1alpha1.Image{URI: "public.ecr.aws/eks-anywhere/boots:latest"},
			Hegel: releasev1alpha1.Image{URI: "public.ecr.aws/eks-anywhere/hegel:latest"},
			Hook: releasev1alpha1.HookBundle{
				Initramfs: releasev1alpha1.HookArch{
					Amd: releasev1alpha1.Archive{
						URI: "https://anywhere-assests.eks.amazonaws.com/tinkerbell/hook/initramfs-x86-64",
					},
				},
			},
			Rufio: releasev1alpha1.Image{
				URI: "public.ecr.aws/eks-anywhere/rufio:latest",
			},
			TinkebellChart: releasev1alpha1.Image{
				Name: helmChartName,
				URI:  helmChartURI,
			},
		},
		KubeVip: releasev1alpha1.Image{
			URI: "public.ecr.aws/eks-anywhere/kube-vip:latest",
		},
		Envoy: releasev1alpha1.Image{
			URI: "public.ecr.aws/eks-anywhere/envoy:latest",
		},
	}
}

func assertYamlFilesEqual(t *testing.T, wantYamlPath, gotYamlPath string) {
	processUpdate(t, wantYamlPath, gotYamlPath)

	if diff := cmp.Diff(unmarshalYamlToObject(t, wantYamlPath), unmarshalYamlToObject(t, gotYamlPath)); diff != "" {
		t.Errorf("Expected file mismatch (-want +got):\n%s", diff)
	}
}

func unmarshalYamlToObject(t *testing.T, filepath string) map[string]interface{} {
	unmarshaledObject := make(map[string]interface{})
	bytes := test.ReadFileAsBytes(t, filepath)
	if err := yaml.Unmarshal(bytes, &unmarshaledObject); err != nil {
		t.Fatalf("failed to unmarshal %s: %v", filepath, err)
	}

	return unmarshaledObject
}

func processUpdate(t *testing.T, goldenFilePath, generatedFilePath string) {
	if *test.UpdateGoldenFiles {
		if err := os.WriteFile(goldenFilePath, test.ReadFileAsBytes(t, generatedFilePath), 0o644); err != nil {
			t.Fatalf("failed to update golden file %s: %v", goldenFilePath, err)
		}
		log.Printf("Golden file updated: %s", goldenFilePath)
	}
}

// Note: This test contains generated files
// To automatically update the generated files, run the following
// go test -timeout 30s -run ^TestTinkerbellStackInstallWithDifferentOptions$ github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack -update -count=1 -v.
func TestTinkerbellStackInstallWithDifferentOptions(t *testing.T) {
	stackTests := []struct {
		name              string
		hookImageOverride string
		expectedFile      string
		installOnDocker   bool
		registryMirror    *registrymirror.RegistryMirror
		proxyConfig       *v1alpha1.ProxyConfiguration
		opts              []stack.InstallOption
	}{
		{
			name:         "with_namespace_create_true",
			expectedFile: "testdata/expected_with_namespace_create_true.yaml",
			opts:         []stack.InstallOption{stack.WithNamespaceCreate(true)},
		},
		{
			name:         "with_namespace_create_false",
			expectedFile: "testdata/expected_with_namespace_create_false.yaml",
			opts:         []stack.InstallOption{stack.WithNamespaceCreate(false)},
		},
		{
			name:            "with_boots_on_docker",
			expectedFile:    "testdata/expected_with_boots_on_docker.yaml",
			installOnDocker: true,
			opts:            []stack.InstallOption{stack.WithBootsOnDocker()},
		},
		{
			name:         "with_boots_on_kubernetes",
			expectedFile: "testdata/expected_with_boots_on_kubernetes.yaml",
			opts:         []stack.InstallOption{stack.WithBootsOnKubernetes()},
		},
		{
			name:         "with_host_port_enabled_true",
			expectedFile: "testdata/expected_with_host_port_enabled_true.yaml",
			opts:         []stack.InstallOption{stack.WithHostPortEnabled(true)},
		},
		{
			name:         "with_host_port_enabled_false",
			expectedFile: "testdata/expected_with_host_port_enabled_false.yaml",
			opts:         []stack.InstallOption{stack.WithHostPortEnabled(false)},
		},
		{
			name:         "with_envoy_enabled_true",
			expectedFile: "testdata/expected_with_envoy_enabled_true.yaml",
			opts:         []stack.InstallOption{stack.WithEnvoyEnabled(true)},
		},
		{
			name:         "with_envoy_enabled_false",
			expectedFile: "testdata/expected_with_envoy_enabled_false.yaml",
			opts:         []stack.InstallOption{stack.WithEnvoyEnabled(false)},
		},
		{
			name:         "with_load_balancer_enabled_true",
			expectedFile: "testdata/expected_with_load_balancer_enabled_true.yaml",
			opts:         []stack.InstallOption{stack.WithLoadBalancerEnabled(true)},
		},
		{
			name:         "with_load_balancer_enabled_false",
			expectedFile: "testdata/expected_with_load_balancer_enabled_false.yaml",
			opts:         []stack.InstallOption{stack.WithLoadBalancerEnabled(false)},
		},
		{
			name:         "with_kubernetes_options",
			expectedFile: "testdata/expected_with_kubernetes_options.yaml",
			opts: []stack.InstallOption{
				stack.WithNamespaceCreate(true),
				stack.WithBootsOnKubernetes(),
				stack.WithEnvoyEnabled(true),
				stack.WithLoadBalancerEnabled(true),
			},
		},
		{
			name:            "with_docker_options",
			expectedFile:    "testdata/expected_with_docker_options.yaml",
			installOnDocker: true,
			opts: []stack.InstallOption{
				stack.WithNamespaceCreate(false),
				stack.WithBootsOnDocker(),
				stack.WithHostPortEnabled(true),
				stack.WithEnvoyEnabled(false),
				stack.WithLoadBalancerEnabled(false),
			},
		},
		{
			name:              "with_hook_override",
			hookImageOverride: "https://my-local-web-server/hook",
			expectedFile:      "testdata/expected_with_hook_override.yaml",
			opts:              []stack.InstallOption{},
		},
		{
			name:         "with_registry_mirror",
			expectedFile: "testdata/expected_with_registry_mirror.yaml",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "1.2.3.4:443",
				NamespacedRegistryMap: map[string]string{
					"public.ecr.aws": "1.2.3.4:443/custom",
				},
				Auth: true,
			},
			opts: []stack.InstallOption{},
		},
		{
			name:         "with_proxy_config",
			expectedFile: "testdata/expected_with_proxy_config.yaml",
			proxyConfig: &v1alpha1.ProxyConfiguration{
				HttpProxy:  "1.2.3.4:3128",
				HttpsProxy: "1.2.3.4:3128",
			},
			opts: []stack.InstallOption{},
		},
	}

	for _, stackTest := range stackTests {
		t.Run(stackTest.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			docker := mocks.NewMockDocker(mockCtrl)
			helm := mocks.NewMockHelm(mockCtrl)
			folder, writer := test.NewWriter(t)
			cluster := &types.Cluster{Name: "test"}
			ctx := context.Background()
			s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, "192.168.0.0/16", stackTest.registryMirror, stackTest.proxyConfig)

			generatedOverridesPath := filepath.Join(folder, "generated", overridesFileName)
			if stackTest.registryMirror != nil && stackTest.registryMirror.Auth {
				t.Setenv("REGISTRY_USERNAME", "username")
				t.Setenv("REGISTRY_PASSWORD", "password")
				helm.EXPECT().RegistryLogin(ctx, "1.2.3.4:443", "username", "password")
				helm.EXPECT().InstallChartWithValuesFile(ctx, helmChartName, "oci://1.2.3.4:443/custom/eks-anywhere/tinkerbell/tinkerbell-chart", helmChartVersion, cluster.KubeconfigFile, generatedOverridesPath)

			} else {
				helm.EXPECT().InstallChartWithValuesFile(ctx, helmChartName, fmt.Sprintf("oci://%s", helmChartPath), helmChartVersion, cluster.KubeconfigFile, generatedOverridesPath)
			}

			if stackTest.installOnDocker {
				docker.EXPECT().Run(ctx, "public.ecr.aws/eks-anywhere/boots:latest",
					boots,
					[]string{"-kubeconfig", "/kubeconfig", "-dhcp-addr", "0.0.0.0:67", "-osie-path-override", "https://anywhere-assests.eks.amazonaws.com/tinkerbell/hook"},
					"-v", gomock.Any(),
					"--network", "host",
					"-e", gomock.Any(),
					"-e", gomock.Any(),
					"-e", gomock.Any(),
					"-e", gomock.Any(),
					"-e", gomock.Any(),
					"-e", gomock.Any(),
					"-e", gomock.Any(),
				)
			}

			if err := s.Install(
				ctx,
				getTinkBundle(),
				testIP,
				cluster.KubeconfigFile,
				stackTest.hookImageOverride,
				stackTest.opts...,
			); err != nil {
				t.Fatalf("failed to install Tinkerbell stack: %v", err)
			}

			assertYamlFilesEqual(t, stackTest.expectedFile, generatedOverridesPath)
		})
	}
}

func TestTinkerbellStackUninstallLocalSucess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()

	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, "192.168.0.0/16", nil, nil)

	docker.EXPECT().ForceRemove(ctx, boots)

	err := s.UninstallLocal(ctx)
	if err != nil {
		t.Fatalf("failed to install Tinkerbell stack: %v", err)
	}
}

func TestTinkerbellStackUninstallLocalFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()

	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, "192.168.0.0/16", nil, nil)

	dockerError := "docker error"
	expectedError := fmt.Sprintf("removing local boots container: %s", dockerError)
	docker.EXPECT().ForceRemove(ctx, boots).Return(errors.New(dockerError))

	err := s.UninstallLocal(ctx)
	assert.EqualError(t, err, expectedError, "Error should be: %v, got: %v", expectedError, err)
}

func TestTinkerbellStackCheckLocalBootsExistenceDoesNotExist(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()

	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, "192.168.0.0/16", nil, nil)

	docker.EXPECT().CheckContainerExistence(ctx, "boots").Return(true, nil)
	docker.EXPECT().ForceRemove(ctx, "boots")

	err := s.CleanupLocalBoots(ctx, true)
	assert.NoError(t, err)
}

func TestTinkerbellStackCheckLocalBootsExistenceDoesExist(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()

	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, "192.168.0.0/16", nil, nil)
	expectedErrorMsg := "boots container already exists, delete the container manually"

	docker.EXPECT().CheckContainerExistence(ctx, "boots").Return(true, nil)

	err := s.CleanupLocalBoots(ctx, false)
	assert.EqualError(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
}

func TestTinkerbellStackCheckLocalBootsExistenceDockerError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()

	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, "192.168.0.0/16", nil, nil)

	docker.EXPECT().CheckContainerExistence(ctx, "boots").Return(false, nil)

	err := s.CleanupLocalBoots(ctx, true)
	assert.NoError(t, err)
}

func TestUpgrade(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)

	folder, writer := test.NewWriter(t)
	valuesFile := filepath.Join(folder, "generated", overridesFileName)
	cluster := &types.Cluster{Name: "test"}
	ctx := context.Background()

	helm.EXPECT().
		UpgradeChartWithValuesFile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any())
	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, "192.168.0.0/16", nil, nil)

	err := s.Upgrade(ctx, getTinkBundle(), testIP, cluster.KubeconfigFile, "")
	assert.NoError(t, err)

	assertYamlFilesEqual(t, "testdata/expected_upgrade.yaml", valuesFile)
}

func TestUpgradeWithRegistryMirrorAuthError(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		docker    = mocks.NewMockDocker(mockCtrl)
		helm      = mocks.NewMockHelm(mockCtrl)
		_, writer = test.NewWriter(t)

		cluster = &types.Cluster{Name: "test"}
		ctx     = context.Background()
	)
	registryMirror := &registrymirror.RegistryMirror{
		BaseRegistry: "1.2.3.4:443",
		NamespacedRegistryMap: map[string]string{
			"public.ecr.aws": "1.2.3.4:443/custom",
		},
		Auth: true,
	}
	t.Setenv("REGISTRY_USERNAME", "username")
	t.Setenv("REGISTRY_PASSWORD", "password")

	expectedErrorMsg := "invalid registry credentials"
	helm.EXPECT().RegistryLogin(ctx, "1.2.3.4:443", "username", "password").Return(fmt.Errorf(expectedErrorMsg))

	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, "192.168.0.0/16", registryMirror, nil)

	err := s.Upgrade(ctx, getTinkBundle(), testIP, cluster.KubeconfigFile, "")
	assert.EqualError(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
}

func TestUpdateStackInstallerNoProxyError(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		docker   = mocks.NewMockDocker(mockCtrl)
		helm     = mocks.NewMockHelm(mockCtrl)
		writer   = filewritermocks.NewMockFileWriter(mockCtrl)
	)
	noProxy := []string{
		"localhost", ".svc",
	}
	proxyConfiguration := &v1alpha1.ProxyConfiguration{
		HttpProxy:  "1.2.3.4",
		HttpsProxy: "1.2.3.4",
		NoProxy:    noProxy,
	}

	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, "192.168.0.0/16", nil, proxyConfiguration)
	s.AddNoProxyIP("2.3.4.5")

	noProxy = append(noProxy, "2.3.4.5")
	if !reflect.DeepEqual(proxyConfiguration.NoProxy, noProxy) {
		t.Fatalf("failed upgrading no proxy list of stack installer")
	}
}

func TestUpgradeWithProxy(t *testing.T) {
	var (
		mockCtrl       = gomock.NewController(t)
		docker         = mocks.NewMockDocker(mockCtrl)
		helm           = mocks.NewMockHelm(mockCtrl)
		folder, writer = test.NewWriter(t)

		valuesFile = filepath.Join(folder, "generated", overridesFileName)
		cluster    = &types.Cluster{Name: "test"}
		ctx        = context.Background()
	)

	helm.EXPECT().UpgradeChartWithValuesFile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

	proxyConfiguration := &v1alpha1.ProxyConfiguration{
		HttpProxy:  "1.2.3.4",
		HttpsProxy: "1.2.3.4",
		NoProxy: []string{
			"localhost", ".svc",
		},
	}

	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace, "192.168.0.0/16", nil, proxyConfiguration)

	err := s.Upgrade(ctx, getTinkBundle(), testIP, cluster.KubeconfigFile, "https://my-local-web-server/hook")
	assert.NoError(t, err)

	assertYamlFilesEqual(t, "testdata/expected_upgrade_with_proxy.yaml", valuesFile)
}

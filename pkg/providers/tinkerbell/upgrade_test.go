package tinkerbell

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	rufiov1 "github.com/tinkerbell/rufio/api/v1alpha1"
	tinkv1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	filewritermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/rufiounreleased"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	stackmocks "github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/yaml"
)

func TestProviderPreCoreComponentsUpgrade_NilClusterSpec(t *testing.T) {
	tconfig := NewPreCoreComponentsUpgradeTestConfig(t)

	provider, err := tconfig.GetProvider()
	if err != nil {
		t.Fatalf("Received unexpected error creating provider: %v", err)
	}

	err = provider.PreCoreComponentsUpgrade(
		context.Background(),
		tconfig.Management,
		nil,
	)

	expect := "cluster spec is nil"
	if err == nil || !strings.Contains(err.Error(), expect) {
		t.Fatalf("Expected error containing '%v'; Received '%v'", expect, err)
	}
}

func TestProviderPreCoreComponentsUpgrade_NilCluster(t *testing.T) {
	tconfig := NewPreCoreComponentsUpgradeTestConfig(t)

	provider, err := tconfig.GetProvider()
	if err != nil {
		t.Fatalf("Received unexpected error creating provider: %v", err)
	}

	err = provider.PreCoreComponentsUpgrade(
		context.Background(),
		nil,
		tconfig.ClusterSpec,
	)

	if err != nil {
		t.Fatalf("Received unexpected error: %v", err)
	}
}

func TestProviderPreCoreComponentsUpgrade_StackUpgradeError(t *testing.T) {
	tconfig := NewPreCoreComponentsUpgradeTestConfig(t)

	expect := "foobar"
	bundle := tconfig.ClusterSpec.RootVersionsBundle()
	tconfig.Installer.EXPECT().
		Upgrade(
			gomock.Any(),
			bundle.Tinkerbell,
			tconfig.DatacenterConfig.Spec.TinkerbellIP,
			tconfig.Management.KubeconfigFile,
			tconfig.DatacenterConfig.Spec.HookImagesURLPath,
		).
		Return(errors.New(expect))

	provider, err := tconfig.GetProvider()
	if err != nil {
		t.Fatalf("Couldn't create the provider: %v", err)
	}

	err = provider.PreCoreComponentsUpgrade(context.Background(), tconfig.Management, tconfig.ClusterSpec)
	if err == nil || !strings.Contains(err.Error(), expect) {
		t.Fatalf("Expected error containing '%v'; Received '%v'", expect, err)
	}
}

func TestProviderPreCoreComponentsUpgrade_HasBaseboardManagementCRDError(t *testing.T) {
	tconfig := NewPreCoreComponentsUpgradeTestConfig(t)

	bundle := tconfig.ClusterSpec.RootVersionsBundle()

	tconfig.Installer.EXPECT().
		Upgrade(
			gomock.Any(),
			bundle.Tinkerbell,
			tconfig.TinkerbellIP,
			tconfig.Management.KubeconfigFile,
			tconfig.DatacenterConfig.Spec.HookImagesURLPath,
		).
		Return(nil)

	expect := "foobar"
	tconfig.KubeClient.EXPECT().
		HasCRD(
			gomock.Any(),
			rufiounreleased.BaseboardManagementResourceName,
			tconfig.Management.KubeconfigFile,
		).
		Return(false, errors.New(expect))

	provider, err := tconfig.GetProvider()
	if err != nil {
		t.Fatalf("Couldn't create the provider: %v", err)
	}

	err = provider.PreCoreComponentsUpgrade(context.Background(), tconfig.Management, tconfig.ClusterSpec)
	if err == nil || !strings.Contains(err.Error(), expect) {
		t.Fatalf("Expected error containing '%v'; Received '%v'", expect, err)
	}
}

func TestProviderPreCoreComponentsUpgrade_NoBaseboardManagementCRD(t *testing.T) {
	tconfig := NewPreCoreComponentsUpgradeTestConfig(t)

	bundle := tconfig.ClusterSpec.RootVersionsBundle()

	tconfig.Installer.EXPECT().
		Upgrade(
			gomock.Any(),
			bundle.Tinkerbell,
			tconfig.TinkerbellIP,
			tconfig.Management.KubeconfigFile,
			tconfig.DatacenterConfig.Spec.HookImagesURLPath,
		).
		Return(nil)

	tconfig.KubeClient.EXPECT().
		HasCRD(gomock.Any(), rufiounreleased.BaseboardManagementResourceName, tconfig.Management.KubeconfigFile).
		Return(false, nil)

	provider, err := tconfig.GetProvider()
	if err != nil {
		t.Fatalf("Couldn't create the provider: %v", err)
	}

	err = provider.PreCoreComponentsUpgrade(context.Background(), tconfig.Management, tconfig.ClusterSpec)
	if err != nil {
		t.Fatalf("Received unexpected error: %v", err)
	}
}

func TestProviderPreCoreComponentsUpgrade_RufioConversions(t *testing.T) {
	stackNamespace := "stack-namespace"

	tests := []struct {
		Name                 string
		Hardware             []tinkv1.Hardware
		BaseboardManagements []rufiounreleased.BaseboardManagement
		ExpectMachines       []rufiov1.Machine
		ExpectHardware       []tinkv1.Hardware
	}{
		{
			Name: "NoBaseboardManagementsOrHardware",
		},
		{
			Name: "SingleBaseboardManagement",
			BaseboardManagements: []rufiounreleased.BaseboardManagement{
				{
					Spec: rufiounreleased.BaseboardManagementSpec{
						Connection: rufiounreleased.Connection{
							Host: "host1",
							Port: 443,
							AuthSecretRef: corev1.SecretReference{
								Name:      "name1",
								Namespace: "namespace1",
							},
							InsecureTLS: true,
						},
					},
				},
			},
			ExpectMachines: []rufiov1.Machine{
				PopulateRufioV1MachineMeta(rufiov1.Machine{
					Spec: rufiov1.MachineSpec{
						Connection: rufiov1.Connection{
							Host: "host1",
							Port: 443,
							AuthSecretRef: corev1.SecretReference{
								Name:      "name1",
								Namespace: "namespace1",
							},
							InsecureTLS: true,
						},
					},
				}),
			},
		},
		{
			Name: "MultiBaseboardManagement",
			BaseboardManagements: []rufiounreleased.BaseboardManagement{
				{
					Spec: rufiounreleased.BaseboardManagementSpec{
						Connection: rufiounreleased.Connection{
							Host: "host1",
							Port: 443,
							AuthSecretRef: corev1.SecretReference{
								Name:      "name1",
								Namespace: "namespace1",
							},
							InsecureTLS: true,
						},
					},
				},
				{
					Spec: rufiounreleased.BaseboardManagementSpec{
						Connection: rufiounreleased.Connection{
							Host: "host2",
							Port: 443,
							AuthSecretRef: corev1.SecretReference{
								Name:      "name2",
								Namespace: "namespace2",
							},
							InsecureTLS: true,
						},
					},
				},
				{
					Spec: rufiounreleased.BaseboardManagementSpec{
						Connection: rufiounreleased.Connection{
							Host: "host3",
							Port: 443,
							AuthSecretRef: corev1.SecretReference{
								Name:      "name3",
								Namespace: "namespace3",
							},
							InsecureTLS: true,
						},
					},
				},
			},
			ExpectMachines: []rufiov1.Machine{
				PopulateRufioV1MachineMeta(rufiov1.Machine{
					Spec: rufiov1.MachineSpec{
						Connection: rufiov1.Connection{
							Host: "host1",
							Port: 443,
							AuthSecretRef: corev1.SecretReference{
								Name:      "name1",
								Namespace: "namespace1",
							},
							InsecureTLS: true,
						},
					},
				}),
				PopulateRufioV1MachineMeta(rufiov1.Machine{
					Spec: rufiov1.MachineSpec{
						Connection: rufiov1.Connection{
							Host: "host2",
							Port: 443,
							AuthSecretRef: corev1.SecretReference{
								Name:      "name2",
								Namespace: "namespace2",
							},
							InsecureTLS: true,
						},
					},
				}),
				PopulateRufioV1MachineMeta(rufiov1.Machine{
					Spec: rufiov1.MachineSpec{
						Connection: rufiov1.Connection{
							Host: "host3",
							Port: 443,
							AuthSecretRef: corev1.SecretReference{
								Name:      "name3",
								Namespace: "namespace3",
							},
							InsecureTLS: true,
						},
					},
				}),
			},
		},
		{
			Name: "SingleHardware",
			Hardware: []tinkv1.Hardware{
				{
					Spec: tinkv1.HardwareSpec{
						BMCRef: &v1.TypedLocalObjectReference{
							Kind: "BaseboardManagement",
							Name: "bm1",
						},
					},
				},
			},
			ExpectHardware: []tinkv1.Hardware{
				{
					Spec: tinkv1.HardwareSpec{
						BMCRef: &v1.TypedLocalObjectReference{
							Kind: "Machine",
							Name: "bm1",
						},
					},
				},
			},
		},
		{
			Name: "MultiHardware",
			Hardware: []tinkv1.Hardware{
				{
					Spec: tinkv1.HardwareSpec{
						BMCRef: &v1.TypedLocalObjectReference{
							Name: "bm1",
						},
					},
				},
				{
					Spec: tinkv1.HardwareSpec{
						BMCRef: &v1.TypedLocalObjectReference{
							Name: "bm2",
						},
					},
				},
				{
					Spec: tinkv1.HardwareSpec{
						BMCRef: &v1.TypedLocalObjectReference{
							Name: "bm3",
						},
					},
				},
			},
			ExpectHardware: []tinkv1.Hardware{
				{
					Spec: tinkv1.HardwareSpec{
						BMCRef: &v1.TypedLocalObjectReference{
							Kind: "Machine",
							Name: "bm1",
						},
					},
				},
				{
					Spec: tinkv1.HardwareSpec{
						BMCRef: &v1.TypedLocalObjectReference{
							Kind: "Machine",
							Name: "bm2",
						},
					},
				},
				{
					Spec: tinkv1.HardwareSpec{
						BMCRef: &v1.TypedLocalObjectReference{
							Kind: "Machine",
							Name: "bm3",
						},
					},
				},
			},
		},
		{
			Name: "HardwareWithoutBMCRef",
			Hardware: []tinkv1.Hardware{
				{
					Spec: tinkv1.HardwareSpec{},
				},
				{
					Spec: tinkv1.HardwareSpec{},
				},
				{
					Spec: tinkv1.HardwareSpec{},
				},
			},
		},
		{
			Name: "MultiBaseboardManagementAndHardware",
			BaseboardManagements: []rufiounreleased.BaseboardManagement{
				{
					Spec: rufiounreleased.BaseboardManagementSpec{
						Connection: rufiounreleased.Connection{
							Host: "host1",
							Port: 443,
							AuthSecretRef: corev1.SecretReference{
								Name:      "name1",
								Namespace: "namespace1",
							},
							InsecureTLS: true,
						},
					},
				},
			},
			ExpectMachines: []rufiov1.Machine{
				PopulateRufioV1MachineMeta(rufiov1.Machine{
					Spec: rufiov1.MachineSpec{
						Connection: rufiov1.Connection{
							Host: "host1",
							Port: 443,
							AuthSecretRef: corev1.SecretReference{
								Name:      "name1",
								Namespace: "namespace1",
							},
							InsecureTLS: true,
						},
					},
				}),
			},
			Hardware: []tinkv1.Hardware{
				{
					Spec: tinkv1.HardwareSpec{
						BMCRef: &v1.TypedLocalObjectReference{
							Kind: "BaseboardManagement",
							Name: "bm1",
						},
					},
				},
				{
					Spec: tinkv1.HardwareSpec{},
				},
			},
			ExpectHardware: []tinkv1.Hardware{
				{
					Spec: tinkv1.HardwareSpec{
						BMCRef: &v1.TypedLocalObjectReference{
							Kind: "Machine",
							Name: "bm1",
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			tconfig := NewPreCoreComponentsUpgradeTestConfig(t)

			bundle := tconfig.ClusterSpec.RootVersionsBundle()

			// Configure the mocks to successfully upgrade the Tinkerbell stack using the installer
			// and identify the need to convert deprecated Rufio custom resources.
			tconfig.Installer.EXPECT().
				Upgrade(
					gomock.Any(),
					bundle.Tinkerbell,
					tconfig.DatacenterConfig.Spec.TinkerbellIP,
					tconfig.Management.KubeconfigFile,
					tconfig.DatacenterConfig.Spec.HookImagesURLPath,
				).
				Return(nil)
			tconfig.KubeClient.EXPECT().
				HasCRD(
					gomock.Any(),
					rufiounreleased.BaseboardManagementResourceName,
					tconfig.Management.KubeconfigFile,
				).
				Return(true, nil)

			// We minimally expect calls out to the cluster to retrieve BaseboardManagement and
			// Hardware resources.
			tconfig.KubeClient.EXPECT().
				AllBaseboardManagements(gomock.Any(), tconfig.Management.KubeconfigFile).
				Return(tc.BaseboardManagements, nil)
			tconfig.KubeClient.EXPECT().
				AllTinkerbellHardware(gomock.Any(), tconfig.Management.KubeconfigFile).
				Return(tc.Hardware, nil)

			if len(tc.ExpectMachines) > 0 {
				tconfig.Installer.EXPECT().
					GetNamespace().
					Return(stackNamespace)

				// Serialize the expected rufiov1#Machine objects into YAML so we can use gomock
				// to expect that value.
				serialized, err := yaml.Serialize(tc.ExpectMachines...)
				if err != nil {
					t.Fatalf("Could not serialize expected machines: %v", serialized)
				}
				expect := yaml.Join(serialized)

				tconfig.KubeClient.EXPECT().
					ApplyKubeSpecFromBytesWithNamespace(
						gomock.Any(),
						tconfig.Management,
						expect,
						stackNamespace,
					).
					Return(nil)
			}

			if len(tc.ExpectHardware) > 0 {
				// Serialize the expected tinkv1#Hardware objects into YAML so we can use gomock
				// to expect that value.
				serialized, err := yaml.Serialize(tc.ExpectHardware...)
				if err != nil {
					t.Fatalf("Could not serialize expected hardware: %v", err)
				}
				expect := yaml.Join(serialized)

				tconfig.KubeClient.EXPECT().
					ApplyKubeSpecFromBytesForce(gomock.Any(), tconfig.Management, expect).
					Return(nil)
			}

			// We always attempt to delete
			tconfig.KubeClient.EXPECT().
				DeleteCRD(
					gomock.Any(),
					rufiounreleased.BaseboardManagementResourceName,
					tconfig.Management.KubeconfigFile,
				).
				Return(nil)

			provider, err := tconfig.GetProvider()
			if err != nil {
				t.Fatalf("Couldn't create the provider: %v", err)
			}

			err = provider.PreCoreComponentsUpgrade(
				context.Background(),
				tconfig.Management,
				tconfig.ClusterSpec,
			)
			if err != nil {
				t.Fatalf("Received unexpected error: %v", err)
			}
		})
	}
}

// PopulateRufioV1MachneMeta populates m's TypeMeta with Rufio v1 Machine API Version and Kind.
func PopulateRufioV1MachineMeta(m rufiov1.Machine) rufiov1.Machine {
	m.TypeMeta = metav1.TypeMeta{
		APIVersion: "bmc.tinkerbell.org/v1alpha1",
		Kind:       "Machine",
	}
	return m
}

// PreCoreComponentsUpgradeTestConfig is a test helper that contains the necessary pieces for
// testing the PreCoreComponentsUpgrade functionality.
type PreCoreComponentsUpgradeTestConfig struct {
	Ctrl       *gomock.Controller
	Docker     *stackmocks.MockDocker
	Helm       *stackmocks.MockHelm
	KubeClient *mocks.MockProviderKubectlClient
	Installer  *stackmocks.MockStackInstaller
	Writer     *filewritermocks.MockFileWriter

	TinkerbellIP string

	ClusterSpec      *cluster.Spec
	DatacenterConfig *v1alpha1.TinkerbellDatacenterConfig
	MachineConfigs   map[string]*v1alpha1.TinkerbellMachineConfig
	Management       *types.Cluster
}

// NewPreCoreComponentsUpgradeTestConfig creates a new PreCoreComponentsUpgradeTestConfig with
// all mocks initialized and test data available.
func NewPreCoreComponentsUpgradeTestConfig(t *testing.T) *PreCoreComponentsUpgradeTestConfig {
	t.Helper()
	ctrl := gomock.NewController(t)
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	cfg := &PreCoreComponentsUpgradeTestConfig{
		Ctrl:             ctrl,
		Docker:           stackmocks.NewMockDocker(ctrl),
		Helm:             stackmocks.NewMockHelm(ctrl),
		KubeClient:       mocks.NewMockProviderKubectlClient(ctrl),
		Installer:        stackmocks.NewMockStackInstaller(ctrl),
		Writer:           filewritermocks.NewMockFileWriter(ctrl),
		TinkerbellIP:     "1.1.1.1",
		ClusterSpec:      clusterSpec,
		DatacenterConfig: datacenterConfig,
		MachineConfigs:   machineConfigs,
		Management:       &types.Cluster{KubeconfigFile: "kubeconfig-file"},
	}
	cfg.DatacenterConfig.Spec.TinkerbellIP = cfg.TinkerbellIP
	return cfg
}

// GetProvider retrieves a new Tinkerbell provider instance build using the mocks initialized
// in t.
func (t *PreCoreComponentsUpgradeTestConfig) GetProvider() (*Provider, error) {
	p, err := NewProvider(
		t.DatacenterConfig,
		t.MachineConfigs,
		t.ClusterSpec.Cluster,
		"",
		t.Writer,
		t.Docker,
		t.Helm,
		t.KubeClient,
		testIP,
		test.FakeNow,
		false,
		false,
	)
	if err != nil {
		return nil, err
	}
	p.SetStackInstaller(t.Installer)
	return p, nil
}

// WithStackUpgrade configures t mocks to get successfully reach Rufio CRD conversion.
func (t *PreCoreComponentsUpgradeTestConfig) WithStackUpgrade() *PreCoreComponentsUpgradeTestConfig {
	bundle := t.ClusterSpec.RootVersionsBundle()

	t.Installer.EXPECT().
		Upgrade(
			gomock.Any(),
			bundle.Tinkerbell,
			t.TinkerbellIP,
			t.Management.KubeconfigFile,
			t.DatacenterConfig.Spec.HookImagesURLPath,
		).
		Return(nil)
	t.KubeClient.EXPECT().
		HasCRD(
			gomock.Any(),
			rufiounreleased.BaseboardManagementResourceName,
			t.Management.KubeconfigFile,
		).
		Return(true, nil)
	return t
}

func newTinkerbellProvider(datacenterConfig *v1alpha1.TinkerbellDatacenterConfig, machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig, clusterConfig *v1alpha1.Cluster, writer filewriter.FileWriter, docker stack.Docker, helm stack.Helm, kubectl ProviderKubectlClient) *Provider {
	hardwareFile := "./testdata/hardware.csv"
	forceCleanup := false

	provider, err := NewProvider(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		hardwareFile,
		writer,
		docker,
		helm,
		kubectl,
		testIP,
		test.FakeNow,
		forceCleanup,
		false,
	)
	if err != nil {
		panic(err)
	}
	return provider
}

func TestProviderSetupAndValidateManagementProxySuccess(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_proxy.yaml"

	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()

	provider := newTinkerbellProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl)
	provider.stackInstaller = stackInstaller

	clusterSpec.ManagementCluster = &types.Cluster{Name: "test", KubeconfigFile: "kubeconfig-file"}
	clusterSpec.Cluster.Spec.ManagementCluster = v1alpha1.ManagementCluster{Name: "test-mgmt"}
	clusterSpec.Cluster.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "1.2.3.4:3128",
		HttpsProxy: "1.2.3.4:3128",
	}

	kubectl.EXPECT().GetUnprovisionedTinkerbellHardware(ctx, clusterSpec.ManagementCluster.KubeconfigFile, constants.EksaSystemNamespace).Return([]tinkv1alpha1.Hardware{}, nil)
	kubectl.EXPECT().GetProvisionedTinkerbellHardware(ctx, clusterSpec.ManagementCluster.KubeconfigFile, constants.EksaSystemNamespace).Return([]tinkv1alpha1.Hardware{}, nil)
	kubectl.EXPECT().GetEksaCluster(ctx, clusterSpec.ManagementCluster, clusterSpec.Cluster.Spec.ManagementCluster.Name).Return(clusterSpec.Cluster, nil)
	stackInstaller.EXPECT().AddNoProxyIP(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host).Return()
	kubectl.EXPECT().ApplyKubeSpecFromBytesForce(ctx, clusterSpec.ManagementCluster, gomock.Any()).Return(nil)
	kubectl.EXPECT().WaitForRufioMachines(ctx, clusterSpec.ManagementCluster, "5m", "Contactable", gomock.Any()).Return(nil)

	err := provider.SetupAndValidateUpgradeCluster(ctx, clusterSpec.ManagementCluster, clusterSpec, clusterSpec)
	if err != nil {
		t.Fatalf("Received unexpected error: %v", err)
	}
}

func TestProviderSetupAndValidateManagementProxyError(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_proxy.yaml"

	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()

	provider := newTinkerbellProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer, docker, helm, kubectl)
	provider.stackInstaller = stackInstaller

	clusterSpec.ManagementCluster = &types.Cluster{Name: "test", KubeconfigFile: "kubeconfig-file"}
	clusterSpec.Cluster.Spec.ManagementCluster = v1alpha1.ManagementCluster{Name: "test-mgmt"}
	clusterSpec.Cluster.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "1.2.3.4:3128",
		HttpsProxy: "1.2.3.4:3128",
	}

	kubectl.EXPECT().GetUnprovisionedTinkerbellHardware(ctx, clusterSpec.ManagementCluster.KubeconfigFile, constants.EksaSystemNamespace).Return([]tinkv1alpha1.Hardware{}, nil)
	kubectl.EXPECT().GetProvisionedTinkerbellHardware(ctx, clusterSpec.ManagementCluster.KubeconfigFile, constants.EksaSystemNamespace).Return([]tinkv1alpha1.Hardware{}, nil)
	kubectl.EXPECT().GetEksaCluster(ctx, clusterSpec.ManagementCluster, clusterSpec.Cluster.Spec.ManagementCluster.Name).Return(clusterSpec.Cluster, fmt.Errorf("error getting management cluster data"))

	err := provider.SetupAndValidateUpgradeCluster(ctx, clusterSpec.ManagementCluster, clusterSpec, clusterSpec)
	assertError(t, "error getting management cluster data", err)
}

package tinkerbell

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	tinkv1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	rufiov1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell/rufio"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	filewritermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/rufiounreleased"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	stackmocks "github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
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
		tconfig.managementComponents,
		tconfig.ClusterSpec,
	)

	if err != nil {
		t.Fatalf("Received unexpected error: %v", err)
	}
}

func TestProviderPreCoreComponentsUpgrade_StackUpgradeError(t *testing.T) {
	tconfig := NewPreCoreComponentsUpgradeTestConfig(t)

	expect := "foobar"
	tconfig.Installer.EXPECT().
		Upgrade(
			gomock.Any(),
			tconfig.managementComponents.Tinkerbell,
			tconfig.DatacenterConfig.Spec.TinkerbellIP,
			tconfig.Management.KubeconfigFile,
			tconfig.DatacenterConfig.Spec.HookImagesURLPath,
			gomock.Any(),
		).
		Return(errors.New(expect))

	provider, err := tconfig.GetProvider()
	if err != nil {
		t.Fatalf("Couldn't create the provider: %v", err)
	}

	err = provider.PreCoreComponentsUpgrade(context.Background(), tconfig.Management, tconfig.managementComponents, tconfig.ClusterSpec)
	if err == nil || !strings.Contains(err.Error(), expect) {
		t.Fatalf("Expected error containing '%v'; Received '%v'", expect, err)
	}
}

func TestProviderPreCoreComponentsUpgrade_HasBaseboardManagementCRDError(t *testing.T) {
	tconfig := NewPreCoreComponentsUpgradeTestConfig(t)

	tconfig.Installer.EXPECT().
		Upgrade(
			gomock.Any(),
			tconfig.managementComponents.Tinkerbell,
			tconfig.TinkerbellIP,
			tconfig.Management.KubeconfigFile,
			tconfig.DatacenterConfig.Spec.HookImagesURLPath,
			gomock.Any(),
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

	err = provider.PreCoreComponentsUpgrade(context.Background(), tconfig.Management, tconfig.managementComponents, tconfig.ClusterSpec)
	if err == nil || !strings.Contains(err.Error(), expect) {
		t.Fatalf("Expected error containing '%v'; Received '%v'", expect, err)
	}
}

func TestProviderPreCoreComponentsUpgrade_NoBaseboardManagementCRD(t *testing.T) {
	tconfig := NewPreCoreComponentsUpgradeTestConfig(t)

	tconfig.Installer.EXPECT().
		Upgrade(
			gomock.Any(),
			tconfig.managementComponents.Tinkerbell,
			tconfig.TinkerbellIP,
			tconfig.Management.KubeconfigFile,
			tconfig.DatacenterConfig.Spec.HookImagesURLPath,
			gomock.Any(),
		).
		Return(nil)

	tconfig.KubeClient.EXPECT().
		HasCRD(gomock.Any(), rufiounreleased.BaseboardManagementResourceName, tconfig.Management.KubeconfigFile).
		Return(false, nil)

	provider, err := tconfig.GetProvider()
	if err != nil {
		t.Fatalf("Couldn't create the provider: %v", err)
	}

	err = provider.PreCoreComponentsUpgrade(context.Background(), tconfig.Management, tconfig.managementComponents, tconfig.ClusterSpec)
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

			// Configure the mocks to successfully upgrade the Tinkerbell stack using the installer
			// and identify the need to convert deprecated Rufio custom resources.
			tconfig.Installer.EXPECT().
				Upgrade(
					gomock.Any(),
					tconfig.managementComponents.Tinkerbell,
					tconfig.DatacenterConfig.Spec.TinkerbellIP,
					tconfig.Management.KubeconfigFile,
					tconfig.DatacenterConfig.Spec.HookImagesURLPath,
					gomock.Any(),
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
				tconfig.managementComponents,
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

	ClusterSpec          *cluster.Spec
	managementComponents *cluster.ManagementComponents

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
		Ctrl:                 ctrl,
		Docker:               stackmocks.NewMockDocker(ctrl),
		Helm:                 stackmocks.NewMockHelm(ctrl),
		KubeClient:           mocks.NewMockProviderKubectlClient(ctrl),
		Installer:            stackmocks.NewMockStackInstaller(ctrl),
		Writer:               filewritermocks.NewMockFileWriter(ctrl),
		TinkerbellIP:         "1.1.1.1",
		ClusterSpec:          clusterSpec,
		managementComponents: cluster.ManagementComponentsFromBundles(clusterSpec.Bundles),
		DatacenterConfig:     datacenterConfig,
		MachineConfigs:       machineConfigs,
		Management:           &types.Cluster{KubeconfigFile: "kubeconfig-file"},
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

func TestProvider_ValidateNewSpec_NoChanges(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
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

	cluster := &types.Cluster{Name: "test", KubeconfigFile: "kubeconfig-file"}
	clusterSpec.ManagementCluster = cluster

	provider := newTinkerbellProvider(datacenterConfig, machineConfigs, clusterSpec.Cluster, writer,
		docker, helm, kubectl)
	provider.stackInstaller = stackInstaller

	kubectl.EXPECT().
		GetEksaCluster(ctx, clusterSpec.ManagementCluster, clusterSpec.Cluster.Spec.ManagementCluster.Name).
		Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().
		GetEksaTinkerbellDatacenterConfig(ctx, clusterSpec.Cluster.Spec.DatacenterRef.Name,
			cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).
		Return(datacenterConfig, nil)

	for _, v := range machineConfigs {
		kubectl.EXPECT().
			GetEksaTinkerbellMachineConfig(ctx, v.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).
			Return(v, nil)
	}

	err := provider.ValidateNewSpec(ctx, clusterSpec.ManagementCluster, clusterSpec)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProvider_ValidateNewSpec_ChangeWorkerNodeGroupMachineRef(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	currentClusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()

	cluster := &types.Cluster{Name: "test", KubeconfigFile: "kubeconfig-file"}
	currentClusterSpec.ManagementCluster = cluster

	desiredClusterSpec := currentClusterSpec.DeepCopy()

	// Change an existing worker node groups machine config reference in the desired spec.
	newMachineCfgName := "additiona-machine-config"
	machineConfigs[newMachineCfgName] = &v1alpha1.TinkerbellMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: newMachineCfgName,
		},
	}
	desiredClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = newMachineCfgName

	provider := newTinkerbellProvider(datacenterConfig, machineConfigs, currentClusterSpec.Cluster, writer,
		docker, helm, kubectl)
	provider.stackInstaller = stackInstaller

	kubectl.EXPECT().
		GetEksaCluster(ctx, desiredClusterSpec.ManagementCluster,
			desiredClusterSpec.Cluster.Spec.ManagementCluster.Name).
		Return(currentClusterSpec.Cluster, nil)
	kubectl.EXPECT().
		GetEksaTinkerbellDatacenterConfig(ctx, currentClusterSpec.Cluster.Spec.DatacenterRef.Name,
			cluster.KubeconfigFile, currentClusterSpec.Cluster.Namespace).
		Return(datacenterConfig, nil)

	err := provider.ValidateNewSpec(ctx, desiredClusterSpec.ManagementCluster, desiredClusterSpec)
	if err == nil || !strings.Contains(err.Error(), "worker node group machine config reference is immutable") {
		t.Fatalf("Expected error containing 'worker node group machine config reference is immutable' but received: %v", err)
	}
}

func TestProvider_ValidateNewSpec_ChangeControlPlaneMachineRefToExisting(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	currentClusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()

	cluster := &types.Cluster{Name: "test", KubeconfigFile: "kubeconfig-file"}
	currentClusterSpec.ManagementCluster = cluster

	desiredClusterSpec := currentClusterSpec.DeepCopy()

	// Change an existing worker node groups machine config reference to an existing machine config.
	desiredClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "test-md"

	provider := newTinkerbellProvider(datacenterConfig, machineConfigs, currentClusterSpec.Cluster, writer,
		docker, helm, kubectl)
	provider.stackInstaller = stackInstaller

	kubectl.EXPECT().
		GetEksaCluster(ctx, desiredClusterSpec.ManagementCluster,
			desiredClusterSpec.Cluster.Spec.ManagementCluster.Name).
		Return(currentClusterSpec.Cluster, nil)
	kubectl.EXPECT().
		GetEksaTinkerbellDatacenterConfig(ctx, currentClusterSpec.Cluster.Spec.DatacenterRef.Name,
			cluster.KubeconfigFile, currentClusterSpec.Cluster.Namespace).
		Return(datacenterConfig, nil)

	err := provider.ValidateNewSpec(ctx, desiredClusterSpec.ManagementCluster, desiredClusterSpec)
	if err == nil || !strings.Contains(err.Error(), "control plane machine config reference is immutable") {
		t.Fatalf("Expected error containing 'control plane machine config reference is immutable' but received: %v", err)
	}
}

func TestProvider_ValidateNewSpec_ChangeWorkerNodeGroupMachineRefToExisting(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	currentClusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()

	cluster := &types.Cluster{Name: "test", KubeconfigFile: "kubeconfig-file"}
	currentClusterSpec.ManagementCluster = cluster

	desiredClusterSpec := currentClusterSpec.DeepCopy()

	// Change an existing worker node groups machine config reference to an existing machine config.
	desiredClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = "test-cp"

	provider := newTinkerbellProvider(datacenterConfig, machineConfigs, currentClusterSpec.Cluster, writer,
		docker, helm, kubectl)
	provider.stackInstaller = stackInstaller

	kubectl.EXPECT().
		GetEksaCluster(ctx, desiredClusterSpec.ManagementCluster,
			desiredClusterSpec.Cluster.Spec.ManagementCluster.Name).
		Return(currentClusterSpec.Cluster, nil)
	kubectl.EXPECT().
		GetEksaTinkerbellDatacenterConfig(ctx, currentClusterSpec.Cluster.Spec.DatacenterRef.Name,
			cluster.KubeconfigFile, currentClusterSpec.Cluster.Namespace).
		Return(datacenterConfig, nil)

	controlPlaneMachineCfgName := currentClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	kubectl.EXPECT().
		GetEksaTinkerbellMachineConfig(ctx, controlPlaneMachineCfgName, cluster.KubeconfigFile,
			desiredClusterSpec.Cluster.Namespace).
		Return(machineConfigs[controlPlaneMachineCfgName], nil).
		// The implementation of ValidateNewSpec iterates over machine references in a non-deterministic
		// order. This means sometimes we inspect the control plane machine config ref first and
		// sometimes we inspect the worker node group machine config ref first. AnyTimes() accounts
		// for the latter ensuring we tolerate 0 attempts to look at the control plane group.
		AnyTimes()

	err := provider.ValidateNewSpec(ctx, desiredClusterSpec.ManagementCluster, desiredClusterSpec)
	if err == nil || !strings.Contains(err.Error(), "worker node group machine config reference is immutable") {
		t.Fatalf("Expected error containing 'worker node group machine config reference is immutable' but received: %v", err)
	}
}

func TestProvider_ValidateNewSpec_NewWorkerNodeGroup(t *testing.T) {
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	currentClusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	docker := stackmocks.NewMockDocker(mockCtrl)
	helm := stackmocks.NewMockHelm(mockCtrl)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	stackInstaller := stackmocks.NewMockStackInstaller(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()

	cluster := &types.Cluster{Name: "test", KubeconfigFile: "kubeconfig-file"}
	currentClusterSpec.ManagementCluster = cluster

	desiredClusterSpec := currentClusterSpec.DeepCopy()

	// Add an extra worker node group to the desired configuration with its associated machine
	// config. The machine configs are plumbed in via the Tinkerbell provider constructor func.
	newMachineCfgName := "additional-machine-config"
	newWorkerNodeGroupName := "additional-worker-node-group"
	desiredWorkerNodeGroups := &desiredClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations
	*desiredWorkerNodeGroups = append(*desiredWorkerNodeGroups, v1alpha1.WorkerNodeGroupConfiguration{
		Name:  newWorkerNodeGroupName,
		Count: ptr.Int(1),
		MachineGroupRef: &v1alpha1.Ref{
			Name: newMachineCfgName,
		},
	})
	machineConfigs[newMachineCfgName] = &v1alpha1.TinkerbellMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: newMachineCfgName,
		},
	}

	provider := newTinkerbellProvider(datacenterConfig, machineConfigs, currentClusterSpec.Cluster, writer,
		docker, helm, kubectl)
	provider.stackInstaller = stackInstaller

	kubectl.EXPECT().
		GetEksaCluster(ctx, desiredClusterSpec.ManagementCluster, desiredClusterSpec.Cluster.Spec.ManagementCluster.Name).
		Return(currentClusterSpec.Cluster, nil)
	kubectl.EXPECT().
		GetEksaTinkerbellDatacenterConfig(ctx, currentClusterSpec.Cluster.Spec.DatacenterRef.Name,
			cluster.KubeconfigFile, currentClusterSpec.Cluster.Namespace).
		Return(datacenterConfig, nil)

	for name, v := range machineConfigs {
		// Don't expect a request when its a new machine config.
		if name == newMachineCfgName {
			continue
		}

		kubectl.EXPECT().
			GetEksaTinkerbellMachineConfig(ctx, v.Name, cluster.KubeconfigFile, desiredClusterSpec.Cluster.Namespace).
			Return(v, nil)
	}

	err := provider.ValidateNewSpec(ctx, desiredClusterSpec.ManagementCluster, desiredClusterSpec)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProviderValidateAvailableHardwareOnlyCPUpgradeSuccess(t *testing.T) {
	clusterSpecManifest := "cluster_osimage_machine_config.yaml"
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
	catalogue := hardware.NewCatalogue()
	newCluster := clusterSpec.DeepCopy()
	newCluster.Cluster.Spec.KubernetesVersion = v1alpha1.Kube122
	cpRef := newCluster.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	machineConfigs[cpRef].Spec.OSImageURL = "https://ubuntu-1-22.gz"
	_ = catalogue.InsertHardware(&tinkv1.Hardware{ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{"type": "cp"},
	}})
	provider.catalogue = catalogue
	err := provider.validateAvailableHardwareForUpgrade(ctx, clusterSpec, newCluster)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProviderValidateAvailableHardwareOnlyCPUpgradeError(t *testing.T) {
	clusterSpecManifest := "cluster_osimage_machine_config.yaml"
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
	catalogue := hardware.NewCatalogue()
	newCluster := clusterSpec.DeepCopy()
	newCluster.Cluster.Spec.KubernetesVersion = v1alpha1.Kube122
	cpRef := newCluster.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	machineConfigs[cpRef].Spec.OSImageURL = "https://ubuntu-1-22.gz"
	provider.catalogue = catalogue
	err := provider.validateAvailableHardwareForUpgrade(ctx, clusterSpec, newCluster)
	if err == nil || !strings.Contains(err.Error(), "for rolling upgrade, minimum hardware count not met for selector '{\"type\":\"cp\"}'") {
		t.Fatal(err)
	}
}

func TestProviderValidateAvailableHardwareOnlyWorkerUpgradeSuccess(t *testing.T) {
	clusterSpecManifest := "cluster_osimage_machine_config.yaml"
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
	catalogue := hardware.NewCatalogue()
	newCluster := clusterSpec.DeepCopy()
	kube122 := v1alpha1.Kube122
	newCluster.Cluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube122
	_ = catalogue.InsertHardware(&tinkv1.Hardware{ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{"type": "worker"},
	}})
	wngRef := newCluster.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	machineConfigs[wngRef].Spec.OSImageURL = "https://ubuntu-1-22.gz"
	provider.catalogue = catalogue
	err := provider.validateAvailableHardwareForUpgrade(ctx, clusterSpec, newCluster)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProviderValidateAvailableHardwareOnlyWorkerUpgradeError(t *testing.T) {
	clusterSpecManifest := "cluster_osimage_machine_config.yaml"
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
	catalogue := hardware.NewCatalogue()
	newCluster := clusterSpec.DeepCopy()
	kube122 := v1alpha1.Kube122
	newCluster.Cluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube122
	wngRef := newCluster.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	machineConfigs[wngRef].Spec.OSImageURL = "https://ubuntu-1-22.gz"
	provider.catalogue = catalogue
	err := provider.validateAvailableHardwareForUpgrade(ctx, clusterSpec, newCluster)
	if err == nil || !strings.Contains(err.Error(), "for rolling upgrade, minimum hardware count not met for selector '{\"type\":\"worker\"}'") {
		t.Fatal(err)
	}
}

func TestProviderValidateAvailableHardwareEksaVersionUpgradeSuccess(t *testing.T) {
	clusterSpecManifest := "cluster_osimage_machine_config.yaml"
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
	catalogue := hardware.NewCatalogue()
	newCluster := clusterSpec.DeepCopy()
	newCluster.Bundles.Spec.Number++
	_ = catalogue.InsertHardware(&tinkv1.Hardware{ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{"type": "cp"},
	}})
	_ = catalogue.InsertHardware(&tinkv1.Hardware{ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{"type": "worker"},
	}})
	provider.catalogue = catalogue
	err := provider.validateAvailableHardwareForUpgrade(ctx, clusterSpec, newCluster)
	if err != nil {
		t.Fatal(err)
	}
}

package vsphere

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockswriter "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/govmomi"
	govmomi_mocks "github.com/aws/eks-anywhere/pkg/govmomi/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	testClusterConfigMainFilename          = "cluster_main.yaml"
	testClusterConfigMain121Filename       = "cluster_main_121.yaml"
	testClusterConfigMain121CPOnlyFilename = "cluster_main_121_cp_only.yaml"
	testClusterConfigWithCPUpgradeStrategy = "cluster_main_121_cp_upgrade_strategy.yaml"
	testClusterConfigWithMDUpgradeStrategy = "cluster_main_121_md_upgrade_strategy.yaml"
	testClusterConfigRedhatFilename        = "cluster_redhat_external_etcd.yaml"
	testDataDir                            = "testdata"
	expectedVSphereName                    = "vsphere"
	expectedVSphereUsername                = "vsphere_username"
	expectedVSpherePassword                = "vsphere_password"
	expectedVSphereServer                  = "vsphere_server"
	expectedExpClusterResourceSet          = "expClusterResourceSetKey"
	eksd119Release                         = "kubernetes-1-19-eks-4"
	eksd119ReleaseTag                      = "eksdRelease:kubernetes-1-19-eks-4"
	eksd121ReleaseTag                      = "eksdRelease:kubernetes-1-21-eks-4"
	ubuntuOSTag                            = "os:ubuntu"
	bottlerocketOSTag                      = "os:bottlerocket"
	testTemplate                           = "/SDDC-Datacenter/vm/Templates/ubuntu-1804-kube-v1.19.6"
)

type DummyProviderGovcClient struct {
	osTag string
}

func NewDummyProviderGovcClient() *DummyProviderGovcClient {
	return &DummyProviderGovcClient{osTag: ubuntuOSTag}
}

func (pc *DummyProviderGovcClient) TemplateHasSnapshot(ctx context.Context, template string) (bool, error) {
	return true, nil
}

func (pc *DummyProviderGovcClient) GetWorkloadAvailableSpace(ctx context.Context, datastore string) (float64, error) {
	return math.MaxFloat64, nil
}

func (pc *DummyProviderGovcClient) DeployTemplate(ctx context.Context, datacenterConfig *v1alpha1.VSphereDatacenterConfig) error {
	return nil
}

func (pc *DummyProviderGovcClient) ValidateVCenterConnection(ctx context.Context, server string) error {
	return nil
}

func (pc *DummyProviderGovcClient) ValidateVCenterAuthentication(ctx context.Context) error {
	return nil
}

func (pc *DummyProviderGovcClient) IsCertSelfSigned(ctx context.Context) bool {
	return false
}

func (pc *DummyProviderGovcClient) GetCertThumbprint(ctx context.Context) (string, error) {
	return "", nil
}

func (pc *DummyProviderGovcClient) ConfigureCertThumbprint(ctx context.Context, server, thumbprint string) error {
	return nil
}

func (pc *DummyProviderGovcClient) DatacenterExists(ctx context.Context, datacenter string) (bool, error) {
	return true, nil
}

func (pc *DummyProviderGovcClient) NetworkExists(ctx context.Context, network string) (bool, error) {
	return true, nil
}

func (pc *DummyProviderGovcClient) ValidateVCenterSetupMachineConfig(ctx context.Context, datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfig *v1alpha1.VSphereMachineConfig, selfSigned *bool) error {
	return nil
}

func (pc *DummyProviderGovcClient) SearchTemplate(ctx context.Context, datacenter, template string) (string, error) {
	return template, nil
}

func (pc *DummyProviderGovcClient) LibraryElementExists(ctx context.Context, library string) (bool, error) {
	return true, nil
}

func (pc *DummyProviderGovcClient) GetLibraryElementContentVersion(ctx context.Context, element string) (string, error) {
	return "", nil
}

func (pc *DummyProviderGovcClient) DeleteLibraryElement(ctx context.Context, element string) error {
	return nil
}

func (pc *DummyProviderGovcClient) CreateLibrary(ctx context.Context, datastore, library string) error {
	return nil
}

func (pc *DummyProviderGovcClient) DeployTemplateFromLibrary(ctx context.Context, templateDir, templateName, library, datacenter, datastore, network, resourcePool string, resizeDisk2 bool) error {
	return nil
}

func (pc *DummyProviderGovcClient) ResizeDisk(ctx context.Context, template, diskName string, diskSizeInGB int) error {
	return nil
}

func (pc *DummyProviderGovcClient) ImportTemplate(ctx context.Context, library, ovaURL, name string) error {
	return nil
}

func (pc *DummyProviderGovcClient) GetVMDiskSizeInGB(ctx context.Context, vm, datacenter string) (int, error) {
	return 25, nil
}

func (pc *DummyProviderGovcClient) GetHardDiskSize(ctx context.Context, vm, datacenter string) (map[string]float64, error) {
	return map[string]float64{"Hard disk 1": 23068672}, nil
}

func (pc *DummyProviderGovcClient) GetResourcePoolInfo(ctx context.Context, datacenter, resourcePool string, args ...string) (map[string]int, error) {
	return map[string]int{"Memory_Available": -1}, nil
}

func (pc *DummyProviderGovcClient) GetTags(ctx context.Context, path string) (tags []string, err error) {
	return []string{eksd119ReleaseTag, eksd121ReleaseTag, pc.osTag}, nil
}

func (pc *DummyProviderGovcClient) ListTags(ctx context.Context) ([]executables.Tag, error) {
	return nil, nil
}

func (pc *DummyProviderGovcClient) CreateTag(ctx context.Context, tag, category string) error {
	return nil
}

func (pc *DummyProviderGovcClient) AddTag(ctx context.Context, path, tag string) error {
	return nil
}

func (pc *DummyProviderGovcClient) ListCategories(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (pc *DummyProviderGovcClient) CreateCategoryForVM(ctx context.Context, name string) error {
	return nil
}

func (pc *DummyProviderGovcClient) AddUserToGroup(ctx context.Context, name string, username string) error {
	return nil
}

func (pc *DummyProviderGovcClient) CreateGroup(ctx context.Context, name string) error {
	return nil
}

func (pc *DummyProviderGovcClient) CreateRole(ctx context.Context, name string, privileges []string) error {
	return nil
}

func (pc *DummyProviderGovcClient) CreateUser(ctx context.Context, username string, password string) error {
	return nil
}

func (pc *DummyProviderGovcClient) UserExists(ctx context.Context, username string) (bool, error) {
	return true, nil
}

func (pc *DummyProviderGovcClient) GroupExists(ctx context.Context, name string) (bool, error) {
	return true, nil
}

func (pc *DummyProviderGovcClient) RoleExists(ctx context.Context, name string) (bool, error) {
	return false, nil
}

func (pc *DummyProviderGovcClient) SetGroupRoleOnObject(ctx context.Context, principal string, role string, object string, domain string) error {
	return nil
}

func givenClusterConfig(t *testing.T, fileName string) *v1alpha1.Cluster {
	return givenClusterSpec(t, fileName).Cluster
}

func givenClusterSpec(t *testing.T, fileName string) *cluster.Spec {
	return test.NewFullClusterSpec(t, path.Join(testDataDir, fileName))
}

func givenEmptyClusterSpec() *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].KubeVersion = "1.19"
		s.VersionsBundles["1.19"].EksD.Name = eksd119Release
		s.Cluster.Namespace = "test-namespace"
		s.VSphereDatacenter = &v1alpha1.VSphereDatacenterConfig{}
	})
}

func givenDatacenterConfig(t *testing.T, fileName string) *v1alpha1.VSphereDatacenterConfig {
	datacenterConfig, err := v1alpha1.GetVSphereDatacenterConfig(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file: %v", err)
	}
	return datacenterConfig
}

func givenProvider(t *testing.T) *vsphereProvider {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider := newProviderWithKubectl(
		t,
		datacenterConfig,
		clusterConfig,
		nil,
		ipValidator,
	)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}
	return provider
}

func workerNodeGroup1MachineDeployment() *clusterv1.MachineDeployment {
	return &clusterv1.MachineDeployment{
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							Name: "test-md-0-template-1234567890000",
						},
					},
				},
			},
		},
	}
}

func workerNodeGroup2MachineDeployment() *clusterv1.MachineDeployment {
	return &clusterv1.MachineDeployment{
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							Name: "test-md-1-template-1234567890000",
						},
					},
				},
			},
		},
	}
}

func setupContext(t *testing.T) {
	t.Setenv(config.EksavSphereUsernameKey, expectedVSphereUsername)
	t.Setenv(vSphereUsernameKey, os.Getenv(config.EksavSphereUsernameKey))
	t.Setenv(config.EksavSpherePasswordKey, expectedVSpherePassword)
	t.Setenv(vSpherePasswordKey, os.Getenv(config.EksavSpherePasswordKey))
	t.Setenv(vSphereServerKey, expectedVSphereServer)
	t.Setenv(expClusterResourceSetKey, expectedExpClusterResourceSet)
}

type providerTest struct {
	*WithT
	t                                  *testing.T
	ctx                                context.Context
	managementCluster, workloadCluster *types.Cluster
	provider                           *vsphereProvider
	cluster                            *v1alpha1.Cluster
	clusterSpec                        *cluster.Spec
	datacenterConfig                   *v1alpha1.VSphereDatacenterConfig
	machineConfigs                     map[string]*v1alpha1.VSphereMachineConfig
	kubectl                            *mocks.MockProviderKubectlClient
	govc                               *mocks.MockProviderGovcClient
	clientBuilder                      *mockVSphereClientBuilder
	ipValidator                        *mocks.MockIPValidator
}

func newProviderTest(t *testing.T) *providerTest {
	setupContext(t)
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(ctrl)
	govc := mocks.NewMockProviderGovcClient(ctrl)
	vscb, _ := newMockVSphereClientBuilder(ctrl)
	ipValidator := mocks.NewMockIPValidator(ctrl)
	spec := givenClusterSpec(t, testClusterConfigMainFilename)
	p := &providerTest{
		t:     t,
		WithT: NewWithT(t),
		ctx:   context.Background(),
		managementCluster: &types.Cluster{
			Name:           "m-cluster",
			KubeconfigFile: "kubeconfig-m.kubeconfig",
		},
		workloadCluster: &types.Cluster{
			Name:           "test",
			KubeconfigFile: "kubeconfig-w.kubeconfig",
		},
		cluster:          spec.Cluster,
		clusterSpec:      spec,
		datacenterConfig: spec.VSphereDatacenter,
		machineConfigs:   spec.VSphereMachineConfigs,
		kubectl:          kubectl,
		govc:             govc,
		clientBuilder:    vscb,
		ipValidator:      ipValidator,
	}
	p.buildNewProvider()

	return p
}

func (tt *providerTest) setExpectationsForDefaultDiskAndCloneModeGovcCalls() {
	for _, m := range tt.machineConfigs {
		tt.govc.EXPECT().GetVMDiskSizeInGB(tt.ctx, m.Spec.Template, tt.datacenterConfig.Spec.Datacenter).Return(25, nil)
		tt.govc.EXPECT().TemplateHasSnapshot(tt.ctx, m.Spec.Template).Return(true, nil)
	}
}

func (tt *providerTest) setExpectationForVCenterValidation() {
	tt.govc.EXPECT().IsCertSelfSigned(tt.ctx).Return(false)
	tt.govc.EXPECT().DatacenterExists(tt.ctx, tt.datacenterConfig.Spec.Datacenter).Return(true, nil)
	tt.govc.EXPECT().NetworkExists(tt.ctx, tt.datacenterConfig.Spec.Network).Return(true, nil)
}

func (tt *providerTest) setExpectationForSetup() {
	tt.govc.EXPECT().ValidateVCenterConnection(tt.ctx, tt.datacenterConfig.Spec.Server).Return(nil)
	tt.govc.EXPECT().ValidateVCenterAuthentication(tt.ctx).Return(nil)
	tt.govc.EXPECT().ConfigureCertThumbprint(tt.ctx, tt.datacenterConfig.Spec.Server, tt.datacenterConfig.Spec.Thumbprint).Return(nil)
}

func (tt *providerTest) setExpectationsForMachineConfigsVCenterValidation() {
	for _, m := range tt.machineConfigs {
		var b bool
		tt.govc.EXPECT().ValidateVCenterSetupMachineConfig(tt.ctx, tt.datacenterConfig, m, &b).Return(nil)
	}
}

func (tt *providerTest) buildNewProvider() {
	tt.provider = newProvider(
		tt.t,
		tt.clusterSpec.VSphereDatacenter,
		tt.clusterSpec.Cluster,
		tt.govc,
		tt.kubectl,
		NewValidator(tt.govc, tt.clientBuilder),
		tt.ipValidator,
	)
}

func TestNewProvider(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	govc := NewDummyProviderGovcClient()
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	_, writer := test.NewWriter(t)
	skipIPCheck := true
	skippedValidations := map[string]bool{}

	provider := NewProvider(
		datacenterConfig,
		clusterConfig,
		govc,
		kubectl,
		writer,
		ipValidator,
		time.Now,
		skipIPCheck,
		skippedValidations,
	)

	if provider == nil {
		t.Fatalf("provider object is nil")
	}
	if provider.validator == nil {
		t.Fatalf("validator not configured")
	}
}

func TestNewProviderCustomNet(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider := newProviderWithKubectl(
		t,
		datacenterConfig,
		clusterConfig,
		kubectl,
		ipValidator,
	)

	if provider == nil {
		t.Fatalf("provider object is nil")
	}
}

func newProviderWithKubectl(t *testing.T, datacenterConfig *v1alpha1.VSphereDatacenterConfig, clusterConfig *v1alpha1.Cluster, kubectl ProviderKubectlClient, ipValidator IPValidator) *vsphereProvider {
	ctrl := gomock.NewController(t)
	govc := NewDummyProviderGovcClient()
	vscb, _ := newMockVSphereClientBuilder(ctrl)
	v := NewValidator(govc, vscb)
	return newProvider(
		t,
		datacenterConfig,
		clusterConfig,
		govc,
		kubectl,
		v,
		ipValidator,
	)
}

func newProviderWithGovc(t *testing.T, datacenterConfig *v1alpha1.VSphereDatacenterConfig, clusterConfig *v1alpha1.Cluster, govc ProviderGovcClient) *vsphereProvider {
	ctrl := gomock.NewController(t)
	vscb, _ := newMockVSphereClientBuilder(ctrl)
	v := NewValidator(govc, vscb)
	kubectl := mocks.NewMockProviderKubectlClient(ctrl)
	ipValidator := mocks.NewMockIPValidator(ctrl)
	return newProvider(
		t,
		datacenterConfig,
		clusterConfig,
		govc,
		kubectl,
		v,
		ipValidator,
	)
}

type mockVSphereClientBuilder struct {
	vsc *govmomi_mocks.MockVSphereClient
}

func (mvscb *mockVSphereClientBuilder) Build(ctx context.Context, host string, username string, password string, insecure bool, datacenter string) (govmomi.VSphereClient, error) {
	return mvscb.vsc, nil
}

func setDefaultVSphereClientMock(vsc *govmomi_mocks.MockVSphereClient) error {
	vsc.EXPECT().Username().Return("foobar").AnyTimes()

	var privs []string
	err := json.Unmarshal([]byte(config.VSphereAdminPrivsFile), &privs)
	if err != nil {
		return err
	}

	vsc.EXPECT().GetPrivsOnEntity(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(privs, nil).AnyTimes()

	return nil
}

func newMockVSphereClientBuilder(ctrl *gomock.Controller) (*mockVSphereClientBuilder, error) {
	vsc := govmomi_mocks.NewMockVSphereClient(ctrl)
	err := setDefaultVSphereClientMock(vsc)
	mvscb := mockVSphereClientBuilder{vsc}
	return &mvscb, err
}

func newProvider(t *testing.T, datacenterConfig *v1alpha1.VSphereDatacenterConfig, clusterConfig *v1alpha1.Cluster, govc ProviderGovcClient, kubectl ProviderKubectlClient, v *Validator, ipValidator IPValidator) *vsphereProvider {
	_, writer := test.NewWriter(t)

	return NewProviderCustomNet(
		datacenterConfig,
		clusterConfig,
		govc,
		kubectl,
		writer,
		ipValidator,
		test.FakeNow,
		false,
		v,
		map[string]bool{},
	)
}

func TestProviderGenerateCAPISpecForUpgradeUpdateMachineTemplate(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantCPFile        string
		wantMDFile        string
	}{
		{
			testName:          "minimal",
			clusterconfigFile: "cluster_minimal.yaml",
			wantCPFile:        "testdata/expected_results_minimal_cp.yaml",
			wantMDFile:        "testdata/expected_results_minimal_md.yaml",
		},
		{
			testName:          "minimal_autoscaler",
			clusterconfigFile: "cluster_minimal_autoscaling.yaml",
			wantCPFile:        "testdata/expected_results_minimal_cp.yaml",
			wantMDFile:        "testdata/expected_results_minimal_autoscaling_md.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			setupContext(t)
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			vsphereDatacenter := &v1alpha1.VSphereDatacenterConfig{
				Spec: v1alpha1.VSphereDatacenterConfigSpec{},
			}
			vsphereMachineConfig := firstMachineConfig(clusterSpec).DeepCopy()

			ipValidator := mocks.NewMockIPValidator(mockCtrl)
			ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereDatacenter, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}

			err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
			if err != nil {
				t.Fatalf("failed to setup and validate: %v", err)
			}

			cp, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, clusterSpec.DeepCopy())
			if err != nil {
				t.Fatalf("failed to generate cluster api spec contents: %v", err)
			}

			test.AssertContentToFile(t, string(cp), tt.wantCPFile)

			test.AssertContentToFile(t, string(md), tt.wantMDFile)
		})
	}
}

func firstMachineConfig(spec *cluster.Spec) *v1alpha1.VSphereMachineConfig {
	var mc *v1alpha1.VSphereMachineConfig
	for _, m := range spec.VSphereMachineConfigs {
		mc = m
		break
	}
	return mc
}

func getMachineConfig(spec *cluster.Spec, name string) *v1alpha1.VSphereMachineConfig {
	if mc, ok := spec.VSphereMachineConfigs[name]; ok {
		return mc
	}
	return nil
}

func TestProviderGenerateCAPISpecForUpgradeOIDC(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantCPFile        string
	}{
		{
			testName:          "with minimal oidc",
			clusterconfigFile: "cluster_minimal_oidc.yaml",
			wantCPFile:        "testdata/expected_results_minimal_oidc_cp.yaml",
		},
		{
			testName:          "with full oidc",
			clusterconfigFile: "cluster_full_oidc.yaml",
			wantCPFile:        "testdata/expected_results_full_oidc_cp.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			setupContext(t)
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			vsphereDatacenter := &v1alpha1.VSphereDatacenterConfig{
				Spec: v1alpha1.VSphereDatacenterConfigSpec{},
			}

			vsphereMachineConfig := firstMachineConfig(clusterSpec).DeepCopy()

			ipValidator := mocks.NewMockIPValidator(mockCtrl)
			ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereDatacenter, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}

			err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
			if err != nil {
				t.Fatalf("failed to setup and validate: %v", err)
			}

			cp, _, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, clusterSpec.DeepCopy())
			if err != nil {
				t.Fatalf("failed to generate cluster api spec contents: %v", err)
			}

			test.AssertContentToFile(t, string(cp), tt.wantCPFile)
		})
	}
}

func TestProviderGenerateCAPISpecForUpgradeMultipleWorkerNodeGroups(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantMDFile        string
	}{
		{
			testName:          "adding a worker node group",
			clusterconfigFile: "cluster_main_multiple_worker_node_groups.yaml",
			wantMDFile:        "testdata/expected_results_minimal_add_worker_node_group.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			setupContext(t)
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			vsphereDatacenter := &v1alpha1.VSphereDatacenterConfig{
				Spec: v1alpha1.VSphereDatacenterConfigSpec{},
			}
			vsphereMachineConfig := firstMachineConfig(clusterSpec).DeepCopy()

			newClusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			newConfig := v1alpha1.WorkerNodeGroupConfiguration{Count: ptr.Int(1), MachineGroupRef: &v1alpha1.Ref{Name: "test-wn", Kind: "VSphereMachineConfig"}, Name: "md-2"}
			newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = append(newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations, newConfig)

			ipValidator := mocks.NewMockIPValidator(mockCtrl)
			ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup2MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereDatacenter, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil).AnyTimes()
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", cluster.Name), map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"}, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster)))

			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}

			err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
			if err != nil {
				t.Fatalf("failed to setup and validate: %v", err)
			}

			_, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, newClusterSpec)
			if err != nil {
				t.Fatalf("failed to generate cluster api spec contents: %v", err)
			}

			test.AssertContentToFile(t, string(md), tt.wantMDFile)
		})
	}
}

func TestProviderGenerateCAPISpecForUpgradeWorkerVersion(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantMDFile        string
	}{
		{
			testName:          "adding worker node group kubernetes version",
			clusterconfigFile: "cluster_main_worker_version.yaml",
			wantMDFile:        "testdata/expected_results_minimal_worker_version.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			setupContext(t)
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			vsphereDatacenter := &v1alpha1.VSphereDatacenterConfig{
				Spec: v1alpha1.VSphereDatacenterConfigSpec{},
			}
			vsphereMachineConfig := firstMachineConfig(clusterSpec).DeepCopy()

			newClusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			newConfig := v1alpha1.WorkerNodeGroupConfiguration{Count: ptr.Int(1), MachineGroupRef: &v1alpha1.Ref{Name: "test-wn", Kind: "VSphereMachineConfig"}, Name: "md-2"}
			newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = append(newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations, newConfig)

			ipValidator := mocks.NewMockIPValidator(mockCtrl)
			ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup2MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereDatacenter, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil).AnyTimes()
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[1].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil).AnyTimes()
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", cluster.Name), map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"}, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster)))

			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}

			err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
			if err != nil {
				t.Fatalf("failed to setup and validate: %v", err)
			}

			_, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, newClusterSpec)
			if err != nil {
				t.Fatalf("failed to generate cluster api spec contents: %v", err)
			}

			test.AssertContentToFile(t, string(md), tt.wantMDFile)
		})
	}
}

func TestProviderGenerateCAPISpecForUpgradeUpdateMachineTemplateExternalEtcd(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantCPFile        string
		wantMDFile        string
	}{
		{
			testName:          "main",
			clusterconfigFile: testClusterConfigMainFilename,
			wantCPFile:        "testdata/expected_results_main_cp.yaml",
			wantMDFile:        "testdata/expected_results_main_md.yaml",
		},
		{
			testName:          "main_with_taints",
			clusterconfigFile: "cluster_main_with_taints.yaml",
			wantCPFile:        "testdata/expected_results_main_with_taints_cp.yaml",
			wantMDFile:        "testdata/expected_results_main_with_taints_md.yaml",
		},
		{
			testName:          "main with node labels",
			clusterconfigFile: "cluster_main_with_node_labels.yaml",
			wantCPFile:        "testdata/expected_results_main_cp.yaml",
			wantMDFile:        "testdata/expected_results_main_node_labels_md.yaml",
		},
		{
			testName:          "main with cp node labels",
			clusterconfigFile: "cluster_main_with_cp_node_labels.yaml",
			wantCPFile:        "testdata/expected_results_main_node_labels_cp.yaml",
			wantMDFile:        "testdata/expected_results_main_md.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			setupContext(t)
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			vsphereDatacenter := &v1alpha1.VSphereDatacenterConfig{
				Spec: v1alpha1.VSphereDatacenterConfigSpec{},
			}
			vsphereMachineConfig := firstMachineConfig(clusterSpec).DeepCopy()

			ipValidator := mocks.NewMockIPValidator(mockCtrl)
			ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereDatacenter, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", cluster.Name), map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"}, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster)))
			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}

			err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
			if err != nil {
				t.Fatalf("failed to setup and validate: %v", err)
			}

			cp, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, clusterSpec)
			if err != nil {
				t.Fatalf("failed to generate cluster api spec contents: %v", err)
			}

			test.AssertContentToFile(t, string(cp), tt.wantCPFile)
			test.AssertContentToFile(t, string(md), tt.wantMDFile)
		})
	}
}

func TestProviderGenerateCAPISpecForUpgradeNotUpdateMachineTemplate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	bootstrapCluster := &types.Cluster{
		Name: "bootstrap-test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	oldCP := &controlplanev1.KubeadmControlPlane{
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: v1.ObjectReference{
					Name: "test-control-plane-template-original",
				},
			},
		},
	}
	oldMD := &clusterv1.MachineDeployment{
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					InfrastructureRef: v1.ObjectReference{
						Name: "test-md-0-original",
					},
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							Name: "test-md-0-template-original",
						},
					},
				},
			},
		},
	}
	etcdadmCluster := &etcdv1.EtcdadmCluster{
		Spec: etcdv1.EtcdadmClusterSpec{
			InfrastructureTemplate: v1.ObjectReference{
				Name: "test-etcd-template-original",
			},
		},
	}

	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	machineDeploymentName := fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Name)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(datacenterConfig, nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, controlPlaneMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName], nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, workerNodeMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName], nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, etcdMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereMachineConfigs[etcdMachineConfigName], nil)
	kubectl.EXPECT().GetKubeadmControlPlane(ctx, cluster, clusterSpec.Cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(oldCP, nil)
	kubectl.EXPECT().GetMachineDeployment(ctx, machineDeploymentName, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(oldMD, nil).Times(2)
	kubectl.EXPECT().GetEtcdadmCluster(ctx, cluster, clusterSpec.Cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(etcdadmCluster, nil)
	cp, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, clusterSpec.DeepCopy())
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_main_no_machinetemplate_update_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_no_machinetemplate_update_md.yaml")
}

func TestProviderGenerateCAPISpecForCreate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_main_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_md.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithControlPlaneTags(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	datacenterConfig := clusterSpec.VSphereDatacenter
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_main_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_md.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithMultipleWorkerNodeGroups(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, "cluster_main_multiple_worker_node_groups.yaml")
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	datacenterConfig := givenDatacenterConfig(t, "cluster_main_multiple_worker_node_groups.yaml")
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	_, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_multiple_worker_node_groups.yaml")
}

func TestProviderGenerateCAPISpecForUpgradeUpdateMachineGroupRefName(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	bootstrapCluster := &types.Cluster{
		Name: "bootstrap-test",
	}
	clusterSpec := givenClusterSpec(t, "cluster_main.yaml")
	vsphereDatacenter := &v1alpha1.VSphereDatacenterConfig{
		Spec: v1alpha1.VSphereDatacenterConfigSpec{},
	}
	vsphereMachineConfig := firstMachineConfig(clusterSpec).DeepCopy()
	wnMachineConfig := getMachineConfig(clusterSpec, "test-wn")

	newClusterSpec := clusterSpec.DeepCopy()
	newMachineConfigName := "new-test-wn"
	newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = newMachineConfigName
	newWorkerMachineConfig := wnMachineConfig.DeepCopy()
	newWorkerMachineConfig.Name = newMachineConfigName
	newClusterSpec.VSphereMachineConfigs[newMachineConfigName] = newWorkerMachineConfig

	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereDatacenter, nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(wnMachineConfig, nil).AnyTimes()
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
	kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", cluster.Name), map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"}, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster)))

	datacenterConfig := givenDatacenterConfig(t, "cluster_main.yaml")
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	provider.templateBuilder.now = test.NewFakeNow
	_, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, newClusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_md_update_machine_template.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithBottlerocketAndExternalEtcd(t *testing.T) {
	clusterSpecManifest := "cluster_bottlerocket_external_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	vscb, _ := newMockVSphereClientBuilder(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	v := NewValidator(govc, vscb)
	govc.osTag = bottlerocketOSTag
	provider := newProvider(
		t,
		datacenterConfig,
		clusterSpec.Cluster,
		govc,
		kubectl,
		v,
		ipValidator,
	)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_bottlerocket_external_etcd_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_bottlerocket_external_etcd_md.yaml")
}

func TestProviderGenerateDeploymentFileForBottleRocketWithMirrorConfig(t *testing.T) {
	clusterSpecManifest := "cluster_bottlerocket_mirror_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	vscb, _ := newMockVSphereClientBuilder(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	v := NewValidator(govc, vscb)
	govc.osTag = bottlerocketOSTag
	provider := newProvider(
		t,
		datacenterConfig,
		clusterSpec.Cluster,
		govc,
		kubectl,
		v,
		ipValidator,
	)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_bottlerocket_mirror_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_bottlerocket_mirror_config_md.yaml")
}

func TestProviderGenerateDeploymentFileForBottleRocketWithMirrorAndCertConfig(t *testing.T) {
	clusterSpecManifest := "cluster_bottlerocket_mirror_with_cert_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	govc.osTag = bottlerocketOSTag
	vscb, _ := newMockVSphereClientBuilder(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	v := NewValidator(govc, vscb)
	provider := newProvider(
		t,
		datacenterConfig,
		clusterSpec.Cluster,
		govc,
		kubectl,
		v,
		ipValidator,
	)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_bottlerocket_mirror_config_with_cert_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_bottlerocket_mirror_config_with_cert_md.yaml")
}

func TestProviderGenerateDeploymentFileForBottleRocketWithMirrorAuthConfig(t *testing.T) {
	clusterSpecManifest := "cluster_bottlerocket_mirror_with_auth_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	t.Setenv("REGISTRY_USERNAME", "username")
	t.Setenv("REGISTRY_PASSWORD", "password")
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	govc.osTag = bottlerocketOSTag
	vscb, _ := newMockVSphereClientBuilder(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	v := NewValidator(govc, vscb)
	provider := newProvider(
		t,
		datacenterConfig,
		clusterSpec.Cluster,
		govc,
		kubectl,
		v,
		ipValidator,
	)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_bottlerocket_mirror_config_with_auth_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_bottlerocket_mirror_config_with_auth_md.yaml")
}

func TestProviderGenerateDeploymentFileWithMirrorConfig(t *testing.T) {
	clusterSpecManifest := "cluster_mirror_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	govc.osTag = ubuntuOSTag
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_mirror_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_mirror_config_md.yaml")
}

func TestProviderGenerateDeploymentFileWithMirrorAndCertConfig(t *testing.T) {
	clusterSpecManifest := "cluster_mirror_with_cert_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	govc.osTag = ubuntuOSTag
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_mirror_config_with_cert_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_mirror_config_with_cert_md.yaml")
}

func TestProviderGenerateDeploymentFileWithMirrorAuth(t *testing.T) {
	clusterSpecManifest := "cluster_mirror_with_auth_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	if err := os.Setenv("REGISTRY_USERNAME", "username"); err != nil {
		t.Fatalf(err.Error())
	}
	if err := os.Setenv("REGISTRY_PASSWORD", "password"); err != nil {
		t.Fatalf(err.Error())
	}
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	govc.osTag = ubuntuOSTag
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_mirror_with_auth_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_mirror_with_auth_config_md.yaml")
}

func TestUpdateKubeConfig(t *testing.T) {
	provider := givenProvider(t)
	content := []byte{}

	err := provider.UpdateKubeConfig(&content, "clusterName")
	if err != nil {
		t.Fatalf("failed UpdateKubeConfig: %v", err)
	}
}

func TestBootstrapClusterOpts(t *testing.T) {
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	bootstrapClusterOps, err := provider.BootstrapClusterOpts(clusterSpec)
	if err != nil {
		t.Fatalf("failed BootstrapClusterOpts: %v", err)
	}
	if bootstrapClusterOps == nil {
		t.Fatalf("expected BootstrapClusterOpts")
	}
}

func TestName(t *testing.T) {
	provider := givenProvider(t)

	if provider.Name() != expectedVSphereName {
		t.Fatalf("unexpected Name %s!=%s", provider.Name(), expectedVSphereName)
	}
}

func TestSetupAndValidateCreateCluster(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	provider.ipValidator = ipValidator

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func thenErrorPrefixExpected(t *testing.T, expected string, err error) {
	if err == nil {
		t.Fatalf("Expected=<%s> actual=<nil>", expected)
	}
	actual := err.Error()
	if !strings.HasPrefix(actual, expected) {
		t.Fatalf("Expected=<%s...> actual=<%s...>", expected, actual)
	}
}

func thenErrorExpected(t *testing.T, expected string, err error) {
	if err == nil {
		t.Fatalf("Expected=<%s> actual=<nil>", expected)
	}
	actual := err.Error()
	if expected != actual {
		t.Fatalf("Expected=<%s> actual=<%s>", expected, actual)
	}
}

func TestSetupAndValidateCreateClusterNoUsername(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	provider := givenProvider(t)
	setupContext(t)
	os.Unsetenv(config.EksavSphereUsernameKey)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: EKSA_VSPHERE_USERNAME is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoPassword(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	provider := givenProvider(t)
	setupContext(t)
	os.Unsetenv(config.EksavSpherePasswordKey)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: EKSA_VSPHERE_PASSWORD is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterMissingPrivError(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	setupContext(t)
	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	provider.ipValidator = ipValidator
	vscb := mocks.NewMockVSphereClientBuilder(mockCtrl)
	vscb.EXPECT().Build(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), clusterSpec.VSphereDatacenter.Spec.Datacenter).Return(nil, fmt.Errorf("error"))
	provider.validator.vSphereClientBuilder = vscb

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "validating vsphere user privileges: error", err)
}

func TestSetupAndValidateUpgradeClusterMissingPrivError(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	cluster := &types.Cluster{}
	provider := givenProvider(t)
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	setupContext(t)

	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil).Times(2)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, gomock.Any(), cluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace()).Times(5)

	vscb := mocks.NewMockVSphereClientBuilder(mockCtrl)
	vscb.EXPECT().Build(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), clusterSpec.VSphereDatacenter.Spec.Datacenter).Return(nil, fmt.Errorf("error"))
	provider.validator.vSphereClientBuilder = vscb

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)

	thenErrorExpected(t, "validating vsphere user privileges: error", err)
}

func TestSetupAndValidateCreateCPUpgradeRolloutStrategy(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigWithCPUpgradeStrategy)
	provider := givenProvider(t)
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: Upgrade rollout strategy customization is not supported for vSphere provider", err)
}

func TestSetupAndValidateCreateMDUpgradeRolloutStrategy(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigWithMDUpgradeStrategy)
	provider := givenProvider(t)
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: Upgrade rollout strategy customization is not supported for vSphere provider", err)
}

func TestSetupAndValidateUpgradeCPUpgradeRolloutStrategy(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigWithCPUpgradeStrategy)
	cluster := &types.Cluster{}
	provider := givenProvider(t)
	setupContext(t)

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: Upgrade rollout strategy customization is not supported for vSphere provider", err)
}

func TestSetupAndValidateUpgradeMDUpgradeRolloutStrategy(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigWithMDUpgradeStrategy)
	cluster := &types.Cluster{}
	provider := givenProvider(t)
	setupContext(t)

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: Upgrade rollout strategy customization is not supported for vSphere provider", err)
}

func TestSetupAndValidateDeleteCPUpgradeRolloutStrategy(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigWithCPUpgradeStrategy)
	provider := givenProvider(t)
	tt := newProviderTest(t)
	setupContext(t)

	err := provider.SetupAndValidateDeleteCluster(ctx, tt.managementCluster, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: Upgrade rollout strategy customization is not supported for vSphere provider", err)
}

func TestSetupAndValidateDeleteMDUpgradeRolloutStrategy(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigWithMDUpgradeStrategy)
	provider := givenProvider(t)
	tt := newProviderTest(t)
	setupContext(t)

	err := provider.SetupAndValidateDeleteCluster(ctx, tt.managementCluster, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: Upgrade rollout strategy customization is not supported for vSphere provider", err)
}

func TestSetupAndValidateCreateWorkloadClusterSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	setupContext(t)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator

	clusterSpec.Cluster.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
	}
	for _, config := range clusterSpec.VSphereMachineConfigs {
		kubectl.EXPECT().SearchVsphereMachineConfig(ctx, config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{}, nil)
	}
	kubectl.EXPECT().SearchVsphereDatacenterConfig(ctx, datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return([]*v1alpha1.VSphereDatacenterConfig{}, nil)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

func TestSetupAndValidateCreateWorkloadClusterFailsIfMachineExists(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator

	clusterSpec.Cluster.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
	}

	idx := 0
	var existingMachine string
	for _, config := range clusterSpec.VSphereMachineConfigs {
		if idx == 0 {
			kubectl.EXPECT().SearchVsphereMachineConfig(ctx, config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{config}, nil)
			existingMachine = config.Name
		} else {
			kubectl.EXPECT().SearchVsphereMachineConfig(ctx, config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{}, nil).MaxTimes(1)
		}
		idx++
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, fmt.Sprintf("VSphereMachineConfig %s already exists", existingMachine), err)
}

func TestSetupAndValidateSelfManagedClusterSkipMachineNameValidateSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator

	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
	}

	kubectl.EXPECT().SearchVsphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

func TestSetupAndValidateCreateWorkloadClusterFailsIfDatacenterExists(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	setupContext(t)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator

	clusterSpec.Cluster.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
	}

	for _, config := range clusterSpec.VSphereMachineConfigs {
		kubectl.EXPECT().SearchVsphereMachineConfig(ctx, config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{}, nil)
	}
	kubectl.EXPECT().SearchVsphereDatacenterConfig(ctx, datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return([]*v1alpha1.VSphereDatacenterConfig{datacenterConfig}, nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, fmt.Sprintf("VSphereDatacenter %s already exists", datacenterConfig.Name), err)
}

func TestSetupAndValidateSelfManagedClusterSkipDatacenterNameValidateSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator

	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
	}

	kubectl.EXPECT().SearchVsphereMachineConfig(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	kubectl.EXPECT().SearchVsphereDatacenterConfig(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

func TestSetupAndValidateDeleteCluster(t *testing.T) {
	tt := newProviderTest(t)

	tt.Expect(
		tt.provider.SetupAndValidateDeleteCluster(tt.ctx, tt.managementCluster, tt.clusterSpec),
	).To(Succeed())
}

func TestSetupAndValidateDeleteClusterNoPassword(t *testing.T) {
	tt := newProviderTest(t)
	os.Unsetenv(config.EksavSpherePasswordKey)

	err := tt.provider.SetupAndValidateDeleteCluster(tt.ctx, tt.managementCluster, tt.clusterSpec)
	thenErrorExpected(t, "failed setup and validations: EKSA_VSPHERE_PASSWORD is not set or is empty", err)
}

func TestSetupAndValidateUpgradeCluster(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	cluster := &types.Cluster{}
	provider := givenProvider(t)
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	setupContext(t)

	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil).Times(3)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, gomock.Any(), cluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace()).Times(5)
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterNoUsername(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	provider := givenProvider(t)
	setupContext(t)
	os.Unsetenv(config.EksavSphereUsernameKey)

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: EKSA_VSPHERE_USERNAME is not set or is empty", err)
}

func TestSetupAndValidateUpgradeClusterNoPassword(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	provider := givenProvider(t)
	setupContext(t)
	os.Unsetenv(config.EksavSpherePasswordKey)

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: EKSA_VSPHERE_PASSWORD is not set or is empty", err)
}

func TestSetupAndValidateUpgradeClusterDatastoreUsageError(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	cluster := &types.Cluster{}
	provider := givenProvider(t)
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	setupContext(t)

	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, gomock.Any(), cluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace()).Return(nil, fmt.Errorf("error"))

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	thenErrorExpected(t, "validating vsphere machine configs datastore usage: calculating datastore usage: error", err)
}

func TestSetupAndValidateUpgradeClusterCPSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil).Times(3)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, gomock.Any(), cluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace()).Times(5)
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterWorkerSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil).Times(3)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, gomock.Any(), cluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace()).Times(5)

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterEtcdSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil).Times(3)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, gomock.Any(), cluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace()).Times(5)

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterSameMachineConfigforCPandEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = etcdMachineConfigName
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil).Times(3)
	for _, mc := range clusterSpec.VSphereMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, gomock.Any(), cluster.KubeconfigFile, clusterSpec.Cluster.GetNamespace()).Return(mc, nil).AnyTimes()
	}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestVersion(t *testing.T) {
	vSphereProviderVersion := "v0.7.10"
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	clusterSpec.VersionsBundles["1.19"].VSphere.Version = vSphereProviderVersion
	setupContext(t)

	result := provider.Version(clusterSpec)
	if result != vSphereProviderVersion {
		t.Fatalf("Unexpected version expected <%s> actual=<%s>", vSphereProviderVersion, result)
	}
}

func TestProviderBootstrapSetup(t *testing.T) {
	ctx := context.Background()
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterConfig, kubectl, ipValidator)
	cluster := types.Cluster{
		Name:           "test",
		KubeconfigFile: "",
	}
	values := map[string]string{
		"clusterName":               clusterConfig.Name,
		"vspherePassword":           expectedVSphereUsername,
		"vsphereUsername":           expectedVSpherePassword,
		"eksaCloudProviderUsername": expectedVSphereUsername,
		"eksaCloudProviderPassword": expectedVSpherePassword,
		"vsphereServer":             datacenterConfig.Spec.Server,
		"vsphereDatacenter":         datacenterConfig.Spec.Datacenter,
		"vsphereNetwork":            datacenterConfig.Spec.Network,
		"eksaLicense":               "",
	}

	setupContext(t)

	tpl, err := template.New("test").Funcs(sprig.TxtFuncMap()).Parse(defaultSecretObject)
	if err != nil {
		t.Fatalf("template create error: %v", err)
	}
	err = tpl.Execute(&bytes.Buffer{}, values)
	if err != nil {
		t.Fatalf("template execute error: %v", err)
	}

	err = provider.PostBootstrapSetup(ctx, clusterConfig, &cluster)
	if err != nil {
		t.Fatalf("BootstrapSetup error %v", err)
	}
}

func TestProviderUpdateSecretSuccess(t *testing.T) {
	ctx := context.Background()
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterConfig, kubectl, ipValidator)
	cluster := types.Cluster{
		Name:           "test",
		KubeconfigFile: "",
	}
	values := map[string]string{
		"clusterName":               clusterConfig.Name,
		"vspherePassword":           expectedVSphereUsername,
		"vsphereUsername":           expectedVSpherePassword,
		"eksaCloudProviderUsername": expectedVSphereUsername,
		"eksaCloudProviderPassword": expectedVSpherePassword,
		"eksaLicense":               "",
		"eksaSystemNamespace":       constants.EksaSystemNamespace,
	}

	setupContext(t)

	kubectl.EXPECT().ApplyKubeSpecFromBytes(ctx, gomock.Any(), gomock.Any())

	template, err := template.New("test").Funcs(sprig.TxtFuncMap()).Parse(defaultSecretObject)
	if err != nil {
		t.Fatalf("template create error: %v", err)
	}
	err = template.Execute(&bytes.Buffer{}, values)
	if err != nil {
		t.Fatalf("template execute error: %v", err)
	}

	err = provider.UpdateSecrets(ctx, &cluster, nil)
	if err != nil {
		t.Fatalf("UpdateSecrets error %v", err)
	}
}

func TestSetupAndValidateCreateClusterNoServer(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.VSphereDatacenter.Spec.Server = ""
	provider := givenProvider(t)
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereDatacenterConfig server is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterInsecure(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.VSphereDatacenter.Spec.Insecure = true
	provider := givenProvider(t)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if clusterSpec.VSphereDatacenter.Spec.Thumbprint != "" {
		t.Fatalf("Expected=<> actual=<%s>", clusterSpec.VSphereDatacenter.Spec.Thumbprint)
	}
}

func TestSetupAndValidateCreateClusterNoDatacenter(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.VSphereDatacenter.Spec.Datacenter = ""
	provider := givenProvider(t)
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereDatacenterConfig datacenter is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoNetwork(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.VSphereDatacenter.Spec.Network = ""
	provider := givenProvider(t)
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereDatacenterConfig VM network is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNotControlPlaneVMsMemoryMiB(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.MemoryMiB = 0
	provider := givenProvider(t)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.MemoryMiB != 8192 {
		t.Fatalf("Expected=<8192> actual=<%d>", clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.MemoryMiB)
	}
}

func TestSetupAndValidateCreateClusterNotControlPlaneVMsNumCPUs(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.NumCPUs = 0
	provider := givenProvider(t)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.NumCPUs != 2 {
		t.Fatalf("Expected=<2> actual=<%d>", clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.NumCPUs)
	}
}

func TestSetupAndValidateCreateClusterNotWorkloadVMsMemoryMiB(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.MemoryMiB = 0
	provider := givenProvider(t)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.MemoryMiB != 8192 {
		t.Fatalf("Expected=<8192> actual=<%d>", clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.MemoryMiB)
	}
}

func TestSetupAndValidateCreateClusterNotWorkloadVMsNumCPUs(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.NumCPUs = 0
	provider := givenProvider(t)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.NumCPUs != 2 {
		t.Fatalf("Expected=<2> actual=<%d>", clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.NumCPUs)
	}
}

func TestSetupAndValidateCreateClusterNotEtcdVMsMemoryMiB(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.MemoryMiB = 0
	provider := givenProvider(t)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.MemoryMiB != 8192 {
		t.Fatalf("Expected=<8192> actual=<%d>", clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.MemoryMiB)
	}
}

func TestSetupAndValidateCreateClusterNotEtcdVMsNumCPUs(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.NumCPUs = 0
	provider := givenProvider(t)
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.NumCPUs != 2 {
		t.Fatalf("Expected=<2> actual=<%d>", clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.NumCPUs)
	}
}

func TestSetupAndValidateCreateClusterBogusIp(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "bogus"
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "cluster controlPlaneConfiguration.Endpoint.Host is invalid: bogus", err)
}

func TestSetupAndValidateCreateClusterUsedIp(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "0.0.0.0"
	setupContext(t)

	ipInUseError := "cluster controlPlaneConfiguration.Endpoint.Host <0.0.0.0> is already in use, please provide a unique IP"

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(fmt.Errorf(ipInUseError))

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, ipInUseError, err)
}

func TestSetupAndValidateCreateClusterNoCloneModeDefaultToLinkedClone(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	govc := NewDummyProviderGovcClient()
	provider := newProviderWithGovc(t,
		clusterSpec.VSphereDatacenter,
		clusterSpec.Cluster,
		govc,
	)
	provider.providerGovcClient = govc
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.CloneMode = ""
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.CloneMode = ""
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.CloneMode = ""
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	assert.NoError(t, err, "No error expected for provider.SetupAndValidateCreateCluster()")

	for _, m := range clusterSpec.VSphereMachineConfigs {
		assert.Equal(t, m.Spec.CloneMode, v1alpha1.LinkedClone)
	}
}

func TestSetupAndValidateCreateClusterNoCloneModeDefaultToFullClone(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	govc := NewDummyProviderGovcClient()
	provider := newProviderWithGovc(t,
		clusterSpec.VSphereDatacenter,
		clusterSpec.Cluster,
		govc,
	)
	provider.providerGovcClient = govc
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.CloneMode = ""
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.DiskGiB = 100
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.CloneMode = ""
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.DiskGiB = 100
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.CloneMode = ""
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.DiskGiB = 100
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	assert.NoError(t, err, "No error expected for provider.SetupAndValidateCreateCluster()")

	for _, m := range clusterSpec.VSphereMachineConfigs {
		assert.Equal(t, m.Spec.CloneMode, v1alpha1.FullClone)
	}
}

func TestSetupAndValidateCreateClusterFullCloneDiskGiBLessThan20TemplateDiskSize25(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	govc := NewDummyProviderGovcClient()
	provider := newProviderWithGovc(t,
		clusterSpec.VSphereDatacenter,
		clusterSpec.Cluster,
		govc,
	)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.CloneMode = v1alpha1.FullClone
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.DiskGiB = 10
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.CloneMode = v1alpha1.FullClone
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.DiskGiB = 10
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.CloneMode = v1alpha1.FullClone
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.DiskGiB = 10
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	assert.NoError(t, err, "No error expected for provider.SetupAndValidateCreateCluster()")

	for _, m := range clusterSpec.VSphereMachineConfigs {
		assert.Equalf(t, m.Spec.DiskGiB, 25, "DiskGiB mismatch for VSphereMachineConfig %s", m.Name)
	}
}

func TestSetupAndValidateCreateClusterFullCloneDiskGiBLessThan20TemplateDiskSize20(t *testing.T) {
	setupContext(t)
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(ctrl)
	govc := mocks.NewMockProviderGovcClient(ctrl)
	vscb, _ := newMockVSphereClientBuilder(ctrl)
	ipValidator := mocks.NewMockIPValidator(ctrl)
	spec := givenClusterSpec(t, testClusterConfigMain121CPOnlyFilename)

	tt := &providerTest{
		t:     t,
		WithT: NewWithT(t),
		ctx:   context.Background(),
		managementCluster: &types.Cluster{
			Name:           "m-cluster",
			KubeconfigFile: "kubeconfig-m.kubeconfig",
		},
		workloadCluster: &types.Cluster{
			Name:           "test",
			KubeconfigFile: "kubeconfig-w.kubeconfig",
		},
		cluster:          spec.Cluster,
		clusterSpec:      spec,
		datacenterConfig: spec.VSphereDatacenter,
		machineConfigs:   spec.VSphereMachineConfigs,
		kubectl:          kubectl,
		govc:             govc,
		clientBuilder:    vscb,
		ipValidator:      ipValidator,
	}
	tt.buildNewProvider()

	controlPlaneMachineConfigName := tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	tt.clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.CloneMode = v1alpha1.FullClone
	tt.clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.DiskGiB = 10

	template := tt.clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Template

	tt.setExpectationForSetup()
	tt.setExpectationForVCenterValidation()
	tt.setExpectationsForMachineConfigsVCenterValidation()

	tt.govc.EXPECT().GetVMDiskSizeInGB(tt.ctx, template, tt.clusterSpec.VSphereDatacenter.Spec.Datacenter)
	tt.govc.EXPECT().TemplateHasSnapshot(tt.ctx, template).Return(false, nil)
	tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, template).Return(template, nil).Times(2) // One for defaults and another time for template validation
	tt.govc.EXPECT().GetTags(tt.ctx, template).Return([]string{"eksdRelease:kubernetes-1-21-eks-4", "os:ubuntu"}, nil)
	tt.govc.EXPECT().ListTags(tt.ctx)
	tt.govc.EXPECT().GetWorkloadAvailableSpace(tt.ctx, tt.clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Datastore).Return(100.0, nil)
	tt.ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(tt.cluster)

	resourcePoolResponse := map[string]int{
		"Memory_Available": -1,
	}
	tt.govc.EXPECT().GetResourcePoolInfo(tt.ctx, tt.clusterSpec.VSphereDatacenter.Spec.Datacenter, tt.clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.ResourcePool).Return(resourcePoolResponse, nil)
	err := tt.provider.SetupAndValidateCreateCluster(context.Background(), tt.clusterSpec)

	assert.NoError(t, err, "No error expected for provider.SetupAndValidateCreateCluster()")

	for _, m := range tt.clusterSpec.VSphereMachineConfigs {
		assert.Equalf(t, m.Spec.DiskGiB, 20, "DiskGiB mismatch for VSphereMachineConfig %s", m.Name)
	}
}

func TestSetupAndValidateCreateClusterLinkedCloneErrorDiskSize(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	govc := NewDummyProviderGovcClient()
	provider := newProviderWithGovc(t,
		clusterSpec.VSphereDatacenter,
		clusterSpec.Cluster,
		govc,
	)
	provider.providerGovcClient = govc
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.DiskGiB = 100
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	assert.ErrorContains(t, err, fmt.Sprintf(
		"diskGiB cannot be customized for VSphereMachineConfig '%s' when using 'linkedClone'; change the cloneMode to 'fullClone' or the diskGiB to match the template's (%s) disk size of 25 GiB",
		controlPlaneMachineConfigName, clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Template,
	))
}

func TestSetupAndValidateCreateClusterLinkedCloneErrorNoSnapshots(t *testing.T) {
	tt := newProviderTest(t)
	controlPlaneMachineConfigName := tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name

	tt.setExpectationForSetup()
	tt.setExpectationForVCenterValidation()
	tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, tt.clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Template).Return(tt.clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Template, nil)
	tt.govc.EXPECT().GetVMDiskSizeInGB(tt.ctx, tt.clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Template, tt.clusterSpec.VSphereDatacenter.Spec.Datacenter)
	tt.govc.EXPECT().TemplateHasSnapshot(tt.ctx, tt.clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Template).Return(false, nil)

	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	assert.Regexp(t,
		"cannot use 'linkedClone' for VSphereMachineConfig '.*' because its template (.*) has no snapshots; create snapshots or change the cloneMode to 'fullClone",
		err.Error(),
	)
}

func TestSetupAndValidateCreateClusterInvalidCloneMode(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	govc := NewDummyProviderGovcClient()
	provider := newProviderWithGovc(t,
		clusterSpec.VSphereDatacenter,
		clusterSpec.Cluster,
		govc,
	)
	provider.providerGovcClient = govc
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	invalidClone := "invalidClone"
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.CloneMode = v1alpha1.CloneMode(invalidClone)
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	assert.ErrorContains(t, err,
		fmt.Sprintf(
			"cloneMode %s is not supported for VSphereMachineConfig %s. Supported clone modes: [%s, %s]",
			invalidClone,
			controlPlaneMachineConfigName,
			v1alpha1.LinkedClone,
			v1alpha1.FullClone,
		),
	)
}

func TestSetupAndValidateCreateClusterDatastoreUsageError(t *testing.T) {
	tt := newProviderTest(t)

	tt.setExpectationForSetup()
	tt.setExpectationForVCenterValidation()
	tt.setExpectationsForDefaultDiskAndCloneModeGovcCalls()
	tt.setExpectationsForMachineConfigsVCenterValidation()

	cpMachineConfig := tt.machineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	for _, mc := range tt.machineConfigs {
		tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, mc.Spec.Template).Return(mc.Spec.Template, nil).AnyTimes()
	}
	tt.govc.EXPECT().GetTags(tt.ctx, cpMachineConfig.Spec.Template).Return([]string{eksd119ReleaseTag, ubuntuOSTag}, nil)
	tt.govc.EXPECT().ListTags(tt.ctx)
	tt.govc.EXPECT().GetWorkloadAvailableSpace(tt.ctx, cpMachineConfig.Spec.Datastore).Return(0.0, fmt.Errorf("error"))

	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)

	thenErrorExpected(t, "validating vsphere machine configs datastore usage: getting datastore details: error", err)
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyCP(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for control plane machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyErrorGenerating(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	g := NewWithT(t)
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	ctrl := gomock.NewController(t)
	writer := mockswriter.NewMockFileWriter(ctrl)
	provider.writer = writer
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""

	writer.EXPECT().Write(
		test.OfType("string"), gomock.Any(), gomock.Not(gomock.Nil()),
	).Return("", errors.New("writing file"))

	setupContext(t)

	g.Expect(
		provider.SetupAndValidateCreateCluster(ctx, clusterSpec),
	).To(MatchError(ContainSubstring(
		"failed setup and validations: generating ssh key pair: writing private key: writing file",
	)))
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyWorker(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for worker node machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyEtcd(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey did not get generated for etcd machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyAllMachineConfigs(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""

	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for control plane machine")
	}
	if clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for worker node machine")
	}
	if clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey not generated for etcd machines")
	}
	if clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] != clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] {
		t.Fatalf("sshAuthorizedKey not the same for controlplane and worker machines")
	}
	if clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] != clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] {
		t.Fatalf("sshAuthorizedKey not the same for controlplane and etcd machines")
	}
}

func TestSetupAndValidateUsersNil(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users = nil
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.Users = nil
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Users = nil
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
}

func TestSetupAndValidateSshAuthorizedKeysNil(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
}

func TestSetupAndValidateCreateClusterCPMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "nonexistent"
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "cannot find VSphereMachineConfig nonexistent for control plane", err)
}

func TestSetupAndValidateCreateClusterWorkerMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = "nonexistent"
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "failed setting default values for vsphere machine configs: cannot find VSphereMachineConfig nonexistent for worker nodes", err)
}

func TestSetupAndValidateCreateClusterEtcdMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "nonexistent"
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "cannot find VSphereMachineConfig nonexistent for etcd machines", err)
}

func TestSetupAndValidateCreateClusterOsFamilyDifferent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.OSFamily = "bottlerocket"
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].Name = "ec2-user"
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "all VSphereMachineConfigs must have the same osFamily specified", err)
}

func TestSetupAndValidateCreateClusterOsFamilyDifferentForEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.OSFamily = "bottlerocket"
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Users[0].Name = "ec2-user"
	setupContext(t)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "all VSphereMachineConfigs must have the same osFamily specified", err)
}

func TestSetupAndValidateCreateClusterOsFamilyEmpty(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	govc := NewDummyProviderGovcClient()
	govc.osTag = bottlerocketOSTag
	provider := newProviderWithGovc(t,
		clusterSpec.VSphereDatacenter,
		clusterSpec.Cluster,
		govc,
	)
	provider.providerGovcClient = govc
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.OSFamily = ""
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Users[0].Name = ""
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.OSFamily = ""
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.Users[0].Name = ""
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.OSFamily = ""
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Users[0].Name = ""
	setupContext(t)

	mockCtrl := gomock.NewController(t)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.OSFamily != v1alpha1.Bottlerocket {
		t.Fatalf("got osFamily for control plane machine as %v, want %v", clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.OSFamily, v1alpha1.Bottlerocket)
	}
	if clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.OSFamily != v1alpha1.Bottlerocket {
		t.Fatalf("got osFamily for control plane machine as %v, want %v", clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.OSFamily, v1alpha1.Bottlerocket)
	}
	if clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.OSFamily != v1alpha1.Bottlerocket {
		t.Fatalf("got osFamily for etcd machine as %v, want %v", clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.OSFamily, v1alpha1.Bottlerocket)
	}
}

func TestSetupAndValidateCreateClusterTemplateDoesNotExist(t *testing.T) {
	tt := newProviderTest(t)

	tt.setExpectationForSetup()
	tt.setExpectationForVCenterValidation()
	for _, mc := range tt.machineConfigs {
		tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, mc.Spec.Template).Return("", nil).MaxTimes(1)
	}

	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)

	thenErrorExpected(t, "failed setting default values for vsphere machine configs: template <"+testTemplate+"> not found", err)
}

func TestSetupAndValidateCreateClusterErrorCheckingTemplate(t *testing.T) {
	tt := newProviderTest(t)
	errorMessage := "failed getting template"

	tt.setExpectationForSetup()
	tt.setExpectationForVCenterValidation()
	for _, mc := range tt.machineConfigs {
		tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, mc.Spec.Template).Return("", errors.New(errorMessage)).MaxTimes(1)
	}

	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)

	thenErrorExpected(t, "failed setting default values for vsphere machine configs: setting template full path: "+errorMessage, err)
}

func TestSetupAndValidateCreateClusterTemplateMissingTags(t *testing.T) {
	tt := newProviderTest(t)

	tt.setExpectationForSetup()
	tt.setExpectationsForDefaultDiskAndCloneModeGovcCalls()
	tt.setExpectationForVCenterValidation()
	tt.setExpectationsForMachineConfigsVCenterValidation()

	for _, mc := range tt.machineConfigs {
		tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, mc.Spec.Template).Return(mc.Spec.Template, nil)
	}
	controlPlaneMachineConfigName := tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	controlPlaneMachineConfig := tt.machineConfigs[controlPlaneMachineConfigName]

	tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, controlPlaneMachineConfig.Spec.Template).Return(controlPlaneMachineConfig.Spec.Template, nil)
	tt.govc.EXPECT().GetTags(tt.ctx, controlPlaneMachineConfig.Spec.Template).Return(nil, nil)

	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)

	thenErrorPrefixExpected(t, "template "+testTemplate+" is missing tag ", err)
}

func TestSetupAndValidateCreateClusterErrorGettingTags(t *testing.T) {
	tt := newProviderTest(t)
	errorMessage := "failed getting tags"

	controlPlaneMachineConfigName := tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	controlPlaneMachineConfig := tt.clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName]

	tt.setExpectationForSetup()
	tt.setExpectationsForDefaultDiskAndCloneModeGovcCalls()
	tt.setExpectationForVCenterValidation()
	tt.setExpectationsForMachineConfigsVCenterValidation()
	for _, mc := range tt.machineConfigs {
		tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, mc.Spec.Template).Return(mc.Spec.Template, nil)
	}
	tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, controlPlaneMachineConfig.Spec.Template).Return(controlPlaneMachineConfig.Spec.Template, nil)
	tt.govc.EXPECT().GetTags(tt.ctx, controlPlaneMachineConfig.Spec.Template).Return(nil, errors.New(errorMessage))

	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)

	thenErrorExpected(t, "validating template tags: failed getting tags", err)
}

func TestSetupAndValidateCreateClusterDefaultTemplateWorker(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.VersionsBundles["1.19"].EksD.Ova.Bottlerocket.URI = "https://amazonaws.com/artifacts/0.0.1/eks-distro/ova/1-19/1-19-4/bottlerocket-eks-a-0.0.1.build.38-amd64.ova"
	clusterSpec.VersionsBundles["1.19"].EksD.Ova.Bottlerocket.SHA256 = "63a8dce1683379cb8df7d15e9c5adf9462a2b9803a544dd79b16f19a4657967f"
	clusterSpec.VersionsBundles["1.19"].EksD.Ova.Bottlerocket.Arch = []string{"amd64"}
	clusterSpec.VersionsBundles["1.19"].EksD.Name = eksd119Release
	clusterSpec.VersionsBundles["1.19"].EksD.KubeVersion = "v1.19.8"
	clusterSpec.VersionsBundles["1.19"].KubeVersion = "1.19"
	clusterSpec.Cluster.Namespace = "test-namespace"
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Template = ""
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName].Spec.Template = ""
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Template = ""
	wantError := fmt.Errorf("failed setting default values for vsphere machine configs: can not import ova for osFamily: ubuntu, please use bottlerocket as osFamily for auto-importing or provide a valid template")

	setupContext(t)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err == nil || err.Error() != wantError.Error() {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = %v", err, wantError)
	}
}

func TestSetupAndValidateCreateClusterDefaultTemplateCP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.VersionsBundles["1.19"].EksD.Ova.Bottlerocket.URI = "https://amazonaws.com/artifacts/0.0.1/eks-distro/ova/1-19/1-19-4/bottlerocket-eks-a-0.0.1.build.38-amd64.ova"
	clusterSpec.VersionsBundles["1.19"].EksD.Ova.Bottlerocket.SHA256 = "63a8dce1683379cb8df7d15e9c5adf9462a2b9803a544dd79b16f19a4657967f"
	clusterSpec.VersionsBundles["1.19"].EksD.Ova.Bottlerocket.Arch = []string{"amd64"}
	clusterSpec.VersionsBundles["1.19"].EksD.Name = eksd119Release
	clusterSpec.VersionsBundles["1.19"].EksD.KubeVersion = "v1.19.8"
	clusterSpec.VersionsBundles["1.19"].KubeVersion = "1.19"
	clusterSpec.Cluster.Namespace = "test-namespace"
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.Template = ""
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	clusterSpec.VSphereMachineConfigs[etcdMachineConfigName].Spec.Template = ""
	wantError := fmt.Errorf("failed setting default values for vsphere machine configs: can not import ova for osFamily: ubuntu, please use bottlerocket as osFamily for auto-importing or provide a valid template")

	setupContext(t)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err == nil || err.Error() != wantError.Error() {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = %v", err, wantError)
	}
}

func TestGetInfrastructureBundleSuccess(t *testing.T) {
	tests := []struct {
		testName    string
		clusterSpec *cluster.Spec
	}{
		{
			testName: "correct Overrides layer",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].VSphere = releasev1alpha1.VSphereBundle{
					Version: "v0.7.8",
					ClusterAPIController: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api-provider-vsphere/release/manager:v0.7.8-35f54b0a7ff0f4f3cb0b8e30a0650acd0e55496a",
					},
					Manager: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/kubernetes/cloud-provider-vsphere/cpi/manager:v1.18.1-2093eaeda5a4567f0e516d652e0b25b1d7abc774",
					},
					KubeVip: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.2-2093eaeda5a4567f0e516d652e0b25b1d7abc774",
					},
					Metadata: releasev1alpha1.Manifest{
						URI: "Metadata.yaml",
					},
					Components: releasev1alpha1.Manifest{
						URI: "Components.yaml",
					},
					ClusterTemplate: releasev1alpha1.Manifest{
						URI: "ClusterTemplate.yaml",
					},
				}
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			p := givenProvider(t)

			infraBundle := p.GetInfrastructureBundle(tt.clusterSpec)
			if infraBundle == nil {
				t.Fatalf("provider.GetInfrastructureBundle() should have an infrastructure bundle")
			}
			assert.Equal(t, "infrastructure-vsphere/v0.7.8/", infraBundle.FolderName, "Incorrect folder name")
			assert.Equal(t, len(infraBundle.Manifests), 3, "Wrong number of files in the infrastructure bundle")
			bundle := tt.clusterSpec.RootVersionsBundle()
			wantManifests := []releasev1alpha1.Manifest{
				bundle.VSphere.Components,
				bundle.VSphere.Metadata,
				bundle.VSphere.ClusterTemplate,
			}
			assert.ElementsMatch(t, infraBundle.Manifests, wantManifests, "Incorrect manifests")
		})
	}
}

func TestGetDatacenterConfig(t *testing.T) {
	tt := newProviderTest(t)

	providerConfig := tt.provider.DatacenterConfig(tt.clusterSpec)
	tt.Expect(providerConfig).To(BeAssignableToTypeOf(&v1alpha1.VSphereDatacenterConfig{}))
	d := providerConfig.(*v1alpha1.VSphereDatacenterConfig)
	tt.Expect(d).To(Equal(tt.clusterSpec.VSphereDatacenter))
}

func TestValidateNewSpecSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()
	setupContext(t)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterVsphereSecret := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Data: map[string][]byte{
			"username": []byte("vsphere_username"),
			"password": []byte("vsphere_password"),
		},
	}

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereDatacenter, nil)
	for _, config := range clusterSpec.VSphereMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterSpec.Cluster.Namespace).Return(config, nil)
	}
	kubectl.EXPECT().GetSecretFromNamespace(gomock.Any(), gomock.Any(), CredentialsObjectName, gomock.Any()).Return(clusterVsphereSecret, nil)

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, newClusterSpec)
	assert.NoError(t, err, "No error should be returned when previous spec == new spec")
}

func TestValidateNewSpecMutableFields(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()
	setupContext(t)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	for _, config := range newClusterSpec.VSphereMachineConfigs {
		config.Spec.ResourcePool = "new-" + config.Spec.ResourcePool
		config.Spec.Folder = "new=" + config.Spec.Folder
	}

	clusterVsphereSecret := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Data: map[string][]byte{
			"username": []byte("vsphere_username"),
			"password": []byte("vsphere_password"),
		},
	}

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(clusterSpec.VSphereDatacenter, nil)
	for _, config := range clusterSpec.VSphereMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterSpec.Cluster.Namespace).Return(config, nil)
	}
	kubectl.EXPECT().GetSecretFromNamespace(gomock.Any(), gomock.Any(), CredentialsObjectName, gomock.Any()).Return(clusterVsphereSecret, nil)

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, newClusterSpec)
	assert.NoError(t, err, "No error should be returned when modifying mutable fields")
}

func TestValidateNewSpecDatacenterImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.VSphereDatacenter.Spec.Datacenter = "new-" + clusterSpec.VSphereDatacenter.Spec.Datacenter

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereDatacenter, nil)
	for _, config := range clusterSpec.VSphereMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterSpec.Cluster.Namespace).Return(config, nil)
	}

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, newClusterSpec)
	assert.Error(t, err, "Datacenter should be immutable")
}

func TestValidateNewSpecMachineConfigNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newClusterSpec.VSphereDatacenter.Spec.Datacenter = "new-" + newClusterSpec.VSphereDatacenter.Spec.Datacenter
	newClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "missing-machine-group"
	newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = "missing-machine-group"
	newClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "missing-machine-group"

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereDatacenter, nil)

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, newClusterSpec)
	assert.Errorf(t, err, "can't find machine config missing-machine-group in vsphere provider machine configs")
}

func TestValidateNewSpecServerImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newClusterSpec.VSphereDatacenter.Spec.Server = "new-" + newClusterSpec.VSphereDatacenter.Spec.Server

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereDatacenter, nil)
	for _, config := range clusterSpec.VSphereMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterSpec.Cluster.Namespace).Return(config, nil)
	}

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, newClusterSpec)
	assert.Error(t, err, "Server should be immutable")
}

func TestValidateNewSpecStoragePolicyNameImmutableControlPlane(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	controlPlaneMachineConfigName := newClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	controlPlaneMachineConfig := newClusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName]
	controlPlaneMachineConfig.Spec.StoragePolicyName = "new-" + controlPlaneMachineConfig.Spec.StoragePolicyName

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereDatacenter, nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName], nil).AnyTimes()

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, newClusterSpec)
	assert.ErrorContains(t, err, "spec.storagePolicyName is immutable", "StoragePolicyName should be immutable")
}

func TestValidateNewSpecStoragePolicyNameImmutableWorker(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	workerMachineConfigName := newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	workerMachineConfig := newClusterSpec.VSphereMachineConfigs[workerMachineConfigName]
	workerMachineConfig.Spec.StoragePolicyName = "new-" + workerMachineConfig.Spec.StoragePolicyName

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereDatacenter, nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereMachineConfigs[workerMachineConfigName], nil).AnyTimes()

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, newClusterSpec)
	assert.ErrorContains(t, err, "spec.storagePolicyName is immutable", "StoragePolicyName should be immutable")
}

func TestValidateNewSpecOSFamilyImmutableControlPlane(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	newClusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName].Spec.OSFamily = "bottlerocket"

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereDatacenter, nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName], nil).AnyTimes()

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, newClusterSpec)
	assert.ErrorContains(t, err, "spec.osFamily is immutable", "OSFamily should be immutable")
}

func TestValidateNewSpecOSFamilyImmutableWorker(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	workerMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	newClusterSpec.VSphereMachineConfigs[workerMachineConfigName].Spec.OSFamily = "bottlerocket"

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterSpec.Cluster, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereDatacenter, nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereMachineConfigs[workerMachineConfigName], nil).AnyTimes()

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, newClusterSpec)
	assert.ErrorContains(t, err, "spec.osFamily is immutable", "OSFamily should be immutable")
}

func TestChangeDiffNoChange(t *testing.T) {
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	assert.Nil(t, provider.ChangeDiff(clusterSpec, clusterSpec))
}

func TestChangeDiffWithChange(t *testing.T) {
	provider := givenProvider(t)
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].VSphere.Version = "v0.3.18"
	})
	newClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].VSphere.Version = "v0.3.19"
	})

	wantDiff := &types.ComponentChangeDiff{
		ComponentName: "vsphere",
		NewVersion:    "v0.3.19",
		OldVersion:    "v0.3.18",
	}

	assert.Equal(t, wantDiff, provider.ChangeDiff(clusterSpec, newClusterSpec))
}

func TestVsphereProviderRunPostControlPlaneUpgrade(t *testing.T) {
	tt := newProviderTest(t)

	tt.Expect(tt.provider.RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, tt.workloadCluster, tt.managementCluster)).To(Succeed())
}

func TestProviderUpgradeNeeded(t *testing.T) {
	testCases := []struct {
		testName               string
		newManager, oldManager string
		newKubeVip, oldKubeVip string
		want                   bool
	}{
		{
			testName:   "different manager",
			newManager: "a", oldManager: "b",
			want: true,
		},
		{
			testName:   "different kubevip",
			newKubeVip: "a", oldKubeVip: "b",
			want: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			provider := givenProvider(t)
			clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].VSphere.Manager.ImageDigest = tt.oldManager
				s.VersionsBundles["1.19"].VSphere.KubeVip.ImageDigest = tt.oldKubeVip
			})

			newClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundles["1.19"].VSphere.Manager.ImageDigest = tt.newManager
				s.VersionsBundles["1.19"].VSphere.KubeVip.ImageDigest = tt.newKubeVip
			})

			g := NewWithT(t)
			g.Expect(provider.UpgradeNeeded(context.Background(), clusterSpec, newClusterSpec, nil)).To(Equal(tt.want))
		})
	}
}

func TestProviderGenerateCAPISpecForCreateWithPodIAMConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.Cluster.Spec.PodIAMConfig = &v1alpha1.PodIAMConfig{ServiceAccountIssuer: "https://test"}

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, _, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_pod_iam_config.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithMultipleMachineTemplate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, "cluster_main_diff_template.yaml")
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	datacenterConfig := givenDatacenterConfig(t, "cluster_main_diff_template.yaml")
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_diff_template_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_md.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithCustomResolvConf(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf = &v1alpha1.ResolvConf{Path: "/etc/my-custom-resolv.conf"}

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, _, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_custom_resolv_conf.yaml")
}

func TestProviderGenerateCAPISpecForCreateVersion121(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMain121Filename)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMain121Filename)
	provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_main_121_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_121_md.yaml")
}

func TestSetupAndValidateCreateManagementDoesNotCheckIfMachineAndDataCenterExist(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	setupContext(t)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider.ipValidator = ipValidator

	for _, config := range clusterSpec.VSphereMachineConfigs {
		kubectl.EXPECT().SearchVsphereMachineConfig(context.TODO(), config.Name, gomock.Any(), config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{}, nil).Times(0)
	}
	kubectl.EXPECT().SearchVsphereDatacenterConfig(context.TODO(), datacenterConfig.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return([]*v1alpha1.VSphereDatacenterConfig{}, nil).Times(0)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

func TestClusterSpecChangedNoChanges(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	cluster := &types.Cluster{
		KubeconfigFile: "test",
	}
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)

	for _, value := range clusterSpec.VSphereMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, value.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(value, nil)
	}
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	provider := newProviderWithKubectl(t, dcConfig, clusterSpec.Cluster, kubectl, ipValidator)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, clusterSpec.Cluster.Spec.DatacenterRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(dcConfig, nil)

	specChanged, err := provider.UpgradeNeeded(ctx, clusterSpec, clusterSpec, cluster)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	if specChanged {
		t.Fatalf("expected no spec change to be detected")
	}
}

func TestClusterSpecChangedDatacenterConfigChanged(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	newClusterSpec := clusterSpec.DeepCopy()
	cluster := &types.Cluster{
		KubeconfigFile: "test",
	}
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newClusterSpec.VSphereDatacenter.Spec.Datacenter = "shiny-new-api-datacenter"

	provider := newProviderWithKubectl(t, dcConfig, clusterSpec.Cluster, kubectl, ipValidator)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, clusterSpec.Cluster.Spec.DatacenterRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereDatacenter, nil)

	specChanged, err := provider.UpgradeNeeded(ctx, newClusterSpec, clusterSpec, cluster)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	if !specChanged {
		t.Fatalf("expected spec change but none was detected")
	}
}

func TestClusterSpecChangedMachineConfigsChanged(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	cluster := &types.Cluster{
		KubeconfigFile: "test",
	}
	dcConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	modifiedMachineConfig := clusterSpec.VSphereMachineConfigs[clusterSpec.Cluster.MachineConfigRefs()[0].Name].DeepCopy()
	modifiedMachineConfig.Spec.NumCPUs = 4
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, gomock.Any(), cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(modifiedMachineConfig, nil)
	provider := newProviderWithKubectl(t, dcConfig, clusterSpec.Cluster, kubectl, ipValidator)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, clusterSpec.Cluster.Spec.DatacenterRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(dcConfig, nil)

	specChanged, err := provider.UpgradeNeeded(ctx, clusterSpec, clusterSpec, cluster)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	if !specChanged {
		t.Fatalf("expected spec change but none was detected")
	}
}

func TestValidateMachineConfigsDatastoreUsageCreateSuccess(t *testing.T) {
	tt := newProviderTest(t)
	machineConfigs := tt.clusterSpec.VSphereMachineConfigs
	machineConfigs[tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Datastore = "test-datastore"
	for _, config := range machineConfigs {
		tt.govc.EXPECT().GetWorkloadAvailableSpace(tt.ctx, config.Spec.Datastore).Return(200.0, nil)
	}
	vSpec := NewSpec(tt.clusterSpec)
	err := tt.provider.validateDatastoreUsageForCreate(tt.ctx, vSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestValidateMachineConfigsDatastoreUsageCreateError(t *testing.T) {
	tt := newProviderTest(t)
	machineConfigs := tt.clusterSpec.VSphereMachineConfigs
	for _, config := range machineConfigs {
		tt.govc.EXPECT().GetWorkloadAvailableSpace(tt.ctx, config.Spec.Datastore).Return(50.0, nil)
	}
	vSpec := NewSpec(tt.clusterSpec)
	err := tt.provider.validateDatastoreUsageForCreate(tt.ctx, vSpec)
	thenErrorExpected(t, fmt.Sprintf("not enough space in datastore %s for given diskGiB and count for respective machine groups", tt.clusterSpec.VSphereMachineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Datastore), err)
}

func TestValidateMachineConfigsDatastoreUsageUpgradeError(t *testing.T) {
	tt := newProviderTest(t)
	cluster := &types.Cluster{
		Name: "test",
	}
	tt.kubectl.EXPECT().GetEksaCluster(tt.ctx, cluster, tt.clusterSpec.Cluster.GetName()).Return(tt.clusterSpec.Cluster.DeepCopy(), nil)
	machineConfigs := tt.clusterSpec.VSphereMachineConfigs
	for _, config := range machineConfigs {
		tt.kubectl.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, config.Name, cluster.KubeconfigFile, config.Namespace).AnyTimes()
		tt.govc.EXPECT().GetWorkloadAvailableSpace(tt.ctx, config.Spec.Datastore).Return(50.0, nil)
	}
	vSpec := NewSpec(tt.clusterSpec)
	err := tt.provider.validateDatastoreUsageForUpgrade(tt.ctx, vSpec, cluster)
	thenErrorExpected(t, fmt.Sprintf("not enough space in datastore %s for given diskGiB and count for respective machine groups", tt.clusterSpec.VSphereMachineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Datastore), err)
}

func TestValidateMachineConfigsMemoryUsageCreateSuccess(t *testing.T) {
	tt := newProviderTest(t)
	machineConfigs := tt.clusterSpec.VSphereMachineConfigs
	datacenter := tt.clusterSpec.VSphereDatacenter.Spec.Datacenter
	machineConfigs[tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.ResourcePool = "test-resourcepool"
	for _, config := range machineConfigs {
		tt.govc.EXPECT().GetResourcePoolInfo(tt.ctx, datacenter, config.Spec.ResourcePool).Return(map[string]int{MemoryAvailable: -1}, nil)
	}
	vSpec := NewSpec(tt.clusterSpec)
	err := tt.provider.validateMemoryUsage(tt.ctx, vSpec, nil)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestValidateMachineConfigsMemoryUsageCreateError(t *testing.T) {
	tt := newProviderTest(t)
	machineConfigs := tt.clusterSpec.VSphereMachineConfigs
	datacenter := tt.clusterSpec.VSphereDatacenter.Spec.Datacenter
	for _, config := range machineConfigs {
		tt.govc.EXPECT().GetResourcePoolInfo(tt.ctx, datacenter, config.Spec.ResourcePool).Return(map[string]int{MemoryAvailable: 10000}, nil)
	}
	vSpec := NewSpec(tt.clusterSpec)
	err := tt.provider.validateMemoryUsage(tt.ctx, vSpec, nil)
	resourcePool := machineConfigs[tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.ResourcePool
	thenErrorExpected(t, fmt.Sprintf("not enough memory avaialable in resource pool %v for given memoryMiB and count for respective machine groups", resourcePool), err)
}

func TestSetupAndValidateCreateClusterMemoryUsageError(t *testing.T) {
	tt := newProviderTest(t)
	tt.setExpectationForSetup()
	tt.setExpectationForVCenterValidation()
	tt.setExpectationsForDefaultDiskAndCloneModeGovcCalls()
	tt.setExpectationsForMachineConfigsVCenterValidation()
	datacenter := tt.clusterSpec.VSphereDatacenter.Spec.Datacenter
	cpMachineConfig := tt.machineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	for _, mc := range tt.machineConfigs {
		tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, mc.Spec.Template).Return(mc.Spec.Template, nil).AnyTimes()
	}
	tt.govc.EXPECT().GetTags(tt.ctx, cpMachineConfig.Spec.Template).Return([]string{eksd119ReleaseTag, ubuntuOSTag}, nil)
	tt.govc.EXPECT().ListTags(tt.ctx)
	tt.govc.EXPECT().GetWorkloadAvailableSpace(tt.ctx, cpMachineConfig.Spec.Datastore).Return(1000.0, nil).AnyTimes()
	tt.govc.EXPECT().GetResourcePoolInfo(tt.ctx, datacenter, cpMachineConfig.Spec.ResourcePool).Return(nil, fmt.Errorf("error"))
	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)
	thenErrorExpected(t, "validating vsphere machine configs resource pool memory usage: calculating memory usage for machine config test-cp: error", err)
}

func TestValidateMachineConfigsMemoryUsageUpgradeSuccess(t *testing.T) {
	tt := newProviderTest(t)
	cluster := &types.Cluster{
		Name: "test",
	}
	tt.kubectl.EXPECT().GetEksaCluster(tt.ctx, cluster, tt.clusterSpec.Cluster.GetName()).Return(tt.clusterSpec.Cluster.DeepCopy(), nil)
	vSpec := NewSpec(tt.clusterSpec)
	vSpec.Cluster.Spec.ControlPlaneConfiguration.Count += 2
	// change the worker node group to test there is no negative count scenario
	wnMachineConfig := getMachineConfig(vSpec.Spec, "test-wn")
	newMachineConfigName := "new-test-wn"
	newWorkerMachineConfig := wnMachineConfig.DeepCopy()
	newWorkerMachineConfig.Name = newMachineConfigName
	vSpec.VSphereMachineConfigs[newMachineConfigName] = newWorkerMachineConfig
	vSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = newMachineConfigName
	machineConfigs := tt.clusterSpec.VSphereMachineConfigs
	datacenter := tt.clusterSpec.VSphereDatacenter.Spec.Datacenter
	for _, config := range machineConfigs {
		tt.kubectl.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, config.Name, cluster.KubeconfigFile, config.Namespace).Return(config, nil).AnyTimes()
		tt.govc.EXPECT().GetResourcePoolInfo(tt.ctx, datacenter, config.Spec.ResourcePool).Return(map[string]int{MemoryAvailable: -1}, nil).AnyTimes()
	}
	err := tt.provider.validateMemoryUsage(tt.ctx, vSpec, cluster)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestValidateMachineConfigsMemoryUsageUpgradeError(t *testing.T) {
	tt := newProviderTest(t)
	cluster := &types.Cluster{
		Name: "test",
	}
	tt.kubectl.EXPECT().GetEksaCluster(tt.ctx, cluster, tt.clusterSpec.Cluster.GetName()).Return(tt.clusterSpec.Cluster.DeepCopy(), nil)
	machineConfigs := tt.clusterSpec.VSphereMachineConfigs
	datacenter := tt.clusterSpec.VSphereDatacenter.Spec.Datacenter
	for _, config := range machineConfigs {
		tt.kubectl.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, config.Name, cluster.KubeconfigFile, config.Namespace).AnyTimes()
		tt.govc.EXPECT().GetResourcePoolInfo(tt.ctx, datacenter, config.Spec.ResourcePool).Return(map[string]int{MemoryAvailable: 10000}, nil)
	}
	vSpec := NewSpec(tt.clusterSpec)
	vSpec.Cluster.Spec.ControlPlaneConfiguration.Count += 2
	err := tt.provider.validateMemoryUsage(tt.ctx, vSpec, cluster)
	resourcePool := machineConfigs[tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.ResourcePool
	thenErrorExpected(t, fmt.Sprintf("not enough memory avaialable in resource pool %v for given memoryMiB and count for respective machine groups", resourcePool), err)
}

func TestSetupAndValidateUpgradeClusterMemoryUsageError(t *testing.T) {
	tt := newProviderTest(t)
	cluster := &types.Cluster{
		Name: "test",
	}
	tt.setExpectationForSetup()
	tt.setExpectationForVCenterValidation()
	tt.setExpectationsForDefaultDiskAndCloneModeGovcCalls()
	tt.setExpectationsForMachineConfigsVCenterValidation()
	tt.kubectl.EXPECT().GetEksaCluster(tt.ctx, cluster, tt.clusterSpec.Cluster.GetName()).Return(tt.clusterSpec.Cluster.DeepCopy(), nil).Times(1)
	tt.kubectl.EXPECT().GetEksaVSphereMachineConfig(tt.ctx, gomock.Any(), cluster.KubeconfigFile, tt.clusterSpec.Cluster.GetNamespace()).AnyTimes()
	cpMachineConfig := tt.machineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, cpMachineConfig.Spec.Template).Return(cpMachineConfig.Spec.Template, nil).AnyTimes()
	tt.govc.EXPECT().GetTags(tt.ctx, cpMachineConfig.Spec.Template).Return([]string{eksd119ReleaseTag, ubuntuOSTag}, nil)
	tt.govc.EXPECT().ListTags(tt.ctx)
	tt.govc.EXPECT().GetWorkloadAvailableSpace(tt.ctx, cpMachineConfig.Spec.Datastore).Return(1000.0, nil).AnyTimes()
	datacenter := tt.clusterSpec.VSphereDatacenter.Spec.Datacenter
	tt.govc.EXPECT().GetResourcePoolInfo(tt.ctx, datacenter, cpMachineConfig.Spec.ResourcePool).Return(nil, fmt.Errorf("error"))
	err := tt.provider.SetupAndValidateUpgradeCluster(tt.ctx, cluster, tt.clusterSpec, tt.clusterSpec)
	thenErrorExpected(t, "validating vsphere machine configs resource pool memory usage: calculating memory usage for machine config test-cp: error", err)
}

func TestValidateMachineConfigsNameUniquenessSuccess(t *testing.T) {
	tt := newProviderTest(t)
	cluster := &types.Cluster{
		Name: "test",
	}
	prevSpec := tt.clusterSpec.DeepCopy()
	prevSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "prev-test-cp"
	prevSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "prev-test-etcd"
	tt.kubectl.EXPECT().GetEksaCluster(tt.ctx, cluster, tt.clusterSpec.Cluster.Name).Return(prevSpec.Cluster, nil)
	machineConfigs := tt.clusterSpec.VSphereMachineConfigs
	for _, config := range machineConfigs {
		tt.kubectl.EXPECT().SearchVsphereMachineConfig(tt.ctx, config.Name, cluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{}, nil).AnyTimes()
	}

	err := tt.provider.validateMachineConfigsNameUniqueness(tt.ctx, cluster, tt.clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestValidateMachineConfigsNameUniquenessError(t *testing.T) {
	tt := newProviderTest(t)
	cluster := &types.Cluster{
		Name: "test",
	}
	prevSpec := tt.clusterSpec.DeepCopy()
	prevSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "prev-test-cp"
	prevSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "prev-test-etcd"
	dummyVsphereMachineConfig := &v1alpha1.VSphereMachineConfig{
		Spec: v1alpha1.VSphereMachineConfigSpec{
			Users: []v1alpha1.UserConfiguration{{Name: "ec2-user"}},
		},
	}
	tt.kubectl.EXPECT().GetEksaCluster(tt.ctx, cluster, tt.clusterSpec.Cluster.Name).Return(prevSpec.Cluster, nil)
	machineConfigs := tt.clusterSpec.VSphereMachineConfigs
	for _, config := range machineConfigs {
		tt.kubectl.EXPECT().SearchVsphereMachineConfig(tt.ctx, config.Name, cluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{dummyVsphereMachineConfig}, nil).AnyTimes()
	}
	err := tt.provider.validateMachineConfigsNameUniqueness(tt.ctx, cluster, tt.clusterSpec)
	thenErrorExpected(t, fmt.Sprintf("control plane VSphereMachineConfig %s already exists", tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name), err)
}

func TestProviderGenerateCAPISpecForCreateCloudProviderCredentials(t *testing.T) {
	tests := []struct {
		testName   string
		wantCPFile string
		envMap     map[string]string
	}{
		{
			testName:   "specify cloud provider credentials",
			wantCPFile: "testdata/expected_results_main_cp_cloud_provider_credentials.yaml",
			envMap:     map[string]string{config.EksavSphereCPUsernameKey: "EksavSphereCPUsername", config.EksavSphereCPPasswordKey: "EksavSphereCPPassword"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			setupContext(t)

			previousValues := map[string]string{}
			for k, v := range tt.envMap {
				previousValues[k] = os.Getenv(k)
				if err := os.Setenv(k, v); err != nil {
					t.Fatalf(err.Error())
				}
			}

			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			ipValidator := mocks.NewMockIPValidator(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

			datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}
			ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

			err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
			if err != nil {
				t.Fatalf("failed to setup and validate: %v", err)
			}

			cp, _, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
			if err != nil {
				t.Fatalf("failed to generate cluster api spec contents: %v", err)
			}
			test.AssertContentToFile(t, string(cp), tt.wantCPFile)
			for k, v := range previousValues {
				if err := os.Setenv(k, v); err != nil {
					t.Fatalf(err.Error())
				}
			}
		})
	}
}

func TestVsphereProviderMachineConfigsSelfManagedCluster(t *testing.T) {
	tt := newProviderTest(t)
	machineConfigs := tt.provider.MachineConfigs(tt.clusterSpec)
	tt.Expect(machineConfigs).To(HaveLen(3))
	for _, m := range machineConfigs {
		tt.Expect(m).To(BeAssignableToTypeOf(&v1alpha1.VSphereMachineConfig{}))
		machineConfig := m.(*v1alpha1.VSphereMachineConfig)
		tt.Expect(machineConfig.IsManaged()).To(BeFalse())

		if machineConfig.Name == tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name {
			tt.Expect(machineConfig.IsControlPlane()).To(BeTrue())
		}

		if machineConfig.Name == tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name {
			tt.Expect(machineConfig.IsEtcd()).To(BeTrue())
		}
	}
}

func TestVsphereProviderMachineConfigsManagedCluster(t *testing.T) {
	tt := newProviderTest(t)
	tt.clusterSpec.Cluster.SetManagedBy("my-management-cluster")
	machineConfigs := tt.provider.MachineConfigs(tt.clusterSpec)
	tt.Expect(machineConfigs).To(HaveLen(3))
	for _, m := range machineConfigs {
		tt.Expect(m).To(BeAssignableToTypeOf(&v1alpha1.VSphereMachineConfig{}))
		machineConfig := m.(*v1alpha1.VSphereMachineConfig)
		tt.Expect(machineConfig.IsManaged()).To(BeTrue())

		if machineConfig.Name == tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name {
			tt.Expect(machineConfig.IsControlPlane()).To(BeTrue())
		}

		if machineConfig.Name == tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name {
			tt.Expect(machineConfig.IsEtcd()).To(BeTrue())
		}
	}
}

func TestProviderGenerateDeploymentFileForBottleRocketWithNTPConfig(t *testing.T) {
	clusterSpecManifest := "cluster_bottlerocket_ntp_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	vscb, _ := newMockVSphereClientBuilder(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	v := NewValidator(govc, vscb)
	govc.osTag = bottlerocketOSTag
	provider := newProvider(
		t,
		datacenterConfig,
		clusterSpec.Cluster,
		govc,
		kubectl,
		v,
		ipValidator,
	)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_bottlerocket_ntp_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_bottlerocket_ntp_config_md.yaml")
}

func TestProviderGenerateDeploymentFileForUbuntuWithNTPConfig(t *testing.T) {
	clusterSpecManifest := "cluster_ubuntu_ntp_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	vscb, _ := newMockVSphereClientBuilder(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	v := NewValidator(govc, vscb)
	provider := newProvider(
		t,
		datacenterConfig,
		clusterSpec.Cluster,
		govc,
		kubectl,
		v,
		ipValidator,
	)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_ubuntu_ntp_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_ubuntu_ntp_config_md.yaml")
}

func TestProviderGenerateDeploymentFileForBottlerocketWithBottlerocketSettingsConfig(t *testing.T) {
	clusterSpecManifest := "cluster_bottlerocket_settings_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	vscb, _ := newMockVSphereClientBuilder(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	v := NewValidator(govc, vscb)
	govc.osTag = bottlerocketOSTag
	provider := newProvider(
		t,
		datacenterConfig,
		clusterSpec.Cluster,
		govc,
		kubectl,
		v,
		ipValidator,
	)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_bottlerocket_settings_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_bottlerocket_settings_config_md.yaml")
}

func TestProviderGenerateDeploymentFileForBottlerocketWithKernelConfig(t *testing.T) {
	clusterSpecManifest := "cluster_bottlerocket_kernel_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	vscb, _ := newMockVSphereClientBuilder(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	v := NewValidator(govc, vscb)
	govc.osTag = bottlerocketOSTag
	provider := newProvider(
		t,
		datacenterConfig,
		clusterSpec.Cluster,
		govc,
		kubectl,
		v,
		ipValidator,
	)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_bottlerocket_kernel_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_bottlerocket_kernel_config_md.yaml")
}

func TestProviderGenerateDeploymentFileForBottlerocketWithBootSettings(t *testing.T) {
	clusterSpecManifest := "cluster_bottlerocket_boot_settings_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	vscb, _ := newMockVSphereClientBuilder(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	v := NewValidator(govc, vscb)
	govc.osTag = bottlerocketOSTag
	provider := newProvider(
		t,
		datacenterConfig,
		clusterSpec.Cluster,
		govc,
		kubectl,
		v,
		ipValidator,
	)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_bottlerocket_boot_settings_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_bottlerocket_boot_settings_config_md.yaml")
}

func TestProviderGenerateDeploymentFileForBottlerocketWithTrustedCertBundles(t *testing.T) {
	clusterSpecManifest := "cluster_bottlerocket_cert_bundles_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	vscb, _ := newMockVSphereClientBuilder(mockCtrl)
	ipValidator := mocks.NewMockIPValidator(mockCtrl)
	ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)
	v := NewValidator(govc, vscb)
	govc.osTag = bottlerocketOSTag
	provider := newProvider(
		t,
		datacenterConfig,
		clusterSpec.Cluster,
		govc,
		kubectl,
		v,
		ipValidator,
	)
	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_bottlerocket_cert_bundles_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_bottlerocket_cert_bundles_config_md.yaml")
}

func TestProviderGenerateCAPISpecForUpgradeEtcdEncryption(t *testing.T) {
	tests := []struct {
		testName          string
		clusterconfigFile string
		wantCPFile        string
		wantMDFile        string
	}{
		{
			testName:          "etcd-encryption",
			clusterconfigFile: "cluster_ubuntu_etcd_encryption.yaml",
			wantCPFile:        "testdata/expected_results_ubuntu_etcd_encryption_cp.yaml",
			wantMDFile:        "testdata/expected_results_main_121_md.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			setupContext(t)
			ctx := context.Background()
			kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			clusterSpec := givenClusterSpec(t, tt.clusterconfigFile)

			oldCP := &controlplanev1.KubeadmControlPlane{
				Spec: controlplanev1.KubeadmControlPlaneSpec{
					MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
						InfrastructureRef: v1.ObjectReference{
							Name: "test-control-plane-template-1234567890000",
						},
					},
				},
			}
			oldMD := &clusterv1.MachineDeployment{
				Spec: clusterv1.MachineDeploymentSpec{
					Template: clusterv1.MachineTemplateSpec{
						Spec: clusterv1.MachineSpec{
							InfrastructureRef: v1.ObjectReference{
								Name: "test-md-0-1234567890000",
							},
							Bootstrap: clusterv1.Bootstrap{
								ConfigRef: &v1.ObjectReference{
									Name: "test-md-0-template-1234567890000",
								},
							},
						},
					},
				},
			}
			etcdadmCluster := &etcdv1.EtcdadmCluster{
				Spec: etcdv1.EtcdadmClusterSpec{
					InfrastructureTemplate: v1.ObjectReference{
						Name: "test-etcd-template-1234567890000",
					},
				},
			}

			ipValidator := mocks.NewMockIPValidator(mockCtrl)
			ipValidator.EXPECT().ValidateControlPlaneIPUniqueness(clusterSpec.Cluster).Return(nil)

			datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
			provider := newProviderWithKubectl(t, datacenterConfig, clusterSpec.Cluster, kubectl, ipValidator)
			if provider == nil {
				t.Fatalf("provider object is nil")
			}

			err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
			if err != nil {
				t.Fatalf("failed to setup and validate: %v", err)
			}

			controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
			workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
			machineDeploymentName := fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Name)
			etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name

			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(datacenterConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, controlPlaneMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereMachineConfigs[controlPlaneMachineConfigName], nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, workerNodeMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereMachineConfigs[workerNodeMachineConfigName], nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, etcdMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(clusterSpec.VSphereMachineConfigs[etcdMachineConfigName], nil)
			kubectl.EXPECT().GetKubeadmControlPlane(ctx, cluster, clusterSpec.Cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(oldCP, nil)
			kubectl.EXPECT().GetMachineDeployment(ctx, machineDeploymentName, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(oldMD, nil).Times(2)
			kubectl.EXPECT().GetEtcdadmCluster(ctx, cluster, clusterSpec.Cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(etcdadmCluster, nil)

			cp, md, err := provider.GenerateCAPISpecForUpgrade(context.Background(), bootstrapCluster, cluster, clusterSpec, clusterSpec.DeepCopy())
			if err != nil {
				t.Fatalf("failed to generate cluster api spec contents: %v", err)
			}

			test.AssertContentToFile(t, string(cp), tt.wantCPFile)
			test.AssertContentToFile(t, string(md), tt.wantMDFile)
		})
	}
}

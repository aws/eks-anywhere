package vsphere

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"path"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/golang/mock/gomock"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	testClusterConfigMainFilename    = "cluster_main.yaml"
	testClusterConfigMain121Filename = "cluster_main_121.yaml"
	testDataDir                      = "testdata"
	expectedVSphereName              = "vsphere"
	expectedVSphereUsername          = "vsphere_username"
	expectedVSpherePassword          = "vsphere_password"
	expectedVSphereServer            = "vsphere_server"
	expectedExpClusterResourceSet    = "expClusterResourceSetKey"
	eksd119Release                   = "kubernetes-1-19-eks-4"
	eksd119ReleaseTag                = "eksdRelease:kubernetes-1-19-eks-4"
	eksd121ReleaseTag                = "eksdRelease:kubernetes-1-21-eks-4"
	ubuntuOSTag                      = "os:ubuntu"
	bottlerocketOSTag                = "os:bottlerocket"
	testTemplate                     = "/SDDC-Datacenter/vm/Templates/ubuntu-1804-kube-v1.19.6"
)

type DummyProviderGovcClient struct {
	osTag string
}

func NewDummyProviderGovcClient() *DummyProviderGovcClient {
	return &DummyProviderGovcClient{osTag: ubuntuOSTag}
}

func (pc *DummyProviderGovcClient) TemplateHasSnapshot(ctx context.Context, template string) (bool, error) {
	return false, nil
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

func (pc *DummyProviderGovcClient) SearchTemplate(ctx context.Context, datacenter string, machineConfig *v1alpha1.VSphereMachineConfig) (string, error) {
	return machineConfig.Spec.Template, nil
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

func (pc *DummyProviderGovcClient) DeployTemplateFromLibrary(ctx context.Context, templateDir, templateName, library, datacenter, datastore, resourcePool string, resizeDisk2 bool) error {
	return nil
}

func (pc *DummyProviderGovcClient) ResizeDisk(ctx context.Context, template, diskName string, diskSizeInGB int) error {
	return nil
}

func (pc *DummyProviderGovcClient) ImportTemplate(ctx context.Context, library, ovaURL, name string) error {
	return nil
}

func (pc *DummyProviderGovcClient) GetTags(ctx context.Context, path string) (tags []string, err error) {
	return []string{eksd119ReleaseTag, eksd121ReleaseTag, pc.osTag}, nil
}

func (pc *DummyProviderGovcClient) ListTags(ctx context.Context) ([]string, error) {
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

type DummyNetClient struct{}

func (n *DummyNetClient) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	// add dummy case for coverage
	if address == "255.255.255.255:22" {
		return &net.IPConn{}, nil
	}
	return nil, errors.New("")
}

func givenClusterConfig(t *testing.T, fileName string) *v1alpha1.Cluster {
	return givenClusterSpec(t, fileName).Cluster
}

func givenClusterSpec(t *testing.T, fileName string) *cluster.Spec {
	return test.NewFullClusterSpec(t, path.Join(testDataDir, fileName))
}

func givenEmptyClusterSpec() *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.KubeVersion = "1.19"
		s.VersionsBundle.EksD.Name = eksd119Release
		s.Cluster.Namespace = "test-namespace"
	})
}

func fillClusterSpecWithClusterConfig(spec *cluster.Spec, clusterConfig *v1alpha1.Cluster) {
	spec.Cluster = clusterConfig
}

func givenDatacenterConfig(t *testing.T, fileName string) *v1alpha1.VSphereDatacenterConfig {
	datacenterConfig, err := v1alpha1.GetVSphereDatacenterConfig(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get datacenter config from file: %v", err)
	}
	return datacenterConfig
}

func givenMachineConfigs(t *testing.T, fileName string) map[string]*v1alpha1.VSphereMachineConfig {
	machineConfigs, err := v1alpha1.GetVSphereMachineConfigs(path.Join(testDataDir, fileName))
	if err != nil {
		t.Fatalf("unable to get machine configs from file")
	}
	return machineConfigs
}

func givenProvider(t *testing.T) *vsphereProvider {
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(
		t,
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		nil,
	)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}
	return provider
}

type testContext struct {
	oldUsername                string
	isUsernameSet              bool
	oldPassword                string
	isPasswordSet              bool
	oldServername              string
	isServernameSet            bool
	oldExpClusterResourceSet   string
	isExpClusterResourceSetSet bool
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

func (tctx *testContext) SaveContext() {
	tctx.oldUsername, tctx.isUsernameSet = os.LookupEnv(EksavSphereUsernameKey)
	tctx.oldPassword, tctx.isPasswordSet = os.LookupEnv(EksavSpherePasswordKey)
	tctx.oldServername, tctx.isServernameSet = os.LookupEnv(vSpherePasswordKey)
	tctx.oldExpClusterResourceSet, tctx.isExpClusterResourceSetSet = os.LookupEnv(vSpherePasswordKey)
	os.Setenv(EksavSphereUsernameKey, expectedVSphereUsername)
	os.Setenv(vSphereUsernameKey, os.Getenv(EksavSphereUsernameKey))
	os.Setenv(EksavSpherePasswordKey, expectedVSpherePassword)
	os.Setenv(vSpherePasswordKey, os.Getenv(EksavSpherePasswordKey))
	os.Setenv(vSphereServerKey, expectedVSphereServer)
	os.Setenv(expClusterResourceSetKey, expectedExpClusterResourceSet)
}

func (tctx *testContext) RestoreContext() {
	if tctx.isUsernameSet {
		os.Setenv(EksavSphereUsernameKey, tctx.oldUsername)
	} else {
		os.Unsetenv(EksavSphereUsernameKey)
	}
	if tctx.isPasswordSet {
		os.Setenv(EksavSpherePasswordKey, tctx.oldPassword)
	} else {
		os.Unsetenv(EksavSpherePasswordKey)
	}
}

func setupContext(t *testing.T) {
	var tctx testContext
	tctx.SaveContext()
	t.Cleanup(func() {
		tctx.RestoreContext()
	})
}

type providerTest struct {
	*WithT
	ctx                                context.Context
	managementCluster, workloadCluster *types.Cluster
	provider                           *vsphereProvider
	cluster                            *v1alpha1.Cluster
	clusterSpec                        *cluster.Spec
	datacenterConfig                   *v1alpha1.VSphereDatacenterConfig
	machineConfigs                     map[string]*v1alpha1.VSphereMachineConfig
	kubectl                            *mocks.MockProviderKubectlClient
	govc                               *mocks.MockProviderGovcClient
	resourceSetManager                 *mocks.MockClusterResourceSetManager
}

func newProviderTest(t *testing.T) *providerTest {
	setupContext(t)
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(ctrl)
	govc := mocks.NewMockProviderGovcClient(ctrl)
	resourceSetManager := mocks.NewMockClusterResourceSetManager(ctrl)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	provider := newProvider(
		t,
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		govc,
		kubectl,
		resourceSetManager,
	)
	return &providerTest{
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
		provider:           provider,
		cluster:            clusterConfig,
		clusterSpec:        givenClusterSpec(t, testClusterConfigMainFilename),
		datacenterConfig:   datacenterConfig,
		machineConfigs:     machineConfigs,
		kubectl:            kubectl,
		govc:               govc,
		resourceSetManager: resourceSetManager,
	}
}

func (tt *providerTest) setExpectationsForDefaultDiskGovcCalls() {
	for _, m := range tt.machineConfigs {
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

func TestNewProvider(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider := newProviderWithKubectl(
		t,
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		kubectl,
	)

	if provider == nil {
		t.Fatalf("provider object is nil")
	}
}

func newProviderWithKubectl(t *testing.T, datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfigs map[string]*v1alpha1.VSphereMachineConfig, clusterConfig *v1alpha1.Cluster, kubectl ProviderKubectlClient) *vsphereProvider {
	ctrl := gomock.NewController(t)
	resourceSetManager := mocks.NewMockClusterResourceSetManager(ctrl)
	return newProvider(
		t,
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		NewDummyProviderGovcClient(),
		kubectl,
		resourceSetManager,
	)
}

func newProviderWithGovc(t *testing.T, datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfigs map[string]*v1alpha1.VSphereMachineConfig, clusterConfig *v1alpha1.Cluster, govc ProviderGovcClient) *vsphereProvider {
	ctrl := gomock.NewController(t)
	resourceSetManager := mocks.NewMockClusterResourceSetManager(ctrl)
	kubectl := mocks.NewMockProviderKubectlClient(ctrl)
	return newProvider(
		t,
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		govc,
		kubectl,
		resourceSetManager,
	)
}

func newProvider(t *testing.T, datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfigs map[string]*v1alpha1.VSphereMachineConfig, clusterConfig *v1alpha1.Cluster, govc ProviderGovcClient, kubectl ProviderKubectlClient, resourceSetManager ClusterResourceSetManager) *vsphereProvider {
	_, writer := test.NewWriter(t)
	return NewProviderCustomNet(
		datacenterConfig,
		machineConfigs,
		clusterConfig,
		govc,
		kubectl,
		writer,
		&DummyNetClient{},
		test.FakeNow,
		false,
		resourceSetManager,
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
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			var tctx testContext
			tctx.SaveContext()
			defer tctx.RestoreContext()
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
			vsphereMachineConfig := &v1alpha1.VSphereMachineConfig{
				Spec: v1alpha1.VSphereMachineConfigSpec{
					Users: []v1alpha1.UserConfiguration{
						{
							Name:              "capv",
							SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
						},
					},
				},
			}

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereDatacenter, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			machineConfigs := givenMachineConfigs(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)
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
			var tctx testContext
			tctx.SaveContext()
			defer tctx.RestoreContext()
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
			vsphereMachineConfig := &v1alpha1.VSphereMachineConfig{
				Spec: v1alpha1.VSphereMachineConfigSpec{
					Users: []v1alpha1.UserConfiguration{
						{
							Name:              "capv",
							SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
						},
					},
				},
			}

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereDatacenter, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			machineConfigs := givenMachineConfigs(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)
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
			var tctx testContext
			tctx.SaveContext()
			defer tctx.RestoreContext()
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
			vsphereMachineConfig := &v1alpha1.VSphereMachineConfig{
				Spec: v1alpha1.VSphereMachineConfigSpec{
					Users: []v1alpha1.UserConfiguration{
						{
							Name:              "capv",
							SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
						},
					},
				},
			}
			newClusterSpec := givenClusterSpec(t, tt.clusterconfigFile)
			newConfig := v1alpha1.WorkerNodeGroupConfiguration{Count: 1, MachineGroupRef: &v1alpha1.Ref{Name: "test-wn", Kind: "VSphereMachineConfig"}, Name: "md-2"}
			newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = append(newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations, newConfig)

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup2MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereDatacenter, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil).AnyTimes()
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", cluster.Name), map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"}, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster)))

			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			machineConfigs := givenMachineConfigs(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)
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
			var tctx testContext
			tctx.SaveContext()
			defer tctx.RestoreContext()
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
			vsphereMachineConfig := &v1alpha1.VSphereMachineConfig{
				Spec: v1alpha1.VSphereMachineConfigSpec{
					Users: []v1alpha1.UserConfiguration{
						{
							Name:              "capv",
							SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
						},
					},
				},
			}

			kubectl.EXPECT().GetMachineDeployment(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(workerNodeGroup1MachineDeployment(), nil)
			kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name).Return(clusterSpec.Cluster, nil)
			kubectl.EXPECT().GetEksaVSphereDatacenterConfig(ctx, cluster.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereDatacenter, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(vsphereMachineConfig, nil)
			kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", cluster.Name), map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"}, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster)))
			datacenterConfig := givenDatacenterConfig(t, tt.clusterconfigFile)
			machineConfigs := givenMachineConfigs(t, tt.clusterconfigFile)
			provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)
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

func TestProviderGenerateCAPISpecForUpgradeNotUpdateMachineTemplate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
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

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)
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
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, controlPlaneMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(machineConfigs[controlPlaneMachineConfigName], nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, workerNodeMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(machineConfigs[workerNodeMachineConfigName], nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(ctx, etcdMachineConfigName, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return(machineConfigs[etcdMachineConfigName], nil)
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
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)
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
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, "cluster_main_multiple_worker_node_groups.yaml")

	datacenterConfig := givenDatacenterConfig(t, "cluster_main_multiple_worker_node_groups.yaml")
	machineConfigs := givenMachineConfigs(t, "cluster_main_multiple_worker_node_groups.yaml")
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)
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

func TestProviderGenerateStorageClass(t *testing.T) {
	provider := givenProvider(t)

	storageClassManifest := provider.GenerateStorageClass()
	if storageClassManifest == nil {
		t.Fatalf("Expected storageClassManifest")
	}
}

func TestProviderGenerateCAPISpecForCreateWithBottlerocketAndExternalEtcd(t *testing.T) {
	clusterSpecManifest := "cluster_bottlerocket_external_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	resourceSetManager := mocks.NewMockClusterResourceSetManager(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	govc.osTag = bottlerocketOSTag
	provider := newProvider(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, govc, kubectl, resourceSetManager)

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
	resourceSetManager := mocks.NewMockClusterResourceSetManager(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	govc.osTag = bottlerocketOSTag
	provider := newProvider(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, govc, kubectl, resourceSetManager)
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
	resourceSetManager := mocks.NewMockClusterResourceSetManager(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	govc.osTag = bottlerocketOSTag
	provider := newProvider(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, govc, kubectl, resourceSetManager)
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

func TestProviderGenerateDeploymentFileWithMirrorConfig(t *testing.T) {
	clusterSpecManifest := "cluster_mirror_config.yaml"
	mockCtrl := gomock.NewController(t)
	setupContext(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	datacenterConfig := givenDatacenterConfig(t, clusterSpecManifest)
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	govc.osTag = ubuntuOSTag
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)

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
	machineConfigs := givenMachineConfigs(t, clusterSpecManifest)
	ctx := context.Background()
	govc := NewDummyProviderGovcClient()
	govc.osTag = ubuntuOSTag
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)

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

	bootstrapClusterOps, err := provider.BootstrapClusterOpts()
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
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

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
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	os.Unsetenv(EksavSphereUsernameKey)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: EKSA_VSPHERE_USERNAME is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoPassword(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	provider := givenProvider(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	os.Unsetenv(EksavSpherePasswordKey)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: EKSA_VSPHERE_PASSWORD is not set or is empty", err)
}

func TestSetupAndValidateCreateWorkloadClusterSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.Cluster.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}
	for _, config := range newMachineConfigs {
		kubectl.EXPECT().SearchVsphereMachineConfig(context.TODO(), config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{}, nil)
	}
	kubectl.EXPECT().SearchVsphereDatacenterConfig(context.TODO(), datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return([]*v1alpha1.VSphereDatacenterConfig{}, nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

func TestSetupAndValidateCreateWorkloadClusterFailsIfMachineExists(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.Cluster.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}

	idx := 0
	var existingMachine string
	for _, config := range newMachineConfigs {
		if idx == 0 {
			kubectl.EXPECT().SearchVsphereMachineConfig(context.TODO(), config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{config}, nil)
			existingMachine = config.Name
		} else {
			kubectl.EXPECT().SearchVsphereMachineConfig(context.TODO(), config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{}, nil).MaxTimes(1)
		}
		idx++
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, fmt.Sprintf("VSphereMachineConfig %s already exists", existingMachine), err)
}

func TestSetupAndValidateSelfManagedClusterSkipMachineNameValidateSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}

	kubectl.EXPECT().SearchVsphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

func TestSetupAndValidateCreateWorkloadClusterFailsIfDatacenterExists(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.Cluster.SetManagedBy("management-cluster")
	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}

	for _, config := range newMachineConfigs {
		kubectl.EXPECT().SearchVsphereMachineConfig(context.TODO(), config.Name, clusterSpec.ManagementCluster.KubeconfigFile, config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{}, nil)
	}
	kubectl.EXPECT().SearchVsphereDatacenterConfig(context.TODO(), datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace).Return([]*v1alpha1.VSphereDatacenterConfig{datacenterConfig}, nil)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, fmt.Sprintf("VSphereDatacenter %s already exists", datacenterConfig.Name), err)
}

func TestSetupAndValidateSelfManagedClusterSkipDatacenterNameValidateSuccess(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	clusterSpec.ManagementCluster = &types.Cluster{
		Name:               "management-cluster",
		KubeconfigFile:     "kc.kubeconfig",
		ExistingManagement: true,
	}

	kubectl.EXPECT().SearchVsphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	kubectl.EXPECT().SearchVsphereDatacenterConfig(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

func TestSetupAndValidateDeleteCluster(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	err := provider.SetupAndValidateDeleteCluster(ctx)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateDeleteClusterNoPassword(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	os.Unsetenv(EksavSpherePasswordKey)

	err := provider.SetupAndValidateDeleteCluster(ctx)

	thenErrorExpected(t, "failed setup and validations: EKSA_VSPHERE_PASSWORD is not set or is empty", err)
}

func TestSetupAndValidateUpgradeCluster(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	cluster := &types.Cluster{}
	provider := givenProvider(t)
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterNoUsername(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	provider := givenProvider(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	os.Unsetenv(EksavSphereUsernameKey)

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: EKSA_VSPHERE_USERNAME is not set or is empty", err)
}

func TestSetupAndValidateUpgradeClusterNoPassword(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	provider := givenProvider(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	os.Unsetenv(EksavSpherePasswordKey)

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)

	thenErrorExpected(t, "failed setup and validations: EKSA_VSPHERE_PASSWORD is not set or is empty", err)
}

func TestSetupAndValidateUpgradeClusterCPSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterWorkerSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestSetupAndValidateUpgradeClusterEtcdSshNotExists(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	cluster := &types.Cluster{}
	kubectl.EXPECT().GetEksaCluster(ctx, cluster, clusterSpec.Cluster.GetName()).Return(clusterSpec.Cluster.DeepCopy(), nil)

	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
}

func TestVersion(t *testing.T) {
	vSphereProviderVersion := "v0.7.10"
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	clusterSpec.VersionsBundle.VSphere.Version = vSphereProviderVersion
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	result := provider.Version(clusterSpec)
	if result != vSphereProviderVersion {
		t.Fatalf("Unexpected version expected <%s> actual=<%s>", vSphereProviderVersion, result)
	}
}

func TestProviderBootstrapSetup(t *testing.T) {
	ctx := context.Background()
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterConfig, kubectl)
	cluster := types.Cluster{
		Name:           "test",
		KubeconfigFile: "",
	}
	values := map[string]string{
		"clusterName":       clusterConfig.Name,
		"vspherePassword":   expectedVSphereUsername,
		"vsphereUsername":   expectedVSpherePassword,
		"vsphereServer":     datacenterConfig.Spec.Server,
		"vsphereDatacenter": datacenterConfig.Spec.Datacenter,
		"vsphereNetwork":    datacenterConfig.Spec.Network,
	}

	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	template, err := template.New("test").Parse(defaultSecretObject)
	if err != nil {
		t.Fatalf("template create error: %v", err)
	}
	err = template.Execute(&bytes.Buffer{}, values)
	if err != nil {
		t.Fatalf("template execute error: %v", err)
	}

	err = provider.PostBootstrapSetup(ctx, clusterConfig, &cluster)
	if err != nil {
		t.Fatalf("BootstrapSetup error %v", err)
	}
}

func TestProviderUpdateSecret(t *testing.T) {
	ctx := context.Background()
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterConfig, kubectl)
	cluster := types.Cluster{
		Name:           "test",
		KubeconfigFile: "",
	}
	values := map[string]string{
		"vspherePassword":     expectedVSphereUsername,
		"vsphereUsername":     expectedVSpherePassword,
		"eksaLicense":         "",
		"eksaSystemNamespace": constants.EksaSystemNamespace,
	}

	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	kubectl.EXPECT().GetNamespace(ctx, gomock.Any(), constants.EksaSystemNamespace).Return(errors.New("test"))
	kubectl.EXPECT().CreateNamespace(ctx, gomock.Any(), constants.EksaSystemNamespace)
	kubectl.EXPECT().ApplyKubeSpecFromBytes(ctx, gomock.Any(), gomock.Any())

	template, err := template.New("test").Parse(defaultSecretObject)
	if err != nil {
		t.Fatalf("template create error: %v", err)
	}
	err = template.Execute(&bytes.Buffer{}, values)
	if err != nil {
		t.Fatalf("template execute error: %v", err)
	}

	err = provider.UpdateSecrets(ctx, &cluster)
	if err != nil {
		t.Fatalf("UpdateSecrets error %v", err)
	}
}

func TestSetupAndValidateCreateClusterNoServer(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	provider := givenProvider(t)
	provider.datacenterConfig.Spec.Server = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereDatacenterConfig server is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterInsecure(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	provider.datacenterConfig.Spec.Insecure = true
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if provider.datacenterConfig.Spec.Thumbprint != "" {
		t.Fatalf("Expected=<> actual=<%s>", provider.datacenterConfig.Spec.Thumbprint)
	}
}

func TestSetupAndValidateCreateClusterNoControlPlaneEndpointIP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoDatacenter(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	provider.datacenterConfig.Spec.Datacenter = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereDatacenterConfig datacenter is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoDatastoreControlPlane(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Datastore = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereMachineConfig datastore for control plane is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoDatastoreWorker(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerMachineConfigName].Spec.Datastore = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereMachineConfig datastore for worker nodes is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoDatastoreEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Datastore = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereMachineConfig datastore for etcd machines is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoResourcePoolControlPlane(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.ResourcePool = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereMachineConfig VM resourcePool for control plane is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoResourcePoolWorker(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerMachineConfigName].Spec.ResourcePool = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereMachineConfig VM resourcePool for worker nodes is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoResourcePoolEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.ResourcePool = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereMachineConfig VM resourcePool for etcd machines is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNoNetwork(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	provider.datacenterConfig.Spec.Network = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "VSphereDatacenterConfig VM network is not set or is empty", err)
}

func TestSetupAndValidateCreateClusterNotControlPlaneVMsMemoryMiB(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.MemoryMiB = 0
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.MemoryMiB != 8192 {
		t.Fatalf("Expected=<8192> actual=<%d>", provider.machineConfigs[controlPlaneMachineConfigName].Spec.MemoryMiB)
	}
}

func TestSetupAndValidateCreateClusterNotControlPlaneVMsNumCPUs(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.NumCPUs = 0
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.NumCPUs != 2 {
		t.Fatalf("Expected=<2> actual=<%d>", provider.machineConfigs[controlPlaneMachineConfigName].Spec.NumCPUs)
	}
}

func TestSetupAndValidateCreateClusterNotWorkloadVMsMemoryMiB(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.MemoryMiB = 0
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if provider.machineConfigs[workerNodeMachineConfigName].Spec.MemoryMiB != 8192 {
		t.Fatalf("Expected=<8192> actual=<%d>", provider.machineConfigs[workerNodeMachineConfigName].Spec.MemoryMiB)
	}
}

func TestSetupAndValidateCreateClusterNotWorkloadVMsNumCPUs(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.NumCPUs = 0
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if provider.machineConfigs[workerNodeMachineConfigName].Spec.NumCPUs != 2 {
		t.Fatalf("Expected=<2> actual=<%d>", provider.machineConfigs[workerNodeMachineConfigName].Spec.NumCPUs)
	}
}

func TestSetupAndValidateCreateClusterNotEtcdVMsMemoryMiB(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.MemoryMiB = 0
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if provider.machineConfigs[etcdMachineConfigName].Spec.MemoryMiB != 8192 {
		t.Fatalf("Expected=<8192> actual=<%d>", provider.machineConfigs[etcdMachineConfigName].Spec.MemoryMiB)
	}
}

func TestSetupAndValidateCreateClusterNotEtcdVMsNumCPUs(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.NumCPUs = 0
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Unexpected error <%v>", err)
	}
	if provider.machineConfigs[etcdMachineConfigName].Spec.NumCPUs != 2 {
		t.Fatalf("Expected=<2> actual=<%d>", provider.machineConfigs[etcdMachineConfigName].Spec.NumCPUs)
	}
}

func TestSetupAndValidateCreateClusterBogusIp(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "bogus"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "cluster controlPlaneConfiguration.Endpoint.Host is invalid: bogus", err)
}

func TestSetupAndValidateCreateClusterUsedIp(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "255.255.255.255"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	thenErrorExpected(t, "cluster controlPlaneConfiguration.Endpoint.Host <255.255.255.255> is already in use, please provide a unique IP", err)
}

func TestSetupAndValidateForCreateSSHAuthorizedKeyInvalidCP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: ssh: no key found", err)
}

func TestSetupAndValidateForCreateSSHAuthorizedKeyInvalidWorker(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: ssh: no key found", err)
}

func TestSetupAndValidateForCreateSSHAuthorizedKeyInvalidEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: ssh: no key found", err)
}

func TestSetupAndValidateForUpgradeSSHAuthorizedKeyInvalidCP(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: ssh: no key found", err)
}

func TestSetupAndValidateForUpgradeSSHAuthorizedKeyInvalidWorker(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: ssh: no key found", err)
}

func TestSetupAndValidateForUpgradeSSHAuthorizedKeyInvalidEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	tempKey := "ssh-rsa AAAA    B3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = tempKey
	var tctx testContext
	tctx.SaveContext()

	cluster := &types.Cluster{}
	err := provider.SetupAndValidateUpgradeCluster(ctx, cluster, clusterSpec)
	thenErrorExpected(t, "failed setup and validations: ssh: no key found", err)
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyCP(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for control plane machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyWorker(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for worker node machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyEtcd(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey did not get generated for etcd machine")
	}
}

func TestSetupAndValidateSSHAuthorizedKeyEmptyAllMachineConfigs(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] = ""

	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for control plane machine")
	}
	if provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey has not changed for worker node machine")
	}
	if provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] == "" {
		t.Fatalf("sshAuthorizedKey not generated for etcd machines")
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] != provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] {
		t.Fatalf("sshAuthorizedKey not the same for controlplane and worker machines")
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] != provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys[0] {
		t.Fatalf("sshAuthorizedKey not the same for controlplane and etcd machines")
	}
}

func TestSetupAndValidateUsersNil(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users = nil
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users = nil
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Users = nil
	var tctx testContext
	tctx.SaveContext()

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
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].SshAuthorizedKeys = nil
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
}

func TestSetupAndValidateCreateClusterCPMachineGroupRefNil(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef = nil
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "must specify machineGroupRef for control plane", err)
	}
}

func TestSetupAndValidateCreateClusterWorkerMachineGroupRefNil(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef = nil
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "must specify machineGroupRef for worker nodes", err)
	}
}

func TestSetupAndValidateCreateClusterEtcdMachineGroupRefNil(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef = nil
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "must specify machineGroupRef for etcd machines", err)
	}
}

func TestSetupAndValidateCreateClusterCPMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "nonexistent"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "cannot find VSphereMachineConfig nonexistent for control plane", err)
	}
}

func TestSetupAndValidateCreateClusterWorkerMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = "nonexistent"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "cannot find VSphereMachineConfig nonexistent for worker nodes", err)
	}
}

func TestSetupAndValidateCreateClusterEtcdMachineGroupRefNonexistent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "nonexistent"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "cannot find VSphereMachineConfig nonexistent for etcd machines", err)
	}
}

func TestSetupAndValidateCreateClusterOsFamilyInvalid(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.OSFamily = "rhel"
	var tctx testContext
	tctx.SaveContext()
	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "control plane osFamily: rhel is not supported, please use one of the following: bottlerocket, ubuntu", err)
}

func TestSetupAndValidateCreateClusterOsFamilyInvalidWorkerNode(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	workerMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerMachineConfigName].Spec.OSFamily = "rhel"
	var tctx testContext
	tctx.SaveContext()
	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "worker node osFamily: rhel is not supported, please use one of the following: bottlerocket, ubuntu", err)
}

func TestSetupAndValidateCreateClusterOsFamilyInvalidEtcdNode(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.OSFamily = "rhel"
	var tctx testContext
	tctx.SaveContext()
	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	thenErrorExpected(t, "etcd node osFamily: rhel is not supported, please use one of the following: bottlerocket, ubuntu", err)
}

func TestSetupAndValidateCreateClusterOsFamilyDifferent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.OSFamily = "bottlerocket"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "control plane and worker nodes must have the same osFamily specified", err)
	}
}

func TestSetupAndValidateCreateClusterOsFamilyDifferentForEtcd(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.OSFamily = "bottlerocket"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "control plane and etcd machines must have the same osFamily specified", err)
	}
}

func TestSetupAndValidateCreateClusterOsFamilyEmpty(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)
	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	govc := NewDummyProviderGovcClient()
	govc.osTag = bottlerocketOSTag
	provider := newProviderWithGovc(t, datacenterConfig, machineConfigs, clusterConfig, govc)
	provider.providerGovcClient = govc
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.OSFamily = ""
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Users[0].Name = ""
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.OSFamily = ""
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Users[0].Name = ""
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.OSFamily = ""
	provider.machineConfigs[etcdMachineConfigName].Spec.Users[0].Name = ""
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	if provider.machineConfigs[controlPlaneMachineConfigName].Spec.OSFamily != v1alpha1.Bottlerocket {
		t.Fatalf("got osFamily for control plane machine as %v, want %v", provider.machineConfigs[controlPlaneMachineConfigName].Spec.OSFamily, v1alpha1.Bottlerocket)
	}
	if provider.machineConfigs[workerNodeMachineConfigName].Spec.OSFamily != v1alpha1.Bottlerocket {
		t.Fatalf("got osFamily for control plane machine as %v, want %v", provider.machineConfigs[controlPlaneMachineConfigName].Spec.OSFamily, v1alpha1.Bottlerocket)
	}
	if provider.machineConfigs[etcdMachineConfigName].Spec.OSFamily != v1alpha1.Bottlerocket {
		t.Fatalf("got osFamily for etcd machine as %v, want %v", provider.machineConfigs[etcdMachineConfigName].Spec.OSFamily, v1alpha1.Bottlerocket)
	}
}

func TestSetupAndValidateCreateClusterTemplateDifferent(t *testing.T) {
	ctx := context.Background()
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Template = "test"
	var tctx testContext
	tctx.SaveContext()

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		thenErrorExpected(t, "control plane and worker nodes must have the same template specified", err)
	}
}

func TestSetupAndValidateCreateClusterTemplateDoesNotExist(t *testing.T) {
	tt := newProviderTest(t)

	tt.setExpectationForSetup()
	tt.setExpectationForVCenterValidation()
	for _, mc := range tt.machineConfigs {
		tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, mc).Return("", nil).MaxTimes(1)
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
		tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, mc).Return("", errors.New(errorMessage)).MaxTimes(1)
	}

	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)

	thenErrorExpected(t, "failed setting default values for vsphere machine configs: setting template full path: "+errorMessage, err)
}

func TestSetupAndValidateCreateClusterTemplateMissingTags(t *testing.T) {
	tt := newProviderTest(t)

	tt.setExpectationForSetup()
	tt.setExpectationsForDefaultDiskGovcCalls()
	tt.setExpectationForVCenterValidation()
	tt.setExpectationsForMachineConfigsVCenterValidation()

	for _, mc := range tt.machineConfigs {
		tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, mc).Return(mc.Spec.Template, nil)
	}
	controlPlaneMachineConfigName := tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	controlPlaneMachineConfig := tt.machineConfigs[controlPlaneMachineConfigName]

	tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, controlPlaneMachineConfig).Return(controlPlaneMachineConfig.Spec.Template, nil)
	tt.govc.EXPECT().GetTags(tt.ctx, controlPlaneMachineConfig.Spec.Template).Return(nil, nil)

	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)

	thenErrorPrefixExpected(t, "template "+testTemplate+" is missing tag ", err)
}

func TestSetupAndValidateCreateClusterErrorGettingTags(t *testing.T) {
	tt := newProviderTest(t)
	errorMessage := "failed getting tags"

	controlPlaneMachineConfigName := tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	controlPlaneMachineConfig := tt.machineConfigs[controlPlaneMachineConfigName]

	tt.setExpectationForSetup()
	tt.setExpectationsForDefaultDiskGovcCalls()
	tt.setExpectationForVCenterValidation()
	tt.setExpectationsForMachineConfigsVCenterValidation()
	for _, mc := range tt.machineConfigs {
		tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, mc).Return(mc.Spec.Template, nil)
	}
	tt.govc.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, controlPlaneMachineConfig).Return(controlPlaneMachineConfig.Spec.Template, nil)
	tt.govc.EXPECT().GetTags(tt.ctx, controlPlaneMachineConfig.Spec.Template).Return(nil, errors.New(errorMessage))

	err := tt.provider.SetupAndValidateCreateCluster(tt.ctx, tt.clusterSpec)

	thenErrorExpected(t, "validating template tags: failed getting tags", err)
}

func TestSetupAndValidateCreateClusterDefaultTemplate(t *testing.T) {
	ctx := context.Background()
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.EksD.Ova.Ubuntu.URI = "https://amazonaws.com/artifacts/0.0.1/eks-distro/ova/1-19/1-19-4/ubuntu-v1.19.8-eks-d-1-19-4-eks-a-0.0.1.build.38-amd64.ova"
		s.VersionsBundle.EksD.Ova.Ubuntu.SHA256 = "63a8dce1683379cb8df7d15e9c5adf9462a2b9803a544dd79b16f19a4657967f"
		s.VersionsBundle.EksD.Ova.Ubuntu.Arch = []string{"amd64"}
		s.VersionsBundle.EksD.Name = eksd119Release
		s.VersionsBundle.EksD.KubeVersion = "v1.19.8"
		s.VersionsBundle.KubeVersion = "1.19"
		s.Cluster.Namespace = "test-namespace"
	})
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	provider := givenProvider(t)
	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	provider.machineConfigs[controlPlaneMachineConfigName].Spec.Template = ""
	workerNodeMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	provider.machineConfigs[workerNodeMachineConfigName].Spec.Template = ""
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	provider.machineConfigs[etcdMachineConfigName].Spec.Template = ""
	wantTemplate := "/SDDC-Datacenter/vm/Templates/ubuntu-v1.19.8-kubernetes-1-19-eks-4-amd64-63a8dce"
	setupContext(t)

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("provider.SetupAndValidateCreateCluster() err = %v, want err = nil", err)
	}
	gotTemplate := provider.machineConfigs[controlPlaneMachineConfigName].Spec.Template
	if gotTemplate != wantTemplate {
		t.Fatalf("provider.SetupAndValidateCreateCluster() template = %s, want %s", gotTemplate, wantTemplate)
	}
	gotTemplate = provider.machineConfigs[workerNodeMachineConfigName].Spec.Template
	if gotTemplate != wantTemplate {
		t.Fatalf("provider.SetupAndValidateCreateCluster() template = %s, want %s", gotTemplate, wantTemplate)
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
				s.VersionsBundle.VSphere = releasev1alpha1.VSphereBundle{
					Version: "v0.7.8",
					ClusterAPIController: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api-provider-vsphere/release/manager:v0.7.8-35f54b0a7ff0f4f3cb0b8e30a0650acd0e55496a",
					},
					Manager: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/kubernetes/cloud-provider-vsphere/cpi/manager:v1.18.1-2093eaeda5a4567f0e516d652e0b25b1d7abc774",
					},
					KubeVip: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/plunder-app/kube-vip:v0.3.2-2093eaeda5a4567f0e516d652e0b25b1d7abc774",
					},
					Driver: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/vsphere-csi-driver/csi/driver:v2.2.0-7c2690c880c6521afdd9ffa8d90443a11c6b817b",
					},
					Syncer: releasev1alpha1.Image{
						URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/vsphere-csi-driver/csi/syncer:v2.2.0-7c2690c880c6521afdd9ffa8d90443a11c6b817b",
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
			wantManifests := []releasev1alpha1.Manifest{
				tt.clusterSpec.VersionsBundle.VSphere.Components,
				tt.clusterSpec.VersionsBundle.VSphere.Metadata,
				tt.clusterSpec.VersionsBundle.VSphere.ClusterTemplate,
			}
			assert.ElementsMatch(t, infraBundle.Manifests, wantManifests, "Incorrect manifests")
		})
	}
}

func TestGetDatacenterConfig(t *testing.T) {
	provider := givenProvider(t)
	provider.datacenterConfig.TypeMeta.Kind = "kind"

	providerConfig := provider.DatacenterConfig(givenEmptyClusterSpec())
	if providerConfig.Kind() != "kind" {
		t.Fatal("Unexpected error DatacenterConfig: kind field not found")
	}
}

func TestValidateNewSpecSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	clusterVsphereSecret := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Data: map[string][]byte{
			"username": []byte("vsphere_username"),
			"password": []byte("vsphere_password"),
		},
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})
	c := &types.Cluster{}

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)
	for _, config := range newMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterConfig.Namespace).Return(config, nil)
	}
	kubectl.EXPECT().GetSecret(gomock.Any(), CredentialsObjectName, gomock.Any()).Return(clusterVsphereSecret, nil)

	err := provider.ValidateNewSpec(context.TODO(), c, clusterSpec)
	assert.NoError(t, err, "No error should be returned when previous spec == new spec")
}

func TestValidateNewSpecMutableFields(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newDatacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	for _, config := range newMachineConfigs {
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

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(newDatacenterConfig, nil)
	for _, config := range newMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterConfig.Namespace).Return(config, nil)
	}
	kubectl.EXPECT().GetSecret(gomock.Any(), CredentialsObjectName, gomock.Any(), gomock.Any()).Return(clusterVsphereSecret, nil)

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.NoError(t, err, "No error should be returned when modifying mutable fields")
}

func TestValidateNewSpecDatacenterImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newProviderConfig.Spec.Datacenter = "new-" + newProviderConfig.Spec.Datacenter

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)
	for _, config := range newMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterConfig.Namespace).Return(config, nil)
	}

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Error(t, err, "Datacenter should be immutable")
}

func TestValidateNewSpecMachineConfigNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newProviderConfig.Spec.Datacenter = "new-" + newProviderConfig.Spec.Datacenter

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Namespace = "test-namespace"
		s.Cluster = clusterConfig.DeepCopy()
		s.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "missing-machine-group"
		s.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name = "missing-machine-group"
		s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "missing-machine-group"
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Errorf(t, err, "can't find machine config missing-machine-group in vsphere provider machine configs")
}

func TestValidateNewSpecServerImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newProviderConfig.Spec.Server = "new-" + newProviderConfig.Spec.Server

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)
	for _, config := range newMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterConfig.Namespace).Return(config, nil)
	}

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Error(t, err, "Server should be immutable")
}

func TestValidateNewSpecStoragePolicyNameImmutableControlPlane(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	controlPlaneMachineConfigName := clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	newMachineConfigs[controlPlaneMachineConfigName].Spec.StoragePolicyName = "new-" + newMachineConfigs[controlPlaneMachineConfigName].Spec.StoragePolicyName

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterConfig.Namespace).Return(newMachineConfigs[controlPlaneMachineConfigName], nil)

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Error(t, err, "StoragePolicyName should be immutable")
}

func TestValidateNewSpecStoragePolicyNameImmutableWorker(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	workerMachineConfigName := clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	newMachineConfigs[workerMachineConfigName].Spec.StoragePolicyName = "new-" + newMachineConfigs[workerMachineConfigName].Spec.StoragePolicyName

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)
	kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterConfig.Namespace).Return(newMachineConfigs[workerMachineConfigName], nil)

	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Error(t, err, "StoragePolicyName should be immutable")
}

func TestValidateNewSpecTLSInsecureImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newProviderConfig.Spec.Insecure = !newProviderConfig.Spec.Insecure

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)
	for _, config := range newMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterConfig.Namespace).Return(config, nil)
	}
	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Error(t, err, "Insecure should be immutable")
}

func TestValidateNewSpecTLSThumbprintImmutable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	clusterConfig := givenClusterConfig(t, testClusterConfigMainFilename)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	newProviderConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newProviderConfig.Spec.Thumbprint = "new-" + newProviderConfig.Spec.Thumbprint

	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Namespace = "test-namespace"
		s.Cluster = clusterConfig
	})

	kubectl.EXPECT().GetEksaCluster(context.TODO(), gomock.Any(), gomock.Any()).Return(clusterConfig, nil)
	kubectl.EXPECT().GetEksaVSphereDatacenterConfig(context.TODO(), clusterConfig.Spec.DatacenterRef.Name, gomock.Any(), clusterConfig.Namespace).Return(newProviderConfig, nil)
	for _, config := range newMachineConfigs {
		kubectl.EXPECT().GetEksaVSphereMachineConfig(context.TODO(), gomock.Any(), gomock.Any(), clusterConfig.Namespace).Return(config, nil)
	}
	err := provider.ValidateNewSpec(context.TODO(), &types.Cluster{}, clusterSpec)
	assert.Error(t, err, "Thumbprint should be immutable")
}

func TestGetMHCSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	provider := givenProvider(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	mhcTemplate := fmt.Sprintf(`apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineHealthCheck
metadata:
  name: test-node-unhealthy-5m
  namespace: %[1]s
spec:
  clusterName: test
  maxUnhealthy: 40%%
  nodeStartupTimeout: 10m
  selector:
    matchLabels:
      cluster.x-k8s.io/deployment-name: "test-md-0"
  unhealthyConditions:
    - type: Ready
      status: Unknown
      timeout: 300s
    - type: Ready
      status: "False"
      timeout: 300s
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineHealthCheck
metadata:
  name: test-kcp-unhealthy-5m
  namespace: %[1]s
spec:
  clusterName: test
  maxUnhealthy: 100%%
  selector:
    matchLabels:
      cluster.x-k8s.io/control-plane: ""
  unhealthyConditions:
    - type: Ready
      status: Unknown
      timeout: 300s
    - type: Ready
      status: "False"
      timeout: 300s
`, constants.EksaSystemNamespace)

	mch, err := provider.GenerateMHC()
	assert.NoError(t, err, "Expected successful execution of GenerateMHC() but got an error", "error", err)
	assert.Equal(t, string(mch), mhcTemplate, "generated MachineHealthCheck is different from the expected one")
}

func TestChangeDiffNoChange(t *testing.T) {
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	assert.Nil(t, provider.ChangeDiff(clusterSpec, clusterSpec))
}

func TestChangeDiffWithChange(t *testing.T) {
	provider := givenProvider(t)
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.VSphere.Version = "v0.3.18"
	})
	newClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.VSphere.Version = "v0.3.19"
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

	tt.resourceSetManager.EXPECT().ForceUpdate(tt.ctx, "test-crs-0", "eksa-system", tt.managementCluster, tt.workloadCluster)
	tt.Expect(tt.provider.RunPostControlPlaneUpgrade(tt.ctx, tt.clusterSpec, tt.clusterSpec, tt.workloadCluster, tt.managementCluster)).To(Succeed())
}

func TestProviderUpgradeNeeded(t *testing.T) {
	testCases := []struct {
		testName               string
		newSyncer, oldSyncer   string
		newDriver, oldDriver   string
		newManager, oldManager string
		newKubeVip, oldKubeVip string
		want                   bool
	}{
		{
			testName:  "different syncer",
			newSyncer: "a", oldSyncer: "b",
			want: true,
		},
		{
			testName:  "different driver",
			newDriver: "a", oldDriver: "b",
			want: true,
		},
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
				s.VersionsBundle.VSphere.Syncer.ImageDigest = tt.oldSyncer
				s.VersionsBundle.VSphere.Driver.ImageDigest = tt.oldDriver
				s.VersionsBundle.VSphere.Manager.ImageDigest = tt.oldManager
				s.VersionsBundle.VSphere.KubeVip.ImageDigest = tt.oldKubeVip
			})

			newClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.VersionsBundle.VSphere.Syncer.ImageDigest = tt.newSyncer
				s.VersionsBundle.VSphere.Driver.ImageDigest = tt.newDriver
				s.VersionsBundle.VSphere.Manager.ImageDigest = tt.newManager
				s.VersionsBundle.VSphere.KubeVip.ImageDigest = tt.newKubeVip
			})

			g := NewWithT(t)
			g.Expect(provider.UpgradeNeeded(context.Background(), clusterSpec, newClusterSpec)).To(Equal(tt.want))
		})
	}
}

func TestProviderGenerateCAPISpecForCreateWithPodIAMConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.Cluster.Spec.PodIAMConfig = &v1alpha1.PodIAMConfig{ServiceAccountIssuer: "https://test"}

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

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

func TestProviderGenerateCAPISpecForCreateWithCustomResolvConf(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMainFilename)
	clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf = &v1alpha1.ResolvConf{Path: "/etc/my-custom-resolv.conf"}

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)
	if provider == nil {
		t.Fatalf("provider object is nil")
	}

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
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMain121Filename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMain121Filename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMain121Filename)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)
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
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_main_121_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_121_md.yaml")
}

func TestProviderGenerateCAPISpecForCreateVersion121Controller(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()
	ctx := context.Background()
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	cluster := &types.Cluster{
		Name: "test",
	}
	clusterSpec := givenClusterSpec(t, testClusterConfigMain121Filename)

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMain121Filename)
	machineConfigs := givenMachineConfigs(t, testClusterConfigMain121Filename)
	provider := newProviderWithKubectl(t, datacenterConfig, machineConfigs, clusterSpec.Cluster, kubectl)

	if provider == nil {
		t.Fatalf("provider object is nil")
	}
	provider.templateBuilder.fromController = true
	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), cluster, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/expected_results_main_121_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_main_121_controller_md.yaml")
}

func TestSetupAndValidateCreateManagementDoesNotCheckIfMachineAndDataCenterExist(t *testing.T) {
	ctx := context.Background()
	provider := givenProvider(t)
	clusterSpec := givenEmptyClusterSpec()
	fillClusterSpecWithClusterConfig(clusterSpec, givenClusterConfig(t, testClusterConfigMainFilename))
	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	datacenterConfig := givenDatacenterConfig(t, testClusterConfigMainFilename)
	newMachineConfigs := givenMachineConfigs(t, testClusterConfigMainFilename)

	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockProviderKubectlClient(mockCtrl)
	provider.providerKubectlClient = kubectl

	for _, config := range newMachineConfigs {
		kubectl.EXPECT().SearchVsphereMachineConfig(context.TODO(), config.Name, gomock.Any(), config.Namespace).Return([]*v1alpha1.VSphereMachineConfig{}, nil).Times(0)
	}
	kubectl.EXPECT().SearchVsphereDatacenterConfig(context.TODO(), datacenterConfig.Name, gomock.Any(), clusterSpec.Cluster.Namespace).Return([]*v1alpha1.VSphereDatacenterConfig{}, nil).Times(0)

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("unexpected failure %v", err)
	}
	assert.NoError(t, err, "No error should be returned")
}

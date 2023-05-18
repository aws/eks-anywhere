package upgradevalidations_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/version"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	filewritermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	mockproviders "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	tinkerbellmocks "github.com/aws/eks-anywhere/pkg/providers/tinkerbell/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	stackmocks "github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	kubeconfigFilePath = "./fakeKubeconfigFilePath"
)

var goodClusterResponse = []types.CAPICluster{{Metadata: types.Metadata{Name: testclustername}}}

func TestPreflightValidationsTinkerbell(t *testing.T) {
	tests := []struct {
		name                    string
		clusterVersion          string
		upgradeVersion          string
		getClusterResponse      []types.CAPICluster
		cpResponse              error
		workerResponse          error
		nodeResponse            error
		crdResponse             error
		wantErr                 error
		modifyFunc              func(s *cluster.Spec)
		modifyDatacenterFunc    func(s *v1alpha1.TinkerbellDatacenterConfig)
		modifyMachineConfigFunc func(s *v1alpha1.TinkerbellMachineConfig)
	}{
		{
			name:               "ValidationSucceeds",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            nil,
		},
		{
			name:               "ValidationFailsMajorVersionPlus2",
			clusterVersion:     "v1.18.16-eks-1-18-4",
			upgradeVersion:     "1.20",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("WARNING: version difference between upgrade version (1.20) and server version (1.18) do not meet the supported version increment of +1"),
		},
		{
			name:               "ValidationFailsMajorVersionMinus1",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.18",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("WARNING: version difference between upgrade version (1.18) and server version (1.19) do not meet the supported version increment of +1"),
		},
		{
			name:               "ValidationFailsClusterDoesNotExist",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: []types.CAPICluster{{Metadata: types.Metadata{Name: "thisIsNotTheClusterYourLookingFor"}}},
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("couldn't find CAPI cluster object for cluster with name testcluster"),
		},
		{
			name:               "ValidationFailsNoClusters",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: []types.CAPICluster{},
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("no CAPI cluster objects present on workload cluster testcluster"),
		},
		{
			name:               "ValidationFailsCpNotReady",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         errors.New("control plane nodes are not ready"),
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("control plane nodes are not ready"),
		},
		{
			name:               "ValidationFailsWorkerNodesNotReady",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     errors.New("2 worker nodes are not ready"),
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("2 worker nodes are not ready"),
		},
		{
			name:               "ValidationFailsNodesNotReady",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       errors.New("node test-node is not ready, currently in Unknown state"),
			crdResponse:        nil,
			wantErr:            composeError("node test-node is not ready, currently in Unknown state"),
		},
		{
			name:               "ValidationFailsNoCrds",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        errors.New("error getting clusters crd: crd not found"),
			wantErr:            composeError("error getting clusters crd: crd not found"),
		},
		{
			name:               "ValidationFailsExplodingCluster",
			clusterVersion:     "v1.18.16-eks-1-18-4",
			upgradeVersion:     "1.20",
			getClusterResponse: []types.CAPICluster{{Metadata: types.Metadata{Name: "thisIsNotTheClusterYourLookingFor"}}},
			cpResponse:         errors.New("control plane nodes are not ready"),
			workerResponse:     errors.New("2 worker nodes are not ready"),
			nodeResponse:       errors.New("node test-node is not ready, currently in Unknown state"),
			crdResponse:        errors.New("error getting clusters crd: crd not found"),
			wantErr:            explodingClusterError,
		},
		{
			name:               "ValidationControlPlaneImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.controlPlaneConfiguration.endpoint is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "2.3.4.5"
			},
		},
		{
			name:               "ValidationClusterNetworkPodsImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.clusterNetwork.Pods is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ClusterNetwork.Pods = v1alpha1.Pods{}
			},
		},
		{
			name:               "ValidationClusterNetworkServicesImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.clusterNetwork.Services is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ClusterNetwork.Services = v1alpha1.Services{}
			},
		},
		{
			name:               "ValidationManagementImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("management flag is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.Cluster.SetManagedBy(fmt.Sprintf("%s-1", s.Cluster.ManagedBy()))
			},
		},
		{
			name:               "ValidationTinkerbellIPImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.TinkerbellIP is immutable. Previous value 4.5.6.7,   New value 1.2.3.4"),
			modifyDatacenterFunc: func(s *v1alpha1.TinkerbellDatacenterConfig) {
				s.Spec.TinkerbellIP = "4.5.6.7"
			},
		},
		{
			name:               "ValidationOSImageURLImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.OSImageURL is immutable. Previous value http://old-os-image-url,   New value http://os-image-url"),
			modifyDatacenterFunc: func(s *v1alpha1.TinkerbellDatacenterConfig) {
				s.Spec.OSImageURL = "http://old-os-image-url"
			},
		},
		{
			name:               "ValidationHookImageURLImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.HookImagesURLPath is immutable. Previous value http://old-hook-image-url,   New value http://hook-image-url"),
			modifyDatacenterFunc: func(s *v1alpha1.TinkerbellDatacenterConfig) {
				s.Spec.HookImagesURLPath = "http://old-hook-image-url"
			},
		},
		{
			name:               "ValidationSSHUsernameImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.Users[0].Name is immutable. Previous value myOldSshUsername,   New value mySshUsername"),
			modifyMachineConfigFunc: func(s *v1alpha1.TinkerbellMachineConfig) {
				s.Spec.Users[0].Name = "myOldSshUsername"
			},
		},
		{
			name:               "ValidationSSHAuthorizedKeysImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.Users[0].SshAuthorizedKeys[0] is immutable. Previous value myOldSshAuthorizedKeys,   New value mySshAuthorizedKey"),
			modifyMachineConfigFunc: func(s *v1alpha1.TinkerbellMachineConfig) {
				s.Spec.Users[0].SshAuthorizedKeys[0] = "myOldSshAuthorizedKeys"
			},
		},
		{
			name:               "ValidationHardwareSelectorImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.HardwareSelector is immutable. Previous value map[type:cp1],   New value map[type:cp]"),
			modifyMachineConfigFunc: func(s *v1alpha1.TinkerbellMachineConfig) {
				s.Spec.HardwareSelector = map[string]string{
					"type": "cp1",
				}
			},
		},
	}

	defaultControlPlane := v1alpha1.ControlPlaneConfiguration{
		Count: 1,
		Endpoint: &v1alpha1.Endpoint{
			Host: "1.1.1.1",
		},
		MachineGroupRef: &v1alpha1.Ref{
			Name: "test-cp",
			Kind: "TinkerbellMachineConfig",
		},
	}

	defaultDatacenterSpec := v1alpha1.TinkerbellDatacenterConfig{
		Spec: v1alpha1.TinkerbellDatacenterConfigSpec{
			TinkerbellIP:      "1.2.3.4",
			OSImageURL:        "http://os-image-url",
			HookImagesURLPath: "http://hook-image-url",
		},
		Status: v1alpha1.TinkerbellDatacenterConfigStatus{},
	}

	defaultTinkerbellMachineConfigSpec := v1alpha1.TinkerbellMachineConfig{
		Spec: v1alpha1.TinkerbellMachineConfigSpec{
			HardwareSelector: map[string]string{
				"type": "cp",
			},
			OSFamily: "ubuntu",
			Users: []v1alpha1.UserConfiguration{{
				Name:              "mySshUsername",
				SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
			}},
		},
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = testclustername
		s.Cluster.Spec.ControlPlaneConfiguration = defaultControlPlane
		s.Cluster.Spec.DatacenterRef = v1alpha1.Ref{
			Kind: v1alpha1.TinkerbellDatacenterKind,
			Name: "tinkerbell test",
		}
		s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
			Pods: v1alpha1.Pods{
				CidrBlocks: []string{
					"1.2.3.4/5",
				},
			},
			Services: v1alpha1.Services{
				CidrBlocks: []string{
					"1.2.3.4/6",
				},
			},
		}
	})

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			_, ctx, workloadCluster, _ := validations.NewKubectl(t)
			workloadCluster.KubeconfigFile = kubeconfigFilePath
			workloadCluster.Name = testclustername

			mockCtrl := gomock.NewController(t)
			k := mocks.NewMockKubectlClient(mockCtrl)
			kubectl := tinkerbellmocks.NewMockProviderKubectlClient(mockCtrl)
			docker := stackmocks.NewMockDocker(mockCtrl)
			helm := stackmocks.NewMockHelm(mockCtrl)
			writer := filewritermocks.NewMockFileWriter(mockCtrl)
			tlsValidator := mocks.NewMockTlsValidator(mockCtrl)

			provider := newProvider(defaultDatacenterSpec, givenTinkerbellMachineConfigs(t), clusterSpec.Cluster, writer, docker, helm, kubectl, false)
			// provider := mockproviders.NewMockProvider(mockCtrl)
			opts := &validations.Opts{
				Kubectl:           k,
				Spec:              clusterSpec,
				WorkloadCluster:   workloadCluster,
				ManagementCluster: workloadCluster,
				Provider:          provider,
				TLSValidator:      tlsValidator,
			}

			clusterSpec.Cluster.Spec.KubernetesVersion = v1alpha1.KubernetesVersion(tc.upgradeVersion)
			existingClusterSpec := clusterSpec.DeepCopy()
			existingProviderSpec := defaultDatacenterSpec.DeepCopy()
			existingMachineConfigSpec := defaultTinkerbellMachineConfigSpec.DeepCopy()
			// existingMachineConfigSpec := defaultTinkerbellMachineConfigSpec.DeepCopy()
			if tc.modifyFunc != nil {
				tc.modifyFunc(existingClusterSpec)
			}
			if tc.modifyDatacenterFunc != nil {
				tc.modifyDatacenterFunc(existingProviderSpec)
			}
			if tc.modifyMachineConfigFunc != nil {
				tc.modifyMachineConfigFunc(existingMachineConfigSpec)
			}
			versionResponse := &executables.VersionResponse{
				ServerVersion: version.Info{
					GitVersion: tc.clusterVersion,
				},
			}

			kubectl.EXPECT().GetEksaCluster(ctx, workloadCluster, clusterSpec.Cluster.Name).Return(existingClusterSpec.Cluster, nil).MaxTimes(1)
			// provider.EXPECT().DatacenterConfig(clusterSpec).Return(existingProviderSpec).MaxTimes(1)
			// provider.EXPECT().MachineConfigs(clusterSpec).Return(existingMachineConfigSpec).MaxTimes(1)
			// provider.EXPECT().ValidateNewSpec(ctx, workloadCluster, clusterSpec).Return(nil).MaxTimes(1)
			kubectl.EXPECT().GetEksaTinkerbellDatacenterConfig(ctx, clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(existingProviderSpec, nil).MaxTimes(1)
			kubectl.EXPECT().GetEksaTinkerbellMachineConfig(ctx, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any(), gomock.Any()).Return(existingMachineConfigSpec, nil).MaxTimes(1)
			k.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			k.EXPECT().ValidateControlPlaneNodes(ctx, workloadCluster, clusterSpec.Cluster.Name).Return(tc.cpResponse)
			k.EXPECT().ValidateWorkerNodes(ctx, workloadCluster.Name, workloadCluster.KubeconfigFile).Return(tc.workerResponse)
			k.EXPECT().ValidateNodes(ctx, kubeconfigFilePath).Return(tc.nodeResponse)
			k.EXPECT().ValidateClustersCRD(ctx, workloadCluster).Return(tc.crdResponse)
			k.EXPECT().GetClusters(ctx, workloadCluster).Return(tc.getClusterResponse, nil)
			k.EXPECT().GetEksaCluster(ctx, workloadCluster, clusterSpec.Cluster.Name).Return(existingClusterSpec.Cluster, nil)
			k.EXPECT().Version(ctx, workloadCluster).Return(versionResponse, nil)
			upgradeValidations := upgradevalidations.New(opts)
			err := validations.ProcessValidationResults(upgradeValidations.PreflightValidations(ctx))
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%s want err=%v\n got err=%v\n", tc.name, tc.wantErr, err)
			}
		})
	}
}

func givenTinkerbellMachineConfigs(t *testing.T) map[string]*v1alpha1.TinkerbellMachineConfig {
	machineConfigs, err := v1alpha1.GetTinkerbellMachineConfigs("./testdata/tinkerbell_clusterconfig.yaml")
	if err != nil {
		t.Fatalf("unable to get machine configs from file: %v", err)
	}
	return machineConfigs
}

func newProvider(datacenterConfig v1alpha1.TinkerbellDatacenterConfig, machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig, clusterConfig *v1alpha1.Cluster, writer filewriter.FileWriter, docker stack.Docker, helm stack.Helm, kubectl tinkerbell.ProviderKubectlClient, forceCleanup bool) *tinkerbell.Provider {
	hardwareFile := "./testdata/hardware.csv"
	provider, err := tinkerbell.NewProvider(
		&datacenterConfig,
		machineConfigs,
		clusterConfig,
		hardwareFile,
		writer,
		docker,
		helm,
		kubectl,
		"1.2.3.4",
		test.FakeNow,
		forceCleanup,
		false,
	)
	if err != nil {
		panic(err)
	}

	return provider
}

func TestPreflightValidationsVsphere(t *testing.T) {
	tests := []struct {
		name                   string
		clusterVersion         string
		upgradeVersion         string
		getClusterResponse     []types.CAPICluster
		cpResponse             error
		workerResponse         error
		nodeResponse           error
		crdResponse            error
		wantErr                error
		modifyExistingSpecFunc func(s *cluster.Spec)
		modifyDefaultSpecFunc  func(s *cluster.Spec)
		additionalKubectlMocks func(k *mocks.MockKubectlClient)
	}{
		{
			name:               "ValidationSucceeds",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            nil,
		},
		{
			name:               "ValidationFailsMajorVersionPlus2",
			clusterVersion:     "v1.18.16-eks-1-18-4",
			upgradeVersion:     "1.20",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("WARNING: version difference between upgrade version (1.20) and server version (1.18) do not meet the supported version increment of +1"),
		},
		{
			name:               "ValidationFailsMajorVersionMinus1",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.18",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("WARNING: version difference between upgrade version (1.18) and server version (1.19) do not meet the supported version increment of +1"),
		},
		{
			name:               "ValidationFailsClusterDoesNotExist",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: []types.CAPICluster{{Metadata: types.Metadata{Name: "thisIsNotTheClusterYourLookingFor"}}},
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("couldn't find CAPI cluster object for cluster with name testcluster"),
		},
		{
			name:               "ValidationFailsNoClusters",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: []types.CAPICluster{},
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("no CAPI cluster objects present on workload cluster testcluster"),
		},
		{
			name:               "ValidationFailsCpNotReady",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         errors.New("control plane nodes are not ready"),
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("control plane nodes are not ready"),
		},
		{
			name:               "ValidationFailsWorkerNodesNotReady",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     errors.New("2 worker nodes are not ready"),
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("2 worker nodes are not ready"),
		},
		{
			name:               "ValidationFailsNodesNotReady",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       errors.New("node test-node is not ready, currently in Unknown state"),
			crdResponse:        nil,
			wantErr:            composeError("node test-node is not ready, currently in Unknown state"),
		},
		{
			name:               "ValidationFailsNoCrds",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        errors.New("error getting clusters crd: crd not found"),
			wantErr:            composeError("error getting clusters crd: crd not found"),
		},
		{
			name:               "ValidationFailsExplodingCluster",
			clusterVersion:     "v1.18.16-eks-1-18-4",
			upgradeVersion:     "1.20",
			getClusterResponse: []types.CAPICluster{{Metadata: types.Metadata{Name: "thisIsNotTheClusterYourLookingFor"}}},
			cpResponse:         errors.New("control plane nodes are not ready"),
			workerResponse:     errors.New("2 worker nodes are not ready"),
			nodeResponse:       errors.New("node test-node is not ready, currently in Unknown state"),
			crdResponse:        errors.New("error getting clusters crd: crd not found"),
			wantErr:            explodingClusterError,
		},
		{
			name:               "ValidationEtcdImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.externalEtcdConfiguration.count is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ExternalEtcdConfiguration.Count++
			},
		},
		{
			name:               "ValidationControlPlaneImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.controlPlaneConfiguration.endpoint is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "2.3.4.5"
			},
		},
		{
			name:               "ValidationAwsIamRegionImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("aws iam identity provider is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.AWSIamConfig.Spec.AWSRegion = "us-east-2"
			},
		},
		{
			name:               "ValidationAwsIamBackEndModeImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("aws iam identity provider is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.AWSIamConfig.Spec.BackendMode = append(s.AWSIamConfig.Spec.BackendMode, "us-east-2")
			},
		},
		{
			name:               "ValidationAwsIamPartitionImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("aws iam identity provider is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.AWSIamConfig.Spec.Partition = "partition2"
			},
		},
		{
			name:               "ValidationAwsIamNameImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("aws iam identity provider is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.IdentityProviderRefs[1] = v1alpha1.Ref{
					Kind: v1alpha1.AWSIamConfigKind,
					Name: "aws-iam2",
				}
			},
		},
		{
			name:               "ValidationAwsIamKindImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("aws iam identity provider is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.IdentityProviderRefs[0] = v1alpha1.Ref{
					Kind: v1alpha1.OIDCConfigKind,
					Name: "oidc",
				}
			},
		},
		{
			name:               "ValidationAwsIamKindImmutableSwapOrder",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            nil,
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.IdentityProviderRefs[1] = v1alpha1.Ref{
					Kind: v1alpha1.AWSIamConfigKind,
					Name: "aws-iam",
				}
				s.Cluster.Spec.IdentityProviderRefs[0] = v1alpha1.Ref{
					Kind: v1alpha1.OIDCConfigKind,
					Name: "oidc",
				}
			},
		},
		{
			name:               "ValidationGitOpsNamespaceImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("gitOps spec.flux.github.fluxSystemNamespace is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.GitOpsConfig.Spec.Flux.Github.FluxSystemNamespace = "new-namespace"
			},
		},
		{
			name:               "ValidationGitOpsBranchImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("gitOps spec.flux.github.branch is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.GitOpsConfig.Spec.Flux.Github.Branch = "new-branch"
			},
		},
		{
			name:               "ValidationGitOpsOwnerImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("gitOps spec.flux.github.owner is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.GitOpsConfig.Spec.Flux.Github.Owner = "new-owner"
			},
		},
		{
			name:               "ValidationGitOpsRepositoryImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("gitOps spec.flux.github.repository is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.GitOpsConfig.Spec.Flux.Github.Repository = "new-repository"
			},
		},
		{
			name:               "ValidationGitOpsPathImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("gitOps spec.flux.github.clusterConfigPath is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.GitOpsConfig.Spec.Flux.Github.ClusterConfigPath = "new-path"
			},
		},
		{
			name:               "ValidationGitOpsPersonalImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("gitOps spec.flux.github.personal is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.GitOpsConfig.Spec.Flux.Github.Personal = !s.GitOpsConfig.Spec.Flux.Github.Personal
			},
		},
		{
			name:               "ValidationOIDCClientIdMutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            nil,
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.ClientId = "new-client-id"
			},
		},
		{
			name:               "ValidationOIDCGroupsClaimMutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            nil,
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.GroupsClaim = "new-groups-claim"
			},
		},
		{
			name:               "ValidationOIDCGroupsPrefixMutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            nil,
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.GroupsPrefix = "new-groups-prefix"
			},
		},
		{
			name:               "ValidationOIDCIssuerUrlMutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            nil,
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.IssuerUrl = "new-issuer-url"
			},
		},
		{
			name:               "ValidationOIDCUsernameClaimMutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            nil,
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.UsernameClaim = "new-username-claim"
			},
		},
		{
			name:               "ValidationOIDCUsernamePrefixMutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            nil,
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.UsernamePrefix = "new-username-prefix"
			},
		},
		{
			name:               "ValidationOIDCRequiredClaimsMutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            nil,
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.RequiredClaims[0].Claim = "new-groups-claim"
			},
		},
		{
			name:               "ValidationClusterNetworkPodsImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.clusterNetwork.Pods is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ClusterNetwork.Pods = v1alpha1.Pods{}
			},
		},
		{
			name:               "ValidationClusterNetworkServicesImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.clusterNetwork.Services is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ClusterNetwork.Services = v1alpha1.Services{}
			},
		},
		{
			name:               "ValidationClusterNetworkDNSImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.clusterNetwork.DNS is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ClusterNetwork.DNS = v1alpha1.DNS{}
			},
		},
		{
			name:               "ValidationProxyConfigurationImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.proxyConfiguration is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
					HttpProxy:  "httpproxy2",
					HttpsProxy: "httpsproxy2",
					NoProxy: []string{
						"noproxy3",
					},
				}
			},
		},
		{
			name:               "ValidationEtcdConfigReplicasImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.externalEtcdConfiguration.count is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ExternalEtcdConfiguration.Count += 1
				s.Cluster.Spec.DatacenterRef = v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
					Name: "vsphere test",
				}
			},
		},
		{
			name:               "ValidationEtcdConfigPreviousSpecEmpty",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("adding or removing external etcd during upgrade is not supported"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ExternalEtcdConfiguration = nil
				s.Cluster.Spec.DatacenterRef = v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
					Name: "vsphere test",
				}
			},
		},
		{
			name:               "ValidationManagementImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("management flag is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.SetManagedBy(fmt.Sprintf("%s-1", s.Cluster.ManagedBy()))
			},
		},
		{
			name:               "ValidationManagementClusterNameImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("management cluster name is immutable"),
			modifyExistingSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ManagementCluster.Name = fmt.Sprintf("%s-1", s.Cluster.ManagedBy())
			},
			modifyDefaultSpecFunc: func(s *cluster.Spec) {
				s.Cluster.Spec.ManagementCluster.Name = fmt.Sprintf("%s-2", s.Cluster.ManagedBy())
			},
		},
	}

	defaultControlPlane := v1alpha1.ControlPlaneConfiguration{
		Count: 1,
		Endpoint: &v1alpha1.Endpoint{
			Host: "1.1.1.1",
		},
		MachineGroupRef: &v1alpha1.Ref{
			Name: "test",
			Kind: "VSphereMachineConfig",
		},
	}

	defaultETCD := &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
	}

	defaultDatacenterSpec := v1alpha1.VSphereDatacenterConfig{
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Datacenter: "datacenter!!!",
			Network:    "network",
			Server:     "server",
			Thumbprint: "thumbprint",
			Insecure:   false,
		},
		Status: v1alpha1.VSphereDatacenterConfigStatus{},
	}

	defaultGitOps := &v1alpha1.GitOpsConfig{
		Spec: v1alpha1.GitOpsConfigSpec{
			Flux: v1alpha1.Flux{
				Github: v1alpha1.Github{
					Owner:               "owner",
					Repository:          "repo",
					FluxSystemNamespace: "flux-system",
					Branch:              "main",
					ClusterConfigPath:   "clusters/" + testclustername,
					Personal:            false,
				},
			},
		},
	}

	defaultOIDC := &v1alpha1.OIDCConfig{
		Spec: v1alpha1.OIDCConfigSpec{
			ClientId:     "client-id",
			GroupsClaim:  "groups-claim",
			GroupsPrefix: "groups-prefix",
			IssuerUrl:    "issuer-url",
			RequiredClaims: []v1alpha1.OIDCConfigRequiredClaim{{
				Claim: "claim",
				Value: "value",
			}},
			UsernameClaim:  "username-claim",
			UsernamePrefix: "username-prefix",
		},
	}

	defaultAWSIAM := &v1alpha1.AWSIamConfig{
		Spec: v1alpha1.AWSIamConfigSpec{
			AWSRegion: "us-east-1",
			MapRoles: []v1alpha1.MapRoles{{
				RoleARN:  "roleARN",
				Username: "username",
				Groups:   []string{"group1", "group2"},
			}},
			MapUsers: []v1alpha1.MapUsers{{
				UserARN:  "userARN",
				Username: "username",
				Groups:   []string{"group1", "group2"},
			}},
			Partition: "partition",
		},
	}

	defaultClusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = testclustername
		s.Cluster.Spec.ControlPlaneConfiguration = defaultControlPlane
		s.Cluster.Spec.ExternalEtcdConfiguration = defaultETCD
		s.Cluster.Spec.DatacenterRef = v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: "vsphere test",
		}
		s.Cluster.Spec.IdentityProviderRefs = []v1alpha1.Ref{
			{
				Kind: v1alpha1.AWSIamConfigKind,
				Name: "aws-iam",
			},
			{
				Kind: v1alpha1.OIDCConfigKind,
				Name: "oidc",
			},
		}
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.GitOpsConfigKind,
			Name: "gitops test",
		}
		s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
			Pods: v1alpha1.Pods{
				CidrBlocks: []string{
					"1.2.3.4/5",
				},
			},
			Services: v1alpha1.Services{
				CidrBlocks: []string{
					"1.2.3.4/6",
				},
			},
			DNS: v1alpha1.DNS{
				ResolvConf: &v1alpha1.ResolvConf{Path: "file.conf"},
			},
		}
		s.Cluster.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
			HttpProxy:  "httpproxy",
			HttpsProxy: "httpsproxy",
			NoProxy: []string{
				"noproxy1",
				"noproxy2",
			},
		}
		s.Cluster.Spec.BundlesRef = &anywherev1.BundlesRef{
			Name:      "bundles-28",
			Namespace: constants.EksaSystemNamespace,
		}

		s.GitOpsConfig = defaultGitOps
		s.OIDCConfig = defaultOIDC
		s.AWSIamConfig = defaultAWSIAM
	})

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			_, ctx, workloadCluster, _ := validations.NewKubectl(t)
			workloadCluster.KubeconfigFile = kubeconfigFilePath
			workloadCluster.Name = testclustername

			mockCtrl := gomock.NewController(t)
			k := mocks.NewMockKubectlClient(mockCtrl)
			tlsValidator := mocks.NewMockTlsValidator(mockCtrl)

			provider := mockproviders.NewMockProvider(mockCtrl)

			clusterSpec := defaultClusterSpec.DeepCopy()
			if tc.modifyDefaultSpecFunc != nil {
				tc.modifyDefaultSpecFunc(clusterSpec)
			}
			opts := &validations.Opts{
				Kubectl:           k,
				Spec:              clusterSpec,
				WorkloadCluster:   workloadCluster,
				ManagementCluster: workloadCluster,
				Provider:          provider,
				TLSValidator:      tlsValidator,
			}

			clusterSpec.Cluster.Spec.KubernetesVersion = v1alpha1.KubernetesVersion(tc.upgradeVersion)
			existingClusterSpec := defaultClusterSpec.DeepCopy()
			existingProviderSpec := defaultDatacenterSpec.DeepCopy()
			if tc.modifyExistingSpecFunc != nil {
				tc.modifyExistingSpecFunc(existingClusterSpec)
			}
			versionResponse := &executables.VersionResponse{
				ServerVersion: version.Info{
					GitVersion: tc.clusterVersion,
				},
			}
			bundlesResponse := &releasev1alpha1.Bundles{
				Spec: releasev1alpha1.BundlesSpec{
					Number: 28,
				},
			}

			provider.EXPECT().DatacenterConfig(clusterSpec).Return(existingProviderSpec).MaxTimes(1)
			provider.EXPECT().ValidateNewSpec(ctx, workloadCluster, clusterSpec).Return(nil).MaxTimes(1)
			k.EXPECT().GetEksaVSphereDatacenterConfig(ctx, clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(existingProviderSpec, nil).MaxTimes(1)
			k.EXPECT().ValidateControlPlaneNodes(ctx, workloadCluster, clusterSpec.Cluster.Name).Return(tc.cpResponse)
			k.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			k.EXPECT().ValidateWorkerNodes(ctx, workloadCluster.Name, workloadCluster.KubeconfigFile).Return(tc.workerResponse)
			k.EXPECT().ValidateNodes(ctx, kubeconfigFilePath).Return(tc.nodeResponse)
			k.EXPECT().ValidateClustersCRD(ctx, workloadCluster).Return(tc.crdResponse)
			k.EXPECT().GetClusters(ctx, workloadCluster).Return(tc.getClusterResponse, nil)
			k.EXPECT().GetEksaCluster(ctx, workloadCluster, clusterSpec.Cluster.Name).Return(existingClusterSpec.Cluster, nil)
			if opts.Spec.Cluster.IsManaged() {
				k.EXPECT().GetEksaCluster(ctx, workloadCluster, workloadCluster.Name).Return(existingClusterSpec.Cluster, nil)
				k.EXPECT().GetBundles(ctx, workloadCluster.KubeconfigFile, existingClusterSpec.Cluster.Spec.BundlesRef.Name, existingClusterSpec.Cluster.Spec.BundlesRef.Namespace).Return(bundlesResponse, nil)
			}
			k.EXPECT().GetEksaGitOpsConfig(ctx, clusterSpec.Cluster.Spec.GitOpsRef.Name, gomock.Any(), gomock.Any()).Return(existingClusterSpec.GitOpsConfig, nil).MaxTimes(1)
			k.EXPECT().GetEksaOIDCConfig(ctx, clusterSpec.Cluster.Spec.IdentityProviderRefs[1].Name, gomock.Any(), gomock.Any()).Return(existingClusterSpec.OIDCConfig, nil).MaxTimes(1)
			k.EXPECT().GetEksaAWSIamConfig(ctx, clusterSpec.Cluster.Spec.IdentityProviderRefs[0].Name, gomock.Any(), gomock.Any()).Return(existingClusterSpec.AWSIamConfig, nil).MaxTimes(1)
			k.EXPECT().Version(ctx, workloadCluster).Return(versionResponse, nil)
			upgradeValidations := upgradevalidations.New(opts)
			err := validations.ProcessValidationResults(upgradeValidations.PreflightValidations(ctx))
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%s want err=%v\n got err=%v\n", tc.name, tc.wantErr, err)
			}
		})
	}
}

func composeError(msgs ...string) *validations.ValidationError {
	var errs []string
	errs = append(errs, msgs...)
	return &validations.ValidationError{Errs: errs}
}

var explodingClusterError = composeError(
	"control plane nodes are not ready",
	"2 worker nodes are not ready",
	"node test-node is not ready, currently in Unknown state",
	"error getting clusters crd: crd not found",
	"couldn't find CAPI cluster object for cluster with name testcluster",
	"WARNING: version difference between upgrade version (1.20) and server version (1.18) do not meet the supported version increment of +1",
)

func TestPreFlightValidationsGit(t *testing.T) {
	tests := []struct {
		name               string
		clusterVersion     string
		upgradeVersion     string
		getClusterResponse []types.CAPICluster
		cpResponse         error
		workerResponse     error
		nodeResponse       error
		crdResponse        error
		wantErr            error
		modifyFunc         func(s *cluster.Spec)
	}{
		{
			name:               "ValidationFluxSshKeyAlgoImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("fluxConfig spec.fluxConfig.spec.git.sshKeyAlgorithm is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.FluxConfig.Spec.Git.SshKeyAlgorithm = "rsa2"
			},
		},
		{
			name:               "ValidationFluxRepoUrlImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("fluxConfig spec.fluxConfig.spec.git.repositoryUrl is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.FluxConfig.Spec.Git.RepositoryUrl = "test2"
			},
		},
	}
	defaultControlPlane := v1alpha1.ControlPlaneConfiguration{
		Count: 1,
		Endpoint: &v1alpha1.Endpoint{
			Host: "1.1.1.1",
		},
		MachineGroupRef: &v1alpha1.Ref{
			Name: "test",
			Kind: "VSphereMachineConfig",
		},
	}

	defaultETCD := &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
	}

	defaultDatacenterSpec := v1alpha1.VSphereDatacenterConfig{
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Datacenter: "datacenter!!!",
			Network:    "network",
			Server:     "server",
			Thumbprint: "thumbprint",
			Insecure:   false,
		},
		Status: v1alpha1.VSphereDatacenterConfigStatus{},
	}

	defaultFlux := &v1alpha1.FluxConfig{
		Spec: v1alpha1.FluxConfigSpec{
			Git: &v1alpha1.GitProviderConfig{
				RepositoryUrl:   "test",
				SshKeyAlgorithm: "rsa",
			},
		},
	}
	defaultOIDC := &v1alpha1.OIDCConfig{
		Spec: v1alpha1.OIDCConfigSpec{
			ClientId:     "client-id",
			GroupsClaim:  "groups-claim",
			GroupsPrefix: "groups-prefix",
			IssuerUrl:    "issuer-url",
			RequiredClaims: []v1alpha1.OIDCConfigRequiredClaim{{
				Claim: "claim",
				Value: "value",
			}},
			UsernameClaim:  "username-claim",
			UsernamePrefix: "username-prefix",
		},
	}

	defaultAWSIAM := &v1alpha1.AWSIamConfig{
		Spec: v1alpha1.AWSIamConfigSpec{
			AWSRegion: "us-east-1",
			MapRoles: []v1alpha1.MapRoles{{
				RoleARN:  "roleARN",
				Username: "username",
				Groups:   []string{"group1", "group2"},
			}},
			MapUsers: []v1alpha1.MapUsers{{
				UserARN:  "userARN",
				Username: "username",
				Groups:   []string{"group1", "group2"},
			}},
			Partition: "partition",
		},
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = testclustername
		s.Cluster.Spec.ControlPlaneConfiguration = defaultControlPlane
		s.Cluster.Spec.ExternalEtcdConfiguration = defaultETCD
		s.Cluster.Spec.DatacenterRef = v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: "vsphere test",
		}
		s.Cluster.Spec.IdentityProviderRefs = []v1alpha1.Ref{
			{
				Kind: v1alpha1.OIDCConfigKind,
				Name: "oidc",
			},
			{
				Kind: v1alpha1.AWSIamConfigKind,
				Name: "aws-iam",
			},
		}
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.FluxConfigKind,
			Name: "flux test",
		}
		s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
			Pods: v1alpha1.Pods{
				CidrBlocks: []string{
					"1.2.3.4/5",
				},
			},
			Services: v1alpha1.Services{
				CidrBlocks: []string{
					"1.2.3.4/6",
				},
			},
			DNS: v1alpha1.DNS{
				ResolvConf: &v1alpha1.ResolvConf{Path: "file.conf"},
			},
		}
		s.Cluster.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
			HttpProxy:  "httpproxy",
			HttpsProxy: "httpsproxy",
			NoProxy: []string{
				"noproxy1",
				"noproxy2",
			},
		}

		s.OIDCConfig = defaultOIDC
		s.AWSIamConfig = defaultAWSIAM
		s.FluxConfig = defaultFlux
	})

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			_, ctx, workloadCluster, _ := validations.NewKubectl(t)
			workloadCluster.KubeconfigFile = kubeconfigFilePath
			workloadCluster.Name = testclustername

			mockCtrl := gomock.NewController(t)
			k := mocks.NewMockKubectlClient(mockCtrl)
			tlsValidator := mocks.NewMockTlsValidator(mockCtrl)
			cliConfig := &config.CliConfig{
				GitPrivateKeyFile:   "testdata/git_nonempty_private_key",
				GitSshKeyPassphrase: "test",
				GitKnownHostsFile:   "testdata/git_nonempty_ssh_known_hosts",
			}

			provider := mockproviders.NewMockProvider(mockCtrl)
			opts := &validations.Opts{
				Kubectl:           k,
				Spec:              clusterSpec,
				WorkloadCluster:   workloadCluster,
				ManagementCluster: workloadCluster,
				Provider:          provider,
				TLSValidator:      tlsValidator,
				CliConfig:         cliConfig,
			}

			clusterSpec.Cluster.Spec.KubernetesVersion = v1alpha1.KubernetesVersion(tc.upgradeVersion)
			existingClusterSpec := clusterSpec.DeepCopy()
			existingProviderSpec := defaultDatacenterSpec.DeepCopy()
			if tc.modifyFunc != nil {
				tc.modifyFunc(existingClusterSpec)
			}
			versionResponse := &executables.VersionResponse{
				ServerVersion: version.Info{
					GitVersion: tc.clusterVersion,
				},
			}

			provider.EXPECT().DatacenterConfig(clusterSpec).Return(existingProviderSpec).MaxTimes(1)
			provider.EXPECT().ValidateNewSpec(ctx, workloadCluster, clusterSpec).Return(nil).MaxTimes(1)
			k.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			k.EXPECT().GetEksaVSphereDatacenterConfig(ctx, clusterSpec.Cluster.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(existingProviderSpec, nil).MaxTimes(1)
			k.EXPECT().ValidateControlPlaneNodes(ctx, workloadCluster, clusterSpec.Cluster.Name).Return(tc.cpResponse)
			k.EXPECT().ValidateWorkerNodes(ctx, workloadCluster.Name, workloadCluster.KubeconfigFile).Return(tc.workerResponse)
			k.EXPECT().ValidateNodes(ctx, kubeconfigFilePath).Return(tc.nodeResponse)
			k.EXPECT().ValidateClustersCRD(ctx, workloadCluster).Return(tc.crdResponse)
			k.EXPECT().GetClusters(ctx, workloadCluster).Return(tc.getClusterResponse, nil)
			k.EXPECT().GetEksaCluster(ctx, workloadCluster, clusterSpec.Cluster.Name).Return(existingClusterSpec.Cluster, nil)
			k.EXPECT().GetEksaFluxConfig(ctx, clusterSpec.Cluster.Spec.GitOpsRef.Name, gomock.Any(), gomock.Any()).Return(existingClusterSpec.FluxConfig, nil).MaxTimes(1)
			k.EXPECT().GetEksaOIDCConfig(ctx, clusterSpec.Cluster.Spec.IdentityProviderRefs[0].Name, gomock.Any(), gomock.Any()).Return(existingClusterSpec.OIDCConfig, nil).MaxTimes(1)
			k.EXPECT().GetEksaAWSIamConfig(ctx, clusterSpec.Cluster.Spec.IdentityProviderRefs[1].Name, gomock.Any(), gomock.Any()).Return(existingClusterSpec.AWSIamConfig, nil).MaxTimes(1)
			k.EXPECT().Version(ctx, workloadCluster).Return(versionResponse, nil)
			upgradeValidations := upgradevalidations.New(opts)
			err := validations.ProcessValidationResults(upgradeValidations.PreflightValidations(ctx))
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%s want err=%v\n got err=%v\n", tc.name, tc.wantErr, err)
			}
		})
	}
}

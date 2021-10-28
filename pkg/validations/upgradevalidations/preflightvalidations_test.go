package upgradevalidations_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/version"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockproviders "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations/mocks"
)

const (
	kubeconfigFilePath = "./fakeKubeconfigFilePath"
)

var goodClusterResponse = []types.CAPICluster{{Metadata: types.Metadata{Name: testclustername}}}

func TestPreflightValidations(t *testing.T) {
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
			name:               "ValidationSucceeds",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,

			workerResponse: nil,
			nodeResponse:   nil,
			crdResponse:    nil,
			wantErr:        nil,
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
			wantErr:            composeError("spec.externalEtcdConfiguration is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.Spec.ExternalEtcdConfiguration.Count++
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
			modifyFunc: func(s *cluster.Spec) {
				s.Spec.ControlPlaneConfiguration.Endpoint.Host = "2.3.4.5"
			},
		},
		{
			name:               "ValidationIdentityProviderRefsImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.identityProviderRefs is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.Spec.IdentityProviderRefs = []v1alpha1.Ref{
					{
						Kind: v1alpha1.OIDCConfigKind,
						Name: "oidc-2",
					},
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
			wantErr:            composeError("gitOps is immutable"),
			modifyFunc: func(s *cluster.Spec) {
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
			wantErr:            composeError("gitOps is immutable"),
			modifyFunc: func(s *cluster.Spec) {
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
			wantErr:            composeError("gitOps is immutable"),
			modifyFunc: func(s *cluster.Spec) {
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
			wantErr:            composeError("gitOps is immutable"),
			modifyFunc: func(s *cluster.Spec) {
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
			wantErr:            composeError("gitOps is immutable"),
			modifyFunc: func(s *cluster.Spec) {
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
			wantErr:            composeError("gitOps is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.GitOpsConfig.Spec.Flux.Github.Personal = !s.GitOpsConfig.Spec.Flux.Github.Personal
			},
		},
		{
			name:               "ValidationOIDCClientIdImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("oidc identity provider is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.ClientId = "new-client-id"
			},
		},
		{
			name:               "ValidationOIDCGroupsClaimImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("oidc identity provider is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.GroupsClaim = "new-groups-claim"
			},
		},
		{
			name:               "ValidationOIDCGroupsPrefixImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("oidc identity provider is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.GroupsPrefix = "new-groups-prefix"
			},
		},
		{
			name:               "ValidationOIDCIssuerUrlImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("oidc identity provider is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.IssuerUrl = "new-issuer-url"
			},
		},
		{
			name:               "ValidationOIDCUsernameClaimImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("oidc identity provider is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.UsernameClaim = "new-username-claim"
			},
		},
		{
			name:               "ValidationOIDCUsernamePrefixImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("oidc identity provider is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.UsernamePrefix = "new-username-prefix"
			},
		},
		{
			name:               "ValidationOIDCRequiredClaimsImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("oidc identity provider is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.OIDCConfig.Spec.RequiredClaims[0].Claim = "new-groups-claim"
			},
		},
		{
			name:               "ValidationClusterNetworkImmutable",
			clusterVersion:     "v1.19.16-eks-1-19-4",
			upgradeVersion:     "1.19",
			getClusterResponse: goodClusterResponse,
			cpResponse:         nil,
			workerResponse:     nil,
			nodeResponse:       nil,
			crdResponse:        nil,
			wantErr:            composeError("spec.clusterNetwork is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{}
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
			modifyFunc: func(s *cluster.Spec) {
				s.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
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
			wantErr:            composeError("spec.externalEtcdConfiguration is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.Spec.ExternalEtcdConfiguration.Count += 1
				s.Spec.DatacenterRef = v1alpha1.Ref{
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
			wantErr:            composeError("spec.externalEtcdConfiguration is immutable"),
			modifyFunc: func(s *cluster.Spec) {
				s.Spec.ExternalEtcdConfiguration = nil
				s.Spec.DatacenterRef = v1alpha1.Ref{
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
			modifyFunc: func(s *cluster.Spec) {
				if s.Spec.Management == nil {
					nb := false
					s.Spec.Management = &nb
				} else {
					*s.Spec.Management = !*s.Spec.Management
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

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = testclustername
		s.Spec.ControlPlaneConfiguration = defaultControlPlane
		s.Spec.ExternalEtcdConfiguration = defaultETCD
		s.Spec.DatacenterRef = v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: "vsphere test",
		}
		s.Spec.IdentityProviderRefs = []v1alpha1.Ref{
			{
				Kind: v1alpha1.OIDCConfigKind,
				Name: "oidc",
			},
		}
		s.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.GitOpsConfigKind,
			Name: "gitops test",
		}
		s.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
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
		s.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
			HttpProxy:  "httpproxy",
			HttpsProxy: "httpsproxy",
			NoProxy: []string{
				"noproxy1",
				"noproxy2",
			},
		}

		s.GitOpsConfig = defaultGitOps
		s.OIDCConfig = defaultOIDC
	})

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			_, ctx, workloadCluster, _ := newKubectl(t)
			workloadCluster.KubeconfigFile = kubeconfigFilePath
			workloadCluster.Name = testclustername

			mockCtrl := gomock.NewController(t)
			k := mocks.NewMockValidationsKubectlClient(mockCtrl)

			provider := mockproviders.NewMockProvider(mockCtrl)
			opts := &upgradevalidations.UpgradeValidationOpts{
				Kubectl:           k,
				Spec:              clusterSpec,
				WorkloadCluster:   workloadCluster,
				ManagementCluster: workloadCluster,
				Provider:          provider,
			}

			clusterSpec.Spec.KubernetesVersion = v1alpha1.KubernetesVersion(tc.upgradeVersion)
			existingClusterSpec := &cluster.Spec{
				Cluster:      clusterSpec.Cluster.DeepCopy(),
				GitOpsConfig: clusterSpec.GitOpsConfig.DeepCopy(),
				OIDCConfig:   clusterSpec.OIDCConfig.DeepCopy(),
			}
			existingProviderSpec := defaultDatacenterSpec.DeepCopy()
			if tc.modifyFunc != nil {
				tc.modifyFunc(existingClusterSpec)
			}
			versionResponse := &executables.VersionResponse{
				ServerVersion: version.Info{
					GitVersion: tc.clusterVersion,
				},
			}

			provider.EXPECT().DatacenterConfig().Return(existingProviderSpec).MaxTimes(1)
			provider.EXPECT().ValidateNewSpec(ctx, workloadCluster, clusterSpec).Return(nil).MaxTimes(1)
			k.EXPECT().GetEksaVSphereDatacenterConfig(ctx, clusterSpec.Spec.DatacenterRef.Name, gomock.Any(), gomock.Any()).Return(existingProviderSpec, nil).MaxTimes(1)
			k.EXPECT().ValidateControlPlaneNodes(ctx, workloadCluster, clusterSpec.Name).Return(tc.cpResponse)
			k.EXPECT().ValidateWorkerNodes(ctx, workloadCluster, workloadCluster.Name).Return(tc.workerResponse)
			k.EXPECT().ValidateNodes(ctx, kubeconfigFilePath).Return(tc.nodeResponse)
			k.EXPECT().ValidateClustersCRD(ctx, workloadCluster).Return(tc.crdResponse)
			k.EXPECT().GetClusters(ctx, workloadCluster).Return(tc.getClusterResponse, nil)
			k.EXPECT().GetEksaCluster(ctx, workloadCluster, clusterSpec.Name).Return(existingClusterSpec.Cluster, nil)
			k.EXPECT().GetEksaGitOpsConfig(ctx, clusterSpec.Spec.GitOpsRef.Name, gomock.Any(), gomock.Any()).Return(existingClusterSpec.GitOpsConfig, nil).MaxTimes(1)
			k.EXPECT().GetEksaOIDCConfig(ctx, clusterSpec.Spec.IdentityProviderRefs[0].Name, gomock.Any(), gomock.Any()).Return(existingClusterSpec.OIDCConfig, nil).MaxTimes(1)
			k.EXPECT().Version(ctx, workloadCluster).Return(versionResponse, nil)
			upgradeValidations := upgradevalidations.New(opts)
			err := upgradeValidations.PreflightValidations(ctx)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%s want err=%v\n got err=%v\n", tc.name, tc.wantErr, err)
			}
		})
	}
}

func composeError(msgs ...string) *upgradevalidations.ValidationError {
	var errs []string
	errs = append(errs, msgs...)
	return &upgradevalidations.ValidationError{Errs: errs}
}

var explodingClusterError = composeError(
	"control plane nodes are not ready",
	"2 worker nodes are not ready",
	"node test-node is not ready, currently in Unknown state",
	"error getting clusters crd: crd not found",
	"couldn't find CAPI cluster object for cluster with name testcluster",
	"WARNING: version difference between upgrade version (1.20) and server version (1.18) do not meet the supported version increment of +1",
)

package upgradevalidations_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	pmock "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestValidateGitOpsImmutableFieldsRef(t *testing.T) {
	tests := []struct {
		name           string
		oldRef, newRef *v1alpha1.Ref
		wantErr        string
	}{
		{
			name:   "old gitRef nil, new gitRef not nil",
			oldRef: nil,
			newRef: &v1alpha1.Ref{
				Kind: "GitOpsConfig",
				Name: "gitops-new",
			},
			wantErr: "",
		},
		{
			name: "old gitRef not nil, new gitRef nil",
			oldRef: &v1alpha1.Ref{
				Kind: "GitOpsConfig",
				Name: "gitops-old",
			},
			newRef:  nil,
			wantErr: "once cluster.spec.gitOpsRef is set, it is immutable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = testclustername
				s.Cluster.Spec.GitOpsRef = tc.newRef
			})
			cluster := &types.Cluster{
				KubeconfigFile: "kubeconfig",
			}

			oldCluster := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					GitOpsRef: tc.oldRef,
				},
			}

			err := upgradevalidations.ValidateGitOpsImmutableFields(ctx, nil, cluster, clusterSpec, oldCluster)
			if tc.wantErr == "" {
				g.Expect(err).To(Succeed())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tc.wantErr)))
			}
		})
	}
}

type gitOpsTest struct {
	*WithT
	ctx context.Context
	k   *mocks.MockKubectlClient
	c   *types.Cluster
	o   *v1alpha1.Cluster
	s   *cluster.Spec
}

func newGitClientTest(t *testing.T) *gitOpsTest {
	ctrl := gomock.NewController(t)
	return &gitOpsTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		k:     mocks.NewMockKubectlClient(ctrl),
		c: &types.Cluster{
			KubeconfigFile: "kubeconfig",
		},
		o: &v1alpha1.Cluster{
			Spec: v1alpha1.ClusterSpec{
				GitOpsRef: &v1alpha1.Ref{
					Name: "test",
					Kind: "GitOpsConfig",
				},
			},
		},
		s: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Name = testclustername
			s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
				Name: "test",
				Kind: "GitOpsConfig",
			}
		}),
	}
}

func TestValidateGitOpsImmutableFieldsGetEksaGitOpsConfigError(t *testing.T) {
	g := newGitClientTest(t)
	g.k.EXPECT().GetEksaGitOpsConfig(g.ctx, g.s.Cluster.Spec.GitOpsRef.Name, "kubeconfig", "").Return(nil, errors.New("error in get gitops config"))
	g.Expect(upgradevalidations.ValidateGitOpsImmutableFields(g.ctx, g.k, g.c, g.s, g.o)).To(MatchError(ContainSubstring("error in get gitops config")))
}

func TestValidateGitOpsImmutableFieldsGetEksaFluxConfigError(t *testing.T) {
	g := newGitClientTest(t)
	g.o.Spec.GitOpsRef.Kind = "FluxConfig"
	g.s.Cluster.Spec.GitOpsRef.Kind = "FluxConfig"
	g.k.EXPECT().GetEksaFluxConfig(g.ctx, g.s.Cluster.Spec.GitOpsRef.Name, "kubeconfig", "").Return(nil, errors.New("error in get flux config"))
	g.Expect(upgradevalidations.ValidateGitOpsImmutableFields(g.ctx, g.k, g.c, g.s, g.o)).To(MatchError(ContainSubstring("error in get flux config")))
}

func TestValidateGitOpsImmutableFieldsFluxConfig(t *testing.T) {
	tests := []struct {
		name     string
		new, old *v1alpha1.FluxConfig
		wantErr  string
	}{
		{
			name: "github repo diff",
			new: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Github: &v1alpha1.GithubProviderConfig{
						Repository: "a",
					},
				},
			},
			old: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Github: &v1alpha1.GithubProviderConfig{
						Repository: "b",
					},
				},
			},
			wantErr: "fluxConfig spec.github.repository is immutable",
		},
		{
			name: "github owner diff",
			new: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Github: &v1alpha1.GithubProviderConfig{
						Owner: "a",
					},
				},
			},
			old: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Github: &v1alpha1.GithubProviderConfig{
						Owner: "b",
					},
				},
			},
			wantErr: "fluxConfig spec.github.owner is immutable",
		},
		{
			name: "github personal diff",
			new: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Github: &v1alpha1.GithubProviderConfig{
						Personal: true,
					},
				},
			},
			old: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Github: &v1alpha1.GithubProviderConfig{
						Personal: false,
					},
				},
			},
			wantErr: "fluxConfig spec.github.personal is immutable",
		},
		{
			name: "branch diff",
			new: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Branch: "a",
				},
			},
			old: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Branch: "b",
				},
			},
			wantErr: "fluxConfig spec.branch is immutable",
		},
		{
			name: "clusterConfigPath diff",
			new: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					ClusterConfigPath: "a",
				},
			},
			old: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					ClusterConfigPath: "b",
				},
			},
			wantErr: "fluxConfig spec.clusterConfigPath is immutable",
		},
		{
			name: "systemNamespace diff",
			new: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					SystemNamespace: "a",
				},
			},
			old: &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					SystemNamespace: "b",
				},
			},
			wantErr: "fluxConfig spec.systemNamespace is immutable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			g := newGitClientTest(t)
			g.o.Spec.GitOpsRef.Kind = "FluxConfig"
			g.s.Cluster.Spec.GitOpsRef.Kind = "FluxConfig"
			g.s.FluxConfig = tc.new

			g.k.EXPECT().GetEksaFluxConfig(g.ctx, g.s.Cluster.Spec.GitOpsRef.Name, "kubeconfig", "").Return(tc.old, nil)

			err := upgradevalidations.ValidateGitOpsImmutableFields(g.ctx, g.k, g.c, g.s, g.o)
			if tc.wantErr == "" {
				g.Expect(err).To(Succeed())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tc.wantErr)))
			}
		})
	}
}

func TestValidateImmutableFields(t *testing.T) {
	tests := []struct {
		Name             string
		ConfigureCurrent func(current *v1alpha1.Cluster)
		ConfigureDesired func(desired *v1alpha1.Cluster)
		ExpectedError    string
	}{
		{
			Name: "Toggle Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade on",
			ConfigureCurrent: func(current *v1alpha1.Cluster) {
				current.Spec.ClusterNetwork.CNIConfig = &v1alpha1.CNIConfig{
					Cilium: &v1alpha1.CiliumConfig{
						SkipUpgrade: ptr.Bool(false),
					},
				}
			},
			ConfigureDesired: func(desired *v1alpha1.Cluster) {
				desired.Spec.ClusterNetwork.CNIConfig = &v1alpha1.CNIConfig{
					Cilium: &v1alpha1.CiliumConfig{
						SkipUpgrade: ptr.Bool(true),
					},
				}
			},
		},
		{
			Name: "Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade unset",
			ConfigureCurrent: func(current *v1alpha1.Cluster) {
				current.Spec.ClusterNetwork.CNIConfig = &v1alpha1.CNIConfig{
					Cilium: &v1alpha1.CiliumConfig{},
				}
			},
			ConfigureDesired: func(desired *v1alpha1.Cluster) {
				desired.Spec.ClusterNetwork.CNIConfig = &v1alpha1.CNIConfig{
					Cilium: &v1alpha1.CiliumConfig{},
				}
			},
		},
		{
			Name: "Toggle Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade off",
			ConfigureCurrent: func(current *v1alpha1.Cluster) {
				current.Spec.ClusterNetwork.CNIConfig = &v1alpha1.CNIConfig{
					Cilium: &v1alpha1.CiliumConfig{
						SkipUpgrade: ptr.Bool(true),
					},
				}
			},
			ConfigureDesired: func(desired *v1alpha1.Cluster) {
				desired.Spec.ClusterNetwork.CNIConfig = &v1alpha1.CNIConfig{
					Cilium: &v1alpha1.CiliumConfig{
						SkipUpgrade: ptr.Bool(false),
					},
				}
			},
			ExpectedError: "spec.clusterNetwork.cniConfig.cilium.skipUpgrade cannot be toggled off",
		},
	}

	clstr := &types.Cluster{}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			g := NewWithT(t)
			ctrl := gomock.NewController(t)

			current := &cluster.Spec{
				Config: &cluster.Config{
					Cluster: &v1alpha1.Cluster{
						Spec: v1alpha1.ClusterSpec{
							WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{}},
						},
					},
				},
				Bundles: &releasev1alpha1.Bundles{},
			}
			desired := current.DeepCopy()

			tc.ConfigureCurrent(current.Config.Cluster)
			tc.ConfigureDesired(desired.Config.Cluster)

			client := mocks.NewMockKubectlClient(ctrl)
			client.EXPECT().
				GetEksaCluster(gomock.Any(), clstr, current.Cluster.Name).
				Return(current.Cluster, nil)

			provider := pmock.NewMockProvider(ctrl)

			// The algorithm calls out to the provider to validate the new spec only if it finds
			// no errors in the generic validation first.
			if tc.ExpectedError == "" {
				provider.EXPECT().
					ValidateNewSpec(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			}

			err := upgradevalidations.ValidateImmutableFields(
				context.Background(),
				client,
				clstr,
				desired,
				provider,
			)

			if tc.ExpectedError == "" {
				g.Expect(err).To(Succeed())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tc.ExpectedError)))
			}
		})
	}
}

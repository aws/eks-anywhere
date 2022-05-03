package createvalidations_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
)

const (
	emptyVar   = ""
	testEnvVar = "test"
)

var eksaGitOpsResourceType = fmt.Sprintf("gitopsconfigs.%s", v1alpha1.GroupVersion.Group)

func TestValidateGitopsForWorkloadClustersPath(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
		github  v1alpha1.Github
	}{
		{
			name:    "Success",
			wantErr: nil,
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "repo",
				FluxSystemNamespace: "flux-system",
				Branch:              "main",
				ClusterConfigPath:   "clusters/management-gitops",
				Personal:            false,
			},
		},
		{
			name:    "Failure, path diff",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "repo",
				FluxSystemNamespace: "flux-system",
				Branch:              "main",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            false,
			},
		},
		{
			name:    "Failure, branch diff",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "repo",
				FluxSystemNamespace: "flux-system",
				Branch:              "dev",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            false,
			},
		},
		{
			name:    "Failure, owner owner",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "janedoe",
				Repository:          "repo",
				FluxSystemNamespace: "flux-system",
				Branch:              "main",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            false,
			},
		},
		{
			name:    "Failure, repo diff",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "diffrepo",
				FluxSystemNamespace: "flux-system",
				Branch:              "main",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            false,
			},
		},
		{
			name:    "Failure, namespace diff",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "repo",
				FluxSystemNamespace: "diff-ns",
				Branch:              "main",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            false,
			},
		},
		{
			name:    "Failure, personal diff",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "repo",
				FluxSystemNamespace: "flux-system",
				Branch:              "main",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            true,
			},
		},
	}

	gitOpsListContent := test.ReadFile(t, "testdata/empty_get_gitops_response.json")
	eksaClusterContent := test.ReadFile(t, "testdata/eksa_cluster_exists.json")
	mgmtGitOpsContent := test.ReadFile(t, "testdata/management_gitops_config.json")

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			defaultGitOps := &v1alpha1.GitOpsConfig{
				Spec: v1alpha1.GitOpsConfigSpec{
					Flux: v1alpha1.Flux{
						Github: tc.github,
					},
				},
			}

			clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = testclustername
				s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
					Kind: v1alpha1.GitOpsConfigKind,
					Name: "gitopstest",
				}
				s.GitOpsConfig = defaultGitOps
				s.Cluster.SetManagedBy("management-cluster")
			})
			cliConfig := &config.CliConfig{}
			k, ctx, cluster, e := validations.NewKubectl(t)
			cluster.Name = "management-cluster"

			e.EXPECT().Execute(
				ctx, []string{
					"get", eksaGitOpsResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
					"--field-selector=metadata.name=gitopstest",
				}).Return(*bytes.NewBufferString(gitOpsListContent), nil)

			e.EXPECT().Execute(
				ctx, []string{
					"get", "clusters.anywhere.eks.amazonaws.com", "-A", "-o", "jsonpath={.items[0]}", "--kubeconfig",
					cluster.KubeconfigFile,
					"--field-selector=metadata.name=management-cluster",
				}).Return(*bytes.NewBufferString(eksaClusterContent), nil)

			e.EXPECT().Execute(
				ctx, []string{
					"get", eksaGitOpsResourceType, "management-gitops", "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
				}).Return(*bytes.NewBufferString(mgmtGitOpsContent), nil)

			err := createvalidations.ValidateGitOps(ctx, k, cluster, clusterSpec, cliConfig)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateGitopsForWorkloadClustersFailure(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		getClusterResponse string
	}{
		{
			name:               "FailureGitOpsNameExists",
			wantErr:            errors.New("gitOpsConfig gitopstest already exists"),
			getClusterResponse: "testdata/gitops_name_exists.json",
		},
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

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = testclustername
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.GitOpsConfigKind,
			Name: "gitopstest",
		}
		s.GitOpsConfig = defaultGitOps
		s.Cluster.SetManagedBy("management-cluster")
	})
	cliConfig := &config.CliConfig{}
	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			fileContent := test.ReadFile(t, tc.getClusterResponse)
			e.EXPECT().Execute(
				ctx, []string{
					"get", eksaGitOpsResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
					"--field-selector=metadata.name=gitopstest",
				}).Return(*bytes.NewBufferString(fileContent), nil)

			err := createvalidations.ValidateGitOps(ctx, k, cluster, clusterSpec, cliConfig)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestSkipValidateGitopsWithNoGitOpts(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "SuccessNoGitops",
			wantErr: nil,
		},
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.SetManagedBy("management-cluster")
		s.Cluster.Name = testclustername

		s.GitOpsConfig = nil
	})
	cliConfig := &config.CliConfig{}

	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			e.EXPECT().Execute(
				ctx, []string{
					"get", eksaGitOpsResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
					"--field-selector=metadata.name=gitopstest",
				}).Times(0)

			err := createvalidations.ValidateGitOps(ctx, k, cluster, clusterSpec, cliConfig)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateGitopsForSelfManagedCluster(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		getClusterResponse string
	}{
		{
			name:               "Skip Validate GitOpsConfig name",
			wantErr:            nil,
			getClusterResponse: "testdata/empty_get_gitops_response.json",
		},
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

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = testclustername
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.GitOpsConfigKind,
			Name: "gitopstest",
		}
		s.GitOpsConfig = defaultGitOps

		s.Cluster.SetSelfManaged()
	})
	cliConfig := &config.CliConfig{}
	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			e.EXPECT().Execute(
				ctx, []string{
					"get", eksaGitOpsResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
					"--field-selector=metadata.name=gitopstest",
				}).Times(0)

			err := createvalidations.ValidateGitOps(ctx, k, cluster, clusterSpec, cliConfig)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateGitOpsGitProviderNoAuthForWorkloadCluster(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   error
		git       *v1alpha1.GitProviderConfig
		cliConfig *config.CliConfig
	}{
		{
			name:    "Empty password and private key",
			wantErr: fmt.Errorf("provide a path to a private key file via the EKSA_GIT_PRIVATE_KEY in order to use the generic git Flux provider"),
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile: emptyVar,
				GitPassword:       emptyVar,
				GitKnownHostsFile: "testdata/git_nonempty_ssh_known_hosts",
			},
		},
		{
			name:      "Empty git config",
			wantErr:   nil,
			git:       nil,
			cliConfig: nil,
		},
		{
			name:    "Empty password",
			wantErr: fmt.Errorf("private key file does not exist at %s or is empty", testEnvVar),
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile: testEnvVar,
				GitPassword:       emptyVar,
				GitKnownHostsFile: "testdata/git_nonempty_ssh_known_hosts",
			},
		},
		{
			name:    "Empty private key file",
			wantErr: fmt.Errorf("private key file does not exist at %s or is empty", "testdata/git_empty_file"),
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile: "testdata/git_empty_file",
				GitPassword:       emptyVar,
				GitKnownHostsFile: "testdata/git_nonempty_ssh_known_hosts",
			},
		},
		{
			name:    "Empty private key",
			wantErr: fmt.Errorf("provide a path to a private key file via the EKSA_GIT_PRIVATE_KEY in order to use the generic git Flux provider"),
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile: emptyVar,
				GitPassword:       testEnvVar,
				GitKnownHostsFile: "testdata/git_nonempty_ssh_known_hosts",
			},
		},
		{
			name:    "Password and private key populated",
			wantErr: nil,
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile: "testdata/git_nonempty_private_key",
				GitPassword:       testEnvVar,
				GitKnownHostsFile: "testdata/git_nonempty_ssh_known_hosts",
			},
		},
		{
			name:    "Empty known hosts",
			wantErr: fmt.Errorf("SSH known hosts file does not exist at testdata/git_empty_file or is empty"),
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile: "testdata/git_nonempty_private_key",
				GitPassword:       testEnvVar,
				GitKnownHostsFile: "testdata/git_empty_file",
			},
		},
		{
			name:    "No known hosts",
			wantErr: fmt.Errorf("SSH known hosts file does not exist at testdata/git_empty_file or is empty"),
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile: "testdata/git_nonempty_private_key",
				GitPassword:       testEnvVar,
				GitKnownHostsFile: "testdata/git_empty_file",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			defaultFlux := &v1alpha1.FluxConfig{
				Spec: v1alpha1.FluxConfigSpec{
					Git: tc.git,
				},
			}
			clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = testclustername
				s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
					Kind: v1alpha1.GitOpsConfigKind,
					Name: "gitopstest",
				}
				s.FluxConfig = defaultFlux
				s.Cluster.SetManagedBy("management-cluster")
			})

			cliConfig := tc.cliConfig
			k, ctx, cluster, _ := validations.NewKubectl(t)
			cluster.Name = "management-cluster"

			err := createvalidations.ValidateGitOps(ctx, k, cluster, clusterSpec, cliConfig)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

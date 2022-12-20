package validations_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/validations"
)

const (
	emptyVar   = ""
	testEnvVar = "test"
)

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
				GitPrivateKeyFile:   emptyVar,
				GitSshKeyPassphrase: emptyVar,
				GitKnownHostsFile:   "testdata/git_nonempty_ssh_known_hosts",
			},
		},
		{
			name:    "cliConfig nil",
			wantErr: nil,
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: nil,
		},
		{
			name:    "Empty known host",
			wantErr: fmt.Errorf("provide a path to an SSH Known Hosts file which contains a valid entry associate with the given private key via the EKSA_GIT_SSH_KNOWN_HOSTS environment variable"),
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile:   "testdata/git_nonempty_private_key",
				GitSshKeyPassphrase: testEnvVar,
				GitKnownHostsFile:   emptyVar,
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
				GitPrivateKeyFile:   testEnvVar,
				GitSshKeyPassphrase: emptyVar,
				GitKnownHostsFile:   "testdata/git_nonempty_ssh_known_hosts",
			},
		},
		{
			name:    "Empty private key file",
			wantErr: fmt.Errorf("private key file does not exist at %s or is empty", "testdata/git_empty_file"),
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile:   "testdata/git_empty_file",
				GitSshKeyPassphrase: emptyVar,
				GitKnownHostsFile:   "testdata/git_nonempty_ssh_known_hosts",
			},
		},
		{
			name:    "Empty private key",
			wantErr: fmt.Errorf("provide a path to a private key file via the EKSA_GIT_PRIVATE_KEY in order to use the generic git Flux provider"),
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile:   emptyVar,
				GitSshKeyPassphrase: testEnvVar,
				GitKnownHostsFile:   "testdata/git_nonempty_ssh_known_hosts",
			},
		},
		{
			name:    "Password and private key populated",
			wantErr: nil,
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile:   "testdata/git_nonempty_private_key",
				GitSshKeyPassphrase: testEnvVar,
				GitKnownHostsFile:   "testdata/git_nonempty_ssh_known_hosts",
			},
		},
		{
			name:    "Empty known hosts",
			wantErr: fmt.Errorf("SSH known hosts file does not exist at testdata/git_empty_file or is empty"),
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile:   "testdata/git_nonempty_private_key",
				GitSshKeyPassphrase: testEnvVar,
				GitKnownHostsFile:   "testdata/git_empty_file",
			},
		},
		{
			name:    "No known hosts",
			wantErr: fmt.Errorf("SSH known hosts file does not exist at testdata/git_empty_file or is empty"),
			git: &v1alpha1.GitProviderConfig{
				RepositoryUrl: "testRepo",
			},
			cliConfig: &config.CliConfig{
				GitPrivateKeyFile:   "testdata/git_nonempty_private_key",
				GitSshKeyPassphrase: testEnvVar,
				GitKnownHostsFile:   "testdata/git_empty_file",
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
				s.Cluster.Name = "testcluster"
				s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
					Kind: v1alpha1.GitOpsConfigKind,
					Name: "gitopstest",
				}
				s.FluxConfig = defaultFlux
				s.Cluster.SetManagedBy("management-cluster")
			})

			_, _, cluster, _ := validations.NewKubectl(t)
			cluster.Name = "management-cluster"

			err := validations.ValidateAuthenticationForGitProvider(clusterSpec, tc.cliConfig)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

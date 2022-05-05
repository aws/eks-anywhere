package v1alpha1

import (
	"os"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EksaGitPrivateKeyTokenEnv = "EKSA_GIT_PRIVATE_KEY"
	EksaGitKnownHostsFileEnv  = "EKSA_GIT_KNOWN_HOSTS"
)

type testContext struct {
	privateKeyFile         string
	isprivateKeyFileSet    bool
	gitKnownHostsFile      string
	isGitKnownHostsFileSet bool
}

func (tctx *testContext) SaveContext() {
	tctx.privateKeyFile, tctx.isprivateKeyFileSet = os.LookupEnv(EksaGitPrivateKeyTokenEnv)
	tctx.gitKnownHostsFile, tctx.isGitKnownHostsFileSet = os.LookupEnv(EksaGitKnownHostsFileEnv)
	os.Setenv(EksaGitPrivateKeyTokenEnv, "my/private/key")
	os.Setenv(EksaGitKnownHostsFileEnv, "my/known/hosts")
}

func (tctx *testContext) RestoreContext() {
	if tctx.isprivateKeyFileSet {
		os.Setenv(EksaGitPrivateKeyTokenEnv, tctx.privateKeyFile)
	} else {
		os.Unsetenv(EksaGitPrivateKeyTokenEnv)
	}
	if tctx.isGitKnownHostsFileSet {
		os.Setenv(EksaGitKnownHostsFileEnv, tctx.gitKnownHostsFile)
	} else {
		os.Unsetenv(EksaGitKnownHostsFileEnv)
	}
}

func setupContext(t *testing.T) {
	var tctx testContext
	tctx.SaveContext()
	t.Cleanup(func() {
		tctx.RestoreContext()
	})
}

func TestGetAndValidateFluxConfig(t *testing.T) {
	tests := []struct {
		testName       string
		fileName       string
		refName        string
		wantFluxConfig *FluxConfig
		clusterConfig  *Cluster
		wantErr        bool
		gitProvider    bool
	}{
		{
			testName:       "file doesn't exist",
			fileName:       "testdata/fake_file.yaml",
			wantFluxConfig: nil,
			wantErr:        true,
		},
		{
			testName:       "not parseable file",
			fileName:       "testdata/not_parseable_fluxconfig.yaml",
			wantFluxConfig: nil,
			wantErr:        true,
		},
		{
			testName: "valid 1.19 github",
			fileName: "testdata/cluster_1_19_flux_github.yaml",
			refName:  "test-flux-github",
			wantFluxConfig: &FluxConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       FluxConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-flux-github",
					Namespace: "default",
				},
				Spec: FluxConfigSpec{
					Github: &GithubProviderConfig{
						Owner:      "janedoe",
						Repository: "flux-fleet",
					},
				},
			},
			clusterConfig: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.19 git",
			fileName: "testdata/cluster_1_19_flux_git.yaml",
			refName:  "test-flux-git",
			wantFluxConfig: &FluxConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       FluxConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-flux-git",
					Namespace: "default",
				},
				Spec: FluxConfigSpec{
					Git: &GitProviderConfig{
						RepositoryUrl: "https://git.com/test/test.git",
					},
				},
			},
			clusterConfig: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
			},
			wantErr:     false,
			gitProvider: true,
		},
		{
			testName: "refName doesn't match",
			fileName: "testdata/cluster_1_19_flux_github.yaml",
			refName:  "wrongName",
			wantErr:  true,
		},
		{
			testName:       "empty owner",
			fileName:       "testdata/cluster_invalid_flux_unset_gitowner.yaml",
			wantFluxConfig: nil,
			wantErr:        true,
		},
		{
			testName:       "empty repo",
			fileName:       "testdata/cluster_invalid_flux_unset_gitrepo.yaml",
			wantFluxConfig: nil,
			wantErr:        true,
		},
		{
			testName:       "empty username",
			fileName:       "testdata/cluster_invalid_flux_unset_gitusername.yaml",
			wantFluxConfig: nil,
			wantErr:        true,
		},
		{
			testName:       "empty repo url",
			fileName:       "testdata/cluster_invalid_flux_unset_gitrepourl.yaml",
			wantFluxConfig: nil,
			wantErr:        true,
		},
		{
			testName:       "invalid repo url",
			fileName:       "testdata/cluster_invalid_flux_gitrepourl.yaml",
			wantFluxConfig: nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.gitProvider {
				setupContext(t)
			}
			got, err := GetAndValidateFluxConfig(tt.fileName, tt.refName, tt.clusterConfig)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAndValidateFluxConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantFluxConfig) {
				t.Fatalf("GetAndValidateFluxConfig() = %#v, want %#v", got, tt.wantFluxConfig)
			}
		})
	}
}

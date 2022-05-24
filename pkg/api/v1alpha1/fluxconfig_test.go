package v1alpha1

import (
	"errors"
	"fmt"
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

func TestValidateFluxConfig(t *testing.T) {
	tests := []struct {
		testName    string
		fluxConfig  *FluxConfig
		wantErr     bool
		gitProvider bool
		error       error
	}{
		{
			testName: "valid fluxconfig github",
			fluxConfig: &FluxConfig{
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
			wantErr: false,
			error:   nil,
		},
		{
			testName: "valid fluxconfig git",
			fluxConfig: &FluxConfig{
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
						RepositoryUrl: "ssh://git@github.com/username/repo.git",
					},
				},
			},
			gitProvider: true,
			wantErr:     false,
			error:       nil,
		},
		{
			testName: "empty owner",
			fluxConfig: &FluxConfig{
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
						Repository: "flux-fleet",
					},
				},
			},
			wantErr: true,
			error:   errors.New("'owner' is not set or empty in githubProviderConfig; owner is a required field"),
		},
		{
			testName: "empty repo",
			fluxConfig: &FluxConfig{
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
						Owner: "janedoe",
					},
				},
			},
			wantErr: true,
			error:   errors.New("'repository' is not set or empty in githubProviderConfig; repository is a required field"),
		},
		{
			testName: "empty repo url",
			fluxConfig: &FluxConfig{
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
						RepositoryUrl: "",
					},
				},
			},
			wantErr: true,
			error:   errors.New("'repositoryUrl' is not set or empty in gitProviderConfig; repositoryUrl is a required field"),
		},
		{
			testName: "invalid repo url",
			fluxConfig: &FluxConfig{
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
						RepositoryUrl: "http://git@github.com/username/repo.git",
					},
				},
			},
			wantErr:     true,
			gitProvider: true,
			error:       fmt.Errorf("invalid repository url scheme: %s", "http"),
		},
		{
			testName: "invalid sshkey algo",
			fluxConfig: &FluxConfig{
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
						RepositoryUrl:   "ssh://git@github.com/username/repo.git",
						SshKeyAlgorithm: "invalid",
					},
				},
			},
			wantErr: true,
			error:   fmt.Errorf("'sshKeyAlgorithm' does not have a valid value in gitProviderConfig; sshKeyAlgorithm must be amongst %s, %s, %s", RsaAlgorithm, EcdsaAlgorithm, Ed25519Algorithm),
		},
		{
			testName: "valid ssh key algo",
			fluxConfig: &FluxConfig{
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
						RepositoryUrl:   "ssh://git@github.com/username/repo.git",
						SshKeyAlgorithm: RsaAlgorithm,
					},
				},
			},
			wantErr:     false,
			gitProvider: true,
			error:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.gitProvider {
				setupContext(t)
			}
			err := tt.fluxConfig.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("FluxConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.error != nil {
				if !reflect.DeepEqual(err, tt.error) {
					t.Fatalf("GetAndValidateFluxConfig() = %#v, want %#v", err, tt.error)
				}
			}
		})
	}
}

package executables_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/types"
)

type testKindOption func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption

func TestKindCreateBootstrapClusterSuccess(t *testing.T) {
	_, writer := test.NewWriter(t)

	clusterName := "test_cluster"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.VersionsBundles["1.19"] = versionBundle
		s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
			Pods: v1alpha1.Pods{
				CidrBlocks: []string{"1.1.1.1"},
			},
			Services: v1alpha1.Services{
				CidrBlocks: []string{"2.2.2.2"},
			},
		}
	})
	eksClusterName := "test_cluster-eks-a-cluster"
	kubeConfigFile := "test_cluster.kind.kubeconfig"
	kindImage := "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.2"

	// Initialize gomock
	mockCtrl := gomock.NewController(t)

	tests := []struct {
		name           string
		wantKubeconfig string
		env            map[string]string
		options        []testKindOption
		wantKindConfig string
	}{
		{
			name:           "No options",
			wantKubeconfig: kubeConfigFile,
			options:        nil,
			env:            map[string]string{},
			wantKindConfig: "testdata/kind_config.yaml",
		},
		{
			name:           "With env option",
			wantKubeconfig: kubeConfigFile,
			options: []testKindOption{
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithEnv(map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"})
				},
			},
			env:            map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"},
			wantKindConfig: "testdata/kind_config.yaml",
		},
		{
			name:           "With docker option",
			wantKubeconfig: kubeConfigFile,
			options: []testKindOption{
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithExtraDockerMounts()
				},
			},
			env:            map[string]string{},
			wantKindConfig: "testdata/kind_config_docker_mount_networking.yaml",
		},
		{
			name:           "With extra port mappings option",
			wantKubeconfig: kubeConfigFile,
			options: []testKindOption{
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithExtraPortMappings([]int{80, 443})
				},
			},
			env:            map[string]string{},
			wantKindConfig: "testdata/kind_config_extra_port_mappings.yaml",
		},
		{
			name:           "With docker option and env option",
			wantKubeconfig: kubeConfigFile,
			options: []testKindOption{
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithEnv(map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"})
				},
				func(k *executables.Kind) bootstrapper.BootstrapClusterClientOption {
					return k.WithExtraDockerMounts()
				},
			},
			env:            map[string]string{"ENV_VAR1": "VALUE1", "ENV_VAR2": "VALUE2"},
			wantKindConfig: "testdata/kind_config_docker_mount_networking.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			executable := mockexecutables.NewMockExecutable(mockCtrl)

			var (
				spec  *cluster.Spec
				image string
			)
			spec = clusterSpec
			image = kindImage

			executable.EXPECT().ExecuteWithEnv(
				ctx,
				tt.env,
				"create", "cluster", "--name", eksClusterName, "--kubeconfig", test.OfType("string"), "--image", image, "--config", test.OfType("string"),
			).Return(bytes.Buffer{}, nil).Times(1).Do(
				func(ctx context.Context, envs map[string]string, args ...string) (stdout bytes.Buffer, err error) {
					gotKindConfig := args[9]
					test.AssertFilesEquals(t, gotKindConfig, tt.wantKindConfig)

					return bytes.Buffer{}, nil
				},
			)

			k := executables.NewKind(executable, writer)
			gotKubeconfig, err := k.CreateBootstrapCluster(ctx, spec, testOptionsToBootstrapOptions(k, tt.options)...)
			if err != nil {
				t.Fatalf("CreateBootstrapCluster() error = %v, wantErr %v", err, nil)
			}

			if !strings.HasSuffix(gotKubeconfig, tt.wantKubeconfig) {
				t.Errorf("CreateBootstrapCluster() gotKubeconfig = %v, want to end with %v", gotKubeconfig, tt.wantKubeconfig)
			}
		})
	}
}

func TestKindCreateBootstrapClusterSuccessWithRegistryMirror(t *testing.T) {
	_, writer := test.NewWriter(t)

	clusterName := "test_cluster"
	eksClusterName := "test_cluster-eks-a-cluster"
	kubeConfigFile := "test_cluster.kind.kubeconfig"
	registryMirror := "registry-mirror.test"
	registryMirrorWithPort := net.JoinHostPort(registryMirror, constants.DefaultHttpsPort)

	// Initialize gomock
	mockCtrl := gomock.NewController(t)

	tests := []struct {
		name           string
		wantKubeconfig string
		env            map[string]string
		clusterSpec    *cluster.Spec
		options        []testKindOption
		wantFiles      map[string]string
	}{
		{
			name:           "With registry mirror option, no CA cert provided",
			wantKubeconfig: kubeConfigFile,
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = clusterName
				s.VersionsBundles["1.19"] = versionBundle
				s.Cluster.Spec.RegistryMirrorConfiguration = &v1alpha1.RegistryMirrorConfiguration{
					Endpoint: registryMirror,
					Port:     constants.DefaultHttpsPort,
					OCINamespaces: []v1alpha1.OCINamespace{
						{
							Registry:  "public.ecr.aws",
							Namespace: "eks-anywhere",
						},
					},
				}
				s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
					Pods: v1alpha1.Pods{
						CidrBlocks: []string{"1.1.1.1"},
					},
					Services: v1alpha1.Services{
						CidrBlocks: []string{"2.2.2.2"},
					},
				}
			}),
			env: map[string]string{},
			wantFiles: map[string]string{
				"kind_tmp.yaml":                               "testdata/kind_config_registry_mirror_insecure.yaml",
				"certs.d/public.ecr.aws/hosts.toml":           "testdata/hosts_toml_insecure_public_ecr_aws.toml",
				"certs.d/registry-mirror.test:443/hosts.toml": "testdata/hosts_toml_insecure_registry_mirror.toml",
			},
		},
		{
			name:           "With registry mirror option, with CA cert",
			wantKubeconfig: kubeConfigFile,
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = clusterName
				s.VersionsBundles["1.19"] = versionBundle
				s.Cluster.Spec.RegistryMirrorConfiguration = &v1alpha1.RegistryMirrorConfiguration{
					Endpoint:      registryMirror,
					Port:          constants.DefaultHttpsPort,
					CACertContent: "test",
				}
				s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
					Pods: v1alpha1.Pods{
						CidrBlocks: []string{"1.1.1.1"},
					},
					Services: v1alpha1.Services{
						CidrBlocks: []string{"2.2.2.2"},
					},
				}
			}),
			env: map[string]string{},
			wantFiles: map[string]string{
				"kind_tmp.yaml":                               "testdata/kind_config_registry_mirror_with_ca.yaml",
				"certs.d/public.ecr.aws/hosts.toml":           "testdata/hosts_toml_with_ca_public_ecr_aws.toml",
				"certs.d/public.ecr.aws/ca.crt":               "testdata/ca.crt",
				"certs.d/registry-mirror.test:443/hosts.toml": "testdata/hosts_toml_with_ca_registry_mirror.toml",
				"certs.d/registry-mirror.test:443/ca.crt":     "testdata/ca.crt",
			},
		},
		{
			name:           "With registry mirror option, with auth",
			wantKubeconfig: kubeConfigFile,
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = clusterName
				s.VersionsBundles["1.19"] = versionBundle
				s.Cluster.Spec.RegistryMirrorConfiguration = &v1alpha1.RegistryMirrorConfiguration{
					Endpoint: registryMirror,
					Port:     constants.DefaultHttpsPort,
					OCINamespaces: []v1alpha1.OCINamespace{
						{
							Registry:  "public.ecr.aws",
							Namespace: "eks-anywhere",
						},
						{
							Registry:  "783794618700.dkr.ecr.us-west-2.amazonaws.com",
							Namespace: "curated-packages",
						},
					},
					Authenticate: true,
				}
				s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
					Pods: v1alpha1.Pods{
						CidrBlocks: []string{"1.1.1.1"},
					},
					Services: v1alpha1.Services{
						CidrBlocks: []string{"2.2.2.2"},
					},
				}
			}),
			env: map[string]string{},
			wantFiles: map[string]string{
				"kind_tmp.yaml":                                           "testdata/kind_config_registry_mirror_with_auth.yaml",
				"certs.d/public.ecr.aws/hosts.toml":                       "testdata/hosts_toml_with_auth_public_ecr_aws.toml",
				"certs.d/783794618700.dkr.ecr.*.amazonaws.com/hosts.toml": "testdata/hosts_toml_with_auth_783794618700_dkr_ecr_star_amazonaws_com.toml",
				"certs.d/registry-mirror.test:443/hosts.toml":             "testdata/hosts_toml_with_auth_registry_mirror.toml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			executable := mockexecutables.NewMockExecutable(mockCtrl)

			var (
				spec  *cluster.Spec
				image string
			)
			spec = tt.clusterSpec
			registry := registryMirrorWithPort
			if r, ok := registrymirror.FromCluster(spec.Cluster).NamespacedRegistryMap[constants.DefaultCoreEKSARegistry]; ok {
				registry = r
			}
			image = fmt.Sprintf("%s/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.2", registry)

			if spec.Cluster.Spec.RegistryMirrorConfiguration.Authenticate {
				t.Setenv("REGISTRY_USERNAME", "username")
				t.Setenv("REGISTRY_PASSWORD", "password")
			}

			executable.EXPECT().ExecuteWithEnv(
				ctx,
				tt.env,
				"create", "cluster", "--name", eksClusterName, "--kubeconfig", test.OfType("string"), "--image", image, "--config", test.OfType("string"),
			).Return(bytes.Buffer{}, nil).Times(1).Do(
				func(ctx context.Context, envs map[string]string, args ...string) (stdout bytes.Buffer, err error) {
					gotKindConfig := args[9]

					for relativePath, expectedFile := range tt.wantFiles {
						var actualFile string
						if relativePath == "kind_tmp.yaml" {
							actualFile = gotKindConfig
						} else {
							// Registry config files are in cluster directory structure
							actualFile = filepath.Join(clusterName, "generated", relativePath)
						}
						test.AssertFilesEquals(t, actualFile, expectedFile)
					}

					return bytes.Buffer{}, nil
				},
			)

			k := executables.NewKind(executable, writer)
			gotKubeconfig, err := k.CreateBootstrapCluster(ctx, spec, testOptionsToBootstrapOptions(k, tt.options)...)
			if err != nil {
				t.Fatalf("CreateBootstrapCluster() error = %v, wantErr %v", err, nil)
			}

			if !strings.HasSuffix(gotKubeconfig, tt.wantKubeconfig) {
				t.Errorf("CreateBootstrapCluster() gotKubeconfig = %v, want to end with %v", gotKubeconfig, tt.wantKubeconfig)
			}
		})
	}
}

func TestKindCreateBootstrapClusterExecutableError(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "clusterName"
		s.VersionsBundles["1.19"] = versionBundle
	})

	ctx := context.Background()
	_, writer := test.NewWriter(t)

	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().ExecuteWithEnv(ctx, map[string]string{}, gomock.Any()).Return(bytes.Buffer{}, errors.New("error from execute with env"))
	k := executables.NewKind(executable, writer)
	gotKubeconfig, err := k.CreateBootstrapCluster(ctx, clusterSpec)
	if err == nil {
		t.Fatal("Kind.CreateBootstrapCluster() error = nil")
	}

	if gotKubeconfig != "" {
		t.Errorf("CreateBootstrapCluster() gotKubeconfig = %v, want empty string", gotKubeconfig)
	}
}

func TestKindCreateBootstrapClusterExecutableWithRegistryMirrorError(t *testing.T) {
	registryMirror := "registry-mirror.test"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "clusterName"
		s.VersionsBundles["1.19"] = versionBundle
		s.Cluster.Spec.RegistryMirrorConfiguration = &v1alpha1.RegistryMirrorConfiguration{
			Endpoint:     registryMirror,
			Port:         constants.DefaultHttpsPort,
			Authenticate: true,
		}
	})

	if err := os.Unsetenv("REGISTRY_USERNAME"); err != nil {
		t.Fatalf(err.Error())
	}
	if err := os.Unsetenv("REGISTRY_PASSWORD"); err != nil {
		t.Fatalf(err.Error())
	}

	ctx := context.Background()
	_, writer := test.NewWriter(t)

	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	k := executables.NewKind(executable, writer)
	gotKubeconfig, err := k.CreateBootstrapCluster(ctx, clusterSpec)
	if err == nil {
		t.Fatal("Kind.CreateBootstrapCluster() error = nil")
	}

	if gotKubeconfig != "" {
		t.Errorf("CreateBootstrapCluster() gotKubeconfig = %v, want empty string", gotKubeconfig)
	}
}

func testOptionsToBootstrapOptions(k *executables.Kind, testOpts []testKindOption) []bootstrapper.BootstrapClusterClientOption {
	opts := make([]bootstrapper.BootstrapClusterClientOption, 0, len(testOpts))
	for _, opt := range testOpts {
		opts = append(opts, opt(k))
	}

	return opts
}

func TestKindDeleteBootstrapClusterSuccess(t *testing.T) {
	cluster := &types.Cluster{
		Name: "clusterName",
	}
	ctx := context.Background()
	_, writer := test.NewWriter(t)
	internalName := fmt.Sprintf("%s-eks-a-cluster", cluster.Name)

	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	expectedParam := []string{"delete", "cluster", "--name", internalName}
	executable.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	k := executables.NewKind(executable, writer)
	if err := k.DeleteBootstrapCluster(ctx, cluster); err != nil {
		t.Fatalf("Kind.DeleteBootstrapCluster() error = %v, want nil", err)
	}
}

func TestKindDeleteBootstrapClusterExecutableError(t *testing.T) {
	cluster := &types.Cluster{
		Name: "clusterName",
	}
	ctx := context.Background()
	_, writer := test.NewWriter(t)
	internalName := fmt.Sprintf("%s-eks-a-cluster", cluster.Name)

	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	expectedParam := []string{"delete", "cluster", "--name", internalName}
	executable.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	k := executables.NewKind(executable, writer)
	if err := k.DeleteBootstrapCluster(ctx, cluster); err == nil {
		t.Fatalf("Kind.DeleteBootstrapCluster() error = %v, want not nil", err)
	}
}

func TestKindClusterExists(t *testing.T) {
	tests := []struct {
		testName     string
		clusterName  string
		internalName string
		kindResponse string
	}{
		{
			testName:     "one cluster",
			clusterName:  "cluster-name",
			internalName: "cluster-name-eks-a-cluster",
			kindResponse: "cluster-name-eks-a-cluster",
		},
		{
			testName:     "3 clusters",
			clusterName:  "cluster-name-2",
			internalName: "cluster-name-2-eks-a-cluster",
			kindResponse: "cluster-name-eks-a-cluster\ncluster-name-eks-a-cluster-6\ncluster-name-2-eks-a-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			_, writer := test.NewWriter(t)

			mockCtrl := gomock.NewController(t)
			executable := mockexecutables.NewMockExecutable(mockCtrl)
			executable.EXPECT().Execute(ctx, "get", "clusters").Return(*bytes.NewBufferString(tt.kindResponse), nil)
			k := executables.NewKind(executable, writer)
			clusterExists, err := k.ClusterExists(ctx, tt.clusterName)
			if err != nil {
				t.Fatalf("Kind.ClusterExists() error = %v, wantErr nil", err)
			}

			if !clusterExists {
				t.Fatal("ClusterExists() clusterExists = false, want true")
			}
		})
	}
}

func TestKindGetKubeconfig(t *testing.T) {
	clusterName := "cluster-name"
	ctx := context.Background()
	_, writer := test.NewWriter(t)

	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "get", "kubeconfig", "--name", fmt.Sprintf("%s-eks-a-cluster", clusterName)).Return(bytes.Buffer{}, nil)
	k := executables.NewKind(executable, writer)
	_, err := k.GetKubeconfig(ctx, clusterName)
	if err != nil {
		t.Fatalf("Kind.GetKubeconfig() error = %v, wantErr nil", err)
	}
}

func TestKindCreateAuditPolicyWriteFailure(t *testing.T) {
	// Create a temporary directory that will be used by the test
	tempDir, err := os.MkdirTemp("", "kind-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	clusterName := "write-fail-cluster"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.VersionsBundles["1.19"] = versionBundle
		s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
			Pods: v1alpha1.Pods{
				CidrBlocks: []string{"1.1.1.1"},
			},
			Services: v1alpha1.Services{
				CidrBlocks: []string{"2.2.2.2"},
			},
		}
	})

	// Create the kubernetes directory with normal permissions
	kubernetesDir := filepath.Join(tempDir, clusterName, "generated", "kubernetes")
	if err := os.MkdirAll(kubernetesDir, 0o755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create a dummy audit-policy.yaml file
	auditPolicyPath := filepath.Join(kubernetesDir, "audit-policy.yaml")
	if err := os.WriteFile(auditPolicyPath, []byte("dummy content"), 0o644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Make the file read-only to force a write error
	// This will allow directory creation to succeed but file writing to fail
	if err := os.Chmod(auditPolicyPath, 0o444); err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}

	// Override the cluster name to point to our test directory
	clusterSpec.Cluster.Name = filepath.Join(tempDir, clusterName)

	ctx := context.Background()
	_, writer := test.NewWriter(t)
	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)

	k := executables.NewKind(executable, writer)

	_, err = k.CreateBootstrapCluster(ctx, clusterSpec)

	if err == nil {
		t.Fatal("Expected an error when trying to write to a read-only file, but got nil")
	}
}

func TestKindCreateAuditPolicyMkdirFailure(t *testing.T) {
	// Create a temporary directory that will be used by the test
	tempDir, err := os.MkdirTemp("", "kind-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	clusterName := "mkdir-fail-cluster"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.VersionsBundles["1.19"] = versionBundle
		s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
			Pods: v1alpha1.Pods{
				CidrBlocks: []string{"1.1.1.1"},
			},
			Services: v1alpha1.Services{
				CidrBlocks: []string{"2.2.2.2"},
			},
		}
	})

	// Create the parent directory structure
	parentDir := filepath.Join(tempDir, clusterName, "generated")
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}

	// Create a file instead of a directory at the kubernetes path
	// This will cause MkdirAll to fail because it can't create a directory where a file exists
	kubernetesFilePath := filepath.Join(parentDir, "kubernetes")
	if err := os.WriteFile(kubernetesFilePath, []byte("This is a file, not a directory"), 0o644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	clusterSpec.Cluster.Name = filepath.Join(tempDir, clusterName)

	ctx := context.Background()
	_, writer := test.NewWriter(t)
	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)

	k := executables.NewKind(executable, writer)

	_, err = k.CreateBootstrapCluster(ctx, clusterSpec)

	if err == nil {
		t.Fatal("Expected an error when trying to create a directory where a file exists, but got nil")
	}
}

func TestKindCreateBootstrapClusterAuditPolicyError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "kind-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	clusterName := "audit-policy-error-cluster"
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = clusterName
		s.VersionsBundles["1.19"] = versionBundle
		s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
			Pods: v1alpha1.Pods{
				CidrBlocks: []string{"1.1.1.1"},
			},
			Services: v1alpha1.Services{
				CidrBlocks: []string{"2.2.2.2"},
			},
		}
	})

	// Create the parent directory structure
	parentDir := filepath.Join(tempDir, clusterName, "generated")
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}

	// Create a file instead of a directory at the kubernetes path to force an error
	kubernetesPath := filepath.Join(parentDir, "kubernetes")
	if err := os.WriteFile(kubernetesPath, []byte("This is a file, not a directory"), 0o644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Override the cluster name to point to our test directory
	clusterSpec.Cluster.Name = filepath.Join(tempDir, clusterName)

	ctx := context.Background()

	_, writer := test.NewWriter(t)

	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)

	k := executables.NewKind(executable, writer)

	_, err = k.CreateBootstrapCluster(ctx, clusterSpec)

	if err == nil {
		t.Fatal("Expected an error when CreateAuditPolicy fails, but got nil")
	}
}

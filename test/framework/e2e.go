package framework

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	defaultClusterConfigFile         = "cluster.yaml"
	defaultBundleReleaseManifestFile = "bin/local-bundle-release.yaml"
	defaultClusterName               = "eksa-test"
	ClusterNameVar                   = "T_CLUSTER_NAME"
	JobIdVar                         = "T_JOB_ID"
	BundlesOverrideVar               = "T_BUNDLES_OVERRIDE"
)

//go:embed testdata/oidc-roles.yaml
var oidcRoles []byte

type E2ETest struct {
	T                     *testing.T
	ClusterConfigLocation string
	ClusterName           string
	ClusterConfig         *v1alpha1.Cluster
	Provider              Provider
	ClusterConfigB        []byte
	ProviderConfigB       []byte
	clusterFillers        []api.ClusterFiller
	KubectlClient         *executables.Kubectl
	GitProvider           git.Provider
	GitWriter             filewriter.FileWriter
	OIDCConfig            *v1alpha1.OIDCConfig
	GitOpsConfig          *v1alpha1.GitOpsConfig
	ProxyConfig           *v1alpha1.ProxyConfiguration
}

type E2ETestOpt func(e *E2ETest)

func NewE2ETest(t *testing.T, provider Provider, opts ...E2ETestOpt) *E2ETest {
	e := &E2ETest{
		T:                     t,
		Provider:              provider,
		ClusterConfigLocation: defaultClusterConfigFile,
		ClusterName:           getClusterName(),
		clusterFillers:        make([]api.ClusterFiller, 0),
		KubectlClient:         buildKubectl(t),
	}

	for _, opt := range opts {
		opt(e)
	}

	provider.Setup()

	return e
}

func WithClusterFiller(f api.ClusterFiller) E2ETestOpt {
	return func(e *E2ETest) {
		e.clusterFillers = append(e.clusterFillers, f)
	}
}

func WithClusterConfigLocationOverride(path string) E2ETestOpt {
	return func(e *E2ETest) {
		e.ClusterConfigLocation = path
	}
}

type Provider interface {
	Name() string
	CustomizeProviderConfig(file string) []byte
	ClusterConfigFillers() []api.ClusterFiller
	Setup()
}

func (e *E2ETest) GenerateClusterConfig() {
	e.RunEKSA("anywhere", "generate", "clusterconfig", e.ClusterName, "-p", e.Provider.Name(), ">", e.ClusterConfigLocation)

	clusterFillersFromProvider := e.Provider.ClusterConfigFillers()
	clusterConfigFillers := make([]api.ClusterFiller, 0, len(e.clusterFillers)+len(clusterFillersFromProvider))
	clusterConfigFillers = append(clusterConfigFillers, e.clusterFillers...)
	clusterConfigFillers = append(clusterConfigFillers, clusterFillersFromProvider...)
	e.ClusterConfigB = e.customizeClusterConfig(clusterConfigFillers...)
	e.ProviderConfigB = e.Provider.CustomizeProviderConfig(e.ClusterConfigLocation)
	e.buildClusterConfigFile()
	e.cleanup(func() {
		os.Remove(e.ClusterConfigLocation)
	})
}

func (e *E2ETest) ImportImages() {
	importImagesArgs := []string{"anywhere", "import-images", "-f", e.ClusterConfigLocation}
	if getBundlesOverride() == "true" {
		importImagesArgs = append(importImagesArgs, "--bundles-override", defaultBundleReleaseManifestFile)
	}
	e.RunEKSA(importImagesArgs...)
}

func (e *E2ETest) CreateCluster() {
	createClusterArgs := []string{"anywhere", "create", "cluster", "-f", e.ClusterConfigLocation, "-v", "4"}
	if getBundlesOverride() == "true" {
		createClusterArgs = append(createClusterArgs, "--bundles-override", defaultBundleReleaseManifestFile)
	}

	e.RunEKSA(createClusterArgs...)
	e.cleanup(func() {
		os.RemoveAll(e.ClusterName)
	})
}

func (e *E2ETest) ValidateCluster(kubeVersion v1alpha1.KubernetesVersion) {
	ctx := context.Background()
	e.T.Log("Validating cluster node status")
	r := retrier.New(10 * time.Minute)
	err := r.Retry(func() error {
		err := e.KubectlClient.ValidateNodes(ctx, e.cluster().KubeconfigFile)
		if err != nil {
			return fmt.Errorf("error validating nodes status: %v", err)
		}
		return nil
	})
	if err != nil {
		e.T.Fatalf("%v", err)
	}
	e.T.Log("Validating cluster node version")
	err = retrier.Retry(180, 1*time.Second, func() error {
		if err = e.KubectlClient.ValidateNodesVersion(ctx, e.cluster().KubeconfigFile, kubeVersion); err != nil {
			return fmt.Errorf("error validating nodes version: %v", err)
		}
		return nil
	})
	if err != nil {
		e.T.Fatal(err)
	}
}

func WithClusterUpgrade(fillers ...api.ClusterFiller) E2ETestOpt {
	return func(e *E2ETest) {
		e.ClusterConfigB = e.customizeClusterConfig(fillers...)
	}
}

func (e *E2ETest) UpgradeCluster(opts ...E2ETestOpt) {
	for _, opt := range opts {
		opt(e)
	}
	e.buildClusterConfigFile()

	upgradeClusterArgs := []string{"anywhere", "upgrade", "cluster", "-f", e.ClusterConfigLocation, "-v", "4"}
	if getBundlesOverride() == "true" {
		upgradeClusterArgs = append(upgradeClusterArgs, "--bundles-override", defaultBundleReleaseManifestFile)
	}
	e.RunEKSA(upgradeClusterArgs...)
}

func (e *E2ETest) buildClusterConfigFile() {
	yamlB := make([][]byte, 0, 4)
	yamlB = append(yamlB, e.ClusterConfigB, e.ProviderConfigB)
	if e.OIDCConfig != nil {
		oidcConfigB, err := yaml.Marshal(e.OIDCConfig)
		if err != nil {
			e.T.Fatalf("error marshalling oidc config: %v", err)
		}
		yamlB = append(yamlB, oidcConfigB)
	}
	if e.GitOpsConfig != nil {
		gitOpsConfigB, err := yaml.Marshal(e.GitOpsConfig)
		if err != nil {
			e.T.Fatalf("error marshalling gitops config: %v", err)
		}
		yamlB = append(yamlB, gitOpsConfigB)
	}
	writer, err := filewriter.NewWriter(filepath.Dir(e.ClusterConfigLocation))
	if err != nil {
		e.T.Fatalf("Error creating writer: %v", err)
	}

	b := templater.AppendYamlResources(yamlB...)

	writtenFile, err := writer.Write(filepath.Base(e.ClusterConfigLocation), b)
	if err != nil {
		e.T.Fatalf("Error writing cluster config to file %s: %v", e.ClusterConfigLocation, err)
	}

	e.ClusterConfigLocation = writtenFile
}

func (e *E2ETest) DeleteCluster() {
	deleteClusterArgs := []string{"anywhere", "delete", "cluster", e.ClusterName, "-v", "4"}
	if getBundlesOverride() == "true" {
		deleteClusterArgs = append(deleteClusterArgs, "--bundles-override", defaultBundleReleaseManifestFile)
	}
	e.RunEKSA(deleteClusterArgs...)
}

func (e *E2ETest) Run(name string, args ...string) {
	command := strings.Join(append([]string{name}, args...), " ")
	shArgs := []string{"-c", command}

	e.T.Log("Running shell command", "[", command, "]")
	cmd := exec.CommandContext(context.Background(), "sh", shArgs...)

	envPath := os.Getenv("PATH")
	workDir, err := os.Getwd()
	if err != nil {
		e.T.Fatalf("Error finding current directory: %v", err)
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s/bin:%s", workDir, envPath))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		e.T.Fatalf("Error running command %s %v: %v", name, args, err)
	}
}

func (e *E2ETest) RunEKSA(args ...string) {
	e.Run("eksctl", args...)
}

func (e *E2ETest) StopIfFailed() {
	if e.T.Failed() {
		e.T.FailNow()
	}
}

func (e *E2ETest) customizeClusterConfig(fillers ...api.ClusterFiller) []byte {
	b, err := api.AutoFillCluster(e.ClusterConfigLocation, fillers...)
	if err != nil {
		e.T.Fatalf("Error filling cluster config: %v", err)
	}

	return b
}

func (e *E2ETest) cleanup(f func()) {
	e.T.Cleanup(func() {
		if !e.T.Failed() {
			f()
		}
	})
}

func (e *E2ETest) cluster() *types.Cluster {
	return &types.Cluster{
		Name:           e.ClusterName,
		KubeconfigFile: e.kubeconfigFilePath(),
	}
}

func (e *E2ETest) kubeconfigFilePath() string {
	return filepath.Join(e.ClusterName, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", e.ClusterName))
}

func (e *E2ETest) GetEksaVSphereMachineConfigs() []v1alpha1.VSphereMachineConfig {
	clusterConfig := e.clusterConfig()
	machineConfigNames := make([]string, 0, len(clusterConfig.Spec.WorkerNodeGroupConfigurations)+1)
	machineConfigNames = append(machineConfigNames, clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	for _, workerNodeConf := range clusterConfig.Spec.WorkerNodeGroupConfigurations {
		machineConfigNames = append(machineConfigNames, workerNodeConf.MachineGroupRef.Name)
	}

	kubeconfig := e.kubeconfigFilePath()
	ctx := context.Background()

	machineConfigs := make([]v1alpha1.VSphereMachineConfig, 0, len(machineConfigNames))
	for _, name := range machineConfigNames {
		m, err := e.KubectlClient.GetEksaVSphereMachineConfig(ctx, name, kubeconfig, clusterConfig.Namespace)
		if err != nil {
			e.T.Fatalf("Failed getting VSphereMachineConfig: %v", err)
		}

		machineConfigs = append(machineConfigs, *m)
	}

	return machineConfigs
}

func (e *E2ETest) clusterConfig() *v1alpha1.Cluster {
	if e.ClusterConfig != nil {
		return e.ClusterConfig
	}

	c, err := v1alpha1.GetClusterConfig(e.ClusterConfigLocation)
	if err != nil {
		e.T.Fatalf("Error fetching cluster config from file: %v", err)
	}
	e.ClusterConfig = c

	return e.ClusterConfig
}

func (e *E2ETest) getJobIdFromEnv() string {
	return os.Getenv(JobIdVar)
}

func getClusterName() string {
	value := os.Getenv(ClusterNameVar)
	if len(value) == 0 {
		return defaultClusterName
	}
	return value
}

func getBundlesOverride() string {
	return os.Getenv(BundlesOverrideVar)
}

package framework

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	_ "embed"
	"fmt"
	"io"
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
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/pbnj"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	defaultClusterConfigFile         = "cluster.yaml"
	defaultBundleReleaseManifestFile = "bin/local-bundle-release.yaml"
	defaultEksaBinaryLocation        = "eksctl anywhere"
	defaultClusterName               = "eksa-test"
	eksctlVersionEnvVar              = "EKSCTL_VERSION"
	eksctlVersionEnvVarDummyVal      = "ham sandwich"
	ClusterNameVar                   = "T_CLUSTER_NAME"
	JobIdVar                         = "T_JOB_ID"
	BundlesOverrideVar               = "T_BUNDLES_OVERRIDE"
	hardwareYamlPath                 = "hardware.yaml"
	hardwareCsvPath                  = "hardware.csv"
)

//go:embed testdata/oidc-roles.yaml
var oidcRoles []byte

type ClusterE2ETest struct {
	T                      *testing.T
	ClusterConfigLocation  string
	ClusterConfigFolder    string
	HardwareConfigLocation string
	HardwareCsvLocation    string
	TestHardware           map[string]*api.Hardware
	HardwarePool           map[string]*api.Hardware
	ClusterName            string
	ClusterConfig          *v1alpha1.Cluster
	Provider               Provider
	ClusterConfigB         []byte
	ProviderConfigB        []byte
	clusterFillers         []api.ClusterFiller
	KubectlClient          *executables.Kubectl
	GitProvider            git.Provider
	GitWriter              filewriter.FileWriter
	OIDCConfig             *v1alpha1.OIDCConfig
	GitOpsConfig           *v1alpha1.GitOpsConfig
	ProxyConfig            *v1alpha1.ProxyConfiguration
	AWSIamConfig           *v1alpha1.AWSIamConfig
	eksaBinaryLocation     string
}

type ClusterE2ETestOpt func(e *ClusterE2ETest)

func NewClusterE2ETest(t *testing.T, provider Provider, opts ...ClusterE2ETestOpt) *ClusterE2ETest {
	e := &ClusterE2ETest{
		T:                     t,
		Provider:              provider,
		ClusterConfigLocation: defaultClusterConfigFile,
		ClusterName:           getClusterName(t),
		clusterFillers:        make([]api.ClusterFiller, 0),
		KubectlClient:         buildKubectl(t),
		eksaBinaryLocation:    defaultEksaBinaryLocation,
	}

	e.ClusterConfigFolder = fmt.Sprintf("%s-config", e.ClusterName)
	e.HardwareConfigLocation = filepath.Join(e.ClusterConfigFolder, hardwareYamlPath)
	e.HardwareCsvLocation = filepath.Join(e.ClusterConfigFolder, hardwareCsvPath)

	for _, opt := range opts {
		opt(e)
	}

	provider.Setup()

	return e
}

func WithHardware(vendor string, requiredCount int) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		hardwarePool := e.GetHardwarePool()

		if e.TestHardware == nil {
			e.TestHardware = make(map[string]*api.Hardware)
		}

		var count int
		for id, h := range hardwarePool {
			if strings.ToLower(h.BmcVendor) == vendor || vendor == api.HardwareVendorUnspecified {
				if _, exists := e.TestHardware[id]; !exists {
					count++
					e.TestHardware[id] = h
				}

				if count == requiredCount {
					break
				}
			}
		}

		if count < requiredCount {
			e.T.Errorf("this test requires at least %d piece(s) of %s hardware", requiredCount, vendor)
		}
	}
}

func (e *ClusterE2ETest) GetHardwarePool() map[string]*api.Hardware {
	if e.HardwarePool == nil {
		csvFilePath := os.Getenv(tinkerbellInventoryCsvFilePathEnvVar)
		var err error
		e.HardwarePool, err = api.NewHardwareMapFromFile(csvFilePath)
		if err != nil {
			e.T.Fatalf("failed to create hardware map from test hardware pool: %v", err)
		}
	}
	return e.HardwarePool
}

func WithClusterFiller(f ...api.ClusterFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.clusterFillers = append(e.clusterFillers, f...)
	}
}

func WithClusterConfigLocationOverride(path string) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ClusterConfigLocation = path
	}
}

func WithEksaVersion(version *semver.Version) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		eksaBinaryLocation, err := GetReleaseBinaryFromVersion(version)
		if err != nil {
			e.T.Fatal(err)
		}
		e.eksaBinaryLocation = eksaBinaryLocation
		err = setEksctlVersionEnvVar()
		if err != nil {
			e.T.Fatal(err)
		}
	}
}

func WithLatestMinorReleaseFromMain() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		eksaBinaryLocation, err := GetLatestMinorReleaseBinaryFromMain()
		if err != nil {
			e.T.Fatal(err)
		}
		e.eksaBinaryLocation = eksaBinaryLocation
		err = setEksctlVersionEnvVar()
		if err != nil {
			e.T.Fatal(err)
		}
	}
}

func WithLatestMinorReleaseFromVersion(version *semver.Version) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		eksaBinaryLocation, err := GetLatestMinorReleaseBinaryFromVersion(version)
		if err != nil {
			e.T.Fatal(err)
		}
		e.eksaBinaryLocation = eksaBinaryLocation
		err = setEksctlVersionEnvVar()
		if err != nil {
			e.T.Fatal(err)
		}
	}
}

func WithEnvVar(key, val string) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		err := os.Setenv(key, val)
		if err != nil {
			e.T.Fatalf("couldn't set env var %s to value %s due to: %v", key, val, err)
		}
	}
}

type Provider interface {
	Name() string
	CustomizeProviderConfig(file string) []byte
	ClusterConfigFillers() []api.ClusterFiller
	Setup()
}

func (e *ClusterE2ETest) GenerateClusterConfig(opts ...CommandOpt) {
	e.GenerateClusterConfigForVersion("", opts...)
}

func (e *ClusterE2ETest) PowerOffHardware() {
	pbnjEndpoint := os.Getenv(tinkerbellPBnJGRPCAuthEnvVar)
	pbnjClient, err := pbnj.NewPBNJClient(pbnjEndpoint)
	if err != nil {
		e.T.Fatalf("failed to create pbnj client: %v", err)
	}

	ctx := context.Background()

	for _, h := range e.TestHardware {
		bmcInfo := api.NewBmcSecretConfig(h)
		err := pbnjClient.PowerOff(ctx, bmcInfo)
		if err != nil {
			e.T.Fatalf("failed to power off hardware: %v", err)
		}
	}
}

func (e *ClusterE2ETest) ValidateHardwareDecommissioned() {
	pbnjEndpoint := os.Getenv(tinkerbellPBnJGRPCAuthEnvVar)
	pbnjClient, err := pbnj.NewPBNJClient(pbnjEndpoint)
	if err != nil {
		e.T.Fatalf("failed to create pbnj client: %v", err)
	}

	ctx := context.Background()

	var failedToDecomm []*api.Hardware
	for _, h := range e.TestHardware {
		bmcInfo := api.NewBmcSecretConfig(h)

		powerState, err := pbnjClient.GetPowerState(ctx, bmcInfo)
		if err != nil {
			e.T.Logf("failed to get power state for hardware (%v): %v", h, err)
		}

		if powerState != pbnj.PowerStateOff {
			e.T.Logf("failed to decommission hardware: id=%s, hostname=%s, bmc_ip=%s", h.Id, h.Hostname, h.BmcIpAddress)
			failedToDecomm = append(failedToDecomm, h)
		}
	}

	if len(failedToDecomm) > 0 {
		e.T.Fatalf("failed to decommision hardware during cluster deletion")
	}
}

func (e *ClusterE2ETest) GenerateHardwareConfig(opts ...CommandOpt) {
	e.generateHardwareConfig(opts...)
}

func (e *ClusterE2ETest) generateHardwareConfig(opts ...CommandOpt) {
	if len(e.TestHardware) == 0 {
		e.T.Fatal("you must provide the ClusterE2ETest the hardware to use for the test run")
	}

	if _, err := os.Stat(e.HardwareCsvLocation); err == nil {
		os.Remove(e.HardwareCsvLocation)
	}

	err := api.WriteHardwareMapToCSV(e.TestHardware, e.HardwareCsvLocation)
	if err != nil {
		e.T.Fatalf("failed to create hardware csv for the test run: %v", err)
	}

	generateHardwareConfigArgs := []string{
		"generate", "hardware",
		"--skip-registration",
		"-f", e.HardwareCsvLocation,
		"-o", e.ClusterConfigFolder,
	}

	e.RunEKSA(generateHardwareConfigArgs, opts...)
}

func (e *ClusterE2ETest) GenerateClusterConfigForVersion(eksaVersion string, opts ...CommandOpt) {
	e.generateClusterConfigObjects(opts...)
	if eksaVersion != "" {
		var err error
		e.ClusterConfigB, err = cleanUpClusterForVersion(e.ClusterConfigB, eksaVersion)
		if err != nil {
			e.T.Fatal(err)
		}
	}

	e.buildClusterConfigFile()
	e.cleanup(func() {
		os.Remove(e.ClusterConfigLocation)
	})
}

func (e *ClusterE2ETest) generateClusterConfigObjects(opts ...CommandOpt) {
	generateClusterConfigArgs := []string{"generate", "clusterconfig", e.ClusterName, "-p", e.Provider.Name(), ">", e.ClusterConfigLocation}
	e.RunEKSA(generateClusterConfigArgs, opts...)

	clusterFillersFromProvider := e.Provider.ClusterConfigFillers()
	clusterConfigFillers := make([]api.ClusterFiller, 0, len(e.clusterFillers)+len(clusterFillersFromProvider))
	clusterConfigFillers = append(clusterConfigFillers, e.clusterFillers...)
	clusterConfigFillers = append(clusterConfigFillers, clusterFillersFromProvider...)
	e.ClusterConfigB = e.customizeClusterConfig(e.ClusterConfigLocation, clusterConfigFillers...)
	e.ProviderConfigB = e.Provider.CustomizeProviderConfig(e.ClusterConfigLocation)
}

func (e *ClusterE2ETest) ImportImages(opts ...CommandOpt) {
	importImagesArgs := []string{"import-images", "-f", e.ClusterConfigLocation}
	e.RunEKSA(importImagesArgs, opts...)
}

func (e *ClusterE2ETest) CreateCluster(opts ...CommandOpt) {
	e.createCluster(opts...)
}

func (e *ClusterE2ETest) createCluster(opts ...CommandOpt) {
	e.T.Logf("Creating cluster %s", e.ClusterName)
	createClusterArgs := []string{"create", "cluster", "-f", e.ClusterConfigLocation, "-v", "4"}
	if getBundlesOverride() == "true" {
		createClusterArgs = append(createClusterArgs, "--bundles-override", defaultBundleReleaseManifestFile)
	}

	if e.Provider.Name() == TinkerbellProviderName {
		createClusterArgs = append(createClusterArgs, "-w", e.HardwareConfigLocation)
	}

	e.RunEKSA(createClusterArgs, opts...)
	e.cleanup(func() {
		os.RemoveAll(e.ClusterName)
	})
}

func (e *ClusterE2ETest) ValidateCluster(kubeVersion v1alpha1.KubernetesVersion) {
	ctx := context.Background()
	e.T.Log("Validating cluster node status")
	r := retrier.New(10 * time.Minute)
	err := r.Retry(func() error {
		err := e.KubectlClient.ValidateNodes(ctx, e.cluster().KubeconfigFile)
		if err != nil {
			return fmt.Errorf("validating nodes status: %v", err)
		}
		return nil
	})
	if err != nil {
		e.T.Fatal(err)
	}
	e.T.Log("Validating cluster node version")
	err = retrier.Retry(180, 1*time.Second, func() error {
		if err = e.KubectlClient.ValidateNodesVersion(ctx, e.cluster().KubeconfigFile, kubeVersion); err != nil {
			return fmt.Errorf("validating nodes version: %v", err)
		}
		return nil
	})
	if err != nil {
		e.T.Fatal(err)
	}
}

func WithClusterUpgrade(fillers ...api.ClusterFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ClusterConfigB = e.customizeClusterConfig(e.ClusterConfigLocation, fillers...)
	}
}

func (e *ClusterE2ETest) UpgradeCluster(clusterOpts []ClusterE2ETestOpt, commandOpts ...CommandOpt) {
	e.upgradeCluster(clusterOpts, commandOpts...)
}

func (e *ClusterE2ETest) upgradeCluster(clusterOpts []ClusterE2ETestOpt, commandOpts ...CommandOpt) {
	for _, opt := range clusterOpts {
		opt(e)
	}
	e.buildClusterConfigFile()

	upgradeClusterArgs := []string{"upgrade", "cluster", "-f", e.ClusterConfigLocation, "-v", "4"}
	if getBundlesOverride() == "true" {
		upgradeClusterArgs = append(upgradeClusterArgs, "--bundles-override", defaultBundleReleaseManifestFile)
	}

	e.RunEKSA(upgradeClusterArgs, commandOpts...)
}

func (e *ClusterE2ETest) generateClusterConfig() []byte {
	yamlB := make([][]byte, 0, 4)
	yamlB = append(yamlB, e.ClusterConfigB, e.ProviderConfigB)
	if e.OIDCConfig != nil {
		oidcConfigB, err := yaml.Marshal(e.OIDCConfig)
		if err != nil {
			e.T.Fatalf("error marshalling oidc config: %v", err)
		}
		yamlB = append(yamlB, oidcConfigB)
	}
	if e.AWSIamConfig != nil {
		awsIamConfigB, err := yaml.Marshal(e.AWSIamConfig)
		if err != nil {
			e.T.Fatalf("error marshalling aws iam config: %v", err)
		}
		yamlB = append(yamlB, awsIamConfigB)
	}
	if e.GitOpsConfig != nil {
		gitOpsConfigB, err := yaml.Marshal(e.GitOpsConfig)
		if err != nil {
			e.T.Fatalf("error marshalling gitops config: %v", err)
		}
		yamlB = append(yamlB, gitOpsConfigB)
	}

	return templater.AppendYamlResources(yamlB...)
}

func (e *ClusterE2ETest) buildClusterConfigFile() {
	b := e.generateClusterConfig()

	writer, err := filewriter.NewWriter(e.ClusterConfigFolder)
	if err != nil {
		e.T.Fatalf("Error creating writer: %v", err)
	}

	writtenFile, err := writer.Write(filepath.Base(e.ClusterConfigLocation), b, filewriter.PersistentFile)
	if err != nil {
		e.T.Fatalf("Error writing cluster config to file %s: %v", e.ClusterConfigLocation, err)
	}
	e.ClusterConfigLocation = writtenFile
}

func (e *ClusterE2ETest) DeleteCluster(opts ...CommandOpt) {
	e.deleteCluster(opts...)
}

func (e *ClusterE2ETest) deleteCluster(opts ...CommandOpt) {
	deleteClusterArgs := []string{"delete", "cluster", e.ClusterName, "-v", "4"}
	if getBundlesOverride() == "true" {
		deleteClusterArgs = append(deleteClusterArgs, "--bundles-override", defaultBundleReleaseManifestFile)
	}
	e.RunEKSA(deleteClusterArgs, opts...)
}

func (e *ClusterE2ETest) Run(name string, args ...string) {
	command := strings.Join(append([]string{name}, args...), " ")
	shArgs := []string{"-c", command}

	e.T.Log("Running shell command", "[", command, "]")
	cmd := exec.CommandContext(context.Background(), "sh", shArgs...)

	envPath := os.Getenv("PATH")

	workDir, err := os.Getwd()
	if err != nil {
		e.T.Fatalf("Error finding current directory: %v", err)
	}

	var stdoutAndErr bytes.Buffer

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s/bin:%s", workDir, envPath))
	cmd.Stderr = io.MultiWriter(os.Stderr, &stdoutAndErr)
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutAndErr)

	if err = cmd.Run(); err != nil {
		scanner := bufio.NewScanner(&stdoutAndErr)
		var errorMessage string
		// Look for the last line of the out put that starts with 'Error:'
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Error:") {
				errorMessage = line
			}
		}

		if err := scanner.Err(); err != nil {
			e.T.Fatalf("Failed reading command output looking for error message: %v", err)
		}

		if errorMessage != "" {
			e.T.Fatalf("Command %s %v failed with error: %v: %s", name, args, err, errorMessage)
		}

		e.T.Fatalf("Error running command %s %v: %v", name, args, err)
	}
}

func (e *ClusterE2ETest) RunEKSA(args []string, opts ...CommandOpt) {
	binaryPath := e.eksaBinaryLocation
	for _, o := range opts {
		err := o(&binaryPath, &args)
		if err != nil {
			e.T.Fatalf("Error executing EKS-A at path %s with args %s: %v", binaryPath, args, err)
		}
	}
	e.Run(binaryPath, args...)
}

func (e *ClusterE2ETest) StopIfFailed() {
	if e.T.Failed() {
		e.T.FailNow()
	}
}

func (e *ClusterE2ETest) customizeClusterConfig(clusterConfigLocation string, fillers ...api.ClusterFiller) []byte {
	b, err := api.AutoFillClusterFromFile(clusterConfigLocation, fillers...)
	if err != nil {
		e.T.Fatalf("Error filling cluster config: %v", err)
	}

	return b
}

func (e *ClusterE2ETest) cleanup(f func()) {
	e.T.Cleanup(func() {
		if !e.T.Failed() {
			f()
		}
	})
}

func (e *ClusterE2ETest) cluster() *types.Cluster {
	return &types.Cluster{
		Name:           e.ClusterName,
		KubeconfigFile: e.kubeconfigFilePath(),
	}
}

func (e *ClusterE2ETest) kubeconfigFilePath() string {
	return filepath.Join(e.ClusterName, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", e.ClusterName))
}

func (e *ClusterE2ETest) managementKubeconfigFilePath() string {
	clusterConfig := e.clusterConfig()
	if clusterConfig.IsSelfManaged() {
		return e.kubeconfigFilePath()
	}
	managementClusterName := e.clusterConfig().ManagedBy()
	return filepath.Join(managementClusterName, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", managementClusterName))
}

func (e *ClusterE2ETest) GetEksaVSphereMachineConfigs() []v1alpha1.VSphereMachineConfig {
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

func (e *ClusterE2ETest) GetEksaCloudStackMachineConfigs() []v1alpha1.CloudStackMachineConfig {
	clusterConfig := e.clusterConfig()
	machineConfigNames := make([]string, 0, len(clusterConfig.Spec.WorkerNodeGroupConfigurations)+1)
	machineConfigNames = append(machineConfigNames, clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	for _, workerNodeConf := range clusterConfig.Spec.WorkerNodeGroupConfigurations {
		machineConfigNames = append(machineConfigNames, workerNodeConf.MachineGroupRef.Name)
	}

	kubeconfig := e.kubeconfigFilePath()
	ctx := context.Background()

	machineConfigs := make([]v1alpha1.CloudStackMachineConfig, 0, len(machineConfigNames))
	for _, name := range machineConfigNames {
		m, err := e.KubectlClient.GetEksaCloudStackMachineConfig(ctx, name, kubeconfig, clusterConfig.Namespace)
		if err != nil {
			e.T.Fatalf("Failed getting CloudStackMachineConfig: %v", err)
		}

		machineConfigs = append(machineConfigs, *m)
	}

	return machineConfigs
}

func (e *ClusterE2ETest) clusterConfig() *v1alpha1.Cluster {
	if e.ClusterConfig != nil {
		return e.ClusterConfig
	}

	c := &v1alpha1.Cluster{}
	if err := yaml.Unmarshal(e.ClusterConfigB, c); err != nil {
		e.T.Fatalf("Error fetching cluster config from file: %v", err)
	}
	e.ClusterConfig = c

	return e.ClusterConfig
}

func (e *ClusterE2ETest) getJobIdFromEnv() string {
	return os.Getenv(JobIdVar)
}

func getClusterName(t *testing.T) string {
	value := os.Getenv(ClusterNameVar)
	if len(value) == 0 {
		h := sha1.New()
		h.Write([]byte(t.Name()))
		testNameHash := fmt.Sprintf("%x", h.Sum(nil))
		// Append hash to make each cluster name unique per test. Using the testname will be too long
		// and would fail validations
		return fmt.Sprintf("%s-%s", defaultClusterName, testNameHash[:7])
	}
	return value
}

func getBundlesOverride() string {
	return os.Getenv(BundlesOverrideVar)
}

func setEksctlVersionEnvVar() error {
	eksctlVersionEnv := os.Getenv(eksctlVersionEnvVar)
	if eksctlVersionEnv == "" {
		err := os.Setenv(eksctlVersionEnvVar, eksctlVersionEnvVarDummyVal)
		if err != nil {
			return fmt.Errorf("couldn't set eksctl version env var %s to value %s", eksctlVersionEnvVar, eksctlVersionEnvVarDummyVal)
		}
	}
	return nil
}

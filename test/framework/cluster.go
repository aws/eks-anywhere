package framework

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	rapi "github.com/tinkerbell/rufio/api/v1alpha1"
	rctrl "github.com/tinkerbell/rufio/controllers"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/git"
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
	ClusterPrefixVar                 = "T_CLUSTER_PREFIX"
	JobIdVar                         = "T_JOB_ID"
	BundlesOverrideVar               = "T_BUNDLES_OVERRIDE"
	CleanupVmsVar                    = "T_CLEANUP_VMS"
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
	GitProvider            git.ProviderClient
	GitClient              git.Client
	HelmInstallConfig      *HelmInstallConfig
	PackageConfig          *PackageConfig
	GitWriter              filewriter.FileWriter
	OIDCConfig             *v1alpha1.OIDCConfig
	GitOpsConfig           *v1alpha1.GitOpsConfig
	FluxConfig             *v1alpha1.FluxConfig
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

	e.ClusterConfigFolder = e.ClusterName
	e.HardwareConfigLocation = filepath.Join(e.ClusterConfigFolder, hardwareYamlPath)
	e.HardwareCsvLocation = filepath.Join(e.ClusterConfigFolder, hardwareCsvPath)

	for _, opt := range opts {
		opt(e)
	}

	provider.Setup()

	e.T.Cleanup(func() {
		e.CleanupVms()
	})

	return e
}

func withHardware(requiredCount int, hardareType string, labels map[string]string) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		hardwarePool := e.GetHardwarePool()

		if e.TestHardware == nil {
			e.TestHardware = make(map[string]*api.Hardware)
		}

		var count int
		for id, h := range hardwarePool {
			if _, exists := e.TestHardware[id]; !exists {
				count++
				h.Labels = labels
				e.TestHardware[id] = h
			}

			if count == requiredCount {
				break
			}
		}

		if count < requiredCount {
			e.T.Errorf("this test requires at least %d piece(s) of %s hardware", requiredCount, hardareType)
		}
	}
}

func WithControlPlaneHardware(requiredCount int) ClusterE2ETestOpt {
	return withHardware(requiredCount, api.ControlPlane, map[string]string{api.HardwareLabelTypeKeyName: api.ControlPlane})
}

func WithWorkerHardware(requiredCount int) ClusterE2ETestOpt {
	return withHardware(requiredCount, api.Worker, map[string]string{api.HardwareLabelTypeKeyName: api.Worker})
}

func WithExternalEtcdHardware(requiredCount int) ClusterE2ETestOpt {
	return withHardware(requiredCount, api.ExternalEtcd, map[string]string{api.HardwareLabelTypeKeyName: api.ExternalEtcd})
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

func (e *ClusterE2ETest) RunClusterFlowWithGitOps(clusterOpts ...ClusterE2ETestOpt) {
	e.GenerateClusterConfig()
	e.createCluster()
	e.UpgradeWithGitOps(clusterOpts...)
	time.Sleep(5 * time.Minute)
	e.deleteCluster()
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
	CleanupVMs(clusterName string) error
}

func (e *ClusterE2ETest) GenerateClusterConfig(opts ...CommandOpt) {
	e.GenerateClusterConfigForVersion("", opts...)
}

func (e *ClusterE2ETest) PowerOffHardware() {
	// Initializing BMC Client
	ctx := context.Background()
	bmcClientFactory := rctrl.NewBMCClientFactoryFunc(ctx)

	for _, h := range e.TestHardware {
		bmcClient, err := bmcClientFactory(ctx, h.BMCIPAddress, "623", h.BMCUsername, h.BMCPassword)
		if err != nil {
			e.T.Fatalf("failed to create bmc client: %v", err)
		}

		defer func() {
			// Close BMC connection after reconcilation
			err = bmcClient.Close(ctx)
			if err != nil {
				e.T.Fatalf("BMC close connection failed: %v", err)
			}
		}()

		_, err = bmcClient.SetPowerState(ctx, string(rapi.Off))
		if err != nil {
			e.T.Fatalf("failed to power off hardware: %v", err)
		}
	}
}

func (e *ClusterE2ETest) PowerOnHardware() {
	// Initializing BMC Client
	ctx := context.Background()
	bmcClientFactory := rctrl.NewBMCClientFactoryFunc(ctx)

	for _, h := range e.TestHardware {
		bmcClient, err := bmcClientFactory(ctx, h.BMCIPAddress, "623", h.BMCUsername, h.BMCPassword)
		if err != nil {
			e.T.Fatalf("failed to create bmc client: %v", err)
		}

		defer func() {
			// Close BMC connection after reconcilation
			err = bmcClient.Close(ctx)
			if err != nil {
				e.T.Fatalf("BMC close connection failed: %v", err)
			}
		}()

		_, err = bmcClient.SetPowerState(ctx, string(rapi.On))
		if err != nil {
			e.T.Fatalf("failed to power on hardware: %v", err)
		}
	}
}

func (e *ClusterE2ETest) ValidateHardwareDecommissioned() {
	// Initializing BMC Client
	ctx := context.Background()
	bmcClientFactory := rctrl.NewBMCClientFactoryFunc(ctx)

	var failedToDecomm []*api.Hardware
	for _, h := range e.TestHardware {
		bmcClient, err := bmcClientFactory(ctx, h.BMCIPAddress, "443", h.BMCUsername, h.BMCPassword)
		if err != nil {
			e.T.Fatalf("failed to create bmc client: %v", err)
		}

		defer func() {
			// Close BMC connection after reconcilation
			err = bmcClient.Close(ctx)
			if err != nil {
				e.T.Fatalf("BMC close connection failed: %v", err)
			}
		}()

		powerState, err := bmcClient.GetPowerState(ctx)
		// add sleep retries to give the machine time to power off
		timeout := 15
		for !strings.EqualFold(powerState, string(rapi.Off)) && timeout > 0 {
			if err != nil {
				e.T.Logf("failed to get power state for hardware (%v): %v", h, err)
			}
			time.Sleep(5 * time.Second)
			timeout = timeout - 5
			powerState, err = bmcClient.GetPowerState(ctx)
			e.T.Logf("hardware power state (id=%s, hostname=%s, bmc_ip=%s): power_state=%s", h.MACAddress, h.Hostname, h.BMCIPAddress, powerState)
		}

		if !strings.EqualFold(powerState, string(rapi.Off)) {
			e.T.Logf("failed to decommission hardware: id=%s, hostname=%s, bmc_ip=%s", h.MACAddress, h.Hostname, h.BMCIPAddress)
			failedToDecomm = append(failedToDecomm, h)
		} else {
			e.T.Logf("successfully decommissioned hardware: id=%s, hostname=%s, bmc_ip=%s", h.MACAddress, h.Hostname, h.BMCIPAddress)
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
		"-z", e.HardwareCsvLocation,
		"-o", e.HardwareConfigLocation,
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
	clusterConfigFillers := make([]api.ClusterFiller, 0, len(e.clusterFillers)+len(clusterFillersFromProvider)+3)
	// This defaults all tests to a 1:1:1 configuration. Since all the fillers defined on each test are run
	// after these 3, if the tests is explicit about any of these, the defaults will be overwritten
	// (@g-gaston) This is a temporary fix to avoid overloading the CI system and we should remove it once we
	// stabilize the test runs
	clusterConfigFillers = append(clusterConfigFillers,
		api.WithControlPlaneCount(1), api.WithWorkerNodeCount(1), api.WithEtcdCountIfExternal(1),
	)
	clusterConfigFillers = append(clusterConfigFillers, e.clusterFillers...)
	clusterConfigFillers = append(clusterConfigFillers, clusterFillersFromProvider...)
	e.ClusterConfigB = e.customizeClusterConfig(e.ClusterConfigLocation, clusterConfigFillers...)
	e.ProviderConfigB = e.Provider.CustomizeProviderConfig(e.ClusterConfigLocation)
}

func (e *ClusterE2ETest) ImportImages(opts ...CommandOpt) {
	importImagesArgs := []string{"import-images", "-f", e.ClusterConfigLocation}
	e.RunEKSA(importImagesArgs, opts...)
}

func (e *ClusterE2ETest) DownloadArtifacts(opts ...CommandOpt) {
	downloadArtifactsArgs := []string{"download", "artifacts", "-f", e.ClusterConfigLocation}
	e.RunEKSA(downloadArtifactsArgs, opts...)
	if _, err := os.Stat("eks-anywhere-downloads.tar.gz"); err != nil {
		e.T.Fatal(err)
	} else {
		e.T.Log("Downloaded artifacts saved at eks-anywhere-downloads.tar.gz")
	}
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
		createClusterArgs = append(createClusterArgs, "-z", e.HardwareCsvLocation)
		tinkBootstrapIP := os.Getenv(tinkerbellBootstrapIPEnvVar)
		e.T.Logf("tinkBootstrapIP: %s", tinkBootstrapIP)
		if tinkBootstrapIP != "" {
			createClusterArgs = append(createClusterArgs, "--tinkerbell-bootstrap-ip", tinkBootstrapIP)
		}
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
			e.T.Fatalf("marshalling oidc config: %v", err)
		}
		yamlB = append(yamlB, oidcConfigB)
	}
	if e.AWSIamConfig != nil {
		awsIamConfigB, err := yaml.Marshal(e.AWSIamConfig)
		if err != nil {
			e.T.Fatalf("marshalling aws iam config: %v", err)
		}
		yamlB = append(yamlB, awsIamConfigB)
	}
	if e.GitOpsConfig != nil {
		gitOpsConfigB, err := yaml.Marshal(e.GitOpsConfig)
		if err != nil {
			e.T.Fatalf("marshalling gitops config: %v", err)
		}
		yamlB = append(yamlB, gitOpsConfigB)
	}
	if e.FluxConfig != nil {
		fluxConfigB, err := yaml.Marshal(e.FluxConfig)
		if err != nil {
			e.T.Fatalf("marshalling gitops config: %v", err)
		}
		yamlB = append(yamlB, fluxConfigB)
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

func (e *ClusterE2ETest) CleanupVms() {
	if !shouldCleanUpVms() {
		e.T.Logf("Skipping VM cleanup")
		return
	}

	if err := e.Provider.CleanupVMs(e.ClusterName); err != nil {
		e.T.Logf("failed to clean up VMs: %v", err)
	}
}

func shouldCleanUpVms() bool {
	shouldCleanupVms, err := getCleanupVmsVar()
	return err == nil && shouldCleanupVms
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

func GetTestNameHash(name string) string {
	h := sha1.New()
	h.Write([]byte(name))
	testNameHash := fmt.Sprintf("%x", h.Sum(nil))
	return testNameHash[:7]
}

func getClusterName(t *testing.T) string {
	value := os.Getenv(ClusterPrefixVar)
	// Append hash to make each cluster name unique per test. Using the testname will be too long
	// and would fail validations
	if len(value) == 0 {
		value = defaultClusterName
	}

	return fmt.Sprintf("%s-%s", value, GetTestNameHash(t.Name()))
}

func getBundlesOverride() string {
	return os.Getenv(BundlesOverrideVar)
}

func getCleanupVmsVar() (bool, error) {
	return strconv.ParseBool(os.Getenv(CleanupVmsVar))
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

func (e *ClusterE2ETest) InstallHelmChart() {
	kubeconfig := e.kubeconfigFilePath()
	ctx := context.Background()

	err := e.HelmInstallConfig.HelmClient.InstallChart(ctx, e.HelmInstallConfig.chartName, e.HelmInstallConfig.chartURI, e.HelmInstallConfig.chartVersion, kubeconfig, e.HelmInstallConfig.chartValues)
	if err != nil {
		e.T.Fatalf("Error installing %s helm chart on the cluster: %v", e.HelmInstallConfig.chartName, err)
	}
}

func (e *ClusterE2ETest) InstallCuratedPackagesController() {
	kubeconfig := e.kubeconfigFilePath()
	// TODO Add a test that installs the controller via the CLI.
	ctx := context.Background()
	err := e.PackageConfig.HelmClient.InstallChart(ctx,
		e.PackageConfig.chartName, e.PackageConfig.chartURI,
		e.PackageConfig.chartVersion, kubeconfig, e.PackageConfig.chartValues)
	if err != nil {
		e.T.Fatalf("Error installing %s helm chart on the cluster: %v",
			e.PackageConfig.chartName, err)
	}
}

func (e *ClusterE2ETest) InstallCuratedPackage(packageName, packagePrefix string) {
	os.Setenv("CURATED_PACKAGES_SUPPORT", "true")
	// The package install command doesn't (yet?) have a --kubeconfig flag.
	os.Setenv("KUBECONFIG", e.kubeconfigFilePath())
	e.RunEKSA([]string{
		"install", "package", packageName,
		"--source=registry", "--registry=public.ecr.aws/l0g8r8j6",
		"--package-name=" + packagePrefix, "-v=9", "--kube-version=1.21",
	})
}

// WithCluster helps with bringing up and tearing down E2E test clusters.
func (e *ClusterE2ETest) WithCluster(f func(e *ClusterE2ETest)) {
	e.GenerateClusterConfig()
	e.CreateCluster()
	defer e.DeleteCluster()
	f(e)
}

func (e *ClusterE2ETest) VerifyHelloPackageInstalled(name string) {
	ctx := context.Background()

	ns := constants.EksaPackagesName
	err := e.KubectlClient.WaitForService(ctx,
		e.cluster().KubeconfigFile, "5m", name, ns)
	if err != nil {
		e.T.Fatalf("waiting for service timed out: %s", err)
	}

	// Ensure that the pod is up before trying to port-forward. In some test
	// environments, the pod might not be running when the port-forward is
	// attempted, and that will cause the port-forward to fail.
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.cluster(), "5m", "Available", "hello-eks-anywhere", ns)
	if err != nil {
		e.T.Fatalf("waiting for hello-eks-anywhere pod timed out: %s", err)
	}

	timedCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	timedOut := timedCtx.Done()
	// It's preferable to configure kubectl to use a random port, which
	// it would write to stdout, indicating when the port-forward is
	// active. However, the current Executable framework doesn't allow
	// for reading stdout before the process exits. Polling provides a
	// workable solution.
	const port = 9980 // ...and hope it's available...
	stopPF, pfErrCh := e.forwardPortToService(timedCtx, name, ns, port)
	defer stopPF()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var resp *http.Response
outer:
	for {
		select {
		case <-timedOut:
			e.T.Fatalf("timed out: %s", timedCtx.Err())
		case err := <-pfErrCh:
			e.T.Fatalf("port forwarding error: %s", err)
		case <-ticker.C:
			url := fmt.Sprintf("http://localhost:%d/index.json", port)
			resp, err = http.Get(url)
			if err != nil {
				e.T.Logf("service error, will retry: %s", err)
				continue
			}
			if resp.StatusCode < http.StatusOK ||
				resp.StatusCode >= http.StatusMultipleChoices {
				resp.Body.Close()
				e.T.Fatalf("expected a 2XX response, got: %d (%s)",
					resp.StatusCode, http.StatusText(resp.StatusCode))
			}
			defer resp.Body.Close()
			break outer
		}
	}

	buf := &bytes.Buffer{}
	// A TeeReader will let us log the entire body in case of an error.
	tee := io.TeeReader(resp.Body, buf)
	respData := map[string]interface{}{}
	if err = json.NewDecoder(tee).Decode(&respData); err != nil {
		_, debugErr := io.ReadAll(tee)
		if debugErr != nil {
			// Just log this, since the test is already a failure.
			e.T.Logf("trying to read the entire response body: %s", debugErr)
		}
		e.T.Fatalf("unmarshaling JSON response: %s\n%s", err, buf.String())
	}

	title, ok := respData["title"].(string)
	if !ok {
		e.T.Fatalf("expected title to be a string, got %T", respData["title"])
	}
	expected := "Amazon EKS Anywhere"
	if !strings.EqualFold(title, expected) {
		e.T.Fatalf("expected title to be %q, got %q", expected, title)
	}
}

func (e *ClusterE2ETest) forwardPortToService(ctx context.Context,
	name, namespace string, port int,
) (func(), <-chan error) {
	// The current Executable framework doesn't allow reading stdout before
	// the command completes, so there's no way to know when the port-forward
	// is available, short of just trying it.
	pfContext, pfCancel := context.WithCancel(ctx)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer close(errCh)
		defer wg.Done()
		_, err := e.KubectlClient.Execute(pfContext, "port-forward",
			"--kubeconfig="+e.kubeconfigFilePath(), "--namespace="+namespace,
			"service/"+name, fmt.Sprintf("%d:80", port))
		if err != nil {
			pfCtxErr := pfContext.Err()
			// A canceled context indicates a controlled shutdown.
			if errors.Is(pfCtxErr, context.Canceled) {
				return
			}
			if pfCtxErr != nil {
				e.T.Logf("port-forward context error: %s", err)
			}
			errCh <- err
		}
	}()

	return func() {
		pfCancel()
		wg.Wait()
	}, errCh
}

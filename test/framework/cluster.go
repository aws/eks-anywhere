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
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	rapi "github.com/tinkerbell/rufio/api/v1alpha1"
	rctrl "github.com/tinkerbell/rufio/controllers"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
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
	ClusterIPPoolEnvVar              = "T_CLUSTER_IP_POOL"
	CleanupVmsVar                    = "T_CLEANUP_VMS"
	hardwareYamlPath                 = "hardware.yaml"
	hardwareCsvPath                  = "hardware.csv"
	EksaPackagesInstallation         = "eks-anywhere-packages"
)

//go:embed testdata/oidc-roles.yaml
var oidcRoles []byte

//go:embed testdata/hpa_busybox.yaml
var hpaBusybox []byte

type ClusterE2ETest struct {
	T                      *testing.T
	ClusterConfigLocation  string
	ClusterConfigFolder    string
	HardwareConfigLocation string
	HardwareCsvLocation    string
	TestHardware           map[string]*api.Hardware
	HardwarePool           map[string]*api.Hardware
	WithNoPowerActions     bool
	ClusterName            string
	ClusterConfig          *cluster.Config
	Provider               Provider
	clusterFillers         []api.ClusterFiller
	KubectlClient          *executables.Kubectl
	GitProvider            git.ProviderClient
	GitClient              git.Client
	HelmInstallConfig      *HelmInstallConfig
	PackageConfig          *PackageConfig
	GitWriter              filewriter.FileWriter
	eksaBinaryLocation     string
	ExpectFailure          bool
}

type ClusterE2ETestOpt func(e *ClusterE2ETest)

func NewClusterE2ETest(t *testing.T, provider Provider, opts ...ClusterE2ETestOpt) *ClusterE2ETest {
	e := &ClusterE2ETest{
		T:                     t,
		Provider:              provider,
		ClusterConfig:         &cluster.Config{},
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

		tinkerbellCIEnvironment := os.Getenv(TinkerbellCIEnvironment)
		if e.Provider.Name() == TinkerbellProviderName && tinkerbellCIEnvironment == "true" {
			e.CleanupDockerEnvironment()
		}
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

func WithNoPowerActions() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.WithNoPowerActions = true
	}
}

func ExpectFailure(expected bool) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ExpectFailure = expected
	}
}

func WithControlPlaneHardware(requiredCount int) ClusterE2ETestOpt {
	return withHardware(requiredCount, api.ControlPlane, map[string]string{api.HardwareLabelTypeKeyName: api.ControlPlane})
}

func WithWorkerHardware(requiredCount int) ClusterE2ETestOpt {
	return withHardware(requiredCount, api.Worker, map[string]string{api.HardwareLabelTypeKeyName: api.Worker})
}

func WithCustomLabelHardware(requiredCount int, label string) ClusterE2ETestOpt {
	return withHardware(requiredCount, api.Worker, map[string]string{api.HardwareLabelTypeKeyName: label})
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

// WithClusterSingleNode helps to create an e2e test option for a single node cluster.
func WithClusterSingleNode(v v1alpha1.KubernetesVersion) ClusterE2ETestOpt {
	return WithClusterFiller(
		api.WithKubernetesVersion(v),
		api.WithControlPlaneCount(1),
		api.WithEtcdCountIfExternal(0),
		api.RemoveAllWorkerNodeGroups(),
	)
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
	// ClusterConfigUpdates allows a provider to modify the default cluster config
	// after this one is generated for the first time. This is not reapplied on every CLI operation.
	// Prefer to call UpdateClusterConfig directly from the tests to make it more explicit.
	ClusterConfigUpdates() []api.ClusterConfigFiller
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

func (e *ClusterE2ETest) PXEBootHardware() {
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

		_, err = bmcClient.SetBootDevice(ctx, string(rapi.PXE), false, true)
		if err != nil {
			e.T.Fatalf("failed to pxe boot hardware: %v", err)
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

	testHardware := e.TestHardware
	if e.WithNoPowerActions {
		hardwareWithNoBMC := make(map[string]*api.Hardware)
		for k, h := range testHardware {
			lessBmc := *h
			lessBmc.BMCIPAddress = ""
			lessBmc.BMCUsername = ""
			lessBmc.BMCPassword = ""
			hardwareWithNoBMC[k] = &lessBmc
		}
		testHardware = hardwareWithNoBMC
	}

	err := api.WriteHardwareMapToCSV(testHardware, e.HardwareCsvLocation)
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
		err := cleanUpClusterForVersion(e.ClusterConfig, eksaVersion)
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

	config, err := cluster.ParseConfigFromFile(e.ClusterConfigLocation)
	if err != nil {
		e.T.Fatalf("Failed parsing generated cluster config: %s", err)
	}

	// Copy all objects that might be generated by the CLI.
	// Don't replace the whole ClusterConfig since somne ClusterE2ETestOpt might
	// have already set some data in it.
	e.ClusterConfig.Cluster = config.Cluster
	e.ClusterConfig.CloudStackDatacenter = config.CloudStackDatacenter
	e.ClusterConfig.VSphereDatacenter = config.VSphereDatacenter
	e.ClusterConfig.DockerDatacenter = config.DockerDatacenter
	e.ClusterConfig.SnowDatacenter = config.SnowDatacenter
	e.ClusterConfig.NutanixDatacenter = config.NutanixDatacenter
	e.ClusterConfig.TinkerbellDatacenter = config.TinkerbellDatacenter
	e.ClusterConfig.VSphereMachineConfigs = config.VSphereMachineConfigs
	e.ClusterConfig.CloudStackMachineConfigs = config.CloudStackMachineConfigs
	e.ClusterConfig.SnowMachineConfigs = config.SnowMachineConfigs
	e.ClusterConfig.NutanixMachineConfigs = config.NutanixMachineConfigs
	e.ClusterConfig.TinkerbellMachineConfigs = config.TinkerbellMachineConfigs
	e.ClusterConfig.TinkerbellTemplateConfigs = config.TinkerbellTemplateConfigs

	clusterFillers := make([]api.ClusterFiller, 0, len(e.clusterFillers)+3)
	// This defaults all tests to a 1:1:1 configuration. Since all the fillers defined on each test are run
	// after these 3, if the tests is explicit about any of these, the defaults will be overwritten
	// (@g-gaston) This is a temporary fix to avoid overloading the CI system and we should remove it once we
	// stabilize the test runs
	clusterFillers = append(clusterFillers,
		api.WithControlPlaneCount(1), api.WithWorkerNodeCount(1), api.WithEtcdCountIfExternal(1),
	)
	clusterFillers = append(clusterFillers, e.clusterFillers...)
	configFillers := []api.ClusterConfigFiller{api.ClusterToConfigFiller(clusterFillers...)}
	configFillers = append(configFillers, e.Provider.ClusterConfigUpdates()...)

	e.UpdateClusterConfig(configFillers...)
}

// UpdateClusterConfig applies the cluster Config provided updates to e.ClusterConfig, marshalls its content
// to yaml and writes it to a file on disk configured by e.ClusterConfigLocation. Call this method when you want
// make changes to the eks-a cluster definition before running a CLI command or API operation.
func (e *ClusterE2ETest) UpdateClusterConfig(fillers ...api.ClusterConfigFiller) {
	api.UpdateClusterConfig(e.ClusterConfig, fillers...)
	e.buildClusterConfigFile()
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
	createClusterArgs := []string{"create", "cluster", "-f", e.ClusterConfigLocation, "-v", "12"}
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

func (e *ClusterE2ETest) WaitForMachineDeploymentReady(machineDeploymentName string) {
	ctx := context.Background()
	e.T.Logf("Waiting for machine deployment %s to be ready for cluster %s", machineDeploymentName, e.ClusterName)
	err := e.KubectlClient.WaitForMachineDeploymentReady(ctx, e.cluster(), "5m", machineDeploymentName)
	if err != nil {
		e.T.Fatal(err)
	}
}

func (e *ClusterE2ETest) GetCapiMachinesForCluster(clusterName string) map[string]types.Machine {
	ctx := context.Background()
	capiMachines, err := e.KubectlClient.GetMachines(ctx, e.cluster(), clusterName)
	if err != nil {
		e.T.Fatal(err)
	}
	machinesMap := make(map[string]types.Machine, 0)
	for _, machine := range capiMachines {
		machinesMap[machine.Metadata.Name] = machine
	}
	return machinesMap
}

func WithClusterUpgrade(fillers ...api.ClusterFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.ClusterToConfigFiller(fillers...))
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

func (e *ClusterE2ETest) generateClusterConfigYaml() []byte {
	childObjs := e.ClusterConfig.ChildObjects()
	yamlB := make([][]byte, 0, len(childObjs)+1)
	clusterConfigB, err := yaml.Marshal(e.ClusterConfig.Cluster)
	if err != nil {
		e.T.Fatal(err)
	}
	yamlB = append(yamlB, clusterConfigB)
	for _, o := range childObjs {
		objB, err := yaml.Marshal(o)
		if err != nil {
			e.T.Fatalf("Failed marshalling %s config: %v", o.GetName(), err)
		}
		yamlB = append(yamlB, objB)
	}

	return templater.AppendYamlResources(yamlB...)
}

func (e *ClusterE2ETest) buildClusterConfigFile() {
	yaml := e.generateClusterConfigYaml()

	writer, err := filewriter.NewWriter(e.ClusterConfigFolder)
	if err != nil {
		e.T.Fatalf("Error creating writer: %v", err)
	}

	writtenFile, err := writer.Write(filepath.Base(e.ClusterConfigLocation), yaml, filewriter.PersistentFile)
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

func (e *ClusterE2ETest) CleanupDockerEnvironment() {
	e.T.Logf("cleanup kind enviornment...")
	e.Run("kind", "delete", "clusters", "--all", "||", "true")
	e.T.Logf("cleanup docker enviornment...")
	e.Run("docker", "rm", "-vf", "$(docker ps -a -q)", "||", "true")
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
			if e.ExpectFailure {
				e.T.Logf("This error was expected. Continuing...")
				return
			}
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
	clusterConfig := e.ClusterConfig.Cluster
	if clusterConfig.IsSelfManaged() {
		return e.kubeconfigFilePath()
	}
	managementClusterName := e.ClusterConfig.Cluster.ManagedBy()
	return filepath.Join(managementClusterName, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", managementClusterName))
}

func (e *ClusterE2ETest) GetEksaVSphereMachineConfigs() []v1alpha1.VSphereMachineConfig {
	clusterConfig := e.ClusterConfig.Cluster
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

	err := e.HelmInstallConfig.HelmClient.InstallChart(ctx, e.HelmInstallConfig.chartName, e.HelmInstallConfig.chartURI, e.HelmInstallConfig.chartVersion, kubeconfig, "", e.HelmInstallConfig.chartValues)
	if err != nil {
		e.T.Fatalf("Error installing %s helm chart on the cluster: %v", e.HelmInstallConfig.chartName, err)
	}
}

func (e *ClusterE2ETest) CreateNamespace(namespace string) {
	kubeconfig := e.kubeconfigFilePath()
	err := e.KubectlClient.CreateNamespace(context.Background(), kubeconfig, namespace)
	if err != nil {
		e.T.Fatalf("Namespace creation failed for %s", namespace)
	}
}

func (e *ClusterE2ETest) DeleteNamespace(namespace string) {
	kubeconfig := e.kubeconfigFilePath()
	err := e.KubectlClient.DeleteNamespace(context.Background(), kubeconfig, namespace)
	if err != nil {
		e.T.Fatalf("Namespace deletion failed for %s", namespace)
	}
}

func (e *ClusterE2ETest) InstallCuratedPackagesController() {
	kubeconfig := e.kubeconfigFilePath()
	// TODO Add a test that installs the controller via the CLI.
	ctx := context.Background()
	charts, err := e.PackageConfig.HelmClient.ListCharts(ctx, kubeconfig)
	if err != nil {
		e.T.Fatalf("Unable to list charts: %v", err)
	}
	installed := false
	for _, c := range charts {
		if c == EksaPackagesInstallation {
			installed = true
			break
		}
	}
	if !installed {
		err = e.PackageConfig.HelmClient.InstallChart(ctx, e.PackageConfig.chartName, e.PackageConfig.chartURI, e.PackageConfig.chartVersion, kubeconfig, "eksa-packages", e.PackageConfig.chartValues)
		if err != nil {
			e.T.Fatalf("Unable to install %s helm chart on the cluster: %v",
				e.PackageConfig.chartName, err)
		}
	}
}

// SetPackageBundleActive will set the current packagebundle to the active state.
func (e *ClusterE2ETest) SetPackageBundleActive() {
	kubeconfig := e.kubeconfigFilePath()
	pbc, err := e.KubectlClient.GetPackageBundleController(context.Background(), kubeconfig, e.ClusterName)
	if err != nil {
		e.T.Fatalf("Error getting PackageBundleController: %v", err)
	}
	pb, err := e.KubectlClient.GetPackageBundleList(context.Background(), e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("Error getting PackageBundle: %v", err)
	}
	if pbc.Spec.ActiveBundle != pb[0].ObjectMeta.Name {
		e.RunEKSA([]string{
			"upgrade", "packages",
			"--bundle-version", pb[0].ObjectMeta.Name, "-v=9",
			"--cluster=" + e.ClusterName,
		})
	}
}

// InstallCuratedPackage will install a curated package in the desired namespace.
func (e *ClusterE2ETest) InstallCuratedPackage(packageName, packagePrefix, kubeconfig, namespace string, opts ...string) {
	os.Setenv("CURATED_PACKAGES_SUPPORT", "true")
	// The package install command doesn't (yet?) have a --kubeconfig flag.
	os.Setenv("KUBECONFIG", kubeconfig)
	e.RunEKSA([]string{
		"install", "package", packageName,
		"--package-name=" + packagePrefix, "-v=9",
		"--cluster=" + e.ClusterName,
		strings.Join(opts, " "),
	})
}

// InstallCuratedPackageFile will install a curated package from a yaml file, this is useful since target namespace isn't supported on the CLI.
func (e *ClusterE2ETest) InstallCuratedPackageFile(packageFile, kubeconfig string, opts ...string) {
	os.Setenv("CURATED_PACKAGES_SUPPORT", "true")
	os.Setenv("KUBECONFIG", kubeconfig)
	e.T.Log("Installing EKS-A Packages file", packageFile)
	e.RunEKSA([]string{
		"apply", "package", "-f", packageFile, "-v=9", strings.Join(opts, " "),
	})
}

func (e *ClusterE2ETest) generatePackageConfig(ns, targetns, prefix, packageName string) []byte {
	yamlB := make([][]byte, 0, 4)
	generatedName := fmt.Sprintf("%s-%s", prefix, packageName)
	if targetns == "" {
		targetns = ns
	}
	ns = fmt.Sprintf("%s-%s", ns, e.ClusterName)
	builtpackage := &packagesv1.Package{
		TypeMeta: metav1.TypeMeta{
			Kind:       packagesv1.PackageKind,
			APIVersion: "packages.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      generatedName,
			Namespace: ns,
		},
		Spec: packagesv1.PackageSpec{
			PackageName:     packageName,
			TargetNamespace: targetns,
		},
	}
	builtpackageB, err := yaml.Marshal(builtpackage)
	if err != nil {
		e.T.Fatalf("marshalling package config file: %v", err)
	}
	yamlB = append(yamlB, builtpackageB)
	return templater.AppendYamlResources(yamlB...)
}

// BuildPackageConfigFile will create the file in the test directory for the curated package.
func (e *ClusterE2ETest) BuildPackageConfigFile(packageName, prefix, ns string) string {
	b := e.generatePackageConfig(ns, ns, prefix, packageName)

	writer, err := filewriter.NewWriter(e.ClusterConfigFolder)
	if err != nil {
		e.T.Fatalf("Error creating writer: %v", err)
	}
	packageFile := fmt.Sprintf("%s.yaml", packageName)

	writtenFile, err := writer.Write(packageFile, b, filewriter.PersistentFile)
	if err != nil {
		e.T.Fatalf("Error writing cluster config to file %s: %v", e.ClusterConfigLocation, err)
	}
	return writtenFile
}

func (e *ClusterE2ETest) CreateResource(ctx context.Context, resource string) {
	err := e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.cluster(), []byte(resource))
	if err != nil {
		e.T.Fatalf("Failed to create required resource (%s): %v", resource, err)
	}
}

func (e *ClusterE2ETest) UninstallCuratedPackage(packagePrefix string, opts ...string) {
	os.Setenv("CURATED_PACKAGES_SUPPORT", "true")
	os.Setenv("KUBECONFIG", e.kubeconfigFilePath())
	e.RunEKSA([]string{
		"delete", "package", packagePrefix, "-v=9",
		strings.Join(opts, " "),
	})
}

func (e *ClusterE2ETest) InstallLocalStorageProvisioner() {
	ctx := context.Background()
	_, err := e.KubectlClient.ExecuteCommand(ctx, "apply", "-f",
		"https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.22/deploy/local-path-storage.yaml",
		"--kubeconfig", e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("Error installing local-path-provisioner: %v", err)
	}
}

// WithCluster helps with bringing up and tearing down E2E test clusters.
func (e *ClusterE2ETest) WithCluster(f func(e *ClusterE2ETest)) {
	e.GenerateClusterConfig()
	e.CreateCluster()
	defer e.DeleteCluster()
	f(e)
}

// Like WithCluster but does not delete the cluster. Useful for debugging.
func (e *ClusterE2ETest) WithPersistentCluster(f func(e *ClusterE2ETest)) {
	configPath := e.kubeconfigFilePath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		e.GenerateClusterConfig()
		e.CreateCluster()
	}
	f(e)
}

// VerifyHarborPackageInstalled is checking if the harbor package gets installed correctly.
func (e *ClusterE2ETest) VerifyHarborPackageInstalled(prefix string, namespace string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deployments := []string{"core", "jobservice", "nginx", "portal", "registry"}
	statefulsets := []string{"database", "redis", "trivy"}

	var wg sync.WaitGroup
	wg.Add(len(deployments) + len(statefulsets))
	errCh := make(chan error, 1)
	okCh := make(chan string, 1)

	time.Sleep(3 * time.Minute)

	for _, name := range deployments {
		go func(name string) {
			defer wg.Done()
			err := e.KubectlClient.WaitForDeployment(ctx,
				e.cluster(), "5m", "Available", fmt.Sprintf("%s-harbor-%s", prefix, name), namespace)
			if err != nil {
				errCh <- err
			}
		}(name)
	}
	for _, name := range statefulsets {
		go func(name string) {
			defer wg.Done()
			err := e.KubectlClient.Wait(ctx, e.kubeconfigFilePath(), "5m", "Ready",
				fmt.Sprintf("pods/%s-harbor-%s-0", prefix, name), namespace)
			if err != nil {
				errCh <- err
			}
		}(name)
	}
	go func() {
		wg.Wait()
		okCh <- "completed"
	}()

	select {
	case err := <-errCh:
		e.T.Fatal(err)
	case <-okCh:
		return
	}
}

// VerifyHelloPackageInstalled is checking if the hello eks anywhere package gets installed correctly.
func (e *ClusterE2ETest) VerifyHelloPackageInstalled(name string, mgmtCluster *types.Cluster) {
	ctx := context.Background()
	ns := constants.EksaPackagesName

	e.T.Log("Waiting for Package", name, "To be installed")
	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		mgmtCluster, name, "5m", fmt.Sprintf("%s-%s", ns, e.ClusterName))
	if err != nil {
		e.T.Fatalf("waiting for hello-eks-anywhere package timed out: %s", err)
	}

	e.T.Log("Waiting for Package", name, "Deployment to be healthy")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.cluster(), "5m", "Available", "hello-eks-anywhere", ns)
	if err != nil {
		e.T.Fatalf("waiting for hello-eks-anywhere deployment timed out: %s", err)
	}

	svcAddress := name + "." + ns + ".svc.cluster.local"
	randomname := fmt.Sprintf("%s-%s", "busybox-test", utilrand.String(7))
	clientPod, err := e.KubectlClient.RunBusyBoxPod(context.TODO(), ns, randomname, e.kubeconfigFilePath(), []string{"curl", svcAddress})
	if err != nil {
		e.T.Fatalf("error launching busybox pod: %s", err)
	}
	e.T.Log("Launching Busybox pod", clientPod, "to test Package", name)

	err = e.KubectlClient.WaitForPodCompleted(ctx,
		e.cluster(), clientPod, "5m", ns)
	if err != nil {
		e.T.Fatalf("waiting for busybox pod timed out: %s", err)
	}

	e.T.Log("Checking Busybox pod logs", clientPod)
	logs, err := e.KubectlClient.GetPodLogs(context.TODO(), ns, clientPod, clientPod, e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("failure getting pod logs %s", err)
	}
	fmt.Printf("Logs from curl Hello EKS Anywhere\n %s\n", logs)
	ok := strings.Contains(logs, "Amazon EKS Anywhere")
	if !ok {
		e.T.Fatalf("expected Amazon EKS Anywhere, got %T", logs)
	}
}

// VerifyAdotPackageInstalled is checking if the ADOT package gets installed correctly.
func (e *ClusterE2ETest) VerifyAdotPackageInstalled(packageName string, targetNamespace string) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", "eksa-packages", e.ClusterName)

	e.T.Log("Waiting for package", packageName, "to be installed")
	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		e.cluster(), packageName, "10m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for adot package install timed out: %s", err)
	}

	e.T.Log("Waiting for package", packageName, "deployment to be available")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.cluster(), "5m", "Available", fmt.Sprintf("%s-aws-otel-collector", packageName), targetNamespace)
	if err != nil {
		e.T.Fatalf("waiting for adot deployment timed out: %s", err)
	}

	e.T.Log("Reading", packageName, "pod logs")
	adotPodName, err := e.KubectlClient.GetPodNameByLabel(context.TODO(), targetNamespace, "app.kubernetes.io/name=aws-otel-collector", e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("unable to get name of the aws-otel-collector pod: %s", err)
	}
	logs, err := e.KubectlClient.GetPodLogs(context.TODO(), targetNamespace, adotPodName, "aws-otel-collector", e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("failure getting pod logs %s", err)
	}
	fmt.Printf("Logs from aws-otel-collector pod\n %s\n", logs)
	ok := strings.Contains(logs, "Everything is ready")
	if !ok {
		e.T.Fatalf("expected to find 'Everything is ready' in the log, got %s", logs)
	}

	e.T.Log("Launching Busybox pod to test Package", packageName)
	podIPAddress, err := e.KubectlClient.GetPodIP(context.TODO(), targetNamespace, adotPodName, e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("unable to get ip of the aws-otel-collector pod: %s", err)
	}
	podFullIPAddress := strings.Trim(podIPAddress, `'"`) + ":8888/metrics"
	busyBoxName := fmt.Sprintf("%s-%s", "busybox-test", utilrand.String(7))
	clientPod, err := e.KubectlClient.RunBusyBoxPod(context.TODO(), targetNamespace, busyBoxName, e.kubeconfigFilePath(), []string{"curl", podFullIPAddress})
	if err != nil {
		e.T.Fatalf("error launching busybox pod: %s", err)
	}
	e.T.Log("Waiting Busybox pod", clientPod, "to be ready")
	err = e.KubectlClient.WaitForPodCompleted(ctx,
		e.cluster(), clientPod, "5m", targetNamespace)
	if err != nil {
		e.T.Fatalf("waiting for busybox pod timed out: %s", err)
	}
	e.T.Log("Checking Busybox pod logs", clientPod)
	logs, err = e.KubectlClient.GetPodLogs(context.TODO(), targetNamespace, clientPod, clientPod, e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("failure getting pod logs %s", err)
	}
	fmt.Printf("Logs from curl adot\n %s\n", logs)
	ok = strings.Contains(logs, "otelcol_exporter")
	if !ok {
		e.T.Fatalf("expected to find otelcol_exporter in the log, got %s", logs)
	}
}

//go:embed testdata/adot_package_deployment.yaml
var adotPackageDeployment []byte

//go:embed testdata/adot_package_daemonset.yaml
var adotPackageDaemonset []byte

// VerifyAdotPackageDeploymentUpdated is checking if deployment config changes trigger resource reloads correctly.
func (e *ClusterE2ETest) VerifyAdotPackageDeploymentUpdated(packageName string, targetNamespace string) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", "eksa-packages", e.ClusterName)

	// Deploy ADOT as a deployment and scrape the apiservers
	e.T.Log("Apply changes to package", packageName)
	e.T.Log("This will update", packageName, "to be a deployment, and scrape the apiservers")
	err := e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, e.cluster(), adotPackageDeployment, packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("Error upgrading adot package: %s", err)
		return
	}
	time.Sleep(30 * time.Second) // Add sleep to allow package to change state

	e.T.Log("Waiting for package", packageName, "to be updated")
	err = e.KubectlClient.WaitForPackagesInstalled(ctx,
		e.cluster(), packageName, "10m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for adot package update timed out: %s", err)
	}

	e.T.Log("Waiting for package", packageName, "deployment to be available")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.cluster(), "5m", "Available", fmt.Sprintf("%s-aws-otel-collector", packageName), targetNamespace)
	if err != nil {
		e.T.Fatalf("waiting for adot deployment timed out: %s", err)
	}

	e.T.Log("Reading", packageName, "pod logs")
	adotPodName, err := e.KubectlClient.GetPodNameByLabel(context.TODO(), targetNamespace, "app.kubernetes.io/name=aws-otel-collector", e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("unable to get name of the aws-otel-collector pod: %s", err)
	}
	logs, err := e.KubectlClient.GetPodLogs(context.TODO(), targetNamespace, adotPodName, "aws-otel-collector", e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("failure getting pod logs %s", err)
	}
	fmt.Printf("Logs from aws-otel-collector pod\n %s\n", logs)
	expectedLogs := "MetricsExporter	{\"kind\": \"exporter\", \"data_type\": \"metrics\", \"name\": \"logging\", \"#metrics\":"
	ok := strings.Contains(logs, expectedLogs)
	if !ok {
		e.T.Fatalf("expected to find %s in the log, got %s", expectedLogs, logs)
	}
}

// VerifyAdotPackageDaemonSetUpdated is checking if daemonset config changes trigger resource reloads correctly.
func (e *ClusterE2ETest) VerifyAdotPackageDaemonSetUpdated(packageName string, targetNamespace string) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", "eksa-packages", e.ClusterName)

	// Deploy ADOT as a daemonset and scrape the node
	e.T.Log("Apply changes to package", packageName)
	e.T.Log("This will update", packageName, "to be a daemonset, and scrape the node")
	err := e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, e.cluster(), adotPackageDaemonset, packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("Error upgrading adot package: %s", err)
		return
	}
	time.Sleep(30 * time.Second) // Add sleep to allow package to change state

	e.T.Log("Waiting for package", packageName, "to be updated")
	err = e.KubectlClient.WaitForPackagesInstalled(ctx,
		e.cluster(), packageName, "10m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for adot package update timed out: %s", err)
	}

	e.T.Log("Waiting for package", packageName, "daemonset to be rolled out")
	err = retrier.New(6 * time.Minute).Retry(func() error {
		return e.KubectlClient.WaitForDaemonsetRolledout(ctx,
			e.cluster(), "5m", fmt.Sprintf("%s-aws-otel-collector-agent", packageName), targetNamespace)
	})
	if err != nil {
		e.T.Fatalf("waiting for adot daemonset timed out: %s", err)
	}

	e.T.Log("Reading", packageName, "pod logs")
	adotPodName, err := e.KubectlClient.GetPodNameByLabel(context.TODO(), targetNamespace, "app.kubernetes.io/name=aws-otel-collector", e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("unable to get name of the aws-otel-collector pod: %s", err)
	}
	expectedLogs := "MetricsExporter	{\"kind\": \"exporter\", \"data_type\": \"metrics\", \"name\": \"logging\", \"#metrics\":"
	err = retrier.New(5 * time.Minute).Retry(func() error {
		logs, err := e.KubectlClient.GetPodLogs(context.TODO(), targetNamespace, adotPodName, "aws-otel-collector", e.kubeconfigFilePath())
		if err != nil {
			e.T.Fatalf("failure getting pod logs %s", err)
		}
		fmt.Printf("Logs from aws-otel-collector pod\n %s\n", logs)
		ok := strings.Contains(logs, expectedLogs)
		if !ok {
			return fmt.Errorf("expected to find %s in the log, got %s", expectedLogs, logs)
		}
		return nil
	})
	if err != nil {
		e.T.Fatalf("unable to finish log comparison: %s", err)
	}
}

//go:embed testdata/emissary_listener.yaml
var emisarryListener []byte

//go:embed testdata/emissary_package.yaml
var emisarryPackage []byte

// VerifyEmissaryPackageInstalled is checking if emissary package gets installed correctly.
func (e *ClusterE2ETest) VerifyEmissaryPackageInstalled(name string, mgmtCluster *types.Cluster) {
	ctx := context.Background()
	ns := constants.EksaPackagesName

	e.T.Log("Waiting for Package", name, "To be installed")
	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		mgmtCluster, name, "5m", fmt.Sprintf("%s-%s", ns, e.ClusterName))
	if err != nil {
		e.T.Fatalf("waiting for emissary package timed out: %s", err)
	}

	e.T.Log("Waiting for Package", name, "Deployment to be healthy")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.cluster(), "5m", "Available", name, ns)
	if err != nil {
		e.T.Fatalf("waiting for emissary deployment timed out: %s", err)
	}
	svcAddress := name + "-admin." + ns + ".svc.cluster.local" + ":8877/ambassador/v0/check_alive"
	randomname := fmt.Sprintf("%s-%s", "busybox-test", utilrand.String(7))
	clientPod, err := e.KubectlClient.RunBusyBoxPod(context.TODO(), ns, randomname, e.kubeconfigFilePath(), []string{"curl", svcAddress})
	if err != nil {
		e.T.Fatalf("error launching busybox pod: %s", err)
	}
	e.T.Log("Launching Busybox pod", clientPod, "to test Package", name)

	err = e.KubectlClient.WaitForPodCompleted(ctx,
		e.cluster(), clientPod, "5m", ns)
	if err != nil {
		e.T.Fatalf("waiting for busybox pod timed out: %s", err)
	}

	e.T.Log("Checking Busybox pod logs", clientPod)
	logs, err := e.KubectlClient.GetPodLogs(context.TODO(), ns, clientPod, clientPod, e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("failure getting pod logs %s", err)
	}
	fmt.Printf("Logs from curl emissary\n %s\n", logs)
	ok := strings.Contains(logs, "Ambassador is alive and well")
	if !ok {
		e.T.Fatalf("expected Ambassador is alive and well, got %s", logs)
	}
}

// TestEmissaryPackageRouting is checking if emissary is able to create Ingress, host, and mapping that function correctly.
func (e *ClusterE2ETest) TestEmissaryPackageRouting(name string, mgmtCluster *types.Cluster) {
	ctx := context.Background()
	ns := constants.EksaPackagesName
	err := e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.cluster(), emisarryPackage)
	if err != nil {
		e.T.Errorf("Error upgrading emissary package: %v", err)
		return
	}
	e.T.Log("Waiting for Package", name, "To be upgraded")
	err = e.KubectlClient.WaitForPackagesInstalled(ctx,
		mgmtCluster, name, "10m", fmt.Sprintf("%s-%s", ns, e.ClusterName))
	if err != nil {
		e.T.Fatalf("waiting for emissary package upgrade timed out: %s", err)
	}
	err = e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.cluster(), emisarryListener)
	if err != nil {
		e.T.Errorf("Error applying roles for oids: %v", err)
		return
	}

	// Functional testing of Emissary Ingress
	ingresssvcAddress := name + "." + ns + ".svc.cluster.local" + "/backend/"
	randomnameIng := fmt.Sprintf("%s-%s", "busybox-test-ing", utilrand.String(7))
	clientPod, err := e.KubectlClient.RunBusyBoxPod(context.TODO(), ns, randomnameIng, e.kubeconfigFilePath(), []string{"curl", ingresssvcAddress})
	if err != nil {
		e.T.Fatalf("error launching busybox pod: %s", err)
	}
	e.T.Log("Launching Busybox pod", clientPod, "to test Package", name)
	err = e.KubectlClient.WaitForPodCompleted(ctx,
		e.cluster(), clientPod, "5m", ns)
	if err != nil {
		e.T.Fatalf("waiting for busybox pod timed out: %s", err)
	}
	e.T.Log("Checking Busybox pod logs", clientPod)
	logs, err := e.KubectlClient.GetPodLogs(context.TODO(), ns, clientPod, clientPod, e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("failure getting pod logs %s", err)
	}
	fmt.Printf("Logs from curl emissary\n %s\n", logs)
	ok := strings.Contains(logs, "quote")
	if !ok {
		e.T.Fatalf("expected quote, got %s", logs)
	}
}

// VerifyPrometheusPackageInstalled is checking if the Prometheus package gets installed correctly.
func (e *ClusterE2ETest) VerifyPrometheusPackageInstalled(packageName string, targetNamespace string) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", "eksa-packages", e.ClusterName)

	e.T.Log("Waiting for package", packageName, "to be installed")
	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		e.cluster(), packageName, "10m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for prometheus package install timed out: %s", err)
	}
}

// VerifyPrometheusPrometheusServerStates is checking if the Prometheus package prometheus-server component is functioning properly.
func (e *ClusterE2ETest) VerifyPrometheusPrometheusServerStates(packageName string, targetNamespace string) {
	ctx := context.Background()

	e.T.Log("Waiting for package", packageName, "deployment prometheus-server to be available")
	err := e.KubectlClient.WaitForDeployment(ctx,
		e.cluster(), "5m", "Available", fmt.Sprintf("%s-server", packageName), targetNamespace)
	if err != nil {
		e.T.Fatalf("waiting for prometheus deployment timed out: %s", err)
	}

	e.T.Log("Reading package", packageName, "pod prometheus-server logs")
	podName, err := e.KubectlClient.GetPodNameByLabel(context.TODO(), targetNamespace, "app=prometheus,component=server", e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("unable to get name of the prometheus-server pod: %s", err)
	}
	logs, err := e.KubectlClient.GetPodLogs(context.TODO(), targetNamespace, podName, "prometheus-server", e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("failure getting pod logs %s", err)
	}
	expectedLogs := "Server is ready to receive web requests"
	fmt.Printf("Logs from pod %s \n %s\n", podName, logs)
	ok := strings.Contains(logs, expectedLogs)
	if !ok {
		e.T.Fatalf("expected to find %s in the log, got %s", expectedLogs, logs)
	}
}

// VerifyPrometheusNodeExporterStates is checking if the Prometheus package node-exporter component is functioning properly.
func (e *ClusterE2ETest) VerifyPrometheusNodeExporterStates(packageName string, targetNamespace string) {
	ctx := context.Background()

	e.T.Log("Waiting for package", packageName, "daemonset node-exporter to be rolled out")
	err := e.KubectlClient.WaitForDaemonsetRolledout(ctx,
		e.cluster(), "5m", fmt.Sprintf("%s-node-exporter", packageName), targetNamespace)
	if err != nil {
		e.T.Fatalf("waiting for prometheus daemonset timed out: %s", err)
	}

	e.T.Log("Launching Busybox pod to test Package", packageName, "service node-exporter")
	svcAddress := packageName + "-node-exporter." + targetNamespace + ".svc.cluster.local" + ":9100/metrics"
	busyBoxName := fmt.Sprintf("%s-%s", "busybox-test", utilrand.String(7))
	clientPod, err := e.KubectlClient.RunBusyBoxPod(context.TODO(), targetNamespace, busyBoxName, e.kubeconfigFilePath(), []string{"curl", svcAddress})
	if err != nil {
		e.T.Fatalf("error launching busybox pod: %s", err)
	}
	e.T.Log("Waiting Busybox pod", clientPod, "to be ready")
	err = e.KubectlClient.WaitForPodCompleted(ctx,
		e.cluster(), clientPod, "5m", targetNamespace)
	if err != nil {
		e.T.Fatalf("waiting for busybox pod timed out: %s", err)
	}
	e.T.Log("Checking Busybox pod logs", clientPod)
	logs, err := e.KubectlClient.GetPodLogs(context.TODO(), targetNamespace, clientPod, clientPod, e.kubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("failure getting pod logs %s", err)
	}
	fmt.Printf("Logs from curl node-exporter service\n %s\n", logs)
	expectedLogs := "HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles"
	ok := strings.Contains(logs, expectedLogs)
	if !ok {
		e.T.Fatalf("expected to find %s in the log, got %s", expectedLogs, logs)
	}
}

// VerifyPackageControllerNotInstalled is verifying that package controller is not installed.
func (e *ClusterE2ETest) VerifyPackageControllerNotInstalled() {
	ctx := context.Background()

	ns := constants.EksaPackagesName
	packageDeployment := "eks-anywhere-packages"

	_, err := e.KubectlClient.GetDeployment(ctx, packageDeployment, ns, e.cluster().KubeconfigFile)

	if !apierrors.IsNotFound(err) {
		e.T.Fatalf("found deployment for package controller in workload cluster %s : %s", e.ClusterName, err)
	}
}

// VerifyAutoScalerPackageInstalled is verifying that the autoscaler package is installed and deployed.
func (e *ClusterE2ETest) VerifyAutoScalerPackageInstalled(name string, targetNamespace string, mgmtCluster *types.Cluster) {
	ctx := context.Background()
	deploymentName := "cluster-autoscaler-clusterapi-cluster-autoscaler"

	e.T.Log("Waiting for Package", name, "To be installed")
	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		mgmtCluster, name, "5m", fmt.Sprintf("%s-%s", targetNamespace, e.ClusterName))
	if err != nil {
		e.T.Fatalf("waiting for Autoscaler Package to be avaliable")
	}

	e.T.Log("Waiting for Package", name, "Deployment to be healthy")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.cluster(), "5m", "Available", deploymentName, targetNamespace)
	if err != nil {
		e.T.Fatalf("waiting for cluster-autoscaler deployment timed out: %s", err)
	}
}

// VerifyMetricServerPackageInstalled is verifying that metrics-server is installed and deployed.
func (e *ClusterE2ETest) VerifyMetricServerPackageInstalled(name string, targetNamespace string, mgmtCluster *types.Cluster) {
	ctx := context.Background()
	deploymentName := "metrics-server"

	e.T.Log("Waiting for Package", name, "To be installed")
	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		mgmtCluster, name, "5m", fmt.Sprintf("%s-%s", targetNamespace, e.ClusterName))
	if err != nil {
		e.T.Fatalf("waiting for Metric Server Package to be avaliable")
	}

	e.T.Log("Waiting for Package", name, "Deployment to be healthy")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.cluster(), "5m", "Available", deploymentName, targetNamespace)
	if err != nil {
		e.T.Fatalf("waiting for Metric Server deployment timed out: %s", err)
	}
}

//go:embed testdata/autoscaler_package.yaml
var autoscalerPackageDeploymentTemplate string

//go:embed testdata/metrics_server_package.yaml
var metricsServerPackageDeploymentTemplate string

// InstallAutoScalerWithMetricServer installs autoscaler and metrics-server with a given target namespace.
func (e *ClusterE2ETest) InstallAutoScalerWithMetricServer(targetNamespace string) {
	ctx := context.Background()
	packageInstallNamespace := fmt.Sprintf("%s-%s", "eksa-packages", e.ClusterName)
	data := map[string]interface{}{
		"targetNamespace": targetNamespace,
		"clusterName":     e.cluster().Name,
	}

	metricsServerPackageDeployment, err := templater.Execute(metricsServerPackageDeploymentTemplate, data)
	if err != nil {
		e.T.Fatalf("Failed creating metrics-erver Package Deployment: %s", err)
	}

	err = e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, e.cluster(), metricsServerPackageDeployment,
		packageInstallNamespace)
	if err != nil {
		e.T.Fatalf("Error installing metrics-sserver pacakge: %s", err)
	}

	autoscalerPackageDeployment, err := templater.Execute(autoscalerPackageDeploymentTemplate, data)
	if err != nil {
		e.T.Fatalf("Failed creating autoscaler Package Deployment: %s", err)
	}

	err = e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, e.cluster(), autoscalerPackageDeployment,
		packageInstallNamespace)
	if err != nil {
		e.T.Fatalf("Error installing cluster autoscaler pacakge: %s", err)
	}
}

// CombinedAutoScalerMetricServerTest verifies that new nodes are spun up after using a HPA to scale a deployment.
func (e *ClusterE2ETest) CombinedAutoScalerMetricServerTest(autoscalerName string, metricServerName string, targetNamespace string, mgmtCluster *types.Cluster) {
	ctx := context.Background()
	ns := "default"
	name := "hpa-busybox-test"
	machineDeploymentName := e.ClusterName + "-" + "md-0"

	e.VerifyMetricServerPackageInstalled(metricServerName, targetNamespace, mgmtCluster)
	e.VerifyAutoScalerPackageInstalled(autoscalerName, targetNamespace, mgmtCluster)

	e.T.Log("Metrics Server and Cluster Autoscaler ready")

	err := e.KubectlClient.ApplyKubeSpecFromBytes(ctx, mgmtCluster, hpaBusybox)
	if err != nil {
		e.T.Fatalf("Failed to apply hpa busybox load %s", err)
	}

	e.T.Log("Deploying test workload")

	err = e.KubectlClient.WaitForDeployment(ctx,
		e.cluster(), "5m", "Available", name, ns)
	if err != nil {
		e.T.Fatalf("Failed waiting for test workload deployent %s", err)
	}

	params := []string{"autoscale", "deployment", name, "--cpu-percent=50", "--min=1", "--max=20", "--kubeconfig", e.kubeconfigFilePath()}
	_, err = e.KubectlClient.ExecuteCommand(ctx, params...)
	if err != nil {
		e.T.Fatalf("Failed to autoscale deployent: %s", err)
	}

	e.T.Log("Waiting for machinedeployment to begin scaling up")
	err = e.KubectlClient.WaitJSONPathLoop(ctx, mgmtCluster.KubeconfigFile, "5m", "status.phase", "ScalingUp",
		fmt.Sprintf("machinedeployments.cluster.x-k8s.io/%s", machineDeploymentName), constants.EksaSystemNamespace)
	if err != nil {
		e.T.Fatalf("Failed to get ScalingUp phase for machinedeployment: %s", err)
	}

	e.T.Log("Waiting for machinedeployment to finish scaling up")
	err = e.KubectlClient.WaitJSONPathLoop(ctx, mgmtCluster.KubeconfigFile, "10m", "status.phase", "Running",
		fmt.Sprintf("machinedeployments.cluster.x-k8s.io/%s", machineDeploymentName), constants.EksaSystemNamespace)
	if err != nil {
		e.T.Fatalf("Failed to get Running phase for machinedeployment: %s", err)
	}

	err = e.KubectlClient.WaitForMachineDeploymentReady(ctx, mgmtCluster, "2m",
		machineDeploymentName)
	if err != nil {
		e.T.Fatalf("Machine deployment stuck in scaling up: %s", err)
	}

	e.T.Log("Finished scaling up machines")
}

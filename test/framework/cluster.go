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
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bmc-toolbox/bmclib/v2"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

const (
	defaultClusterConfigFile               = "cluster.yaml"
	defaultBundleReleaseManifestFile       = "bin/local-bundle-release.yaml"
	defaultEksaBinaryLocation              = "eksctl anywhere"
	defaultClusterName                     = "eksa-test"
	defaultDownloadArtifactsOutputLocation = "eks-anywhere-downloads.tar.gz"
	defaultDownloadImagesOutputLocation    = "images.tar"
	eksctlVersionEnvVar                    = "EKSCTL_VERSION"
	eksctlVersionEnvVarDummyVal            = "ham sandwich"
	ClusterPrefixVar                       = "T_CLUSTER_PREFIX"
	JobIdVar                               = "T_JOB_ID"
	BundlesOverrideVar                     = "T_BUNDLES_OVERRIDE"
	ClusterIPPoolEnvVar                    = "T_CLUSTER_IP_POOL"
	CleanupVmsVar                          = "T_CLEANUP_VMS"
	hardwareYamlPath                       = "hardware.yaml"
	hardwareCsvPath                        = "hardware.csv"
	EksaPackagesInstallation               = "eks-anywhere-packages"
)

//go:embed testdata/oidc-roles.yaml
var oidcRoles []byte

//go:embed testdata/hpa_busybox.yaml
var hpaBusybox []byte

type ClusterE2ETest struct {
	T                            T
	ClusterConfigLocation        string
	ClusterConfigFolder          string
	HardwareConfigLocation       string
	HardwareCsvLocation          string
	TestHardware                 map[string]*api.Hardware
	HardwarePool                 map[string]*api.Hardware
	WithNoPowerActions           bool
	WithOOBConfiguration         bool
	ClusterName                  string
	ClusterConfig                *cluster.Config
	clusterStateValidationConfig *clusterf.StateValidationConfig
	Provider                     Provider
	// TODO(g-gaston): migrate uses of clusterFillers to clusterConfigFillers
	clusterFillers       []api.ClusterFiller
	clusterConfigFillers []api.ClusterConfigFiller
	KubectlClient        *executables.Kubectl
	GitProvider          git.ProviderClient
	GitClient            git.Client
	HelmInstallConfig    *HelmInstallConfig
	PackageConfig        *PackageConfig
	GitWriter            filewriter.FileWriter
	eksaBinaryLocation   string
	OSFamily             v1alpha1.OSFamily
	ExpectFailure        bool
	// PersistentCluster avoids creating the clusters if it finds a kubeconfig
	// in the corresponding cluster folder. Useful for local development of tests.
	// When generating a new base cluster config, it will read from disk instead of
	// using the CLI generate command and will preserve the previous CP endpoint.
	PersistentCluster bool
}

type ClusterE2ETestOpt func(e *ClusterE2ETest)

// NewClusterE2ETest is a support structure for defining an end-to-end test.
func NewClusterE2ETest(t T, provider Provider, opts ...ClusterE2ETestOpt) *ClusterE2ETest {
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

	for _, opt := range opts {
		opt(e)
	}

	if e.ClusterConfigFolder == "" {
		e.ClusterConfigFolder = e.ClusterName
	}
	if e.HardwareConfigLocation == "" {
		e.HardwareConfigLocation = filepath.Join(e.ClusterConfigFolder, hardwareYamlPath)
	}
	if e.HardwareCsvLocation == "" {
		e.HardwareCsvLocation = filepath.Join(e.ClusterConfigFolder, hardwareCsvPath)
	}

	e.ClusterConfigLocation = filepath.Join(e.ClusterConfigFolder, e.ClusterName+"-eks-a.yaml")

	if err := os.MkdirAll(e.ClusterConfigFolder, os.ModePerm); err != nil {
		t.Fatalf("Failed creating cluster config folder for test: %s", err)
	}

	provider.Setup()

	e.T.Cleanup(func() {
		e.CleanupVms()

		tinkerbellCIEnvironment := os.Getenv(TinkerbellCIEnvironment)
		if e.Provider.Name() == tinkerbellProviderName && tinkerbellCIEnvironment == "true" {
			e.CleanupDockerEnvironment()
		}
	})

	return e
}

// UpdateClusterName updates the cluster name for the test. This will drive both the name of the eks-a
// cluster config objects as well as the cluster config file name and the folder where the cluster config
// file is stored.
// The cluster config folder will be updated to the new cluster name only if it was using the default value.
func (e *ClusterE2ETest) UpdateClusterName(name string) {
	oldName := e.ClusterName
	e.ClusterName = name
	if e.ClusterConfigFolder == oldName {
		// Only update the folder if it was using the old name. This is the default value.
		e.ClusterConfigFolder = e.ClusterName
		if err := os.MkdirAll(e.ClusterConfigFolder, os.ModePerm); err != nil {
			e.T.Fatalf("Failed creating cluster config folder for test: %s", err)
		}
	}
	e.ClusterConfigLocation = filepath.Join(e.ClusterConfigFolder, e.ClusterName+"-eks-a.yaml")
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
	return withHardware(
		requiredCount,
		api.ControlPlane,
		map[string]string{api.HardwareLabelTypeKeyName: api.ControlPlane},
	)
}

func WithWorkerHardware(requiredCount int) ClusterE2ETestOpt {
	return withHardware(requiredCount, api.Worker, map[string]string{api.HardwareLabelTypeKeyName: api.Worker})
}

func WithCustomLabelHardware(requiredCount int, label string) ClusterE2ETestOpt {
	return withHardware(requiredCount, api.Worker, map[string]string{api.HardwareLabelTypeKeyName: label})
}

func WithExternalEtcdHardware(requiredCount int) ClusterE2ETestOpt {
	return withHardware(
		requiredCount,
		api.ExternalEtcd,
		map[string]string{api.HardwareLabelTypeKeyName: api.ExternalEtcd},
	)
}

// WithClusterName sets the name that will be used for the cluster. This will drive both the name of the eks-a
// cluster config objects as well as the cluster config file name.
func WithClusterName(name string) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ClusterName = name
	}
}

// PersistentCluster  avoids creating the clusters if it finds a kubeconfig
// in the corresponding cluster folder. Useful for local development of tests.
func PersistentCluster() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.PersistentCluster = true
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
	UpdateKubeConfig(content *[]byte, clusterName string) error
	ClusterStateValidations() []clusterf.StateValidation
	WithKubeVersionAndOS(kubeVersion v1alpha1.KubernetesVersion, os OS, release *releasev1.EksARelease) api.ClusterConfigFiller
	WithNewWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup) api.ClusterConfigFiller
}

func (e *ClusterE2ETest) GenerateClusterConfig(opts ...CommandOpt) {
	e.GenerateClusterConfigForVersion("", opts...)
}

func newBmclibClient(log logr.Logger, hostIP, username, password string) *bmclib.Client {
	o := []bmclib.Option{}
	log = log.WithValues("host", hostIP, "username", username)
	o = append(o, bmclib.WithLogger(log))
	client := bmclib.NewClient(hostIP, username, password, o...)
	client.Registry.Drivers = client.Registry.PreferProtocol("redfish")

	return client
}

// powerOffHardware issues power off calls to all Hardware. This function does not fail the test if it encounters an error.
// This function is a helper and not part of the code path that we are testing.
// For this reason, we are only logging the errors and not failing the test.
// This function exists not because we need the hardware to be powered off before a test run,
// but because we want to make sure that no other Tinkerbell Boots DHCP server is running.
// Another Boots DHCP server running can cause netboot issues with hardware.
func (e *ClusterE2ETest) powerOffHardware() {
	for _, h := range e.TestHardware {
		ctx, done := context.WithTimeout(context.Background(), 2*time.Minute)
		defer done()
		bmcClient := newBmclibClient(logr.Discard(), h.BMCIPAddress, h.BMCUsername, h.BMCPassword)

		if err := bmcClient.Open(ctx); err != nil {
			md := bmcClient.GetMetadata()
			e.T.Logf("Failed to open connection to BMC: %v, hardware: %v, providersAttempted: %v, failedProviderDetail: %v", err, h.BMCIPAddress, md.ProvidersAttempted, md.SuccessfulOpenConns)

			continue
		}
		md := bmcClient.GetMetadata()
		e.T.Logf("Connected to BMC: hardware: %v, providersAttempted: %v, successfulProvider: %v", h.BMCIPAddress, md.ProvidersAttempted, md.SuccessfulOpenConns)

		defer func() {
			if err := bmcClient.Close(ctx); err != nil {
				md := bmcClient.GetMetadata()
				e.T.Logf("BMC close connection failed: %v, hardware: %v, providersAttempted: %v, failedProviderDetail: %v", err, h.BMCIPAddress, md.ProvidersAttempted, md.FailedProviderDetail)
			}
		}()

		state, err := bmcClient.GetPowerState(ctx)
		if err != nil {
			state = "unknown"
		}
		if strings.Contains(strings.ToLower(state), "off") {
			return
		}

		if _, err := bmcClient.SetPowerState(ctx, "off"); err != nil {
			md := bmcClient.GetMetadata()
			e.T.Logf("failed to power off hardware: %v, hardware: %v, providersAttempted: %v, failedProviderDetail: %v", err, h.BMCIPAddress, md.ProvidersAttempted, md.SuccessfulOpenConns)
			continue
		}
	}
}

// ValidateHardwareDecommissioned checks that the all hardware was powered off during the cluster deletion.
// This function tests that the hardware was powered off during the cluster deletion. If any hardware are not powered off
// this func calls powerOffHardware to power off the hardware and then fails this test.
func (e *ClusterE2ETest) ValidateHardwareDecommissioned() {
	var failedToDecomm []*api.Hardware
	for _, h := range e.TestHardware {
		ctx, done := context.WithTimeout(context.Background(), 2*time.Minute)
		defer done()
		bmcClient := newBmclibClient(logr.Discard(), h.BMCIPAddress, h.BMCUsername, h.BMCPassword)

		if err := bmcClient.Open(ctx); err != nil {
			md := bmcClient.GetMetadata()
			e.T.Logf("Failed to open connection to BMC: %v, hardware: %v, providersAttempted: %v, failedProviderDetail: %v", err, h.BMCIPAddress, md.ProvidersAttempted, md.SuccessfulOpenConns)

			continue
		}
		md := bmcClient.GetMetadata()
		e.T.Logf("Connected to BMC: hardware: %v, providersAttempted: %v, successfulProvider: %v", h.BMCIPAddress, md.ProvidersAttempted, md.SuccessfulOpenConns)

		defer func() {
			if err := bmcClient.Close(ctx); err != nil {
				md := bmcClient.GetMetadata()
				e.T.Logf("BMC close connection failed: %v, hardware: %v, providersAttempted: %v, failedProviderDetail: %v", err, h.BMCIPAddress, md.ProvidersAttempted, md.FailedProviderDetail)
			}
		}()

		// add sleep retries to give the machine time to power off
		wait := 5 * time.Second
		for tries := 1; tries <= 4; tries++ {
			time.Sleep(wait)
			powerState, err := bmcClient.GetPowerState(ctx)
			e.T.Logf("hardware power state (id=%s, hostname=%s, bmc_ip=%s): power_state=%s", h.MACAddress, h.Hostname, h.BMCIPAddress, powerState)
			if err != nil {
				md := bmcClient.GetMetadata()
				e.T.Logf("failed to get power state for hardware: id=%s, hostname=%s, bmc_ip=%s, providersAttempted: %v, failedProviderDetail: %v",
					h.MACAddress,
					h.Hostname,
					h.BMCIPAddress,
					md.ProvidersAttempted,
					md.FailedProviderDetail,
				)
				continue
			}
			if !strings.Contains(strings.ToLower(powerState), "off") {
				e.T.Logf("failed to decommission hardware: id=%s, hostname=%s, bmc_ip=%s, powerState=%v", h.MACAddress, h.Hostname, h.BMCIPAddress, powerState)
				failedToDecomm = append(failedToDecomm, h)
			} else {
				e.T.Logf("successfully decommissioned hardware: id=%s, hostname=%s, bmc_ip=%s", h.MACAddress, h.Hostname, h.BMCIPAddress)
				break
			}
		}
	}

	if len(failedToDecomm) > 0 {
		e.powerOffHardware()
		e.T.Fatalf("failed to decommission all hardware during cluster deletion")
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
	// create hardware CSV with no bmc username/password
	if e.WithOOBConfiguration {
		hardwareWithNoUsernamePassword := make(map[string]*api.Hardware)
		for k, h := range testHardware {
			lessBmc := *h
			lessBmc.BMCUsername = ""
			lessBmc.BMCPassword = ""
			hardwareWithNoUsernamePassword[k] = &lessBmc
		}
		testHardware = hardwareWithNoUsernamePassword
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
}

func (e *ClusterE2ETest) generateClusterConfigObjects(opts ...CommandOpt) {
	e.generateClusterConfigWithCLI(opts...)
	config, err := cluster.ParseConfigFromFile(e.ClusterConfigLocation)
	if err != nil {
		e.T.Fatalf("Failed parsing generated cluster config: %s", err)
	}

	// Copy all objects that might be generated by the CLI.
	// Don't replace the whole ClusterConfig since some ClusterE2ETestOpt might
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
	e.ClusterConfig.SnowIPPools = config.SnowIPPools
	e.ClusterConfig.NutanixMachineConfigs = config.NutanixMachineConfigs
	e.ClusterConfig.TinkerbellMachineConfigs = config.TinkerbellMachineConfigs
	e.ClusterConfig.TinkerbellTemplateConfigs = config.TinkerbellTemplateConfigs

	e.UpdateClusterConfig(e.baseClusterConfigUpdates()...)
}

// UpdateClusterConfig applies the cluster Config provided updates to e.ClusterConfig, marshalls its content
// to yaml and writes it to a file on disk configured by e.ClusterConfigLocation. Call this method when you want
// make changes to the eks-a cluster definition before running a CLI command or API operation.
func (e *ClusterE2ETest) UpdateClusterConfig(fillers ...api.ClusterConfigFiller) {
	e.T.Log("Updating cluster config")
	api.UpdateClusterConfig(e.ClusterConfig, fillers...)
	e.T.Logf("Writing cluster config to file: %s", e.ClusterConfigLocation)
	e.buildClusterConfigFile()
}

func (e *ClusterE2ETest) baseClusterConfigUpdates(opts ...CommandOpt) []api.ClusterConfigFiller {
	clusterFillers := make([]api.ClusterFiller, 0, len(e.clusterFillers)+3)
	// This defaults all tests to a 1:1:1 configuration. Since all the fillers defined on each test are run
	// after these 3, if the tests is explicit about any of these, the defaults will be overwritten
	clusterFillers = append(clusterFillers,
		api.WithControlPlaneCount(1), api.WithWorkerNodeCount(1), api.WithEtcdCountIfExternal(1),
	)
	clusterFillers = append(clusterFillers, e.clusterFillers...)
	configFillers := make([]api.ClusterConfigFiller, 0, len(e.clusterConfigFillers)+1)
	configFillers = append(configFillers, api.ClusterToConfigFiller(clusterFillers...))
	configFillers = append(configFillers, e.clusterConfigFillers...)
	configFillers = append(configFillers, e.Provider.ClusterConfigUpdates()...)

	// If we are persisting an existing cluster, set the control plane endpoint back to the original, since
	// it is immutable
	if e.ClusterConfig.Cluster.Spec.DatacenterRef.Kind != v1alpha1.DockerDatacenterKind && e.PersistentCluster && e.ClusterConfig.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host != "" {
		endpoint := e.ClusterConfig.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host
		e.T.Logf("Resetting CP endpoint for persistent cluster to %s", endpoint)
		configFillers = append(configFillers,
			api.ClusterToConfigFiller(api.WithControlPlaneEndpointIP(endpoint)),
		)
	}

	return configFillers
}

func (e *ClusterE2ETest) generateClusterConfigWithCLI(opts ...CommandOpt) {
	if e.PersistentCluster && fileExists(e.ClusterConfigLocation) {
		e.T.Log("Skipping CLI cluster generation since this is a persistent cluster that already had one cluster config generated")
		return
	}

	if err := os.MkdirAll(filepath.Dir(e.ClusterConfigLocation), os.ModePerm); err != nil {
		e.T.Fatalf("Failed creating cluster config folder for test: %s", err)
	}

	generateClusterConfigArgs := []string{"generate", "clusterconfig", e.ClusterName, "-p", e.Provider.Name(), ">", e.ClusterConfigLocation}
	e.RunEKSA(generateClusterConfigArgs, opts...)
	e.T.Log("Cluster config generated with CLI")
}

func (e *ClusterE2ETest) parseClusterConfigFromDisk(file string) {
	e.T.Logf("Parsing cluster config from disk: %s", file)
	config, err := cluster.ParseConfigFromFile(file)
	if err != nil {
		e.T.Fatalf("Failed parsing generated cluster config: %s", err)
	}
	e.ClusterConfig = config
}

// WithClusterConfig generates a base cluster config using the CLI `generate clusterconfig` command
// and updates them with the provided fillers. Helpful for defining the initial Cluster config
// before running a create operation.
func (e *ClusterE2ETest) WithClusterConfig(fillers ...api.ClusterConfigFiller) *ClusterE2ETest {
	e.T.Logf("Generating base config for cluster %s", e.ClusterName)
	e.generateClusterConfigWithCLI()
	e.parseClusterConfigFromDisk(e.ClusterConfigLocation)
	base := e.baseClusterConfigUpdates()
	allUpdates := make([]api.ClusterConfigFiller, 0, len(base)+len(fillers))
	allUpdates = append(allUpdates, base...)
	allUpdates = append(allUpdates, fillers...)
	e.UpdateClusterConfig(allUpdates...)
	return e
}

// DownloadArtifacts runs the EKS-A `download artifacts` command with appropriate args.
func (e *ClusterE2ETest) DownloadArtifacts(opts ...CommandOpt) {
	downloadArtifactsArgs := []string{"download", "artifacts", "-f", e.ClusterConfigLocation}
	if getBundlesOverride() == "true" {
		downloadArtifactsArgs = append(downloadArtifactsArgs, "--bundles-override", defaultBundleReleaseManifestFile)
	}
	e.RunEKSA(downloadArtifactsArgs, opts...)
	if _, err := os.Stat(defaultDownloadArtifactsOutputLocation); err != nil {
		e.T.Fatal(err)
	} else {
		e.T.Logf("Downloaded artifacts tarball saved at %s", defaultDownloadArtifactsOutputLocation)
	}
}

// ExtractDownloadedArtifacts extracts the downloaded artifacts.
func (e *ClusterE2ETest) ExtractDownloadedArtifacts(opts ...CommandOpt) {
	e.T.Log("Extracting downloaded artifacts")
	e.Run("tar", "-xf", defaultDownloadArtifactsOutputLocation)
}

// CleanupDownloadedArtifactsAndImages cleans up the downloaded artifacts and images.
func (e *ClusterE2ETest) CleanupDownloadedArtifactsAndImages(opts ...CommandOpt) {
	e.T.Log("Cleaning up downloaded artifacts and images")
	e.Run("rm", "-rf", defaultDownloadArtifactsOutputLocation, defaultDownloadImagesOutputLocation)
}

// DownloadImages runs the EKS-A `download images` command with appropriate args.
func (e *ClusterE2ETest) DownloadImages(opts ...CommandOpt) {
	downloadImagesArgs := []string{"download", "images", "-o", defaultDownloadImagesOutputLocation}
	if getBundlesOverride() == "true" {
		var bundleManifestLocation string
		if _, err := os.Stat(defaultDownloadArtifactsOutputLocation); err == nil {
			bundleManifestLocation = "eks-anywhere-downloads/bundle-release.yaml"
		} else {
			bundleManifestLocation = defaultBundleReleaseManifestFile
		}
		downloadImagesArgs = append(downloadImagesArgs, "--bundles-override", bundleManifestLocation)
	}
	e.RunEKSA(downloadImagesArgs, opts...)
	if _, err := os.Stat(defaultDownloadImagesOutputLocation); err != nil {
		e.T.Fatal(err)
	} else {
		e.T.Logf("Downloaded images archive saved at %s", defaultDownloadImagesOutputLocation)
	}
}

// ImportImages runs the EKS-A `import images` command with appropriate args.
func (e *ClusterE2ETest) ImportImages(opts ...CommandOpt) {
	clusterConfig := e.ClusterConfig.Cluster
	registyMirrorEndpoint, registryMirrorPort := clusterConfig.Spec.RegistryMirrorConfiguration.Endpoint, clusterConfig.Spec.RegistryMirrorConfiguration.Port
	registryMirrorHost := net.JoinHostPort(registyMirrorEndpoint, registryMirrorPort)
	var bundleManifestLocation string
	if _, err := os.Stat(defaultDownloadArtifactsOutputLocation); err == nil {
		bundleManifestLocation = "eks-anywhere-downloads/bundle-release.yaml"
	} else {
		bundleManifestLocation = defaultBundleReleaseManifestFile
	}
	importImagesArgs := []string{"import images", "--input", defaultDownloadImagesOutputLocation, "--bundles", bundleManifestLocation, "--registry", registryMirrorHost, "--insecure"}
	e.RunEKSA(importImagesArgs, opts...)
}

// ChangeInstanceSecurityGroup modifies the security group of the instance to the provided value.
func (e *ClusterE2ETest) ChangeInstanceSecurityGroup(securityGroup string) {
	e.T.Logf("Changing instance security group to %s", securityGroup)
	e.Run(fmt.Sprintf("INSTANCE_ID=$(ec2-metadata -i | awk '{print $2}') && aws ec2 modify-instance-attribute --instance-id $INSTANCE_ID --groups %s", securityGroup))
}

func (e *ClusterE2ETest) CreateCluster(opts ...CommandOpt) {
	e.setFeatureFlagForUnreleasedKubernetesVersion(e.ClusterConfig.Cluster.Spec.KubernetesVersion)
	e.createCluster(opts...)
}

func (e *ClusterE2ETest) createCluster(opts ...CommandOpt) {
	if e.PersistentCluster {
		if fileExists(e.KubeconfigFilePath()) {
			e.T.Logf("Persisent cluster: kubeconfig found for cluster %s, skipping cluster creation", e.ClusterName)
			return
		}
	}

	e.T.Logf("Creating cluster %s", e.ClusterName)
	createClusterArgs := []string{"create", "cluster", "-f", e.ClusterConfigLocation, "-v", "12"}

	dumpFile("Create cluster from file:", e.ClusterConfigLocation, e.T)

	if getBundlesOverride() == "true" {
		createClusterArgs = append(createClusterArgs, "--bundles-override", defaultBundleReleaseManifestFile)
	}

	if e.Provider.Name() == tinkerbellProviderName {
		createClusterArgs = append(createClusterArgs, "-z", e.HardwareCsvLocation)
		dumpFile("Hardware csv file:", e.HardwareCsvLocation, e.T)
		tinkBootstrapIP := os.Getenv(tinkerbellBootstrapIPEnvVar)
		e.T.Logf("tinkBootstrapIP: %s", tinkBootstrapIP)
		if tinkBootstrapIP != "" {
			createClusterArgs = append(createClusterArgs, "--tinkerbell-bootstrap-ip", tinkBootstrapIP)
		}
	}

	e.RunEKSA(createClusterArgs, opts...)
}

func (e *ClusterE2ETest) ValidateCluster(kubeVersion v1alpha1.KubernetesVersion) {
	ctx := context.Background()
	e.T.Log("Validating cluster node status")
	r := retrier.New(10 * time.Minute)
	err := r.Retry(func() error {
		err := e.KubectlClient.ValidateNodes(ctx, e.Cluster().KubeconfigFile)
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
		if err = e.KubectlClient.ValidateNodesVersion(ctx, e.Cluster().KubeconfigFile, kubeVersion); err != nil {
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
	err := e.KubectlClient.WaitForMachineDeploymentReady(ctx, e.Cluster(), "5m", machineDeploymentName)
	if err != nil {
		e.T.Fatal(err)
	}
}

// GetEKSACluster retrieves the EKSA cluster from the runtime environment using kubectl.
func (e *ClusterE2ETest) GetEKSACluster() *v1alpha1.Cluster {
	ctx := context.Background()
	clus, err := e.KubectlClient.GetEksaCluster(ctx, e.Cluster(), e.ClusterName)
	if err != nil {
		e.T.Fatal(err)
	}
	return clus
}

func (e *ClusterE2ETest) GetCapiMachinesForCluster(clusterName string) map[string]types.Machine {
	machines, err := e.CapiMachinesForCluster(clusterName)
	if err != nil {
		e.T.Fatal(err)
	}
	return machines
}

// CapiMachinesForCluster reads all the CAPI Machines for a particular cluster and returns them
// index by their name.
func (e *ClusterE2ETest) CapiMachinesForCluster(clusterName string) (map[string]types.Machine, error) {
	ctx := context.Background()
	capiMachines, err := e.KubectlClient.GetMachines(ctx, e.Cluster(), clusterName)
	if err != nil {
		return nil, err
	}
	machinesMap := make(map[string]types.Machine, 0)
	for _, machine := range capiMachines {
		machinesMap[machine.Metadata.Name] = machine
	}
	return machinesMap, nil
}

// ApplyClusterManifest uses client-side logic to create/update objects defined in a cluster yaml manifest.
func (e *ClusterE2ETest) ApplyClusterManifest() {
	ctx := context.Background()
	e.T.Logf("Applying cluster %s spec located at %s", e.ClusterName, e.ClusterConfigLocation)
	e.applyClusterManifest(ctx)
}

func (e *ClusterE2ETest) applyClusterManifest(ctx context.Context) {
	if err := e.KubectlClient.ApplyManifest(ctx, e.KubeconfigFilePath(), e.ClusterConfigLocation); err != nil {
		e.T.Fatalf("Failed to apply cluster config: %s", err)
	}
}

// WithClusterUpgrade adds a cluster upgrade.
func WithClusterUpgrade(fillers ...api.ClusterFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.ClusterToConfigFiller(fillers...))
	}
}

// WithUpgradeClusterConfig adds a cluster upgrade.
// When we migrate usages of ClusterFiller to ClusterConfigFiller we can rename this to WithClusterUpgrade.
func WithUpgradeClusterConfig(fillers ...api.ClusterConfigFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(fillers...)
	}
}

// LoadClusterConfigGeneratedByCLI loads the full cluster config from the file generated when a cluster is created using the CLI.
func (e *ClusterE2ETest) LoadClusterConfigGeneratedByCLI(fillers ...api.ClusterConfigFiller) {
	fullClusterConfigLocation := filepath.Join(e.ClusterConfigFolder, e.ClusterName+"-eks-a-cluster.yaml")
	e.parseClusterConfigFromDisk(fullClusterConfigLocation)
}

// UpgradeClusterWithNewConfig applies the test options, re-generates the cluster config file and runs the CLI upgrade command.
func (e *ClusterE2ETest) UpgradeClusterWithNewConfig(clusterOpts []ClusterE2ETestOpt, commandOpts ...CommandOpt) {
	e.upgradeCluster(clusterOpts, commandOpts...)
}

func (e *ClusterE2ETest) upgradeCluster(clusterOpts []ClusterE2ETestOpt, commandOpts ...CommandOpt) {
	for _, opt := range clusterOpts {
		opt(e)
	}
	e.buildClusterConfigFile()
	e.setFeatureFlagForUnreleasedKubernetesVersion(e.ClusterConfig.Cluster.Spec.KubernetesVersion)
	e.UpgradeCluster(commandOpts...)
}

// UpgradeCluster runs the CLI upgrade command.
func (e *ClusterE2ETest) UpgradeCluster(commandOpts ...CommandOpt) {
	upgradeClusterArgs := []string{"upgrade", "cluster", "-f", e.ClusterConfigLocation, "-v", "9"}
	if getBundlesOverride() == "true" {
		upgradeClusterArgs = append(upgradeClusterArgs, "--bundles-override", defaultBundleReleaseManifestFile)
	}

	e.RunEKSA(upgradeClusterArgs, commandOpts...)
}

func (e *ClusterE2ETest) generateClusterConfigYaml() []byte {
	childObjs := e.ClusterConfig.ChildObjects()
	yamlB := make([][]byte, 0, len(childObjs)+1)

	if e.PackageConfig != nil {
		e.ClusterConfig.Cluster.Spec.Packages = e.PackageConfig.packageConfiguration
	}

	// This is required because Flux requires a namespace be specified for objects
	// to be able to reconcile right.
	if e.ClusterConfig.Cluster.Namespace == "" {
		e.ClusterConfig.Cluster.Namespace = "default"
	}
	clusterConfigB, err := yaml.Marshal(e.ClusterConfig.Cluster)
	if err != nil {
		e.T.Fatal(err)
	}
	yamlB = append(yamlB, clusterConfigB)
	for _, o := range childObjs {
		// This is required because Flux requires a namespace be specified for objects
		// to be able to reconcile right.
		if o.GetNamespace() == "" {
			o.SetNamespace("default")
		}
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
	e.T.Logf("Written cluster config to %v", writtenFile)
	e.ClusterConfigLocation = writtenFile
}

func (e *ClusterE2ETest) DeleteCluster(opts ...CommandOpt) {
	e.deleteCluster(opts...)
}

// CleanupVms is a helper to clean up VMs. It is a noop if the T_CLEANUP_VMS environment variable
// is false or unset.
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

// GenerateSupportBundleOnCleanupIfTestFailed does what it says on the tin.
//
// It uses testing.T.Cleanup to register a handler that checks if the test
// failed, and generates a support bundle only in the event of a failure.
func (e *ClusterE2ETest) GenerateSupportBundleOnCleanupIfTestFailed(opts ...CommandOpt) {
	e.T.Cleanup(func() {
		if e.T.Failed() {
			e.T.Log("Generating support bundle for failed test")
			generateSupportBundleArgs := []string{"generate", "support-bundle", "-f", e.ClusterConfigLocation}
			e.RunEKSA(generateSupportBundleArgs, opts...)
		}
	})
}

func (e *ClusterE2ETest) Run(name string, args ...string) {
	cmd, err := prepareCommand(name, args...)
	if err != nil {
		e.T.Fatalf("Error preparing command: %v", err)
	}
	e.T.Log("Running shell command", "[", cmd.String(), "]")

	var stdoutAndErr bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stdoutAndErr)
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutAndErr)

	if err = cmd.Run(); err != nil {
		e.T.Log("Command failed, scanning output for error")
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

// Cluster builds a cluster obj using the ClusterE2ETest name and kubeconfig.
func (e *ClusterE2ETest) Cluster() *types.Cluster {
	return &types.Cluster{
		Name:           e.ClusterName,
		KubeconfigFile: e.KubeconfigFilePath(),
	}
}

func (e *ClusterE2ETest) managementCluster() *types.Cluster {
	return &types.Cluster{
		Name:           e.ClusterConfig.Cluster.ManagedBy(),
		KubeconfigFile: e.managementKubeconfigFilePath(),
	}
}

// KubeconfigFilePath retrieves the Kubeconfig path used for the workload cluster.
func (e *ClusterE2ETest) KubeconfigFilePath() string {
	return filepath.Join(e.ClusterConfigFolder, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", e.ClusterName))
}

// BuildWorkloadClusterClient creates a client for the workload cluster created by e.
func (e *ClusterE2ETest) BuildWorkloadClusterClient() (client.Client, error) {
	var clusterClient client.Client
	// Adding the retry logic here because the connection to the client does not always
	// succedd on the first try due to connection failure after the kubeconfig becomes
	// available in the cluster.
	err := retrier.Retry(12, 5*time.Second, func() error {
		c, err := kubernetes.NewRuntimeClientFromFileName(e.KubeconfigFilePath())
		if err != nil {
			return fmt.Errorf("failed to build cluster client: %v", err)
		}
		clusterClient = c
		return nil
	})

	return clusterClient, err
}

func (e *ClusterE2ETest) managementKubeconfigFilePath() string {
	clusterConfig := e.ClusterConfig.Cluster
	if clusterConfig.IsSelfManaged() {
		return e.KubeconfigFilePath()
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

	kubeconfig := e.KubeconfigFilePath()
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

func GetTestNameHash(name string) string {
	h := sha1.New()
	h.Write([]byte(name))
	testNameHash := fmt.Sprintf("%x", h.Sum(nil))
	return testNameHash[:7]
}

func getClusterName(t T) string {
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
			return fmt.Errorf(
				"couldn't set eksctl version env var %s to value %s",
				eksctlVersionEnvVar,
				eksctlVersionEnvVarDummyVal,
			)
		}
	}
	return nil
}

func (e *ClusterE2ETest) InstallHelmChart() {
	kubeconfig := e.KubeconfigFilePath()
	ctx := context.Background()

	err := e.HelmInstallConfig.HelmClient.InstallChart(ctx, e.HelmInstallConfig.chartName, e.HelmInstallConfig.chartURI, e.HelmInstallConfig.chartVersion, kubeconfig, "", "", false, e.HelmInstallConfig.chartValues)
	if err != nil {
		e.T.Fatalf("Error installing %s helm chart on the cluster: %v", e.HelmInstallConfig.chartName, err)
	}
}

// CreateNamespace creates a namespace.
func (e *ClusterE2ETest) CreateNamespace(namespace string) {
	kubeconfig := e.KubeconfigFilePath()
	err := e.KubectlClient.CreateNamespace(context.Background(), kubeconfig, namespace)
	if err != nil {
		e.T.Fatalf("Namespace creation failed for %s", namespace)
	}
}

// DeleteNamespace deletes a namespace.
func (e *ClusterE2ETest) DeleteNamespace(namespace string) {
	kubeconfig := e.KubeconfigFilePath()
	err := e.KubectlClient.DeleteNamespace(context.Background(), kubeconfig, namespace)
	if err != nil {
		e.T.Fatalf("Namespace deletion failed for %s", namespace)
	}
}

// SetPackageBundleActive will set the current packagebundle to the active state.
func (e *ClusterE2ETest) SetPackageBundleActive() {
	kubeconfig := e.KubeconfigFilePath()
	pbc, err := e.KubectlClient.GetPackageBundleController(context.Background(), kubeconfig, e.ClusterName)
	if err != nil {
		e.T.Fatalf("Error getting PackageBundleController: %v", err)
	}
	pb, err := e.KubectlClient.GetPackageBundleList(context.Background(), e.KubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("Error getting PackageBundle: %v", err)
	}
	os.Setenv("KUBECONFIG", kubeconfig)
	if pbc.Spec.ActiveBundle != pb[0].ObjectMeta.Name {
		e.RunEKSA([]string{
			"upgrade", "packages",
			"--bundle-version", pb[0].ObjectMeta.Name, "-v=9",
			"--cluster=" + e.ClusterName,
		})
	}
}

// ValidatingNoPackageController make sure there is no package controller.
func (e *ClusterE2ETest) ValidatingNoPackageController() {
	kubeconfig := e.KubeconfigFilePath()
	_, err := e.KubectlClient.GetPackageBundleController(context.Background(), kubeconfig, e.ClusterName)
	if err == nil {
		e.T.Fatalf("Error unexpected PackageBundleController: %v", err)
	}
}

// InstallCuratedPackage will install a curated package.
func (e *ClusterE2ETest) InstallCuratedPackage(packageName, packagePrefix, kubeconfig string, opts ...string) {
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
	err := e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.Cluster(), []byte(resource))
	if err != nil {
		e.T.Fatalf("Failed to create required resource (%s): %v", resource, err)
	}
}

func (e *ClusterE2ETest) UninstallCuratedPackage(packagePrefix string, opts ...string) {
	e.RunEKSA([]string{
		"delete", "package", packagePrefix, "-v=9",
		"--cluster=" + e.ClusterName,
		strings.Join(opts, " "),
	})
}

func (e *ClusterE2ETest) InstallLocalStorageProvisioner() {
	ctx := context.Background()
	_, err := e.KubectlClient.ExecuteCommand(ctx, "apply", "-f",
		"https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.22/deploy/local-path-storage.yaml",
		"--kubeconfig", e.KubeconfigFilePath())
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
	configPath := e.KubeconfigFilePath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		e.GenerateClusterConfig()
		e.CreateCluster()
	}
	f(e)
}

// VerifyHarborPackageInstalled is checking if the harbor package gets installed correctly.
func (e *ClusterE2ETest) VerifyHarborPackageInstalled(prefix, namespace string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deployments := []string{"core", "jobservice", "nginx", "portal", "registry"}
	statefulsets := []string{"database", "redis", "trivy"}

	var wg sync.WaitGroup
	wg.Add(len(deployments) + len(statefulsets))
	errCh := make(chan error, 1)
	okCh := make(chan string, 1)

	time.Sleep(5 * time.Minute)

	// Log Package/Deployment outputs
	defer func() {
		e.printDeploymentSpec(ctx, namespace)
	}()

	for _, name := range deployments {
		go func(name string) {
			defer wg.Done()
			err := e.KubectlClient.WaitForDeployment(ctx,
				e.Cluster(), "20m", "Available", fmt.Sprintf("%s-harbor-%s", prefix, name), namespace)
			if err != nil {
				errCh <- err
			}
		}(name)
	}
	for _, name := range statefulsets {
		go func(name string) {
			defer wg.Done()
			err := e.KubectlClient.Wait(ctx, e.KubeconfigFilePath(), "20m", "Ready",
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

func (e *ClusterE2ETest) printPackageSpec(ctx context.Context, params []string) {
	bytes, _ := e.KubectlClient.Execute(ctx, params...)
	response := &packagesv1.Package{}
	_ = json.Unmarshal(bytes.Bytes(), response)
	formatted, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(formatted))
}

func (e *ClusterE2ETest) printDeploymentSpec(ctx context.Context, ns string) {
	response, _ := e.KubectlClient.GetDeployments(ctx,
		executables.WithKubeconfig(e.managementKubeconfigFilePath()),
		executables.WithNamespace(ns),
	)
	formatted, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(formatted))
}

// VerifyHelloPackageInstalled is checking if the hello eks anywhere package gets installed correctly.
func (e *ClusterE2ETest) VerifyHelloPackageInstalled(packageName string, mgmtCluster *types.Cluster) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)
	e.GenerateSupportBundleOnCleanupIfTestFailed()

	// Log Package/Deployment outputs
	defer func() {
		params := []string{"get", "package", packageName, "-o", "json", "-n", packageMetadatNamespace, "--kubeconfig", e.KubeconfigFilePath()}
		e.printPackageSpec(ctx, params)
		e.printDeploymentSpec(ctx, constants.EksaPackagesName)
	}()

	e.T.Log("Waiting for Package", packageName, "To be installed")
	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		mgmtCluster, packageName, "20m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for hello-eks-anywhere package timed out: %s", err)
	}

	e.T.Log("Waiting for Package", packageName, "Deployment to be healthy")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.Cluster(), "20m", "Available", "hello-eks-anywhere", constants.EksaPackagesName)
	if err != nil {
		e.T.Fatalf("waiting for hello-eks-anywhere deployment timed out: %s", err)
	}

	svcAddress := packageName + "." + constants.EksaPackagesName + ".svc.cluster.local"
	e.T.Log("Validate content at endpoint", svcAddress)
	expectedLogs := "Amazon EKS Anywhere"
	e.ValidateEndpointContent(svcAddress, constants.EksaPackagesName, expectedLogs)
}

// VerifyAdotPackageInstalled is checking if the ADOT package gets installed correctly.
func (e *ClusterE2ETest) VerifyAdotPackageInstalled(packageName, targetNamespace string) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)
	e.GenerateSupportBundleOnCleanupIfTestFailed()

	e.T.Log("Waiting for package", packageName, "to be installed")
	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		e.Cluster(), packageName, "20m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for adot package install timed out: %s", err)
	}

	// Log Package/Deployment outputs
	defer func() {
		params := []string{"get", "package", packageName, "-o", "json", "-n", packageMetadatNamespace, "--kubeconfig", e.KubeconfigFilePath()}
		e.printPackageSpec(ctx, params)
		e.printDeploymentSpec(ctx, targetNamespace)
	}()

	e.T.Log("Waiting for package", packageName, "deployment to be available")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.Cluster(), "20m", "Available", fmt.Sprintf("%s-aws-otel-collector", packageName), targetNamespace)
	if err != nil {
		e.T.Fatalf("waiting for adot deployment timed out: %s", err)
	}

	e.T.Log("Reading", packageName, "pod logs")
	adotPodName, err := e.KubectlClient.GetPodNameByLabel(context.TODO(), targetNamespace, "app.kubernetes.io/name=aws-otel-collector", e.KubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("unable to get name of the aws-otel-collector pod: %s", err)
	}
	expectedLogs := "Everything is ready"
	e.MatchLogs(targetNamespace, adotPodName, "aws-otel-collector", expectedLogs, 5*time.Minute)

	podIPAddress, err := e.KubectlClient.GetPodIP(context.TODO(), targetNamespace, adotPodName, e.KubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("unable to get ip of the aws-otel-collector pod: %s", err)
	}
	podFullIPAddress := strings.Trim(podIPAddress, `'"`) + ":8888/metrics"
	e.T.Log("Validate content at endpoint", podFullIPAddress)
	expectedLogs = "HTTP/1.1 200 OK"
	e.ValidateEndpointContent(podFullIPAddress, targetNamespace, expectedLogs, "-I")
}

//go:embed testdata/adot_package_deployment.yaml
var adotPackageDeployment []byte

//go:embed testdata/adot_package_daemonset.yaml
var adotPackageDaemonset []byte

// VerifyAdotPackageDeploymentUpdated is checking if deployment config changes trigger resource reloads correctly.
func (e *ClusterE2ETest) VerifyAdotPackageDeploymentUpdated(packageName, targetNamespace string) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)

	// Deploy ADOT as a deployment and scrape the apiservers
	e.T.Log("Apply changes to package", packageName)
	e.T.Log("This will update", packageName, "to be a deployment, and scrape the apiservers")
	err := e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, e.Cluster(), adotPackageDeployment, packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("Error upgrading adot package: %s", err)
		return
	}
	time.Sleep(30 * time.Second) // Add sleep to allow package to change state

	// Log Package/Deployment outputs
	defer func() {
		params := []string{"get", "package", packageName, "-o", "json", "-n", packageMetadatNamespace, "--kubeconfig", e.KubeconfigFilePath()}
		e.printPackageSpec(ctx, params)
		e.printDeploymentSpec(ctx, targetNamespace)
	}()

	e.T.Log("Waiting for package", packageName, "to be updated")
	err = e.KubectlClient.WaitForPackagesInstalled(ctx,
		e.Cluster(), packageName, "20m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for adot package update timed out: %s", err)
	}

	e.T.Log("Waiting for package", packageName, "deployment to be available")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.Cluster(), "20m", "Available", fmt.Sprintf("%s-aws-otel-collector", packageName), targetNamespace)
	if err != nil {
		e.T.Fatalf("waiting for adot deployment timed out: %s", err)
	}

	e.T.Log("Reading", packageName, "pod logs")
	adotPodName, err := e.KubectlClient.GetPodNameByLabel(context.TODO(), targetNamespace, "app.kubernetes.io/name=aws-otel-collector", e.KubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("unable to get name of the aws-otel-collector pod: %s", err)
	}
	logs, err := e.KubectlClient.GetPodLogs(context.TODO(), targetNamespace, adotPodName, "aws-otel-collector", e.KubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("failure getting pod logs %s", err)
	}
	fmt.Printf("Logs from aws-otel-collector pod\n %s\n", logs)
	expectedLogs := "Everything is ready"
	ok := strings.Contains(logs, expectedLogs)
	if !ok {
		e.T.Fatalf("expected to find %s in the log, got %s", expectedLogs, logs)
	}
}

// VerifyAdotPackageDaemonSetUpdated is checking if daemonset config changes trigger resource reloads correctly.
func (e *ClusterE2ETest) VerifyAdotPackageDaemonSetUpdated(packageName, targetNamespace string) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)

	// Deploy ADOT as a daemonset and scrape the node
	e.T.Log("Apply changes to package", packageName)
	e.T.Log("This will update", packageName, "to be a daemonset, and scrape the node")
	err := e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, e.Cluster(), adotPackageDaemonset, packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("Error upgrading adot package: %s", err)
		return
	}
	time.Sleep(30 * time.Second) // Add sleep to allow package to change state

	// Log Package/Deployment outputs
	defer func() {
		params := []string{"get", "package", packageName, "-o", "json", "-n", packageMetadatNamespace, "--kubeconfig", e.KubeconfigFilePath()}
		e.printPackageSpec(ctx, params)
		e.printDeploymentSpec(ctx, targetNamespace)
	}()

	e.T.Log("Waiting for package", packageName, "to be updated")
	err = e.KubectlClient.WaitForPackagesInstalled(ctx,
		e.Cluster(), packageName, "20m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for adot package update timed out: %s", err)
	}

	e.T.Log("Waiting for package", packageName, "daemonset to be rolled out")
	err = retrier.New(6 * time.Minute).Retry(func() error {
		return e.KubectlClient.WaitForResourceRolledout(ctx,
			e.Cluster(), "20m", fmt.Sprintf("%s-aws-otel-collector-agent", packageName), targetNamespace, "daemonset")
	})
	if err != nil {
		e.T.Fatalf("waiting for adot daemonset timed out: %s", err)
	}

	e.T.Log("Reading", packageName, "pod logs")
	adotPodName, err := e.KubectlClient.GetPodNameByLabel(context.TODO(), targetNamespace, "app.kubernetes.io/name=aws-otel-collector", e.KubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("unable to get name of the aws-otel-collector pod: %s", err)
	}
	expectedLogs := "Everything is ready"
	err = retrier.New(5 * time.Minute).Retry(func() error {
		logs, err := e.KubectlClient.GetPodLogs(context.TODO(), targetNamespace, adotPodName, "aws-otel-collector", e.KubeconfigFilePath())
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
func (e *ClusterE2ETest) VerifyEmissaryPackageInstalled(packageName string, mgmtCluster *types.Cluster) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)

	e.T.Log("Waiting for Package", packageName, "To be installed")
	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		mgmtCluster, packageName, "20m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for emissary package timed out: %s", err)
	}

	// Log Package/Deployment outputs
	defer func() {
		params := []string{"get", "package", packageName, "-o", "json", "-n", packageMetadatNamespace, "--kubeconfig", e.KubeconfigFilePath()}
		e.printPackageSpec(ctx, params)
		e.printDeploymentSpec(ctx, constants.EksaPackagesName)
	}()

	e.T.Log("Waiting for Package", packageName, "Deployment to be healthy")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.Cluster(), "20m", "Available", packageName, constants.EksaPackagesName)
	if err != nil {
		e.T.Fatalf("waiting for emissary deployment timed out: %s", err)
	}
	svcAddress := packageName + "-admin." + constants.EksaPackagesName + ".svc.cluster.local" + ":8877/ambassador/v0/check_alive"
	e.T.Log("Validate content at endpoint", svcAddress)
	expectedLogs := "Ambassador is alive and well"
	e.ValidateEndpointContent(svcAddress, constants.EksaPackagesName, expectedLogs)
}

// TestEmissaryPackageRouting is checking if emissary is able to create Ingress, host, and mapping that function correctly.
func (e *ClusterE2ETest) TestEmissaryPackageRouting(packageName, checkName string, mgmtCluster *types.Cluster) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)

	err := e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.Cluster(), emisarryPackage)
	if err != nil {
		e.T.Errorf("Error upgrading emissary package: %v", err)
		return
	}

	// Log Package/Deployment outputs
	defer func() {
		params := []string{"get", "package", packageName, "-o", "json", "-n", packageMetadatNamespace, "--kubeconfig", e.KubeconfigFilePath()}
		e.printPackageSpec(ctx, params)
		e.printDeploymentSpec(ctx, constants.EksaPackagesName)
	}()

	e.T.Log("Waiting for Package", packageName, "To be upgraded")
	err = e.KubectlClient.WaitForPackagesInstalled(ctx,
		mgmtCluster, packageName, "20m", fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName))
	if err != nil {
		e.T.Fatalf("waiting for emissary package upgrade timed out: %s", err)
	}
	err = e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.Cluster(), emisarryListener)
	if err != nil {
		e.T.Errorf("Error applying roles for oids: %v", err)
		return
	}

	// Functional testing of Emissary Ingress
	ingresssvcAddress := checkName + "." + constants.EksaPackagesName + ".svc.cluster.local"
	e.T.Log("Validate content at endpoint", ingresssvcAddress)
	expectedLogs := "Thank you for using"
	e.ValidateEndpointContent(ingresssvcAddress, constants.EksaPackagesName, expectedLogs)
}

// VerifyPrometheusPackageInstalled is checking if the Prometheus package gets installed correctly.
func (e *ClusterE2ETest) VerifyPrometheusPackageInstalled(packageName, targetNamespace string) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)

	e.T.Log("Waiting for package", packageName, "to be installed")
	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		e.Cluster(), packageName, "20m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for prometheus package install timed out: %s", err)
	}
}

// VerifyCertManagerPackageInstalled is checking if the cert manager package gets installed correctly.
func (e *ClusterE2ETest) VerifyCertManagerPackageInstalled(prefix, namespace, packageName string, mgmtCluster *types.Cluster) {
	ctx, cancel := context.WithCancel(context.Background())
	packageMetadatNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)
	defer cancel()

	deployments := []string{"cert-manager", "cert-manager-cainjector", "cert-manager-webhook"}

	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	okCh := make(chan string, 1)

	e.T.Log("Waiting for Package", packageName, "To be installed")

	// Log Package/Deployment outputs
	defer func() {
		params := []string{"get", "package", packageName, "-o", "json", "-n", packageMetadatNamespace, "--kubeconfig", e.KubeconfigFilePath()}
		e.printPackageSpec(ctx, params)
		e.printDeploymentSpec(ctx, namespace)
	}()

	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		mgmtCluster, prefix+"-"+packageName, "5m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for cert-manager package timed out: %s", err)
	}

	e.T.Log("Waiting for Package", packageName, "Deployment to be healthy")

	for _, name := range deployments {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			err := e.KubectlClient.WaitForDeployment(ctx,
				e.Cluster(), "20m", "Available", fmt.Sprintf("%s-%s", prefix, name), namespace)
			if err != nil {
				errCh <- err
			}
		}(name)
	}

	e.T.Log("Waiting for Self Signed certificate to be issued")
	err = e.verifySelfSignedCertificate(mgmtCluster)
	if err != nil {
		errCh <- err
	}

	e.T.Log("Waiting for Let's Encrypt certificate to be issued")
	err = e.verifyLetsEncryptCert(mgmtCluster)
	if err != nil {
		errCh <- err
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

//go:embed testdata/certmanager/certmanager_selfsignedissuer.yaml
var certManagerSelfSignedIssuer []byte

//go:embed testdata/certmanager/certmanager_selfsignedcert.yaml
var certManagerSelfSignedCert []byte

func (e *ClusterE2ETest) verifySelfSignedCertificate(mgmtCluster *types.Cluster) error {
	ctx := context.Background()
	selfsignedCert := "my-selfsigned-ca"
	err := e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.Cluster(), certManagerSelfSignedIssuer)
	if err != nil {
		return fmt.Errorf("error installing Cluster issuer for cert manager: %v", err)
	}

	err = e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.Cluster(), certManagerSelfSignedCert)
	if err != nil {
		return fmt.Errorf("error applying certificate for cert manager: %v", err)
	}

	err = e.KubectlClient.WaitJSONPathLoop(ctx, e.Cluster().KubeconfigFile, "5m", "status.conditions[?(@.type=='Ready')].status", "True",
		fmt.Sprintf("certificates.cert-manager.io/%s", selfsignedCert), constants.EksaPackagesName)
	if err != nil {
		return fmt.Errorf("failed to issue a self signed certificate: %v", err)
	}
	return nil
}

//go:embed testdata/certmanager/certmanager_letsencrypt_issuer.yaml
var certManagerLetsEncryptIssuer string

//go:embed testdata/certmanager/certmanager_letsencrypt_cert.yaml
var certManagerLetsEncryptCert []byte

//go:embed testdata/certmanager/certmanager_secret.yaml
var certManagerSecret string

func (e *ClusterE2ETest) verifyLetsEncryptCert(mgmtCluster *types.Cluster) error {
	ctx := context.Background()
	letsEncryptCert := "test-cert"
	accessKey, secretAccess, region, zoneID := GetRoute53Configs()
	data := map[string]interface{}{
		"route53SecretAccessKey": secretAccess,
	}

	certManagerSecretData, err := templater.Execute(certManagerSecret, data)
	if err != nil {
		return fmt.Errorf("failed creating cert manager secret: %v", err)
	}

	err = e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.Cluster(), certManagerSecretData)
	if err != nil {
		return fmt.Errorf("error creating cert manager secret: %v", err)
	}

	data = map[string]interface{}{
		"route53AccessKeyId": accessKey,
		"route53ZoneId":      zoneID,
		"route53Region":      region,
	}

	certManagerIssuerData, err := templater.Execute(certManagerLetsEncryptIssuer, data)
	if err != nil {
		return fmt.Errorf("failed creating lets encrypt issuer: %v", err)
	}

	err = e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.Cluster(), certManagerIssuerData)
	if err != nil {
		return fmt.Errorf("error creating cert manager let's encrypt issuer: %v", err)
	}

	err = e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.Cluster(), certManagerLetsEncryptCert)
	if err != nil {
		return fmt.Errorf("error creating cert manager let's encrypt issuer: %v", err)
	}

	err = e.KubectlClient.WaitJSONPathLoop(ctx, e.Cluster().KubeconfigFile, "5m", "status.conditions[?(@.type=='Ready')].status", "True",
		fmt.Sprintf("certificates.cert-manager.io/%s", letsEncryptCert), constants.EksaPackagesName)
	if err != nil {
		return fmt.Errorf("failed to issue a let's encrypt certificate: %v", err)
	}

	return nil
}

// CleanupCerts cleans up letsencrypt certificates.
func (e *ClusterE2ETest) CleanupCerts(mgmtCluster *types.Cluster) error {
	ctx := context.Background()
	letsEncryptCert := "test-cert"
	opts := &kubernetes.KubectlDeleteOptions{
		Name:      letsEncryptCert,
		Namespace: constants.EksaPackagesName,
	}
	err := e.KubectlClient.Delete(ctx, "certificates.cert-manager.io", e.Cluster().KubeconfigFile, opts)
	if err != nil {
		return fmt.Errorf("failed to cleanup let's encrypt certificate: %v", err)
	}

	return nil
}

// VerifyPrometheusPrometheusServerStates is checking if the Prometheus package prometheus-server component is functioning properly.
func (e *ClusterE2ETest) VerifyPrometheusPrometheusServerStates(packageName, targetNamespace, mode string) {
	ctx := context.Background()

	e.T.Log("Waiting for package", packageName, mode, "prometheus-server to be rolled out")
	err := retrier.New(6 * time.Minute).Retry(func() error {
		return e.KubectlClient.WaitForResourceRolledout(ctx,
			e.Cluster(), "5m", fmt.Sprintf("%s-server", packageName), targetNamespace, mode)
	})
	if err != nil {
		e.T.Fatalf("waiting for prometheus-server %s timed out: %s", mode, err)
	}

	e.T.Log("Reading package", packageName, "pod prometheus-server logs")
	podName, err := e.KubectlClient.GetPodNameByLabel(context.TODO(), targetNamespace, "app=prometheus,component=server", e.KubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("unable to get name of the prometheus-server pod: %s", err)
	}

	expectedLogs := "Server is ready to receive web requests"
	e.MatchLogs(targetNamespace, podName, "prometheus-server", expectedLogs, 5*time.Minute)
}

// VerifyPrometheusNodeExporterStates is checking if the Prometheus package node-exporter component is functioning properly.
func (e *ClusterE2ETest) VerifyPrometheusNodeExporterStates(packageName, targetNamespace string) {
	ctx := context.Background()

	e.T.Log("Waiting for package", packageName, "daemonset node-exporter to be rolled out")
	err := retrier.New(6 * time.Minute).Retry(func() error {
		return e.KubectlClient.WaitForResourceRolledout(ctx,
			e.Cluster(), "5m", fmt.Sprintf("%s-node-exporter", packageName), targetNamespace, "daemonset")
	})
	if err != nil {
		e.T.Fatalf("waiting for prometheus daemonset timed out: %s", err)
	}

	svcAddress := packageName + "-node-exporter." + targetNamespace + ".svc.cluster.local" + ":9100/metrics"
	e.T.Log("Validate content at endpoint", svcAddress)
	expectedLogs := "HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles"
	e.ValidateEndpointContent(svcAddress, targetNamespace, expectedLogs)
}

//go:embed testdata/prometheus_package_deployment.yaml
var prometheusPackageDeployment []byte

//go:embed testdata/prometheus_package_statefulset.yaml
var prometheusPackageStatefulSet []byte

// ApplyPrometheusPackageServerDeploymentFile is checking if deployment config changes trigger resource reloads correctly.
func (e *ClusterE2ETest) ApplyPrometheusPackageServerDeploymentFile(packageName, targetNamespace string) {
	e.T.Log("Update", packageName, "to be a deployment, and scrape the api-servers")
	e.ApplyPackageFile(packageName, targetNamespace, prometheusPackageDeployment)
}

// ApplyPrometheusPackageServerStatefulSetFile is checking if statefulset config changes trigger resource reloads correctly.
func (e *ClusterE2ETest) ApplyPrometheusPackageServerStatefulSetFile(packageName, targetNamespace string) {
	e.T.Log("Update", packageName, "to be a statefulset, and scrape the api-servers")
	e.ApplyPackageFile(packageName, targetNamespace, prometheusPackageStatefulSet)
}

// VerifyPackageControllerNotInstalled is verifying that package controller is not installed.
func (e *ClusterE2ETest) VerifyPackageControllerNotInstalled() {
	ctx := context.Background()

	packageDeployment := "eks-anywhere-packages"
	_, err := e.KubectlClient.GetDeployment(ctx, packageDeployment, constants.EksaPackagesName, e.Cluster().KubeconfigFile)

	if !apierrors.IsNotFound(err) {
		e.T.Fatalf("found deployment for package controller in workload cluster %s : %s", e.ClusterName, err)
	}
}

// VerifyAutoScalerPackageInstalled is verifying that the autoscaler package is installed and deployed.
func (e *ClusterE2ETest) VerifyAutoScalerPackageInstalled(packageName, targetNamespace string, mgmtCluster *types.Cluster) {
	ctx := context.Background()
	deploymentName := "cluster-autoscaler-clusterapi-cluster-autoscaler"
	packageMetadatNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)

	e.T.Log("Waiting for Package", packageName, "To be installed")

	// Log Package/Deployment outputs
	defer func() {
		params := []string{"get", "package", packageName, "-o", "json", "-n", packageMetadatNamespace, "--kubeconfig", e.KubeconfigFilePath()}
		e.printPackageSpec(ctx, params)
		e.printDeploymentSpec(ctx, targetNamespace)
	}()

	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		mgmtCluster, packageName, "20m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for Autoscaler Package to be avaliable")
	}

	e.T.Log("Waiting for Package", packageName, "Deployment to be healthy")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.Cluster(), "20m", "Available", deploymentName, targetNamespace)
	if err != nil {
		e.T.Fatalf("waiting for cluster-autoscaler deployment timed out: %s", err)
	}
}

// VerifyMetricServerPackageInstalled is verifying that metrics-server is installed and deployed.
func (e *ClusterE2ETest) VerifyMetricServerPackageInstalled(packageName, targetNamespace string, mgmtCluster *types.Cluster) {
	ctx := context.Background()
	deploymentName := "metrics-server"
	packageMetadatNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)

	e.T.Log("Waiting for Package", packageName, "To be installed")

	// Log Package/Deployment outputs
	defer func() {
		params := []string{"get", "package", packageName, "-o", "json", "-n", packageMetadatNamespace, "--kubeconfig", e.KubeconfigFilePath()}
		e.printPackageSpec(ctx, params)
		e.printDeploymentSpec(ctx, targetNamespace)
	}()

	err := e.KubectlClient.WaitForPackagesInstalled(ctx,
		mgmtCluster, packageName, "20m", packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("waiting for Metric Server Package to be avaliable")
	}

	e.T.Log("Waiting for Package", packageName, "Deployment to be healthy")
	err = e.KubectlClient.WaitForDeployment(ctx,
		e.Cluster(), "20m", "Available", deploymentName, targetNamespace)
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
	packageMetadatNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)
	data := map[string]interface{}{
		"targetNamespace": targetNamespace,
		"clusterName":     e.Cluster().Name,
	}

	metricsServerPackageDeployment, err := templater.Execute(metricsServerPackageDeploymentTemplate, data)
	if err != nil {
		e.T.Fatalf("Failed creating metrics-erver Package Deployment: %s", err)
	}

	err = e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, e.Cluster(), metricsServerPackageDeployment,
		packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("Error installing metrics-sserver pacakge: %s", err)
	}

	autoscalerPackageDeployment, err := templater.Execute(autoscalerPackageDeploymentTemplate, data)
	if err != nil {
		e.T.Fatalf("Failed creating autoscaler Package Deployment: %s", err)
	}

	err = e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, e.Cluster(), autoscalerPackageDeployment,
		packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("Error installing cluster autoscaler pacakge: %s", err)
	}
}

// CombinedAutoScalerMetricServerTest verifies that new nodes are spun up after using a HPA to scale a deployment.
func (e *ClusterE2ETest) CombinedAutoScalerMetricServerTest(autoscalerName, metricServerName, targetNamespace string, mgmtCluster *types.Cluster) {
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
		e.Cluster(), "5m", "Available", name, ns)
	if err != nil {
		e.T.Fatalf("Failed waiting for test workload deployent %s", err)
	}

	params := []string{"autoscale", "deployment", name, "--cpu-percent=50", "--min=1", "--max=20", "--kubeconfig", e.KubeconfigFilePath()}
	_, err = e.KubectlClient.ExecuteCommand(ctx, params...)
	if err != nil {
		e.T.Fatalf("Failed to autoscale deployent: %s", err)
	}

	e.T.Log("Waiting for machinedeployment to begin scaling up")
	err = e.KubectlClient.WaitJSONPathLoop(ctx, mgmtCluster.KubeconfigFile, "20m", "status.phase", "ScalingUp",
		fmt.Sprintf("machinedeployments.cluster.x-k8s.io/%s", machineDeploymentName), constants.EksaSystemNamespace)
	if err != nil {
		e.T.Fatalf("Failed to get ScalingUp phase for machinedeployment: %s", err)
	}

	e.T.Log("Waiting for machinedeployment to finish scaling up")
	err = e.KubectlClient.WaitJSONPathLoop(ctx, mgmtCluster.KubeconfigFile, "15m", "status.phase", "Running",
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

// ValidateClusterState runs a set of validations against the cluster to identify an invalid cluster state.
func (e *ClusterE2ETest) ValidateClusterState() {
	validateClusterState(e.T.(*testing.T), e)
}

// ValidateClusterStateWithT runs a set of validations against the cluster to identify an invalid cluster state and accepts *testing.T as a parameter.
func (e *ClusterE2ETest) ValidateClusterStateWithT(t *testing.T) {
	validateClusterState(t, e)
}

func validateClusterState(t *testing.T, e *ClusterE2ETest) {
	t.Logf("Validating cluster %s", e.ClusterName)
	ctx := context.Background()
	e.buildClusterStateValidationConfig(ctx)
	clusterStateValidator := newClusterStateValidator(e.clusterStateValidationConfig)
	clusterStateValidator.WithValidations(validationsForExpectedObjects()...)
	clusterStateValidator.WithValidations(e.Provider.ClusterStateValidations()...)
	if err := clusterStateValidator.Validate(ctx); err != nil {
		e.T.Fatalf("failed to validate cluster %v", err)
	}
}

// ApplyPackageFile is applying a package file in the cluster.
func (e *ClusterE2ETest) ApplyPackageFile(packageName, targetNamespace string, PackageFile []byte) {
	ctx := context.Background()
	packageMetadatNamespace := fmt.Sprintf("%s-%s", "eksa-packages", e.ClusterName)

	e.T.Log("Apply changes to package", packageName)
	err := e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, e.Cluster(), PackageFile, packageMetadatNamespace)
	if err != nil {
		e.T.Fatalf("Error upgrading package: %s", err)
		return
	}
	time.Sleep(30 * time.Second) // Add sleep to allow package to change state
}

// CurlEndpoint creates a pod with command to curl the target endpoint,
// and returns the created pod name.
func (e *ClusterE2ETest) CurlEndpoint(endpoint, namespace string, extraCurlArgs ...string) string {
	ctx := context.Background()

	e.T.Log("Launching pod to curl endpoint", endpoint)
	randomname := fmt.Sprintf("%s-%s", "curl-test", utilrand.String(7))
	curlPodName, err := e.KubectlClient.RunCurlPod(context.TODO(),
		namespace, randomname, e.KubeconfigFilePath(), append([]string{"curl", endpoint}, extraCurlArgs...))
	if err != nil {
		e.T.Fatalf("error launching pod: %s", err)
	}

	err = e.KubectlClient.WaitForPodCompleted(ctx,
		e.Cluster(), curlPodName, "5m", namespace)
	if err != nil {
		e.T.Fatalf("waiting for pod %s timed out: %s", curlPodName, err)
	}

	return curlPodName
}

// MatchLogs matches the log from a container to the expected content. Given it
// takes time for logs to be populated, a retrier with configurable timeout duration
// is added.
func (e *ClusterE2ETest) MatchLogs(targetNamespace, targetPodName string,
	targetContainerName, expectedLogs string, timeout time.Duration,
) {
	e.T.Logf("Match logs for pod %s, container %s in namespace %s", targetPodName,
		targetContainerName, targetNamespace)
	e.GenerateSupportBundleOnCleanupIfTestFailed()

	err := retrier.New(timeout).Retry(func() error {
		logs, err := e.KubectlClient.GetPodLogs(context.TODO(), targetNamespace,
			targetPodName, targetContainerName, e.KubeconfigFilePath())
		if err != nil {
			return fmt.Errorf("failure getting pod logs %s", err)
		}
		fmt.Printf("Logs from pod\n %s\n", logs)
		ok := strings.Contains(logs, expectedLogs)
		if !ok {
			return fmt.Errorf("expected to find %s in the log, got %s", expectedLogs, logs)
		}
		return nil
	})
	if err != nil {
		e.T.Fatalf("unable to match logs: %s", err)
	}
}

// ValidateEndpointContent validates the contents at the target endpoint.
func (e *ClusterE2ETest) ValidateEndpointContent(endpoint, namespace, expectedContent string, extraCurlArgs ...string) {
	curlPodName := e.CurlEndpoint(endpoint, namespace, extraCurlArgs...)
	e.MatchLogs(namespace, curlPodName, curlPodName, expectedContent, 5*time.Minute)
}

// AirgapDockerContainers airgap docker containers. Outside network should not be reached during airgapped deployment.
func (e *ClusterE2ETest) AirgapDockerContainers(localCIDRs string) {
	e.T.Logf("Airgap docker containers...")
	e.Run(fmt.Sprintf("sudo iptables -F DOCKER-USER && sudo iptables -I DOCKER-USER -j DROP && sudo iptables -I DOCKER-USER -s %s,172.0.0.0/8,127.0.0.1 -j ACCEPT", localCIDRs))
}

// CreateAirgappedUser create airgapped user and setup the iptables rule. Notice that OUTPUT chain is flushed each time.
func (e *ClusterE2ETest) CreateAirgappedUser(localCIDR string) {
	e.Run("if ! id airgap; then sudo useradd airgap -G docker; fi")
	e.Run("mkdir ./eksa-cli-logs || chmod 777 ./eksa-cli-logs") // Allow the airgap user to access logs folder
	e.Run("chmod -R 777 ./")                                    // Allow the airgap user to access working dir
	e.Run("sudo iptables -F OUTPUT")
	e.Run(fmt.Sprintf("sudo iptables -A OUTPUT -d %s,172.0.0.0/8,127.0.0.1 -m owner --uid-owner airgap -j ACCEPT", localCIDR))
	e.Run("sudo iptables -A OUTPUT -m owner --uid-owner airgap -j REJECT")
}

// AssertAirgappedNetwork make sure that the admin machine is indeed airgapped.
func (e *ClusterE2ETest) AssertAirgappedNetwork() {
	cmd := exec.Command("docker", "run", "--rm", "busybox", "ping", "8.8.8.8", "-c", "1", "-W", "2")
	out, err := cmd.Output()
	e.T.Log(string(out))
	if err == nil {
		e.T.Fatalf("Docker container is not airgapped")
	}

	cmd = exec.Command("sudo", "-u", "airgap", "ping", "8.8.8.8", "-c", "1", "-W", "2")
	out, err = cmd.Output()
	e.T.Log(string(out))
	if err == nil {
		e.T.Fatalf("Airgap user is not airgapped")
	}
}

func dumpFile(description, path string, t T) {
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s:\n%s\n", description, string(b))
}

func (e *ClusterE2ETest) setFeatureFlagForUnreleasedKubernetesVersion(version v1alpha1.KubernetesVersion) {
	// Update this variable to equal the feature flagged k8s version when applicable.
	// For example, if k8s 1.26 is under a feature flag, we would set this to v1alpha1.Kube126
	unreleasedK8sVersion := v1alpha1.Kube129

	if version == unreleasedK8sVersion {
		// Set feature flag for the unreleased k8s version when applicable
		e.T.Logf("Setting k8s version support feature flag...")
		os.Setenv(features.K8s129SupportEnvVar, "true")
	}
}

// CreateCloudStackCredentialsSecretFromEnvVar parses the cloudstack credentials from an environment variable,
// builds a new secret object from the credentials in the provided profile and creates it in the cluster.
func (e *ClusterE2ETest) CreateCloudStackCredentialsSecretFromEnvVar(name, profileName string) {
	ctx := context.Background()

	execConfig, err := decoder.ParseCloudStackCredsFromEnv()
	if err != nil {
		e.T.Fatalf("error parsing cloudstack credentials from env: %v", err)
		return
	}

	var selectedProfile *decoder.CloudStackProfileConfig
	for _, p := range execConfig.Profiles {
		if profileName == p.Name {
			selectedProfile = &p
			break
		}
	}

	if selectedProfile == nil {
		e.T.Fatalf("error finding profile with the name %s", profileName)
		return
	}

	data := map[string][]byte{}
	data[decoder.APIKeyKey] = []byte(selectedProfile.ApiKey)
	data[decoder.SecretKeyKey] = []byte(selectedProfile.SecretKey)
	data[decoder.APIUrlKey] = []byte(selectedProfile.ManagementUrl)
	data[decoder.VerifySslKey] = []byte(selectedProfile.VerifySsl)

	// Create a new secret with the credentials from the profile, but with a new name.
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: data,
	}

	secretContent, err := yaml.Marshal(secret)
	if err != nil {
		e.T.Fatalf("error marshalling credentials secret : %v", err)
		return
	}

	err = e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, e.Cluster(), secretContent,
		constants.EksaSystemNamespace)
	if err != nil {
		e.T.Fatalf("error applying credentials secret to cluster %s: %v", e.Cluster().Name, err)
		return
	}
}

func (e *ClusterE2ETest) addClusterConfigFillers(fillers ...api.ClusterConfigFiller) {
	e.clusterConfigFillers = append(e.clusterConfigFillers, fillers...)
}

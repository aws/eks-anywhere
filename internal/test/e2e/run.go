package e2e

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	e2etest "github.com/aws/eks-anywhere/test/e2e"
)

const (
	testResultPass       = "pass"
	testResultFail       = "fail"
	testResultError      = "error"
	nonAirgappedHardware = "nonAirgappedHardware"
	airgappedHardware    = "AirgappedHardware"
	maxIPPoolSize        = 10
	minIPPoolSize        = 1
	tinkerbellIPPoolSize = 2

	// Default timeout for E2E test instance.
	e2eTimeout           = 150 * time.Minute
	e2eSSMTimeoutPadding = 10 * time.Minute

	// Default timeout used for all SSM commands besides running the actual E2E test.
	ssmTimeout = 10 * time.Minute
)

type ParallelRunConf struct {
	TestInstanceConfigFile string
	MaxInstances           int
	MaxConcurrentTests     int
	InstanceProfileName    string
	StorageBucket          string
	JobId                  string
	Regex                  string
	TestsToSkip            []string
	BundlesOverride        bool
	CleanupResources       bool
	TestReportFolder       string
	BranchName             string
	BaremetalBranchName    string
	Logger                 logr.Logger

	infraConfig *TestInfraConfig
	session     *session.Session
}

// isBaremetal checks is the test run is for Baremetal tests.
func (c ParallelRunConf) isBaremetal() bool {
	return strings.EqualFold(c.BranchName, c.BaremetalBranchName)
}

// init initializes ParallelRunConf. It needs to be called before the config is used.
// This is not thread safe.
func (c *ParallelRunConf) init() error {
	infraConfig, err := NewTestRunnerConfigFromFile(c.Logger, c.TestInstanceConfigFile)
	if err != nil {
		return fmt.Errorf("creating test runner config for tests: %v", err)
	}
	c.infraConfig = infraConfig

	awsSession, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("creating aws session for tests: %v", err)
	}
	c.session = awsSession

	return nil
}

type (
	testCommandResult    = ssm.RunOutput
	instanceTestsResults struct {
		conf              *instanceRunConf
		testCommandResult *testCommandResult
		err               error
	}
)

// RunTestsInParallel Run Tests in parallel by spawning multiple admin machines.
func RunTestsInParallel(conf ParallelRunConf) error {
	if err := conf.init(); err != nil {
		return err
	}

	testsList, skippedTests, err := listTests(conf.Regex, conf.TestsToSkip)
	if err != nil {
		return err
	}

	conf.Logger.Info("Running tests", "selected", testsList, "skipped", skippedTests)

	if conf.TestReportFolder != "" {
		if err = os.MkdirAll(conf.TestReportFolder, os.ModePerm); err != nil {
			return err
		}
	}

	var wg sync.WaitGroup

	instancesConf, err := splitTests(testsList, conf)
	if err != nil {
		return fmt.Errorf("failed to split tests: %v", err)
	}

	logTestGroups(conf.Logger, instancesConf)

	work := make(chan *instanceRunConf)
	results := make(chan instanceTestsResults)
	go func() {
		for _, instanceConf := range instancesConf {
			work <- instanceConf
		}
		close(work)
	}()

	numWorkers := conf.MaxConcurrentTests
	if numWorkers > len(instancesConf) {
		numWorkers = len(instancesConf)
	}
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for c := range work {
				r := instanceTestsResults{conf: c}

				r.conf.InstanceID, r.testCommandResult, err = RunTests(c)
				if err != nil {
					r.err = err
				}

				results <- r
			}
		}()
		wg.Add(1)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	failedInstances := 0
	totalInstances := len(instancesConf)
	completedInstances := 0
	for r := range results {
		var result string
		// This variable can be used in cloudwatch log insights query for e2e test success rate
		succeeded := 0
		// TODO: keeping the old logs temporarily for compatibility with the test tool
		// Once the tool is updated to support the unified message, remove them
		if r.err != nil {
			result = testResultError
			conf.Logger.Error(r.err, "Failed running e2e tests for instance", "jobId", r.conf.JobID, "instanceId", r.conf.InstanceID, "tests", r.conf.Regex, "status", testResultFail)
			failedInstances++
		} else if !r.testCommandResult.Successful() {
			result = testResultFail
			conf.Logger.Info("An e2e instance run has failed", "jobId", r.conf.JobID, "instanceId", r.conf.InstanceID, "commandId", r.testCommandResult.CommandId, "tests", r.conf.Regex, "status", testResultFail)
			failedInstances++
		} else {
			result = testResultPass
			succeeded = 1
			conf.Logger.Info("Instance tests completed successfully", "jobId", r.conf.JobID, "instanceId", r.conf.InstanceID, "commandId", r.testCommandResult.CommandId, "tests", r.conf.Regex, "status", testResultPass)
		}
		completedInstances++
		conf.Logger.Info("Instance tests run finished",
			"result", result,
			"tests", r.conf.Regex,
			"jobId", r.conf.JobID,
			"instanceId", r.conf.InstanceID,
			"completedInstances", completedInstances,
			"totalInstances", totalInstances,
			"succeeded", succeeded,
			"ssmStatusDetails", r.testCommandResult.StatusDetails(),
		)
		putInstanceTestResultMetrics(r)
	}

	if failedInstances > 0 {
		return fmt.Errorf("%d/%d e2e instances failed", failedInstances, totalInstances)
	}

	return nil
}

type instanceRunConf struct {
	InstanceProfileName     string
	StorageBucket           string
	JobID                   string
	ParentJobID             string
	Regex                   string
	InstanceID              string
	TestReportFolder        string
	BranchName              string
	IPPool                  networkutils.IPPool
	Hardware                []*api.Hardware
	HardwareCount           int
	TinkerbellAirgappedTest bool
	BundlesOverride         bool
	TestRunnerType          TestRunnerType
	TestRunnerConfig        TestInfraConfig
	CleanupResources        bool
	Logger                  logr.Logger
	Session                 *session.Session

	// hardwareCatalogue holds the hardware inventory for Tinkerbell tests.
	// Before running the tests for this instance, the hardware needs to be
	// reserved through this inventory and set in the Hardware field.
	// After the tests are run, the hardware needs to be released.
	hardwareCatalogue *hardwareCatalogue
}

// reserveHardware allocates the required hardware for the test run.
// If hardware is not available, it will block until it is available.
func (c *instanceRunConf) reserveHardware() error {
	if c.HardwareCount == 0 {
		return nil
	}

	reservedTinkerbellHardware, err := c.hardwareCatalogue.reserveHardware(c.HardwareCount)
	if err != nil {
		return fmt.Errorf("timed out waiting for hardware")
	}
	c.Hardware = reservedTinkerbellHardware
	logTinkerbellTestHardwareInfo(c, "Reserved")

	return nil
}

// releaseHardware de-allocates the hardware reserved for this test run.
// This should be called after the test run is complete.
func (c *instanceRunConf) releaseHardware() {
	if c.HardwareCount == 0 {
		return
	}
	logTinkerbellTestHardwareInfo(c, "Releasing")
	c.hardwareCatalogue.releaseHardware(c.Hardware)
}

//nolint:gocyclo, revive // RunTests responsible launching test runner to run tests is complex.
func RunTests(conf *instanceRunConf) (testInstanceID string, testCommandResult *testCommandResult, err error) {
	testRunner, err := newTestRunner(conf.TestRunnerType, conf.TestRunnerConfig)
	if err != nil {
		return "", nil, err
	}

	if err = conf.reserveHardware(); err != nil {
		return "", nil, err
	}
	defer conf.releaseHardware()

	conf.Logger.Info("Creating runner instance",
		"instance_profile_name", conf.InstanceProfileName, "storage_bucket", conf.StorageBucket,
		"parent_job_id", conf.ParentJobID, "regex", conf.Regex, "test_report_folder", conf.TestReportFolder,
		"branch_name", conf.BranchName, "ip_pool", conf.IPPool.ToString(),
		"hardware_count", conf.HardwareCount, "tinkerbell_airgapped_test", conf.TinkerbellAirgappedTest,
		"bundles_override", conf.BundlesOverride, "test_runner_type", conf.TestRunnerType,
		"cleanup_resources", conf.CleanupResources)

	instanceId, err := testRunner.createInstance(conf)
	if err != nil {
		return "", nil, err
	}
	conf.Logger = conf.Logger.WithValues("instance_id", instanceId)
	conf.Logger.Info("TestRunner instance has been created")

	defer func() {
		err := testRunner.decommInstance(conf)
		if err != nil {
			conf.Logger.V(1).Info("WARN: Failed to decomm e2e test runner instance", "error", err)
		}
	}()

	session, err := newE2ESession(instanceId, conf)
	if err != nil {
		return "", nil, err
	}

	err = session.setup(conf.Regex)
	if err != nil {
		return session.instanceId, nil, err
	}

	testCommandResult, err = session.runTests(conf.Regex)
	if err != nil {
		return session.instanceId, nil, err
	}

	if err = conf.runPostTestsProcessing(session, testCommandResult); err != nil {
		return session.instanceId, nil, err
	}

	// Tagging only successful e2e test instances.
	// The aws cleanup periodic job deletes the tagged EC2 instances and long lived instances.
	if testCommandResult.Successful() {
		key := "Integration-Test-Done"
		value := "TRUE"
		err = testRunner.tagInstance(conf, key, value)
		if err != nil {
			return session.instanceId, nil, fmt.Errorf("tagging instance for e2e success: %v", err)
		}
	}

	return session.instanceId, testCommandResult, nil
}

func (e *E2ESession) runTests(regex string) (testCommandResult *testCommandResult, err error) {
	e.logger.V(1).Info("Running e2e tests", "regex", regex)
	command := "GOVERSION=go1.16.6 gotestsum --junitfile=junit-testing.xml --raw-command --format=standard-verbose --hide-summary=all --ignore-non-json-output-lines -- test2json -t -p e2e ./bin/e2e.test -test.v"

	if regex != "" {
		command = fmt.Sprintf("%s -test.run \"^(%s)$\" -test.timeout %s", command, regex, e2eTimeout)
	}

	command = e.commandWithEnvVars(command)

	opt := ssm.WithOutputToCloudwatch()

	testCommandResult, err = ssm.RunCommand(
		e.session,
		e.logger.V(4),
		e.instanceId,
		command,
		e2eTimeout+e2eSSMTimeoutPadding,
		opt,
	)
	if err != nil {
		return nil, fmt.Errorf("running e2e tests on instance %s: %v", e.instanceId, err)
	}

	return testCommandResult, nil
}

func (c instanceRunConf) runPostTestsProcessing(e *E2ESession, testCommandResult *testCommandResult) error {
	regex := strings.Trim(c.Regex, "\"")
	tests := strings.Split(regex, "|")

	for _, testName := range tests {
		e.uploadJUnitReportFromInstance(testName)
		if c.TestReportFolder != "" {
			e.downloadJUnitReportToLocalDisk(testName, c.TestReportFolder)
		}

		if !testCommandResult.Successful() {
			// For Tinkerbell tests we run multiple tests on the same instance.
			// Hence upload fails for passed tests within the instance.
			// TODO (pokearu): Find a way to only upload for failed tests within the instance.
			e.uploadGeneratedFilesFromInstance(testName)
			e.uploadDiagnosticArchiveFromInstance(testName)
		}
	}

	return nil
}

func (e *E2ESession) commandWithEnvVars(command string) string {
	fullCommand := make([]string, 0, len(e.testEnvVars)+1)

	for k, v := range e.testEnvVars {
		fullCommand = append(fullCommand, fmt.Sprintf("export %s=\"%s\"", k, v))
	}
	fullCommand = append(fullCommand, command)

	return strings.Join(fullCommand, "; ")
}

func splitTests(testsList []string, conf ParallelRunConf) ([]*instanceRunConf, error) {
	testPerInstance := len(testsList) / conf.MaxInstances
	if testPerInstance == 0 {
		testPerInstance = 1
	}

	vsphereTestsRe := regexp.MustCompile(vsphereRegex)
	tinkerbellTestsRe := regexp.MustCompile(tinkerbellTestsRe)
	cloudstackTestRe := regexp.MustCompile(cloudstackRegex)
	nutanixTestsRe := regexp.MustCompile(nutanixRegex)
	privateNetworkTestsRe := regexp.MustCompile(`^.*(Proxy|RegistryMirror).*$`)
	multiClusterTestsRe := regexp.MustCompile(`^.*Multicluster.*$`)

	runConfs := make([]*instanceRunConf, 0, conf.MaxInstances)
	vsphereIPMan := newE2EIPManager(conf.Logger, os.Getenv(vsphereCidrVar))
	vspherePrivateIPMan := newE2EIPManager(conf.Logger, os.Getenv(vspherePrivateNetworkCidrVar))
	cloudstackIPMan := newE2EIPManager(conf.Logger, os.Getenv(cloudstackCidrVar))
	nutanixIPMan := newE2EIPManager(conf.Logger, os.Getenv(nutanixCidrVar))

	testsInEC2Instance := make([]string, 0, testPerInstance)
	for i, testName := range testsList {
		if tinkerbellTestsRe.MatchString(testName) {
			continue
		}

		testsInEC2Instance = append(testsInEC2Instance, testName)
		multiClusterTest := multiClusterTestsRe.MatchString(testName)

		var ips networkutils.IPPool
		if vsphereTestsRe.MatchString(testName) {
			if privateNetworkTestsRe.MatchString(testName) {
				if multiClusterTest {
					ips = vspherePrivateIPMan.reserveIPPool(maxIPPoolSize)
				} else {
					ips = vspherePrivateIPMan.reserveIPPool(minIPPoolSize)
				}
			} else {
				if multiClusterTest {
					ips = vsphereIPMan.reserveIPPool(maxIPPoolSize)
				} else {
					ips = vsphereIPMan.reserveIPPool(minIPPoolSize)
				}
			}
		} else if nutanixTestsRe.MatchString(testName) {
			ips = nutanixIPMan.reserveIPPool(minIPPoolSize)
		} else if cloudstackTestRe.MatchString(testName) {
			if multiClusterTest {
				ips = cloudstackIPMan.reserveIPPool(maxIPPoolSize)
			} else {
				ips = cloudstackIPMan.reserveIPPool(minIPPoolSize)
			}
		}

		if len(testsInEC2Instance) == testPerInstance || (len(testsList)-1) == i {
			runConfs = append(runConfs, newInstanceRunConf(conf.session, conf, len(runConfs), strings.Join(testsInEC2Instance, "|"), ips, []*api.Hardware{}, 0, false, Ec2TestRunnerType, conf.infraConfig))
			testsInEC2Instance = make([]string, 0, testPerInstance)
		}
	}

	if conf.isBaremetal() {
		confs, err := testConfigurationsForTinkerbell(testsList, conf)
		if err != nil {
			return nil, err
		}
		runConfs = append(runConfs, confs...)
	}

	return runConfs, nil
}

func testConfigurationsForTinkerbell(testsList []string, conf ParallelRunConf) ([]*instanceRunConf, error) {
	runConfs := []*instanceRunConf{}

	tinkerbellIPManager := newE2EIPManager(conf.Logger, os.Getenv(tinkerbellControlPlaneNetworkCidrEnvVar))
	confs, err := nonAirgappedTinkerbellRunConfs(testsList, conf, tinkerbellIPManager)
	if err != nil {
		return nil, fmt.Errorf("failed to split Tinkerbell tests: %v", err)
	}
	runConfs = append(runConfs, confs...)

	confs, err = airgappedTinkerbellRunConfs(testsList, conf, tinkerbellIPManager)
	if err != nil {
		return nil, fmt.Errorf("failed to run airgapped Tinkerbell tests: %v", err)
	}
	runConfs = append(runConfs, confs...)

	return runConfs, nil
}

//nolint:gocyclo // This legacy function is complex but the team too busy to simplify it
func nonAirgappedTinkerbellRunConfs(testsList []string, conf ParallelRunConf, ipManager *E2EIPManager) ([]*instanceRunConf, error) {
	nonAirgappedTinkerbellTests := getTinkerbellNonAirgappedTests(testsList)
	conf.Logger.V(1).Info("INFO:", "tinkerbellTests", len(nonAirgappedTinkerbellTests))

	testPerInstance := len(nonAirgappedTinkerbellTests) / conf.MaxInstances
	if testPerInstance == 0 {
		testPerInstance = 1
	}
	testsInVSphereInstance := make([]string, 0, testPerInstance)
	nonAirgappedTinkerbellTestsWithCount, err := getTinkerbellTestsWithCount(nonAirgappedTinkerbellTests, conf)
	if err != nil {
		return nil, err
	}
	hardware, err := nonAirgappedHardwarePool(conf.session, conf.StorageBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to get non-airgapped hardware inventory for Tinkerbell Tests: %v", err)
	}
	catalogue := newHardwareCatalogue(hardware)

	runConfs := make([]*instanceRunConf, 0, len(nonAirgappedTinkerbellTestsWithCount))
	for i, test := range nonAirgappedTinkerbellTestsWithCount {
		testsInVSphereInstance = append(testsInVSphereInstance, test.Name)
		ipPool := ipManager.reserveIPPool(tinkerbellIPPoolSize)
		if len(testsInVSphereInstance) == testPerInstance || (len(testsList)-1) == i {
			c := newInstanceRunConf(conf.session, conf, len(runConfs), strings.Join(testsInVSphereInstance, "|"), ipPool, []*api.Hardware{}, test.Count, false, VSphereTestRunnerType, conf.infraConfig)
			c.hardwareCatalogue = catalogue
			runConfs = append(runConfs, c)
			testsInVSphereInstance = make([]string, 0, testPerInstance)
		}
	}

	return runConfs, nil
}

func airgappedTinkerbellRunConfs(testsList []string, conf ParallelRunConf, ipManager *E2EIPManager) ([]*instanceRunConf, error) {
	airgappedTinkerbellTests := getTinkerbellAirgappedTests(testsList)
	if len(airgappedTinkerbellTests) == 0 {
		conf.Logger.V(1).Info("No tinkerbell airgapped test to run")
		return nil, nil
	}
	conf.Logger.V(1).Info("INFO:", "tinkerbellAirGappedTests", len(airgappedTinkerbellTests))
	testPerInstance := len(airgappedTinkerbellTests) / conf.MaxInstances
	if testPerInstance == 0 {
		testPerInstance = 1
	}
	testsInVSphereInstance := make([]string, 0, testPerInstance)
	airgappedTinkerbellTestsWithCount, err := getTinkerbellTestsWithCount(airgappedTinkerbellTests, conf)
	if err != nil {
		return nil, err
	}

	hardware, err := airgappedHardwarePool(conf.session, conf.StorageBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to get airgapped hardware inventory for Tinkerbell Tests: %v", err)
	}
	catalogue := newHardwareCatalogue(hardware)

	runConfs := make([]*instanceRunConf, 0, len(airgappedTinkerbellTestsWithCount))
	for i, test := range airgappedTinkerbellTestsWithCount {
		testsInVSphereInstance = append(testsInVSphereInstance, test.Name)
		ipPool := ipManager.reserveIPPool(tinkerbellIPPoolSize)
		if len(testsInVSphereInstance) == testPerInstance || (len(testsList)-1) == i {
			c := newInstanceRunConf(conf.session, conf, len(runConfs), strings.Join(testsInVSphereInstance, "|"), ipPool, []*api.Hardware{}, test.Count, true, VSphereTestRunnerType, conf.infraConfig)
			c.hardwareCatalogue = catalogue
			runConfs = append(runConfs, c)
			testsInVSphereInstance = make([]string, 0, testPerInstance)
		}
	}

	return runConfs, nil
}

func getTinkerbellTestsWithCount(tinkerbellTests []string, conf ParallelRunConf) ([]TinkerbellTest, error) {
	tinkerbellTestsHwRequirements, err := e2etest.GetTinkerbellTestsHardwareRequirements()
	if err != nil {
		return nil, fmt.Errorf("failed to get tinkerbell tests hardware count: %v", err)
	}

	tinkerbellTestsWithCount := make([]TinkerbellTest, 0, len(tinkerbellTests))

	for _, testName := range tinkerbellTests {
		hwCount, ok := tinkerbellTestsHwRequirements[testName]
		if !ok {
			conf.Logger.V(1).Info(fmt.Sprintf("WARN: Test not found in %s", e2etest.TinkerbellHardwareCountFile), "test", testName)
		} else {
			tinkerbellTestsWithCount = append(tinkerbellTestsWithCount, TinkerbellTest{Name: testName, Count: hwCount})
		}
	}
	// sort tests by Hardware count, to enable running larger tests first for Tinkerbell Provider
	sort.Slice(tinkerbellTestsWithCount, func(i, j int) bool {
		return tinkerbellTestsWithCount[i].Count > tinkerbellTestsWithCount[j].Count
	})

	return tinkerbellTestsWithCount, nil
}

func newInstanceRunConf(awsSession *session.Session, conf ParallelRunConf, jobNumber int, testRegex string, ipPool networkutils.IPPool, hardware []*api.Hardware, hardwareCount int, tinkerbellAirgappedTest bool, testRunnerType TestRunnerType, testRunnerConfig *TestInfraConfig) *instanceRunConf {
	jobID := fmt.Sprintf("%s-%d", conf.JobId, jobNumber)
	return &instanceRunConf{
		Session:                 awsSession,
		InstanceProfileName:     conf.InstanceProfileName,
		StorageBucket:           conf.StorageBucket,
		JobID:                   jobID,
		ParentJobID:             conf.JobId,
		Regex:                   testRegex,
		IPPool:                  ipPool,
		Hardware:                hardware,
		HardwareCount:           hardwareCount,
		TinkerbellAirgappedTest: tinkerbellAirgappedTest,
		BundlesOverride:         conf.BundlesOverride,
		TestReportFolder:        conf.TestReportFolder,
		BranchName:              conf.BranchName,
		CleanupResources:        conf.CleanupResources,
		TestRunnerType:          testRunnerType,
		TestRunnerConfig:        *testRunnerConfig,
		Logger:                  conf.Logger.WithValues("jobID", jobID, "test", testRegex),
	}
}

func logTestGroups(logger logr.Logger, instancesConf []*instanceRunConf) {
	testGroups := make([]string, 0, len(instancesConf))
	for _, i := range instancesConf {
		testGroups = append(testGroups, i.Regex)
	}
	logger.V(1).Info("Running tests in parallel", "testsGroups", testGroups)
}

func logTinkerbellTestHardwareInfo(conf *instanceRunConf, action string) {
	var hardwareInfo []string
	for _, hardware := range conf.Hardware {
		hardwareInfo = append(hardwareInfo, hardware.Hostname)
	}
	conf.Logger.V(1).Info(action+" hardware for TestRunner", "hardwarePool", strings.Join(hardwareInfo, ", "))
}

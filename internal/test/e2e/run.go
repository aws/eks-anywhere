package e2e

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

const (
	testResultPass  = "pass"
	testResultFail  = "fail"
	testResultError = "error"

	maxIPPoolSize = 10
	minIPPoolSize = 1

	// Default timeout for E2E test instance.
	e2eTimeout           = 300 * time.Minute
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
	CleanupVms             bool
	TestReportFolder       string
	BranchName             string
	BaremetalBranchName    string
	Logger                 logr.Logger
}

type (
	testCommandResult    = ssm.RunOutput
	instanceTestsResults struct {
		conf              instanceRunConf
		testCommandResult *testCommandResult
		err               error
	}
)

// RunTestsInParallel Run Tests in parallel by spawning multiple admin machines.
func RunTestsInParallel(conf ParallelRunConf) error {
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

	results := make([]instanceTestsResults, 0, len(instancesConf))
	logTestGroups(conf.Logger, instancesConf)
	maxConcurrentTests := conf.MaxConcurrentTests
	// Add a blocking channel to only allow for certain number of tests to run at a time
	queue := make(chan struct{}, maxConcurrentTests)
	for _, instanceConf := range instancesConf {
		queue <- struct{}{}
		go func(c instanceRunConf) {
			defer wg.Done()
			r := instanceTestsResults{conf: c}

			r.conf.instanceId, r.testCommandResult, err = RunTests(c)
			if err != nil {
				r.err = err
			}

			results = append(results, r)
			<-queue
		}(instanceConf)
		wg.Add(1)
	}

	wg.Wait()
	close(queue)

	failedInstances := 0
	totalInstances := len(instancesConf)
	completedInstances := 0
	for _, r := range results {
		var result string
		// TODO: keeping the old logs temporarily for compatibility with the test tool
		// Once the tool is updated to support the unified message, remove them
		if r.err != nil {
			result = testResultError
			conf.Logger.Error(r.err, "Failed running e2e tests for instance", "jobId", r.conf.jobId, "instanceId", r.conf.instanceId, "tests", r.conf.regex, "status", testResultFail)
			failedInstances += 1
		} else if !r.testCommandResult.Successful() {
			result = testResultFail
			conf.Logger.Info("An e2e instance run has failed", "jobId", r.conf.jobId, "instanceId", r.conf.instanceId, "commandId", r.testCommandResult.CommandId, "tests", r.conf.regex, "status", testResultFail)
			failedInstances += 1
		} else {
			result = testResultPass
			conf.Logger.Info("Instance tests completed successfully", "jobId", r.conf.jobId, "instanceId", r.conf.instanceId, "commandId", r.testCommandResult.CommandId, "tests", r.conf.regex, "status", testResultPass)
		}
		completedInstances += 1
		conf.Logger.Info("Instance tests run finished",
			"result", result,
			"tests", r.conf.regex,
			"jobId", r.conf.jobId,
			"instanceId", r.conf.instanceId,
			"completedInstances", completedInstances,
			"totalInstances", totalInstances,
		)
	}

	if failedInstances > 0 {
		return fmt.Errorf("%d/%d e2e instances failed", failedInstances, totalInstances)
	}

	return nil
}

type instanceRunConf struct {
	session                                                                   *session.Session
	instanceProfileName, storageBucket, jobId, parentJobId, regex, instanceId string
	testReportFolder, branchName                                              string
	ipPool                                                                    networkutils.IPPool
	hardware                                                                  []*api.Hardware
	bundlesOverride                                                           bool
	testRunnerType                                                            TestRunnerType
	testRunnerConfig                                                          TestInfraConfig
	cleanupVms                                                                bool
	logger                                                                    logr.Logger
}

func RunTests(conf instanceRunConf) (testInstanceID string, testCommandResult *testCommandResult, err error) {
	testRunner, err := newTestRunner(conf.testRunnerType, conf.testRunnerConfig)
	if err != nil {
		return "", nil, err
	}

	instanceId, err := testRunner.createInstance(conf)
	if err != nil {
		return "", nil, err
	}
	conf.logger.V(1).Info("TestRunner instance has been created", "instanceId", instanceId)

	defer func() {
		err := testRunner.decommInstance(conf)
		if err != nil {
			conf.logger.V(1).Info("WARN: Failed to decomm e2e test runner instance", "error", err)
		}
	}()

	session, err := newE2ESession(instanceId, conf)
	if err != nil {
		return "", nil, err
	}

	err = session.setup(conf.regex)
	if err != nil {
		return session.instanceId, nil, err
	}

	testCommandResult, err = session.runTests(conf.regex)
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
	regex := strings.Trim(c.regex, "\"")
	tests := strings.Split(regex, "|")

	for _, testName := range tests {
		e.uploadJUnitReportFromInstance(testName)
		if c.testReportFolder != "" {
			e.downloadJUnitReportToLocalDisk(testName, c.testReportFolder)
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

func splitTests(testsList []string, conf ParallelRunConf) ([]instanceRunConf, error) {
	testPerInstance := len(testsList) / conf.MaxInstances
	if testPerInstance == 0 {
		testPerInstance = 1
	}

	vsphereTestsRe := regexp.MustCompile(vsphereRegex)
	tinkerbellTestsRe := regexp.MustCompile(tinkerbellTestsRe)
	nutanixTestsRe := regexp.MustCompile(nutanixRegex)
	privateNetworkTestsRe := regexp.MustCompile(`^.*(Proxy|RegistryMirror).*$`)
	multiClusterTestsRe := regexp.MustCompile(`^.*Multicluster.*$`)

	runConfs := make([]instanceRunConf, 0, conf.MaxInstances)
	vsphereIPMan := newE2EIPManager(conf.Logger, os.Getenv(vsphereCidrVar))
	vspherePrivateIPMan := newE2EIPManager(conf.Logger, os.Getenv(vspherePrivateNetworkCidrVar))
	nutanixIPMan := newE2EIPManager(conf.Logger, os.Getenv(nutanixCidrVar))

	awsSession, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("creating aws session for tests: %v", err)
	}

	testRunnerConfig, err := NewTestRunnerConfigFromFile(conf.Logger, conf.TestInstanceConfigFile)
	if err != nil {
		return nil, fmt.Errorf("creating test runner config for tests: %v", err)
	}

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
		}
		if nutanixTestsRe.MatchString(testName) {
			ips = nutanixIPMan.reserveIPPool(minIPPoolSize)
		}

		if len(testsInEC2Instance) == testPerInstance || (len(testsList)-1) == i {
			runConfs = append(runConfs, newInstanceRunConf(awsSession, conf, len(runConfs), strings.Join(testsInEC2Instance, "|"), ips, []*api.Hardware{}, Ec2TestRunnerType, testRunnerConfig))
			testsInEC2Instance = make([]string, 0, testPerInstance)
		}
	}

	if strings.EqualFold(conf.BranchName, conf.BaremetalBranchName) {
		tinkerbellIPManager := newE2EIPManager(conf.Logger, os.Getenv(tinkerbellControlPlaneNetworkCidrEnvVar))
		runConfs, err = appendNonAirgappedTinkerbellRunConfs(awsSession, testsList, conf, testRunnerConfig, runConfs, tinkerbellIPManager)
		if err != nil {
			return nil, fmt.Errorf("failed to split Tinkerbell tests: %v", err)
		}

		runConfs, err = appendAirgappedTinkerbellRunConfs(awsSession, testsList, conf, testRunnerConfig, runConfs, tinkerbellIPManager)
		if err != nil {
			return nil, fmt.Errorf("failed to run airgapped Tinkerbell tests: %v", err)
		}
	}

	return runConfs, nil
}

//nolint:gocyclo // This legacy function is complex but the team too busy to simplify it
func appendNonAirgappedTinkerbellRunConfs(awsSession *session.Session, testsList []string, conf ParallelRunConf, testRunnerConfig *TestInfraConfig, runConfs []instanceRunConf, ipManager *E2EIPManager) ([]instanceRunConf, error) {
	err := s3.DownloadToDisk(awsSession, os.Getenv(tinkerbellHardwareS3FileKeyEnvVar), conf.StorageBucket, e2eHardwareCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download tinkerbell hardware csv: %v", err)
	}

	hardware, err := api.NewHardwareSliceFromFile(e2eHardwareCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tinkerbell hardware: %v", err)
	}

	maxHardwarePerE2ETest := tinkerbellDefaultMaxHardwarePerE2ETest
	maxHardwareEnvValue := os.Getenv(maxHardwarePerE2ETestEnvVar)
	if len(maxHardwareEnvValue) > 0 {
		maxHardwarePerE2ETest, err = strconv.Atoi(maxHardwareEnvValue)
		if err != nil {
			return nil, fmt.Errorf("failed to get Tinkerbell max hardware per test env var: %v", err)
		}
	}

	conf.Logger.V(1).Info("INFO:", "totalHardware", len(hardware))

	tinkerbellTests := getTinkerbellNonAirgappedTests(testsList)
	conf.Logger.V(1).Info("INFO:", "tinkerbellTests", len(tinkerbellTests))

	tinkTestInstances := len(hardware) / maxHardwarePerE2ETest
	conf.Logger.V(1).Info("INFO:", "tinkTestInstances", tinkTestInstances)

	tinkTestsPerInstance := 1
	var remainingTests int
	overflowTests := false
	if len(tinkerbellTests) > tinkTestInstances {
		tinkTestsPerInstance = len(tinkerbellTests) / tinkTestInstances
		remainingTests = len(tinkerbellTests) % tinkTestInstances
		if remainingTests != 0 {
			tinkTestsPerInstance++
			overflowTests = true
		}
	}

	conf.Logger.V(1).Info("INFO:", "tinkTestsPerInstance", tinkTestsPerInstance)
	conf.Logger.V(1).Info("INFO:", "tinkTestInstances", tinkTestInstances)
	conf.Logger.V(1).Info("INFO:", "remainingTests", remainingTests)

	hardwareChunks := api.SplitHardware(hardware, maxHardwarePerE2ETest)

	testsInVSphereInstance := make([]string, 0, tinkTestsPerInstance)
	for i, testName := range tinkerbellTests {
		testsInVSphereInstance = append(testsInVSphereInstance, testName)

		if len(testsInVSphereInstance) == tinkTestsPerInstance || (len(testsList)-1) == i {
			conf.Logger.V(1).Info("INFO:", "hardwareChunksSize", len(hardwareChunks))
			conf.Logger.V(1).Info("INFO:", "hardwareSize", len(hardware))

			// each tinkerbell test requires 2 floating ip's (cp & tink server)
			ips := ipManager.reserveIPPool(tinkTestsPerInstance * 2)

			if len(hardwareChunks) > 0 {
				hardware, hardwareChunks = hardwareChunks[0], hardwareChunks[1:]
			}

			runConfs = append(runConfs, newInstanceRunConf(awsSession, conf, len(runConfs), strings.Join(testsInVSphereInstance, "|"), ips, hardware, VSphereTestRunnerType, testRunnerConfig))

			if remainingTests > 0 {
				remainingTests--
			}

			if remainingTests == 0 && overflowTests {
				tinkTestsPerInstance--
				overflowTests = false
			}

			testsInVSphereInstance = make([]string, 0, tinkTestsPerInstance)
		}
	}

	return runConfs, nil
}

func appendAirgappedTinkerbellRunConfs(awsSession *session.Session, testsList []string, conf ParallelRunConf, testRunnerConfig *TestInfraConfig, runConfs []instanceRunConf, ipManager *E2EIPManager) ([]instanceRunConf, error) {
	tinkerbellTests := getTinkerbellAirgappedTests(testsList)
	if len(tinkerbellTests) == 0 {
		conf.Logger.V(1).Info("No tinkerbell airgapped test to run")
		return runConfs, nil
	}

	conf.Logger.V(1).Info("INFO:", "tinkerbellAirGappedTests", len(tinkerbellTests))

	err := s3.DownloadToDisk(awsSession, os.Getenv(tinkerbellAirgappedHardwareS3FileKeyEnvVar), conf.StorageBucket, e2eAirgappedHardwareCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("downloading tinkerbell airgapped hardware csv: %v", err)
	}

	hardware, err := api.NewHardwareSliceFromFile(e2eAirgappedHardwareCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tinkerbell hardware: %v", err)
	}

	conf.Logger.V(1).Info("INFO:", "totalAirgappedHardware", len(hardware))

	pool := ipManager.reserveIPPool(len(tinkerbellTests) * 2)
	runConfs = append(runConfs, newInstanceRunConf(awsSession, conf, len(runConfs), strings.Join(tinkerbellTests, "|"), pool, hardware, VSphereTestRunnerType, testRunnerConfig))

	return runConfs, nil
}

func newInstanceRunConf(awsSession *session.Session, conf ParallelRunConf, jobNumber int, testRegex string, ipPool networkutils.IPPool, hardware []*api.Hardware, testRunnerType TestRunnerType, testRunnerConfig *TestInfraConfig) instanceRunConf {
	jobID := fmt.Sprintf("%s-%d", conf.JobId, jobNumber)
	return instanceRunConf{
		session:             awsSession,
		instanceProfileName: conf.InstanceProfileName,
		storageBucket:       conf.StorageBucket,
		jobId:               jobID,
		parentJobId:         conf.JobId,
		regex:               testRegex,
		ipPool:              ipPool,
		hardware:            hardware,
		bundlesOverride:     conf.BundlesOverride,
		testReportFolder:    conf.TestReportFolder,
		branchName:          conf.BranchName,
		cleanupVms:          conf.CleanupVms,
		testRunnerType:      testRunnerType,
		testRunnerConfig:    *testRunnerConfig,
		logger:              conf.Logger.WithValues("jobID", jobID, "test", testRegex),
	}
}

func logTestGroups(logger logr.Logger, instancesConf []instanceRunConf) {
	testGroups := make([]string, 0, len(instancesConf))
	for _, i := range instancesConf {
		testGroups = append(testGroups, i.regex)
	}
	logger.V(1).Info("Running tests in parallel", "testsGroups", testGroups)
}

package e2e

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

const (
	passedStatus = "pass"
	failedStatus = "fail"

	maxIPPoolSize = 10
	minIPPoolSize = 1
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
}

type (
	testCommandResult    = ssm.RunOutput
	instanceTestsResults struct {
		conf              instanceRunConf
		testCommandResult *testCommandResult
		err               error
	}
)

func RunTestsInParallel(conf ParallelRunConf) error {
	testsList, skippedTests, err := listTests(conf.Regex, conf.TestsToSkip)
	if err != nil {
		return err
	}

	logger.Info("Running tests", "selected", testsList, "skipped", skippedTests)

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
	logTestGroups(instancesConf)
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
	for _, r := range results {
		if r.err != nil {
			logger.Error(r.err, "Failed running e2e tests for instance", "jobId", r.conf.jobId, "instanceId", r.conf.instanceId, "tests", r.conf.regex, "status", failedStatus)
			failedInstances += 1
		} else if !r.testCommandResult.Successful() {
			logger.Info("An e2e instance run has failed", "jobId", r.conf.jobId, "instanceId", r.conf.instanceId, "commandId", r.testCommandResult.CommandId, "tests", r.conf.regex, "status", failedStatus)
			logResult(r.testCommandResult)
			failedInstances += 1
		} else {
			logger.Info("Instance tests completed successfully", "jobId", r.conf.jobId, "instanceId", r.conf.instanceId, "commandId", r.testCommandResult.CommandId, "tests", r.conf.regex, "status", passedStatus)
			logResult(r.testCommandResult)
		}
	}

	if failedInstances > 0 {
		return fmt.Errorf("%d/%d e2e instances failed", failedInstances, len(instancesConf))
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

	defer func() {
		err := testRunner.decommInstance(conf)
		if err != nil {
			logger.V(1).Info("WARN: Failed to decomm e2e test runner instance", "error", err)
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

	key := "Integration-Test-Done"
	value := "TRUE"
	err = testRunner.tagInstance(conf, key, value)
	if err != nil {
		return session.instanceId, nil, fmt.Errorf("tagging instance for e2e success: %v", err)
	}

	return session.instanceId, testCommandResult, nil
}

func (e *E2ESession) runTests(regex string) (testCommandResult *testCommandResult, err error) {
	logger.V(1).Info("Running e2e tests", "regex", regex)
	command := "GOVERSION=go1.16.6 gotestsum --junitfile=junit-testing.xml --raw-command --format=standard-verbose --hide-summary=all --ignore-non-json-output-lines -- test2json -t -p e2e ./bin/e2e.test -test.v"

	if regex != "" {
		command = fmt.Sprintf("%s -test.run \"%s\"", command, regex)
	}

	command = e.commandWithEnvVars(command)

	opt := ssm.WithOutputToCloudwatch()

	testCommandResult, err = ssm.RunCommand(
		e.session,
		e.instanceId,
		command,
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
			e.uploadGeneratedFilesFromInstance(testName)
			e.uploadDiagnosticArchiveFromInstance(testName)
			return nil
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
	privateNetworkTestsRe := regexp.MustCompile(`^.*(Proxy|RegistryMirror).*$`)
	multiClusterTestsRe := regexp.MustCompile(`^.*Multicluster.*$`)

	runConfs := make([]instanceRunConf, 0, conf.MaxInstances)
	ipman := newE2EIPManager(os.Getenv(cidrVar), os.Getenv(privateNetworkCidrVar))

	awsSession, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("creating aws session for tests: %v", err)
	}

	testRunnerConfig, err := NewTestRunnerConfigFromFile(conf.TestInstanceConfigFile)
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
		if privateNetworkTestsRe.MatchString(testName) {
			if multiClusterTest {
				ips = ipman.reservePrivateIPPool(maxIPPoolSize)
			} else {
				ips = ipman.reservePrivateIPPool(minIPPoolSize)
			}
		} else if vsphereTestsRe.MatchString(testName) {
			if multiClusterTest {
				ips = ipman.reserveIPPool(maxIPPoolSize)
			} else {
				ips = ipman.reserveIPPool(minIPPoolSize)
			}
		}

		if len(testsInEC2Instance) == testPerInstance || (len(testsList)-1) == i {
			runConfs = append(runConfs, newInstanceRunConf(awsSession, conf, len(runConfs), strings.Join(testsInEC2Instance, "|"), ips, []*api.Hardware{}, Ec2TestRunnerType, testRunnerConfig))
			testsInEC2Instance = make([]string, 0, testPerInstance)
		}
	}

	if strings.EqualFold(conf.BranchName, "main") {
		runConfs, err = splitTinkerbellTests(awsSession, testsList, conf, testRunnerConfig, runConfs)
		if err != nil {
			return nil, fmt.Errorf("failed to split Tinkerbell tests: %v", err)
		}
	}

	return runConfs, nil
}

func splitTinkerbellTests(awsSession *session.Session, testsList []string, conf ParallelRunConf, testRunnerConfig *TestInfraConfig, runConfs []instanceRunConf) ([]instanceRunConf, error) {
	err := s3.DownloadToDisk(awsSession, os.Getenv(tinkerbellHardwareS3FileKeyEnvVar), conf.StorageBucket, e2eHardwareCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download tinkerbell hardware csv: %v", err)
	}

	hardware, err := api.NewHardwareSliceFromFile(e2eHardwareCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tinkerbell hardware: %v", err)
	}

	maxHardwarePerE2ETest := TinkerbellDefaultMaxHardwarePerE2ETest
	maxHardwareEnvValue := os.Getenv(MaxHardwarePerE2ETestEnvVar)
	if len(maxHardwareEnvValue) > 0 {
		maxHardwarePerE2ETest, err = strconv.Atoi(maxHardwareEnvValue)
		if err != nil {
			return nil, fmt.Errorf("failed to get Tinkerbell max hardware per test env var: %v", err)
		}
	}

	logger.V(1).Info("INFO:", "totalHardware", len(hardware))

	tinkerbellTests := getTinkerbellTests(testsList)
	logger.V(1).Info("INFO:", "tinkerbellTests", len(tinkerbellTests))

	tinkTestInstances := len(hardware) / maxHardwarePerE2ETest
	logger.V(1).Info("INFO:", "tinkTestInstances", tinkTestInstances)

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

	logger.V(1).Info("INFO:", "tinkTestsPerInstance", tinkTestsPerInstance)
	logger.V(1).Info("INFO:", "tinkTestInstances", tinkTestInstances)
	logger.V(1).Info("INFO:", "remainingTests", remainingTests)

	hardwareChunks := api.SplitHardware(hardware, maxHardwarePerE2ETest)

	testsInVSphereInstance := make([]string, 0, tinkTestsPerInstance)
	for i, testName := range tinkerbellTests {
		testsInVSphereInstance = append(testsInVSphereInstance, testName)

		if len(testsInVSphereInstance) == tinkTestsPerInstance || (len(testsList)-1) == i {
			logger.V(1).Info("INFO:", "hardwareChunksSize", len(hardwareChunks))
			logger.V(1).Info("INFO:", "hardwareSize", len(hardware))

			if len(hardwareChunks) > 0 {
				hardware, hardwareChunks = hardwareChunks[0], hardwareChunks[1:]
			}

			runConfs = append(runConfs, newInstanceRunConf(awsSession, conf, len(runConfs), strings.Join(testsInVSphereInstance, "|"), networkutils.IPPool{}, hardware, VSphereTestRunnerType, testRunnerConfig))

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

func newInstanceRunConf(awsSession *session.Session, conf ParallelRunConf, jobNumber int, testRegex string, ipPool networkutils.IPPool, hardware []*api.Hardware, testRunnerType TestRunnerType, testRunnerConfig *TestInfraConfig) instanceRunConf {
	return instanceRunConf{
		session:             awsSession,
		instanceProfileName: conf.InstanceProfileName,
		storageBucket:       conf.StorageBucket,
		jobId:               fmt.Sprintf("%s-%d", conf.JobId, jobNumber),
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
	}
}

func logTestGroups(instancesConf []instanceRunConf) {
	testGroups := make([]string, 0, len(instancesConf))
	for _, i := range instancesConf {
		testGroups = append(testGroups, i.regex)
	}
	logger.V(1).Info("Running tests in parallel", "testsGroups", testGroups)
}

func logResult(t *testCommandResult) {
	// Because of the way we run tests with gotestsum and test2json
	// both cli output and test logs get conveniently combined in stdout
	fmt.Println(string(t.StdOut))
}

package e2e

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/aws/eks-anywhere/internal/pkg/ec2"
	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	passedStatus = "pass"
	failedStatus = "fail"
)

type ParallelRunConf struct {
	MaxInstances        int
	AmiId               string
	InstanceProfileName string
	StorageBucket       string
	JobId               string
	SubnetId            string
	Regex               string
	TestsToSkip         []string
	BundlesOverride     bool
	CleanupVms          bool
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

	var wg sync.WaitGroup
	resultCh := make(chan instanceTestsResults)

	instancesConf := splitTests(testsList, conf)
	logTestGroups(instancesConf)
	for _, instanceConf := range instancesConf {
		go func(c instanceRunConf) {
			defer wg.Done()
			r := instanceTestsResults{conf: c}

			r.conf.instanceId, r.testCommandResult, err = RunTests(c)
			if err != nil {
				r.err = err
			}

			resultCh <- r
		}(instanceConf)
		wg.Add(1)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	failedInstances := 0
	for r := range resultCh {
		if r.err != nil {
			logger.Error(r.err, "Failed running e2e tests for instance", "jobId", r.conf.jobId, "instanceId", r.conf.instanceId, "tests", r.conf.regex, "status", failedStatus)
			failedInstances += 1
		} else if !r.testCommandResult.Successful() {
			logger.Info("An e2e instance run has failed", "jobId", r.conf.jobId, "instanceId", r.conf.instanceId, "commandId", r.testCommandResult.CommandId, "tests", r.conf.regex, "status", failedStatus)
			logResult(r.testCommandResult)
			failedInstances += 1
		} else {
			logger.Info("Ec2 instance tests completed successfully", "jobId", r.conf.jobId, "instanceId", r.conf.instanceId, "commandId", r.testCommandResult.CommandId, "tests", r.conf.regex, "status", passedStatus)
			logResult(r.testCommandResult)
		}
		if conf.CleanupVms {
			err = CleanUpVsphereTestResources(context.Background(), r.conf.instanceId)
			if err != nil {
				logger.Error(err, "Failed to clean up VSphere cluster VMs", "clusterName", r.conf.instanceId)
			}
		}
	}

	if failedInstances > 0 {
		return fmt.Errorf("%d/%d e2e instances failed", failedInstances, len(instancesConf))
	}

	return nil
}

type instanceRunConf struct {
	amiId, instanceProfileName, storageBucket, jobId, parentJobId, subnetId, regex, instanceId, controlPlaneIP string
	bundlesOverride                                                                                            bool
}

func RunTests(conf instanceRunConf) (testInstanceID string, testCommandResult *testCommandResult, err error) {
	session, err := newSession(conf.amiId, conf.instanceProfileName, conf.storageBucket, conf.jobId, conf.subnetId, conf.controlPlaneIP, conf.bundlesOverride)
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

	return session.instanceId, testCommandResult, nil
}

func (e *E2ESession) runTests(regex string) (testCommandResult *testCommandResult, err error) {
	logger.V(1).Info("Running e2e tests", "regex", regex)
	command := "GOVERSION=go1.16.6 gotestsum --junitfile=junit-testing.xml --raw-command --format=standard-verbose --ignore-non-json-output-lines -- test2json -t -p e2e ./bin/e2e.test -test.v"

	if regex != "" {
		command = fmt.Sprintf("%s -test.run %s", command, regex)
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
		return nil, fmt.Errorf("error running e2e tests on instance %s: %v", e.instanceId, err)
	}

	e.uploadJUnitReport(regex)

	if !testCommandResult.Successful() {
		e.uploadGeneratedFilesFromInstance(regex)
		e.uploadDiagnosticArchiveFromInstance(regex)
		return testCommandResult, nil
	}

	key := "Integration-Test-Done"
	value := "TRUE"
	err = ec2.TagInstance(e.session, e.instanceId, key, value)
	if err != nil {
		return nil, fmt.Errorf("error tagging instance for e2e success: %v", err)
	}

	return testCommandResult, nil
}

func (e *E2ESession) commandWithEnvVars(command string) string {
	fullCommand := make([]string, 0, len(e.testEnvVars)+1)

	for k, v := range e.testEnvVars {
		fullCommand = append(fullCommand, fmt.Sprintf("export %s=\"%s\"", k, v))
	}
	fullCommand = append(fullCommand, command)

	return strings.Join(fullCommand, "; ")
}

func splitTests(testsList []string, conf ParallelRunConf) []instanceRunConf {
	testPerInstance := len(testsList) / conf.MaxInstances
	if testPerInstance == 0 {
		testPerInstance = 1
	}

	vsphereTestsRe := regexp.MustCompile(vsphereRegex)
	privateNetworkTestsRe := regexp.MustCompile(`^.*(Proxy|RegistryMirror).*$`)

	runConfs := make([]instanceRunConf, 0, conf.MaxInstances)
	ipman := newE2EIPManager(os.Getenv(cidrVar), os.Getenv(privateNetworkCidrVar))

	testsInCurrentInstance := make([]string, 0, testPerInstance)
	for _, testName := range testsList {
		testsInCurrentInstance = append(testsInCurrentInstance, testName)
		ip := ""
		if privateNetworkTestsRe.MatchString(testName) {
			ip = ipman.getPrivateIP()
		} else if vsphereTestsRe.MatchString(testName) {
			ip = ipman.getIP()
		}
		if len(testsInCurrentInstance) == testPerInstance {
			runConfs = append(runConfs, instanceRunConf{
				amiId:               conf.AmiId,
				instanceProfileName: conf.InstanceProfileName,
				storageBucket:       conf.StorageBucket,
				jobId:               fmt.Sprintf("%s-%d", conf.JobId, len(runConfs)),
				parentJobId:         conf.JobId,
				subnetId:            conf.SubnetId,
				regex:               strings.Join(testsInCurrentInstance, "|"),
				bundlesOverride:     conf.BundlesOverride,
				controlPlaneIP:      ip,
			})

			testsInCurrentInstance = make([]string, 0, testPerInstance)
		}
	}

	return runConfs
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

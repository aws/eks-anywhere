package logfetcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	awscodebuild "github.com/aws/aws-sdk-go/service/codebuild"

	"github.com/aws/eks-anywhere-test-tool/pkg/cloudwatch"
	"github.com/aws/eks-anywhere-test-tool/pkg/codebuild"
	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

type FetchLogsOpt func(options *fetchLogsConfig) (err error)

func WithCodebuildBuild(buildId string) FetchLogsOpt {
	return func(options *fetchLogsConfig) (err error) {
		options.buildId = buildId
		logger.Info("user provided build ID detected", "buildId", buildId)
		return err
	}
}

func WithCodebuildProject(project string) FetchLogsOpt {
	return func(options *fetchLogsConfig) (err error) {
		options.project = project
		logger.Info("user provided project ID detected", "project", project)
		return err
	}
}

type fetchLogsConfig struct {
	buildId string
	project string
}

var (
	ssmCommandExecutionLogStreamTemplate = "%s/%s/aws-runShellScript/%s"
	individualTestLogFileTemplate        = "%s-%s-%s"
)

type testResult struct {
	InstanceId string `json:"instanceId"`
	JobId      string `json:"jobId"`
	CommandId  string `json:"commandId"`
	Tests      string `json:"tests"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}

func (t *testResult) OutputFileName() string {
	return fmt.Sprintf(individualTestLogFileTemplate, t.Tests, t.InstanceId, t.JobId)
}

type (
	testFilter        func(logs []*cloudwatchlogs.OutputLogEvent) (filteredTestsLogs *bytes.Buffer, filteredTestResults []testResult, err error)
	codebuildConsumer func(*awscodebuild.Build) error
	messagesConsumer  func(allMessages, filteredMessages *bytes.Buffer) error
	testConsumer      func(testName string, logs []*cloudwatchlogs.OutputLogEvent) error
)

type LogFetcherOpt func(*testLogFetcher)

func WithTestFilterByName(tests []string) LogFetcherOpt {
	return func(l *testLogFetcher) {
		l.filterTests = newTestFilterByName(tests)
	}
}

func WithLogStdout() LogFetcherOpt {
	return func(l *testLogFetcher) {
		l.processCodebuild = func(*awscodebuild.Build) error { return nil }
		l.processMessages = func(allMessages, filteredMessages *bytes.Buffer) error { return nil }
		l.processTest = logTest
	}
}

type testLogFetcher struct {
	buildAccountCwClient        *cloudwatch.Cloudwatch
	testAccountCwClient         *cloudwatch.Cloudwatch
	buildAccountCodebuildClient *codebuild.Codebuild
	writer                      *testsWriter
	filterTests                 testFilter
	processCodebuild            codebuildConsumer
	processMessages             messagesConsumer
	processTest                 testConsumer
}

func New(buildAccountCwClient *cloudwatch.Cloudwatch, testAccountCwClient *cloudwatch.Cloudwatch, buildAccountCodebuildClient *codebuild.Codebuild, opts ...LogFetcherOpt) *testLogFetcher {
	l := &testLogFetcher{
		buildAccountCwClient:        buildAccountCwClient,
		testAccountCwClient:         testAccountCwClient,
		buildAccountCodebuildClient: buildAccountCodebuildClient,
	}
	for _, o := range opts {
		o(l)
	}

	defultOutputFolder := time.Now().Format(time.RFC3339 + "-logs")

	if l.filterTests == nil {
		l.filterTests = filterFailedTests
	}

	if l.processCodebuild == nil {
		_ = l.ensureWriter(defultOutputFolder)
		l.processCodebuild = l.writer.writeCodeBuild
	}

	if l.processMessages == nil {
		_ = l.ensureWriter(defultOutputFolder)
		l.processMessages = l.writer.writeMessages
	}

	if l.processTest == nil {
		_ = l.ensureWriter(defultOutputFolder)
		l.processTest = l.writer.writeTest
	}

	return l
}

func (l *testLogFetcher) FetchLogs(opts ...FetchLogsOpt) error {
	config := &fetchLogsConfig{
		project: constants.EksATestCodebuildProject,
	}

	for _, opt := range opts {
		err := opt(config)
		if err != nil {
			return fmt.Errorf("failed to set options on fetch logs config: %v", err)
		}
	}

	if config.buildId == "" {
		p, err := l.buildAccountCodebuildClient.FetchLatestBuildForProject(config.project)
		if err != nil {
			return fmt.Errorf("failed to get latest build for project: %v", err)
		}
		config.buildId = *p.Id
		logger.Info("Using latest build for selected project", "buildID", config.buildId, "project", config.project)
	}

	failedTests, err := l.GetBuildProjectLogs(config.project, config.buildId)
	if err != nil {
		return err
	}

	err = l.FetchTestLogs(failedTests)
	if err != nil {
		return err
	}
	return nil
}

func (l *testLogFetcher) GetBuildProjectLogs(project string, buildId string) ([]testResult, error) {
	logger.Info("Fetching build project logs...")
	build, err := l.buildAccountCodebuildClient.FetchBuildForProject(buildId)
	if err != nil {
		return nil, fmt.Errorf("fetching build project logs for project %s: %v", project, err)
	}

	g := build.Logs.GroupName
	s := build.Logs.StreamName

	logs, err := l.buildAccountCwClient.GetLogs(*g, *s)
	if err != nil {
		return nil, fmt.Errorf("fetching cloudwatch logs: %v", err)
	}

	allMsg := allMessages(logs)
	filteredMsg, filteredTests, err := l.filterTests(logs)
	if err != nil {
		return nil, err
	}

	if err = l.processCodebuild(build); err != nil {
		return nil, err
	}

	if err = l.processMessages(allMsg, filteredMsg); err != nil {
		return nil, err
	}

	return filteredTests, nil
}

func (l *testLogFetcher) FetchTestLogs(tests []testResult) error {
	logger.Info("Fetching individual test logs...")
	for _, test := range tests {
		stdout := fmt.Sprintf(ssmCommandExecutionLogStreamTemplate, test.CommandId, test.InstanceId, "stdout")
		logs, err := l.testAccountCwClient.GetLogs(constants.E2eIndividualTestLogGroup, stdout)
		if err != nil {
			logger.Info("error when fetching cloudwatch logs", "error", err)
			return err
		}

		if err := l.processTest(test.Tests, logs); err != nil {
			return nil
		}
	}
	return nil
}

func (l *testLogFetcher) ensureWriter(folderPath string) error {
	if l.writer != nil {
		return nil
	}

	var err error
	l.writer, err = newTestsWriter(folderPath)
	if err != nil {
		return err
	}

	return nil
}

func filterFailedTests(logs []*cloudwatchlogs.OutputLogEvent) (failedTestMessages *bytes.Buffer, failedTestResults []testResult, err error) {
	var failedTests []testResult
	failedTestMessages = &bytes.Buffer{}
	for _, event := range logs {
		if strings.Contains(*event.Message, constants.FailedMessage) {
			msg := *event.Message
			i := strings.Index(msg, "{")
			subMsg := msg[i:]
			var r testResult
			err = json.Unmarshal([]byte(subMsg), &r)
			if err != nil {
				logger.Info("error when unmarshalling json of test results", "error", err)
				return nil, nil, err
			}
			failedTests = append(failedTests, r)
			failedTestMessages.WriteString(subMsg)
		}
	}

	return failedTestMessages, failedTests, nil
}

func allMessages(logs []*cloudwatchlogs.OutputLogEvent) *bytes.Buffer {
	allMsg := new(bytes.Buffer)
	for _, event := range logs {
		allMsg.WriteString(*event.Message)
	}

	return allMsg
}

func newTestFilterByName(tests []string) testFilter {
	lookup := types.SliceToLookup(tests)
	return func(logs []*cloudwatchlogs.OutputLogEvent) (filteredTestsLogs *bytes.Buffer, filteredTestResults []testResult, err error) {
		filteredTestsLogs = &bytes.Buffer{}
		for _, event := range logs {
			if !isResultMessage(*event.Message) {
				continue
			}

			msg := *event.Message
			i := strings.Index(msg, "{")
			subMsg := msg[i:]
			var r testResult
			err = json.Unmarshal([]byte(subMsg), &r)
			if err != nil {
				logger.Info("error when unmarshalling json of test results", "error", err)
				return nil, nil, err
			}

			if !lookup.IsPresent(r.Tests) {
				continue
			}

			filteredTestResults = append(filteredTestResults, r)
			filteredTestsLogs.WriteString(subMsg)
		}

		return filteredTestsLogs, filteredTestResults, nil
	}
}

func isResultMessage(message string) bool {
	return strings.Contains(message, constants.FailedMessage) || strings.Contains(message, constants.SuccessMEssage)
}

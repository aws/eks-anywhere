package logfetcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/aws/eks-anywhere-test-tool/pkg/cloudwatch"
	"github.com/aws/eks-anywhere-test-tool/pkg/codebuild"
	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere-test-tool/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
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

type testLogFetcher struct {
	buildAccountCwClient        *cloudwatch.Cloudwatch
	testAccountCwClient         *cloudwatch.Cloudwatch
	buildAccountCodebuildClient *codebuild.Codebuild
	writer                      filewriter.FileWriter
}

func New(buildAccountCwClient *cloudwatch.Cloudwatch, testAccountCwClient *cloudwatch.Cloudwatch, buildAccountCodebuildCient *codebuild.Codebuild, writer filewriter.FileWriter) *testLogFetcher {
	return &testLogFetcher{
		buildAccountCwClient:        buildAccountCwClient,
		testAccountCwClient:         testAccountCwClient,
		buildAccountCodebuildClient: buildAccountCodebuildCient,
		writer:                      writer,
	}
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

func (l *testLogFetcher) GetBuildProjectLogs(project string, buildId string) (failed []testResult, err error) {
	logger.Info("Fetching build project logs...")
	build, err := l.buildAccountCodebuildClient.FetchBuildForProject(buildId)
	if err != nil {
		return nil, fmt.Errorf("error fetching build project logs for project %s: %v", project, err)
	}

	g := build.Logs.GroupName
	s := build.Logs.StreamName

	logs, err := l.buildAccountCwClient.GetLogs(*g, *s)
	if err != nil {
		return nil, fmt.Errorf("error when fetching cloudwatch logs: %v", err)
	}

	allMsg, failedMsg, failedTests, err := l.extractMessagesFromLog(logs)
	if err != nil {
		return nil, err
	}

	_, err = l.writer.Write(constants.BuildDescriptionFile, []byte(build.String()), filewriter.PersistentFile)
	if err != nil {
		return nil, fmt.Errorf("error when writing build description: %v", err)
	}

	_, err = l.writer.Write(constants.FailedTestsFile, failedMsg.Bytes(), filewriter.PersistentFile)
	if err != nil {
		return nil, err
	}

	_, err = l.writer.Write(constants.LogOutputFile, allMsg.Bytes(), filewriter.PersistentFile)
	if err != nil {
		return nil, err
	}

	return failedTests, nil
}

func (l *testLogFetcher) FetchTestLogs(tests []testResult) error {
	logger.Info("Fetching individual test logs...")
	for _, test := range tests {
		stderr := fmt.Sprintf(ssmCommandExecutionLogStreamTemplate, test.CommandId, test.InstanceId, "stderr")
		logs, err := l.testAccountCwClient.GetLogs(constants.E2eIndividualTestLogGroup, stderr)
		if err != nil {
			logger.Info("error when fetching cloudwatch logs", "error", err)
			return err
		}

		buf := new(bytes.Buffer)
		for _, log := range logs {
			buf.WriteString(*log.Message)
		}

		_, err = l.writer.Write(test.Tests, buf.Bytes(), filewriter.PersistentFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *testLogFetcher) extractMessagesFromLog(logs []*cloudwatchlogs.OutputLogEvent) (allMessages *bytes.Buffer, failedTestMessages *bytes.Buffer, failedTestResults []testResult, err error) {
	var failedTests []testResult
	failedTestsBuf := new(bytes.Buffer)
	buf := new(bytes.Buffer)
	for _, event := range logs {
		buf.WriteString(*event.Message)
		if strings.Contains(*event.Message, constants.FailedMessage) {
			msg := *event.Message
			i := strings.Index(msg, "{")
			subMsg := msg[i:]
			var r testResult
			err = json.Unmarshal([]byte(subMsg), &r)
			if err != nil {
				logger.Info("error when unmarshalling json of test results", "error", err)
				return nil, nil, nil, err
			}
			failedTests = append(failedTests, r)
			failedTestsBuf.WriteString(subMsg)
		}
	}

	return buf, failedTestsBuf, failedTests, nil
}

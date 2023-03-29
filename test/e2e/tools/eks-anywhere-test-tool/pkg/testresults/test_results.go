package testresults

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

type TestFilter func(logs []*cloudwatchlogs.OutputLogEvent) (filteredTestsLogs *bytes.Buffer, filteredTestResults []TestResult, err error)

type TestResult struct {
	InstanceId string `json:"instanceId"`
	JobId      string `json:"jobId"`
	CommandId  string `json:"commandId"`
	Tests      string `json:"tests"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}

func GetFailedTests(logs []*cloudwatchlogs.OutputLogEvent) (failedTestMessages *bytes.Buffer, failedTestResults []TestResult, err error) {
	var failedTests []TestResult
	failedTestMessages = &bytes.Buffer{}
	for _, event := range logs {
		if strings.Contains(*event.Message, constants.FailedMessage) {
			msg := *event.Message
			i := strings.Index(msg, "{")
			subMsg := msg[i:]
			var r TestResult
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

func NewTestFilterByName(tests []string) TestFilter {
	lookup := types.SliceToLookup(tests)
	return func(logs []*cloudwatchlogs.OutputLogEvent) (filteredTestsLogs *bytes.Buffer, filteredTestResults []TestResult, err error) {
		filteredTestsLogs = &bytes.Buffer{}
		for _, event := range logs {
			if !isResultMessage(*event.Message) {
				continue
			}

			msg := *event.Message
			i := strings.Index(msg, "{")
			subMsg := msg[i:]
			var r TestResult
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

func TestResultsJobIdMap(tests []TestResult) map[string]bool {
	m := make(map[string]bool, len(tests))

	for _, test := range tests {
		m[test.JobId] = true
	}
	return m
}

func isResultMessage(message string) bool {
	return strings.Contains(message, constants.FailedMessage) || strings.Contains(message, constants.SuccessMessage)
}

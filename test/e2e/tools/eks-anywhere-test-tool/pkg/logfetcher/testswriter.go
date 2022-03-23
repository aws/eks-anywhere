package logfetcher

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	awscodebuild "github.com/aws/aws-sdk-go/service/codebuild"

	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere-test-tool/pkg/filewriter"
)

type testsWriter struct {
	filewriter.FileWriter
}

func newTestsWriter(folderPath string) (*testsWriter, error) {
	writer, err := filewriter.NewWriter(folderPath)
	if err != nil {
		return nil, fmt.Errorf("setting up tests writer: %v", err)
	}

	return &testsWriter{FileWriter: writer}, nil
}

func (w *testsWriter) writeCodeBuild(build *awscodebuild.Build) error {
	if _, err := w.Write(constants.BuildDescriptionFile, []byte(build.String()), filewriter.PersistentFile); err != nil {
		return fmt.Errorf("writing build description: %v", err)
	}

	return nil
}

func (w *testsWriter) writeMessages(allMessages, filteredMessages *bytes.Buffer) error {
	if _, err := w.Write(constants.FailedTestsFile, filteredMessages.Bytes(), filewriter.PersistentFile); err != nil {
		return err
	}

	if _, err := w.Write(constants.LogOutputFile, allMessages.Bytes(), filewriter.PersistentFile); err != nil {
		return err
	}

	return nil
}

func (w *testsWriter) writeTest(testName string, logs []*cloudwatchlogs.OutputLogEvent) error {
	buf := new(bytes.Buffer)
	for _, log := range logs {
		buf.WriteString(*log.Message)
	}

	if _, err := w.Write(testName, buf.Bytes(), filewriter.PersistentFile); err != nil {
		return err
	}

	return nil
}

package e2e

import (
	"fmt"
	"path/filepath"

	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/logger"
)

func (e *E2ESession) uploadGeneratedFilesFromInstance(testName string) {
	logger.V(1).Info("Uploading log files to s3 bucket")
	command := fmt.Sprintf("aws s3 cp /home/e2e/%s/ %s/%s/ --recursive",
		e.instanceId, e.generatedArtifactsBucketPath(), testName)

	if err := ssm.Run(e.session, e.instanceId, command); err != nil {
		logger.Error(err, "error uploading log files from instance")
	} else {
		logger.V(1).Info("Successfully uploaded log files to S3")
	}
}

func (e *E2ESession) uploadDiagnosticArchiveFromInstance(testName string) {
	bundleNameFormat := "support-bundle-*.tar.gz"
	logger.V(1).Info("Uploading diagnostic bundle to s3 bucket")
	command := fmt.Sprintf("aws s3 cp /home/e2e/ %s/%s/ --recursive --exclude \"*\" --include \"%s\"",
		e.generatedArtifactsBucketPath(), testName, bundleNameFormat)

	if err := ssm.Run(e.session, e.instanceId, command); err != nil {
		logger.Error(err, "error uploading diagnostic bundle from instance")
	} else {
		logger.V(1).Info("Successfully uploaded diagnostic bundle files to S3")
	}
}

func (e *E2ESession) uploadJUnitReport(testName string) {
	junitFile := "junit-testing.xml"
	logger.V(1).Info("Uploading JUnit report to s3 bucket")
	command := fmt.Sprintf("aws s3 cp /home/e2e/ %s/%s/ --recursive --exclude \"*\" --include \"%s\"",
		e.generatedArtifactsBucketPath(), testName, junitFile)

	if err := ssm.Run(e.session, e.instanceId, command); err != nil {
		logger.Error(err, "error uploading JUnit report from instance")
	} else {
		logger.V(1).Info("Successfully uploaded JUnit report files to S3")
	}
}

func (e *E2ESession) downloadJUnitReport(testName, destinationFolder string) {
	junitFile := "junit-testing.xml"
	key := filepath.Join(e.generatedArtifactsPath(), testName, junitFile)
	dst := filepath.Join(destinationFolder, fmt.Sprintf("junit-testing-%s.xml", testName))

	logger.V(1).Info("Downloading JUnit report to disk", "dst", dst)
	if err := s3.DownloadToDisk(e.session, key, e.storageBucket, dst); err != nil {
		logger.Error(err, "Error downloading JUnit report from s3")
	}
}

func (e *E2ESession) generatedArtifactsBucketPath() string {
	return fmt.Sprintf("s3://%s/%s", e.storageBucket, e.generatedArtifactsPath())
}

func (e *E2ESession) generatedArtifactsPath() string {
	return filepath.Join(e.jobId, "generated-artifacts")
}

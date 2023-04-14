package e2e

import (
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/internal/pkg/ssm"
)

const e2eHomeFolder = "/home/e2e/"

func (e *E2ESession) uploadGeneratedFilesFromInstance(testName string) {
	e.logger.V(1).Info("Uploading log files to s3 bucket")
	command := newCopyCommand().from(
		e2eHomeFolder, e.clusterName(e.branchName, e.instanceId, testName),
	).to(
		e.generatedArtifactsBucketPath(), testName,
	).recursive().String()

	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		e.logger.Error(err, "error uploading log files from instance")
	} else {
		e.logger.V(1).Info("Successfully uploaded log files to S3")
	}
}

func (e *E2ESession) uploadDiagnosticArchiveFromInstance(testName string) {
	bundleNameFormat := "support-bundle-*.tar.gz"
	e.logger.V(1).Info("Uploading diagnostic bundle to s3 bucket")
	command := newCopyCommand().from(e2eHomeFolder).to(
		e.generatedArtifactsBucketPath(), testName,
	).recursive().exclude("*").include(bundleNameFormat).String()

	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		e.logger.Error(err, "error uploading diagnostic bundle from instance")
	} else {
		e.logger.V(1).Info("Successfully uploaded diagnostic bundle files to S3")
	}
}

func (e *E2ESession) uploadJUnitReportFromInstance(testName string) {
	junitFile := "junit-testing.xml"
	e.logger.V(1).Info("Uploading JUnit report to s3 bucket")
	command := newCopyCommand().from(e2eHomeFolder).to(
		e.generatedArtifactsBucketPath(), testName,
	).recursive().exclude("*").include(junitFile).String()

	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		e.logger.Error(err, "error uploading JUnit report from instance")
	} else {
		e.logger.V(1).Info("Successfully uploaded JUnit report files to S3")
	}
}

func (e *E2ESession) downloadJUnitReportToLocalDisk(testName, destinationFolder string) {
	junitFile := "junit-testing.xml"
	key := filepath.Join(e.generatedArtifactsPath(), testName, junitFile)
	dst := filepath.Join(destinationFolder, fmt.Sprintf("junit-testing-%s.xml", testName))

	e.logger.V(1).Info("Downloading JUnit report to disk", "dst", dst)
	if err := s3.DownloadToDisk(e.session, key, e.storageBucket, dst); err != nil {
		e.logger.Error(err, "Error downloading JUnit report from s3")
	}
}

func (e *E2ESession) generatedArtifactsBucketPath() string {
	return fmt.Sprintf("s3://%s/%s", e.storageBucket, e.generatedArtifactsPath())
}

func (e *E2ESession) generatedArtifactsPath() string {
	return filepath.Join(e.jobId, "generated-artifacts")
}

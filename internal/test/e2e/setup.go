package e2e

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/eks-anywhere/internal/pkg/ec2"
	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

var requiredFiles = []string{cliBinary, e2eBinary, eksctlBinary}

const (
	cliBinary                  = "eksctl-anywhere"
	e2eBinary                  = "e2e.test"
	eksctlBinary               = "eksctl"
	bundlesReleaseManifestFile = "local-bundle-release.yaml"
	eksAComponentsManifestFile = "local-eksa-components.yaml"
	testNameFile               = "e2e-test-name"
)

type E2ESession struct {
	session             *session.Session
	amiId               string
	instanceProfileName string
	storageBucket       string
	jobId               string
	subnetId            string
	instanceId          string
	controlPlaneIP      string
	testEnvVars         map[string]string
	bundlesOverride     bool
	requiredFiles       []string
}

func newSession(amiId, instanceProfileName, storageBucket, jobId, subnetId, controlPlaneIP string, bundlesOverride bool) (*E2ESession, error) {
	session, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("error creating session: %v", err)
	}

	e := &E2ESession{
		session:             session,
		amiId:               amiId,
		instanceProfileName: instanceProfileName,
		storageBucket:       storageBucket,
		jobId:               jobId,
		subnetId:            subnetId,
		controlPlaneIP:      controlPlaneIP,
		testEnvVars:         make(map[string]string),
		bundlesOverride:     bundlesOverride,
		requiredFiles:       requiredFiles,
	}

	return e, nil
}

func (e *E2ESession) setup(regex string) error {
	err := e.uploadRequiredFiles()
	if err != nil {
		return err
	}

	key := "Integration-Test"
	tag := "EKSA-E2E"
	name := fmt.Sprintf("eksa-e2e-%s", e.jobId)
	logger.V(1).Info("Creating ec2 instance", "name", name)
	instanceId, err := ec2.CreateInstance(e.session, e.amiId, key, tag, e.instanceProfileName, e.subnetId, name)
	if err != nil {
		return fmt.Errorf("error creating instance for e2e tests: %v", err)
	}
	logger.V(1).Info("Instance created", "instance-id", instanceId)
	e.instanceId = instanceId

	logger.V(1).Info("Waiting until SSM is ready")
	err = ssm.WaitForSSMReady(e.session, instanceId)
	if err != nil {
		return fmt.Errorf("error waiting for ssm in new instance: %v", err)
	}

	err = e.createTestNameFile(regex)
	if err != nil {
		return err
	}

	err = e.downloadRequiredFilesInInstance()
	if err != nil {
		return err
	}

	err = e.setupOIDC(regex)
	if err != nil {
		return err
	}

	err = e.setupVSphereEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupFluxEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupProxyEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupRegistryMirrorEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupAwsIam(regex)
	if err != nil {
		return err
	}

	// Adding JobId to Test Env variables
	e.testEnvVars[e2etests.JobIdVar] = e.jobId
	e.testEnvVars[e2etests.BundlesOverrideVar] = strconv.FormatBool(e.bundlesOverride)
	e.testEnvVars[e2etests.ClusterNameVar] = instanceId
	return nil
}

func (e *E2ESession) uploadRequiredFile(file string) error {
	uploadFile := fmt.Sprintf("bin/%s", file)
	key := fmt.Sprintf("%s/%s", e.jobId, file)
	logger.V(1).Info("Uploading file to s3 bucket", "file", file, "key", key)
	err := s3.UploadFile(e.session, uploadFile, key, e.storageBucket)
	if err != nil {
		return fmt.Errorf("error uploading file [%s]: %v", file, err)
	}

	return nil
}

func (e *E2ESession) uploadRequiredFiles() error {
	if e.bundlesOverride {
		e.requiredFiles = append(e.requiredFiles, bundlesReleaseManifestFile)
		if _, err := os.Stat(eksAComponentsManifestFile); err == nil {
			e.requiredFiles = append(e.requiredFiles, eksAComponentsManifestFile)
		} else if errors.Is(err, os.ErrNotExist) {
			logger.V(0).Info("WARNING: no components manifest override found, but bundle override is present. " +
				"If the EKS-A components have changed be sure to provide a components override!")
		} else {
			return err
		}
	}
	for _, file := range e.requiredFiles {
		if file != "eksctl" {
			err := e.uploadRequiredFile(file)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *E2ESession) downloadRequiredFileInInstance(file string) error {
	logger.V(1).Info("Downloading from s3 in instance", "file", file)

	var command string
	if file == "eksctl" {
		command = fmt.Sprintf("aws s3 cp s3://%s/eksctl/%[2]s ./bin/ && chmod 645 ./bin/%[2]s", e.storageBucket, file)
	} else {
		command = fmt.Sprintf("aws s3 cp s3://%s/%s/%[3]s ./bin/ && chmod 645 ./bin/%[3]s", e.storageBucket, e.jobId, file)
	}
	_, err := ssm.Run(e.session, e.instanceId, command)
	if err != nil {
		return fmt.Errorf("error downloading file in instance: %v", err)
	}
	logger.V(1).Info("Successfully downloaded file")

	return nil
}

func (e *E2ESession) uploadGeneratedFilesFromInstance(testName string) {
	logger.V(1).Info("Uploading log files to s3 bucket")
	command := fmt.Sprintf("aws s3 cp /home/e2e/%s/ %s/%s/ --recursive",
		e.instanceId, e.generatedArtifactsBucketPath(), testName)

	_, err := ssm.Run(e.session, e.instanceId, command)
	if err != nil {
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

	_, err := ssm.Run(e.session, e.instanceId, command)
	if err != nil {
		logger.Error(err, "error uploading diagnostic bundle from instance")
	} else {
		logger.V(1).Info("Successfully uploaded diagnostic bundle files to S3")
	}
}

func (e *E2ESession) generatedArtifactsBucketPath() string {
	return fmt.Sprintf("s3://%s/%s/generated-artifacts", e.storageBucket, e.jobId)
}

func (e *E2ESession) downloadRequiredFilesInInstance() error {
	for _, file := range e.requiredFiles {
		err := e.downloadRequiredFileInInstance(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *E2ESession) createTestNameFile(testName string) error {
	command := fmt.Sprintf("echo %s > %s", testName, testNameFile)
	_, err := ssm.Run(e.session, e.instanceId, command)
	if err != nil {
		return fmt.Errorf("error creating test name file in instance: %v", err)
	}
	logger.V(1).Info("Successfully created test name file")

	return nil
}

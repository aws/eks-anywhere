package e2e

import (
	"fmt"
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
)

type E2ESession struct {
	session             *session.Session
	amiId               string
	instanceProfileName string
	storageBucket       string
	jobId               string
	subnetId            string
	instanceId          string
	testEnvVars         map[string]string
	bundlesOverride     bool
}

func newSession(amiId, instanceProfileName, storageBucket, jobId, subnetId string, bundlesOverride bool) (*E2ESession, error) {
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
		testEnvVars:         make(map[string]string),
		bundlesOverride:     bundlesOverride,
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

	// Adding JobId to Test Env variables
	e.testEnvVars[e2etests.JobIdVar] = e.jobId
	e.testEnvVars[e2etests.BundlesOverrideVar] = strconv.FormatBool(e.bundlesOverride)

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
		requiredFiles = append(requiredFiles, bundlesReleaseManifestFile, eksAComponentsManifestFile)
	}
	for _, file := range requiredFiles {
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
		command = fmt.Sprintf("aws s3 cp s3://%s/eksctl/%s ./bin/ && chmod 645 ./bin/%s", e.storageBucket, file, file)
	} else {
		command = fmt.Sprintf("aws s3 cp s3://%s/%s/%s ./bin/ && chmod 645 ./bin/%s", e.storageBucket, e.jobId, file, file)
	}
	err := ssm.Run(e.session, e.instanceId, command)
	if err != nil {
		return fmt.Errorf("error downloading file in instance: %v", err)
	}
	logger.V(1).Info("Successfully downloaded file")

	return nil
}

func (e *E2ESession) downloadRequiredFilesInInstance() error {
	if e.bundlesOverride {
		requiredFiles = append(requiredFiles, bundlesReleaseManifestFile, eksAComponentsManifestFile)
	}
	for _, file := range requiredFiles {
		err := e.downloadRequiredFileInInstance(file)
		if err != nil {
			return err
		}
	}

	return nil
}

package e2e

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/eks-anywhere/internal/pkg/ec2"
	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

var binaries = []string{cliBinary, e2eBinary, eksctlBinary}

const (
	cliBinary    = "eksctl-anywhere"
	e2eBinary    = "e2e.test"
	eksctlBinary = "eksctl"
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
}

func newSession(amiId, instanceProfileName, storageBucket, jobId, subnetId string) (*E2ESession, error) {
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
	}

	return e, nil
}

func (e *E2ESession) setup(regex string) error {
	err := e.uploadBinaries()
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

	err = e.downloadBinariesInInstance()
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

	return nil
}

func (e *E2ESession) uploadBinary(binary string) error {
	file := fmt.Sprintf("bin/%s", binary)
	key := fmt.Sprintf("%s/%s", e.jobId, binary)
	logger.V(1).Info("Uploading binary to s3 bucket", "binary", binary, "key", key)
	err := s3.UploadFile(e.session, file, key, e.storageBucket)
	if err != nil {
		return fmt.Errorf("error uploading binary [%s]: %v", binary, err)
	}

	return nil
}

func (e *E2ESession) uploadBinaries() error {
	for _, binary := range binaries {
		if binary != "eksctl" {
			err := e.uploadBinary(binary)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *E2ESession) downloadBinaryInInstance(binary string) error {
	logger.V(1).Info("Downloading from s3 in instance", "binary", binary)

	var command string
	if binary == "eksctl" {
		command = fmt.Sprintf("aws s3 cp s3://%s/eksctl/%s ./bin/ && chmod 645 ./bin/%s", e.storageBucket, binary, binary)
	} else {
		command = fmt.Sprintf("aws s3 cp s3://%s/%s/%s ./bin/ && chmod 645 ./bin/%s", e.storageBucket, e.jobId, binary, binary)
	}
	err := ssm.Run(e.session, e.instanceId, command)
	if err != nil {
		return fmt.Errorf("error downloading binary in instance: %v", err)
	}
	logger.V(1).Info("Successfully downloaded binary")

	return nil
}

func (e *E2ESession) downloadBinariesInInstance() error {
	for _, binary := range binaries {
		err := e.downloadBinaryInInstance(binary)
		if err != nil {
			return err
		}
	}

	return nil
}

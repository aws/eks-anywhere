package e2e

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

const (
	tinkerbellInventoryCsvFilePathEnvVar       = "T_TINKERBELL_INVENTORY_CSV"
	tinkerbellControlPlaneNetworkCidrEnvVar    = "T_TINKERBELL_CP_NETWORK_CIDR"
	tinkerbellHardwareS3FileKeyEnvVar          = "T_TINKERBELL_S3_INVENTORY_CSV_KEY"
	tinkerbellAirgappedHardwareS3FileKeyEnvVar = "T_TINKERBELL_S3_AG_INVENTORY_CSV_KEY"
	tinkerbellTestsRe                          = `^.*Tinkerbell.*$`
	e2eHardwareCsvFilePath                     = "e2e-inventory.csv"
	e2eAirgappedHardwareCsvFilePath            = "e2e-ag-inventory.csv"
	maxHardwarePerE2ETestEnvVar                = "T_TINKERBELL_MAX_HARDWARE_PER_TEST"
	tinkerbellDefaultMaxHardwarePerE2ETest     = 4
	tinkerbellBootstrapInterfaceEnvVar         = "T_TINKERBELL_BOOTSTRAP_INTERFACE"
	tinkerbellCIEnvironmentEnvVar              = "T_TINKERBELL_CI_ENVIRONMENT"

	// tinkerbellJobTag is the tag used to map vm runners and SSM activations to an e2e job.
	tinkerbellJobTag = "eksa-tinkerbell-e2e-job"
)

// TinkerbellTest maps each Tinkbell test with the hardware count needed for the test.
type TinkerbellTest struct {
	Name  string `yaml:"name"`
	Count int    `yaml:"count"`
}

func (e *E2ESession) setupTinkerbellEnv(testRegex string) error {
	re := regexp.MustCompile(tinkerbellTestsRe)
	if !re.MatchString(testRegex) {
		return nil
	}

	requiredEnvVars := e2etests.RequiredTinkerbellEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	inventoryFileName := fmt.Sprintf("%s.csv", getTestRunnerName(e.logger, e.jobId))
	inventoryFilePath := fmt.Sprintf("bin/%s", inventoryFileName)

	if _, err := os.Stat(inventoryFilePath); err == nil {
		err = os.Remove(inventoryFilePath)
		if err != nil {
			e.logger.V(1).Info("WARN: Failed to clean up existing inventory csv", "file", inventoryFilePath)
		}
	}

	err := api.WriteHardwareSliceToCSV(e.hardware, inventoryFilePath)
	if err != nil {
		return fmt.Errorf("failed to setup tinkerbell test environment: %v", err)
	}

	err = e.uploadRequiredFile(inventoryFileName)
	if err != nil {
		return fmt.Errorf("failed to upload tinkerbell inventory file (%s) : %v", inventoryFileName, err)
	}

	err = e.downloadRequiredFileInInstance(inventoryFileName)
	if err != nil {
		return fmt.Errorf("failed to download tinkerbell inventory file (%s) to test instance : %v", inventoryFileName, err)
	}

	tinkInterface := os.Getenv(tinkerbellBootstrapInterfaceEnvVar)
	if tinkInterface == "" {
		return fmt.Errorf("tinkerbell bootstrap interface env var is required: %s", tinkerbellBootstrapInterfaceEnvVar)
	}

	err = e.setTinkerbellBootstrapIPInInstance(tinkInterface)
	if err != nil {
		return fmt.Errorf("failed to set tinkerbell boostrap ip on interface (%s) in test instance : %v", tinkInterface, err)
	}

	e.testEnvVars[tinkerbellInventoryCsvFilePathEnvVar] = inventoryFilePath
	e.testEnvVars[tinkerbellCIEnvironmentEnvVar] = "true"

	return nil
}

func (e *E2ESession) setTinkerbellBootstrapIPInInstance(tinkInterface string) error {
	e.logger.V(1).Info("Setting Tinkerbell Bootstrap IP in instance")

	command := fmt.Sprintf("export T_TINKERBELL_BOOTSTRAP_IP=$(/sbin/ip -o -4 addr list %s | awk '{print $4}' | cut -d/ -f1) && echo T_TINKERBELL_BOOTSTRAP_IP=\"$T_TINKERBELL_BOOTSTRAP_IP\" | tee -a /etc/environment", tinkInterface)
	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		return fmt.Errorf("setting tinkerbell boostrap ip: %v", err)
	}

	e.logger.V(1).Info("Successfully set tinkerbell boostrap ip")

	return nil
}

// Get non airgapped, normal tinkerbell tests.
func getTinkerbellNonAirgappedTests(tests []string) []string {
	tinkerbellTestsRe := regexp.MustCompile(tinkerbellTestsRe)
	airgappedRe := regexp.MustCompile(`^.*Airgapped.*$`)
	var tinkerbellTests []string

	for _, testName := range tests {
		if tinkerbellTestsRe.MatchString(testName) && !airgappedRe.MatchString(testName) {
			tinkerbellTests = append(tinkerbellTests, testName)
		}
	}
	return tinkerbellTests
}

func getTinkerbellAirgappedTests(tests []string) []string {
	tinkerbellTestsRe := regexp.MustCompile(tinkerbellTestsRe)
	airgappedRe := regexp.MustCompile(`^.*Airgapped.*$`)
	var tinkerbellTests []string

	for _, testName := range tests {
		if tinkerbellTestsRe.MatchString(testName) && airgappedRe.MatchString(testName) {
			tinkerbellTests = append(tinkerbellTests, testName)
		}
	}
	return tinkerbellTests
}

// ReadTinkerbellMachinePool returns the list of baremetal machines designated for e2e tests.
func ReadTinkerbellMachinePool(session *session.Session, bucketName string) ([]*api.Hardware, error) {
	hardware := []*api.Hardware{}
	machines, err := nonAirgappedHardwarePool(session, bucketName)
	if err != nil {
		return nil, err
	}
	hardware = append(hardware, machines...)

	machines, err = airgappedHardwarePool(session, bucketName)
	if err != nil {
		return nil, err
	}
	hardware = append(hardware, machines...)

	return hardware, nil
}

func nonAirgappedHardwarePool(session *session.Session, storageBucket string) ([]*api.Hardware, error) {
	err := s3.DownloadToDisk(session, os.Getenv(tinkerbellHardwareS3FileKeyEnvVar), storageBucket, e2eHardwareCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download tinkerbell hardware csv: %v", err)
	}

	hardware, err := api.ReadTinkerbellHardwareFromFile(e2eHardwareCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tinkerbell hardware: %v", err)
	}
	return hardware, nil
}

// airgappedHardwarePool returns the hardware pool for airgapped tinkerbell tests.
// Airgapped tinkerbell tests have special hardware requirements that doesn't have internet connectivity.
func airgappedHardwarePool(session *session.Session, storageBucket string) ([]*api.Hardware, error) {
	err := s3.DownloadToDisk(session, os.Getenv(tinkerbellAirgappedHardwareS3FileKeyEnvVar), storageBucket, e2eAirgappedHardwareCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("downloading tinkerbell airgapped hardware csv: %v", err)
	}

	hardware, err := api.ReadTinkerbellHardwareFromFile(e2eAirgappedHardwareCsvFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tinkerbell hardware: %v", err)
	}

	return hardware, nil
}

type TinkerbellSSMInstances struct {
	// InstanceIDs is a list of SSM instance IDs created for the vm runners.
	InstanceIDs []string
	// ActivationIDs is a list of SSM activation IDs created for the vm runners.
	ActivationIDs []string
}

// ListTinkerbellSSMInstances returns a list of SSM instances created for the tinkerbell vm runners.
func ListTinkerbellSSMInstances(ctx context.Context, session *session.Session) (*TinkerbellSSMInstances, error) {
	runnerInstances := &TinkerbellSSMInstances{}

	instances, err := ssm.ListInstancesByTags(ctx, session, ssm.Tag{Key: tinkerbellJobTag, Value: "*"})
	if err != nil {
		return nil, fmt.Errorf("listing tinkerbell runners: %v", err)
	}

	for _, instance := range instances {
		runnerInstances.ActivationIDs = append(runnerInstances.ActivationIDs, *instance.ActivationId)
		runnerInstances.InstanceIDs = append(runnerInstances.InstanceIDs, *instance.InstanceId)
	}

	return runnerInstances, nil
}

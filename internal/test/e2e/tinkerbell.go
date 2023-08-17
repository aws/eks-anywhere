package e2e

import (
	"fmt"
	"os"
	"regexp"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/internal/pkg/api"
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
	e.testEnvVars[e2etests.TinkerbellCIEnvironment] = "true"

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

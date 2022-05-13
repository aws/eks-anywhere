package e2e

import (
	"fmt"
	"os"
	"regexp"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

const (
	tinkerbellInventoryCsvFilePathEnvVar = "T_TINKERBELL_INVENTORY_CSV"
	tinkerbellHardwareS3FileKeyEnvVar    = "T_TINKERBELL_S3_INVENTORY_CSV_KEY"
	tinkerbellTestsRe                    = `^.*Tinkerbell.*$`
	e2eHardwareCsvFilePath               = "e2e-inventory.csv"
	MaxHardwarePerE2ETestEnvVar          = "T_TINKERBELL_MAX_HARDWARE_PER_TEST"
)

func (e *E2ESession) setupTinkerbellEnv(testRegex string) error {
	re := regexp.MustCompile(tinkerbellTestsRe)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running Tinkerbell tests, skipping Env variable setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredTinkerbellEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	inventoryFileName := fmt.Sprintf("%s.csv", getTestRunnerName(e.jobId))
	inventoryFilePath := fmt.Sprintf("bin/%s", inventoryFileName)
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

	e.testEnvVars[tinkerbellInventoryCsvFilePathEnvVar] = inventoryFilePath

	return nil
}

func getTinkerbellTests(tests []string) []string {
	tinkerbellTestsRe := regexp.MustCompile(tinkerbellTestsRe)
	var tinkerbellTests []string
	for _, testName := range tests {
		if tinkerbellTestsRe.MatchString(testName) {
			tinkerbellTests = append(tinkerbellTests, testName)
		}
	}
	return tinkerbellTests
}

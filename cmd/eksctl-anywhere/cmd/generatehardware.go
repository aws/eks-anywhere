package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

type hardwareOptions struct {
	csvPath    string
	outputPath string
}

// Flag name constants
const generateHardwareFilenameFlagName = "filename"

var hOpts = &hardwareOptions{}

var generateHardwareCmd = &cobra.Command{
	Use:   "hardware",
	Short: "Generate hardware files",
	Long: `
Generate Kubernetes hardware YAML manifests for each Hardware entry in the source.
`,
	RunE: hOpts.generateHardware,
}

func init() {
	generateCmd.AddCommand(generateHardwareCmd)

	flags := generateHardwareCmd.Flags()

	flags.StringVarP(&hOpts.csvPath, generateHardwareFilenameFlagName, "f", "", "CSV file path")
	if err := generateHardwareCmd.MarkFlagRequired(generateHardwareFilenameFlagName); err != nil {
		panic(err)
	}

	flags.StringVarP(&hOpts.outputPath, "output", "o", "", "directory path to output hardware files; Tinkerbell JSON files are stored under a \"json\" subdirectory")
}

func (hOpts *hardwareOptions) generateHardware(cmd *cobra.Command, args []string) error {
	csvFile, err := os.Open(hOpts.csvPath)
	if err != nil {
		return fmt.Errorf("csv: %v", err)
	}

	outputDir, err := hardware.CreateManifestDir(hOpts.outputPath)
	if err != nil {
		return err
	}

	hardwareYAML, err := os.Create(filepath.Join(outputDir, hardware.DefaultHardwareManifestYAMLFilename))
	if err != nil {
		return fmt.Errorf("tinkerbell manifest yaml: %v", err)
	}
	yamlWriter := hardware.NewTinkerbellManifestYAML(hardwareYAML)

	reader, err := hardware.NewCSVReader(csvFile)
	if err != nil {
		return fmt.Errorf("csv: %v", err)
	}

	validator := hardware.NewDefaultMachineValidator()

	if err := hardware.TranslateAll(reader, yamlWriter, validator); err != nil {
		return err
	}

	return nil
}

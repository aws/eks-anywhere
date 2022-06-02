package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

type hardwareOptions struct {
	csvPath    string
	outputPath string
}

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

	flags.StringVarP(&hOpts.csvPath, tinkerbellHardwareCSVFlag, "w", "", "CSV file path")
	if err := generateHardwareCmd.MarkFlagRequired(tinkerbellHardwareCSVFlag); err != nil {
		panic(err)
	}

	flags.StringVarP(&hOpts.outputPath, "output", "o", "", "directory path to output hardware YAML")
}

func (hOpts *hardwareOptions) generateHardware(cmd *cobra.Command, args []string) error {
	csvFile, err := os.Open(hOpts.csvPath)
	if err != nil {
		return fmt.Errorf("csv: %v", err)
	}

	reader, err := hardware.NewCSVReader(csvFile)
	if err != nil {
		return fmt.Errorf("csv: %v", err)
	}

	fh, err := hardware.CreateOrStdout(hOpts.outputPath)
	if err != nil {
		return err
	}
	bufferedWriter := bufio.NewWriter(fh)
	defer bufferedWriter.Flush()
	writer := hardware.NewTinkerbellManifestYAML(bufferedWriter)

	validator := hardware.NewDefaultMachineValidator()

	return hardware.TranslateAll(reader, writer, validator)
}

package cmd

import (
	"bufio"
	"fmt"

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
	flags.StringVarP(&hOpts.outputPath, "output", "o", "", "Path to output hardware YAML.")
	flags.StringVarP(
		&hOpts.csvPath,
		TinkerbellHardwareCSVFlagName,
		TinkerbellHardwareCSVFlagAlias,
		"",
		TinkerbellHardwareCSVFlagDescription,
	)

	if err := generateHardwareCmd.MarkFlagRequired(TinkerbellHardwareCSVFlagName); err != nil {
		panic(err)
	}
}

func (hOpts *hardwareOptions) generateHardware(cmd *cobra.Command, args []string) error {
	hardwareYaml, err := hardware.BuildHardwareYAML(hOpts.csvPath)
	if err != nil {
		return fmt.Errorf("building hardware yaml from csv: %v", err)
	}

	fh, err := hardware.CreateOrStdout(hOpts.outputPath)
	if err != nil {
		return err
	}
	bufferedWriter := bufio.NewWriter(fh)
	defer bufferedWriter.Flush()
	_, err = bufferedWriter.Write(hardwareYaml)
	if err != nil {
		return fmt.Errorf("writing hardware yaml to output: %v", err)
	}

	return nil
}

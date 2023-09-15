package cmd

import (
	"bufio"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

type hardwareOptions struct {
	csvPath    string
	outputPath string
	config     tinkerbell.Config
}

var hOpts = &hardwareOptions{}

var generateHardwareCmd = &cobra.Command{
	Use:     "hardware",
	Short:   "Generate hardware files",
	PreRunE: bindFlagsToViper,
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

	generateHardwareCmd.Flags().StringVar(&cc.tinkerbellConfig.Rufio.WebhookSecret, "webhook-secrets", "", "Comma separated list of secrets for use with the bare metal webhook provider")
	markFlagHidden(generateHardwareCmd.Flags(), "webhook-secrets")
	generateHardwareCmd.Flags().StringVar(&cc.tinkerbellConfig.Rufio.ConsumerURL, "consumer-url", "", "URL for the bare metal webhook consumer")
	markFlagHidden(generateHardwareCmd.Flags(), "webhook-url")
}

func (hOpts *hardwareOptions) generateHardware(cmd *cobra.Command, args []string) error {
	hardwareYaml, err := hardware.BuildHardwareYAML(hOpts.csvPath, hOpts.config.Rufio.WebhookSecret)
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

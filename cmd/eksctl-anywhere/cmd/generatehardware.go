package cmd

import (
	"bufio"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

type hardwareOptions struct {
	csvPath         string
	outputPath      string
	providerOptions *dependencies.ProviderOptions
}

var hOpts = &hardwareOptions{
	providerOptions: &dependencies.ProviderOptions{
		Tinkerbell: &dependencies.TinkerbellOptions{
			BMCOptions: &hardware.BMCOptions{
				RPC: &hardware.RPCOpts{},
			},
		},
	},
}

var generateHardwareCmd = &cobra.Command{
	Use:     "hardware",
	Short:   "Generate hardware files",
	Long:    `Generate Kubernetes hardware YAML manifests for each Hardware entry in the source.`,
	RunE:    hOpts.generateHardware,
	PreRunE: bindFlagsToViper,
}

func init() {
	generateCmd.AddCommand(generateHardwareCmd)

	fset := generateHardwareCmd.Flags()
	fset.StringVarP(&hOpts.outputPath, "output", "o", "", "Path to output hardware YAML.")
	fset.StringVarP(
		&hOpts.csvPath,
		TinkerbellHardwareCSVFlagName,
		TinkerbellHardwareCSVFlagAlias,
		"",
		TinkerbellHardwareCSVFlagDescription,
	)

	if err := generateHardwareCmd.MarkFlagRequired(TinkerbellHardwareCSVFlagName); err != nil {
		panic(err)
	}
	tinkerbellFlags(fset, hOpts.providerOptions.Tinkerbell.BMCOptions.RPC)
}

func (hOpts *hardwareOptions) generateHardware(cmd *cobra.Command, args []string) error {
	hardwareYaml, err := hardware.BuildHardwareYAML(hOpts.csvPath, hOpts.providerOptions.Tinkerbell.BMCOptions)
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

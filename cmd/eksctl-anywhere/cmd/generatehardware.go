package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/types"
)

type hardwareOptions struct {
	csvPath      string
	outputPath   string
	tinkerbellIp string
	grpcPort     string
	certPort     string
	dryRun       bool
}

const (
	defaultGrpcPort = "42113"
	defaultCertPort = "42114"
)

var hOpts = &hardwareOptions{}

var generateHardwareCmd = &cobra.Command{
	Use:   "hardware",
	Short: "Generate hardware files",
	Long:  "This command is used to generate hardware JSON and YAML files used for tinkerbell provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		return hOpts.generateHardware(cmd.Context())
	},
}

func init() {
	generateCmd.AddCommand(generateHardwareCmd)
	generateHardwareCmd.Flags().StringVarP(&hOpts.csvPath, "filename", "f", "", "path to csv file")
	generateHardwareCmd.Flags().StringVarP(&hOpts.outputPath, "output", "o", "", "directory path to output hardware files")
	generateHardwareCmd.Flags().StringVar(&hOpts.tinkerbellIp, "tinkerbell-ip", "", "Tinkerbell stack IP, required unless --dry-run flag is set")
	generateHardwareCmd.Flags().StringVar(&hOpts.grpcPort, "grpc-port", defaultGrpcPort, "Tinkerbell GRPC Authority port")
	generateHardwareCmd.Flags().StringVar(&hOpts.certPort, "cert-port", defaultCertPort, "Tinkerbell Cert URL port")
	generateHardwareCmd.Flags().BoolVar(&hOpts.dryRun, "dry-run", false, "set this flag to skip pushing Hardware to tinkerbell stack automatically")
	err := generateHardwareCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func (hOpts *hardwareOptions) generateHardware(ctx context.Context) error {
	if err := validateOptions(hOpts); err != nil {
		return err
	}

	csvFile, err := os.Open(hOpts.csvPath)
	if err != nil {
		return fmt.Errorf("csv: %v", err)
	}

	manifestFile, err := os.Create(hardware.DefaultHardwareManifestYamlFilename)
	if err != nil {
		return fmt.Errorf("tinkerbell manifest yaml: %v", err)
	}
	yamlWriter := hardware.NewTinkerbellManifestYaml(manifestFile)

	var journal hardware.Journal
	jsonFactory := hardware.RecordingTinkerbellHardwareJsonFactory(hardware.DefaultTinkerbellHardwareJsonDir, &journal)
	jsonWriter := hardware.NewTinkerbellHardwareJsonWriter(jsonFactory)

	reader, err := hardware.NewCsvReader(csvFile)
	if err != nil {
		return fmt.Errorf("csv: %v", err)
	}

	writer := hardware.NewTeeWriterWith(yamlWriter, jsonWriter)
	validator := hardware.NewDefaultMachineValidator()

	if err := hardware.TranslateAll(reader, writer, validator); err != nil {
		return err
	}

	if !hOpts.dryRun {
		tink, close, err := tinkExecutableFromOpts(ctx, hOpts)
		if err != nil {
			return err
		}
		defer close.Close(ctx)

		if err := hardware.RegisterTinkerbellHardware(ctx, tink, journal); err != nil {
			return err
		}
	}

	return nil
}

func validateOptions(opts *hardwareOptions) error {
	if err := networkutils.ValidateIP(opts.tinkerbellIp); err != nil {
		return fmt.Errorf("invalid tinkerbell-ip: %v", err)
	}

	if !networkutils.IsPortValid(opts.grpcPort) {
		return fmt.Errorf("invalid grpc-port: %v", opts.certPort)
	}

	if !networkutils.IsPortValid(opts.certPort) {
		return fmt.Errorf("invalid cert-port: %v", opts.certPort)
	}

	return nil
}

func tinkExecutableFromOpts(ctx context.Context, opts *hardwareOptions) (*executables.Tink, types.Closer, error) {
	executableBuilder, close, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return nil, nil, fmt.Errorf("initialize executables: %v", err)
	}

	cert := fmt.Sprintf("http://%s:%s/cert", opts.tinkerbellIp, opts.certPort)
	grpc := fmt.Sprintf("%s:%s", opts.tinkerbellIp, opts.grpcPort)

	return executableBuilder.BuildTinkExecutable(cert, grpc), close, nil
}

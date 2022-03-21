package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/types"
)

type hardwareOptions struct {
	csvPath          string
	outputPath       string
	tinkerbellIp     string
	grpcPort         string
	certPort         string
	skipRegistration bool
}

const (
	defaultGrpcPort = "42113"
	defaultCertPort = "42114"
)

// Flag name constants
const (
	generateHardwareFilenameFlagName     = "filename"
	generateHardwareTinkerbellIpFlagName = "tinkerbell-ip"
)

var hOpts = &hardwareOptions{}

var generateHardwareCmd = &cobra.Command{
	Use:   "hardware",
	Short: "Generate hardware files",
	Long: `
Generate hardware JSON and YAML files used for Tinkerbell provider. Tinkerbell 
hardware JSON are registered with a Tinkerbell stack. Use --skip-registration 
to prevent Tinkerbell stack interactions.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return hOpts.generateHardware(cmd.Context())
	},
}

func init() {
	generateCmd.AddCommand(generateHardwareCmd)

	flags := generateHardwareCmd.Flags()

	flags.StringVarP(&hOpts.csvPath, generateHardwareFilenameFlagName, "f", "", "CSV file path")
	if err := generateHardwareCmd.MarkFlagRequired(generateHardwareFilenameFlagName); err != nil {
		panic(err)
	}

	flags.StringVar(&hOpts.tinkerbellIp, generateHardwareTinkerbellIpFlagName, "", "Tinkerbell stack IP address; not required with --skip-registration")
	if err := generateHardwareCmd.MarkFlagRequired(generateHardwareTinkerbellIpFlagName); err != nil {
		panic(err)
	}

	flags.StringVarP(&hOpts.outputPath, "output", "o", "", "directory path to output hardware files; Tinkerbell JSON files are stored under a \"json\" subdirectory")
	flags.StringVar(&hOpts.grpcPort, "grpc-port", defaultGrpcPort, "Tinkerbell GRPC Authority port")
	flags.StringVar(&hOpts.certPort, "cert-port", defaultCertPort, "Tinkerbell Cert URL port")
	flags.BoolVar(&hOpts.skipRegistration, "skip-registration", false, "skip hardware registration with the Tinkerbell stack")
}

func (hOpts *hardwareOptions) generateHardware(ctx context.Context) error {
	if !hOpts.skipRegistration {
		if err := validateOptions(hOpts); err != nil {
			return err
		}
	}

	csvFile, err := os.Open(hOpts.csvPath)
	if err != nil {
		return fmt.Errorf("csv: %v", err)
	}

	outputDir, err := hardware.CreateManifestDir(hOpts.outputPath)
	if err != nil {
		return err
	}

	jsonDir, err := hardware.CreateDefaultJsonDir(outputDir)
	if err != nil {
		return err
	}

	hardwareYaml, err := os.Create(filepath.Join(outputDir, hardware.DefaultHardwareManifestYamlFilename))
	if err != nil {
		return fmt.Errorf("tinkerbell manifest yaml: %v", err)
	}
	yamlWriter := hardware.NewTinkerbellManifestYaml(hardwareYaml)

	var journal hardware.Journal
	jsonFactory, err := hardware.RecordingTinkerbellHardwareJsonFactory(jsonDir, &journal)
	if err != nil {
		return err
	}
	jsonWriter := hardware.NewTinkerbellHardwareJsonWriter(jsonFactory)

	reader, err := hardware.NewCsvReader(csvFile)
	if err != nil {
		return fmt.Errorf("csv: %v", err)
	}

	writer := hardware.MultiMachineWriter(yamlWriter, jsonWriter)
	validator := hardware.NewDefaultMachineValidator()

	if err := hardware.TranslateAll(reader, writer, validator); err != nil {
		return err
	}

	if !hOpts.skipRegistration {
		tink, closer, err := tinkExecutableFromOpts(ctx, hOpts)
		if err != nil {
			return err
		}
		defer closer.Close(ctx)

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

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/hardware"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type hardwareOptions struct {
	csvPath      string
	tinkerbellIp string
	grpcPort     string
	certPort     string
	skipPush     bool
}

const (
	defaultGrpcPort = "42113"
	defaultCertPort = "42114"
)

var hOpts = &hardwareOptions{}

var generateHardwareCmd = &cobra.Command{
	Use:    "hardware",
	Short:  "Generate hardware files",
	Long:   "This command is used to generate hardware JSON and YAML files used for tinkerbell provider",
	PreRun: preRunGenerateHardware,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := hOpts.generateHardware(cmd.Context())
		if err != nil {
			log.Fatalf("Error filling the provider config: %v", err)
		}
		return nil
	},
}

func init() {
	generateCmd.AddCommand(generateHardwareCmd)
	generateHardwareCmd.Flags().StringVarP(&hOpts.csvPath, "filename", "f", "", "path to csv file")
	generateHardwareCmd.Flags().StringVar(&hOpts.tinkerbellIp, "tinkerbell-ip", "", "tinkerbell stack IP, required unless --skip-push flag is specified")
	generateHardwareCmd.Flags().StringVar(&hOpts.grpcPort, "grpc-port", defaultGrpcPort, "tinkerbell GRPC Authority port [Default: 42113]")
	generateHardwareCmd.Flags().StringVar(&hOpts.certPort, "cert-port", defaultCertPort, "tinkerbell Cert URL port [Default: 42114]")
	generateHardwareCmd.Flags().BoolVar(&hOpts.skipPush, "skip-push", false, "set this flag to skip pushing Hardware to tinkerbell stack automatically")
	err := generateHardwareCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func preRunGenerateHardware(cmd *cobra.Command, args []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
}

func (hOpts *hardwareOptions) generateHardware(ctx context.Context) error {
	if !hOpts.skipPush && hOpts.tinkerbellIp == "" {
		return errors.New("tinkerbell-ip is required, please specify it using --tinkerbell-ip")
	}

	csv, err := hardware.NewCsvParser(hOpts.csvPath)
	if err != nil {
		return err
	}

	defer csv.Close()

	json, err := hardware.NewJsonParser()
	if err != nil {
		return err
	}

	defer json.CleanUp()

	yaml, err := hardware.NewYamlParser()
	if err != nil {
		return err
	}

	defer yaml.Close()

	var tink *executables.Tink
	if !hOpts.skipPush {

		if err := networkutils.ValidateIP(hOpts.tinkerbellIp); err != nil {
			return fmt.Errorf("tinkerbell-ip is not valid: %v", err)
		}

		executableBuilder, close, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
		if err != nil {
			return fmt.Errorf("unable to initialize executables: %v", err)
		}
		defer close.CheckErr(ctx)

		if !networkutils.IsPortValid(hOpts.grpcPort) {
			return fmt.Errorf("grpc-port %s is invalid", hOpts.certPort)
		}

		if !networkutils.IsPortValid(hOpts.certPort) {
			return fmt.Errorf("cert-port %s is invalid", hOpts.certPort)
		}

		cert := fmt.Sprintf("http://%s:%s/cert", hOpts.tinkerbellIp, hOpts.certPort)
		grpc := fmt.Sprintf("%s:%s", hOpts.tinkerbellIp, hOpts.grpcPort)
		tink = executableBuilder.BuildTinkExecutable(cert, grpc)
	}

	for {
		items, err := csv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed reading csv: %v", err)
		}

		id := uuid.New().String()
		hardware, err := json.GetHardwareJson(id, items[csv.HostnameIndex], items[csv.IpAddressIndex], items[csv.GatewayIndex], items[csv.NetmaskIndex], items[csv.MacIndex], items[csv.NameServerIndex])
		if err != nil {
			return fmt.Errorf("error getting hardware json: %v", err)
		}

		filename := fmt.Sprintf("%s.json", items[csv.HostnameIndex])
		logger.V(4).Info("Writing hardware json", "Filename", filename)
		if err := json.Write(filename, hardware); err != nil {
			return err
		}

		if !hOpts.skipPush {
			logger.V(4).Info("Pushing hardware", "Hardware", items[csv.MacIndex])
			if err := tink.PushHardware(ctx, hardware); err != nil {
				return err
			}
		}

		if err := yaml.WriteHardwareYaml(id, items[csv.HostnameIndex], items[csv.BmcIpIndex], items[csv.VendorIndex], items[csv.BmcUsernameIndex], items[csv.BmcPasswordIndex]); err != nil {
			return fmt.Errorf("error writing hardware yaml: %v", err)
		}

	}
	return nil
}

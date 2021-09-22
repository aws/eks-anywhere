package cmd

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers/factory"
	support "github.com/aws/eks-anywhere/pkg/support"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/version"
)

type generateSupportBundleOptions struct {
	fileName string
}

var gsbo = &generateSupportBundleOptions{}

var generateBundleConfigCmd = &cobra.Command{
	Use:     "support-bundle-config",
	Short:   "Generate support bundle config",
	Long:    "This command is used to generate a default support bundle config yaml",
	PreRunE: preRunGenerateBundleConfigCmd,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := gsbo.validateCmdInput()
		if err != nil {
			return fmt.Errorf("command input validation failed: %v", err)
		}
		bundle, err := gsbo.generateBundleConfig(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to generate bunlde config: %v", err)
		}
		err = bundle.PrintBundleConfig()
		if err != nil {
			return fmt.Errorf("failed to print bundle config: %v", err)
		}
		return nil
	},
}

func init() {
	generateCmd.AddCommand(generateBundleConfigCmd)
	generateBundleConfigCmd.Flags().StringVarP(&gsbo.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
}

func preRunGenerateBundleConfigCmd(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func (gsbo *generateSupportBundleOptions) validateCmdInput() error {
	f := gsbo.fileName
	if f != "" {
		clusterConfigFileExist := validations.FileExists(f)
		if !clusterConfigFileExist {
			return fmt.Errorf("the cluster config file %s does not exist", f)
		}
		_, err := v1alpha1.ValidateClusterConfig(f)
		if err != nil {
			return fmt.Errorf("unable to get cluster config from file: %v", err)
		}
	}
	return nil
}

func (gsbo *generateSupportBundleOptions) generateBundleConfig(ctx context.Context) (*support.EksaDiagnosticBundle, error) {
	f := gsbo.fileName
	if f == "" {
		return support.NewDiagnosticBundleDefault(support.NewAnalyzerFactory(), support.NewCollectorFactory("")), nil
	}

	clusterSpec, err := cluster.NewSpec(f, version.Get())
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	eksaToolsImage := clusterSpec.VersionsBundle.Eksa.CliTools
	image := eksaToolsImage.VersionedImage()
	executableBuilder, err := executables.NewExecutableBuilder(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize executables: %v", err)
	}

	writerDir := fmt.Sprintf(clusterSpec.Name)
	writer, err := filewriter.NewWriter(writerDir)
	if err != nil {
		return nil, fmt.Errorf("unable to write: %v", err)
	}

	clusterawsadm := executableBuilder.BuildClusterAwsAdmExecutable()
	kubectl := executableBuilder.BuildKubectlExecutable()
	govc := executableBuilder.BuildGovcExecutable(writer)
	docker := executables.BuildDockerExecutable()

	providerFactory := &factory.ProviderFactory{
		AwsClient:            clusterawsadm,
		DockerClient:         docker,
		DockerKubectlClient:  kubectl,
		VSphereGovcClient:    govc,
		VSphereKubectlClient: kubectl,
		Writer:               writer,
		SkipIpCheck:          cc.skipIpCheck,
	}
	provider, err := providerFactory.BuildProvider(cc.fileName, clusterSpec.Cluster)
	if err != nil {
		return nil, err
	}

	collectorImage := clusterSpec.VersionsBundle.Eksa.DiagnosticCollector.VersionedImage()
	af := support.NewAnalyzerFactory()
	cf := support.NewCollectorFactory(collectorImage)
	opts := support.EksaDiagnosticBundleOpts{
		AnalyzerFactory:  af,
		CollectorFactory: cf,
	}
	return support.NewDiagnosticBundleFromSpec(clusterSpec, provider, gsbo.kubeConfig(clusterSpec.Name), opts)
}

func (gsbo *generateSupportBundleOptions) kubeConfig(clusterName string) string {
	if csbo.wConfig == "" {
		return filepath.Join(clusterName, fmt.Sprintf(kubeconfigPattern, clusterName))
	}
	return csbo.wConfig
}

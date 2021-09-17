package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
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
		bundle, err := gsbo.generateBundleConfig()
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

func (gsbo *generateSupportBundleOptions) generateBundleConfig() (*support.EksaDiagnosticBundle, error) {
	f := gsbo.fileName
	if f == "" {
		return support.NewDiagnosticBundleDefault(support.NewAnalyzerFactory(), support.NewCollectorFactory()), nil
	}
	clusterSpec, err := cluster.NewSpec(f, version.Get())
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster config from file: %v", err)
	}
	opts := support.EksaDiagnosticBundleOpts{
		AnalyzerFactory:  support.NewAnalyzerFactory(),
		CollectorFactory: support.NewCollectorFactory(),
		ClusterSpec:      clusterSpec,
	}
	return support.NewDiagnosticBundleFromSpec(opts), nil
}

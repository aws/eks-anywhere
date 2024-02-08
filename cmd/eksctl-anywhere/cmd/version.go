package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/version"
)

type versionOptions struct {
	output string
}

var vo = &versionOptions{}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get the eksctl anywhere version",
	Long:  "This command prints the version of eksctl anywhere",
	RunE: func(cmd *cobra.Command, args []string) error {
		return vo.printVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().StringVarP(&vo.output, "output", "o", "", "specifies the output format (valid option: json)")
}

func (vo *versionOptions) printVersion() error {
	versionInfo, bundlesErr := version.GetFullVersionInfo()
	switch vo.output {
	case "":
		fmt.Printf("Version: %s\n", versionInfo.GitVersion)
		if bundlesErr != nil {
			return fmt.Errorf("error getting bundle manifest URL for version %s: %v", versionInfo.GitVersion, bundlesErr)
		}
		fmt.Printf("Release Manifest URL: %s\n", versionInfo.ReleaseManifestURL)
		fmt.Printf("Bundle Manifest URL: %s\n", versionInfo.BundleManifestURL)
	case "json":
		versionInfoJSON, unmarshalErr := json.Marshal(versionInfo)
		if unmarshalErr != nil {
			return fmt.Errorf("error marshaling version info to JSON: %v", unmarshalErr)
		}
		fmt.Printf("%s\n", versionInfoJSON)
		if bundlesErr != nil {
			return fmt.Errorf("error getting bundle manifest URL for version %s: %v", versionInfo, bundlesErr)
		}
	case "yaml":
		versionInfoYAML, unmarshalErr := yaml.Marshal(versionInfo)
		if unmarshalErr != nil {
			return fmt.Errorf("error marshaling version info to YAML: %v", unmarshalErr)
		}
		fmt.Printf("%s\n", versionInfoYAML)
		if bundlesErr != nil {
			return fmt.Errorf("error getting bundle manifest URL for version %s: %v", versionInfo, bundlesErr)
		}
	default:
		return fmt.Errorf("invalid output format: %s", vo.output)
	}

	return nil
}

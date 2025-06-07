package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/certificates"
)

type renewCertificatesOptions struct {
	configFile string
	component  string
}

var rc = &renewCertificatesOptions{}

var renewCertificatesCmd = &cobra.Command{
	Use:          "certificates",
	Short:        "Renew certificates",
	Long:         "Renew external ETCD and control plane certificates",
	PreRunE:      bindFlagsToViper,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return rc.renewCertificates(cmd)
	},
}

func init() {
	renewCmd.AddCommand(renewCertificatesCmd)
	renewCertificatesCmd.Flags().StringVarP(&rc.configFile, "config", "f", "", "Config file containing node and SSH information")
	renewCertificatesCmd.Flags().StringVarP(&rc.component, "component", "c", "", "Component to renew certificates for (etcd or control-plane). If not specified, renews both.")
	if err := renewCertificatesCmd.MarkFlagRequired("config"); err != nil {
		log.Fatalf("marking config as required: %s", err)
	}
}

func validateComponent(component string) error {
	if component != "" && component != "etcd" && component != "control-plane" {
		return fmt.Errorf("invalid component %q, must be either 'etcd' or 'control-plane'", component)
	}
	return nil
}

func (rc *renewCertificatesOptions) renewCertificates(_ *cobra.Command) error {
	if err := validateComponent(rc.component); err != nil {
		return err
	}

	_, err := certificates.ParseConfig(rc.configFile)
	if err != nil {
		return fmt.Errorf("parsing config file: %v", err)
	}
	return nil
}

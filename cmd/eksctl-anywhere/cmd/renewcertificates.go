package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/certificates"
	"github.com/aws/eks-anywhere/pkg/types"
)

type renewCertificatesOptions struct {
	configFile string
	component  string
}

var rc = &renewCertificatesOptions{}

var renewCertificatesCmd = &cobra.Command{
	Use:   "certificates",
	Short: "Renew certificates for cluster components",
	Long: `Renew certificates for etcd and control plane nodes in EKS Anywhere cluster.

This command supports certificate renewal for both etcd and control plane components.
You can choose to renew certificates for all components or specify a single component.

Example config file (yaml):
  clusterName: my-cluster
  controlPlane:
    nodes:
    - 192.168.1.10
    - 192.168.1.11
    - 192.168.1.12
    os: ubuntu    # ubuntu, rhel, or bottlerocket
    sshKey: /path/to/ssh/private-key
    sshUser: ec2-user
  etcd:
    nodes:
    - 192.168.1.20
    - 192.168.1.21
    - 192.168.1.22
    os: ubuntu    # ubuntu, rhel, or bottlerocket
    sshKey: /path/to/ssh/private-key
    sshUser: ec2-user`,
	PreRunE:      bindFlagsToViper,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return rc.renewCertificates(cmd)
	},
}

func init() {
	renewCmd.AddCommand(renewCertificatesCmd)

	renewCertificatesCmd.Flags().StringVarP(&rc.configFile, "config", "f", "", "Config file containing node and SSH information")
	renewCertificatesCmd.Flags().StringVarP(&rc.component, "component", "c", "", "Component to renew certificates for (etcd or control-plane). If not specified, renews both.")

	if err := renewCertificatesCmd.MarkFlagRequired("config"); err != nil {
		panic(err)
	}
}

func validateComponent(component string) error {
	if component != "" && component != "etcd" && component != "control-plane" {
		return fmt.Errorf("invalid component %q, must be either 'etcd' or 'control-plane'", component)
	}
	return nil
}

func (rc *renewCertificatesOptions) renewCertificates(cmd *cobra.Command) error {
	if err := validateComponent(rc.component); err != nil {
		return err
	}

	config, err := certificates.ParseConfig(rc.configFile)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	cluster := &types.Cluster{
		Name: config.ClusterName,
	}

	renewer, err := certificates.NewRenewer()
	if err != nil {
		return fmt.Errorf("failed to create renewer: %v", err)
	}
	return renewer.RenewCertificates(cmd.Context(), cluster, config, rc.component)
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/certificates"
	"github.com/aws/eks-anywhere/pkg/types"
)

type renewCertificatesOptions struct {
	configFile  string
	clusterName string
	component   string
	sshKey      string
}

var rc = &renewCertificatesOptions{}

var renewCertificatesCmd = &cobra.Command{
	Use:   "certificates",
	Short: "Renew certificates for cluster components",
	Long: `Renew certificates for etcd and control plane nodes in EKS Anywhere cluster.

This command supports certificate renewal for both etcd and control plane components.
You can choose to renew certificates for all components or specify a single component.

The command can be used in two modes:

1. Using cluster name (--cluster-name):
   For functional clusters with expiring certificates, provide the cluster name and SSH key.
   The command will automatically:
   - Get node information from the cluster
   - Detect node operating systems
   - Get SSH username from cluster configuration

   Required:
   - KUBECONFIG environment variable must be set
   - Cluster configuration file must exist in ./<cluster-name>/<cluster-name>-eks-a-cluster.yaml
   - SSH private key must be provided via --ssh-key flag

   Example:
     export CLUSTER_NAME=my-cluster
     export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
     eksctl anywhere certificates renew --cluster-name ${CLUSTER_NAME} --ssh-key ~/.ssh/id_ed25519

2. Using config file (--config):
   For clusters with expired certificates, provide a YAML file with node and SSH information.

   Example config file:
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
	renewCertificatesCmd.Flags().StringVarP(&rc.clusterName, "cluster-name", "n", "", "Name of the cluster to renew certificates for")
	renewCertificatesCmd.Flags().StringVarP(&rc.component, "component", "c", "", "Component to renew certificates for (etcd or control-plane). If not specified, renews both.")
	renewCertificatesCmd.Flags().StringVar(&rc.sshKey, "ssh-key", "", "SSH private key file (required when using --cluster-name)")
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

	if (rc.configFile == "") == (rc.clusterName == "") {
		return fmt.Errorf("must specify exactly one of --config or --cluster-name")
	}

	if rc.clusterName != "" && rc.sshKey == "" {
		return fmt.Errorf("--ssh-key is required when using --cluster-name")
	}

	var config *certificates.RenewalConfig
	var err error

	if rc.configFile != "" {
		config, err = certificates.ParseConfig(rc.configFile)
		if err != nil {
			return fmt.Errorf("failed to parse config file: %v", err)
		}
	} else {
		if os.Getenv("KUBECONFIG") == "" {
			return fmt.Errorf("KUBECONFIG environment variable must be set when using --cluster-name")
		}
		config, err = certificates.BuildConfigFromCluster(rc.clusterName, rc.sshKey)
		if err != nil {
			return fmt.Errorf("failed to build config from cluster: %v", err)
		}
	}

	cluster := &types.Cluster{
		Name: rc.clusterName,
	}
	if rc.configFile != "" {
		cluster.Name = config.ClusterName
	}

	renewer, err := certificates.NewRenewer()
	if err != nil {
		return fmt.Errorf("failed to create renewer: %v", err)
	}
	return renewer.RenewCertificates(cmd.Context(), cluster, config, rc.component)
}

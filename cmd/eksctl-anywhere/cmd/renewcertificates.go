package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/certificates"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
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
	RunE:         rc.renewCertificates,
}

func init() {
	renewCmd.AddCommand(renewCertificatesCmd)
	renewCertificatesCmd.Flags().StringVarP(&rc.configFile, "config", "f", "", "Config file containing node and SSH information")
	renewCertificatesCmd.Flags().StringVarP(&rc.component, "component", "c", "", fmt.Sprintf("Component to renew certificates for (%s or %s). If not specified, renews both.", constants.EtcdComponent, constants.ControlPlaneComponent))

	if err := renewCertificatesCmd.MarkFlagRequired("config"); err != nil {
		logger.Fatal(err, "marking config as required")
	}
}

func (rc *renewCertificatesOptions) renewCertificates(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	cfg, err := certificates.ParseConfig(rc.configFile)
	if err != nil {
		return err
	}

	deps, err := dependencies.NewFactory().
		WithExecutableBuilder().
		WithKubectl().
		WithUnAuthKubeClient().
		Build(ctx)
	if err != nil {
		return err
	}

	kubeCfgPath := kubeconfig.FromClusterName(cfg.ClusterName)

	kubeClient := deps.UnAuthKubeClient.KubeconfigClient(kubeCfgPath)

	cluster := &types.Cluster{
		Name:           cfg.ClusterName,
		KubeconfigFile: kubeCfgPath,
	}

	if err := certificates.PopulateConfig(ctx, cfg, kubeClient, cluster); err != nil {
		return err
	}

	if err := certificates.ValidateConfig(cfg, rc.component); err != nil {
		return err
	}

	os := cfg.OS
	if os == string(v1alpha1.Ubuntu) || os == string(v1alpha1.RedHat) {
		os = string(certificates.OSTypeLinux)
	}

	renewer, err := certificates.NewRenewer(kubeClient, os, cfg)
	if err != nil {
		return err
	}

	return renewer.RenewCertificates(ctx, cfg, rc.component)
}

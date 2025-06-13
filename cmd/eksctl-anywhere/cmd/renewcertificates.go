package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/certificates"
	"github.com/aws/eks-anywhere/pkg/constants"
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
	renewCertificatesCmd.Flags().IntVarP(&certificates.VerbosityLevel, "verbosity", "v", 0, "Set the verbosity level")

	if err := renewCertificatesCmd.MarkFlagRequired("config"); err != nil {
		logger.Fatal(err, "marking config as required")
	}
}

func (rc *renewCertificatesOptions) renewCertificates(cmd *cobra.Command, _ []string) error {
	if err := certificates.ValidateComponent(rc.component); err != nil {
		return err
	}

	config, err := certificates.ParseConfig(rc.configFile)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	if err := certificates.ValidateComponentWithConfig(rc.component, config); err != nil {
		return err
	}

	osType := certificates.DetermineOSType(rc.component, config)

	renewer, err := certificates.NewRenewerWithClusterName(osType, config.ClusterName)
	if err != nil {
		return fmt.Errorf("failed to create renewer: %v", err)
	}

	return rc.executeRenewal(cmd.Context(), config, osType, renewer)
}

func (rc *renewCertificatesOptions) executeRenewal(ctx context.Context, config *certificates.RenewalConfig, _ string, renewer *certificates.Renewer) error {
	cluster := &types.Cluster{
		Name: config.ClusterName,
	}

	return renewer.RenewCertificates(ctx, cluster, config, rc.component)
}

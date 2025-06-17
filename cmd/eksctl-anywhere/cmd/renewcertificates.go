package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/certificates"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/logger"
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

// newRenewerForCmd builds dependencies & returns a ready to-use Renewer.
func newRenewerForCmd(ctx context.Context, cfg *certificates.RenewalConfig) (*certificates.Renewer, error) {
	mountDirs := certificates.GetSSHKeyDirs(cfg)

	deps, err := dependencies.NewFactory().
		WithExecutableBuilder().
		WithExecutableMountDirs(mountDirs...). // ssh key in container
		WithKubectl().
		WithUnAuthKubeClient().
		Build(ctx)
	if err != nil {
		return nil, err
	}

	sshRunner, err := certificates.NewSSHRunner(cfg.ControlPlane.SSH)
	if err != nil {
		return nil, err
	}

	// kubeCfgPath := kubeconfig.FromClusterName(cfg.ClusterName)

	kubeCfgPath, err := certificates.ResolveKubeconfigPath(cfg.ClusterName)
	if err != nil {
		return nil, err
	}

	kubeClient := deps.UnAuthKubeClient.KubeconfigClient(kubeCfgPath)

	osKey := cfg.OS
	if osKey == string(v1alpha1.Ubuntu) || osKey == string(v1alpha1.RedHat) {
		osKey = string(certificates.OSTypeLinux)
	}
	osRenewer, err := certificates.BuildOSRenewer(osKey)
	if err != nil {
		return nil, err
	}

	return certificates.NewRenewer(kubeClient, sshRunner, osRenewer)
}

func (rc *renewCertificatesOptions) renewCertificates(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	cfg, err := certificates.ParseConfig(rc.configFile)
	if err != nil {
		return err
	}
	if err = certificates.ValidateComponentWithConfig(rc.component, cfg); err != nil {
		return err
	}

	renewer, err := newRenewerForCmd(ctx, cfg)
	if err != nil {
		return err
	}

	return renewer.RenewCertificates(ctx, cfg, rc.component)
}

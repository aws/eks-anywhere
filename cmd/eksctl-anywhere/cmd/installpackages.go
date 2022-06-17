package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/version"
)

type installPackageOptions struct {
	source        curatedpackages.BundleSource
	kubeVersion   string
	packageName   string
	registry      string
	customConfigs []string
	showOptions   bool
}

var ipo = &installPackageOptions{}

func init() {
	installCmd.AddCommand(installPackageCommand)
	installPackageCommand.Flags().Var(&ipo.source, "source", "Location to find curated packages: (cluster, registry)")
	if err := installPackageCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	installPackageCommand.Flags().StringVar(&ipo.kubeVersion, "kube-version", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	installPackageCommand.Flags().StringVarP(&ipo.packageName, "package-name", "n", "", "Custom name of the curated package to install")
	if err := installPackageCommand.MarkFlagRequired("package-name"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	installPackageCommand.Flags().StringVar(&ipo.registry, "registry", "", "Used to specify an alternative registry for discovery")
	installPackageCommand.Flags().BoolVar(&ipo.showOptions, "show-options", false, "Used to specify the package options to be used for configuration")
	installPackageCommand.Flags().StringArrayVar(&ipo.customConfigs, "set", []string{}, "Provide custom configurations for curated packages. Format key:value")
}

var installPackageCommand = &cobra.Command{
	Use:          "package [package] [flags]",
	Aliases:      []string{"package"},
	Short:        "Install package",
	Long:         "This command is used to Install a curated package. Use list to discover curated packages",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE:         runInstallPackages,
	Args:         cobra.ExactArgs(1),
}

func runInstallPackages(cmd *cobra.Command, args []string) error {
	if err := curatedpackages.ValidateKubeVersion(ipo.kubeVersion, ipo.source); err != nil {
		return err
	}

	return installPackages(cmd.Context(), args)
}

func installPackages(ctx context.Context, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	deps, err := NewDependenciesForPackages(ctx, WithRegistryName(ipo.registry), WithKubeVersion(ipo.kubeVersion), WithMountPaths(kubeConfig))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	bm := curatedpackages.CreateBundleManager(ipo.kubeVersion)

	b := curatedpackages.NewBundleReader(
		kubeConfig,
		ipo.kubeVersion,
		ipo.source,
		deps.Kubectl,
		bm,
		version.Get(),
		deps.BundleRegistry,
	)

	bundle, err := b.GetLatestBundle(ctx)
	if err != nil {
		return err
	}

	packages := curatedpackages.NewPackageClient(
		deps.Kubectl,
		curatedpackages.WithBundle(bundle),
		curatedpackages.WithCustomConfigs(ipo.customConfigs),
		curatedpackages.WithShowOptions(ipo.showOptions),
	)

	p, err := packages.GetPackageFromBundle(args[0])
	if err != nil {
		return err
	}

	if ipo.showOptions {
		configs := curatedpackages.GetConfigurationsFromBundle(p)
		curatedpackages.DisplayConfigurationOptions(configs)
		return nil
	}

	curatedpackages.PrintLicense()
	err = packages.InstallPackage(ctx, p, ipo.packageName, kubeConfig)
	if err != nil {
		return err
	}
	return nil
}

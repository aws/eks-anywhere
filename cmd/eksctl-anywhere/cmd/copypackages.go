package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/registry"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// copyPackagesCmd is the context for the copy packages command.
var copyPackagesCmd = &cobra.Command{
	Use:          "packages",
	Short:        "Copy package images and charts from a source to a destination",
	Long:         `Copy all the EKS Anywhere package images and helm charts from a source to a destination.`,
	SilenceUsage: true,
	RunE:         runCopyPackages,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err == nil {
			return nil
		}
		return fmt.Errorf("A destination must be specified as an argument")
	},
}

func init() {
	copyCmd.AddCommand(copyPackagesCmd)

	copyPackagesCmd.Flags().StringVarP(&copyPackagesCommand.bundleFile, "bundle", "b", "", "EKS-A bundle file to read artifact dependencies from")
	if err := copyPackagesCmd.MarkFlagRequired("bundle"); err != nil {
		log.Fatalf("Cannot mark 'bundle' flag as required: %s", err)
	}
	copyPackagesCmd.Flags().StringVarP(&copyPackagesCommand.dstCert, "dst-cert", "", "", "Package bundles file to read artifact dependencies from")
	copyPackagesCmd.Flags().StringVarP(&copyPackagesCommand.srcCert, "src-cert", "", "", "Package bundles file to read artifact dependencies from")
	copyPackagesCmd.Flags().BoolVar(&copyPackagesCommand.insecure, "insecure", false, "Skip TLS verification while copying images and charts")
	copyPackagesCmd.Flags().BoolVar(&copyPackagesCommand.dryRun, "dry-run", false, "Dry run copy to print images that would be copied")
}

var copyPackagesCommand = CopyPackagesCommand{}

// CopyPackagesCommand copies packages specified in a bundle to a destination.
type CopyPackagesCommand struct {
	destination   string
	bundleFile    string
	srcCert       string
	dstCert       string
	insecure      bool
	dryRun        bool
	registryCache *registry.Cache
	dstRegistry   registry.StorageClient
}

func runCopyPackages(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	copyPackagesCommand.destination = args[0]
	return copyPackagesCommand.call(ctx)
}

func (c CopyPackagesCommand) call(ctx context.Context) error {
	factory := dependencies.NewFactory()
	deps, err := factory.
		WithManifestReader().
		Build(ctx)
	if err != nil {
		return err
	}

	eksaBundle, err := bundles.Read(deps.ManifestReader, c.bundleFile)
	if err != nil {
		return err
	}
	bundleReader := curatedpackages.NewPackageReader(deps.ManifestReader)

	bundleReader.AllImages = false
	imageList := bundleReader.ReadChartsFromBundles(ctx, eksaBundle)

	c.registryCache = registry.NewCache()
	dstRegistry, err := c.registryCache.Get(c.destination, c.dstCert, c.insecure)
	if err != nil {
		return fmt.Errorf("error with repository %s: %v", c.destination, err)
	}
	c.dstRegistry = *dstRegistry

	err = c.copyImages(ctx, imageList)
	if err != nil {
		return err
	}

	imageList, err = bundleReader.ReadImagesFromBundles(ctx, eksaBundle)
	if err != nil {
		return err
	}
	c.dstRegistry.SetProject("curated-packages/")
	return c.copyImages(ctx, imageList)
}

func (c CopyPackagesCommand) copyImages(ctx context.Context, imageList []releasev1.Image) error {
	for _, image := range imageList {
		host := image.Registry()
		srcRegistry, err := c.registryCache.Get(host, c.srcCert, c.insecure)
		if err != nil {
			return fmt.Errorf("error with repository %s: %v", host, err)
		}

		if c.dryRun {
			continue
		}

		err = (*srcRegistry).Copy(ctx, image, c.dstRegistry)
		if err != nil {
			return err
		}
	}
	return nil
}

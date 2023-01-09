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
	Short:        "Copy curated package images and charts from a source to a destination",
	Long:         `Copy all the EKS Anywhere curated package images and helm charts from a source to a destination.`,
	SilenceUsage: true,
	RunE:         runCopyPackages,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return fmt.Errorf("A destination must be specified as an argument")
		}
		return nil
	},
}

func init() {
	copyCmd.AddCommand(copyPackagesCmd)

	copyPackagesCmd.Flags().StringVarP(&copyPackagesCommand.bundleFile, "bundle", "b", "", "EKS-A bundle file to read artifact dependencies from")
	if err := copyPackagesCmd.MarkFlagRequired("bundle"); err != nil {
		log.Fatalf("Cannot mark 'bundle' flag as required: %s", err)
	}
	copyPackagesCmd.Flags().StringVarP(&copyPackagesCommand.dstCert, "dst-cert", "", "", "TLS certificate for destination registry")
	copyPackagesCmd.Flags().StringVarP(&copyPackagesCommand.srcCert, "src-cert", "", "", "TLS certificate for source registry")
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
}

func runCopyPackages(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	copyPackagesCommand.destination = args[0]

	credentialStore := registry.NewCredentialStore()
	err := credentialStore.Init()
	if err != nil {
		return err
	}

	return copyPackagesCommand.call(ctx, credentialStore)
}

func (c CopyPackagesCommand) call(ctx context.Context, credentialStore registry.CredentialStore) error {
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

	imageList := bundleReader.ReadChartsFromBundles(ctx, eksaBundle)

	certificates, err := registry.GetCertificates(c.dstCert)
	if err != nil {
		return err
	}

	dstContext := registry.NewStorageContext(c.destination, credentialStore, certificates, c.insecure)
	c.registryCache = registry.NewCache()
	dstRegistry, err := c.registryCache.Get(dstContext)
	if err != nil {
		return fmt.Errorf("error with repository %s: %v", c.destination, err)
	}

	err = c.copyImages(ctx, dstRegistry, credentialStore, imageList)
	if err != nil {
		return err
	}

	imageList, err = bundleReader.ReadImagesFromBundles(ctx, eksaBundle)
	if err != nil {
		return err
	}
	dstRegistry.SetProject("curated-packages/")
	return c.copyImages(ctx, dstRegistry, credentialStore, imageList)
}

func (c CopyPackagesCommand) copyImages(ctx context.Context, dstRegistry registry.StorageClient, credentialStore registry.CredentialStore, imageList []releasev1.Image) error {
	certificates, err := registry.GetCertificates(c.srcCert)
	if err != nil {
		return err
	}

	for _, image := range imageList {
		host := image.Registry()

		srcContext := registry.NewStorageContext(host, credentialStore, certificates, c.insecure)
		srcRegistry, err := c.registryCache.Get(srcContext)
		if err != nil {
			return fmt.Errorf("error with repository %s: %v", host, err)
		}

		artifact := registry.NewArtifact(image.Registry(), image.Repository(), image.Version(), image.Digest())
		log.Println(dstRegistry.Destination(artifact))
		if c.dryRun {
			continue
		}

		err = registry.Copy(ctx, srcRegistry, dstRegistry, artifact)
		if err != nil {
			return err
		}
	}
	return nil
}

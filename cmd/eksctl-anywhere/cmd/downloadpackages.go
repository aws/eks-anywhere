package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"oras.land/oras/cmd/oras/root"

	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/registry"
)

// downloadPackagesCmd is the context for the download packages command.
var downloadPackagesCmd = &cobra.Command{
	Use:          "packages <destination-registry>",
	Short:        "Download curated package images and charts",
	Long:         `Download all the EKS Anywhere curated package images and helm charts.`,
	SilenceUsage: true,
	RunE:         runDownloadPackages,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return fmt.Errorf("A destination directory must be specified as an argument")
		}
		return nil
	},
}

func init() {
	downloadCmd.AddCommand(downloadPackagesCmd)

	downloadPackagesCmd.Flags().StringVarP(&downloadPackagesCommand.bundleFile, "bundle", "b", "", "EKS-A bundle file to read artifact dependencies from")
	if err := downloadPackagesCmd.MarkFlagRequired("bundle"); err != nil {
		logger.Fatal(err, "Cannot mark 'bundle' flag as required")
	}
	downloadPackagesCmd.Flags().StringVarP(&downloadPackagesCommand.awsRegion, "aws-region", "", os.Getenv(config.EksaRegionEnv), "Region to download images from")
}

var downloadPackagesCommand = DownloadPackagesCommand{}

// DownloadPackagesCommand copies packages specified in a bundle to a destination.
type DownloadPackagesCommand struct {
	destination   string
	bundleFile    string
	awsRegion     string
	registryCache *registry.Cache
}

func runDownloadPackages(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	downloadPackagesCommand.destination = args[0]

	credentialStore := registry.NewCredentialStore()
	err := credentialStore.Init()
	if err != nil {
		return err
	}

	return downloadPackagesCommand.call(ctx, credentialStore)
}

func (c DownloadPackagesCommand) call(ctx context.Context, credentialStore *registry.CredentialStore) error {
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

	c.registryCache = registry.NewCache()
	bundleReader := curatedpackages.NewPackageReader(c.registryCache, credentialStore, c.awsRegion)

	imageList := bundleReader.ReadChartsFromBundles(ctx, eksaBundle)

	logger.Info("Downloading curated packages helm charts from public ECR", "destination", c.destination)
	err = c.downloadImages(imageList, c.destination, "curated-packages")
	if err != nil {
		return err
	}

	imageList, err = bundleReader.ReadImagesFromBundles(ctx, eksaBundle)
	if err != nil {
		return err
	}
	logger.Info("Downloading curated packages images from private ECR", "destination", c.destination)
	return c.downloadImages(imageList, c.destination, "curated-packages")
}

func (c DownloadPackagesCommand) downloadImages(imageList []registry.Artifact, destination, project string) error {
	for _, image := range imageList {
		source := image.VersionedImage()
		dst := path.Join(destination, project, image.Repository)
		err := downloadSingleImage(source, dst)
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadSingleImage(source, destination string) error {
	args := []string{"copy", source, destination}
	cmd := root.New()
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("reading %s: %v", source, err)
	}
	return nil
}

package cmd

import (
	"context"
	"github.com/aws/eks-anywhere/pkg/registry"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/oras"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/tar"
	"github.com/aws/eks-anywhere/pkg/version"
)

// imagesCmd represents the images command.
var downloadImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Download all eks-a images to disk",
	Long: `Creates a tarball containing all necessary images
to create an eks-a cluster for any of the supported
Kubernetes versions.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		return downloadImagesRunner.Run(ctx)
	},
}

func init() {
	downloadCmd.AddCommand(downloadImagesCmd)

	downloadImagesCmd.Flags().StringVarP(&downloadImagesRunner.outputFile, "output", "o", "", "Output tarball containing all downloaded images")
	if err := downloadImagesCmd.MarkFlagRequired("output"); err != nil {
		log.Fatalf("Cannot mark 'output' flag as required: %s", err)
	}

	downloadImagesCmd.Flags().BoolVar(&downloadImagesRunner.includePackages, "include-packages", false, "Flag to indicate inclusion of curated packages in downloaded images")
	downloadImagesCmd.Flags().BoolVar(&downloadImagesRunner.insecure, "insecure", false, "Flag to indicate skipping TLS verification while downloading helm charts")
}

var downloadImagesRunner = downloadImagesCommand{}

type downloadImagesCommand struct {
	outputFile      string
	includePackages bool
	insecure        bool
}

func (c downloadImagesCommand) Run(ctx context.Context) error {
	factory := dependencies.NewFactory()
	helmOpts := []executables.HelmOpt{}
	if c.insecure {
		helmOpts = append(helmOpts, executables.WithInsecure())
	}
	deps, err := factory.
		WithManifestReader().
		WithStorageClient(c.insecure).
		WithHelm(helmOpts...).
		Build(ctx)
	if err != nil {
		return err
	}
	defer deps.Close(ctx)

	dockerClient := executables.BuildDockerExecutable()
	downloadFolder := "tmp-eks-a-artifacts-download"
	imagesFile := filepath.Join(downloadFolder, imagesTarFile)
	eksaToolsImageFile := filepath.Join(downloadFolder, eksaToolsImageTarFile)

	downloadArtifacts := artifacts.Download{
		Reader: fetchReader(deps.ManifestReader, deps.StorageClient, c.includePackages),
		BundlesImagesDownloader: docker.NewImageMover(
			docker.NewOriginalRegistrySource(dockerClient),
			docker.NewDiskDestination(dockerClient, imagesFile),
		),
		EksaToolsImageDownloader: docker.NewImageMover(
			docker.NewOriginalRegistrySource(dockerClient),
			docker.NewDiskDestination(dockerClient, eksaToolsImageFile),
		),
		ChartDownloader:    helm.NewChartRegistryDownloader(deps.Helm, downloadFolder),
		Version:            version.Get(),
		TmpDowloadFolder:   downloadFolder,
		DstFile:            c.outputFile,
		Packager:           packagerForFile(c.outputFile),
		ManifestDownloader: oras.NewBundleDownloader(downloadFolder),
	}

	return downloadArtifacts.Run(ctx)
}

type packager interface {
	UnPackage(orgFile, dstFolder string) error
	Package(sourceFolder, dstFile string) error
}

func packagerForFile(file string) packager {
	if strings.HasSuffix(file, ".tar.gz") {
		return tar.NewGzipPackager()
	} else {
		return tar.NewPackager()
	}
}

func fetchReader(reader *manifests.Reader, storageClient registry.StorageClient, includePackages bool) artifacts.Reader {
	if includePackages {
		return curatedpackages.NewPackageReader(reader, storageClient)
	}
	return reader
}

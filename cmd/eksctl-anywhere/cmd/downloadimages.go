package cmd

import (
	"context"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/oras"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/helm"
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

	downloadImagesCmd.Flags().BoolVar(&downloadImagesRunner.includePackages, "include-packages", false, "this flag no longer works, use copy packages instead")
	downloadImagesCmd.Flag("include-packages").Deprecated = "use copy packages command"
	downloadImagesCmd.Flags().StringVarP(&downloadImagesRunner.bundlesOverride, "bundles-override", "", "", "Override default Bundles manifest (not recommended)")
	downloadImagesCmd.Flags().BoolVar(&downloadImagesRunner.insecure, "insecure", false, "Flag to indicate skipping TLS verification while downloading helm charts")
}

var downloadImagesRunner = downloadImagesCommand{}

type downloadImagesCommand struct {
	outputFile      string
	bundlesOverride string
	includePackages bool
	insecure        bool
}

func (c downloadImagesCommand) Run(ctx context.Context) error {
	factory := dependencies.NewFactory()
	helmOpts := []helm.Opt{}
	if c.insecure {
		helmOpts = append(helmOpts, helm.WithInsecure())
	}
	deps, err := factory.
		WithFileReader().
		WithManifestReader().
		WithHelm(helmOpts...).
		WithLogger().
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
		Reader:     deps.ManifestReader,
		FileReader: deps.FileReader,
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
		ManifestDownloader: oras.NewBundleDownloader(deps.Logger, downloadFolder),
		BundlesOverride:    c.bundlesOverride,
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

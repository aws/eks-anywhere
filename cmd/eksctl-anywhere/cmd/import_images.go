/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/oras"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
)

// imagesCmd represents the images command
var importImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Import images and charts to a registry from a tarball",
	Long: `Import all the images and helm charts necessary for EKS Anywhere clusters into a registry.
Use this command in conjunction with download images, passing it output tarball as input to this command.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return importImagesCommand.Call(ctx)
	},
}

func init() {
	importCmd.AddCommand(importImagesCmd)

	importImagesCmd.Flags().StringVarP(&importImagesCommand.InputFile, "input", "i", "", "Input tarball containing all images and charts to import")
	if err := importImagesCmd.MarkFlagRequired("input"); err != nil {
		log.Fatalf("Cannot mark 'input' as required: %s", err)
	}
	importImagesCmd.Flags().StringVarP(&importImagesCommand.RegistryEndpoint, "registry", "r", "", "Registry where to import images and charts")
	if err := importImagesCmd.MarkFlagRequired("registry"); err != nil {
		log.Fatalf("Cannot mark 'registry' as required: %s", err)
	}
	importImagesCmd.Flags().StringVarP(&importImagesCommand.BundlesFile, "bundles", "b", "", "Bundles file to read artifact dependencies from")
	if err := importImagesCmd.MarkFlagRequired("bundles"); err != nil {
		log.Fatalf("Cannot mark 'bundles' as required: %s", err)
	}
	importImagesCmd.Flags().BoolVar(&importImagesCommand.includePackages, "include-packages", false, "Flag to indicate inclusion of curated packages in imported images")
}

var importImagesCommand = ImportImagesCommand{}

type ImportImagesCommand struct {
	InputFile        string
	RegistryEndpoint string
	BundlesFile      string
	includePackages  bool
}

func (c ImportImagesCommand) Call(ctx context.Context) error {
	if !features.IsActive(features.CuratedPackagesSupport()) && c.includePackages {
		return fmt.Errorf("curated packages installation is not supported in this release")
	}

	username, password, err := config.ReadCredentials()
	if err != nil {
		return err
	}

	factory := dependencies.NewFactory()
	deps, err := factory.
		WithManifestReader().
		Build(ctx)
	if err != nil {
		return err
	}

	bundle, err := bundles.Read(deps.ManifestReader, c.BundlesFile)
	if err != nil {
		return err
	}

	artifactsFolder := "tmp-eks-a-artifacts"
	dockerClient := executables.BuildDockerExecutable()
	toolsImageFile := filepath.Join(artifactsFolder, eksaToolsImageTarFile)

	// Import the eksa tools image into the registry first, so it can be used immediately
	// after to build the helm executable
	importToolsImage := artifacts.ImportToolsImage{
		Bundles:            bundle,
		InputFile:          c.InputFile,
		TmpArtifactsFolder: artifactsFolder,
		UnPackager:         packagerForFile(c.InputFile),
		ImageMover: docker.NewImageMover(
			docker.NewDiskSource(dockerClient, toolsImageFile),
			docker.NewRegistryDestination(dockerClient, c.RegistryEndpoint),
		),
	}

	if err = importToolsImage.Run(ctx); err != nil {
		return err
	}

	deps, err = factory.
		WithRegistryMirror(c.RegistryEndpoint).
		UseExecutableImage(bundle.DefaultEksAToolsImage().VersionedImage()).
		WithHelm().
		Build(ctx)
	if err != nil {
		return err
	}
	defer deps.Close(ctx)

	imagesFile := filepath.Join(artifactsFolder, "images.tar")
	importArtifacts := artifacts.Import{
		Reader:  fetchReader(deps.ManifestReader, c.includePackages),
		Bundles: bundle,
		ImageMover: docker.NewImageMover(
			docker.NewDiskSource(dockerClient, imagesFile),
			docker.NewRegistryDestination(dockerClient, c.RegistryEndpoint),
		),
		ChartImporter: helm.NewChartRegistryImporter(
			deps.Helm, artifactsFolder,
			c.RegistryEndpoint,
			username,
			password,
		),
		TmpArtifactsFolder: artifactsFolder,
		FileImporter:       fetchFileRegistry(c.RegistryEndpoint, username, password, artifactsFolder, c.includePackages),
	}

	return importArtifacts.Run(ctx)
}

func fetchFileRegistry(registryEndpoint, username, password, artifactsFolder string, includePackages bool) artifacts.FileImporter {
	if includePackages {
		return oras.NewFileRegistryImporter(registryEndpoint, username, password, artifactsFolder)
	}
	return &artifacts.Noop{}
}

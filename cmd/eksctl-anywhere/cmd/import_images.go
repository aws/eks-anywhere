/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts"
	"github.com/aws/eks-anywhere/pkg/bundles"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/tar"
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
}

var importImagesCommand = ImportImagesCommand{}

type ImportImagesCommand struct {
	InputFile        string
	RegistryEndpoint string
	BundlesFile      string
}

func (c ImportImagesCommand) Call(ctx context.Context) error {
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
		UnPackager:         tar.NewPackager(),
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

	username, password, err := readRegistryCredentials()
	if err != nil {
		return err
	}

	imagesFile := filepath.Join(artifactsFolder, "images.tar")
	importArtifacts := artifacts.Import{
		Reader:  deps.ManifestReader,
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
	}

	return importArtifacts.Run(ctx)
}

func readRegistryCredentials() (username, password string, err error) {
	username, ok := os.LookupEnv("REGISTRY_USERNAME")
	if !ok {
		return "", "", errors.New("please set REGISTRY_USERNAME env var")
	}

	password, ok = os.LookupEnv("REGISTRY_PASSWORD")
	if !ok {
		return "", "", errors.New("please set REGISTRY_PASSWORD env var")
	}

	return username, password, nil
}

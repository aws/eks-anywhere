/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/version"
)

// imagesCmd represents the images command
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
}

var downloadImagesRunner = downloadImagesCommand{}

type downloadImagesCommand struct {
	outputFile string
}

func (c downloadImagesCommand) Run(ctx context.Context) error {
	deps, err := dependencies.NewFactory().
		WithManifestReader().
		Build(ctx)
	if err != nil {
		return err
	}
	defer deps.Close(ctx)

	dockerClient := executables.BuildDockerExecutable()
	mover := docker.NewImageMover(
		docker.NewOriginalRegistrySource(dockerClient),
		docker.NewDiskDestination(dockerClient, c.outputFile),
	)
	moveImages := commands.MoveImages{
		Reader:  deps.ManifestReader,
		Mover:   mover,
		Version: version.Get(),
	}

	return moveImages.Run(ctx)
}

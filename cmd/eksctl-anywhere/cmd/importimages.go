package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/version"
)

type importImagesOptions struct {
	fileName         string
	registryEndpoint string
}

var opts = &importImagesOptions{}

func init() {
	rootCmd.AddCommand(importImagesCmd)
	importImagesCmd.Flags().StringVarP(&opts.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	importImagesCmd.Flags().StringVarP(&opts.registryEndpoint, "endpoint", "e", "", "Local registry endpoint (host name and port)")
	err := importImagesCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking filename flag as required: %v", err)
	}
	err = importImagesCmd.MarkFlagRequired("endpoint")
	if err != nil {
		log.Fatalf("Error marking endpoint flag as required: %v", err)
	}
}

var importImagesCmd = &cobra.Command{
	Use:          "import-images",
	Short:        "Push EKS Anywhere images to a private registry",
	Long:         "This command is used to import images from an EKS Anywhere release bundle into a private registry",
	PreRunE:      preRunImportImagesCmd,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := importImages(cmd.Context(), opts.fileName, opts.registryEndpoint); err != nil {
			return err
		}
		return nil
	},
}

func importImages(context context.Context, spec string, endpoint string) error {
	clusterSpec, err := cluster.NewSpec(spec, version.Get())
	if err != nil {
		return err
	}
	de := executables.BuildDockerExecutable()

	bundle := clusterSpec.VersionsBundle

	for _, image := range bundle.Images() {
		if err := importImage(context, de, image.URI, endpoint); err != nil {
			return fmt.Errorf("error importig image %s: %v", image.URI, err)
		}
	}

	return nil
}

func importImage(ctx context.Context, docker *executables.Docker, image string, endpoint string) error {
	if err := docker.Pull(ctx, image); err != nil {
		return err
	}

	if err := docker.TagImage(ctx, image, endpoint); err != nil {
		return err
	}

	return docker.PushImage(ctx, image, endpoint)
}

func preRunImportImagesCmd(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

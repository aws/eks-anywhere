package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/executables"
)

type importImagesOptions struct {
	clusterOptions
}

var iio = &importImagesOptions{}

func init() {
	rootCmd.AddCommand(importImagesCmd)
	importImagesCmd.Flags().StringVarP(&iio.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	importImagesCmd.Flags().StringVar(&iio.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	err := importImagesCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking filename flag as required: %v", err)
	}
}

var importImagesCmd = &cobra.Command{
	Use:          "import-images",
	Short:        "Push EKS Anywhere images to a private registry",
	Long:         "This command is used to import images from an EKS Anywhere release bundle into a private registry",
	PreRunE:      preRunImportImagesCmd,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := importImages(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func importImages(context context.Context) error {
	clusterSpec, err := newClusterSpec(iio.clusterOptions)
	if err != nil {
		return err
	}
	de := executables.BuildDockerExecutable()

	if clusterSpec.Spec.RegistryMirrorConfiguration == nil || clusterSpec.Spec.RegistryMirrorConfiguration.Endpoint == "" {
		return fmt.Errorf("it is necessary to define a valid endpoint in your spec (registryMirrorConfiguration.endpoint)")
	}
	endpoint := clusterSpec.Spec.RegistryMirrorConfiguration.Endpoint

	images, err := getImages(clusterSpec)
	if err != nil {
		return err
	}
	for _, image := range images {
		if err := importImage(context, de, image.URI, endpoint); err != nil {
			return fmt.Errorf("error importing image %s: %v", image.URI, err)
		}
	}
	return nil
}

func importImage(ctx context.Context, docker *executables.Docker, image string, endpoint string) error {
	if err := docker.PullImage(ctx, image); err != nil {
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

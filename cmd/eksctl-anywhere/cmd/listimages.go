package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type listImagesOptions struct {
	clusterOptions
}

var lio = &listImagesOptions{}

func init() {
	listCmd.AddCommand(listImagesCommand)
	listImagesCommand.Flags().StringVarP(&lio.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	listImagesCommand.Flags().StringVar(&lio.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	err := listImagesCommand.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking filename flag as required: %v", err)
	}
}

var listImagesCommand = &cobra.Command{
	Use:   "images",
	Short: "Generate a list of images used by EKS Anywhere",
	Long:  "This command is used to generate a list of images used by EKS-Anywhere for cluster provisioning",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if err := viper.BindPFlag(flag.Name, flag); err != nil {
				log.Fatalf("Error initializing flags: %v", err)
			}
		})
		return nil
	},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listImages()
	},
}

func listImages() error {
	clusterSpec, err := newClusterSpec(lio.clusterOptions)
	if err != nil {
		return err
	}

	images, err := getImages(clusterSpec)
	if err != nil {
		return err
	}

	for _, image := range images {
		if image.ImageDigest != "" {
			fmt.Printf("%s@%s\n", image.URI, image.ImageDigest)
		} else {
			fmt.Printf("%s\n", image.URI)
		}
	}

	return nil
}

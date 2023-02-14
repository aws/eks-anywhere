package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/version"
)

type checkImagesOptions struct {
	fileName string
}

var cio = &checkImagesOptions{}

func init() {
	rootCmd.AddCommand(checkImagesCommand)
	checkImagesCommand.Flags().StringVarP(&cio.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	err := checkImagesCommand.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking filename flag as required: %v", err)
	}
}

var checkImagesCommand = &cobra.Command{
	Use:   "check-images",
	Short: "Check images used by EKS Anywhere do exist in the target registry",
	Long:  "This command is used to check images used by EKS-Anywhere for cluster provisioning do exist in the target registry",
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
		return checkImages(cmd.Context(), cio.fileName)
	},
}

func checkImages(context context.Context, clusterSpecPath string) error {
	images, err := getImages(clusterSpecPath, "")
	if err != nil {
		return err
	}

	clusterSpec, err := readAndValidateClusterSpec(clusterSpecPath, version.Get())
	if err != nil {
		return err
	}

	checkImageExistence := artifacts.CheckImageExistence{}
	for _, image := range images {
		myImageURI := registrymirror.FromCluster(clusterSpec.Cluster).ReplaceRegistry(image.URI)
		checkImageExistence.ImageUri = myImageURI
		if err = checkImageExistence.Run(context); err != nil {
			fmt.Println(err.Error())
			logger.MarkFail(myImageURI)
		} else {
			logger.MarkPass(myImageURI)
		}
	}

	return nil
}

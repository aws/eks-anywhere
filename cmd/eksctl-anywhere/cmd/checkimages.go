package cmd

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/logger"
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

func checkImages(context context.Context, spec string) error {
	images, err := getImages(spec)
	if err != nil {
		return err
	}

	clusterSpec, err := readAndValidateClusterSpec(spec, version.Get())
	if err != nil {
		return err
	}

	myRegistry := constants.DefaultRegistry
	namespace := ""

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		host := clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint
		if len(host) > 0 {
			port := clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Port
			if port == "" {
				port = constants.DefaultHttpsPort
			}
			myRegistry = net.JoinHostPort(host, port)
			namespace = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Namespace
		}
	}

	checkImageExistence := artifacts.CheckImageExistence{}
	for _, image := range images {
		myImageUri := docker.ReplaceHostWithNamespacedEndpoint(image.URI, myRegistry, namespace)
		checkImageExistence.ImageUri = myImageUri
		if err = checkImageExistence.Run(context); err != nil {
			fmt.Println(err.Error())
			logger.MarkFail(myImageUri)
		} else {
			logger.MarkPass(myImageUri)
		}
	}

	return nil
}

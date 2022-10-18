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
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type checkImagesOptions struct {
	fileName string
	insecure bool
}

var cio = &checkImagesOptions{}

func init() {
	rootCmd.AddCommand(checkImagesCommand)
	checkImagesCommand.Flags().StringVarP(&cio.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	err := checkImagesCommand.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking filename flag as required: %v", err)
	}
	checkImagesCommand.Flags().BoolVar(&cio.insecure, "insecure", false, "Flag to indicate skipping TLS verification while downloading helm charts")
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
		return checkImages(cmd.Context(), cio)
	},
}

func checkImages(context context.Context, options *checkImagesOptions) error {
	images, err := getImages(cio.fileName)
	if err != nil {
		return err
	}

	clusterSpec, err := readAndValidateClusterSpec(cio.fileName, version.Get())
	if err != nil {
		return err
	}

	myRegistry := constants.DefaultRegistry
	ociNamespace := ""
	packageOCINamespace := ""

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		host := clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint
		if len(host) > 0 {
			port := clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Port
			if port == "" {
				port = constants.DefaultHTTPSPort
			}
			myRegistry = net.JoinHostPort(host, port)
			ociNamespace = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.OCINamespace
			packageOCINamespace = clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.PackageOCINamespace
		}
	}

	factory := dependencies.NewFactory()
	helmOpts := []executables.HelmOpt{}
	if cio.insecure {
		helmOpts = append(helmOpts, executables.WithInsecure())
	}
	deps, err := factory.
		WithManifestReader().
		WithRegistryMirror(myRegistry, ociNamespace, packageOCINamespace, false).
		WithHelm(helmOpts...).
		Build(context)
	if err != nil {
		return err
	}
	defer deps.Close(context)

	reader := curatedpackages.NewPackageReader(deps.ManifestReader)
	bundle, err := reader.ReadBundlesForVersion(version.Get().GitVersion)
	if err != nil {
		return err
	}
	packageImages, err := reader.ReadPackageImagesFromBundles(context, bundle)
	if err != nil {
		return err
	}
	packageImages = append(packageImages, reader.ReadPackageChartsFromBundles(context, bundle)...)
	packageImageSet := buildPackageImageNamesSet(packageImages)

	checkImageExistence := artifacts.CheckImageExistence{}
	for _, image := range images {
		myImageURI := docker.ReplaceHostWithNamespacedEndpoint(image.URI, myRegistry, ociNamespace)
		if _, ok := packageImageSet[image.URI]; ok {
			myImageURI = docker.ReplaceHostWithNamespacedEndpoint(image.URI, myRegistry, packageOCINamespace)
		}
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

func buildPackageImageNamesSet(packageImages []v1alpha1.Image) map[string]bool {
	set := make(map[string]bool)
	for _, image := range packageImages {
		set[image.URI] = true
	}
	return set
}

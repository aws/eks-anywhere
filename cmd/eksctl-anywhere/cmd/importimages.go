package cmd

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type importImagesOptions struct {
	fileName string
}

var opts = &importImagesOptions{}

const ociPrefix = "oci://"

func init() {
	rootCmd.AddCommand(importImagesCmd)
	importImagesCmd.Flags().StringVarP(&opts.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
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
		if err := importImages(cmd.Context(), opts.fileName); err != nil {
			return err
		}
		return nil
	},
}

func importImages(ctx context.Context, spec string) error {
	clusterSpec, err := cluster.NewSpecFromClusterConfig(spec, version.Get())
	if err != nil {
		return err
	}

	de := executables.BuildDockerExecutable()

	bundle := clusterSpec.VersionsBundle
	executableBuilder, closer, err := executables.NewExecutableBuilder(ctx, bundle.Eksa.CliTools.VersionedImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	defer closer.CheckErr(ctx)
	helmExecutable := executableBuilder.BuildHelmExecutable()

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration == nil || clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint == "" {
		return fmt.Errorf("endpoint not set. It is necessary to define a valid endpoint in your spec (registryMirrorConfiguration.endpoint)")
	}
	host := clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint
	port := clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Port
	if port == "" {
		logger.V(1).Info("RegistryMirrorConfiguration.Port is not specified, default port will be used", "Default Port", constants.DefaultHttpsPort)
		port = constants.DefaultHttpsPort
	}
	if !networkutils.IsPortValid(clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Port) {
		return fmt.Errorf("registry mirror port %s is invalid, please provide a valid port", clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Port)
	}

	images, err := getImages(spec)
	if err != nil {
		return err
	}
	for _, image := range images {
		if err := importImage(ctx, de, image.URI, net.JoinHostPort(host, port)); err != nil {
			return fmt.Errorf("importing image %s: %v", image.URI, err)
		}
	}

	endpoint := clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint
	return importCharts(ctx, helmExecutable, bundle.Charts(), endpoint)
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

func importCharts(ctx context.Context, helm *executables.Helm, charts map[string]*v1alpha1.Image, endpoint string) error {
	for _, chart := range charts {
		if err := importChart(ctx, helm, *chart, endpoint); err != nil {
			return err
		}
	}
	return nil
}

func importChart(ctx context.Context, helm *executables.Helm, chart v1alpha1.Image, endpoint string) error {
	uri, chartVersion := getChartUriAndVersion(chart)
	if err := helm.PullChart(ctx, uri, chartVersion); err != nil {
		return err
	}
	return helm.PushChart(ctx, chart.ChartName(), fmt.Sprintf("%s%s/%s", ociPrefix, endpoint, chart.Name))
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

func getChartUriAndVersion(chart v1alpha1.Image) (uri, version string) {
	uri = fmt.Sprintf("%s%s", ociPrefix, chart.Image())
	version = chart.Tag()
	return uri, version
}

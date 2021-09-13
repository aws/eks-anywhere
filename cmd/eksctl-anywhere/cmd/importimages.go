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

	if err := importImage(context, de, bundle.Aws.Controller.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Aws.Controller.URI, err)
	}
	if err := importImage(context, de, bundle.Aws.KubeProxy.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Aws.KubeProxy.URI, err)
	}

	if err := importImage(context, de, bundle.Bootstrap.Controller.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Bootstrap.Controller.URI, err)
	}
	if err := importImage(context, de, bundle.Bootstrap.KubeProxy.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Bootstrap.KubeProxy.URI, err)
	}

	if err := importImage(context, de, bundle.BottleRocketBootstrap.Bootstrap.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.BottleRocketBootstrap.Bootstrap.URI, err)
	}

	if err := importImage(context, de, bundle.CertManager.Acmesolver.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.CertManager.Acmesolver.URI, err)
	}
	if err := importImage(context, de, bundle.CertManager.Cainjector.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.CertManager.Cainjector.URI, err)
	}
	if err := importImage(context, de, bundle.CertManager.Controller.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.CertManager.Controller.URI, err)
	}
	if err := importImage(context, de, bundle.CertManager.Webhook.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.CertManager.Webhook.URI, err)
	}

	if err := importImage(context, de, bundle.Cilium.Cilium.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Cilium.Cilium.URI, err)
	}
	if err := importImage(context, de, bundle.Cilium.Operator.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Cilium.Operator.URI, err)
	}

	if err := importImage(context, de, bundle.ClusterAPI.Controller.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.ClusterAPI.Controller.URI, err)
	}
	if err := importImage(context, de, bundle.ClusterAPI.KubeProxy.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.ClusterAPI.KubeProxy.URI, err)
	}

	if err := importImage(context, de, bundle.ControlPlane.Controller.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.ControlPlane.Controller.URI, err)
	}
	if err := importImage(context, de, bundle.ControlPlane.KubeProxy.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.ControlPlane.KubeProxy.URI, err)
	}

	if err := importImage(context, de, bundle.Docker.KubeProxy.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Docker.KubeProxy.URI, err)
	}
	if err := importImage(context, de, bundle.Docker.Manager.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Docker.Manager.URI, err)
	}

	if err := importImage(context, de, bundle.EksD.KindNode.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.EksD.KindNode.URI, err)
	}

	if err := importImage(context, de, bundle.Eksa.CliTools.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Eksa.CliTools.URI, err)
	}

	if err := importImage(context, de, bundle.Flux.HelmController.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Flux.HelmController.URI, err)
	}
	if err := importImage(context, de, bundle.Flux.KustomizeController.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Flux.KustomizeController.URI, err)
	}
	if err := importImage(context, de, bundle.Flux.NotificationController.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Flux.NotificationController.URI, err)
	}
	if err := importImage(context, de, bundle.Flux.SourceController.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.Flux.SourceController.URI, err)
	}

	if err := importImage(context, de, bundle.VSphere.ClusterAPIController.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.VSphere.ClusterAPIController.URI, err)
	}
	if err := importImage(context, de, bundle.VSphere.Driver.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.VSphere.Driver.URI, err)
	}
	if err := importImage(context, de, bundle.VSphere.KubeProxy.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.VSphere.KubeProxy.URI, err)
	}
	if err := importImage(context, de, bundle.VSphere.KubeVip.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.VSphere.KubeVip.URI, err)
	}
	if err := importImage(context, de, bundle.VSphere.Manager.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.VSphere.Manager.URI, err)
	}
	if err := importImage(context, de, bundle.VSphere.Syncer.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.VSphere.Syncer.URI, err)
	}

	if err := importImage(context, de, bundle.ExternalEtcdBootstrap.Controller.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.ExternalEtcdBootstrap.Controller.URI, err)
	}
	if err := importImage(context, de, bundle.ExternalEtcdBootstrap.KubeProxy.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.ExternalEtcdBootstrap.KubeProxy.URI, err)
	}

	if err := importImage(context, de, bundle.ExternalEtcdController.KubeProxy.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.ExternalEtcdController.KubeProxy.URI, err)
	}
	if err := importImage(context, de, bundle.ExternalEtcdController.Controller.URI, endpoint); err != nil {
		return fmt.Errorf("error importig image %s: %v", bundle.ExternalEtcdController.Controller.URI, err)
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

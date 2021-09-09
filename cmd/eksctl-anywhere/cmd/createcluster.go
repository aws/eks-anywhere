package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/addonmanager/addonclients"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	fluxclient "github.com/aws/eks-anywhere/pkg/clients/flux"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networking"
	"github.com/aws/eks-anywhere/pkg/providers/factory"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type createClusterOptions struct {
	fileName    string
	forceClean  bool
	skipIpCheck bool
}

var cc = &createClusterOptions{}

var createClusterCmd = &cobra.Command{
	Use:          "cluster",
	Short:        "Create workload cluster",
	Long:         "This command is used to create workload clusters",
	PreRunE:      preRunCreateCluster,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cc.validate(cmd.Context()); err != nil {
			return err
		}
		if err := cc.createCluster(cmd.Context()); err != nil {
			return fmt.Errorf("failed to create cluster: %v", err)
		}
		return nil
	},
}

func init() {
	createCmd.AddCommand(createClusterCmd)
	createClusterCmd.Flags().StringVarP(&cc.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	createClusterCmd.Flags().BoolVar(&cc.forceClean, "force-cleanup", false, "Force deletion of previously created bootstrap cluster")
	createClusterCmd.Flags().BoolVar(&cc.skipIpCheck, "skip-ip-check", false, "Skip check for whether cluster control plane ip is in use")
	err := createClusterCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func preRunCreateCluster(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

type clusterManagerClient struct {
	*executables.Clusterctl
	*executables.Kubectl
}

type bootstrapperClient struct {
	*executables.Kind
	*executables.Kubectl
}

func (cc *createClusterOptions) validate(ctx context.Context) error {
	clusterConfig, err := commonValidation(ctx, cc.fileName)
	if err != nil {
		return err
	}
	if validations.KubeConfigExists(clusterConfig.Name, clusterConfig.Name, "", kubeconfigPattern) {
		return fmt.Errorf("old cluster config file exists under %s, please use a different clusterName to proceed", clusterConfig.Name)
	}
	return nil
}

func (cc *createClusterOptions) createCluster(ctx context.Context) error {
	clusterSpec, err := cluster.NewSpec(cc.fileName, version.Get())
	if err != nil {
		return fmt.Errorf("unable to get cluster config from file: %v", err)
	}
	if clusterSpec.HasOverrideClusterSpecFile() {
		logger.Info("Warning: Override Cluster Spec file is configured. All other values in EKS-A spec will be ignored.")
	}

	writer, err := filewriter.NewWriter(clusterSpec.Name)
	if err != nil {
		return fmt.Errorf("unable to write: %v", err)
	}
	eksaToolsImage := clusterSpec.VersionsBundle.Eksa.CliTools
	image := eksaToolsImage.VersionedImage()
	executableBuilder, err := executables.NewExecutableBuilder(ctx, image)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	clusterawsadm := executableBuilder.BuildClusterAwsAdmExecutable()
	kind := executableBuilder.BuildKindExecutable(writer)
	clusterctl := executableBuilder.BuildClusterCtlExecutable(writer)
	kubectl := executableBuilder.BuildKubectlExecutable()
	govc := executableBuilder.BuildGovcExecutable(writer)
	docker := executables.BuildDockerExecutable()
	flux := executableBuilder.BuildFluxExecutable()

	providerFactory := &factory.ProviderFactory{
		AwsClient:            clusterawsadm,
		DockerClient:         docker,
		DockerKubectlClient:  kubectl,
		VSphereGovcClient:    govc,
		VSphereKubectlClient: kubectl,
		Writer:               writer,
		SkipIpCheck:          cc.skipIpCheck,
	}
	provider, err := providerFactory.BuildProvider(cc.fileName, clusterSpec.Cluster)
	if err != nil {
		return err
	}

	bootstrapper := bootstrapper.New(&bootstrapperClient{kind, kubectl})

	clusterManager := clustermanager.New(
		&clusterManagerClient{
			clusterctl,
			kubectl,
		},
		networking.NewCilium(),
		writer,
	)

	gitOpts, err := addonclients.NewGitOptions(ctx, clusterSpec.Cluster, clusterSpec.GitOpsConfig, writer)
	if err != nil {
		return err
	}
	addonClient := addonclients.NewFluxAddonClient(
		&fluxclient.FluxKubectl{
			Flux:    flux,
			Kubectl: kubectl,
		},
		gitOpts,
	)

	createCluster := workflows.NewCreate(
		bootstrapper,
		provider,
		clusterManager,
		addonClient,
		writer,
	)
	err = createCluster.Run(ctx, clusterSpec, cc.forceClean)
	if err == nil {
		writer.CleanUpTemp()
	}
	return err
}

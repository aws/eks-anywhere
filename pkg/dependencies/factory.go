package dependencies

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/aws/eks-anywhere/pkg/addonmanager/addonclients"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/clients/flux"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/networking/kindnetd"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/providers/factory"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/pbnj"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Dependencies struct {
	Provider                  providers.Provider
	ClusterAwsCli             *executables.Clusterawsadm
	DockerClient              *executables.Docker
	Kubectl                   *executables.Kubectl
	Govc                      *executables.Govc
	Cmk                       *executables.Cmk
	Tink                      *executables.Tink
	Pbnj                      *pbnj.Pbnj
	TinkerbellClients         tinkerbell.TinkerbellClients
	Writer                    filewriter.FileWriter
	Kind                      *executables.Kind
	Clusterctl                *executables.Clusterctl
	Flux                      *executables.Flux
	Troubleshoot              *executables.Troubleshoot
	Helm                      *executables.Helm
	Networking                clustermanager.Networking
	AwsIamAuth                clustermanager.AwsIamAuth
	ClusterManager            *clustermanager.ClusterManager
	Bootstrapper              *bootstrapper.Bootstrapper
	FluxAddonClient           *addonclients.FluxAddonClient
	AnalyzerFactory           diagnostics.AnalyzerFactory
	CollectorFactory          diagnostics.CollectorFactory
	DignosticCollectorFactory diagnostics.DiagnosticBundleFactory
	CAPIManager               *clusterapi.Manager
	ResourceSetManager        *clusterapi.ResourceSetManager
	closers                   []types.Closer
}

func (d *Dependencies) Close(ctx context.Context) error {
	// Reverse the loop so we close like LIFO
	for i := len(d.closers) - 1; i >= 0; i-- {
		if err := d.closers[i].Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

func ForSpec(ctx context.Context, clusterSpec *cluster.Spec) *Factory {
	eksaToolsImage := clusterSpec.VersionsBundle.Eksa.CliTools
	return NewFactory().
		WithExecutableImage(clusterSpec.Cluster.UseImageMirror(eksaToolsImage.VersionedImage())).
		WithWriterFolder(clusterSpec.Cluster.Name).
		WithDiagnosticCollectorImage(clusterSpec.VersionsBundle.Eksa.DiagnosticCollector.VersionedImage())
}

type Factory struct {
	executableBuilder        *executables.ExecutableBuilder
	providerFactory          *factory.ProviderFactory
	executablesImage         string
	executablesMountDirs     []string
	writerFolder             string
	diagnosticCollectorImage string
	buildSteps               []buildStep
	dependencies             Dependencies
}

type buildStep func(ctx context.Context) error

func NewFactory() *Factory {
	return &Factory{
		writerFolder: "./",
		buildSteps:   make([]buildStep, 0),
	}
}

func (f *Factory) Build(ctx context.Context) (*Dependencies, error) {
	for _, step := range f.buildSteps {
		if err := step(ctx); err != nil {
			return nil, err
		}
	}

	// clean up stack
	f.buildSteps = make([]buildStep, 0)

	// Make copy of dependencies since its attributes are public
	d := f.dependencies

	return &d, nil
}

func (f *Factory) WithWriterFolder(folder string) *Factory {
	f.writerFolder = folder
	return f
}

func (f *Factory) WithExecutableImage(image string) *Factory {
	f.executablesImage = image
	return f
}

func (f *Factory) WithExecutableMountDirs(mountDirs ...string) *Factory {
	f.executablesMountDirs = mountDirs
	return f
}

func (f *Factory) WithExecutableBuilder() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.executableBuilder != nil {
			return nil
		}

		b, close, err := executables.NewExecutableBuilder(ctx, f.executablesImage, f.executablesMountDirs...)
		if err != nil {
			return err
		}

		f.dependencies.closers = append(f.dependencies.closers, close)

		f.executableBuilder = b
		return nil
	})

	return f
}

func (f *Factory) WithProvider(clusterConfigFile string, clusterConfig *v1alpha1.Cluster, skipIpCheck bool, hardwareConfigFile string, skipPowerActions bool) *Factory {
	f.WithProviderFactory(clusterConfigFile, clusterConfig)
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Provider != nil {
			return nil
		}

		var err error
		f.dependencies.Provider, err = f.providerFactory.BuildProvider(clusterConfigFile, clusterConfig, skipIpCheck, hardwareConfigFile, skipPowerActions)
		if err != nil {
			return err
		}

		return nil
	})

	return f
}

func (f *Factory) WithProviderFactory(clusterConfigFile string, clusterConfig *v1alpha1.Cluster) *Factory {
	switch clusterConfig.Spec.DatacenterRef.Kind {
	case v1alpha1.VSphereDatacenterKind:
		f.WithKubectl().WithGovc().WithWriter().WithCAPIClusterResourceSetManager()
	case v1alpha1.CloudStackDatacenterKind:
		f.WithKubectl().WithCmk().WithWriter()
	case v1alpha1.DockerDatacenterKind:
		f.WithDocker().WithKubectl()
	case v1alpha1.TinkerbellDatacenterKind:
		f.WithKubectl().WithTink(clusterConfigFile).WithPbnj(clusterConfigFile)
	}

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.providerFactory != nil {
			return nil
		}

		f.providerFactory = &factory.ProviderFactory{
			DockerClient:              f.dependencies.DockerClient,
			DockerKubectlClient:       f.dependencies.Kubectl,
			CloudStackCmkClient:       f.dependencies.Cmk,
			CloudStackKubectlClient:   f.dependencies.Kubectl,
			VSphereGovcClient:         f.dependencies.Govc,
			VSphereKubectlClient:      f.dependencies.Kubectl,
			SnowKubectlClient:         f.dependencies.Kubectl,
			TinkerbellKubectlClient:   f.dependencies.Kubectl,
			TinkerbellClients:         tinkerbell.TinkerbellClients{ProviderTinkClient: f.dependencies.Tink, ProviderPbnjClient: f.dependencies.Pbnj},
			Writer:                    f.dependencies.Writer,
			ClusterResourceSetManager: f.dependencies.ResourceSetManager,
		}

		return nil
	})

	return f
}

func (f *Factory) WithClusterAwsCli() *Factory {
	f.WithExecutableBuilder()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.ClusterAwsCli != nil {
			return nil
		}

		f.dependencies.ClusterAwsCli = f.executableBuilder.BuildClusterAwsAdmExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithDocker() *Factory {
	f.WithExecutableBuilder()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.DockerClient != nil {
			return nil
		}

		f.dependencies.DockerClient = executables.BuildDockerExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithKubectl() *Factory {
	f.WithExecutableBuilder()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Kubectl != nil {
			return nil
		}

		f.dependencies.Kubectl = f.executableBuilder.BuildKubectlExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithGovc() *Factory {
	f.WithExecutableBuilder().WithWriter()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Govc != nil {
			return nil
		}

		f.dependencies.Govc = f.executableBuilder.BuildGovcExecutable(f.dependencies.Writer)
		f.dependencies.closers = append(f.dependencies.closers, f.dependencies.Govc)

		return nil
	})

	return f
}

func (f *Factory) WithCmk() *Factory {
	f.WithExecutableBuilder().WithWriter()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Cmk != nil {
			return nil
		}
		execConfig, err := decoder.ParseCloudStackSecret()
		if err != nil {
			return fmt.Errorf("building cmk executable: %v", err)
		}

		f.dependencies.Cmk = f.executableBuilder.BuildCmkExecutable(f.dependencies.Writer, *execConfig)
		f.dependencies.closers = append(f.dependencies.closers, f.dependencies.Cmk)

		return nil
	})

	return f
}

func (f *Factory) WithTink(clusterConfigFile string) *Factory {
	f.WithExecutableBuilder()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Tink != nil {
			return nil
		}
		tinkerbellDatacenterConfig, err := v1alpha1.GetTinkerbellDatacenterConfig(clusterConfigFile)
		if err != nil {
			return err
		}
		f.dependencies.Tink = f.executableBuilder.BuildTinkExecutable(tinkerbellDatacenterConfig.Spec.TinkerbellCertURL, tinkerbellDatacenterConfig.Spec.TinkerbellGRPCAuth)

		return nil
	})

	return f
}

func (f *Factory) WithPbnj(clusterConfigFile string) *Factory {
	f.WithExecutableBuilder()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Pbnj != nil {
			return nil
		}
		tinkerbellDatacenterConfig, err := v1alpha1.GetTinkerbellDatacenterConfig(clusterConfigFile)
		if err != nil {
			return err
		}

		pbnjClient, err := pbnj.NewPBNJClient(tinkerbellDatacenterConfig.Spec.TinkerbellPBnJGRPCAuth)
		if err != nil {
			return err
		}
		f.dependencies.Pbnj = pbnjClient
		return nil
	})

	return f
}

func (f *Factory) WithWriter() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Writer != nil {
			return nil
		}

		var err error
		f.dependencies.Writer, err = filewriter.NewWriter(f.writerFolder)
		if err != nil {
			return err
		}

		return nil
	})

	return f
}

func (f *Factory) WithKind() *Factory {
	f.WithExecutableBuilder().WithWriter()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Kind != nil {
			return nil
		}

		f.dependencies.Kind = f.executableBuilder.BuildKindExecutable(f.dependencies.Writer)
		return nil
	})

	return f
}

func (f *Factory) WithClusterctl() *Factory {
	f.WithExecutableBuilder().WithWriter()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Clusterctl != nil {
			return nil
		}

		f.dependencies.Clusterctl = f.executableBuilder.BuildClusterCtlExecutable(f.dependencies.Writer)
		return nil
	})

	return f
}

func (f *Factory) WithFlux() *Factory {
	f.WithExecutableBuilder()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Flux != nil {
			return nil
		}

		f.dependencies.Flux = f.executableBuilder.BuildFluxExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithTroubleshoot() *Factory {
	f.WithExecutableBuilder()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Troubleshoot != nil {
			return nil
		}

		f.dependencies.Troubleshoot = f.executableBuilder.BuildTroubleshootExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithHelm() *Factory {
	f.WithExecutableBuilder()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Helm != nil {
			return nil
		}

		f.dependencies.Helm = f.executableBuilder.BuildHelmExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithNetworking(clusterConfig *v1alpha1.Cluster) *Factory {
	var networkingBuilder func() clustermanager.Networking
	if clusterConfig.Spec.ClusterNetwork.CNIConfig.Kindnetd != nil {
		f.WithKubectl()
		networkingBuilder = func() clustermanager.Networking {
			return kindnetd.NewKindnetd(f.dependencies.Kubectl)
		}
	} else {
		f.WithKubectl().WithHelm()
		networkingBuilder = func() clustermanager.Networking {
			return cilium.NewCilium(f.dependencies.Kubectl, f.dependencies.Helm)
		}
	}

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Networking != nil {
			return nil
		}
		f.dependencies.Networking = networkingBuilder()

		return nil
	})

	return f
}

func (f *Factory) WithAwsIamAuth() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.AwsIamAuth != nil {
			return nil
		}
		certgen := crypto.NewCertificateGenerator()
		clusterId := uuid.New()
		f.dependencies.AwsIamAuth = awsiamauth.NewAwsIamAuth(certgen, clusterId)
		return nil
	})

	return f
}

type bootstrapperClient struct {
	*executables.Kind
	*executables.Kubectl
}

func (f *Factory) WithBootstrapper() *Factory {
	f.WithKind().WithKubectl()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Bootstrapper != nil {
			return nil
		}

		f.dependencies.Bootstrapper = bootstrapper.New(&bootstrapperClient{f.dependencies.Kind, f.dependencies.Kubectl})
		return nil
	})

	return f
}

type clusterManagerClient struct {
	*executables.Clusterctl
	*executables.Kubectl
}

func (f *Factory) WithClusterManager(clusterConfig *v1alpha1.Cluster) *Factory {
	f.WithClusterctl().WithKubectl().WithNetworking(clusterConfig).WithWriter().WithDiagnosticBundleFactory().WithAwsIamAuth()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.ClusterManager != nil {
			return nil
		}

		f.dependencies.ClusterManager = clustermanager.New(
			&clusterManagerClient{
				f.dependencies.Clusterctl,
				f.dependencies.Kubectl,
			},
			f.dependencies.Networking,
			f.dependencies.Writer,
			f.dependencies.DignosticCollectorFactory,
			f.dependencies.AwsIamAuth,
		)
		return nil
	})

	return f
}

func (f *Factory) WithFluxAddonClient(ctx context.Context, clusterConfig *v1alpha1.Cluster, gitOpsConfig *v1alpha1.GitOpsConfig) *Factory {
	f.WithWriter().WithFlux().WithKubectl()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.FluxAddonClient != nil {
			return nil
		}

		gitOpts, err := addonclients.NewGitOptions(ctx, clusterConfig, gitOpsConfig, f.dependencies.Writer)
		if err != nil {
			return err
		}

		f.dependencies.FluxAddonClient = addonclients.NewFluxAddonClient(
			&flux.FluxKubectl{
				Flux:    f.dependencies.Flux,
				Kubectl: f.dependencies.Kubectl,
			},
			gitOpts,
		)

		return nil
	})

	return f
}

func (f *Factory) WithDiagnosticBundleFactory() *Factory {
	f.WithWriter().WithTroubleshoot().WithCollectorFactory().WithAnalyzerFactory().WithKubectl()
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.DignosticCollectorFactory != nil {
			return nil
		}

		opts := diagnostics.EksaDiagnosticBundleFactoryOpts{
			AnalyzerFactory:  f.dependencies.AnalyzerFactory,
			Client:           f.dependencies.Troubleshoot,
			CollectorFactory: f.dependencies.CollectorFactory,
			Kubectl:          f.dependencies.Kubectl,
			Writer:           f.dependencies.Writer,
		}

		f.dependencies.DignosticCollectorFactory = diagnostics.NewFactory(opts)
		return nil
	})

	return f
}

func (f *Factory) WithAnalyzerFactory() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.AnalyzerFactory != nil {
			return nil
		}

		f.dependencies.AnalyzerFactory = diagnostics.NewAnalyzerFactory()
		return nil
	})

	return f
}

func (f *Factory) WithDiagnosticCollectorImage(diagnosticCollectorImage string) *Factory {
	f.diagnosticCollectorImage = diagnosticCollectorImage
	return f
}

func (f *Factory) WithCollectorFactory() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.CollectorFactory != nil {
			return nil
		}

		if f.diagnosticCollectorImage == "" {
			f.dependencies.CollectorFactory = diagnostics.NewDefaultCollectorFactory()
		} else {
			f.dependencies.CollectorFactory = diagnostics.NewCollectorFactory(f.diagnosticCollectorImage)
		}
		return nil
	})

	return f
}

func (f *Factory) WithCAPIManager() *Factory {
	f.WithClusterctl()
	f.WithKubectl()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.CAPIManager != nil {
			return nil
		}

		f.dependencies.CAPIManager = clusterapi.NewManager(f.dependencies.Clusterctl, f.dependencies.Kubectl)
		return nil
	})

	return f
}

func (f *Factory) WithCAPIClusterResourceSetManager() *Factory {
	f.WithKubectl()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.ResourceSetManager != nil {
			return nil
		}

		f.dependencies.ResourceSetManager = clusterapi.NewResourceSetManager(f.dependencies.Kubectl)
		return nil
	})

	return f
}

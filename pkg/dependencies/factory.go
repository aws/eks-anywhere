package dependencies

import (
	"context"
	"errors"

	"github.com/aws/eks-anywhere/pkg/addonmanager/addonclients"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/clients/flux"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/networking"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/factory"
)

type Dependencies struct {
	Provider         providers.Provider
	ClusterAwsCli    *executables.Clusterawsadm
	DockerClient     *executables.Docker
	Kubectl          *executables.Kubectl
	Govc             *executables.Govc
	Writer           filewriter.FileWriter
	Kind             *executables.Kind
	Clusterctl       *executables.Clusterctl
	Flux             *executables.Flux
	Troubleshoot     *executables.Troubleshoot
	Networking       clustermanager.Networking
	ClusterManager   *clustermanager.ClusterManager
	Bootstrapper     *bootstrapper.Bootstrapper
	FluxAddonClient  *addonclients.FluxAddonClient
	AnalyzerFactory  diagnostics.AnalyzerFactory
	CollectorFactory diagnostics.CollectorFactory
	CAPIUpgrader     *clusterapi.Upgrader
}

func ForSpec(ctx context.Context, clusterSpec *cluster.Spec) *Factory {
	eksaToolsImage := clusterSpec.VersionsBundle.Eksa.CliTools
	return NewFactory().
		WithExecutableBuilder(ctx, clusterSpec.UseImageMirror(eksaToolsImage.VersionedImage())).
		WithWriterFolder(clusterSpec.Name).
		WithDiagnosticCollectorImage(clusterSpec.VersionsBundle.Eksa.DiagnosticCollector.VersionedImage())
}

type Factory struct {
	executableBuilder        *executables.ExecutableBuilder
	providerFactory          *factory.ProviderFactory
	writerFolder             string
	diagnosticCollectorImage string
	buildSteps               []func() error
	dependencies             Dependencies
}

func NewFactory() *Factory {
	return &Factory{
		executableBuilder: executables.NewLocalExecutableBuilder(),
		writerFolder:      "./",
		buildSteps:        make([]func() error, 0),
	}
}

func (f *Factory) Build() (*Dependencies, error) {
	for _, step := range f.buildSteps {
		if err := step(); err != nil {
			return nil, err
		}
	}

	// clean up stack
	f.buildSteps = make([]func() error, 0)

	// Make copy of dependencies since its attributes are public
	d := f.dependencies

	return &d, nil
}

func (f *Factory) WithWriterFolder(folder string) *Factory {
	f.writerFolder = folder
	return f
}

func (f *Factory) WithExecutableBuilder(ctx context.Context, image string) *Factory {
	f.buildSteps = append(f.buildSteps, func() error {
		b, err := executables.NewExecutableBuilder(ctx, image)
		if err != nil {
			return err
		}

		f.executableBuilder = b
		return nil
	})

	return f
}

func (f *Factory) WithProvider(clusterConfigFile string, clusterConfig *v1alpha1.Cluster, skipIpCheck bool) *Factory {
	f.WithProviderFactory()

	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.Provider != nil {
			return nil
		}

		var err error
		f.dependencies.Provider, err = f.providerFactory.BuildProvider(clusterConfigFile, clusterConfig, skipIpCheck)
		if err != nil {
			return err
		}

		return nil
	})

	return f
}

func (f *Factory) WithProviderFactory() *Factory {
	f.WithDocker().WithKubectl().WithGovc().WithWriter()

	f.buildSteps = append(f.buildSteps, func() error {
		if f.providerFactory != nil {
			return nil
		}

		f.providerFactory = &factory.ProviderFactory{
			DockerClient:         f.dependencies.DockerClient,
			DockerKubectlClient:  f.dependencies.Kubectl,
			VSphereGovcClient:    f.dependencies.Govc,
			VSphereKubectlClient: f.dependencies.Kubectl,
			Writer:               f.dependencies.Writer,
		}

		return nil
	})

	return f
}

func (f *Factory) WithClusterAwsCli() *Factory {
	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.ClusterAwsCli != nil {
			return nil
		}

		f.dependencies.ClusterAwsCli = f.executableBuilder.BuildClusterAwsAdmExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithDocker() *Factory {
	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.DockerClient != nil {
			return nil
		}

		f.dependencies.DockerClient = executables.BuildDockerExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithKubectl() *Factory {
	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.Kubectl != nil {
			return nil
		}

		f.dependencies.Kubectl = f.executableBuilder.BuildKubectlExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithGovc() *Factory {
	f.WithWriter()

	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.Govc != nil {
			return nil
		}

		f.dependencies.Govc = f.executableBuilder.BuildGovcExecutable(f.dependencies.Writer)
		return nil
	})

	return f
}

func (f *Factory) WithWriter() *Factory {
	f.buildSteps = append(f.buildSteps, func() error {
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
	f.WithWriter()

	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.Kind != nil {
			return nil
		}

		f.dependencies.Kind = f.executableBuilder.BuildKindExecutable(f.dependencies.Writer)
		return nil
	})

	return f
}

func (f *Factory) WithClusterctl() *Factory {
	f.WithWriter()

	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.Clusterctl != nil {
			return nil
		}

		f.dependencies.Clusterctl = f.executableBuilder.BuildClusterCtlExecutable(f.dependencies.Writer)
		return nil
	})

	return f
}

func (f *Factory) WithFlux() *Factory {
	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.Flux != nil {
			return nil
		}

		f.dependencies.Flux = f.executableBuilder.BuildFluxExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithTroubleshoot() *Factory {
	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.Troubleshoot != nil {
			return nil
		}

		f.dependencies.Troubleshoot = f.executableBuilder.BuildTroubleshootExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithNetworking() *Factory {
	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.Networking != nil {
			return nil
		}

		f.dependencies.Networking = networking.NewCilium()
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

	f.buildSteps = append(f.buildSteps, func() error {
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

func (f *Factory) WithClusterManager() *Factory {
	f.WithClusterctl().WithKubectl().WithNetworking().WithWriter()

	f.buildSteps = append(f.buildSteps, func() error {
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
		)
		return nil
	})

	return f
}

func (f *Factory) WithFluxAddonClient(ctx context.Context, clusterConfig *v1alpha1.Cluster, gitOpsConfig *v1alpha1.GitOpsConfig) *Factory {
	f.WithWriter().WithFlux().WithKubectl()

	f.buildSteps = append(f.buildSteps, func() error {
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

func (f *Factory) WithAnalyzerFactory() *Factory {
	f.buildSteps = append(f.buildSteps, func() error {
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
	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.CollectorFactory != nil {
			return nil
		}

		if f.diagnosticCollectorImage == "" {
			return errors.New("diagnostic collector image is required to build CollectorFactory")
		}

		f.dependencies.CollectorFactory = diagnostics.NewCollectorFactory(f.diagnosticCollectorImage)
		return nil
	})

	return f
}

func (f *Factory) WithCAPIUpgrader() *Factory {
	f.WithClusterctl()

	f.buildSteps = append(f.buildSteps, func() error {
		if f.dependencies.CAPIUpgrader != nil {
			return nil
		}

		f.dependencies.CAPIUpgrader = clusterapi.NewUpgrader(f.dependencies.Clusterctl)
		return nil
	})

	return f
}

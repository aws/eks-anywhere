package controllers

import (
	"context"

	"github.com/go-logr/logr"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
	snowreconciler "github.com/aws/eks-anywhere/pkg/providers/snow/reconciler"
	vspherereconciler "github.com/aws/eks-anywhere/pkg/providers/vsphere/reconciler"
)

type Manager = manager.Manager

type Factory struct {
	buildSteps        []buildStep
	dependencyFactory *dependencies.Factory
	manager           Manager
	registryBuilder   *clusters.ProviderClusterReconcilerRegistryBuilder
	reconcilers       Reconcilers

	tracker                  *remote.ClusterCacheTracker
	registry                 *clusters.ProviderClusterReconcilerRegistry
	vsphereClusterReconciler *vspherereconciler.Reconciler
	snowClusterReconciler    *snowreconciler.Reconciler
	logger                   logr.Logger
	deps                     *dependencies.Dependencies
}

type Reconcilers struct {
	ClusterReconciler           *ClusterReconciler
	VSphereDatacenterReconciler *VSphereDatacenterReconciler
	SnowMachineConfigReconciler *SnowMachineConfigReconciler
}

type buildStep func(ctx context.Context) error

func NewFactory(logger logr.Logger, manager Manager) *Factory {
	return &Factory{
		buildSteps:        make([]buildStep, 0),
		dependencyFactory: dependencies.NewFactory().WithLocalExecutables(),
		manager:           manager,
		logger:            logger,
	}
}

func (f *Factory) Build(ctx context.Context) (*Reconcilers, error) {
	deps, err := f.dependencyFactory.Build(ctx)
	if err != nil {
		return nil, err
	}

	f.deps = deps

	for _, step := range f.buildSteps {
		if err := step(ctx); err != nil {
			return nil, err
		}
	}

	f.buildSteps = make([]buildStep, 0)

	return &f.reconcilers, nil
}

func (f *Factory) WithClusterReconciler(capiProviders []clusterctlv1.Provider) *Factory {
	f.dependencyFactory.WithGovc()
	f.withTracker().WithProviderClusterReconcilerRegistry(capiProviders)

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.reconcilers.ClusterReconciler != nil {
			return nil
		}

		f.reconcilers.ClusterReconciler = NewClusterReconciler(
			f.manager.GetClient(),
			f.logger,
			f.registry,
		)

		return nil
	})
	return f
}

func (f *Factory) WithVSphereDatacenterReconciler() *Factory {
	f.dependencyFactory.WithVSphereDefaulter().WithVSphereValidator()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.reconcilers.VSphereDatacenterReconciler != nil {
			return nil
		}

		f.reconcilers.VSphereDatacenterReconciler = NewVSphereDatacenterReconciler(
			f.manager.GetClient(),
			f.logger,
			f.deps.VSphereValidator,
			f.deps.VSphereDefaulter,
		)

		return nil
	})
	return f
}

func (f *Factory) WithSnowMachineConfigReconciler() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.reconcilers.SnowMachineConfigReconciler != nil {
			return nil
		}

		client := f.manager.GetClient()
		f.reconcilers.SnowMachineConfigReconciler = NewSnowMachineConfigReconciler(
			client,
			f.logger,
			snow.NewValidator(snowreconciler.NewAwsClientBuilder(client)),
		)
		return nil
	})
	return f
}

func (f *Factory) withTracker() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.tracker != nil {
			return nil
		}

		logger := f.logger.WithName("remote").WithName("ClusterCacheTracker")
		tracker, err := remote.NewClusterCacheTracker(
			f.manager,
			remote.ClusterCacheTrackerOptions{
				Log:     &logger,
				Indexes: remote.DefaultIndexes,
			},
		)
		if err != nil {
			return err
		}

		f.tracker = tracker

		return nil
	})
	return f
}

const (
	snowProviderName    = "snow"
	vSphereProviderName = "vsphere"
)

func (f *Factory) WithProviderClusterReconcilerRegistry(capiProviders []clusterctlv1.Provider) *Factory {
	f.registryBuilder = clusters.NewProviderClusterReconcilerRegistryBuilder()

	for _, p := range capiProviders {
		if p.Type != string(clusterctlv1.InfrastructureProviderType) {
			continue
		}

		switch p.ProviderName {
		case snowProviderName:
			f.withSnowClusterReconciler()
		case vSphereProviderName:
			f.withVSphereClusterReconciler()
		default:
			f.logger.Info("Found unknown CAPI provider, ignoring", "providerName", p.ProviderName)
		}
	}

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.registry != nil {
			return nil
		}

		r := f.registryBuilder.Build()
		f.registry = &r

		return nil
	})
	return f
}

func (f *Factory) withVSphereClusterReconciler() *Factory {
	f.dependencyFactory.WithVSphereDefaulter().WithVSphereValidator()
	f.withTracker()
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.vsphereClusterReconciler != nil {
			return nil
		}

		f.vsphereClusterReconciler = vspherereconciler.New(
			f.manager.GetClient(),
			f.deps.VSphereValidator,
			f.deps.VSphereDefaulter,
			f.tracker,
		)
		f.registryBuilder.Add(anywherev1.VSphereDatacenterKind, f.vsphereClusterReconciler)

		return nil
	})

	return f
}

func (f *Factory) withSnowClusterReconciler() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.snowClusterReconciler != nil {
			return nil
		}

		f.snowClusterReconciler = snowreconciler.New(
			f.manager.GetClient(),
			f.deps.SnowValidator,
			f.deps.SnowDefaulters,
		)
		f.registryBuilder.Add(anywherev1.SnowDatacenterKind, f.snowClusterReconciler)

		return nil
	})

	return f
}

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/aws/eks-anywhere/pkg/dependencies"
)

type Manager = manager.Manager

type Factory struct {
	buildSteps        []buildStep
	dependencyFactory *dependencies.Factory
	manager           Manager
	reconcilers       Reconcilers

	tracker *remote.ClusterCacheTracker
	logger  logr.Logger
	deps    *dependencies.Dependencies
}

type Reconcilers struct {
	ClusterReconciler           *ClusterReconciler
	VSphereDatacenterReconciler *VSphereDatacenterReconciler
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

func (f *Factory) WithClusterReconciler() *Factory {
	f.dependencyFactory.WithGovc()
	f.withTracker()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.reconcilers.ClusterReconciler != nil {
			return nil
		}

		f.reconcilers.ClusterReconciler = NewClusterReconciler(
			f.manager.GetClient(),
			f.logger,
			f.manager.GetScheme(),
			f.deps.Govc,
			f.tracker,
			BuildProviderReconciler,
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

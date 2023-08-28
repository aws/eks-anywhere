package reconciler

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
)

type preflightInstallation struct {
	daemonSet  *appsv1.DaemonSet
	deployment *appsv1.Deployment
}

func (p *preflightInstallation) installed() bool {
	return p.daemonSet != nil && p.deployment != nil
}

func getPreflightInstallation(ctx context.Context, client client.Client) (*preflightInstallation, error) {
	ds, err := getDaemonSet(ctx, client, cilium.PreflightDaemonSetName)
	if err != nil {
		return nil, err
	}

	deployment, err := getDeployment(ctx, client, cilium.PreflightDeploymentName)
	if err != nil {
		return nil, err
	}

	return &preflightInstallation{
		daemonSet:  ds,
		deployment: deployment,
	}, nil
}

func getDeployment(ctx context.Context, client client.Client, name string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	key := types.NamespacedName{
		Name:      name,
		Namespace: constants.KubeSystemNamespace,
	}
	err := client.Get(ctx, key, deployment)
	switch {
	case apierrors.IsNotFound(err):
		return nil, nil
	case err != nil:
		return nil, err
	}

	return deployment, nil
}

func getDaemonSet(ctx context.Context, client client.Client, name string) (*appsv1.DaemonSet, error) {
	ds := &appsv1.DaemonSet{}
	key := types.NamespacedName{Name: name, Namespace: constants.KubeSystemNamespace}
	err := client.Get(ctx, key, ds)
	switch {
	case apierrors.IsNotFound(err):
		return nil, nil
	case err != nil:
		return nil, err
	}

	return ds, nil
}

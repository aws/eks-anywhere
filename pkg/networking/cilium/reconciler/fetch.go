package reconciler

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/networking/cilium"
)

func getInstallation(ctx context.Context, client client.Client) (*cilium.Installation, error) {
	ds, err := getCiliumDaemonSet(ctx, client)
	if err != nil {
		return nil, err
	}

	operator, err := getCiliumDeployment(ctx, client)
	if err != nil {
		return nil, err
	}

	return &cilium.Installation{
		DaemonSet: ds,
		Operator:  operator,
	}, nil
}

func getCiliumDaemonSet(ctx context.Context, client client.Client) (*appsv1.DaemonSet, error) {
	return getDaemonSet(ctx, client, cilium.DaemonSetName, "kube-system")
}

func getPreflightDaemonSet(ctx context.Context, client client.Client) (*appsv1.DaemonSet, error) {
	return getDaemonSet(ctx, client, cilium.PreflightDaemonSetName, "kube-system")
}

func getDaemonSet(ctx context.Context, client client.Client, name, namespace string) (*appsv1.DaemonSet, error) {
	ds := &appsv1.DaemonSet{}
	err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, ds)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func getCiliumDeployment(ctx context.Context, client client.Client) (*appsv1.Deployment, error) {
	return getDeployment(ctx, client, cilium.DeploymentName, "kube-system")
}

func getPreflightDeployment(ctx context.Context, client client.Client) (*appsv1.Deployment, error) {
	return getDeployment(ctx, client, cilium.PreflightDeploymentName, "kube-system")
}

func getDeployment(ctx context.Context, client client.Client, name, namespace string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, deployment)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

type preflightInstallation struct {
	daemonSet  *appsv1.DaemonSet
	deployment *appsv1.Deployment
}

func (p *preflightInstallation) installed() bool {
	return p.daemonSet != nil && p.deployment != nil
}

func getPreflightInstallation(ctx context.Context, client client.Client) (*preflightInstallation, error) {
	ds, err := getPreflightDaemonSet(ctx, client)
	if err != nil {
		return nil, err
	}

	deployment, err := getPreflightDeployment(ctx, client)
	if err != nil {
		return nil, err
	}

	return &preflightInstallation{
		daemonSet:  ds,
		deployment: deployment,
	}, nil
}

package cilium

import (
	"context"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	ciliumConfigMapName   = "cilium-config"
	ciliumConfigNamespace = "kube-system"
)

// Installation is an installation of EKSA Cilium components.
type Installation struct {
	DaemonSet *appsv1.DaemonSet
	Operator  *appsv1.Deployment
	ConfigMap *corev1.ConfigMap
}

// Installed determines if all EKS-A Embedded Cilium components are present. It identifies
// EKS-A Embedded Cilium by the image name. If the ConfigMap doesn't exist we still considered
// Cilium is installed. The installation might not be complete but it can be functional.
func (i Installation) Installed() bool {
	var isEKSACilium bool
	if i.DaemonSet != nil {
		for _, c := range i.DaemonSet.Spec.Template.Spec.Containers {
			isEKSACilium = isEKSACilium || strings.Contains(c.Image, "eksa")
		}
	}
	return i.DaemonSet != nil && i.Operator != nil && isEKSACilium
}

// GetInstallation creates a new Installation instance. The returned installation's DaemonSet,
// Operator and ConfigMap fields will be nil if they could not be found within the target cluster.
func GetInstallation(ctx context.Context, client client.Client) (*Installation, error) {
	ds, err := getDaemonSet(ctx, client)
	if err != nil {
		return nil, err
	}

	operator, err := getDeployment(ctx, client)
	if err != nil {
		return nil, err
	}

	cm, err := getConfigMap(ctx, client, ciliumConfigMapName, ciliumConfigNamespace)
	if err != nil {
		return nil, err
	}

	return &Installation{
		DaemonSet: ds,
		Operator:  operator,
		ConfigMap: cm,
	}, nil
}

func getDaemonSet(ctx context.Context, client client.Client) (*appsv1.DaemonSet, error) {
	ds := &appsv1.DaemonSet{}
	err := client.Get(ctx, types.NamespacedName{Name: DaemonSetName, Namespace: constants.KubeSystemNamespace}, ds)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func getConfigMap(ctx context.Context, client client.Client, name string, namespace string) (*corev1.ConfigMap, error) {
	c := &corev1.ConfigMap{}
	err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, c)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return c, nil
}

func getDeployment(ctx context.Context, client client.Client) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	key := types.NamespacedName{
		Name:      DeploymentName,
		Namespace: constants.KubeSystemNamespace,
	}
	err := client.Get(ctx, key, deployment)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

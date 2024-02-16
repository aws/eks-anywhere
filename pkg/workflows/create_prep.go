package workflows

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// CreateNamespaceIfNotPresent creates the namespace on the cluster if it does not already exist.
func CreateNamespaceIfNotPresent(ctx context.Context, namespace, kubeconfig string, clientFactory interfaces.ClientFactory) error {
	client, err := clientFactory.BuildClientFromKubeconfig(kubeconfig)
	if err != nil {
		return err
	}

	if err := client.Get(ctx, namespace, "", &corev1.Namespace{}); err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil {
		return nil
	}

	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	if err = client.Create(ctx, ns); err != nil {
		return err
	}

	return nil
}

package workflows

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// CreateNamespaceIfNotPresent creates the namespace on the cluster if it does not already exist.
func CreateNamespaceIfNotPresent(ctx context.Context, namespace string, client kubernetes.Client) error {
	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	if err := client.Create(ctx, ns); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

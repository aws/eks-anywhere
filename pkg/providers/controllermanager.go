package providers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// SetupProviderManagerDeployment gets the cap[infrastucture] deployment and applies tolerations corresponding to control plane taints.
func SetupProviderManagerDeployment(ctx context.Context, client kubernetes.Client, name string, namespace string) error {
	manager := appsv1.Deployment{}
	err := client.Get(ctx, name, namespace, &manager)

	if err != nil {
		return fmt.Errorf("%s not found", name)
	}

	tolerations := []corev1.Toleration{
		corev1.Toleration{
			Key:    "node-role.kubernetes.io/master",
			Effect: corev1.TaintEffectNoSchedule,
		},
		corev1.Toleration{
			Key:    "node-role.kubernetes.io/control-plane",
			Effect: corev1.TaintEffectNoSchedule,
		},
	}

	set := false
	for _, tol := range manager.Spec.Template.Spec.Tolerations {
		if tol.Key == tolerations[0].Key {
			set = true
			break
		}
	}

	if set {
		return nil
	}

	manager.Spec.Template.Spec.Tolerations = append(manager.Spec.Template.Spec.Tolerations, tolerations...)

	kerr := client.Update(ctx, &manager)

	if kerr != nil {
		return fmt.Errorf("Could not apply changes to %s deployment", name)
	}

	return nil
}

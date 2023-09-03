package providers

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	eksaerrors "github.com/aws/eks-anywhere/pkg/errors"
	"github.com/aws/eks-anywhere/pkg/validation"
)

// ValidateSSHKeyPresentForUpgrade checks that all machine configs in the cluster spec
// contain at least one SSH key.
func ValidateSSHKeyPresentForUpgrade(_ context.Context, spec *cluster.Spec) error {
	machines := make([]machineWithUsers, 0)

	// We don't add snow since SnowMachineConfig's don't have []User

	machines = appendMachinesWithUsers(machines, spec.VSphereMachineConfigs)
	machines = appendMachinesWithUsers(machines, spec.CloudStackMachineConfigs)
	machines = appendMachinesWithUsers(machines, spec.NutanixMachineConfigs)
	machines = appendMachinesWithUsers(machines, spec.TinkerbellMachineConfigs)

	if err := validateAtLeastOneSSHKey(machines); err != nil {
		return validation.WithRemediation(err, "Please include at least one SSH key per machine config. If your keys were autogenerated during the create cluster operation, make sure you include them in your cluster config for all lifecycle operations")
	}

	return nil
}

type machineWithUsers interface {
	metav1.Object
	runtime.Object
	Users() []v1alpha1.UserConfiguration
}

func validateAtLeastOneSSHKey(machines []machineWithUsers) error {
	var errs []error

Machines:
	for _, m := range machines {
		for _, user := range m.Users() {
			for _, key := range user.SshAuthorizedKeys {
				if key != "" {
					continue Machines
				}
			}
		}

		errs = append(errs,
			errors.Errorf(
				"%s %s is invalid: it should contain at least one SSH key",
				m.GetName(),
				m.GetObjectKind().GroupVersionKind().Kind,
			),
		)
	}

	return eksaerrors.NewAggregate(errs)
}

func appendMachinesWithUsers[O machineWithUsers](m []machineWithUsers, addMap map[string]O) []machineWithUsers {
	for _, machine := range addMap {
		m = append(m, machine)
	}

	return m
}

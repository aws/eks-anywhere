package resource

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/common"
)

// needsVSphereNewKubeadmConfigTemplate determines if a new KubeadmConfigTemplate is needed
// based on the changes in the worker node config and vsphere machine config.
// SSH keys comparison is not strict and equivalent keys are considered the same.
// This is useful when comparing VSphereMachineConfigs that have been inferred from a
// vsphere machine template, since comments are lost in the conversion.
func needsVSphereNewKubeadmConfigTemplate(
	newWorkerNodeGroup, oldWorkerNodeGroup *anywherev1.WorkerNodeGroupConfiguration,
	oldWorkerNodeVmc, newWorkerNodeVmc *anywherev1.VSphereMachineConfig,
) bool {
	return !anywherev1.TaintsSliceEqual(newWorkerNodeGroup.Taints, oldWorkerNodeGroup.Taints) || !anywherev1.MapEqual(newWorkerNodeGroup.Labels, oldWorkerNodeGroup.Labels) ||
		!equivalentUsers(oldWorkerNodeVmc.Spec.Users, newWorkerNodeVmc.Spec.Users)
}

// equivalentUsers compares two slices of UserConfiguration
// The logic is not order sensitive
// SSH keys comparison is no strict and equivalent keys are considered the same,
// for example keys with different users.
func equivalentUsers(a, b []anywherev1.UserConfiguration) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string][]string, len(a))
	for _, user := range a {
		m[user.Name] = user.SshAuthorizedKeys
	}
	for _, user := range b {
		if keys, ok := m[user.Name]; !ok {
			return false
		} else if !equivalentSSHKeys(user.SshAuthorizedKeys, keys) {
			return false
		}
	}
	return true
}

// equivalentSSHKeys compares two SSH keys slices
// SSH keys comparison is not strict and equivalent keys are considered the same,
// all comments are stripped before comparison
// This useful when comparing VSphereMachineConfigs that have been inferred from a
// vsphere machine template, since comments are lost in the conversion
// The logic is not order sensitive.
func equivalentSSHKeys(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]struct{}, len(a))
	for _, k := range a {
		normalizedKey, err := common.StripSshAuthorizedKeyComment(k)
		if err != nil {
			m[k] = struct{}{}
		} else {
			m[normalizedKey] = struct{}{}
		}
	}
	for _, k := range b {
		normalizedKey, err := common.StripSshAuthorizedKeyComment(k)
		if err != nil {
			normalizedKey = k
		}

		if _, ok := m[normalizedKey]; !ok {
			return false
		}
	}
	return true
}

package cilium

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

const (
	// PolicyEnforcementConfigMapKey is the key used in the "cilium-config" ConfigMap to
	// store the value for the PolicyEnforcementMode.
	PolicyEnforcementConfigMapKey = "enable-policy"

	// PolicyEnforcementComponentName is the ConfigComponentUpdatePlan name for the
	// PolicyEnforcement configuration component.
	PolicyEnforcementComponentName = "PolicyEnforcementMode"

	// EgressMasqueradeInterfacesMapKey is the key used in the "cilium-config" ConfigMap to
	// store the value for the EgressMasqueradeInterfaces.
	EgressMasqueradeInterfacesMapKey = "egress-masquerade-interfaces"

	// EgressMasqueradeInterfacesComponentName is the ConfigComponentUpdatePlan name for the
	// egressMasqueradeInterfaces configuration component.
	EgressMasqueradeInterfacesComponentName = "EgressMasqueradeInterfaces"
)

// UpgradePlan contains information about a Cilium installation upgrade.
type UpgradePlan struct {
	DaemonSet VersionedComponentUpgradePlan
	Operator  VersionedComponentUpgradePlan
	ConfigMap ConfigUpdatePlan
}

// Needed determines if an upgrade is needed or not
// Returns true if any of the installation components needs an upgrade.
func (c UpgradePlan) Needed() bool {
	return c.VersionUpgradeNeeded() || c.ConfigUpdateNeeded()
}

// VersionUpgradeNeeded determines if a version upgrade is needed or not
// Returns true if any of the installation components needs an upgrade.
func (c UpgradePlan) VersionUpgradeNeeded() bool {
	return c.DaemonSet.Needed() || c.Operator.Needed()
}

// ConfigUpdateNeeded determines if an upgrade is needed on the cilium config or not.
func (c UpgradePlan) ConfigUpdateNeeded() bool {
	return c.ConfigMap.Needed()
}

// Reason returns the reason why an upgrade might be needed
// If no upgrade needed, returns empty string
// For multiple components with needed upgrades, it composes their reasons into one.
func (c UpgradePlan) Reason() string {
	components := []interface {
		reason() string
	}{
		c.DaemonSet,
		c.Operator,
		c.ConfigMap,
	}

	s := make([]string, 0, 3)
	for _, component := range components {
		if reason := component.reason(); reason != "" {
			s = append(s, reason)
		}
	}

	return strings.Join(s, " - ")
}

// VersionedComponentUpgradePlan contains upgrade information for a Cilium versioned component.
type VersionedComponentUpgradePlan struct {
	UpgradeReason string
	OldImage      string
	NewImage      string
}

// Needed determines if an upgrade is needed or not.
func (c VersionedComponentUpgradePlan) Needed() bool {
	return c.UpgradeReason != ""
}

// reason returns the reason for the upgrade if needed.
// If upgrade is not needed, it returns an empty string.
func (c VersionedComponentUpgradePlan) reason() string {
	return c.UpgradeReason
}

// ConfigUpdatePlan contains update information for the Cilium config.
type ConfigUpdatePlan struct {
	UpdateReason string
	Components   []ConfigComponentUpdatePlan
}

// Needed determines if an upgrade is needed or not.
func (c ConfigUpdatePlan) Needed() bool {
	return c.UpdateReason != ""
}

// reason returns the reason for the upgrade if needed.
// If upgrade is not needed, it returns an empty string.
func (c ConfigUpdatePlan) reason() string {
	return c.UpdateReason
}

// generateUpdateReasonFromComponents reads the update reasons for the components
// and generates a compounded update reason. This is not thread safe.
func (c *ConfigUpdatePlan) generateUpdateReasonFromComponents() {
	r := make([]string, 0, len(c.Components))
	for _, component := range c.Components {
		if reason := component.UpdateReason; reason != "" {
			r = append(r, reason)
		}
	}

	if newReason := strings.Join(r, " - "); newReason != "" {
		c.UpdateReason = newReason
	}
}

// ConfigComponentUpdatePlan contains update information for a Cilium config component.
type ConfigComponentUpdatePlan struct {
	Name               string
	UpdateReason       string
	OldValue, NewValue string
}

// BuildUpgradePlan generates the upgrade plan information for a cilium installation by comparing it
// with a desired cluster Spec.
func BuildUpgradePlan(installation *Installation, clusterSpec *cluster.Spec) UpgradePlan {
	return UpgradePlan{
		DaemonSet: daemonSetUpgradePlan(installation.DaemonSet, clusterSpec),
		Operator:  operatorUpgradePlan(installation.Operator, clusterSpec),
		ConfigMap: configMapUpgradePlan(installation.ConfigMap, clusterSpec),
	}
}

func daemonSetUpgradePlan(ds *appsv1.DaemonSet, clusterSpec *cluster.Spec) VersionedComponentUpgradePlan {
	versionsBundle := clusterSpec.RootVersionsBundle()
	dsImage := versionsBundle.Cilium.Cilium.VersionedImage()
	info := VersionedComponentUpgradePlan{
		NewImage: dsImage,
	}

	if ds == nil {
		info.UpgradeReason = "DaemonSet doesn't exist"
		return info
	}

	oldDSImage := ds.Spec.Template.Spec.Containers[0].Image
	info.OldImage = oldDSImage

	containers := make([]corev1.Container, 0, len(ds.Spec.Template.Spec.Containers)+len(ds.Spec.Template.Spec.InitContainers))
	containers = append(containers, ds.Spec.Template.Spec.Containers...)
	containers = append(containers, ds.Spec.Template.Spec.InitContainers...)
	for _, c := range containers {
		if c.Image != dsImage {
			info.OldImage = c.Image
			info.UpgradeReason = fmt.Sprintf("DaemonSet container %s doesn't match image [%s] -> [%s]", c.Name, c.Image, dsImage)
			return info
		}
	}

	return info
}

func operatorUpgradePlan(operator *appsv1.Deployment, clusterSpec *cluster.Spec) VersionedComponentUpgradePlan {
	versionsBundle := clusterSpec.RootVersionsBundle()
	newImage := versionsBundle.Cilium.Operator.VersionedImage()
	info := VersionedComponentUpgradePlan{
		NewImage: newImage,
	}

	if operator == nil {
		info.UpgradeReason = "Operator deployment doesn't exist"
		return info
	}

	if len(operator.Spec.Template.Spec.Containers) == 0 {
		info.UpgradeReason = "Operator deployment doesn't have any containers"
		return info
	}

	oldImage := operator.Spec.Template.Spec.Containers[0].Image
	info.OldImage = oldImage

	if oldImage != newImage {
		info.UpgradeReason = fmt.Sprintf("Operator container doesn't match the provided image [%s] -> [%s]", oldImage, newImage)
		return info
	}

	return info
}

func configMapUpgradePlan(configMap *corev1.ConfigMap, clusterSpec *cluster.Spec) ConfigUpdatePlan {
	updatePlan := &ConfigUpdatePlan{}

	var newEnforcementPolicy string
	if clusterSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode == "" {
		newEnforcementPolicy = "default"
	} else {
		newEnforcementPolicy = string(clusterSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode)
	}

	policyEnforcementUpdate := ConfigComponentUpdatePlan{
		Name:     PolicyEnforcementComponentName,
		NewValue: newEnforcementPolicy,
	}

	if configMap == nil {
		updatePlan.UpdateReason = "Cilium config doesn't exist"
	} else if val, ok := configMap.Data[PolicyEnforcementConfigMapKey]; ok && val != "" {
		policyEnforcementUpdate.OldValue = val
		if policyEnforcementUpdate.OldValue != policyEnforcementUpdate.NewValue {
			policyEnforcementUpdate.UpdateReason = fmt.Sprintf("Cilium enable-policy changed: [%s] -> [%s]", policyEnforcementUpdate.OldValue, policyEnforcementUpdate.NewValue)
		}
	} else {
		policyEnforcementUpdate.UpdateReason = "Cilium enable-policy field is not present in config"
	}

	updatePlan.Components = append(updatePlan.Components, policyEnforcementUpdate)

	newEgressMasqueradeInterfaces := clusterSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.EgressMasqueradeInterfaces

	egressMasqueradeUpdate := ConfigComponentUpdatePlan{
		Name:     EgressMasqueradeInterfacesComponentName,
		NewValue: newEgressMasqueradeInterfaces,
	}

	if configMap == nil {
		updatePlan.UpdateReason = "Cilium config doesn't exist"
	} else if val, ok := configMap.Data[EgressMasqueradeInterfacesMapKey]; ok && val != "" {
		egressMasqueradeUpdate.OldValue = val
		if egressMasqueradeUpdate.OldValue != egressMasqueradeUpdate.NewValue {
			egressMasqueradeUpdate.UpdateReason = fmt.Sprintf("Egress masquerade interfaces changed: [%s] -> [%s]", egressMasqueradeUpdate.OldValue, egressMasqueradeUpdate.NewValue)
		}
	} else if egressMasqueradeUpdate.NewValue != "" {
		egressMasqueradeUpdate.UpdateReason = "Egress masquerade interfaces field is not present in config but is configured in cluster spec"
	}

	updatePlan.Components = append(updatePlan.Components, egressMasqueradeUpdate)

	updatePlan.generateUpdateReasonFromComponents()

	return *updatePlan
}

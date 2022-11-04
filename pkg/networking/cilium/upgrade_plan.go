package cilium

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

// UpgradePlan contains information about a Cilium installation upgrade.
type UpgradePlan struct {
	DaemonSet ComponentUpgradePlan
	Operator  ComponentUpgradePlan
}

// Needed determines if an upgrade is needed or not
// Returns true if any of the installation components needs an upgrade.
func (c UpgradePlan) Needed() bool {
	return c.DaemonSet.Needed() || c.Operator.Needed()
}

// Reason returns the reason why an upgrade might be needed
// If no upgrade needed, returns empty string
// For multiple components with needed upgrades, it composes their reasons into one.
func (c UpgradePlan) Reason() string {
	if !c.Needed() {
		return ""
	}

	s := make([]string, 0, 2)
	if c.DaemonSet.UpgradeReason != "" {
		s = append(s, c.DaemonSet.UpgradeReason)
	}
	if c.Operator.UpgradeReason != "" {
		s = append(s, c.Operator.UpgradeReason)
	}

	return strings.Join(s, " - ")
}

// ComponentUpgradePlan contains upgrade information for a Cilium component.
type ComponentUpgradePlan struct {
	UpgradeReason string
	OldImage      string
	NewImage      string
}

// Needed determines if an upgrade is needed or not.
func (c ComponentUpgradePlan) Needed() bool {
	return c.UpgradeReason != ""
}

// BuildUpgradePlan generates the upgrade plan information for a cilium installation by comparing it
// with a desired cluster Spec.
func BuildUpgradePlan(installation *Installation, clusterSpec *cluster.Spec) UpgradePlan {
	return UpgradePlan{
		DaemonSet: daemonSetUpgradePlan(installation.DaemonSet, clusterSpec),
		Operator:  operatorUpgradePlan(installation.Operator, clusterSpec),
	}
}

func daemonSetUpgradePlan(ds *appsv1.DaemonSet, clusterSpec *cluster.Spec) ComponentUpgradePlan {
	dsImage := clusterSpec.VersionsBundle.Cilium.Cilium.VersionedImage()
	info := ComponentUpgradePlan{
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
			info.UpgradeReason = fmt.Sprintf("DaemonSet container %s doesn't match image", c.Name)
			return info
		}
	}

	return info
}

func operatorUpgradePlan(operator *appsv1.Deployment, clusterSpec *cluster.Spec) ComponentUpgradePlan {
	newImage := clusterSpec.VersionsBundle.Cilium.Operator.VersionedImage()
	info := ComponentUpgradePlan{
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
		info.UpgradeReason = "Operator container doesn't match the provided image"
		return info
	}

	return info
}

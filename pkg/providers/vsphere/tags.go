package vsphere

import (
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func requiredTemplateTags(machineConfig *v1alpha1.VSphereMachineConfig, versionsBundle *cluster.VersionsBundle) []string {
	tagsByCategory := requiredTemplateTagsByCategory(machineConfig, versionsBundle)
	tags := make([]string, 0, len(tagsByCategory))
	for _, t := range tagsByCategory {
		tags = append(tags, t...)
	}

	return tags
}

func requiredTemplateTagsByCategory(machineConfig *v1alpha1.VSphereMachineConfig, versionsBundle *cluster.VersionsBundle) map[string][]string {
	osFamily := machineConfig.Spec.OSFamily
	return map[string][]string{
		"eksdRelease": {fmt.Sprintf("eksdRelease:%s", versionsBundle.EksD.Name)},
		"os":          {fmt.Sprintf("os:%s", strings.ToLower(string(osFamily)))},
	}
}

package cluster

import anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"

func gitOpsEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.GitOpsConfigKind: func() APIObject {
				return &anywherev1.GitOpsConfig{}
			},
		},
		Processors: []ParsedProcessor{processGitOps},
	}
}

func processGitOps(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.GitOpsRef == nil {
		return
	}

	if c.Cluster.Spec.GitOpsRef.Kind == anywherev1.GitOpsConfigKind {
		gitOps := objects.GetFromRef(c.Cluster.APIVersion, *c.Cluster.Spec.GitOpsRef)
		if gitOps == nil {
			return
		}

		c.GitOpsConfig = gitOps.(*anywherev1.GitOpsConfig)
	}
}

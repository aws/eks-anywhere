package features

// These are environment variables used as flags to enable/disable features.
const (
	CloudStackKubeVipDisabledEnvVar   = "CLOUDSTACK_KUBE_VIP_DISABLED"
	CheckpointEnabledEnvVar           = "CHECKPOINT_ENABLED"
	UseNewWorkflowsEnvVar             = "USE_NEW_WORKFLOWS"
	UseControllerForCli               = "USE_CONTROLLER_FOR_CLI"
	VSphereInPlaceEnvVar              = "VSPHERE_IN_PLACE_UPGRADE"
	APIServerExtraArgsEnabledEnvVar   = "API_SERVER_EXTRA_ARGS_ENABLED"
	K8s132SupportEnvVar               = "K8S_1_32_SUPPORT"
	VSPhereFailureDomainEnabledEnvVar = "VSPHERE_FAILURE_DOMAIN_ENABLED"
)

func FeedGates(featureGates []string) {
	globalFeatures.feedGates(featureGates)
}

type Feature struct {
	Name     string
	IsActive func() bool
}

func IsActive(feature Feature) bool {
	return feature.IsActive()
}

// ClearCache is mainly used for unit tests as of now.
func ClearCache() {
	globalFeatures.clearCache()
}

func CloudStackKubeVipDisabled() Feature {
	return Feature{
		Name:     "Kube-vip support disabled in CloudStack provider",
		IsActive: globalFeatures.isActiveForEnvVar(CloudStackKubeVipDisabledEnvVar),
	}
}

func CheckpointEnabled() Feature {
	return Feature{
		Name:     "Checkpoint to rerun commands enabled",
		IsActive: globalFeatures.isActiveForEnvVar(CheckpointEnabledEnvVar),
	}
}

func UseNewWorkflows() Feature {
	return Feature{
		Name:     "Use new workflow logic for cluster management operations",
		IsActive: globalFeatures.isActiveForEnvVar(UseNewWorkflowsEnvVar),
	}
}

// VSphereInPlaceUpgradeEnabled is the feature flag for performing in-place upgrades with the vSphere provider.
func VSphereInPlaceUpgradeEnabled() Feature {
	return Feature{
		Name:     "Perform in-place upgrades with the vSphere provider",
		IsActive: globalFeatures.isActiveForEnvVar(VSphereInPlaceEnvVar),
	}
}

// APIServerExtraArgsEnabled is the feature flag for configuring api server extra args.
func APIServerExtraArgsEnabled() Feature {
	return Feature{
		Name:     "Configure api server extra args",
		IsActive: globalFeatures.isActiveForEnvVar(APIServerExtraArgsEnabledEnvVar),
	}
}

// K8s132Support is the feature flag for Kubernetes 1.32 support.
func K8s132Support() Feature {
	return Feature{
		Name:     "Kubernetes version 1.32 support",
		IsActive: globalFeatures.isActiveForEnvVar(K8s132SupportEnvVar),
	}
}

func VsphereFailureDomainEnabled() Feature {
	return Feature{
		Name:     "Vsphere Failure Domains Enabled",
		IsActive: globalFeatures.isActiveForEnvVar(VSPhereFailureDomainEnabledEnvVar),
	}
}

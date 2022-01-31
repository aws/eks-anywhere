package features

const (
	TaintsSupportEnvVar      = "TAINTS_SUPPORT"
	NodeLabelsSupportEnvVar  = "NODE_LABELS_SUPPORT"
	TinkerbellProviderEnvVar = "TINKERBELL_PROVIDER"
	FullLifecycleAPIEnvVar   = "FULL_LIFECYCLE_API"
	FullLifecycleGate        = "FullLifecycleAPI"
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

func TaintsSupport() Feature {
	return Feature{
		Name:     "Taints support",
		IsActive: globalFeatures.isActiveForEnvVar(TaintsSupportEnvVar),
	}
}

func NodeLabelsSupport() Feature {
	return Feature{
		Name:     "Node labels support",
		IsActive: globalFeatures.isActiveForEnvVar(NodeLabelsSupportEnvVar),
	}
}

func FullLifecycleAPI() Feature {
	return Feature{
		Name:     "Full lifecycle API support through the EKS-A controller",
		IsActive: globalFeatures.isActiveForEnvVarOrGate(FullLifecycleAPIEnvVar, FullLifecycleGate),
	}
}

func TinkerbellProvider() Feature {
	return Feature{
		Name:     "Tinkerbell provider support",
		IsActive: globalFeatures.isActiveForEnvVar(TinkerbellProviderEnvVar),
	}
}

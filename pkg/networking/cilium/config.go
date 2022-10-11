package cilium

import "github.com/aws/eks-anywhere/pkg/cluster"

// Config defines a Cilium installation for an EKS-A cluster.
type Config struct {
	// AllowedNamespaces defines k8s namespaces from/which traffic is allowed
	// when PolicyEnforcementMode is Always. For other values of PolicyEnforcementMode
	// it is ignored.
	AllowedNamespaces []string

	// Spec is the complete EKS-A cluster definition
	Spec *cluster.Spec
	// TODO(gaslor): we should try to reduce down the dependency here and narrow it down
	// to the bare minimun. This requires to refactor the templater to not depend on the
	// cluster spec
}

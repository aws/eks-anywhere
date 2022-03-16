package cluster

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

func ValidateConfig(c *Config) error {
	return manager().Validate(c)
}

type namespaceObject interface {
	runtime.Object
	GetNamespace() string
}

func validateSameNamespace(c *Config, o namespaceObject) error {
	if c.Cluster.Namespace != o.GetNamespace() {
		return fmt.Errorf("%s and Cluster objects must have the same namespace specified", o.GetObjectKind().GroupVersionKind().Kind)
	}

	return nil
}

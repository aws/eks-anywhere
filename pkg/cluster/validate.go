package cluster

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

type validableValidation func(*Config, validable) error

var validableValidations = []validableValidation{
	validateSameNamespace,
}

func ValidateConfig(c *Config) error {
	var allErrs []error
	validables := getValidables(c)

	for _, v := range validables {
		if err := v.Validate(); err != nil {
			allErrs = append(allErrs, err)
		}

		for _, validation := range validableValidations {
			if err := validation(c, v); err != nil {
				allErrs = append(allErrs, err)
			}
		}
	}

	if len(allErrs) > 0 {
		aggregate := utilerrors.NewAggregate(allErrs)
		return fmt.Errorf("invalid cluster config: %v", aggregate)
	}

	return nil
}

type validable interface {
	Validate() error
	GetNamespace() string
	GetObjectKind() schema.ObjectKind
}

func getValidables(c *Config) []validable {
	v := make([]validable, 0, 3)
	v = appendIfNotNil(v, c.Cluster, c.VSphereDatacenter, c.DockerDatacenter, c.GitOpsConfig)

	for _, e := range c.VSphereMachineConfigs {
		v = appendIfNotNil(v, e)
	}

	for _, e := range c.OIDCConfigs {
		v = appendIfNotNil(v, e)
	}

	for _, e := range c.AWSIAMConfigs {
		v = appendIfNotNil(v, e)
	}

	return v
}

func appendIfNotNil(validables []validable, elems ...validable) []validable {
	for _, e := range elems {
		// Since we receive interfaces, these will never be nil since they contain
		// the type of the original implementing struct
		// I can't find another clean option of doing this
		if !reflect.ValueOf(e).IsNil() {
			validables = append(validables, e)
		}
	}

	return validables
}

func validateSameNamespace(c *Config, v validable) error {
	if c.Cluster.Namespace != v.GetNamespace() {
		return fmt.Errorf("%s and Cluster objects must have the same namespace specified", v.GetObjectKind().GroupVersionKind().Kind)
	}

	return nil
}

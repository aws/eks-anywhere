package cluster

import (
	"context"

	"k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

// StateValidation defines a validation that can be registered to the StateValidator.
type StateValidation = func(ctx context.Context, vc StateValidationConfig) error

// RetriableStateValidation returns a StateValidation that is executed with the provided retrier.
func RetriableStateValidation(retrier *retrier.Retrier, validation StateValidation) StateValidation {
	return func(ctx context.Context, vc StateValidationConfig) error {
		err := retrier.Retry(func() error {
			return validation(ctx, vc)
		})
		return err
	}
}

// StateValidator is responsible for checking if a cluster is valid against the spec that is provided.
type StateValidator struct {
	Config      StateValidationConfig
	validations []StateValidation
}

// WithValidations registers multiple validations to the StateValidator that will be run when Validate is called.
func (c *StateValidator) WithValidations(validations ...StateValidation) {
	c.validations = append(c.validations, validations...)
}

// Validate runs through the set registered validations and returns an error if any of them fail after a number of retries.
func (c *StateValidator) Validate(ctx context.Context) error {
	errList := make([]error, 0)
	for _, validate := range c.validations {
		err := validate(ctx, c.Config)
		if err != nil {
			errList = append(errList, err)
		}
	}
	return errors.NewAggregate(errList)
}

// Opt represents is a function that represents an option to configure a StateValidator.
type Opt = func(cv *StateValidator)

// NewStateValidator returns a cluster validator which can be configured by passing Opt arguments.
func NewStateValidator(config StateValidationConfig, opts ...Opt) *StateValidator {
	cv := StateValidator{
		Config:      config,
		validations: []StateValidation{},
	}

	for _, opt := range opts {
		opt(&cv)
	}
	return &cv
}

// StateValidationConfig represents the input for the performing validations on the cluster.
type StateValidationConfig struct {
	ClusterClient           client.Client // the client for the cluster
	ManagementClusterClient client.Client // the client for the management cluster
	ClusterSpec             *cluster.Spec // the cluster spec
}

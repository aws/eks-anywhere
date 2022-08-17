package reconciler

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
)

type ClientRegistry interface {
	Get(ctx context.Context) (aws.Clients, error)
}

type validatorBuilder struct {
	clientRegistry ClientRegistry
}

func NewValidatorBuilder(clientRegistry ClientRegistry) *validatorBuilder {
	return &validatorBuilder{
		clientRegistry: clientRegistry,
	}
}

func (b *validatorBuilder) Build(ctx context.Context) (snow.Validator, error) {
	deviceClientMap, err := b.clientRegistry.Get(ctx)
	if err != nil {
		return nil, err
	}
	return snow.NewValidator(deviceClientMap), nil
}

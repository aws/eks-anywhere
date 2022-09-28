package oras

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
)

type Puller struct {
	registryPuller *artifacts.RegistryPuller
}

func NewPuller(registryPuller *artifacts.RegistryPuller) *Puller {
	return &Puller{
		registryPuller: registryPuller,
	}
}

func (bp *Puller) PullLatestBundle(ctx context.Context, art string) ([]byte, error) {

	data, err := bp.registryPuller.Pull(ctx, art)
	if err != nil {
		return nil, fmt.Errorf("unable to pull artifacts %v", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("latest package bundle artifact is empty")
	}

	return data, nil
}

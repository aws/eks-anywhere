package oras

import (
	"context"
	"fmt"

	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"

	"github.com/aws/eks-anywhere/pkg/logger"
)

type Pusher struct {
	registry *content.Registry
	store    *content.Memory
}

func NewPusher(registry *content.Registry, store *content.Memory) *Pusher {
	return &Pusher{
		registry: registry,
		store:    store,
	}
}

func (p *Pusher) PushBundle(ctx context.Context, ref, fileName string, fileContent []byte) error {
	desc, err := p.store.Add("bundle.yaml", "", fileContent)
	if err != nil {
		return err
	}

	manifest, manifestDesc, config, configDesc, err := content.GenerateManifestAndConfig(nil, nil, desc)
	if err != nil {
		return err
	}

	p.store.Set(configDesc, config)
	err = p.store.StoreManifest(ref, manifestDesc, manifest)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Pushing %s to %s...", fileName, ref))
	desc, err = oras.Copy(ctx, p.store, ref, p.registry, "")
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Pushed to %s with digest %s", ref, desc.Digest))
	return nil
}

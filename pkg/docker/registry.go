package docker

import (
	"context"
)

// ImageRegistryDestination implements the ImageDestination interface, writing images and tags from
// from the local docker cache to an external registry
type ImageRegistryDestination struct {
	client   ImageTaggerPusher
	endpoint string
}

func NewRegistryDestination(client ImageTaggerPusher, registryEndpoint string) *ImageRegistryDestination {
	return &ImageRegistryDestination{
		client:   client,
		endpoint: registryEndpoint,
	}
}

// Write pushes images and tags from from the local docker cache to an external registry
func (d *ImageRegistryDestination) Write(ctx context.Context, images ...string) error {
	for _, i := range images {
		if err := d.client.TagImage(ctx, i, d.endpoint); err != nil {
			return err
		}

		if err := d.client.PushImage(ctx, i, d.endpoint); err != nil {
			return err
		}
	}

	return nil
}

// ImageOriginalRegistrySource implements the ImageSource interface, pulling images and tags from
// their original registry into the local docker cache
type ImageOriginalRegistrySource struct {
	client ImagePuller
}

func NewOriginalRegistrySource(client ImagePuller) *ImageOriginalRegistrySource {
	return &ImageOriginalRegistrySource{
		client: client,
	}
}

// Load pulls images and tags from their original registry into the local docker cache
func (s *ImageOriginalRegistrySource) Load(ctx context.Context, images ...string) error {
	for _, i := range images {
		if err := s.client.PullImage(ctx, i); err != nil {
			return err
		}
	}

	return nil
}

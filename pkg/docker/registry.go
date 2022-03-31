package docker

import (
	"context"
	"runtime"

	"github.com/aws/eks-anywhere/pkg/logger"
)

// ImageRegistryDestination implements the ImageDestination interface, writing images and tags from
// from the local docker cache to an external registry
type ImageRegistryDestination struct {
	client    ImageTaggerPusher
	endpoint  string
	processor *ConcurrentImageProcessor
}

func NewRegistryDestination(client ImageTaggerPusher, registryEndpoint string) *ImageRegistryDestination {
	return &ImageRegistryDestination{
		client:    client,
		endpoint:  registryEndpoint,
		processor: NewConcurrentImageProcessor(runtime.GOMAXPROCS(0) / 2),
	}
}

// Write pushes images and tags from from the local docker cache to an external registry
func (d *ImageRegistryDestination) Write(ctx context.Context, images ...string) error {
	logger.Info("Writing images to registry")
	logger.V(3).Info("Starting registry write", "numberOfImages", len(images))
	err := d.processor.Process(ctx, images, func(ctx context.Context, image string) error {
		if err := d.client.TagImage(ctx, image, d.endpoint); err != nil {
			return err
		}

		if err := d.client.PushImage(ctx, image, d.endpoint); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ImageOriginalRegistrySource implements the ImageSource interface, pulling images and tags from
// their original registry into the local docker cache
type ImageOriginalRegistrySource struct {
	client    ImagePuller
	processor *ConcurrentImageProcessor
}

func NewOriginalRegistrySource(client ImagePuller) *ImageOriginalRegistrySource {
	return &ImageOriginalRegistrySource{
		client:    client,
		processor: NewConcurrentImageProcessor(runtime.GOMAXPROCS(0) / 2),
	}
}

// Load pulls images and tags from their original registry into the local docker cache
func (s *ImageOriginalRegistrySource) Load(ctx context.Context, images ...string) error {
	logger.Info("Pulling images from origin, this might take a while")
	logger.V(3).Info("Starting pull", "numberOfImages", len(images))

	err := s.processor.Process(ctx, images, func(ctx context.Context, image string) error {
		if err := s.client.PullImage(ctx, image); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

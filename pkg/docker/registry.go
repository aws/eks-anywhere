package docker

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/utils/urls"
)

// ImageRegistryDestination implements the ImageDestination interface, writing images and tags from
// from the local docker cache to an external registry
type ImageRegistryDestination struct {
	client    ImageTaggerPusher
	endpoint  string
	namespace string
	processor *ConcurrentImageProcessor
}

func NewRegistryDestination(client ImageTaggerPusher, registryEndpoint, namespace string) *ImageRegistryDestination {
	return &ImageRegistryDestination{
		client:    client,
		endpoint:  registryEndpoint,
		namespace: namespace,
		processor: NewConcurrentImageProcessor(runtime.GOMAXPROCS(0)),
	}
}

// Write pushes images and tags from from the local docker cache to an external registry
func (d *ImageRegistryDestination) Write(ctx context.Context, images ...string) error {
	logger.Info("Writing images to registry")
	logger.V(3).Info("Starting registry write", "numberOfImages", len(images))
	err := d.processor.Process(ctx, images, func(ctx context.Context, image string) error {
		image = removeDigestReference(image)
		if err := d.client.TagImage(ctx, image, d.endpoint, d.namespace); err != nil {
			return err
		}

		if err := d.client.PushImage(ctx, image, d.endpoint, d.namespace); err != nil {
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
		processor: NewConcurrentImageProcessor(runtime.GOMAXPROCS(0)),
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

func ReplaceHostWithNamespacedEndpoint(uri, mirrorRegistry, namespace string) string {
	if mirrorRegistry == "" {
		return uri
	}
	uri = urls.ReplaceHost(uri, mirrorRegistry)
	if namespace == "" {
		return uri
	}
	return strings.ReplaceAll(uri, mirrorRegistry, mirrorRegistry+"/"+namespace)
}

// Curated packages are currently referenced by digest
// Docker doesn't support tagging images with digest
// This method extracts any @ in the image tag
func removeDigestReference(image string) string {
	imageSplit := strings.Split(image, "@")
	if len(imageSplit) < 2 {
		return image
	}
	imageLocation, digest := imageSplit[0], imageSplit[1]
	digestSplit := strings.Split(digest, ":")
	return fmt.Sprintf("%s:%s", imageLocation, digestSplit[1])
}

package docker

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

// These constants are temporary since currently there is a limitation on harbor
// Harbor requires root level projects but curated packages private account currently
// doesn't have support for root level.
const (
	packageProdDomain = "783794618700.dkr.ecr.us-west-2.amazonaws.com"
	packageDevDomain  = "857151390494.dkr.ecr.us-west-2.amazonaws.com"
	publicProdECRName = "eks-anywhere"
	publicDevECRName  = "l0g8r8j6"
)

// ImageRegistryDestination implements the ImageDestination interface, writing images and tags from
// from the local docker cache to an external registry.
type ImageRegistryDestination struct {
	client    ImageTaggerPusher
	endpoint  string
	processor *ConcurrentImageProcessor
}

func NewRegistryDestination(client ImageTaggerPusher, registryEndpoint string) *ImageRegistryDestination {
	return &ImageRegistryDestination{
		client:    client,
		endpoint:  registryEndpoint,
		processor: NewConcurrentImageProcessor(runtime.GOMAXPROCS(0)),
	}
}

// Write pushes images and tags from from the local docker cache to an external registry.
func (d *ImageRegistryDestination) Write(ctx context.Context, images ...string) error {
	logger.Info("Writing images to registry")
	logger.V(3).Info("Starting registry write", "numberOfImages", len(images))
	err := d.processor.Process(ctx, images, func(ctx context.Context, image string) error {
		endpoint := getUpdatedEndpoint(d.endpoint, image)
		image = removeDigestReference(image)
		if err := d.client.TagImage(ctx, image, endpoint); err != nil {
			return err
		}

		if err := d.client.PushImage(ctx, image, endpoint); err != nil {
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
// their original registry into the local docker cache.
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

// Load pulls images and tags from their original registry into the local docker cache.
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

// Currently private curated packages don't have a root level project
// This method adds a root level projectName to the endpoint.
func getUpdatedEndpoint(originalEndpoint, image string) string {
	if strings.Contains(image, packageDevDomain) {
		return originalEndpoint + "/" + publicDevECRName
	}
	if strings.Contains(image, packageProdDomain) {
		return originalEndpoint + "/" + publicProdECRName
	}
	return originalEndpoint
}

// Curated packages are currently referenced by digest
// Docker doesn't support tagging images with digest
// This method extracts any @ in the image tag.
func removeDigestReference(image string) string {
	imageSplit := strings.Split(image, "@")
	if len(imageSplit) < 2 {
		return image
	}
	imageLocation, digest := imageSplit[0], imageSplit[1]
	digestSplit := strings.Split(digest, ":")
	return fmt.Sprintf("%s:%s", imageLocation, digestSplit[1])
}

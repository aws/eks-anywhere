package docker

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/logger"
)

// ImageDiskSource implements the ImageSource interface, loading images and tags from
// a tarbal into the local docker cache.
type ImageDiskSource struct {
	client ImageDiskLoader
	file   string
}

func NewDiskSource(client ImageDiskLoader, file string) *ImageDiskSource {
	return &ImageDiskSource{
		client: client,
		file:   file,
	}
}

// Load reads images and tags from a tarbal into the local docker cache.
func (s *ImageDiskSource) Load(ctx context.Context, images ...string) error {
	if cache, ok := s.client.(imageCache); ok && len(images) > 0 {
		if ids, err := archiveImageIDs(s.file); err == nil && imagesCached(ctx, cache, ids, images) {
			logger.Info("Using verified images already in the docker cache")
			return nil
		}
	}
	logger.Info("Loading images from disk")
	return s.client.LoadFromFile(ctx, s.file)
}

type imageCache interface {
	ImageID(context.Context, string) (string, error)
}

func imagesCached(ctx context.Context, cache imageCache, ids map[string]string, images []string) bool {
	for _, image := range images {
		expected, ok := ids[image]
		if !ok {
			return false
		}
		actual, err := cache.ImageID(ctx, image)
		if err != nil || actual != expected {
			return false
		}
	}
	return true
}

// ImageDiskDestination implements the ImageDestination interface, writing images and tags from
// from the local docker cache into a tarbal.
type ImageDiskDestination struct {
	client ImageDiskWriter
	file   string
}

func NewDiskDestination(client ImageDiskWriter, file string) *ImageDiskDestination {
	return &ImageDiskDestination{
		client: client,
		file:   file,
	}
}

// Write creates a tarball including images and tags from the the local docker cache.
func (s *ImageDiskDestination) Write(ctx context.Context, images ...string) error {
	logger.Info("Writing images to disk")
	return s.client.SaveToFile(ctx, s.file, images...)
}

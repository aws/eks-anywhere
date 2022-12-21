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
	logger.Info("Loading images from disk")
	return s.client.LoadFromFile(ctx, s.file)
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

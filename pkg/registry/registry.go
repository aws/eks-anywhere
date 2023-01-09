package registry

import (
	"context"
	"fmt"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
)

// Copy an image from a source to a destination.
func Copy(ctx context.Context, srcClient *OCIRegistryClient, dstClient StorageClient, image Artifact) (err error) {
	srcStorage, err := srcClient.GetStorage(ctx, image)
	if err != nil {
		return fmt.Errorf("registry copy source: %v", err)
	}

	dstStorage, err := dstClient.GetStorage(ctx, image)
	if err != nil {
		return fmt.Errorf("registry copy destination: %v", err)
	}

	var desc ocispec.Descriptor
	srcClient.registry.Reference.Reference = image.VersionedImage()
	desc, err = srcStorage.Resolve(ctx, srcClient.registry.Reference.Reference)
	if err != nil {
		return fmt.Errorf("registry source resolve: %v", err)
	}

	extendedCopyOptions := oras.DefaultExtendedCopyOptions
	err = oras.CopyGraph(ctx, srcStorage, dstStorage, desc, extendedCopyOptions.CopyGraphOptions)
	if err != nil {
		return fmt.Errorf("registry copy: %v", err)
	}

	return nil
}

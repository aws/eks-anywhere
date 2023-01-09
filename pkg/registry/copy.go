package registry

import (
	"context"
	"fmt"
)

// Copy an image from a source to a destination.
func Copy(ctx context.Context, srcClient StorageClient, dstClient StorageClient, image Artifact) (err error) {
	srcStorage, err := srcClient.GetStorage(ctx, image)
	if err != nil {
		return fmt.Errorf("registry copy source: %v", err)
	}

	dstStorage, err := dstClient.GetStorage(ctx, image)
	if err != nil {
		return fmt.Errorf("registry copy destination: %v", err)
	}

	desc, err := srcClient.Resolve(ctx, srcStorage, image.VersionedImage())
	if err != nil {
		return fmt.Errorf("registry source resolve: %v", err)
	}

	err = srcClient.CopyGraph(ctx, srcStorage, dstStorage, desc)
	if err != nil {
		return fmt.Errorf("registry copy: %v", err)
	}

	return nil
}

package registry

import (
	"context"
	"fmt"
)

// Copy an image from a source to a destination.
func Copy(ctx context.Context, srcClient StorageClient, dstClient StorageClient, image Artifact) (err error) {
	srcStorage, err := srcClient.GetStorage(ctx, image)
	if err != nil {
		return fmt.Errorf("repository source: %v", err)
	}

	dstStorage, err := dstClient.GetStorage(ctx, image)
	if err != nil {
		return fmt.Errorf("repository destination: %v", err)
	}

	err = srcClient.CopyGraph(ctx, srcStorage, image.VersionedImage(), dstStorage, dstClient.Destination(image))
	if err != nil {
		return fmt.Errorf("registry copy: %v", err)
	}

	return nil
}

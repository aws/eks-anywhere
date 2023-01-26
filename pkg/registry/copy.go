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

	desc, err := srcClient.CopyGraph(ctx, srcStorage, image.VersionedImage(), dstStorage, dstClient.Destination(image))
	if err != nil {
		return fmt.Errorf("registry copy: %v", err)
	}

	if len(image.Tag) > 0 {
		err = dstClient.Tag(ctx, dstStorage, desc, image.Tag)
		if err != nil {
			return fmt.Errorf("image tag: %v", err)
		}
	}
	return nil
}

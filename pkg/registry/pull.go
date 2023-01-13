package registry

import (
	"context"
	"encoding/json"
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// PullBytes a resource from the registry.
func PullBytes(ctx context.Context, sc StorageClient, artifact Artifact) (data []byte, err error) {
	srcStorage, err := sc.GetStorage(ctx, artifact)
	if err != nil {
		return nil, fmt.Errorf("repository source: %v", err)
	}

	_, data, err = sc.FetchBytes(ctx, srcStorage, artifact)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %v", err)
	}

	var mani ocispec.Manifest
	if err := json.Unmarshal(data, &mani); err != nil {
		return nil, fmt.Errorf("unmarshal manifest: %v", err)
	}
	if len(mani.Layers) < 1 {
		return nil, fmt.Errorf("missing layer")
	}

	data, err = sc.FetchBlob(ctx, srcStorage, mani.Layers[0])
	if err != nil {
		return nil, fmt.Errorf("fetch blob: %v", err)
	}
	return data, err
}

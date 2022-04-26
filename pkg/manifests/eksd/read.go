package eksd

import (
	"fmt"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"sigs.k8s.io/yaml"
)

type Reader interface {
	ReadFile(url string) ([]byte, error)
}

func ReadManifest(reader Reader, url string) (*eksdv1.Release, error) {
	content, err := reader.ReadFile(url)
	if err != nil {
		return nil, err
	}

	eksd := &eksdv1.Release{}
	if err = yaml.Unmarshal(content, eksd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal eksd manifest: %v", err)
	}

	return eksd, nil
}

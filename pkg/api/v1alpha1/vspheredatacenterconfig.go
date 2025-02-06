package v1alpha1

import (
	"fmt"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const VSphereDatacenterKind = "VSphereDatacenterConfig"

type folderType string

const (
	networkFolderType folderType = "network"
)

// Used for generating yaml for generate clusterconfig command.
func NewVSphereDatacenterConfigGenerate(clusterName string) *VSphereDatacenterConfigGenerate {
	return &VSphereDatacenterConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       VSphereDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: VSphereDatacenterConfigSpec{},
	}
}

func (c *VSphereDatacenterConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *VSphereDatacenterConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *VSphereDatacenterConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetVSphereDatacenterConfig(fileName string) (*VSphereDatacenterConfig, error) {
	var clusterConfig VSphereDatacenterConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}

func generateFullVCenterPath(foldType folderType, folderPath string, datacenter string) string {
	if folderPath == "" {
		return folderPath
	}

	prefix := fmt.Sprintf("/%s", datacenter)
	modPath := folderPath
	if !strings.HasPrefix(folderPath, prefix) {
		modPath = fmt.Sprintf("%s/%s/%s", prefix, foldType, folderPath)
		logger.V(4).Info(fmt.Sprintf("Relative %s path specified, using path %s", foldType, modPath))
		return modPath
	}

	return modPath
}

func validatePath(foldType folderType, folderPath string, datacenter string) error {
	prefix := filepath.Join(fmt.Sprintf("/%s", datacenter), string(foldType))
	if !strings.HasPrefix(folderPath, prefix) {
		return fmt.Errorf("invalid path, expected path [%s] to be under [%s]", folderPath, prefix)
	}

	return nil
}

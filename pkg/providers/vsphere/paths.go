package vsphere

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

type folderType string

const (
	networkFolderType folderType = "network"
)

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

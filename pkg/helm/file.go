package helm

import (
	"path/filepath"
	"strings"
)

func ChartFileName(chart string) string {
	return strings.Replace(filepath.Base(chart), ":", "-", 1) + ".tgz"
}

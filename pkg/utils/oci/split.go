package oci

import (
	"fmt"
	"path/filepath"
	"strings"
)

const OCIPrefix = "oci://"

func Split(artifact string) (path, tag string) {
	lastInd := strings.LastIndex(artifact, ":")
	if lastInd == -1 {
		return artifact, ""
	}

	if lastInd == len(artifact)-1 {
		return artifact[:lastInd], ""
	}

	return artifact[:lastInd], artifact[lastInd+1:]
}

func ChartURLAndVersion(chart string) (url, version string) {
	path, version := Split(chart)
	return URL(path), version
}

func ChartPushURL(chart string) string {
	path, _ := Split(chart)
	path = filepath.Dir(path)
	return URL(path)
}

func URL(artifactPath string) string {
	return fmt.Sprintf("%s%s", OCIPrefix, artifactPath)
}

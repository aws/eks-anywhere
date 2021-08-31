package cluster

import (
	_ "embed"
	"strings"
)

//go:embed objects/coredns_clusterrole.yaml
var coreDNSClusterRole []byte

type KubeObjects map[string][]byte

func BuildExtraObjects(clusterSpec *Spec) KubeObjects {
	objects := make(map[string][]byte)
	if needExtraCoreDNSRole(clusterSpec.VersionsBundle) {
		objects["core-dns-clusterrole"] = coreDNSClusterRole
	}

	return objects
}

func needExtraCoreDNSRole(bundle *VersionsBundle) bool {
	return (bundle.KubeVersion == "1.20" || bundle.KubeVersion == "1.21") && imageVersion(bundle.KubeDistro.CoreDNS) == "v1.8.3"
}

func imageVersion(image VersionedRepository) string {
	tag := image.Tag
	if idx := strings.Index(tag, "-"); idx != -1 {
		return tag[:idx]
	}

	return tag
}

func (objs KubeObjects) Values() [][]byte {
	v := make([][]byte, 0, len(objs))
	for _, o := range objs {
		v = append(v, o)
	}

	return v
}

func (objs KubeObjects) Names() []string {
	v := make([]string, 0, len(objs))
	for n := range objs {
		v = append(v, n)
	}

	return v
}

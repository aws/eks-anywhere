//go:build !e2e

package e2e

import "testing"

func TestGitFluxMulticlusterIPPool(t *testing.T) {
	requirements, err := LoadIPPoolRequirements()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		want int
	}{
		{"TestVSphereKubernetes133MulticlusterWorkloadClusterGitHubFluxAPI", 4},
		{"TestVSphereKubernetes133MulticlusterWorkloadClusterGitFluxAPI", 4},
		{"TestVSphereKubernetes137MulticlusterWorkloadClusterGitHubFluxAPI", 4},
		{"TestVSphereKubernetes137MulticlusterWorkloadClusterGitFluxAPI", 4},
		{"TestVSphereKubernetes137MulticlusterWorkloadClusterAPI", 5},
		{"TestUnrelated", defaultIPPoolSize},
	} {
		if size := GetIPPoolSize(tt.name, requirements); size != tt.want {
			t.Errorf("%s reserves %d IPs, want %d", tt.name, size, tt.want)
		}
	}
}

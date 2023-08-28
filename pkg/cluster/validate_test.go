package cluster_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name string
		// Using testdata file here to avoid specifying structs in code that
		// we already have. If you need to test specific logic, create the structs
		// in this package to avoid testadat file explosion
		config *cluster.Config
	}{
		{
			name:   "vsphere cluster",
			config: clusterConfigFromFile(t, "testdata/cluster_1_19.yaml"),
		},
		{
			name:   "docker cluster gitops",
			config: clusterConfigFromFile(t, "testdata/docker_cluster_oidc_awsiam_gitops.yaml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(cluster.ValidateConfig(tt.config)).To(Succeed())
		})
	}
}

func clusterConfigFromFile(t *testing.T, path string) *cluster.Config {
	t.Helper()
	c, err := cluster.ParseConfigFromFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return c
}

func TestValidateConfigDifferentNamespace(t *testing.T) {
	g := NewWithT(t)
	c := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       anywherev1.ClusterKind,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eksa-unit-test",
				Namespace: "ns-1",
			},
		},
		VSphereDatacenter: &anywherev1.VSphereDatacenterConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       anywherev1.VSphereDatacenterKind,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eksa-unit-test",
				Namespace: "ns-2",
			},
		},
	}

	g.Expect(cluster.ValidateConfig(c)).To(
		MatchError(ContainSubstring("VSphereDatacenterConfig and Cluster objects must have the same namespace specified")),
	)
}

func TestValidateTinkerbellConfigDifferentNamespace(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name      string
		clusterNS string
		dcNS      string
		mcNS      string
		wantErr   string
	}{
		{
			name:      "cluster, datacenter config, machine config in the same namespace",
			clusterNS: constants.EksaSystemNamespace,
			dcNS:      constants.EksaSystemNamespace,
			mcNS:      constants.EksaSystemNamespace,
			wantErr:   "",
		},
		{
			name:      "datacenter config not at the same namespace",
			clusterNS: constants.EksaSystemNamespace,
			dcNS:      "dc-ns",
			mcNS:      constants.EksaSystemNamespace,
			wantErr:   "TinkerbellDatacenterConfig and Cluster objects must have the same namespace specified",
		},
		{
			name:      "machine config not at the same namespace",
			clusterNS: constants.EksaSystemNamespace,
			dcNS:      constants.EksaSystemNamespace,
			mcNS:      "mc-ns",
			wantErr:   "TinkerbellMachineConfig and Cluster objects must have the same namespace specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := clusterConfigFromFile(t, "testdata/cluster_tinkerbell_1_19.yaml")
			config.Cluster.Namespace = tt.clusterNS
			config.TinkerbellDatacenter.Namespace = tt.dcNS
			for _, mc := range config.TinkerbellMachineConfigs {
				mc.Namespace = tt.mcNS
			}
			if tt.wantErr != "" {
				g.Expect(cluster.ValidateConfig(config)).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				g.Expect(cluster.ValidateConfig(config)).To(BeNil())
			}
		})
	}
}

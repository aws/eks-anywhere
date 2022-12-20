package clusterapi_test

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

func TestSetControllerManagerArgs(t *testing.T) {
	tests := []struct {
		name        string
		clusterSpec *cluster.Spec
		want        clusterapi.ExtraArgs
	}{
		{
			name:        "without Node CIDR mask",
			clusterSpec: givenClusterSpec(),
			want:        map[string]string{"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
		},
		{
			name:        "with Node CIDR mask",
			clusterSpec: givenClusterSpecWithNodeCIDR(),
			want:        map[string]string{"node-cidr-mask-size": "28", "tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clusterapi.ControllerManagerArgs(tt.clusterSpec)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerManagerArgs()/%s got = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func givenClusterSpecWithNodeCIDR() *cluster.Spec {
	cluster := givenClusterSpec()
	nodeCidrMaskSize := new(int)
	*nodeCidrMaskSize = 28
	cluster.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
		Nodes: &v1alpha1.Nodes{CIDRMaskSize: nodeCidrMaskSize},
	}
	return cluster
}

func givenClusterSpec() *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "snow-test",
				Namespace: "test-namespace",
			},
			Spec: v1alpha1.ClusterSpec{
				ClusterNetwork: v1alpha1.ClusterNetwork{
					CNI: v1alpha1.Cilium,
					Pods: v1alpha1.Pods{
						CidrBlocks: []string{
							"10.1.0.0/16",
						},
					},
					Services: v1alpha1.Services{
						CidrBlocks: []string{
							"10.96.0.0/12",
						},
					},
				},
			},
		}
	})
}

func tlsCipherSuitesArgs() map[string]string {
	return map[string]string{"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}
}

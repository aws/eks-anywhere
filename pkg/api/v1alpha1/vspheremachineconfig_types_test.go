package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestVSphereMachineConfig_ResourcePaths(t *testing.T) {
	tests := []struct {
		name   string
		config VSphereMachineConfig
		want   map[string]string
	}{
		{
			name: "complete config",
			config: VSphereMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VSphereMachineConfig",
					APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-machine-config",
				},
				Spec: VSphereMachineConfigSpec{
					Folder:       "/dc/vm/folder1",
					Datastore:    "/dc/datastore/ds1",
					ResourcePool: "/dc/host/cluster1/Resources",
					Template:     "ubuntu-2204-kube-v1.27",
					NumCPUs:      2,
					MemoryMiB:    8192,
					OSFamily:     Ubuntu,
				},
			},
			want: map[string]string{
				"folder":       "/dc/vm/folder1",
				"datastore":    "/dc/datastore/ds1",
				"resourcePool": "/dc/host/cluster1/Resources",
				"template":     "ubuntu-2204-kube-v1.27",
			},
		},
		{
			name: "empty paths",
			config: VSphereMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VSphereMachineConfig",
					APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-paths-config",
				},
				Spec: VSphereMachineConfigSpec{
					Folder:       "",
					Datastore:    "",
					ResourcePool: "",
					Template:     "",
					NumCPUs:      2,
					MemoryMiB:    4096,
					OSFamily:     Bottlerocket,
				},
			},
			want: map[string]string{
				"folder":       "",
				"datastore":    "",
				"resourcePool": "",
				"template":     "",
			},
		},
		{
			name: "partial paths",
			config: VSphereMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VSphereMachineConfig",
					APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "partial-paths-config",
				},
				Spec: VSphereMachineConfigSpec{
					Folder:       "/dc/vm/folder2",
					Datastore:    "/dc/datastore/ds2",
					ResourcePool: "",
					Template:     "bottlerocket-kube-v1.28",
					NumCPUs:      4,
					MemoryMiB:    16384,
					OSFamily:     Bottlerocket,
				},
			},
			want: map[string]string{
				"folder":       "/dc/vm/folder2",
				"datastore":    "/dc/datastore/ds2",
				"resourcePool": "",
				"template":     "bottlerocket-kube-v1.28",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ResourcePaths()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VSphereMachineConfig.ResourcePaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

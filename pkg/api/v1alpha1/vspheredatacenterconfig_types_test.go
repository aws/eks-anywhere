package v1alpha1

import (
	"reflect"
	"testing"
)

func TestFailureDomain_ResourcePaths(t *testing.T) {
	tests := []struct {
		name          string
		failureDomain FailureDomain
		want          map[string]string
	}{
		{
			name: "complete failure domain",
			failureDomain: FailureDomain{
				Name:           "test-fd",
				ComputeCluster: "test-cluster",
				ResourcePool:   "test-pool",
				Datastore:      "test-datastore",
				Folder:         "test-folder",
				Network:        "test-network",
			},
			want: map[string]string{
				"computeCluster": "test-cluster",
				"resourcePool":   "test-pool",
				"datastore":      "test-datastore",
				"folder":         "test-folder",
			},
		},
		{
			name: "failure domain with empty values",
			failureDomain: FailureDomain{
				Name:           "empty-fd",
				ComputeCluster: "",
				ResourcePool:   "",
				Datastore:      "",
				Folder:         "",
				Network:        "",
			},
			want: map[string]string{
				"computeCluster": "",
				"resourcePool":   "",
				"datastore":      "",
				"folder":         "",
			},
		},
		{
			name: "failure domain with inventory paths",
			failureDomain: FailureDomain{
				Name:           "path-fd",
				ComputeCluster: "/dc/compute/cluster1",
				ResourcePool:   "/dc/compute/cluster1/Resources",
				Datastore:      "/dc/datastore/ds1",
				Folder:         "/dc/vm/folder1",
				Network:        "/dc/network/network1",
			},
			want: map[string]string{
				"computeCluster": "/dc/compute/cluster1",
				"resourcePool":   "/dc/compute/cluster1/Resources",
				"datastore":      "/dc/datastore/ds1",
				"folder":         "/dc/vm/folder1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.failureDomain.ResourcePaths()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FailureDomain.ResourcePaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

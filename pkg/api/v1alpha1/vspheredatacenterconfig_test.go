package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/features"
)

func TestGetVSphereDatacenterConfig(t *testing.T) {
	tests := []struct {
		testName              string
		fileName              string
		wantVSphereDatacenter *VSphereDatacenterConfig
		wantErr               bool
	}{
		{
			testName:              "file doesn't exist",
			fileName:              "testdata/fake_file.yaml",
			wantVSphereDatacenter: nil,
			wantErr:               true,
		},
		{
			testName:              "not parseable file",
			fileName:              "testdata/not_parseable_cluster.yaml",
			wantVSphereDatacenter: nil,
			wantErr:               true,
		},
		{
			testName: "valid 1.18",
			fileName: "testdata/cluster_1_18.yaml",
			wantVSphereDatacenter: &VSphereDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       VSphereDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: VSphereDatacenterConfigSpec{
					Datacenter: "myDatacenter",
					Network:    "myNetwork",
					Server:     "myServer",
					Thumbprint: "myTlsThumbprint",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19.yaml",
			wantVSphereDatacenter: &VSphereDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       VSphereDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: VSphereDatacenterConfigSpec{
					Datacenter: "myDatacenter",
					Network:    "myNetwork",
					Server:     "myServer",
					Thumbprint: "myTlsThumbprint",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters.yaml",
			wantVSphereDatacenter: &VSphereDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       VSphereDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: VSphereDatacenterConfigSpec{
					Datacenter: "myDatacenter",
					Network:    "myNetwork",
					Server:     "myServer",
					Thumbprint: "myTlsThumbprint",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20.yaml",
			wantVSphereDatacenter: &VSphereDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       VSphereDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: VSphereDatacenterConfigSpec{
					Datacenter: "myDatacenter",
					Network:    "myNetwork",
					Server:     "myServer",
					Thumbprint: "myTlsThumbprint",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName:              "invalid kind",
			fileName:              "testdata/cluster_invalid_kinds.yaml",
			wantVSphereDatacenter: nil,
			wantErr:               true,
		},
		{
			testName: "valid failure domain",
			fileName: "testdata/vsphere_cluster_valid_failuredomains.yaml",
			wantVSphereDatacenter: &VSphereDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       VSphereDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: VSphereDatacenterConfigSpec{
					Datacenter: "myDatacenter",
					Network:    "myNetwork",
					Server:     "myServer",
					Thumbprint: "myTlsThumbprint",
					Insecure:   false,
					FailureDomains: []FailureDomain{
						{
							Name:           "fd-1",
							ComputeCluster: "myComputeCluster",
							ResourcePool:   "myResourcePool",
							Datastore:      "myDatastore",
							Folder:         "myFolder",
							Network:        "/myDatacenter/network/myNetwork",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetVSphereDatacenterConfig(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetVSphereDatacenterConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantVSphereDatacenter) {
				t.Fatalf("GetVSphereDatacenterConfig() = %#v, want %#v", got, tt.wantVSphereDatacenter)
			}
		})
	}
}

func TestValidateVSphereDatacenterConfig(t *testing.T) {
	tests := []struct {
		testName              string
		expectedError         string
		modifyFunc            func(*VSphereDatacenterConfig)
	}{
		{
			testName:              "valid VSphereDatacenterConfig with FailureDomain",
			modifyFunc: func(v *VSphereDatacenterConfig) {
				v.Spec.FailureDomains = []FailureDomain{
					{
						Name:           "fd-1",
						ComputeCluster: "myComputeCluster",
						ResourcePool:   "myResourcePool",
						Datastore:      "myDatastore",
						Folder:         "myFolder",
						Network:        "/myDatacenter/network/myNetwork",
					},
				}
			},
		},
		{
			testName:              "Invalid VSphereDatacenterConfig with missing name in FailureDomain",
			modifyFunc: func(v *VSphereDatacenterConfig) {
				v.Spec.FailureDomains = []FailureDomain{
					{
						ComputeCluster: "myComputeCluster",
						ResourcePool:   "myResourcePool",
						Datastore:      "myDatastore",
						Folder:         "myFolder",
						Network:        "/myDatacenter/network/myNetwork",
					},
				}
			},
			expectedError: "name is not set or is empty",
		},
		{
			testName:              "Invalid VSphereDatacenterConfig with missing computeCluster in FailureDomain",
			modifyFunc: func(v *VSphereDatacenterConfig) {
				v.Spec.FailureDomains = []FailureDomain{
					{
						Name:         "fd-1",
						ResourcePool: "myResourcePool",
						Datastore:    "myDatastore",
						Folder:       "myFolder",
						Network:      "/myDatacenter/network/myNetwork",
					},
				}
			},
			expectedError: "computeCluster is not set or is empty",
		},
		{
			testName:              "Invalid VSphereDatacenterConfig with missing resourcePool in FailureDomain",
			modifyFunc: func(v *VSphereDatacenterConfig) {
				v.Spec.FailureDomains = []FailureDomain{
					{
						Name:           "fd-1",
						ComputeCluster: "myComputeCluster",
						Datastore:      "myDatastore",
						Folder:         "myFolder",
						Network:        "/myDatacenter/network/myNetwork",
					},
				}
			},
			expectedError: "resourcePool is not set or is empty",
		},
		{
			testName:              "Invalid VSphereDatacenterConfig with missing datastore in FailureDomain",
			modifyFunc: func(v *VSphereDatacenterConfig) {
				v.Spec.FailureDomains = []FailureDomain{
					{
						Name:           "fd-1",
						ComputeCluster: "myComputeCluster",
						ResourcePool:   "myResourcePool",
						Folder:         "myFolder",
						Network:        "/myDatacenter/network/myNetwork",
					},
				}
			},
			expectedError: "datastore is not set or is empty",
		},
		{
			testName:              "Invalid VSphereDatacenterConfig with missing folder in FailureDomain",
			modifyFunc: func(v *VSphereDatacenterConfig) {
				v.Spec.FailureDomains = []FailureDomain{
					{
						Name:           "fd-1",
						ComputeCluster: "myComputeCluster",
						ResourcePool:   "myResourcePool",
						Datastore:      "myDatastore",
						Network:        "/myDatacenter/network/myNetwork",
					},
				}
			},
			expectedError: "folder is not set or is empty",
		},
		{
			testName:              "Invalid VSphereDatacenterConfig with missing network in FailureDomain",
			modifyFunc: func(v *VSphereDatacenterConfig) {
				v.Spec.FailureDomains = []FailureDomain{
					{
						Name:           "fd-1",
						ComputeCluster: "myComputeCluster",
						ResourcePool:   "myResourcePool",
						Datastore:      "myDatastore",
						Folder:         "myFolder",
					},
				}
			},
			expectedError: "network is not set or is empty",
		},
		{
			testName:              "Invalid VSphereDatacenterConfig with invalid network in FailureDomain",
			modifyFunc: func(v *VSphereDatacenterConfig) {
				v.Spec.FailureDomains = []FailureDomain{
					{
						Name:           "fd-1",
						ComputeCluster: "myComputeCluster",
						ResourcePool:   "myResourcePool",
						Datastore:      "myDatastore",
						Folder:         "myFolder",
						Network:        "network",
					},
				}
			},
			expectedError: "invalid path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			vSphereDatacenter := generateVSphereDataCenterConfig()
			tt.modifyFunc(vSphereDatacenter)
			err := vSphereDatacenter.Validate()
			if tt.expectedError != "" {
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
		features.ClearCache()
	}
}

func generateVSphereDataCenterConfig() *VSphereDatacenterConfig {
	return &VSphereDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       VSphereDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "eksa-unit-test",
		},
		Spec: VSphereDatacenterConfigSpec{
			Datacenter: "myDatacenter",
			Network:    "/myDatacenter/network/myNetwork",
			Server:     "myServer",
			Thumbprint: "myTlsThumbprint",
			Insecure:   false,
			FailureDomains: []FailureDomain{
				{
					Name:           "fd-1",
					ComputeCluster: "myComputeCluster",
					ResourcePool:   "myResourcePool",
					Datastore:      "myDatastore",
					Folder:         "myFolder",
					Network:        "/myDatacenter/network/myNetwork",
				},
			},
		},
	}
}

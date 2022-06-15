package v1alpha1

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetCloudStackDatacenterConfig(t *testing.T) {
	tests := []struct {
		testName                 string
		fileName                 string
		wantCloudStackDatacenter *CloudStackDatacenterConfig
		wantErr                  bool
	}{
		{
			testName:                 "file doesn't exist",
			fileName:                 "testdata/fake_file.yaml",
			wantCloudStackDatacenter: nil,
			wantErr:                  true,
		},
		{
			testName:                 "not parseable file",
			fileName:                 "testdata/not_parseable_cluster_cloudstack.yaml",
			wantCloudStackDatacenter: nil,
			wantErr:                  true,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19_cloudstack.yaml",
			wantCloudStackDatacenter: &CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudStackDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudStackDatacenterConfigSpec{
					FailureDomains: []CloudStackFailureDomain{
						{
							Domain:  "domain1",
							Account: "admin",
							Zone: CloudStackZone{
								Name: "zone1",
								Network: CloudStackResourceIdentifier{
									Name: "net1",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.21",
			fileName: "testdata/cluster_1_21_cloudstack.yaml",
			wantCloudStackDatacenter: &CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudStackDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudStackDatacenterConfigSpec{
					FailureDomains: []CloudStackFailureDomain{
						{
							Domain:  "domain1",
							Account: "admin",
							Zone: CloudStackZone{
								Id: "zoneId",
								Network: CloudStackResourceIdentifier{
									Id: "netId",
								},
							},
							ManagementApiEndpoint: "https://127.0.0.1:8080/client/api",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters_cloudstack.yaml",
			wantCloudStackDatacenter: &CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudStackDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudStackDatacenterConfigSpec{
					FailureDomains: []CloudStackFailureDomain{
						{
							Domain:  "domain1",
							Account: "admin",
							Zone: CloudStackZone{
								Name: "zone1",
								Network: CloudStackResourceIdentifier{
									Name: "net1",
								},
							},
							ManagementApiEndpoint: "https://127.0.0.1:8080/client/api",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20_cloudstack.yaml",
			wantCloudStackDatacenter: &CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudStackDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudStackDatacenterConfigSpec{
					FailureDomains: []CloudStackFailureDomain{
						{
							Domain:  "domain1",
							Account: "admin",
							Zone: CloudStackZone{
								Name: "zone1",
								Network: CloudStackResourceIdentifier{
									Name: "net1",
								},
							},
							ManagementApiEndpoint: "https://127.0.0.1:8080/client/api",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName:                 "invalid kind",
			fileName:                 "testdata/cluster_invalid_kinds.yaml",
			wantCloudStackDatacenter: nil,
			wantErr:                  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetCloudStackDatacenterConfig(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCloudStackDatacenterConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantCloudStackDatacenter) {
				t.Fatalf("GetCloudStackDatacenterConfig() = %#v, want %#v", got, tt.wantCloudStackDatacenter)
			}
		})
	}
}

var cloudStackDatacenterConfigSpec1 = &CloudStackDatacenterConfigSpec{
	FailureDomains: []CloudStackFailureDomain{
		{
			Domain:  "domain1",
			Account: "admin",
			Zone: CloudStackZone{
				Name: "zone1",
				Network: CloudStackResourceIdentifier{
					Name: "net1",
				},
			},
			ManagementApiEndpoint: "testEndpoint",
		},
	},
}

func TestCloudStackDatacenterConfigSpecEqual(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpec1.DeepCopy()
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeTrue(), "deep copy CloudStackDatacenterConfigSpec showing as non-equal")
}

func TestCloudStackDatacenterConfigSpecNotEqualEndpoint(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpec1.DeepCopy()
	cloudStackDatacenterConfigSpec2.FailureDomains[0].ManagementApiEndpoint = "newEndpoint"
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "ManagementApiEndpoint comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualDomain(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpec1.DeepCopy()
	cloudStackDatacenterConfigSpec2.FailureDomains[0].Domain = "newDomain"
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "Domain comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualAccount(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpec1.DeepCopy()
	cloudStackDatacenterConfigSpec2.FailureDomains[0].Account = "newAccount"
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "Account comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualZonesNil(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpec1.DeepCopy()
	cloudStackDatacenterConfigSpec2.FailureDomains[0].Zone = CloudStackZone{}
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "Zones comparison in CloudStackDatacenterConfigSpec not detected")
}

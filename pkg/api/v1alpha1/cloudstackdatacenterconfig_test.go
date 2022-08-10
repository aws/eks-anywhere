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
					AvailabilityZones: []CloudStackAvailabilityZone{
						{
							Name: "default-az-0",
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
					AvailabilityZones: []CloudStackAvailabilityZone{{
						Name: "default-az-0",
						Domain:  "domain1",
						Account: "admin",
						Zone: CloudStackZone{
							Id: "zoneId",
							Network: CloudStackResourceIdentifier{
								Id: "netId",
							},
						},
						ManagementApiEndpoint: "https://127.0.0.1:8080/client/api",
					}},
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
					AvailabilityZones: []CloudStackAvailabilityZone{{
						Name: "default-az-0",
						Domain:  "domain1",
						Account: "admin",
						Zone: CloudStackZone{
							Name: "zone1",
							Network: CloudStackResourceIdentifier{
								Name: "net1",
							},
						},
						ManagementApiEndpoint: "https://127.0.0.1:8080/client/api",
					}},
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
					AvailabilityZones: []CloudStackAvailabilityZone{{
						Name: "default-az-0",
						Domain:  "domain1",
						Account: "admin",
						Zone: CloudStackZone{
							Name: "zone1",
							Network: CloudStackResourceIdentifier{
								Name: "net1",
							},
						},
						ManagementApiEndpoint: "https://127.0.0.1:8080/client/api",
					}},
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
	Domain:  "domain1",
	Account: "admin",
	Zones: []CloudStackZone{
		{
			Name: "zone1",
			Network: CloudStackResourceIdentifier{
				Name: "net1",
			},
		},
	},
	ManagementApiEndpoint: "testEndpoint",
}

var cloudStackDatacenterConfigSpecAzs = &CloudStackDatacenterConfigSpec{
	AvailabilityZones: []CloudStackAvailabilityZone{
		{
			Name:           "default-az-0",
			CredentialsRef: "global",
			Zone: CloudStackZone{
				Name: "zone1",
				Network: CloudStackResourceIdentifier{
					Name: "net1",
				},
			},
			Account:               "admin",
			Domain:                "domain1",
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
	cloudStackDatacenterConfigSpec2.ManagementApiEndpoint = "newEndpoint"
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "ManagementApiEndpoint comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualDomain(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpec1.DeepCopy()
	cloudStackDatacenterConfigSpec2.Domain = "newDomain"
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "Domain comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualAccount(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpec1.DeepCopy()
	cloudStackDatacenterConfigSpec2.Account = "newAccount"
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "Account comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualZonesNil(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpec1.DeepCopy()
	cloudStackDatacenterConfigSpec2.Zones = nil
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "Zones comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualAvailabilityZonesNil(t *testing.T) {
	g := NewWithT(t)
	g.Expect(cloudStackDatacenterConfigSpecAzs.AvailabilityZones[0].Equal(nil)).To(BeFalse(), "Zones comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualAvailabilityZonesEmpty(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpecAzs.DeepCopy()
	cloudStackDatacenterConfigSpec2.AvailabilityZones = []CloudStackAvailabilityZone{}
	g.Expect(cloudStackDatacenterConfigSpecAzs.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "AvailabilityZones comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualAvailabilityZonesModified(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpecAzs.DeepCopy()
	cloudStackDatacenterConfigSpec2.AvailabilityZones[0].Account = "differentAccount"
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "AvailabilityZones comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackAvailabilityZonesEqual(t *testing.T) {
	g := NewWithT(t)
	cloudStackAvailabilityZoneSpec2 := cloudStackDatacenterConfigSpecAzs.AvailabilityZones[0].DeepCopy()
	g.Expect(cloudStackDatacenterConfigSpecAzs.AvailabilityZones[0].Equal(cloudStackAvailabilityZoneSpec2)).To(BeTrue(), "AvailabilityZones comparison in CloudStackAvailabilityZoneSpec not detected")
}

func TestCloudStackAvailabilityZonesSame(t *testing.T) {
	g := NewWithT(t)
	g.Expect(cloudStackDatacenterConfigSpecAzs.AvailabilityZones[0].Equal(&cloudStackDatacenterConfigSpecAzs.AvailabilityZones[0])).To(BeTrue(), "AvailabilityZones comparison in CloudStackAvailabilityZoneSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualAvailabilityZonesManagementApiEndpoint(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpecAzs.DeepCopy()
	cloudStackDatacenterConfigSpec2.AvailabilityZones[0].ManagementApiEndpoint = "fake-endpoint"
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "AvailabilityZones comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualAvailabilityZonesAccount(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpecAzs.DeepCopy()
	cloudStackDatacenterConfigSpec2.AvailabilityZones[0].Account = "fake-acc"
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "AvailabilityZones comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSpecNotEqualAvailabilityZonesDomain(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfigSpec2 := cloudStackDatacenterConfigSpecAzs.DeepCopy()
	cloudStackDatacenterConfigSpec2.AvailabilityZones[0].Domain = "fake-domain"
	g.Expect(cloudStackDatacenterConfigSpec1.Equal(cloudStackDatacenterConfigSpec2)).To(BeFalse(), "AvailabilityZones comparison in CloudStackDatacenterConfigSpec not detected")
}

func TestCloudStackDatacenterConfigSetDefaults(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfig := CloudStackDatacenterConfig{
		Spec: *cloudStackDatacenterConfigSpec1.DeepCopy(),
	}
	cloudStackDatacenterConfig.SetDefaults()
	g.Expect(cloudStackDatacenterConfig.Spec.Equal(cloudStackDatacenterConfigSpecAzs)).To(BeTrue(), "AvailabilityZones comparison in CloudStackDatacenterConfigSpec not equal")
	g.Expect(len(cloudStackDatacenterConfigSpec1.Zones)).To(Equal(len(cloudStackDatacenterConfig.Spec.AvailabilityZones)), "AvailabilityZones count in CloudStackDatacenterConfigSpec not equal to zone count")
}

func TestCloudStackDatacenterConfigValidate(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfig := CloudStackDatacenterConfig{
		Spec: *cloudStackDatacenterConfigSpec1.DeepCopy(),
	}

	// Spec.Accout validation
	err := cloudStackDatacenterConfig.Validate()
	g.Expect(err).NotTo(BeNil())

	// Spec.Domain validation
	cloudStackDatacenterConfig.Spec.Account = ""
	err = cloudStackDatacenterConfig.Validate()
	g.Expect(err).NotTo(BeNil())

	// Spec.ManagementApiEndpoint validation
	cloudStackDatacenterConfig.Spec.Domain = ""
	err = cloudStackDatacenterConfig.Validate()
	g.Expect(err).NotTo(BeNil())

	// Spec.Zones validation
	cloudStackDatacenterConfig.Spec.ManagementApiEndpoint = ""
	err = cloudStackDatacenterConfig.Validate()
	g.Expect(err).NotTo(BeNil())

	// Spec.AvailabilityZones validation #1 (Length)
	cloudStackDatacenterConfig.Spec.Zones = []CloudStackZone{}
	err = cloudStackDatacenterConfig.Validate()
	g.Expect(err).NotTo(BeNil())
}

func TestCloudStackDatacenterConfigValidateAfterSetDefaults(t *testing.T) {
	g := NewWithT(t)
	cloudStackDatacenterConfig := CloudStackDatacenterConfig{
		Spec: *cloudStackDatacenterConfigSpec1.DeepCopy(),
	}

	cloudStackDatacenterConfig.SetDefaults()
	err := cloudStackDatacenterConfig.Validate()
	g.Expect(err).To(BeNil())

	// Spec.AvailabilityZones validation #2 (Name uniqueness)
	cloudStackDatacenterConfig.Spec.AvailabilityZones = append(cloudStackDatacenterConfig.Spec.AvailabilityZones, cloudStackDatacenterConfig.Spec.AvailabilityZones[0])
	err = cloudStackDatacenterConfig.Validate()
	g.Expect(err).NotTo(BeNil())
}

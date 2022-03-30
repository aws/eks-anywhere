package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetCloudStackMachineConfigs(t *testing.T) {
	tests := []struct {
		testName                     string
		fileName                     string
		wantCloudStackMachineConfigs map[string]*CloudStackMachineConfig
		wantErr                      bool
	}{
		{
			testName:                     "file doesn't exist",
			fileName:                     "testdata/fake_file.yaml",
			wantCloudStackMachineConfigs: nil,
			wantErr:                      true,
		},
		{
			testName:                     "not parseable file",
			fileName:                     "testdata/not_parseable_cluster.yaml",
			wantCloudStackMachineConfigs: nil,
			wantErr:                      true,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19_cloudstack.yaml",
			wantCloudStackMachineConfigs: map[string]*CloudStackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudStackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudStackMachineConfigSpec{
						Template: CloudStackResourceIdentifier{
							Name: "centos7-k8s-119",
						},
						ComputeOffering: CloudStackResourceIdentifier{
							Name: "m4-large",
						},
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.21",
			fileName: "testdata/cluster_1_21_cloudstack.yaml",
			wantCloudStackMachineConfigs: map[string]*CloudStackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudStackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudStackMachineConfigSpec{
						Template: CloudStackResourceIdentifier{
							Id: "centos7-k8s-121-id",
						},
						ComputeOffering: CloudStackResourceIdentifier{
							Id: "m4-large-id",
						},
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters_cloudstack.yaml",
			wantCloudStackMachineConfigs: map[string]*CloudStackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudStackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudStackMachineConfigSpec{
						Template: CloudStackResourceIdentifier{
							Name: "centos7-k8s-118",
						},
						ComputeOffering: CloudStackResourceIdentifier{
							Name: "m4-large",
						},
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20_cloudstack.yaml",
			wantCloudStackMachineConfigs: map[string]*CloudStackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudStackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudStackMachineConfigSpec{
						Template: CloudStackResourceIdentifier{
							Name: "centos7-k8s-120",
						},
						ComputeOffering: CloudStackResourceIdentifier{
							Name: "m4-large",
						},
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid different machine configs",
			fileName: "testdata/cluster_different_machine_configs_cloudstack.yaml",
			wantCloudStackMachineConfigs: map[string]*CloudStackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudStackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudStackMachineConfigSpec{
						Template: CloudStackResourceIdentifier{
							Name: "centos7-k8s-118",
						},
						ComputeOffering: CloudStackResourceIdentifier{
							Name: "m4-large",
						},
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
					},
				},
				"eksa-unit-test-2": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudStackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test-2",
					},
					Spec: CloudStackMachineConfigSpec{
						Template: CloudStackResourceIdentifier{
							Name: "centos7-k8s-118",
						},
						ComputeOffering: CloudStackResourceIdentifier{
							Name: "m5-xlarge",
						},
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
					},
				},
			},
			wantErr: false,
		},
		{
			testName:                     "invalid kind",
			fileName:                     "testdata/cluster_invalid_kinds.yaml",
			wantCloudStackMachineConfigs: nil,
			wantErr:                      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetCloudStackMachineConfigs(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCloudStackMachineConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantCloudStackMachineConfigs) {
				t.Fatalf("GetCloudStackMachineConfigs() = %#v, want %#v", got, tt.wantCloudStackMachineConfigs)
			}
		})
	}
}

func TestEqualCloudStackMachineConfigSpec(t *testing.T) {
	cloudStackMachineConfigSpec1 := &CloudStackMachineConfigSpec{
		Template: CloudStackResourceIdentifier{
			Name: "template1",
		},
		ComputeOffering: CloudStackResourceIdentifier{
			Name: "offering1",
		},
		Users: []UserConfiguration{
			{
				Name:              "zone1",
				SshAuthorizedKeys: []string{"key"},
			},
		},
		UserCustomDetails: map[string]string{
			"foo": "bar",
		},
		AffinityGroupIds: []string{"affinityGroupId1"},
	}
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	assert.True(t, cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2), "deep copy CloudStackMachineConfigSpec showing as non-equal")

	cloudStackMachineConfigSpec2.Template.Name = "newName"
	assert.False(t, cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2), "Template name comparison in CloudStackMachineConfigSpec not detected")
	cloudStackMachineConfigSpec2 = cloudStackMachineConfigSpec1.DeepCopy()

	cloudStackMachineConfigSpec2.Template.Id = "newId"
	assert.False(t, cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2), "Template id comparison in CloudStackMachineConfigSpec not detected")
	cloudStackMachineConfigSpec2 = cloudStackMachineConfigSpec1.DeepCopy()

	cloudStackMachineConfigSpec2.ComputeOffering.Name = "newComputeOffering"
	assert.False(t, cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2), "Compute offering name comparison in CloudStackMachineConfigSpec not detected")
	cloudStackMachineConfigSpec2 = cloudStackMachineConfigSpec1.DeepCopy()

	cloudStackMachineConfigSpec2.ComputeOffering.Id = "newComputeOffering"
	assert.False(t, cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2), "Compute offering id comparison in CloudStackMachineConfigSpec not detected")
	cloudStackMachineConfigSpec2 = cloudStackMachineConfigSpec1.DeepCopy()

	cloudStackMachineConfigSpec2.Users = nil
	assert.False(t, cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2), "Account comparison in CloudStackMachineConfigSpec not detected")
	cloudStackMachineConfigSpec2 = cloudStackMachineConfigSpec1.DeepCopy()

	cloudStackMachineConfigSpec2.Users = append(cloudStackMachineConfigSpec2.Users, UserConfiguration{Name: "newUser", SshAuthorizedKeys: []string{"newKey"}})
	assert.False(t, cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2), "Account comparison in CloudStackMachineConfigSpec not detected")
	cloudStackMachineConfigSpec2 = cloudStackMachineConfigSpec1.DeepCopy()

	cloudStackMachineConfigSpec2.UserCustomDetails = nil
	assert.False(t, cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2), "UserCustomDetails comparison in CloudStackMachineConfigSpec not detected")
	cloudStackMachineConfigSpec2 = cloudStackMachineConfigSpec1.DeepCopy()

	cloudStackMachineConfigSpec2.UserCustomDetails["i"] = "j"
	assert.False(t, cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2), "UserCustomDetails comparison in CloudStackMachineConfigSpec not detected")
}

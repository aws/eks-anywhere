package v1alpha1

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var cloudStackMachineConfigSpec1 = &CloudStackMachineConfigSpec{
	Template: CloudStackResourceIdentifier{
		Name: "template1",
	},
	ComputeOffering: CloudStackResourceIdentifier{
		Name: "offering1",
	},
	DiskOffering: &CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
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
	Symlinks: map[string]string{
		"/var/log/kubernetes": "/data/var/log/kubernetes",
	},
	AffinityGroupIds: []string{"affinityGroupId1"},
}

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
			testName:                     "non-splitable manifest",
			fileName:                     "testdata/invalid_manifest.yaml",
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
						DiskOffering: &CloudStackResourceDiskOffering{
							CloudStackResourceIdentifier: CloudStackResourceIdentifier{
								Name: "Small",
							},
							MountPath:  "/data-small",
							Device:     "/dev/vdb",
							Filesystem: "ext4",
							Label:      "data_disk",
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
						DiskOffering: &CloudStackResourceDiskOffering{
							CloudStackResourceIdentifier: CloudStackResourceIdentifier{
								Name: "Small",
							},
							MountPath:  "/data-small",
							Device:     "/dev/vdb",
							Filesystem: "ext4",
							Label:      "data_disk",
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
						DiskOffering: &CloudStackResourceDiskOffering{
							CloudStackResourceIdentifier: CloudStackResourceIdentifier{
								Name: "Medium",
							},
							MountPath:  "/data-medium",
							Device:     "/dev/vdb",
							Filesystem: "ext4",
							Label:      "data_disk",
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
	for i, tt := range tests {
		if i != 2 {
			continue
		}
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetCloudStackMachineConfigs(tt.fileName)
			fmt.Println(err)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCloudStackMachineConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantCloudStackMachineConfigs) {
				t.Fatalf("GetCloudStackMachineConfigs() = %#v, want %#v", got, tt.wantCloudStackMachineConfigs)
			}
		})
	}
}

func TestCloudStackMachineConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		obj     *CloudStackMachineConfig
		wantErr string
	}{
		{
			name: "valid config",
			obj: &CloudStackMachineConfig{
				Spec: *cloudStackMachineConfigSpec1,
			},
			wantErr: "",
		},
		{
			name: "disk offering empty",
			obj: &CloudStackMachineConfig{
				Spec: CloudStackMachineConfigSpec{
					Template: CloudStackResourceIdentifier{
						Name: "template1",
					},
					ComputeOffering: CloudStackResourceIdentifier{
						Name: "offering1",
					},
					DiskOffering: &CloudStackResourceDiskOffering{},
					Users: []UserConfiguration{
						{
							Name:              "zone1",
							SshAuthorizedKeys: []string{"key"},
						},
					},
					UserCustomDetails: map[string]string{
						"foo": "bar",
					},
					Symlinks: map[string]string{
						"/var/log/kubernetes": "/data/var/log/kubernetes",
					},
					AffinityGroupIds: []string{"affinityGroupId1"},
				},
			},
			wantErr: "",
		},
		{
			name: "invalid - bad mount path",
			obj: &CloudStackMachineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: CloudStackMachineConfigSpec{
					Template: CloudStackResourceIdentifier{
						Name: "template1",
					},
					ComputeOffering: CloudStackResourceIdentifier{
						Name: "offering1",
					},
					DiskOffering: &CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: CloudStackResourceIdentifier{
							Name: "diskOffering1",
						},
						MountPath:  "/",
						Device:     "/dev/vdb",
						Filesystem: "ext4",
						Label:      "data_disk",
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
					Symlinks: map[string]string{
						"/var/log/kubernetes": "/data/var/log/kubernetes",
					},
					Affinity: "pro",
				},
			},
			wantErr: "mountPath: / invalid, must be non-empty and start with /",
		},
		{
			name: "invalid - empty device",
			obj: &CloudStackMachineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: CloudStackMachineConfigSpec{
					Template: CloudStackResourceIdentifier{
						Name: "template1",
					},
					ComputeOffering: CloudStackResourceIdentifier{
						Name: "offering1",
					},
					DiskOffering: &CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: CloudStackResourceIdentifier{
							Name: "diskOffering1",
						},
						MountPath:  "/data",
						Device:     "",
						Filesystem: "ext4",
						Label:      "data_disk",
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
					Symlinks: map[string]string{
						"/var/log/kubernetes": "/data/var/log/kubernetes",
					},
					AffinityGroupIds: []string{"affinityGroupId1"},
				},
			},
			wantErr: "device:  invalid, empty device",
		},
		{
			name: "invalid - empty filesystem",
			obj: &CloudStackMachineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: CloudStackMachineConfigSpec{
					Template: CloudStackResourceIdentifier{
						Name: "template1",
					},
					ComputeOffering: CloudStackResourceIdentifier{
						Name: "offering1",
					},
					DiskOffering: &CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: CloudStackResourceIdentifier{
							Name: "diskOffering1",
						},
						MountPath:  "/data",
						Device:     "/dev/vdb",
						Filesystem: "",
						Label:      "data_disk",
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
					Symlinks: map[string]string{
						"/var/log/kubernetes": "/data/var/log/kubernetes",
					},
					Affinity: "pro",
				},
			},
			wantErr: "filesystem:  invalid, empty filesystem",
		},
		{
			name: "invalid - empty label",
			obj: &CloudStackMachineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: CloudStackMachineConfigSpec{
					Template: CloudStackResourceIdentifier{
						Name: "template1",
					},
					ComputeOffering: CloudStackResourceIdentifier{
						Name: "offering1",
					},
					DiskOffering: &CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: CloudStackResourceIdentifier{
							Name: "diskOffering1",
						},
						MountPath:  "/data",
						Device:     "/dev/vdb",
						Filesystem: "ext4",
						Label:      "",
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
					Symlinks: map[string]string{
						"/var/log/kubernetes": "/data/var/log/kubernetes",
					},
					AffinityGroupIds: []string{"affinityGroupId1"},
				},
			},
			wantErr: "label:  invalid, empty label",
		},
		{
			name: "invalid - restricted user details",
			obj: &CloudStackMachineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: CloudStackMachineConfigSpec{
					Template: CloudStackResourceIdentifier{
						Name: "template1",
					},
					ComputeOffering: CloudStackResourceIdentifier{
						Name: "offering1",
					},
					DiskOffering: &CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: CloudStackResourceIdentifier{
							Name: "diskOffering1",
						},
						MountPath:  "/data",
						Device:     "/dev/vdb",
						Filesystem: "ext4",
						Label:      "data_disk",
					},
					Users: []UserConfiguration{
						{
							Name:              "zone1",
							SshAuthorizedKeys: []string{"key"},
						},
					},
					UserCustomDetails: map[string]string{"keyboard": "test"},
					Symlinks: map[string]string{
						"/var/log/kubernetes": "/data/var/log/kubernetes",
					},
					AffinityGroupIds: []string{"affinityGroupId1"},
				},
			},
			wantErr: "restricted key keyboard found in custom user details",
		},
		{
			name: "bad affinity type",
			obj: &CloudStackMachineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: CloudStackMachineConfigSpec{
					Template: CloudStackResourceIdentifier{
						Name: "template1",
					},
					ComputeOffering: CloudStackResourceIdentifier{
						Name: "offering1",
					},
					DiskOffering: &CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: CloudStackResourceIdentifier{
							Name: "diskOffering1",
						},
						MountPath:  "/data",
						Device:     "/dev/vdb",
						Filesystem: "ext4",
						Label:      "data_disk",
					},
					Users: []UserConfiguration{
						{
							Name:              "zone1",
							SshAuthorizedKeys: []string{"key"},
						},
					},
					UserCustomDetails: map[string]string{"foo": "bar"},
					Symlinks: map[string]string{
						"/var/log/kubernetes": "/data/var/log/kubernetes",
					},
					Affinity: "xxx",
				},
			},
			wantErr: "invalid affinity type xxx for CloudStackMachineConfig test",
		},
		{
			name: "both affinity and affinityGroupIds are defined",
			obj: &CloudStackMachineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: CloudStackMachineConfigSpec{
					Template: CloudStackResourceIdentifier{
						Name: "template1",
					},
					ComputeOffering: CloudStackResourceIdentifier{
						Name: "offering1",
					},
					DiskOffering: &CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: CloudStackResourceIdentifier{
							Name: "diskOffering1",
						},
						MountPath:  "/data",
						Device:     "/dev/vdb",
						Filesystem: "ext4",
						Label:      "data_disk",
					},
					Users: []UserConfiguration{
						{
							Name:              "zone1",
							SshAuthorizedKeys: []string{"key"},
						},
					},
					UserCustomDetails: map[string]string{"foo": "bar"},
					Symlinks: map[string]string{
						"/var/log/kubernetes": "/data/var/log/kubernetes",
					},
					AffinityGroupIds: []string{"affinityGroupId1"},
					Affinity:         "pro",
				},
			},
			wantErr: "affinity and affinityGroupIds cannot be set at the same time for CloudStackMachineConfig test. Please provide either one of them or none",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := tt.obj.Validate()
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestCloudStackMachineConfigSpecEqual(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeTrue(), "deep copy CloudStackMachineConfigSpec showing as non-equal")
}

func TestCloudStackMachineNotEqualTemplateName(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.Template.Name = "newName"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Template name comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualTemplateId(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.Template.Id = "newId"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Template id comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualComputeOfferingName(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.ComputeOffering.Name = "newComputeOffering"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Compute offering name comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualComputeOfferingId(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.ComputeOffering.Id = "newComputeOffering"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Compute offering id comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualDiskOfferingName(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering = (*cloudStackMachineConfigSpec1.DiskOffering).DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering.Name = "newDiskOffering"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Disk offering name comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualDiskOfferingId(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering = (*cloudStackMachineConfigSpec1.DiskOffering).DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering.Id = "newDiskOffering"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Disk offering id comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualDiskOfferingMountPath(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering = (*cloudStackMachineConfigSpec1.DiskOffering).DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering.MountPath = "newDiskOfferingPath"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Disk offering path comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualDiskOfferingDevice(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering = (*cloudStackMachineConfigSpec1.DiskOffering).DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering.Device = "/dev/sdb"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Disk offering device comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualDiskOfferingLabel(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering = (*cloudStackMachineConfigSpec1.DiskOffering).DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering.Label = "data_disk_new"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Disk offering label comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualDiskOfferingFilesystem(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering = (*cloudStackMachineConfigSpec1.DiskOffering).DeepCopy()
	cloudStackMachineConfigSpec2.DiskOffering.Filesystem = "ext3"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Disk offering filesystem comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualAffinity(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.Affinity = "anti"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Affinity comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualUsersNil(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.Users = nil
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Account comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualUsers(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.Users = append(cloudStackMachineConfigSpec2.Users, UserConfiguration{Name: "newUser", SshAuthorizedKeys: []string{"newKey"}})
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Account comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualUserCustomDetailsNil(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.UserCustomDetails = nil
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "UserCustomDetails comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualSymlinksNil(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.Symlinks = nil
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Symlinks comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualUserCustomDetails(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.UserCustomDetails["i"] = "j"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "UserCustomDetails comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualSymlinks(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	cloudStackMachineConfigSpec2.Symlinks["i"] = "j"
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Symlinks comparison in CloudStackMachineConfigSpec not detected")
}

func TestCloudStackMachineNotEqualSymlinksDifferentTargetSameKey(t *testing.T) {
	g := NewWithT(t)
	cloudStackMachineConfigSpec2 := cloudStackMachineConfigSpec1.DeepCopy()
	for k, v := range cloudStackMachineConfigSpec2.Symlinks {
		cloudStackMachineConfigSpec2.Symlinks[k] = "/different" + v
	}
	g.Expect(cloudStackMachineConfigSpec1.Equal(cloudStackMachineConfigSpec2)).To(BeFalse(), "Symlinks comparison in CloudStackMachineConfigSpec not detected")
}

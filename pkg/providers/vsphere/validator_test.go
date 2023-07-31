package vsphere

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/govmomi"
	"github.com/aws/eks-anywhere/pkg/govmomi/mocks"
	govcmocks "github.com/aws/eks-anywhere/pkg/providers/vsphere/mocks"
)

func TestValidatorValidatePrivs(t *testing.T) {
	v := Validator{}

	ctrl := gomock.NewController(t)
	vsc := mocks.NewMockVSphereClient(ctrl)

	ctx := context.Background()
	networkPath := "/Datacenter/network/path/foo"

	objects := []PrivAssociation{
		{
			objectType:   govmomi.VSphereTypeNetwork,
			privsContent: config.VSphereUserPrivsFile,
			path:         networkPath,
		},
	}

	var privs []string
	err := json.Unmarshal([]byte(config.VSphereAdminPrivsFile), &privs)
	if err != nil {
		t.Fatalf("failed to validate privs: %v", err)
	}
	vsc.EXPECT().Username().Return("foobar")
	vsc.EXPECT().GetPrivsOnEntity(ctx, networkPath, govmomi.VSphereTypeNetwork, "foobar").Return(privs, nil)

	passed, err := v.validatePrivs(ctx, objects, vsc)
	if passed != true || err != nil {
		t.Fatalf("failed to validate privs passed=%v, err=%v", passed, err)
	}
}

func TestValidatorValidatePrivsError(t *testing.T) {
	v := Validator{}

	ctrl := gomock.NewController(t)
	vsc := mocks.NewMockVSphereClient(ctrl)

	ctx := context.Background()
	networkPath := "/Datacenter/network/path/foo"

	objects := []PrivAssociation{
		{
			objectType:   govmomi.VSphereTypeNetwork,
			privsContent: config.VSphereUserPrivsFile,
			path:         networkPath,
		},
	}

	var privs []string
	err := json.Unmarshal([]byte(config.VSphereAdminPrivsFile), &privs)
	if err != nil {
		t.Fatalf("failed to validate privs: %v", err)
	}
	errMsg := "Could not retrieve privs"
	g := NewWithT(t)
	vsc.EXPECT().Username().Return("foobar")
	vsc.EXPECT().GetPrivsOnEntity(ctx, networkPath, govmomi.VSphereTypeNetwork, "foobar").Return(nil, fmt.Errorf(errMsg))

	_, err = v.validatePrivs(ctx, objects, vsc)
	g.Expect(err).To(MatchError(ContainSubstring(errMsg)))
}

func TestValidatorValidatePrivsMissing(t *testing.T) {
	v := Validator{}

	ctrl := gomock.NewController(t)
	vsc := mocks.NewMockVSphereClient(ctrl)

	ctx := context.Background()
	folderPath := "/Datacenter/vm/path/foo"

	objects := []PrivAssociation{
		{
			objectType:   govmomi.VSphereTypeFolder,
			privsContent: config.VSphereAdminPrivsFile,
			path:         folderPath,
		},
	}

	var privs []string
	err := json.Unmarshal([]byte(config.VSphereUserPrivsFile), &privs)
	if err != nil {
		t.Fatalf("failed to validate privs: %v", err)
	}
	g := NewWithT(t)
	vsc.EXPECT().Username().Return("foobar")
	vsc.EXPECT().GetPrivsOnEntity(ctx, folderPath, govmomi.VSphereTypeFolder, "foobar").Return(privs, nil)

	passed, err := v.validatePrivs(ctx, objects, vsc)

	g.Expect(passed).To(BeEquivalentTo(false))
	g.Expect(err).NotTo(BeNil())
}

func TestValidatorValidatePrivsBadJson(t *testing.T) {
	v := Validator{}

	ctrl := gomock.NewController(t)
	vsc := mocks.NewMockVSphereClient(ctrl)
	vsc.EXPECT().Username().Return("foobar")

	ctx := context.Background()
	networkPath := "/Datacenter/network/path/foo"
	g := NewWithT(t)
	errMsg := "invalid character 'h' in literal true (expecting 'r')"

	objects := []PrivAssociation{
		{
			objectType:   govmomi.VSphereTypeNetwork,
			privsContent: "this is bad json",
			path:         networkPath,
		},
	}

	_, err := v.validatePrivs(ctx, objects, vsc)
	g.Expect(err).To(MatchError(ContainSubstring(errMsg)))
}

func clusterSpec() Spec {
	cpMachineConfig := &v1alpha1.VSphereMachineConfig{
		Spec: v1alpha1.VSphereMachineConfigSpec{
			Datastore:    "datastore",
			ResourcePool: "pool",
			Folder:       "folder",
			Template:     "temp",
		},
	}

	return Spec{
		Spec: &cluster.Spec{
			Config: &cluster.Config{
				VSphereDatacenter: &v1alpha1.VSphereDatacenterConfig{
					Spec: v1alpha1.VSphereDatacenterConfigSpec{
						Datacenter: "SDDC-Datacenter",
						Server:     "server",
					},
				},
				VSphereMachineConfigs: map[string]*v1alpha1.VSphereMachineConfig{
					"test-cp": cpMachineConfig,
				},
				Cluster: &v1alpha1.Cluster{
					Spec: v1alpha1.ClusterSpec{
						ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
							MachineGroupRef: &v1alpha1.Ref{
								Name: "test-cp",
							},
						},
					},
				},
			},
		},
	}
}

func TestValidatorValidateVsphereUserPrivsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	govc := govcmocks.NewMockProviderGovcClient(ctrl)
	vscb := govcmocks.NewMockVSphereClientBuilder(ctrl)

	v := Validator{
		govc:                 govc,
		vSphereClientBuilder: vscb,
	}

	spec := clusterSpec()

	ctx := context.Background()
	vscb.EXPECT().Build(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), spec.VSphereDatacenter.Spec.Datacenter).Return(nil, fmt.Errorf("error"))

	g := NewWithT(t)

	err := v.validateVsphereUserPrivs(ctx, &spec)
	g.Expect(err).To(MatchError(ContainSubstring("error")))
}

func TestValidatorValidateVsphereCPUserPrivsError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	govc := govcmocks.NewMockProviderGovcClient(ctrl)
	vscb := govcmocks.NewMockVSphereClientBuilder(ctrl)
	vsc := mocks.NewMockVSphereClient(ctrl)

	wantEnv := map[string]string{
		config.EksavSphereUsernameKey:   "foo",
		config.EksavSpherePasswordKey:   "bar",
		config.EksavSphereCPUsernameKey: "foo2",
		config.EksavSphereCPPasswordKey: "bar2",
	}
	for k, v := range wantEnv {
		t.Setenv(k, v)
	}

	v := Validator{
		govc:                 govc,
		vSphereClientBuilder: vscb,
	}

	var privs []string
	err := json.Unmarshal([]byte(config.VSphereAdminPrivsFile), &privs)
	if err != nil {
		t.Fatalf("failed to validate privs: %v", err)
	}

	spec := clusterSpec()

	vsc.EXPECT().Username().Return("foobar").AnyTimes()
	vsc.EXPECT().GetPrivsOnEntity(ctx, gomock.Any(), gomock.Any(), "foobar").Return(privs, nil).AnyTimes()

	vscb.EXPECT().Build(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), spec.VSphereDatacenter.Spec.Datacenter).Return(vsc, nil)
	vscb.EXPECT().Build(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), spec.VSphereDatacenter.Spec.Datacenter).Return(nil, fmt.Errorf("error"))
	g := NewWithT(t)

	err = v.validateVsphereUserPrivs(ctx, &spec)
	g.Expect(err).To(MatchError(ContainSubstring("error")))
}

func TestValidatorValidateMachineConfigTagsExistErrorListingTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	govc := govcmocks.NewMockProviderGovcClient(ctrl)
	ctx := context.Background()
	g := NewWithT(t)

	v := Validator{
		govc: govc,
	}

	machineConfigs := []*v1alpha1.VSphereMachineConfig{
		{
			Spec: v1alpha1.VSphereMachineConfigSpec{
				TagIDs: []string{"tag-1", "tag-2"},
			},
		},
	}

	govc.EXPECT().ListTags(ctx).Return(nil, errors.New("error listing tags"))

	err := v.validateMachineConfigTagsExist(ctx, machineConfigs)
	g.Expect(err).To(Not(BeNil()))
}

func TestValidatorValidateMachineConfigTagsExistSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	govc := govcmocks.NewMockProviderGovcClient(ctrl)
	ctx := context.Background()
	g := NewWithT(t)

	v := Validator{
		govc: govc,
	}

	machineConfigs := []*v1alpha1.VSphereMachineConfig{
		{
			Spec: v1alpha1.VSphereMachineConfigSpec{
				TagIDs: []string{"tag-1", "tag-2"},
			},
		},
	}

	tagIDs := []executables.Tag{
		{
			Id: "tag-1",
		},
		{
			Id: "tag-2",
		},
		{
			Id: "tag-3",
		},
	}

	govc.EXPECT().ListTags(ctx).Return(tagIDs, nil)

	err := v.validateMachineConfigTagsExist(ctx, machineConfigs)
	g.Expect(err).To(BeNil())
}

func TestValidatorValidateMachineConfigTagsExistTagDoesNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	govc := govcmocks.NewMockProviderGovcClient(ctrl)
	ctx := context.Background()
	g := NewWithT(t)

	v := Validator{
		govc: govc,
	}

	machineConfigs := []*v1alpha1.VSphereMachineConfig{
		{
			Spec: v1alpha1.VSphereMachineConfigSpec{
				TagIDs: []string{"tag-1", "tag-2"},
			},
		},
	}

	tagIDs := []executables.Tag{
		{
			Id: "tag-1",
		},
		{
			Id: "tag-3",
		},
	}

	govc.EXPECT().ListTags(ctx).Return(tagIDs, nil)

	err := v.validateMachineConfigTagsExist(ctx, machineConfigs)
	g.Expect(err).To(Not(BeNil()))
}

func TestValidateBRHardDiskSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	govc := govcmocks.NewMockProviderGovcClient(ctrl)
	ctx := context.Background()

	v := Validator{
		govc: govc,
	}

	machineConfig := v1alpha1.VSphereMachineConfig{
		Spec: v1alpha1.VSphereMachineConfigSpec{
			Template: "bottlerocket-kube-v1-21",
		},
	}
	spec := Spec{
		Spec: &cluster.Spec{
			Config: &cluster.Config{
				VSphereDatacenter: &v1alpha1.VSphereDatacenterConfig{
					Spec: v1alpha1.VSphereDatacenterConfigSpec{
						Datacenter: "SDDC-Datacenter",
					},
				},
			},
		},
	}
	govcErr := errors.New("error GetHardDiskSize()")
	tests := []struct {
		testName      string
		returnDiskMap map[string]float64
		ifErr         error
		wantErr       error
	}{
		{
			testName:      "getHardDiskSize_govc_error",
			returnDiskMap: map[string]float64{},
			ifErr:         govcErr,
			wantErr:       fmt.Errorf("validating hard disk size: %v", govcErr),
		},
		{
			testName:      "getHardDiskSize_empty_map_error",
			returnDiskMap: map[string]float64{},
			ifErr:         nil,
			wantErr:       fmt.Errorf("no hard disks found for template: %v", "bottlerocket-kube-v1-21"),
		},
		{
			testName:      "check_disk1_wrong_size",
			returnDiskMap: map[string]float64{"Hard disk 1": 100, "Hard disk 2": 20971520},
			ifErr:         nil,
			wantErr:       fmt.Errorf("Incorrect disk size for disk1 - expected: 2097152 kB got: %v", 100),
		},
		{
			testName:      "check_disk2_wrong_size",
			returnDiskMap: map[string]float64{"Hard disk 1": 2097152, "Hard disk 2": 100},
			ifErr:         nil,
			wantErr:       fmt.Errorf("Incorrect disk size for disk2 - expected: 20971520 kB got: %v", 100),
		},
		{
			testName:      "check_singleDisk_wrong_size",
			returnDiskMap: map[string]float64{"Hard disk 1": 100},
			ifErr:         nil,
			wantErr:       fmt.Errorf("Incorrect disk size for disk1 - expected: 23068672 kB got: %v", 100),
		},
		{
			testName:      "check_happy_flow",
			returnDiskMap: map[string]float64{"Hard disk 1": 2097152, "Hard disk 2": 20971520},
			ifErr:         nil,
			wantErr:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			gt := NewWithT(t)
			govc.EXPECT().GetHardDiskSize(ctx, machineConfig.Spec.Template, spec.Config.VSphereDatacenter.Spec.Datacenter).Return(tt.returnDiskMap, tt.ifErr)
			err := v.validateBRHardDiskSize(ctx, &spec, &machineConfig)
			if err == nil {
				gt.Expect(err).To(BeNil())
			} else {
				gt.Expect(err.Error()).To(Equal(tt.wantErr.Error()))
			}
		})
	}
}

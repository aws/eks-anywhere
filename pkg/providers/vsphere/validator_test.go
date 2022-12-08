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
	g.Expect(err).To(BeNil())
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

package setupuser_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/vsphere/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/setupuser"
)

func TestSetupUserRun(t *testing.T) {
	ctx := context.Background()

	defaults := struct {
		GroupExists      bool
		GlobalRoleExists bool
		UserRoleExists   bool
		AdminRoleExists  bool
	}{
		GroupExists:      false,
		GlobalRoleExists: false,
		UserRoleExists:   false,
		AdminRoleExists:  false,
	}

	tests := []struct {
		name     string
		filepath string
		wantErr  string
	}{
		{
			name:     "test generateconfig read file happy path",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			c, err := setupuser.GenerateConfig(ctx, tt.filepath)
			if err != nil {
				t.Fatalf("failed to generate config from %s with %s", tt.filepath, err)
			}
			ctrl := gomock.NewController(t)
			gc := mocks.NewMockProviderGovcClient(ctrl)
			gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
			gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
			gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
			gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(defaults.GlobalRoleExists, nil)
			gc.EXPECT().CreateRole(ctx, c.Spec.GlobalRole, gomock.Any()).Return(nil)
			gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(defaults.UserRoleExists, nil)
			gc.EXPECT().CreateRole(ctx, c.Spec.UserRole, gomock.Any()).Return(nil)
			gc.EXPECT().RoleExists(ctx, c.Spec.AdminRole).Return(defaults.AdminRoleExists, nil)
			gc.EXPECT().CreateRole(ctx, c.Spec.AdminRole, gomock.Any()).Return(nil)

			gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.GlobalRole, "/", c.Spec.VSphereDomain)

			gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.AdminRole, c.Spec.Objects.Folders[0], c.Spec.VSphereDomain)
			gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.AdminRole, c.Spec.Objects.Templates[0], c.Spec.VSphereDomain)

			gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.UserRole, c.Spec.Objects.Networks[0], c.Spec.VSphereDomain)
			gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.UserRole, c.Spec.Objects.Datastores[0], c.Spec.VSphereDomain)
			gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.UserRole, c.Spec.Objects.ResourcePools[0], c.Spec.VSphereDomain)
			err = setupuser.Run(ctx, c, gc)

			if tt.wantErr == "" {
				g.Expect(err).To(Succeed())
				g.Expect(c).ToNot(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		},
		)
	}
}

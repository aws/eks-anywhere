package setupuser_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/vsphere/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/setupuser"
)

type testConfig struct {
	GroupExists      bool
	GlobalRoleExists bool
	UserRoleExists   bool
	AdminRoleExists  bool
}

func TestSetupUserRun(t *testing.T) {
	defaults := testConfig{
		GroupExists:      false,
		GlobalRoleExists: false,
		UserRoleExists:   false,
		AdminRoleExists:  false,
	}

	tests := []struct {
		name     string
		filepath string
		wantErr  string
		prepare  func(context.Context, *setupuser.VSphereSetupUserConfig, *mocks.MockProviderGovcClient, testConfig)
	}{
		{
			name:     "test generateconfig read file happy path",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockProviderGovcClient, defaults testConfig) {
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
			},
		},
		{
			name:     "test GroupExists error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "failed to connect to govc",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockProviderGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, fmt.Errorf("failed to connect to govc"))
			},
		},
		{
			name:     "test CreateGroup error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "failed to create group",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockProviderGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(fmt.Errorf("failed to create group"))
			},
		},
		{
			name:     "test AddUserToGroup error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "failed to add user to group",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockProviderGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(fmt.Errorf("failed to add user to group"))
			},
		},
		{
			name:     "test RoleExists GlobalRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockProviderGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(defaults.GlobalRoleExists, fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test RoleExists GlobalRole true",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockProviderGovcClient, defaults testConfig) {
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
			},
		},
		{
			name:     "test RoleExists UserRole true",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockProviderGovcClient, defaults testConfig) {
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
			},
		},
		{
			name:     "test RoleExists AdminRole true",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockProviderGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)

				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(true, nil)

				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(true, nil)

				gc.EXPECT().RoleExists(ctx, c.Spec.AdminRole).Return(true, nil)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.GlobalRole, "/", c.Spec.VSphereDomain)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.AdminRole, c.Spec.Objects.Folders[0], c.Spec.VSphereDomain)
				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.AdminRole, c.Spec.Objects.Templates[0], c.Spec.VSphereDomain)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.UserRole, c.Spec.Objects.Networks[0], c.Spec.VSphereDomain)
				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.UserRole, c.Spec.Objects.Datastores[0], c.Spec.VSphereDomain)
				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.UserRole, c.Spec.Objects.ResourcePools[0], c.Spec.VSphereDomain)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			ctx := context.Background()

			c, err := setupuser.GenerateConfig(ctx, tt.filepath)
			if err != nil {
				t.Fatalf("failed to generate config from %s with %s", tt.filepath, err)
			}
			ctrl := gomock.NewController(t)
			gc := mocks.NewMockProviderGovcClient(ctrl)
			tt.prepare(ctx, c, gc, defaults)

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

func TestNewVsphereUserConfig(t *testing.T) {
	filepath := "./testdata/configs/valid.yaml"
	ctx := context.Background()
	c, err := setupuser.GenerateConfig(ctx, filepath)
	if err != nil {
		t.Fatal(err)
	}
	wantEnv := map[string]string{
		"GOVC_URL":        c.Spec.Connection.Server,
		"GOVC_INSECURE":   "false",
		"GOVC_DATACENTER": c.Spec.Datacenter,
	}

	err = setupuser.SetupGOVCEnv(ctx, c)
	if err != nil {
		t.Fatal(err)
	}
	g := NewWithT(t)

	for k, want := range wantEnv {
		v := os.Getenv(k)
		g.Expect(v).To(BeIdenticalTo(want))
	}
}

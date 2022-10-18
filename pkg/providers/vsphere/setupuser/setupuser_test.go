package setupuser_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/vsphere/setupuser"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/setupuser/mocks"
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
		prepare  func(context.Context, *setupuser.VSphereSetupUserConfig, *mocks.MockGovcClient, testConfig)
	}{
		{
			name:     "test setup vsphere user happy path",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
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
			name:     "test setup vsphere user happy path group exists",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(true, nil)
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
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, fmt.Errorf("failed to connect to govc"))
			},
		},
		{
			name:     "test CreateGroup error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "failed to create group",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(fmt.Errorf("failed to create group"))
			},
		},
		{
			name:     "test AddUserToGroup error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "failed to add user to group",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(fmt.Errorf("failed to add user to group"))
			},
		},
		{
			name:     "test RoleExists GlobalRole true",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
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
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
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
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
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
		{
			name:     "test RoleExists GlobalRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(defaults.GlobalRoleExists, fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test RoleExists UserRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(false, fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test RoleExists AdminRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.AdminRole).Return(false, fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test createRole GlobalRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(false, nil)
				gc.EXPECT().CreateRole(ctx, c.Spec.GlobalRole, gomock.Any()).Return(fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test createRole UserRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(false, nil)
				gc.EXPECT().CreateRole(ctx, c.Spec.UserRole, gomock.Any()).Return(fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test createRole AdminRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.AdminRole).Return(false, nil)
				gc.EXPECT().CreateRole(ctx, c.Spec.AdminRole, gomock.Any()).Return(fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test SetGroupRoleOnObject GlobalRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.AdminRole).Return(true, nil)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.GlobalRole, "/", c.Spec.VSphereDomain).Return(fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test SetGroupRoleOnObject AdminRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.AdminRole).Return(true, nil)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.GlobalRole, "/", c.Spec.VSphereDomain).Return(nil)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.AdminRole, c.Spec.Objects.Folders[0], c.Spec.VSphereDomain).Return(fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test SetGroupRoleOnObject UserRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.AdminRole).Return(true, nil)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.GlobalRole, "/", c.Spec.VSphereDomain).Return(nil)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.AdminRole, c.Spec.Objects.Folders[0], c.Spec.VSphereDomain).Return(nil)
				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.AdminRole, c.Spec.Objects.Templates[0], c.Spec.VSphereDomain)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.UserRole, c.Spec.Objects.Networks[0], c.Spec.VSphereDomain).Return(fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test SetGroupRoleOnObject UserRole exists",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient, defaults testConfig) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(defaults.GroupExists, nil)
				gc.EXPECT().CreateGroup(ctx, c.Spec.GroupName).Return(nil)
				gc.EXPECT().AddUserToGroup(ctx, c.Spec.GroupName, c.Spec.Username).Return(nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(true, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.AdminRole).Return(true, nil)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.GlobalRole, "/", c.Spec.VSphereDomain).Return(nil)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.AdminRole, c.Spec.Objects.Folders[0], c.Spec.VSphereDomain).Return(nil)
				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.AdminRole, c.Spec.Objects.Templates[0], c.Spec.VSphereDomain)

				gc.EXPECT().SetGroupRoleOnObject(ctx, c.Spec.GroupName, c.Spec.UserRole, c.Spec.Objects.Networks[0], c.Spec.VSphereDomain).Return(fmt.Errorf("govc error"))
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
			gc := mocks.NewMockGovcClient(ctrl)
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

func TestSetupGOVCEnv(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		filepath string
		wantErr  string
		prepare  func(context.Context, *setupuser.VSphereSetupUserConfig) map[string]string
	}{
		{
			name:     "test SetupGOVCEnv happy path",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig) map[string]string {
				wantEnv := map[string]string{
					"GOVC_URL":        c.Spec.Connection.Server,
					"GOVC_INSECURE":   "false",
					"GOVC_DATACENTER": c.Spec.Datacenter,
				}
				return wantEnv
			},
		},
		{
			name:     "test SetupGOVCEnv happy path insecure=true",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig) map[string]string {
				c.Spec.Connection.Insecure = true

				wantEnv := map[string]string{
					"GOVC_URL":        c.Spec.Connection.Server,
					"GOVC_INSECURE":   "true",
					"GOVC_DATACENTER": c.Spec.Datacenter,
				}
				return wantEnv
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			c, err := setupuser.GenerateConfig(ctx, tt.filepath)
			if err != nil {
				t.Fatalf("failed to generate config from %s with %s", tt.filepath, err)
			}
			wantEnv := tt.prepare(ctx, c)

			err = setupuser.SetupGOVCEnv(ctx, c)
			if len(tt.wantErr) > 0 {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}

			for k, want := range wantEnv {
				v := os.Getenv(k)
				g.Expect(v).To(BeIdenticalTo(want))
			}

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

package setupuser_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/vsphere/setupuser"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/setupuser/mocks"
)

func TestGenerateConfigReadFile(t *testing.T) {
	ctx := context.Background()

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
		{
			name:     "test generateconfig read file bad yaml",
			filepath: "./testdata/configs/not_yaml.yaml",
			wantErr:  "failed to parse ./testdata/configs/not_yaml.yaml, err = yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `this is...` ",
		},
		{
			name:     "test generateconfig read file does not exist",
			filepath: "./testdata/configs/not_a_file.yaml",
			wantErr:  "failed to read file ./testdata/configs/not_a_file.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			c, err := setupuser.GenerateConfig(ctx, tt.filepath)

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

func TestGenerateConfigValidations(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		filepath string
		wantErr  string
	}{
		{
			name:     "test read file happy path",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
		},
		{
			name:     "test validate datacenter not empty",
			filepath: "./testdata/configs/empty.yaml",
			wantErr:  "datacenter cannot be empty",
		},
		{
			name:     "test validate vspheredomain not empty",
			filepath: "./testdata/configs/empty.yaml",
			wantErr:  "vSphereDomain cannot be empty",
		},
		{
			name:     "test validate connection",
			filepath: "./testdata/configs/empty.yaml",
			wantErr:  "server cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			c, err := setupuser.GenerateConfig(ctx, tt.filepath)

			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
				g.Expect(c).ToNot(BeNil())
			} else {
				g.Expect(err.Error()).To(ContainSubstring(tt.wantErr))
			}
		},
		)
	}
}

func TestGenerateConfigSetDefaults(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		filepath string
	}{
		{
			name:     "test populating config with defaults happy path",
			filepath: "./testdata/configs/valid_minimal.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			c, err := setupuser.GenerateConfig(ctx, tt.filepath)
			g.Expect(err).To(Succeed())

			g.Expect(c.Spec.Username).To(Equal(setupuser.DefaultUsername))
			g.Expect(c.Spec.GroupName).To(Equal(setupuser.DefaultGroup))
			g.Expect(c.Spec.GlobalRole).To(Equal(setupuser.DefaultGlobalRole))
			g.Expect(c.Spec.UserRole).To(Equal(setupuser.DefaultUserRole))
			g.Expect(c.Spec.AdminRole).To(Equal(setupuser.DefaultAdminRole))
		},
		)
	}
}

func TestValidateVSphereObjects(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		filepath string
		wantErr  string
		prepare  func(context.Context, *setupuser.VSphereSetupUserConfig, *mocks.MockGovcClient)
	}{
		{
			name:     "test validate happy path",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.AdminRole).Return(false, nil)
			},
		},
		{
			name:     "test GroupExists error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(false, fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test RoleExists GlobalRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(false, fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test validate RoleExists UserRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(false, fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test validate RoleExists AdminRole error",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "govc error",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.AdminRole).Return(false, fmt.Errorf("govc error"))
			},
		},
		{
			name:     "test validate RoleExists AdminRole true",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "role MyExistingEKSAAdminRole already exists, please use force=true to ignore",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.GlobalRole).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.UserRole).Return(false, nil)
				gc.EXPECT().RoleExists(ctx, c.Spec.AdminRole).Return(true, nil)
			},
		},
		{
			name:     "test validate GroupExists true",
			filepath: "./testdata/configs/valid.yaml",
			wantErr:  "group MyExistingGroup already exists, please use force=true to ignore",
			prepare: func(ctx context.Context, c *setupuser.VSphereSetupUserConfig, gc *mocks.MockGovcClient) {
				gc.EXPECT().GroupExists(ctx, c.Spec.GroupName).Return(true, nil)
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
			ctrl := gomock.NewController(t)
			gc := mocks.NewMockGovcClient(ctrl)
			tt.prepare(ctx, c, gc)

			err = setupuser.ValidateVSphereObjects(ctx, c, gc)

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

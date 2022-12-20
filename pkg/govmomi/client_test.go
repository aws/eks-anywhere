package govmomi_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	govmomi_internal "github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/aws/eks-anywhere/pkg/govmomi"
	"github.com/aws/eks-anywhere/pkg/govmomi/mocks"
)

type fields struct {
	AuthorizationManager *mocks.MockVMOMIAuthorizationManager
	Finder               *mocks.MockVMOMIFinder
	Path                 string
}

func TestGetPrivsOnEntity(t *testing.T) {
	ctx := context.Background()
	username := "foobar"
	wantPrivs := []string{"DoManyThings", "DoFewThings"}
	results := []types.UserPrivilegeResult{
		{
			Privileges: wantPrivs,
		},
	}
	errMsg := "No entity found"

	tests := []struct {
		name    string
		objType string
		path    string
		// prepare lets us initialize our mocks within the `tests` slice. Oftentimes it also proves useful for other initialization
		prepare   func(f *fields)
		wantPrivs []string
		wantErr   string
	}{
		{
			name:      "test folder call happy path",
			objType:   govmomi.VSphereTypeFolder,
			path:      "Datacenter/vm/my/directory",
			wantPrivs: wantPrivs,
			wantErr:   "",
			prepare: func(f *fields) {
				obj := object.Folder{}
				objRefs := []types.ManagedObjectReference{obj.Common.Reference()}
				f.AuthorizationManager.EXPECT().FetchUserPrivilegeOnEntities(ctx, objRefs, username).Return(results, nil)
				f.Finder.EXPECT().Folder(ctx, f.Path).Return(&obj, nil)
			},
		},
		{
			name:      "test datastore call happy path",
			objType:   govmomi.VSphereTypeDatastore,
			path:      "Datacenter/datastore/LargeDatastore1",
			wantPrivs: wantPrivs,
			wantErr:   "",
			prepare: func(f *fields) {
				obj := object.Datastore{}
				objRefs := []types.ManagedObjectReference{obj.Common.Reference()}
				f.AuthorizationManager.EXPECT().FetchUserPrivilegeOnEntities(ctx, objRefs, username).Return(results, nil)
				f.Finder.EXPECT().Datastore(ctx, f.Path).Return(&obj, nil)
			},
		},
		{
			name:      "test resource pool call happy path",
			objType:   govmomi.VSphereTypeResourcePool,
			path:      "Datacenter/host/cluster-02/MyResourcePool",
			wantPrivs: wantPrivs,
			wantErr:   "",
			prepare: func(f *fields) {
				obj := object.ResourcePool{}
				objRefs := []types.ManagedObjectReference{obj.Common.Reference()}
				f.AuthorizationManager.EXPECT().FetchUserPrivilegeOnEntities(ctx, objRefs, username).Return(results, nil)
				f.Finder.EXPECT().ResourcePool(ctx, f.Path).Return(&obj, nil)
			},
		},
		{
			name:      "test virtual machine call happy path",
			objType:   govmomi.VSphereTypeVirtualMachine,
			path:      "Datacenter/vm/Templates/MyVMTemplate",
			wantPrivs: wantPrivs,
			wantErr:   "",
			prepare: func(f *fields) {
				obj := object.VirtualMachine{}
				objRefs := []types.ManagedObjectReference{obj.Common.Reference()}
				f.AuthorizationManager.EXPECT().FetchUserPrivilegeOnEntities(ctx, objRefs, username).Return(results, nil)
				f.Finder.EXPECT().VirtualMachine(ctx, f.Path).Return(&obj, nil)
			},
		},
		{
			name:      "test network call happy path",
			objType:   govmomi.VSphereTypeNetwork,
			path:      "Datacenter/network/VM Network",
			wantPrivs: wantPrivs,
			wantErr:   "",
			prepare: func(f *fields) {
				obj := object.Network{}
				objRefs := []types.ManagedObjectReference{obj.Reference()}
				f.AuthorizationManager.EXPECT().FetchUserPrivilegeOnEntities(ctx, objRefs, username).Return(results, nil)
				f.Finder.EXPECT().Network(ctx, f.Path).Return(&obj, nil)
			},
		},
		{
			name:      "test network call missing object",
			objType:   govmomi.VSphereTypeNetwork,
			path:      "Datacenter/network/VM Network",
			wantPrivs: []string{},
			wantErr:   errMsg,
			prepare: func(f *fields) {
				f.Finder.EXPECT().Network(ctx, f.Path).Return(nil, fmt.Errorf(errMsg))
			},
		},
		{
			name:      "test virtual machine call no privs",
			objType:   govmomi.VSphereTypeVirtualMachine,
			path:      "Datacenter/vm/Templates/MyVMTemplate",
			wantPrivs: []string{},
			wantErr:   errMsg,
			prepare: func(f *fields) {
				obj := object.VirtualMachine{}
				objRefs := []types.ManagedObjectReference{obj.Common.Reference()}
				f.AuthorizationManager.EXPECT().FetchUserPrivilegeOnEntities(ctx, objRefs, username).Return(nil, fmt.Errorf(errMsg))
				f.Finder.EXPECT().VirtualMachine(ctx, f.Path).Return(&obj, nil)
			},
		},
		{
			name:      "test resource pool call missing object",
			objType:   govmomi.VSphereTypeResourcePool,
			path:      "Datacenter/host/cluster-02/MyResourcePool",
			wantPrivs: []string{},
			wantErr:   errMsg,
			prepare: func(f *fields) {
				f.Finder.EXPECT().ResourcePool(ctx, f.Path).Return(nil, fmt.Errorf(errMsg))
			},
		},
		{
			name:      "test folder call empty object results",
			objType:   govmomi.VSphereTypeFolder,
			path:      "Datacenter/vm/my/directory",
			wantPrivs: []string{},
			wantErr:   "",
			prepare: func(f *fields) {
				obj := object.Folder{}
				objRefs := []types.ManagedObjectReference{obj.Common.Reference()}
				f.AuthorizationManager.EXPECT().FetchUserPrivilegeOnEntities(ctx, objRefs, username).Return(nil, nil)
				f.Finder.EXPECT().Folder(ctx, f.Path).Return(&obj, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			am := mocks.NewMockVMOMIAuthorizationManager(ctrl)
			finder := mocks.NewMockVMOMIFinder(ctrl)
			f := &fields{
				AuthorizationManager: am,
				Finder:               finder,
				Path:                 tt.path,
			}
			tt.prepare(f)

			g := NewWithT(t)

			vsc := govmomi.NewVMOMIClientCustom(nil, finder, username, am)

			privs, err := vsc.GetPrivsOnEntity(ctx, tt.path, tt.objType, username)
			if tt.wantErr == "" {
				g.Expect(err).To(Succeed())
				if !reflect.DeepEqual(privs, tt.wantPrivs) {
					t.Fatalf("privs = %v, want %v", privs, wantPrivs)
				}
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		},
		)
	}
}

func TestVMOMISessionBuilderBuild(t *testing.T) {
	insecure := false
	datacenter := "mydatacenter"
	datacenterObject := object.Datacenter{}
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	gcb := mocks.NewMockVMOMISessionBuilder(ctrl)

	c := &govmomi_internal.Client{
		Client: &vim25.Client{},
	}

	gcb.EXPECT().Build(ctx, gomock.Any(), insecure).Return(c, nil)

	mockFinder := mocks.NewMockVMOMIFinder(ctrl)
	mockFinder.EXPECT().Datacenter(ctx, datacenter).Return(&datacenterObject, nil)
	mockFinder.EXPECT().SetDatacenter(gomock.Any())

	vfb := mocks.NewMockVMOMIFinderBuilder(ctrl)
	vfb.EXPECT().Build(c.Client, true).Return(mockFinder)

	amb := mocks.NewMockVMOMIAuthorizationManagerBuilder(ctrl)
	amb.EXPECT().Build(c.Client)

	vcb := govmomi.NewVMOMIClientBuilderOverride(vfb, gcb, amb)
	_, err := vcb.Build(ctx, "myhost", "myusername", "mypassword", insecure, datacenter)
	if err != nil {
		t.Fatalf("Failed to build client with %s", err)
	}
}

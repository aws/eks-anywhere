package govmomi

import (
	"context"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	VSphereTypeFolder         = "Folder"
	VSphereTypeNetwork        = "Network"
	VSphereTypeResourcePool   = "ResourcePool"
	VSphereTypeDatastore      = "Datastore"
	VSphereTypeVirtualMachine = "VirtualMachine"
)

type VMOMIAuthorizationManager interface {
	FetchUserPrivilegeOnEntities(ctx context.Context, entities []types.ManagedObjectReference, userName string) ([]types.UserPrivilegeResult, error)
}

type VMOMIFinder interface {
	Datastore(ctx context.Context, path string) (*object.Datastore, error)
	Folder(ctx context.Context, path string) (*object.Folder, error)
	Network(ctx context.Context, path string) (object.NetworkReference, error)
	ResourcePool(ctx context.Context, path string) (*object.ResourcePool, error)
	VirtualMachine(ctx context.Context, path string) (*object.VirtualMachine, error)
	Datacenter(ctx context.Context, path string) (*object.Datacenter, error)
	SetDatacenter(dc *object.Datacenter) *find.Finder
}

type VMOMIClient struct {
	Gcvm                 *govmomi.Client
	Finder               VMOMIFinder
	username             string
	AuthorizationManager VMOMIAuthorizationManager
}

func NewVMOMIClientCustom(gcvm *govmomi.Client, f VMOMIFinder, username string, am VMOMIAuthorizationManager) *VMOMIClient {
	return &VMOMIClient{
		Gcvm:                 gcvm,
		Finder:               f,
		username:             username,
		AuthorizationManager: am,
	}
}

func (vsc *VMOMIClient) Username() string {
	return vsc.username
}

func (vsc *VMOMIClient) GetPrivsOnEntity(ctx context.Context, path string, objType string, username string) ([]string, error) {
	var vSphereObjectReference types.ManagedObjectReference
	emptyResult := []string{}
	var err error

	switch objType {

	case VSphereTypeFolder:
		vSphereObjectReference, err = vsc.getFolder(ctx, path)
	case VSphereTypeNetwork:
		vSphereObjectReference, err = vsc.getNetwork(ctx, path)
	case VSphereTypeDatastore:
		vSphereObjectReference, err = vsc.getDatastore(ctx, path)
	case VSphereTypeResourcePool:
		vSphereObjectReference, err = vsc.getResourcePool(ctx, path)
	case VSphereTypeVirtualMachine:
		vSphereObjectReference, err = vsc.getVirtualMachine(ctx, path)
	}

	if err != nil {
		return emptyResult, err
	}
	refs := []types.ManagedObjectReference{vSphereObjectReference}

	result, err := vsc.AuthorizationManager.FetchUserPrivilegeOnEntities(ctx, refs, username)
	if err != nil {
		return emptyResult, err
	}

	if len(result) > 0 {
		return result[0].Privileges, nil
	} else {
		return emptyResult, nil
	}
}

func (vsc *VMOMIClient) getFolder(ctx context.Context, path string) (types.ManagedObjectReference, error) {
	obj, err := vsc.Finder.Folder(ctx, path)
	if err != nil {
		return types.ManagedObjectReference{}, err
	} else {
		return obj.Common.Reference(), nil
	}
}

func (vsc *VMOMIClient) getNetwork(ctx context.Context, path string) (types.ManagedObjectReference, error) {
	obj, err := vsc.Finder.Network(ctx, path)
	if err != nil {
		return types.ManagedObjectReference{}, err
	} else {
		return obj.Reference(), nil
	}
}

func (vsc *VMOMIClient) getDatastore(ctx context.Context, path string) (types.ManagedObjectReference, error) {
	obj, err := vsc.Finder.Datastore(ctx, path)
	if err != nil {
		return types.ManagedObjectReference{}, err
	} else {
		return obj.Common.Reference(), nil
	}
}

func (vsc *VMOMIClient) getResourcePool(ctx context.Context, path string) (types.ManagedObjectReference, error) {
	obj, err := vsc.Finder.ResourcePool(ctx, path)
	if err != nil {
		return types.ManagedObjectReference{}, err
	} else {
		return obj.Common.Reference(), nil
	}
}

func (vsc *VMOMIClient) getVirtualMachine(ctx context.Context, path string) (types.ManagedObjectReference, error) {
	obj, err := vsc.Finder.VirtualMachine(ctx, path)
	if err != nil {
		return types.ManagedObjectReference{}, err
	} else {
		return obj.Common.Reference(), nil
	}
}

package setupuser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/executables"
)

func SetupGOVCEnv(ctx context.Context, vsuc *VSphereSetupUserConfig) error {
	os.Setenv("GOVC_URL", vsuc.Spec.Connection.Server)
	os.Setenv("GOVC_INSECURE", strconv.FormatBool(vsuc.Spec.Connection.Insecure))
	os.Setenv("GOVC_DATACENTER", vsuc.Spec.Datacenter)
	return nil
}

func CreateUser(ctx context.Context, govc *executables.Govc, vsuc *VSphereSetupUserConfig, password string) error {
	// create user
	err := govc.CreateUser(ctx, vsuc.Spec.Username, password)
	if err != nil {
		return err
	}

	return nil
}

type SetupVSphereUserTask struct {
	vsuc *VSphereSetupUserConfig
	govc *executables.Govc
	force bool
}

func NewSetupVSphereUserTask(vsuc *VSphereSetupUserConfig, govc *executables.Govc, force bool) *SetupVSphereUserTask{
	return &SetupVSphereUserTask{
		vsuc: vsuc,
		govc: govc,
		force: force,
	}
}
	

func (s *SetupVSphereUserTask) CreateGroup(ctx context.Context) error {

	var exists bool
	var err error

	exists, err = s.govc.GroupExists(ctx, s.vsuc.Spec.GroupName)
	if err != nil {
		return err
	} else if exists && !s.force {
		return fmt.Errorf("Group %s already exists", s.vsuc.Spec.GroupName)
	}

	err = s.govc.CreateGroup(ctx, s.vsuc.Spec.GroupName)
	if err != nil {
		return err
	}
	// else {
	// 	fmt.Printf("Skipping creating %s because it already exists\n", vsuc.Spec.GroupName)
	// }

	return nil
}

func (s *SetupVSphereUserTask) CreateRoles(ctx context.Context) error {
	// create roles
	for _, r := range getRoles(s.vsuc) {
		exists, err := s.govc.RoleExists(ctx, r.name)

		if exists && !s.force {
			return fmt.Errorf("Role %s exists", r.name)
		}

		if !exists {
			err = s.govc.CreateRole(ctx, r.name, r.privs)
			if err != nil {
				fmt.Printf("Failed to create %s role with %v\n", r.name, r.privs)
				return err
			}
			fmt.Printf("Created %s role\n", r.name)
		} else {
			fmt.Printf("Skipping creating %s role because it already exists\n", r.name)
		}
	}

	return nil
}

func (s *SetupVSphereUserTask) AssociatePermissions(ctx context.Context) error {
	// global on root
	// admin to template and vm folders
	// user on User on network, datastore, and resourcepool

	// associate roles to objects and group
	// for _, ra := range getRequiredAccesses(vsuc) {

	// 	switch ra.privsContent {
	// 	case config.VSphereGlobalPrivsFile:
	// 		err = govc.SetPermission(ctx, vsuc.GroupName, vsuc.GlobalRole, ra.path, vsuc.Domain)
	// 		fmt.Printf("Set role %s on %s for group %s\n", vsuc.GlobalRole, ra.path, vsuc.GroupName)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	case config.VSphereUserPrivsFile:
	// 		err = govc.SetPermission(ctx, vsuc.GroupName, vsuc.UserRole, ra.path, vsuc.Domain)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		fmt.Printf("Set role %s on %s for group %s\n", vsuc.GlobalRole, ra.path, vsuc.GroupName)
	// 	}
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// err = govc.SetPermission(ctx, vsuc.GroupName, vsuc.CloudAdmin, vsuc.Templates, vsuc.Domain)
	// if err != nil {
	// 	return err
	// }
	// err = govc.SetPermission(ctx, vsuc.GroupName, vsuc.CloudAdmin, vsuc.VirtualMachines, vsuc.Domain)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (s *SetupVSphereUserTask) AddUserToGroup(ctx context.Context) error {

	// associate user to group
	err := s.govc.AddUserToGroup(ctx, s.vsuc.Spec.GroupName, s.vsuc.Spec.Username)
	fmt.Printf("Adding user %s to group %s\n", s.vsuc.Spec.Username, s.vsuc.Spec.GroupName)
	if err != nil {
		return err
	}

	return nil
}

func (s *SetupVSphereUserTask) Run(ctx context.Context) error {


	err := s.CreateGroup(ctx)
	if err != nil {
		return err
	}

	err = s.AddUserToGroup(ctx)
	if err != nil {
		return err
	}

	err = s.CreateRoles(ctx)
	if err != nil {
		return err
	}

	return nil
}

func getPrivsFromFile(privsContent string) ([]string, error) {
	var requiredPrivs []string
	err := json.Unmarshal([]byte(privsContent), &requiredPrivs)
	if err != nil {
		return nil, err
	}
	return requiredPrivs, nil
}

type vsphereRole struct {
	name  string
	privs []string
}

func getRoles(vsuc *VSphereSetupUserConfig) []vsphereRole {
	globalPrivs, _ := getPrivsFromFile(config.VSphereGlobalPrivsFile)
	userPrivs, _ := getPrivsFromFile(config.VSphereUserPrivsFile)
	cloudAdminPrivs, _ := getPrivsFromFile(config.VSphereAdminPrivsFile)
	return []vsphereRole{
		{
			name:  vsuc.Spec.GlobalRole,
			privs: globalPrivs,
		},
		{
			name:  vsuc.Spec.UserRole,
			privs: userPrivs,
		},
		{
			name:  vsuc.Spec.AdminRole,
			privs: cloudAdminPrivs,
		},
	}
}

// func getRequiredAccesses(vsuc *config.VSphereSetupUserConfig) []RequiredAccess {
// 	privObjs := []RequiredAccess{
// 		// validate global root priv settings are correct
// 		{
// 			objectType:   vsphereTypeFolder,
// 			privsContent: config.VSphereGlobalPrivsFile,
// 			path:         vsphereRootPath,
// 		},

// 		// validate object-level priv settings are correct
// 		{
// 			objectType:   vsphereTypeDatastore,
// 			privsContent: config.VSphereUserPrivsFile,
// 			path:         vsuc.Datastore,
// 		},
// 		{
// 			objectType:   vsphereTypeResourcePool,
// 			privsContent: config.VSphereUserPrivsFile,
// 			path:         vsuc.ResourcePool,
// 		},
// 		{
// 			objectType:   vsphereTypeNetwork,
// 			privsContent: config.VSphereUserPrivsFile,
// 			path:         vsuc.Network,
// 		},
// 		// validate Administrator role (all privs) on VM folder and Template folder
// 		{
// 			objectType:   vsphereTypeFolder,
// 			privsContent: config.VSphereAdminPrivsFile,
// 			path:         vsuc.VirtualMachines,
// 		},
// 		// {
// 		// 	objectType:   "VirtualMachine",
// 		// 	privsContent: config.VSphereAdminPrivsFile,
// 		// 	path:         vsuc.Templates,
// 		// },
// 		{
// 			objectType:   vsphereTypeFolder,
// 			privsContent: config.VSphereAdminPrivsFile,
// 			path:         vsuc.Templates,
// 		},
// 	}

// 	return privObjs
// }


package vsphere

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/executables"
)

func ConfigureVSphere(ctx context.Context, govc *executables.Govc, vusc *config.VSphereSetupUserConfig) error {
	// create user
	// assume user already exists, I don't want to deal with passwords right now
	// err := govc.CreateUser(ctx, vusc.Username)q
	// if err != nil {
	// 	return err
	// }

	// create group
	var err error
	var exists bool
	exists, err = govc.GroupExists(ctx, vusc.GroupName)
	if err != nil {
		return err
	}

	if !exists {
		err := govc.CreateGroup(ctx, vusc.GroupName)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("Skipping creating %s because it already exists\n", vusc.GroupName)
	}
	if err != nil {
		return err
	}

	// associate user to group
	err = govc.AddGroupUser(ctx, vusc.GroupName, vusc.Username)
	fmt.Printf("Adding user %s to group %s\n", vusc.Username, vusc.GroupName)
	if err != nil {
		return err
	}

	// create roles
	for _, r := range getRoles(vusc) {
		exists, err = govc.RoleExists(ctx, r.name)
		if !exists {
			err = govc.CreateRole(ctx, r.name, r.privs)
			if err != nil {
				fmt.Printf("Failed to create %s role with %v\n", r.name, r.privs)
				return err
			}
			fmt.Printf("Created %s role\n", r.name)
		} else {
			fmt.Printf("Skipping creating %s role because it already exists\n", r.name)
		}
	}

	// associate roles to objects and group
	for _, ra := range getRequiredAccesses(vusc) {

		switch ra.privsContent {
		case config.VSphereGlobalPrivsFile:
			err = govc.SetPermission(ctx, vusc.GroupName, vusc.GlobalRole, ra.path, vusc.Domain)
			fmt.Printf("Set role %s on %s for group %s\n", vusc.GlobalRole, ra.path, vusc.GroupName)
			if err != nil {
				return err
			}
		case config.VSphereUserPrivsFile:
			err = govc.SetPermission(ctx, vusc.GroupName, vusc.UserRole, ra.path, vusc.Domain)
			if err != nil {
				return err
			}
			fmt.Printf("Set role %s on %s for group %s\n", vusc.GlobalRole, ra.path, vusc.GroupName)
		}
		if err != nil {
			return err
		}
	}
	err = govc.SetPermission(ctx, vusc.GroupName, vusc.CloudAdmin, vusc.Templates, vusc.Domain)
	if err != nil {
		return err
	}
	err = govc.SetPermission(ctx, vusc.GroupName, vusc.CloudAdmin, vusc.VirtualMachines, vusc.Domain)
	if err != nil {
		return err
	}

	return nil
}

type vsphereRole struct {
	name  string
	privs []string
}

func getRoles(vusc *config.VSphereSetupUserConfig) []vsphereRole {
	globalPrivs, _ := getPrivsFromFile(config.VSphereGlobalPrivsFile)
	userPrivs, _ := getPrivsFromFile(config.VSphereUserPrivsFile)
	cloudAdminPrivs, _ := getPrivsFromFile(config.VSphereAdminPrivsFile)
	return []vsphereRole{
		{
			name:  vusc.GlobalRole,
			privs: globalPrivs,
		},
		{
			name:  vusc.UserRole,
			privs: userPrivs,
		},
		{
			name:  vusc.CloudAdmin,
			privs: cloudAdminPrivs,
		},
	}
}

func getRequiredAccesses(vusc *config.VSphereSetupUserConfig) []RequiredAccess {
	privObjs := []RequiredAccess{
		// validate global root priv settings are correct
		{
			objectType:   vsphereTypeFolder,
			privsContent: config.VSphereGlobalPrivsFile,
			path:         vsphereRootPath,
		},

		// validate object-level priv settings are correct
		{
			objectType:   vsphereTypeDatastore,
			privsContent: config.VSphereUserPrivsFile,
			path:         vusc.Datastore,
		},
		{
			objectType:   vsphereTypeResourcePool,
			privsContent: config.VSphereUserPrivsFile,
			path:         vusc.ResourcePool,
		},
		{
			objectType:   vsphereTypeNetwork,
			privsContent: config.VSphereUserPrivsFile,
			path:         vusc.Network,
		},
		// validate Administrator role (all privs) on VM folder and Template folder
		{
			objectType:   vsphereTypeFolder,
			privsContent: config.VSphereAdminPrivsFile,
			path:         vusc.VirtualMachines,
		},
		// {
		// 	objectType:   "VirtualMachine",
		// 	privsContent: config.VSphereAdminPrivsFile,
		// 	path:         vusc.Templates,
		// },
		{
			objectType:   vsphereTypeFolder,
			privsContent: config.VSphereAdminPrivsFile,
			path:         vusc.Templates,
		},
	}

	return privObjs
}

func getPrivsFromFile(privsContent string) ([]string, error) {
	var requiredPrivs []string
	err := json.Unmarshal([]byte(privsContent), &requiredPrivs)
	if err != nil {
		return nil, err
	}
	return requiredPrivs, nil
}

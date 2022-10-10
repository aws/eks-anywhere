package setupuser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

const (
	vSphereRootPath = "/"
)

func SetupGOVCEnv(ctx context.Context, vsuc *VSphereSetupUserConfig) error {
	os.Setenv("GOVC_URL", vsuc.Spec.Connection.Server)
	os.Setenv("GOVC_INSECURE", strconv.FormatBool(vsuc.Spec.Connection.Insecure))
	os.Setenv("GOVC_DATACENTER", vsuc.Spec.Datacenter)
	return nil
}

func Run(ctx context.Context, vsuc *VSphereSetupUserConfig, govc vsphere.ProviderGovcClient) error {
	err := CreateGroup(ctx, vsuc, govc)
	if err != nil {
		return err
	}

	err = AddUserToGroup(ctx, vsuc, govc)
	if err != nil {
		return err
	}

	err = CreateRoles(ctx, vsuc, govc)
	if err != nil {
		return err
	}

	err = AssociateRolesToObjects(ctx, vsuc, govc)
	if err != nil {
		return err
	}

	return nil
}

func CreateUser(ctx context.Context, vsuc *VSphereSetupUserConfig, govc vsphere.ProviderGovcClient, password string) error {
	// create user
	err := govc.CreateUser(ctx, vsuc.Spec.Username, password)
	if err != nil {
		return err
	}

	return nil
}

func CreateGroup(ctx context.Context, vsuc *VSphereSetupUserConfig, govc vsphere.ProviderGovcClient) error {
	exists, err := govc.GroupExists(ctx, vsuc.Spec.GroupName)
	if err != nil {
		return err
	}
	if !exists {
		err = govc.CreateGroup(ctx, vsuc.Spec.GroupName)
	} else {
		logger.V(0).Info(fmt.Sprintf("Skipping creating %s because it already exists\n", vsuc.Spec.GroupName))
	}
	if err != nil {
		return err
	}

	return nil
}

func CreateRoles(ctx context.Context, vsuc *VSphereSetupUserConfig, govc vsphere.ProviderGovcClient) error {
	// create roles
	for _, r := range getRoles(vsuc) {
		exists, err := govc.RoleExists(ctx, r.name)
		if err != nil {
			return err
		}

		if !exists {
			err = govc.CreateRole(ctx, r.name, r.privs)
			if err != nil {
				logger.V(0).Info(fmt.Sprintf("Failed to create %s role with %v\n", r.name, r.privs))
				return err
			}
			logger.V(0).Info(fmt.Sprintf("Created %s role\n", r.name))
		} else {
			logger.V(0).Info(fmt.Sprintf("Skipping creating %s role because it already exists\n", r.name))
		}
	}

	return nil
}

func AssociateRolesToObjects(ctx context.Context, vsuc *VSphereSetupUserConfig, govc vsphere.ProviderGovcClient) error {
	// global on root
	// admin to template and vm folders
	// user on network, datastore, and resourcepool

	err := setGroupRoleOnObjects(ctx, vsuc, govc, vsuc.Spec.GlobalRole, []string{vSphereRootPath})
	if err != nil {
		return err
	}

	adminRoleObjects := append(vsuc.Spec.Objects.Folders, vsuc.Spec.Objects.Templates...)
	err = setGroupRoleOnObjects(ctx, vsuc, govc, vsuc.Spec.AdminRole, adminRoleObjects)
	if err != nil {
		return err
	}

	userRoleObjects := getUserRoleObjects(vsuc)
	err = setGroupRoleOnObjects(ctx, vsuc, govc, vsuc.Spec.UserRole, userRoleObjects)
	if err != nil {
		return err
	}

	return nil
}

func AddUserToGroup(ctx context.Context, vsuc *VSphereSetupUserConfig, govc vsphere.ProviderGovcClient) error {
	// associate user to group
	err := govc.AddUserToGroup(ctx, vsuc.Spec.GroupName, vsuc.Spec.Username)
	logger.V(0).Info(fmt.Sprintf("Adding user %s to group %s\n", vsuc.Spec.Username, vsuc.Spec.GroupName))
	if err != nil {
		return err
	}

	return nil
}

func setGroupRoleOnObjects(ctx context.Context, vsuc *VSphereSetupUserConfig, govc vsphere.ProviderGovcClient, role string, objects []string) error {
	for _, obj := range objects {

		err := govc.SetGroupRoleOnObject(ctx, vsuc.Spec.GroupName, role, obj, vsuc.Spec.VSphereDomain)
		if err != nil {
			return err
		}
		logger.V(0).Info(fmt.Sprintf("Set role %s on %s for group %s\n", role, obj, vsuc.Spec.GroupName))
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

func getUserRoleObjects(vsuc *VSphereSetupUserConfig) []string {
	objects := append(vsuc.Spec.Objects.Networks, vsuc.Spec.Objects.Datastores...)
	objects = append(objects, vsuc.Spec.Objects.ResourcePools...)
	return objects
}

package setupuser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	vSphereRootPath = "/"
)

// GovcClient specifies govc functions required to configure a vsphere user.
type GovcClient interface {
	CreateUser(ctx context.Context, username string, password string) error
	UserExists(ctx context.Context, username string) (bool, error)
	CreateGroup(ctx context.Context, name string) error
	GroupExists(ctx context.Context, name string) (bool, error)
	AddUserToGroup(ctx context.Context, name string, username string) error
	RoleExists(ctx context.Context, name string) (bool, error)
	CreateRole(ctx context.Context, name string, privileges []string) error
	SetGroupRoleOnObject(ctx context.Context, principal string, role string, object string, domain string) error
}

// SetupGOVCEnv creates appropriate govc environment variables to build govc client.
func SetupGOVCEnv(ctx context.Context, vsuc *VSphereSetupUserConfig) error {
	err := os.Setenv("GOVC_URL", vsuc.Spec.Connection.Server)
	if err != nil {
		return err
	}
	err = os.Setenv("GOVC_INSECURE", strconv.FormatBool(vsuc.Spec.Connection.Insecure))
	if err != nil {
		return err
	}
	err = os.Setenv("GOVC_DATACENTER", vsuc.Spec.Datacenter)
	if err != nil {
		return err
	}
	return nil
}

// Run sets up a vSphere user with appropriate group, role, and permissions to create EKS-A kubernetes clusters.
func Run(ctx context.Context, vsuc *VSphereSetupUserConfig, govc GovcClient) error {
	err := createGroup(ctx, vsuc, govc)
	if err != nil {
		return err
	}

	err = addUserToGroup(ctx, vsuc, govc)
	if err != nil {
		return err
	}

	err = createRoles(ctx, vsuc, govc)
	if err != nil {
		return err
	}

	err = associateRolesToObjects(ctx, vsuc, govc)
	if err != nil {
		return err
	}

	return nil
}

func createGroup(ctx context.Context, vsuc *VSphereSetupUserConfig, govc GovcClient) error {
	exists, err := govc.GroupExists(ctx, vsuc.Spec.GroupName)
	if err != nil {
		return err
	}
	if !exists {
		err = govc.CreateGroup(ctx, vsuc.Spec.GroupName)
	} else {
		logger.V(0).Info(fmt.Sprintf("Skipping creating group %s because it already exists", vsuc.Spec.GroupName))
	}
	if err != nil {
		return err
	}

	return nil
}

func createRoles(ctx context.Context, vsuc *VSphereSetupUserConfig, govc GovcClient) error {
	roles, err := getRoles(vsuc)
	if err != nil {
		return err
	}

	for _, r := range roles {
		exists, err := govc.RoleExists(ctx, r.name)
		if err != nil {
			return err
		}

		if !exists {
			err = govc.CreateRole(ctx, r.name, r.privs)
			if err != nil {
				logger.V(0).Info(fmt.Sprintf("Failed to create %s role with %v", r.name, r.privs))
				return err
			}
			logger.V(0).Info(fmt.Sprintf("Created %s role", r.name))
		} else {
			logger.V(0).Info(fmt.Sprintf("Skipping creating %s role because it already exists", r.name))
		}
	}

	return nil
}

func associateRolesToObjects(ctx context.Context, vsuc *VSphereSetupUserConfig, govc GovcClient) error {
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

func addUserToGroup(ctx context.Context, vsuc *VSphereSetupUserConfig, govc GovcClient) error {
	// associate user to group
	err := govc.AddUserToGroup(ctx, vsuc.Spec.GroupName, vsuc.Spec.Username)
	logger.V(0).Info(fmt.Sprintf("Adding user %s to group %s", vsuc.Spec.Username, vsuc.Spec.GroupName))
	if err != nil {
		return err
	}

	return nil
}

func setGroupRoleOnObjects(ctx context.Context, vsuc *VSphereSetupUserConfig, govc GovcClient, role string, objects []string) error {
	for _, obj := range objects {

		err := govc.SetGroupRoleOnObject(ctx, vsuc.Spec.GroupName, role, obj, vsuc.Spec.VSphereDomain)
		if err != nil {
			return err
		}
		logger.V(0).Info(fmt.Sprintf("Set role %s on %s for group %s", role, obj, vsuc.Spec.GroupName))
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

func getRoles(vsuc *VSphereSetupUserConfig) ([]vsphereRole, error) {
	globalPrivs, err := getPrivsFromFile(config.VSphereGlobalPrivsFile)
	if err != nil {
		return []vsphereRole{}, err
	}
	userPrivs, err := getPrivsFromFile(config.VSphereUserPrivsFile)
	if err != nil {
		return []vsphereRole{}, err
	}
	cloudAdminPrivs, err := getPrivsFromFile(config.VSphereAdminPrivsFile)
	if err != nil {
		return []vsphereRole{}, err
	}
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
	}, nil
}

func getUserRoleObjects(vsuc *VSphereSetupUserConfig) []string {
	objects := append(vsuc.Spec.Objects.Networks, vsuc.Spec.Objects.Datastores...)
	objects = append(objects, vsuc.Spec.Objects.ResourcePools...)
	return objects
}

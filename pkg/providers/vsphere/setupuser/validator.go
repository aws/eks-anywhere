package setupuser

import (
	"context"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	DefaultUsername   = "eksa"
	DefaultGroup      = "EKSAUsers"
	DefaultGlobalRole = "EKSAGlobalRole"
	DefaultUserRole   = "EKSAUserRole"
	DefaultAdminRole  = "EKSACloudAdminRole"
)

type Connection struct {
	Server   string `yaml:"server"`
	Insecure bool   `yaml:"insecure"`
}

type Objects struct {
	Networks      []string `yaml:"networks"`
	Datastores    []string `yaml:"datastores"`
	ResourcePools []string `yaml:"resourcePools"`
	Folders       []string `yaml:"folders"`
	Templates     []string `yaml:"templates"`
}

type VSphereUserSpec struct {
	Datacenter    string     `yaml:"datacenter"`
	VSphereDomain string     `yaml:"vSphereDomain"`
	Connection    Connection `yaml:"connection"`
	Objects       Objects    `yaml:"objects"`
	// Below are optional fields with defaults
	Username   string `yaml:"username"`
	GroupName  string `yaml:"group"`
	GlobalRole string `yaml:"globalRole"`
	UserRole   string `yaml:"userRole"`
	AdminRole  string `yaml:"adminRole"`
}

type VSphereSetupUserConfig struct {
	ApiVersion string          `yaml:"apiVersion"`
	Kind       string          `yaml:"kind"`
	Spec       VSphereUserSpec `yaml:"spec"`
}

func GenerateConfig(ctx context.Context, filepath string) (*VSphereSetupUserConfig, error) {
	c, err := readConfig(ctx, filepath)
	if err != nil {
		return nil, err
	}

	err = validate(c)
	if err != nil {
		return nil, err
	}

	setDefaults(c)

	return c, nil
}

func readConfig(ctx context.Context, filepath string) (*VSphereSetupUserConfig, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s, err = %v", filepath, err)
	}

	c := VSphereSetupUserConfig{}

	if err = yaml.Unmarshal(file, &c); err != nil {
		return nil, fmt.Errorf("failed to parse %s, err = %v", filepath, err)
	}

	return &c, nil
}

func validate(c *VSphereSetupUserConfig) error {
	errs := []string{}

	if c.Spec.Datacenter == "" {
		errs = append(errs, "datacenter cannot be empty")
	}

	if c.Spec.VSphereDomain == "" {
		errs = append(errs, "vSphereDomain cannot be empty")
	}

	if c.Spec.Connection.Server == "" {
		errs = append(errs, "server cannot be empty")
	}

	if len(errs) > 0 {
		return fmt.Errorf("validations failed: %s", strings.Join(errs[:], ","))
	}

	return nil
}

func setDefaults(c *VSphereSetupUserConfig) {
	if c.Spec.GlobalRole == "" {
		c.Spec.GlobalRole = DefaultGlobalRole
	}

	if c.Spec.UserRole == "" {
		c.Spec.UserRole = DefaultUserRole
	}

	if c.Spec.AdminRole == "" {
		c.Spec.AdminRole = DefaultAdminRole
	}

	if c.Spec.GroupName == "" {
		c.Spec.GroupName = DefaultGroup
	}

	if c.Spec.Username == "" {
		c.Spec.Username = DefaultUsername
	}
}

// ValidateVSphereObjects validates objects do not exist before configuring user.
func ValidateVSphereObjects(ctx context.Context, c *VSphereSetupUserConfig, govc GovcClient) error {
	exists, err := govc.GroupExists(ctx, c.Spec.GroupName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("group %s already exists, please use force=true to ignore", c.Spec.GroupName)
	}

	roles := []string{c.Spec.GlobalRole, c.Spec.UserRole, c.Spec.AdminRole}
	for _, r := range roles {
		exists, err := govc.RoleExists(ctx, r)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("role %s already exists, please use force=true to ignore", r)
		}
	}

	return nil
}

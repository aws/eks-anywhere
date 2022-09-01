package setupuser

import (
	"context"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/aws/eks-anywhere/pkg/validations"
)

const (
	DefaultUsername   = "eksa"
	DefaultGroup      = "EKSAUsers"
	DefaultGlobalRole = "EKSAGlobalRole"
	DefaultUserRole   = "EKSAUserRole"
	DefaultAdminRole  = "EKSACloudAdminRole"
)

type connection struct {
	Server   string `yaml:"server"`
	Insecure bool   `yaml:"insecure"`
}

type objects struct {
	Networks      []string `yaml:"networks"`
	Datastores    []string `yaml:"datastores"`
	ResourcePools []string `yaml:"resourcePools"`
	Folders       []string `yaml:"folders"`
	Templates     []string `yaml:"templates"`
}

type vSphereUserSpec struct {
	Datacenter    string     `yaml:"datacenter"`
	VSphereDomain string     `yaml:"vSphereDomain"`
	Connection    connection `yaml:"connection"`
	Objects       objects    `yaml:"objects"`
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
	Spec       vSphereUserSpec `yaml:"spec"`
}

func GenerateConfig(ctx context.Context, filepath string) (*VSphereSetupUserConfig, error) {
	c, err := readConfig(ctx, filepath)
	if err != nil {
		return nil, err
	}

	err = validate(ctx, c)
	if err != nil {
		return nil, err
	}

	setDefaults(ctx, c)

	return c, nil
}

func readConfig(ctx context.Context, filepath string) (*VSphereSetupUserConfig, error) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file %s, err = %v", filepath, err)
	}

	c := VSphereSetupUserConfig{}

	if err = yaml.Unmarshal(file, &c); err != nil {
		return nil, fmt.Errorf("Failed to parse %s, err = %v", filepath, err)
	}

	return &c, nil
}

func validate(ctx context.Context, c *VSphereSetupUserConfig) error {
	results := []validations.ValidationResult{
		{
			Name:        "validate datacenter",
			Remediation: "",
			Err:         validateDatacenter(ctx, c),
		},
		{
			Name:        "validate vspheredomain",
			Remediation: "",
			Err:         validateVSphereDomain(ctx, c),
		},
		{
			Name:        "validate connection",
			Remediation: "",
			Err:         validateConnection(ctx, c),
		},
	}

	errs := []string{}
	for _, r := range results {
		if r.Err != nil {
			errs = append(errs, r.Err.Error())
		}
	}

	if len(errs) > 0 {
		return &validations.ValidationError{Errs: errs}
	}

	return nil
}

func validateDatacenter(ctx context.Context, c *VSphereSetupUserConfig) error {
	if c.Spec.Datacenter == "" {
		return fmt.Errorf("datacenter cannot be empty")
	}
	return nil
}

func validateVSphereDomain(ctx context.Context, c *VSphereSetupUserConfig) error {
	if c.Spec.VSphereDomain == "" {
		return fmt.Errorf("vSphereDomain cannot be empty")
	}
	return nil
}

func validateConnection(ctx context.Context, c *VSphereSetupUserConfig) error {
	if c.Spec.Connection.Server == "" {
		return fmt.Errorf("server cannot be empty")
	}
	return nil
}

func setDefaults(ctx context.Context, c *VSphereSetupUserConfig) {
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

	if c.Spec.Objects.Networks == nil {
		c.Spec.Objects.Networks = []string{}
	}

	if c.Spec.Objects.Networks == nil {
		c.Spec.Objects.Networks = []string{}
	}

	if c.Spec.Objects.Networks == nil {
		c.Spec.Objects.Networks = []string{}
	}

	if c.Spec.Objects.Datastores == nil {
		c.Spec.Objects.Datastores = []string{}
	}

	if c.Spec.Objects.ResourcePools == nil {
		c.Spec.Objects.ResourcePools = []string{}
	}

	if c.Spec.Objects.Folders == nil {
		c.Spec.Objects.Folders = []string{}
	}

	if c.Spec.Objects.Folders == nil {
		c.Spec.Objects.Folders = []string{}
	}
}

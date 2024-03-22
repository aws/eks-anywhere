package vsphere

import (
	"fmt"
	"os"
	"strconv"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
)

func SetupEnvVars(datacenterConfig *anywherev1.VSphereDatacenterConfig) error {
	// TODO(cxbrowne): We set environment variables here in response to existing of other environment
	// variables. Investigate why this is done, and possible remove the need for this.
	// https://github.com/aws/eks-anywhere-internal/issues/2192
	if vSphereUsername, ok := os.LookupEnv(config.EksavSphereUsernameKey); ok && len(vSphereUsername) > 0 {
		if err := os.Setenv(vSphereUsernameKey, vSphereUsername); err != nil {
			return fmt.Errorf("unable to set %s: %v", config.EksavSphereUsernameKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", config.EksavSphereUsernameKey)
	}

	if vSpherePassword, ok := os.LookupEnv(config.EksavSpherePasswordKey); ok && len(vSpherePassword) > 0 {
		if err := os.Setenv(vSpherePasswordKey, vSpherePassword); err != nil {
			return fmt.Errorf("unable to set %s: %v", config.EksavSpherePasswordKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", config.EksavSpherePasswordKey)
	}

	if err := os.Setenv(vSphereServerKey, datacenterConfig.Spec.Server); err != nil {
		return fmt.Errorf("unable to set %s: %v", vSphereServerKey, err)
	}

	if err := os.Setenv(expClusterResourceSetKey, "true"); err != nil {
		return fmt.Errorf("unable to set %s: %v", expClusterResourceSetKey, err)
	}

	// TODO: move this somewhere else since it's not vSphere specific
	if _, ok := os.LookupEnv(eksaLicense); !ok {
		if err := os.Setenv(eksaLicense, ""); err != nil {
			return fmt.Errorf("unable to set %s: %v", eksaLicense, err)
		}
	}

	if err := os.Setenv(govcInsecure, strconv.FormatBool(datacenterConfig.Spec.Insecure)); err != nil {
		return fmt.Errorf("unable to set %s: %v", govcInsecure, err)
	}
	if err := os.Setenv(govcDatacenterKey, datacenterConfig.Spec.Datacenter); err != nil {
		return fmt.Errorf("unable to set %s: %v", govcDatacenterKey, err)
	}
	return nil
}

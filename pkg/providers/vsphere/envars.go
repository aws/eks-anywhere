package vsphere

import (
	"fmt"
	"os"
	"strconv"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func setupEnvVars(datacenterConfig *anywherev1.VSphereDatacenterConfig) error {
	if vSphereUsername, ok := os.LookupEnv(eksavSphereUsernameKey); ok && len(vSphereUsername) > 0 {
		if err := os.Setenv(vSphereUsernameKey, vSphereUsername); err != nil {
			return fmt.Errorf("unable to set %s: %v", eksavSphereUsernameKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", eksavSphereUsernameKey)
	}

	if vSpherePassword, ok := os.LookupEnv(eksavSpherePasswordKey); ok && len(vSpherePassword) > 0 {
		if err := os.Setenv(vSpherePasswordKey, vSpherePassword); err != nil {
			return fmt.Errorf("unable to set %s: %v", eksavSpherePasswordKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", eksavSpherePasswordKey)
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
	return nil
}

package nutanix

import (
	"errors"
	"os"
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func fakeOSSetenv(failureKey string) func(key string, value string) error {
	return func(key string, value string) error {
		if key == failureKey {
			return errors.New("os.Setenv failed")
		}

		return nil
	}
}

func restoreOSSetenv(replace func(key string, value string) error) {
	osSetenv = replace
}

func TestUsernameIsNotSet(t *testing.T) {
	config := &v1alpha1.NutanixDatacenterConfig{
		Spec: v1alpha1.NutanixDatacenterConfigSpec{
			Endpoint: "test",
			Insecure: false,
			Port:     9440,
		},
	}

	os.Clearenv()
	if err := setupEnvVars(config); err == nil {
		t.Fatalf("setupEnvVars() err = nil, want err not nil: %#v", err)
	}
}

func TestSetEnvUsernameError(t *testing.T) {
	storedOSSetenv := osSetenv
	osSetenv = fakeOSSetenv(constants.NutanixUsernameKey)
	defer restoreOSSetenv(storedOSSetenv)

	config := &v1alpha1.NutanixDatacenterConfig{
		Spec: v1alpha1.NutanixDatacenterConfigSpec{
			Endpoint: "test",
			Insecure: false,
			Port:     9440,
		},
	}

	t.Setenv(constants.EksaNutanixUsernameKey, "test")
	t.Setenv(constants.EksaNutanixPasswordKey, "test")
	if err := setupEnvVars(config); err == nil {
		t.Fatalf("setupEnvVars() err = nil, want err not nil: %#v", err)
	}
}

func TestPasswordIsNotSetError(t *testing.T) {
	config := &v1alpha1.NutanixDatacenterConfig{
		Spec: v1alpha1.NutanixDatacenterConfigSpec{
			Endpoint: "test",
			Insecure: false,
			Port:     9440,
		},
	}

	os.Clearenv()
	t.Setenv(constants.EksaNutanixUsernameKey, "test")

	if err := setupEnvVars(config); err == nil {
		t.Fatalf("setupEnvVars() err = nil, want err not nil: %#v", err)
	}
}

func TestPasswordSetEnvVarError(t *testing.T) {
	storedOSSetenv := osSetenv
	osSetenv = fakeOSSetenv(constants.NutanixPasswordKey)
	defer restoreOSSetenv(storedOSSetenv)

	config := &v1alpha1.NutanixDatacenterConfig{
		Spec: v1alpha1.NutanixDatacenterConfigSpec{
			Endpoint: "test",
			Insecure: false,
			Port:     9440,
		},
	}

	t.Setenv(constants.EksaNutanixUsernameKey, "test")
	t.Setenv(constants.EksaNutanixPasswordKey, "test")
	if err := setupEnvVars(config); err == nil {
		t.Fatalf("setupEnvVars() err = nil, want err not nil: %#v", err)
	}
}

func TestSetEnvEndpointError(t *testing.T) {
	storedOSSetenv := osSetenv
	osSetenv = fakeOSSetenv(nutanixEndpointKey)
	defer restoreOSSetenv(storedOSSetenv)

	config := &v1alpha1.NutanixDatacenterConfig{
		Spec: v1alpha1.NutanixDatacenterConfigSpec{
			Endpoint: "test",
			Insecure: false,
			Port:     9440,
		},
	}

	t.Setenv(constants.EksaNutanixUsernameKey, "test")
	t.Setenv(constants.EksaNutanixPasswordKey, "test")
	if err := setupEnvVars(config); err == nil {
		t.Fatalf("setupEnvVars() err = nil, want err not nil: %#v", err)
	}
}

func TestSetEnvCRSKeyError(t *testing.T) {
	storedOSSetenv := osSetenv
	osSetenv = fakeOSSetenv(expClusterResourceSetKey)
	defer restoreOSSetenv(storedOSSetenv)

	config := &v1alpha1.NutanixDatacenterConfig{
		Spec: v1alpha1.NutanixDatacenterConfigSpec{
			Endpoint: "test",
			Insecure: false,
			Port:     9440,
		},
	}

	t.Setenv(constants.EksaNutanixUsernameKey, "test")
	t.Setenv(constants.EksaNutanixPasswordKey, "test")
	if err := setupEnvVars(config); err == nil {
		t.Fatalf("setupEnvVars() err = %v, want err not nil", err)
	}
}

func TestSetupEnvVarsSuccess(t *testing.T) {
	config := &v1alpha1.NutanixDatacenterConfig{
		Spec: v1alpha1.NutanixDatacenterConfigSpec{
			Endpoint: "test",
			Insecure: false,
			Port:     9440,
		},
	}

	os.Clearenv()
	t.Setenv(constants.EksaNutanixUsernameKey, "test")
	t.Setenv(constants.EksaNutanixPasswordKey, "test")

	if err := setupEnvVars(config); err != nil {
		t.Fatalf("setupEnvVars() err = %v, want err nil", err)
	}
}

func TestGetCredsFromEnv(t *testing.T) {
	os.Clearenv()
	t.Setenv(constants.EksaNutanixUsernameKey, "test")
	t.Setenv(constants.EksaNutanixPasswordKey, "test")

	creds := GetCredsFromEnv()

	if creds.PrismCentral.BasicAuth.Username != "test" {
		t.Fatalf("getCredsFromEnv() username = %s, want username test", creds.PrismCentral.BasicAuth.Username)
	}

	if creds.PrismCentral.BasicAuth.Password != "test" {
		t.Fatalf("getCredsFromEnv() password = %s, want password test", creds.PrismCentral.BasicAuth.Password)
	}
}

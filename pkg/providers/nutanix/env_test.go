package nutanix

import (
	"errors"
	"os"
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func fakeOSSetenv(key, value string) error {
	return errors.New("os.Setenv failed")
}

func restoreOSSetenv(replace func(key, value string) error) {
	osSetenv = replace
}

func TestSetupEnvVarsErrorDatacenter(t *testing.T) {
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

	t.Setenv(constants.EksaNutanixUsernameKey, "test")
	if err := setupEnvVars(config); err == nil {
		t.Fatalf("setupEnvVars() err = nil, want err not nil: %#v", err)
	}
}

func TestSetupEnvVarsErrorDatacenterSetenvFailures(t *testing.T) {
	storedOSSetenv := osSetenv
	osSetenv = fakeOSSetenv
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

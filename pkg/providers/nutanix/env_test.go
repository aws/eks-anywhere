package nutanix

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

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
}

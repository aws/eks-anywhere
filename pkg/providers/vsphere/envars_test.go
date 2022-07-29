package vsphere_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

func TestSetupEnvVarsErrorDatacenter(t *testing.T) {
	config := &v1alpha1.VSphereDatacenterConfig{
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Server:     "test",
			Insecure:   false,
			Datacenter: string([]byte{0}),
		},
	}
	if err := vsphere.SetupEnvVars(config); err == nil {
		t.Fatal("SetupEnvVars() err = nil, want err not nil")
	}
}

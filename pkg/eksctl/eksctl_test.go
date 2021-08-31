package eksctl_test

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/pkg/eksctl"
)

func TestValidateVersionSuccess(t *testing.T) {
	os.Setenv(eksctl.VersionEnvVar, "dev")
	err := eksctl.ValidateVersion()
	if err != nil {
		t.Fatalf("ValidateVersion() error = %v, wantErr <nil>", err)
	}
	os.Unsetenv(eksctl.VersionEnvVar)
}

func TestValidateVersionError(t *testing.T) {
	expected := "unable to retrieve version. Please use the 'eksctl anywhere' command to use EKS-A"
	err := eksctl.ValidateVersion()
	if err == nil {
		t.Fatalf("ValidateVersion() error = <nil>, want error = %v", expected)
	}
	actual := err.Error()
	if expected != actual {
		t.Fatalf("Expected=<%s> actual=<%s>", expected, actual)
	}
}

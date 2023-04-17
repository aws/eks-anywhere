package executables_test

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
)

func TestRedactCreds(t *testing.T) {
	str := "My username is username123. My password is password456"
	t.Setenv(constants.VSphereUsernameKey, "username123")
	os.Unsetenv(constants.VSpherePasswordKey)
	os.Unsetenv("var")
	envMap := map[string]string{"var": "value", constants.VSpherePasswordKey: "password456"}

	expected := "My username is *****. My password is *****"

	redactedStr := executables.RedactCreds(str, envMap)
	if redactedStr != expected {
		t.Fatalf("executables.RedactCreds expected = %s, got = %s", expected, redactedStr)
	}
}

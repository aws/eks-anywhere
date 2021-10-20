package features

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

const fakeFeatureEnvVar = "fakeFeatureEnvVar"

func fakeFeature() Feature {
	return Feature{
		Name:     "Core components upgrade",
		IsActive: isActiveForEnvVar(fakeFeatureEnvVar),
	}
}

func setupContext(t *testing.T) {
	envVarOrgValue, set := os.LookupEnv(fakeFeatureEnvVar)
	t.Cleanup(func() {
		// cleanup cache
		cache = newMutexMap()
		if set {
			os.Setenv(fakeFeatureEnvVar, envVarOrgValue)
		} else {
			os.Unsetenv(fakeFeatureEnvVar)
		}
	})
}

func TestIsActiveEnvVarUnset(t *testing.T) {
	g := NewWithT(t)
	setupContext(t)

	os.Unsetenv(fakeFeatureEnvVar)
	g.Expect(IsActive(fakeFeature())).To(BeFalse())
}

func TestIsActiveEnvVarSetFalse(t *testing.T) {
	g := NewWithT(t)
	setupContext(t)

	os.Setenv(fakeFeatureEnvVar, "false")
	g.Expect(IsActive(fakeFeature())).To(BeFalse())
}

func TestIsActiveEnvVarSetTrue(t *testing.T) {
	g := NewWithT(t)
	setupContext(t)

	g.Expect(os.Setenv(fakeFeatureEnvVar, "true")).To(Succeed())
	g.Expect(IsActive(fakeFeature())).To(BeTrue())
}

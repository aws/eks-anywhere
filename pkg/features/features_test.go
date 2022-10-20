package features

import (
	"fmt"
	"os"
	"sync"
	"testing"

	. "github.com/onsi/gomega"
)

const (
	fakeFeatureEnvVar = "fakeFeatureEnvVar"
	fakeFeatureGate   = "fakeFeatureGate"
)

func fakeFeature() Feature {
	return Feature{
		Name:     "Core components upgrade",
		IsActive: globalFeatures.isActiveForEnvVar(fakeFeatureEnvVar),
	}
}

func fakeFeatureWithGate() Feature {
	return Feature{
		Name:     "Core components upgrade",
		IsActive: globalFeatures.isActiveForEnvVarOrGate(fakeFeatureEnvVar, fakeFeatureGate),
	}
}

func setupContext(t *testing.T) {
	t.Cleanup(func() {
		// cleanup cache
		globalFeatures.cache = newMutexMap()
		globalFeatures.initGates = sync.Once{}
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

	t.Setenv(fakeFeatureEnvVar, "false")
	g.Expect(IsActive(fakeFeature())).To(BeFalse())
}

func TestIsActiveEnvVarSetTrue(t *testing.T) {
	g := NewWithT(t)
	setupContext(t)

	t.Setenv(fakeFeatureEnvVar, "true")
	g.Expect(IsActive(fakeFeature())).To(BeTrue())
}

func TestIsActiveWithFeatureGatesTrue(t *testing.T) {
	g := NewWithT(t)
	setupContext(t)

	featureGates := []string{"gate1=", "gate2=false", fmt.Sprintf("%s=true", fakeFeatureGate), ""}
	FeedGates(featureGates)

	g.Expect(IsActive(fakeFeatureWithGate())).To(BeTrue())
}

func TestWithNutanixFeatureFlag(t *testing.T) {
	g := NewWithT(t)
	setupContext(t)

	t.Setenv(NutanixProviderEnvVar, "true")
	g.Expect(IsActive(NutanixProvider())).To(BeTrue())
}

func TestWithK8s124FeatureFlag(t *testing.T) {
	g := NewWithT(t)
	setupContext(t)

	t.Setenv(K8s124SupportEnvVar, "true")
	g.Expect(IsActive(K8s124Support())).To(BeTrue())
}

package e2e

import (
	"testing"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

func TestPackagesUpgradeFromBundle(t *testing.T) {
	tests := []struct {
		name     string
		stage    string
		env      string
		wantFrom string
		wantOK   bool
	}{
		{name: "dev without the variable", stage: "dev", wantOK: false},
		{name: "staging defaults to prod", stage: "staging", wantFrom: "prod", wantOK: true},
		{name: "staging keeps an explicit bundle", stage: "staging", env: "v1-36-2026-10-30-1793381261", wantFrom: "v1-36-2026-10-30-1793381261", wantOK: true},
		{name: "dev with the variable", stage: "dev", env: "prod", wantFrom: "prod", wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(e2etests.PackagesUpgradeFromBundleVar, tt.env)
			from, ok := packagesUpgradeFromBundle(tt.stage)
			if from != tt.wantFrom || ok != tt.wantOK {
				t.Errorf("packagesUpgradeFromBundle(%q) = (%q, %t), want (%q, %t)", tt.stage, from, ok, tt.wantFrom, tt.wantOK)
			}
		})
	}
}

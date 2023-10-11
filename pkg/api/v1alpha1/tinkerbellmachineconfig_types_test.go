package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTinkerbellMachineConfigValidateSucceed(t *testing.T) {
	machineConfig := CreateTinkerbellMachineConfig()

	g := NewWithT(t)
	g.Expect(machineConfig.Validate()).To(Succeed())
}

func TestTinkerbellMachineConfigValidateFail(t *testing.T) {
	tests := []struct {
		name          string
		machineConfig *TinkerbellMachineConfig
		expectedErr   string
	}{
		{
			name: "Invalid object meta",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.ObjectMeta.Name = ""
			}),
			expectedErr: "TinkerbellMachineConfig: missing name",
		},
		{
			name: "Empty hardware selector",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.HardwareSelector = nil
			}),
			expectedErr: "TinkerbellMachineConfig: missing spec.hardwareSelector",
		},
		{
			name: "Multiple hardware selectors",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.HardwareSelector["type2"] = "cp2"
			}),
			expectedErr: "TinkerbellMachineConfig: spec.hardwareSelector must contain only 1 key-value pair",
		},
		{
			name: "Empty OS family",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.OSFamily = ""
			}),
			expectedErr: "TinkerbellMachineConfig: missing spec.osFamily",
		},
		{
			name: "Invalid OS family",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.OSFamily = "invalid OS"
			}),
			expectedErr: "unsupported spec.osFamily (invalid OS); Please use one of the following",
		},
		{
			name: "Invalid hostOSConfiguration",
			machineConfig: CreateTinkerbellMachineConfig(
				withHostOSConfiguration(
					&HostOSConfiguration{
						NTPConfiguration: &NTPConfiguration{},
					},
				),
			),
			expectedErr: "HostOSConfiguration is invalid for TinkerbellMachineConfig tinkerbellmachineconfig: NTPConfiguration.Servers can not be empty",
		},
		{
			name: "Invalid OS Image URL",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.OSImageURL = "test"
			}),
			expectedErr: "parsing osImageOverride: parse \"test\": invalid URI for request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tc.machineConfig.Validate()).To(MatchError(ContainSubstring(tc.expectedErr)))
		})
	}
}

type tinkerbellMachineConfigOpt func(mc *TinkerbellMachineConfig)

func withHostOSConfiguration(config *HostOSConfiguration) tinkerbellMachineConfigOpt {
	return func(mc *TinkerbellMachineConfig) {
		mc.Spec.HostOSConfiguration = config
	}
}

func CreateTinkerbellMachineConfig(options ...tinkerbellMachineConfigOpt) *TinkerbellMachineConfig {
	defaultMachineConfig := &TinkerbellMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tinkerbellmachineconfig",
		},
		Spec: TinkerbellMachineConfigSpec{
			HardwareSelector: map[string]string{
				"type1": "cp1",
			},
			OSFamily: Ubuntu,
			Users: []UserConfiguration{{
				Name:              "mySshUsername",
				SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
			}},
		},
	}

	for _, opt := range options {
		opt(defaultMachineConfig)
	}

	return defaultMachineConfig
}

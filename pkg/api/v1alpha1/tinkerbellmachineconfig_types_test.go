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

func TestTinkerbellMachineConfigValidateWithAffinitySucceed(t *testing.T) {
	machineConfig := CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
		mc.Spec.HardwareSelector = nil
		mc.Spec.HardwareAffinity = &HardwareAffinity{
			Required: []HardwareAffinityTerm{
				{
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{"type": "cp"},
					},
				},
			},
		}
	})

	g := NewWithT(t)
	g.Expect(machineConfig.Validate()).To(Succeed())
}

func TestTinkerbellMachineConfigValidateWithPreferredAffinitySucceed(t *testing.T) {
	machineConfig := CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
		mc.Spec.HardwareSelector = nil
		mc.Spec.HardwareAffinity = &HardwareAffinity{
			Required: []HardwareAffinityTerm{
				{
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{"type": "cp"},
					},
				},
			},
			Preferred: []WeightedHardwareAffinityTerm{
				{
					Weight: 50,
					HardwareAffinityTerm: HardwareAffinityTerm{
						LabelSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{"rack": "rack1"},
						},
					},
				},
			},
		}
	})

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
			name: "Neither hardware selector nor affinity specified",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.HardwareSelector = nil
			}),
			expectedErr: "TinkerbellMachineConfig: either hardwareSelector or hardwareAffinity must be specified",
		},
		{
			name: "Both hardware selector and affinity specified",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.HardwareAffinity = &HardwareAffinity{
					Required: []HardwareAffinityTerm{
						{
							LabelSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{"type": "cp"},
							},
						},
					},
				}
			}),
			expectedErr: "TinkerbellMachineConfig: hardwareSelector and hardwareAffinity are mutually exclusive",
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
		{
			name: "HardwareAffinity with empty required terms",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.HardwareSelector = nil
				mc.Spec.HardwareAffinity = &HardwareAffinity{
					Required: []HardwareAffinityTerm{},
				}
			}),
			expectedErr: "TinkerbellMachineConfig: hardwareAffinity.required must have at least one term",
		},
		{
			name: "HardwareAffinity with invalid preferred weight too low",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.HardwareSelector = nil
				mc.Spec.HardwareAffinity = &HardwareAffinity{
					Required: []HardwareAffinityTerm{
						{
							LabelSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{"type": "cp"},
							},
						},
					},
					Preferred: []WeightedHardwareAffinityTerm{
						{
							Weight: 0,
							HardwareAffinityTerm: HardwareAffinityTerm{
								LabelSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"rack": "rack1"},
								},
							},
						},
					},
				}
			}),
			expectedErr: "preferred term weight must be in range [1, 100], got 0",
		},
		{
			name: "HardwareAffinity with invalid preferred weight too high",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.HardwareSelector = nil
				mc.Spec.HardwareAffinity = &HardwareAffinity{
					Required: []HardwareAffinityTerm{
						{
							LabelSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{"type": "cp"},
							},
						},
					},
					Preferred: []WeightedHardwareAffinityTerm{
						{
							Weight: 101,
							HardwareAffinityTerm: HardwareAffinityTerm{
								LabelSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"rack": "rack1"},
								},
							},
						},
					},
				}
			}),
			expectedErr: "preferred term weight must be in range [1, 100], got 101",
		},
		{
			name: "HardwareAffinity with invalid matchExpression operator",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.HardwareSelector = nil
				mc.Spec.HardwareAffinity = &HardwareAffinity{
					Required: []HardwareAffinityTerm{
						{
							LabelSelector: metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "type",
										Operator: "InvalidOp",
										Values:   []string{"cp"},
									},
								},
							},
						},
					},
				}
			}),
			expectedErr: "invalid matchExpression operator 'InvalidOp'",
		},
		{
			name: "HardwareAffinity with In operator and empty values",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.HardwareSelector = nil
				mc.Spec.HardwareAffinity = &HardwareAffinity{
					Required: []HardwareAffinityTerm{
						{
							LabelSelector: metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "type",
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{},
									},
								},
							},
						},
					},
				}
			}),
			expectedErr: "matchExpression with operator In must have non-empty values",
		},
		{
			name: "HardwareAffinity with Exists operator and values",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.HardwareSelector = nil
				mc.Spec.HardwareAffinity = &HardwareAffinity{
					Required: []HardwareAffinityTerm{
						{
							LabelSelector: metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "type",
										Operator: metav1.LabelSelectorOpExists,
										Values:   []string{"should-not-be-here"},
									},
								},
							},
						},
					},
				}
			}),
			expectedErr: "matchExpression with operator Exists must not have values",
		},
		{
			name: "HardwareAffinity preferred term with invalid matchExpression",
			machineConfig: CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
				mc.Spec.HardwareSelector = nil
				mc.Spec.HardwareAffinity = &HardwareAffinity{
					Required: []HardwareAffinityTerm{
						{
							LabelSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{"type": "cp"},
							},
						},
					},
					Preferred: []WeightedHardwareAffinityTerm{
						{
							Weight: 50,
							HardwareAffinityTerm: HardwareAffinityTerm{
								LabelSelector: metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "rack",
											Operator: "InvalidOp",
											Values:   []string{"rack1"},
										},
									},
								},
							},
						},
					},
				}
			}),
			expectedErr: "invalid matchExpression operator 'InvalidOp' in preferred[0]",
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

func TestValidateHardwareAffinityOperator(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		expected bool
	}{
		{"In operator", "In", true},
		{"NotIn operator", "NotIn", true},
		{"Exists operator", "Exists", true},
		{"DoesNotExist operator", "DoesNotExist", true},
		{"Invalid operator", "Invalid", false},
		{"Empty operator", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(ValidateHardwareAffinityOperator(tc.operator)).To(Equal(tc.expected))
		})
	}
}

func TestValidateHardwareAffinityWeight(t *testing.T) {
	tests := []struct {
		name     string
		weight   int32
		expected bool
	}{
		{"Weight 1 (min)", 1, true},
		{"Weight 50 (mid)", 50, true},
		{"Weight 100 (max)", 100, true},
		{"Weight 0 (below min)", 0, false},
		{"Weight 101 (above max)", 101, false},
		{"Weight -1 (negative)", -1, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(ValidateHardwareAffinityWeight(tc.weight)).To(Equal(tc.expected))
		})
	}
}

func TestValidateLabelSelectorRequirement(t *testing.T) {
	tests := []struct {
		name        string
		requirement metav1.LabelSelectorRequirement
		expectErr   bool
		errContains string
	}{
		{
			name: "Valid In operator with values",
			requirement: metav1.LabelSelectorRequirement{
				Key:      "type",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"cp", "worker"},
			},
			expectErr: false,
		},
		{
			name: "Valid NotIn operator with values",
			requirement: metav1.LabelSelectorRequirement{
				Key:      "type",
				Operator: metav1.LabelSelectorOpNotIn,
				Values:   []string{"etcd"},
			},
			expectErr: false,
		},
		{
			name: "Valid Exists operator without values",
			requirement: metav1.LabelSelectorRequirement{
				Key:      "type",
				Operator: metav1.LabelSelectorOpExists,
			},
			expectErr: false,
		},
		{
			name: "Valid DoesNotExist operator without values",
			requirement: metav1.LabelSelectorRequirement{
				Key:      "type",
				Operator: metav1.LabelSelectorOpDoesNotExist,
			},
			expectErr: false,
		},
		{
			name: "Invalid operator",
			requirement: metav1.LabelSelectorRequirement{
				Key:      "type",
				Operator: "InvalidOp",
			},
			expectErr:   true,
			errContains: "invalid operator",
		},
		{
			name: "In operator with empty values",
			requirement: metav1.LabelSelectorRequirement{
				Key:      "type",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{},
			},
			expectErr:   true,
			errContains: "requires non-empty values",
		},
		{
			name: "NotIn operator with empty values",
			requirement: metav1.LabelSelectorRequirement{
				Key:      "type",
				Operator: metav1.LabelSelectorOpNotIn,
				Values:   []string{},
			},
			expectErr:   true,
			errContains: "requires non-empty values",
		},
		{
			name: "Exists operator with values",
			requirement: metav1.LabelSelectorRequirement{
				Key:      "type",
				Operator: metav1.LabelSelectorOpExists,
				Values:   []string{"should-not-be-here"},
			},
			expectErr:   true,
			errContains: "must not have values",
		},
		{
			name: "DoesNotExist operator with values",
			requirement: metav1.LabelSelectorRequirement{
				Key:      "type",
				Operator: metav1.LabelSelectorOpDoesNotExist,
				Values:   []string{"should-not-be-here"},
			},
			expectErr:   true,
			errContains: "must not have values",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			err := ValidateLabelSelectorRequirement(tc.requirement)
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tc.errContains))
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}

func TestTinkerbellMachineConfigValidateWithMatchExpressions(t *testing.T) {
	machineConfig := CreateTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
		mc.Spec.HardwareSelector = nil
		mc.Spec.HardwareAffinity = &HardwareAffinity{
			Required: []HardwareAffinityTerm{
				{
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{"type": "cp"},
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "rack",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"rack1", "rack2"},
							},
							{
								Key:      "vendor",
								Operator: metav1.LabelSelectorOpExists,
							},
						},
					},
				},
			},
		}
	})

	g := NewWithT(t)
	g.Expect(machineConfig.Validate()).To(Succeed())
}

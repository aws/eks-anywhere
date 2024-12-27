/*
Copyright 2022 The Tinkerbell Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1_test

import (
	"testing"

	. "github.com/onsi/gomega" //nolint:revive // one day we will remove gomega
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell/capt/v1beta1"
)

func Test_valid_tinkerbell_machine(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	existingValidMachine := &v1beta1.TinkerbellMachine{}

	for _, machine := range []v1beta1.TinkerbellMachine{
		// preferred affinity weight ranges
		{
			Spec: v1beta1.TinkerbellMachineSpec{
				HardwareAffinity: &v1beta1.HardwareAffinity{
					Preferred: []v1beta1.WeightedHardwareAffinityTerm{
						{
							Weight: 1,
							HardwareAffinityTerm: v1beta1.HardwareAffinityTerm{
								LabelSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"foo": "bar"},
								},
							},
						},
					},
				},
			},
		},
		{
			Spec: v1beta1.TinkerbellMachineSpec{
				HardwareAffinity: &v1beta1.HardwareAffinity{
					Preferred: []v1beta1.WeightedHardwareAffinityTerm{
						{
							Weight: 50,
							HardwareAffinityTerm: v1beta1.HardwareAffinityTerm{
								LabelSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"foo": "bar"},
								},
							},
						},
					},
				},
			},
		},
		{
			Spec: v1beta1.TinkerbellMachineSpec{
				HardwareAffinity: &v1beta1.HardwareAffinity{
					Preferred: []v1beta1.WeightedHardwareAffinityTerm{
						{
							Weight: 100,
							HardwareAffinityTerm: v1beta1.HardwareAffinityTerm{
								LabelSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"foo": "bar"},
								},
							},
						},
					},
				},
			},
		},
	} {
		_, err := machine.ValidateCreate()
		g.Expect(err).ToNot(HaveOccurred())
		_, err = machine.ValidateUpdate(existingValidMachine)
		g.Expect(err).ToNot(HaveOccurred())
	}
}

func Test_invalid_tinkerbell_machine(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	existingValidMachine := &v1beta1.TinkerbellMachine{}

	for _, machine := range []v1beta1.TinkerbellMachine{
		// invalid preferred affinity weight values
		{
			Spec: v1beta1.TinkerbellMachineSpec{
				HardwareAffinity: &v1beta1.HardwareAffinity{
					Preferred: []v1beta1.WeightedHardwareAffinityTerm{
						{
							Weight: -1,
							HardwareAffinityTerm: v1beta1.HardwareAffinityTerm{
								LabelSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"foo": "bar"},
								},
							},
						},
					},
				},
			},
		},
		{
			Spec: v1beta1.TinkerbellMachineSpec{
				HardwareAffinity: &v1beta1.HardwareAffinity{
					Preferred: []v1beta1.WeightedHardwareAffinityTerm{
						{
							Weight: 101,
							HardwareAffinityTerm: v1beta1.HardwareAffinityTerm{
								LabelSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"foo": "bar"},
								},
							},
						},
					},
				},
			},
		},
	} {
		_, err := machine.ValidateCreate()
		g.Expect(err).To(HaveOccurred())
		_, err = machine.ValidateUpdate(existingValidMachine)
		g.Expect(err).To(HaveOccurred())
	}
}

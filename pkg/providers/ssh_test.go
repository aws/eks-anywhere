package providers_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
)

func TestValidateSSHKeyPresentForUpgrade(t *testing.T) {
	testCases := []struct {
		name   string
		config *cluster.Config
		want   []string
	}{
		{
			name: "all machines have keys",
			config: &cluster.Config{
				VSphereMachineConfigs: map[string]*anywherev1.VSphereMachineConfig{
					"machine1": {
						Spec: anywherev1.VSphereMachineConfigSpec{
							Users: []anywherev1.UserConfiguration{
								{
									Name: "user1",
								},
								{
									Name: "user2",
									SshAuthorizedKeys: []string{
										"mykey",
									},
								},
							},
						},
					},
					"machine2": {
						Spec: anywherev1.VSphereMachineConfigSpec{
							Users: []anywherev1.UserConfiguration{
								{
									Name:              "user1",
									SshAuthorizedKeys: []string{""},
								},
								{
									Name: "user2",
									SshAuthorizedKeys: []string{
										"mykey",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "all machines are missing keys",
			config: &cluster.Config{
				VSphereMachineConfigs: map[string]*anywherev1.VSphereMachineConfig{
					"machine1": {
						TypeMeta: metav1.TypeMeta{
							Kind: "VSphereMachineConfig",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "machine1",
						},
						Spec: anywherev1.VSphereMachineConfigSpec{
							Users: []anywherev1.UserConfiguration{
								{
									Name: "user1",
								},
								{
									Name:              "user2",
									SshAuthorizedKeys: []string{""},
								},
							},
						},
					},
					"machine2": {
						TypeMeta: metav1.TypeMeta{
							Kind: "VSphereMachineConfig",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "machine2",
						},
						Spec: anywherev1.VSphereMachineConfigSpec{
							Users: []anywherev1.UserConfiguration{
								{
									Name:              "user1",
									SshAuthorizedKeys: []string{""},
								},
							},
						},
					},
				},
				CloudStackMachineConfigs: map[string]*anywherev1.CloudStackMachineConfig{
					"machine1": {
						TypeMeta: metav1.TypeMeta{
							Kind: "CloudStackMachineConfig",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "machine1",
						},
					},
				},
				NutanixMachineConfigs: map[string]*anywherev1.NutanixMachineConfig{
					"machine1": {
						TypeMeta: metav1.TypeMeta{
							Kind: "NutanixMachineConfig",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "machine1",
						},
					},
				},
				TinkerbellMachineConfigs: map[string]*anywherev1.TinkerbellMachineConfig{
					"machine1": {
						TypeMeta: metav1.TypeMeta{
							Kind: "TinkerbellMachineConfig",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "machine1",
						},
					},
				},
			},
			want: []string{
				"machine1 VSphereMachineConfig is invalid: it should contain at least one SSH key",
				"machine2 VSphereMachineConfig is invalid: it should contain at least one SSH key",
				"machine1 CloudStackMachineConfig is invalid: it should contain at least one SSH key",
				"machine1 NutanixMachineConfig is invalid: it should contain at least one SSH key",
				"machine1 TinkerbellMachineConfig is invalid: it should contain at least one SSH key",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			g := NewWithT(t)
			spec := &cluster.Spec{
				Config: tc.config,
			}

			if len(tc.want) == 0 {
				g.Expect(providers.ValidateSSHKeyPresentForUpgrade(ctx, spec)).To(Succeed())
			} else {
				err := providers.ValidateSSHKeyPresentForUpgrade(ctx, spec)
				for _, errMsg := range tc.want {
					g.Expect(err).To(MatchError(ContainSubstring(errMsg)))
				}
			}
		})
	}
}

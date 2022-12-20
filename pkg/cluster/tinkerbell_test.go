package cluster_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestParseConfigFromFileTinkerbellCluster(t *testing.T) {
	g := NewWithT(t)
	got, err := cluster.ParseConfigFromFile("testdata/cluster_tinkerbell_1_19.yaml")

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(got.Cluster).To(BeComparableTo(
		&anywherev1.Cluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Cluster",
				APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test-namespace",
			},
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: "1.19",
				ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
					Count:    1,
					Endpoint: &anywherev1.Endpoint{Host: "1.2.3.4"},
					MachineGroupRef: &anywherev1.Ref{
						Kind: "TinkerbellMachineConfig",
						Name: "test-cp",
					},
					Taints:                 nil,
					Labels:                 nil,
					UpgradeRolloutStrategy: nil,
				},
				WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
					{
						Count: ptr.Int(1),
						MachineGroupRef: &anywherev1.Ref{
							Kind: "TinkerbellMachineConfig",
							Name: "test-md",
						},
					},
				},
				DatacenterRef: anywherev1.Ref{
					Kind: "TinkerbellDatacenterConfig",
					Name: "test",
				},
				IdentityProviderRefs: nil,
				GitOpsRef:            nil,
				ClusterNetwork: anywherev1.ClusterNetwork{
					Pods: anywherev1.Pods{
						CidrBlocks: []string{"192.168.0.0/16"},
					},
					Services: anywherev1.Services{
						CidrBlocks: []string{"10.96.0.0/12"},
					},
					CNI: "cilium",
				},
				ExternalEtcdConfiguration: &anywherev1.ExternalEtcdConfiguration{
					Count: 1,
					MachineGroupRef: &anywherev1.Ref{
						Kind: "TinkerbellMachineConfig",
						Name: "test-cp",
					},
				},
				ManagementCluster: anywherev1.ManagementCluster{Name: "test"},
			},
		},
	))
	g.Expect(got.TinkerbellDatacenter).To(BeComparableTo(
		&anywherev1.TinkerbellDatacenterConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "TinkerbellDatacenterConfig",
				APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: anywherev1.TinkerbellDatacenterConfigSpec{
				TinkerbellIP: "1.2.3.4",
			},
		},
	))

	g.Expect(got.TinkerbellMachineConfigs).To(HaveLen(1))
	g.Expect(got.TinkerbellMachineConfigs["test-cp"]).To(
		BeComparableTo(
			&anywherev1.TinkerbellMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TinkerbellMachineConfig",
					APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cp",
					Namespace: "test-namespace",
				},
				Spec: anywherev1.TinkerbellMachineConfigSpec{
					TemplateRef: anywherev1.Ref{
						Kind: "TinkerbellTemplateConfig",
						Name: "tink-test",
					},
					OSFamily: "ubuntu",
					Users: []anywherev1.UserConfiguration{
						{
							Name:              "tink-user",
							SshAuthorizedKeys: []string{"ssh-rsa AAAAB3"},
						},
					},
				},
			},
		),
	)

	g.Expect(got.TinkerbellTemplateConfigs).To(HaveLen(1))
	g.Expect(got.TinkerbellTemplateConfigs["tink-test"]).To(
		BeComparableTo(
			&anywherev1.TinkerbellTemplateConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TinkerbellTemplateConfig",
					APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "tink-test",
				},
				Spec: anywherev1.TinkerbellTemplateConfigSpec{
					Template: tinkerbell.Workflow{
						Version:       "0.1",
						Name:          "tink-test",
						GlobalTimeout: 6000,
						Tasks: []tinkerbell.Task{
							{
								Name:       "tink-test",
								WorkerAddr: "{{.device_1}}",
								Actions: []tinkerbell.Action{
									{
										Name:    "stream-image",
										Image:   "image2disk:v1.0.0",
										Timeout: 360,
										Environment: map[string]string{
											"COMPRESSED": "true",
											"DEST_DISK":  "/dev/sda",
											"IMG_URL":    "",
										},
									},
								},
								Volumes: []string{
									"/dev:/dev",
									"/dev/console:/dev/console",
									"/lib/firmware:/lib/firmware:ro",
								},
							},
						},
					},
				},
			},
		),
	)
}

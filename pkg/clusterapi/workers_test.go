package clusterapi_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bootstrapv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
)

type (
	dockerGroup   = clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]
	dockerWorkers = clusterapi.Workers[*dockerv1.DockerMachineTemplate]
)

func TestWorkersUpdateImmutableObjectNamesError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	group1 := dockerGroup{
		MachineDeployment: machineDeployment(),
	}
	group2 := dockerGroup{
		MachineDeployment: machineDeployment(),
	}

	workers := &dockerWorkers{
		Groups: []dockerGroup{group1, group2},
	}
	client := test.NewFakeKubeClientAlwaysError()

	g.Expect(
		workers.UpdateImmutableObjectNames(ctx, client, dummyRetriever, noChangesCompare),
	).NotTo(Succeed())
}

func TestWorkersUpdateImmutableObjectNamesSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	group1 := dockerGroup{
		MachineDeployment: machineDeployment(),
	}
	group2 := dockerGroup{
		MachineDeployment: machineDeployment(),
	}

	workers := &dockerWorkers{
		Groups: []dockerGroup{group1, group2},
	}
	client := test.NewFakeKubeClient()

	g.Expect(
		workers.UpdateImmutableObjectNames(ctx, client, dummyRetriever, noChangesCompare),
	).To(Succeed())
}

func TestWorkerObjects(t *testing.T) {
	g := NewWithT(t)
	group1 := dockerGroup{
		MachineDeployment: machineDeployment(),
	}
	group2 := dockerGroup{
		MachineDeployment: machineDeployment(),
	}

	workers := &dockerWorkers{
		Groups: []dockerGroup{group1, group2},
	}

	objects := workers.WorkerObjects()
	wantObjects := []kubernetes.Object{
		group1.KubeadmConfigTemplate,
		group1.MachineDeployment,
		group1.ProviderMachineTemplate,
		group2.KubeadmConfigTemplate,
		group2.MachineDeployment,
		group2.ProviderMachineTemplate,
	}

	g.Expect(objects).To(ConsistOf(wantObjects))
}

func TestWorkerGroupUpdateImmutableObjectNamesNoMachineDeployment(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	group := &dockerGroup{
		MachineDeployment: machineDeployment(),
	}
	client := test.NewFakeKubeClient()

	g.Expect(group.UpdateImmutableObjectNames(ctx, client, dummyRetriever, noChangesCompare)).To(Succeed())
}

func TestWorkerGroupUpdateImmutableObjectNamesErrorReadingMachineDeployment(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	group := &dockerGroup{
		MachineDeployment: machineDeployment(),
	}
	client := test.NewFakeKubeClientAlwaysError()

	g.Expect(
		group.UpdateImmutableObjectNames(ctx, client, dummyRetriever, noChangesCompare),
	).To(MatchError(ContainSubstring("reading current machine deployment from API")))
}

func TestWorkerGroupUpdateImmutableObjectNamesErrorUpdatingMachineTemplateName(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	group := &dockerGroup{
		MachineDeployment:       machineDeployment(),
		ProviderMachineTemplate: dockerMachineTemplate(),
		KubeadmConfigTemplate:   kubeadmConfigTemplate(),
	}
	group.MachineDeployment.Spec.Template.Spec.InfrastructureRef = contractReference(group.ProviderMachineTemplate)
	group.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef = contractReference(group.KubeadmConfigTemplate)
	client := test.NewFakeKubeClient(group.MachineDeployment)

	g.Expect(
		group.UpdateImmutableObjectNames(ctx, client, errorRetriever, noChangesCompare),
	).To(MatchError(ContainSubstring("reading DockerMachineTemplate eksa-system/mt-1 from API")))
}

func TestWorkerGroupUpdateImmutableObjectNamesErrorUpdatingKubeadmConfigTemplate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	group := &dockerGroup{
		MachineDeployment:       machineDeployment(),
		ProviderMachineTemplate: dockerMachineTemplate(),
		KubeadmConfigTemplate:   kubeadmConfigTemplate(),
	}
	group.MachineDeployment.Spec.Template.Spec.InfrastructureRef = contractReference(group.ProviderMachineTemplate)

	// Set TypeMeta on the object being tested to get proper error message with Kind
	group.KubeadmConfigTemplate.TypeMeta = metav1.TypeMeta{
		Kind:       "KubeadmConfigTemplate",
		APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
	}
	group.KubeadmConfigTemplate.Name = "invalid-name"
	group.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef = contractReference(group.KubeadmConfigTemplate)
	client := test.NewFakeKubeClient(group.MachineDeployment, group.KubeadmConfigTemplate, group.ProviderMachineTemplate)
	group.KubeadmConfigTemplate.Spec.Template.Spec.PostKubeadmCommands = []string{"ls"}

	g.Expect(
		group.UpdateImmutableObjectNames(ctx, client, dummyRetriever, noChangesCompare),
	).To(MatchError(ContainSubstring("incrementing name for KubeadmConfigTemplate eksa-system/invalid-name")))
}

func TestWorkerGroupUpdateImmutableObjectNamesSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	group := &dockerGroup{
		MachineDeployment:       machineDeployment(),
		ProviderMachineTemplate: dockerMachineTemplate(),
		KubeadmConfigTemplate:   kubeadmConfigTemplate(),
	}
	group.MachineDeployment.Spec.Template.Spec.InfrastructureRef = contractReference(group.ProviderMachineTemplate)
	group.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef = contractReference(group.KubeadmConfigTemplate)
	client := test.NewFakeKubeClient(group.MachineDeployment, group.KubeadmConfigTemplate, group.ProviderMachineTemplate)
	group.KubeadmConfigTemplate.Spec.Template.Spec.PostKubeadmCommands = []string{"ls"}

	g.Expect(
		group.UpdateImmutableObjectNames(ctx, client, dummyRetriever, withChangesCompare),
	).To(Succeed())
	g.Expect(group.KubeadmConfigTemplate.Name).To(Equal("template-2"))
	g.Expect(group.ProviderMachineTemplate.Name).To(Equal("mt-2"))
	g.Expect(group.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef.Name).To(Equal(group.KubeadmConfigTemplate.Name))
	g.Expect(group.MachineDeployment.Spec.Template.Spec.InfrastructureRef.Name).To(Equal(group.ProviderMachineTemplate.Name))
}

func TestGetKubeadmConfigTemplateSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	k := kubeadmConfigTemplate()
	client := test.NewFakeKubeClient(k)

	g.Expect(clusterapi.GetKubeadmConfigTemplate(ctx, client, k.Name, k.Namespace)).To(Equal(k))
}

func TestGetKubeadmConfigTemplateError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	k := kubeadmConfigTemplate()
	client := test.NewFakeKubeClientAlwaysError()

	_, err := clusterapi.GetKubeadmConfigTemplate(ctx, client, k.Name, k.Namespace)
	g.Expect(err).To(HaveOccurred())
}

func TestKubeadmConfigTemplateEqual(t *testing.T) {
	tests := []struct {
		name     string
		new, old *bootstrapv1beta2.KubeadmConfigTemplate
		want     bool
	}{
		{
			name: "equal",
			new: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									Taints: &[]corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									Taints: &[]corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "diff taints",
			new: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									Taints: &[]corev1.Taint{},
								},
							},
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									Taints: &[]corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "diff labels",
			new: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									KubeletExtraArgs: clusterapi.ExtraArgs{
										"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
										"cgroup-driver":     "cgroupfs",
										"eviction-hard":     "nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%",
									}.ToArgs(),
								},
							},
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									KubeletExtraArgs: clusterapi.ExtraArgs{
										"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
										"cgroup-driver":     "cgroupfs",
										"eviction-hard":     "nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%",
										"node-labels":       "foo-bar",
									}.ToArgs(),
								},
							},
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "new JoinConfiguration nil",
			new: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									Taints: &[]corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "old JoinConfiguration nil",
			new: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									Taints: &[]corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "diff spec",
			new: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									Taints: &[]corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									Taints: &[]corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []bootstrapv1beta2.File{
								{
									Owner: "you",
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "diff spec files",
			new: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									Taints: &[]corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []bootstrapv1beta2.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &bootstrapv1beta2.KubeadmConfigTemplate{
				Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
					Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
						Spec: bootstrapv1beta2.KubeadmConfigSpec{
							JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
								NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
									Taints: &[]corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(clusterapi.KubeadmConfigTemplateEqual(tt.new, tt.old)).To(Equal(tt.want))
		})
	}
}

func TestWorkerGroupDeepCopy(t *testing.T) {
	g := NewWithT(t)
	group := &dockerGroup{
		MachineDeployment:       machineDeployment(),
		KubeadmConfigTemplate:   kubeadmConfigTemplate(),
		ProviderMachineTemplate: dockerMachineTemplate(),
	}

	g.Expect(group.DeepCopy()).To(Equal(group))
}

func kubeadmConfigTemplate() *bootstrapv1beta2.KubeadmConfigTemplate {
	return &bootstrapv1beta2.KubeadmConfigTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "template-1",
			Namespace:       constants.EksaSystemNamespace,
			ResourceVersion: "1",
		},
		Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
			Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
				Spec: bootstrapv1beta2.KubeadmConfigSpec{
					Files: []bootstrapv1beta2.File{
						{
							Owner: "me",
						},
					},
				},
			},
		},
	}
}

func machineDeployment() *clusterv1beta2.MachineDeployment {
	return &clusterv1beta2.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineDeployment",
			APIVersion: "cluster.x-k8s.io/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deployment",
			Namespace: constants.EksaSystemNamespace,
		},
	}
}

func objectReference(obj client.Object) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
		APIVersion: obj.GetObjectKind().GroupVersionKind().Version,
		Name:       obj.GetName(),
		Namespace:  obj.GetNamespace(),
	}
}

func contractReference(obj client.Object) clusterv1beta2.ContractVersionedObjectReference {
	return clusterv1beta2.ContractVersionedObjectReference{
		Kind:     obj.GetObjectKind().GroupVersionKind().Kind,
		APIGroup: obj.GetObjectKind().GroupVersionKind().Group,
		Name:     obj.GetName(),
	}
}

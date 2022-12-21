package clusterapi_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
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
	group.MachineDeployment.Spec.Template.Spec.InfrastructureRef = *objectReference(group.ProviderMachineTemplate)
	group.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef = objectReference(group.KubeadmConfigTemplate)
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
	group.MachineDeployment.Spec.Template.Spec.InfrastructureRef = *objectReference(group.ProviderMachineTemplate)
	group.KubeadmConfigTemplate.Name = "invalid-name"
	group.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef = objectReference(group.KubeadmConfigTemplate)
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
	group.MachineDeployment.Spec.Template.Spec.InfrastructureRef = *objectReference(group.ProviderMachineTemplate)
	group.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef = objectReference(group.KubeadmConfigTemplate)
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
		new, old *kubeadmv1.KubeadmConfigTemplate
		want     bool
	}{
		{
			name: "equal",
			new: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							JoinConfiguration: &kubeadmv1.JoinConfiguration{
								NodeRegistration: kubeadmv1.NodeRegistrationOptions{
									Taints: []corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []kubeadmv1.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							JoinConfiguration: &kubeadmv1.JoinConfiguration{
								NodeRegistration: kubeadmv1.NodeRegistrationOptions{
									Taints: []corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []kubeadmv1.File{
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
			new: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							JoinConfiguration: &kubeadmv1.JoinConfiguration{
								NodeRegistration: kubeadmv1.NodeRegistrationOptions{
									Taints: []corev1.Taint{},
								},
							},
							Files: []kubeadmv1.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							JoinConfiguration: &kubeadmv1.JoinConfiguration{
								NodeRegistration: kubeadmv1.NodeRegistrationOptions{
									Taints: []corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []kubeadmv1.File{
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
			new: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							JoinConfiguration: &kubeadmv1.JoinConfiguration{
								NodeRegistration: kubeadmv1.NodeRegistrationOptions{
									KubeletExtraArgs: map[string]string{
										"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
										"cgroup-driver":     "cgroupfs",
										"eviction-hard":     "nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%",
									},
								},
							},
							Files: []kubeadmv1.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							JoinConfiguration: &kubeadmv1.JoinConfiguration{
								NodeRegistration: kubeadmv1.NodeRegistrationOptions{
									KubeletExtraArgs: map[string]string{
										"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
										"cgroup-driver":     "cgroupfs",
										"eviction-hard":     "nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%",
										"node-labels":       "foo-bar",
									},
								},
							},
							Files: []kubeadmv1.File{
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
			new: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							Files: []kubeadmv1.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							JoinConfiguration: &kubeadmv1.JoinConfiguration{
								NodeRegistration: kubeadmv1.NodeRegistrationOptions{
									Taints: []corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []kubeadmv1.File{
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
			name: "old JoinConfiguration nil",
			new: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							JoinConfiguration: &kubeadmv1.JoinConfiguration{
								NodeRegistration: kubeadmv1.NodeRegistrationOptions{
									Taints: []corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []kubeadmv1.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							Files: []kubeadmv1.File{
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
			new: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							JoinConfiguration: &kubeadmv1.JoinConfiguration{
								NodeRegistration: kubeadmv1.NodeRegistrationOptions{
									Taints: []corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []kubeadmv1.File{
								{
									Owner: "me",
								},
							},
						},
					},
				},
			},
			old: &kubeadmv1.KubeadmConfigTemplate{
				Spec: kubeadmv1.KubeadmConfigTemplateSpec{
					Template: kubeadmv1.KubeadmConfigTemplateResource{
						Spec: kubeadmv1.KubeadmConfigSpec{
							JoinConfiguration: &kubeadmv1.JoinConfiguration{
								NodeRegistration: kubeadmv1.NodeRegistrationOptions{
									Taints: []corev1.Taint{
										{
											Key: "key",
										},
									},
								},
							},
							Files: []kubeadmv1.File{
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

func kubeadmConfigTemplate() *kubeadmv1.KubeadmConfigTemplate {
	return &kubeadmv1.KubeadmConfigTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmConfigTemplate",
			APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "template-1",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: kubeadmv1.KubeadmConfigTemplateSpec{
			Template: kubeadmv1.KubeadmConfigTemplateResource{
				Spec: kubeadmv1.KubeadmConfigSpec{
					Files: []kubeadmv1.File{
						{
							Owner: "me",
						},
					},
				},
			},
		},
	}
}

func machineDeployment() *clusterv1.MachineDeployment {
	return &clusterv1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineDeployment",
			APIVersion: "cluster.x-k8s.io/v1beta1",
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

package yaml_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type (
	dockerWorkers     = clusterapi.Workers[*dockerv1.DockerMachineTemplate]
	dockerWorkerGroup = clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]
)

func TestNewWorkersParserAndBuilderSuccessParsing(t *testing.T) {
	g := NewWithT(t)
	parser, builder, err := yaml.NewWorkersParserAndBuilder(
		test.NewNullLogger(),
		yamlutil.NewMapping(
			"DockerMachineTemplate",
			func() *dockerv1.DockerMachineTemplate {
				return &dockerv1.DockerMachineTemplate{}
			},
		),
	)

	g.Expect(err).To(BeNil())

	yaml := []byte(`apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
kind: KubeadmConfigTemplate
metadata:
  name: workers-1
  namespace: eksa-system
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  name: workers-1
  namespace: eksa-system
spec:
  template:
    spec:
      bootstrap:
        configRef:
          apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
          kind: KubeadmConfigTemplate
          name: workers-1
          namespace: eksa-system
      infrastructureRef:
        apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
        kind: DockerMachineTemplate
        name: workers-1
        namespace: eksa-system
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: DockerMachineTemplate
metadata:
  name: workers-1
  namespace: eksa-system
---
apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
kind: KubeadmConfigTemplate
metadata:
  name: workers-2
  namespace: eksa-system
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  name: workers-2
  namespace: eksa-system
spec:
  template:
    spec:
      bootstrap:
        configRef:
          apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
          kind: KubeadmConfigTemplate
          name: workers-2
          namespace: eksa-system
      infrastructureRef:
        apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
        kind: DockerMachineTemplate
        name: workers-2
        namespace: eksa-system
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: DockerMachineTemplate
metadata:
  name: workers-2
  namespace: eksa-system
`)

	g.Expect(parser.Parse(yaml, builder)).To(Succeed())
	g.Expect(builder.Workers.Groups).To(
		ConsistOf(*group("workers-1"), *group("workers-2")),
	)
}

func TestNewWorkersParserAndBuilderErrorFromMappings(t *testing.T) {
	g := NewWithT(t)
	_, _, err := yaml.NewWorkersParserAndBuilder(
		test.NewNullLogger(),
		yamlutil.NewMapping(
			"MachineDeployment",
			func() *dockerv1.DockerMachineTemplate {
				return &dockerv1.DockerMachineTemplate{}
			},
		),
	)

	g.Expect(err).To(MatchError(ContainSubstring("registering provider worker mappings")))
}

func TestRegisterWorkerMappingsError(t *testing.T) {
	g := NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	g.Expect(
		parser.RegisterMapping("KubeadmConfigTemplate", func() yamlutil.APIObject { return nil }),
	).To(Succeed())
	g.Expect(
		yaml.RegisterWorkerMappings(parser),
	).To(
		MatchError(ContainSubstring("registering base worker mappings")),
	)
}

func TestRegisterWorkerMappingsSuccess(t *testing.T) {
	g := NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	g.Expect(
		yaml.RegisterWorkerMappings(parser),
	).To(Succeed())
}

func TestProcessWorkerObjects(t *testing.T) {
	g := NewWithT(t)
	wantGroup1 := group("workers-1")
	wantGroup2 := group("workers-2")
	lookup := yamlutil.NewObjectLookupBuilder().Add(
		wantGroup1.MachineDeployment,
		wantGroup1.KubeadmConfigTemplate,
		wantGroup1.ProviderMachineTemplate,
		wantGroup2.MachineDeployment,
		wantGroup2.KubeadmConfigTemplate,
		wantGroup2.ProviderMachineTemplate,
	).Build()
	w := &dockerWorkers{}

	yaml.ProcessWorkerObjects(w, lookup)
	g.Expect(w.Groups).To(ConsistOf(*wantGroup1, *wantGroup2))
}

func TestProcessWorkerGroupObjectsNoKubeadmConfigTemplate(t *testing.T) {
	g := NewWithT(t)
	group := group("workers-1")
	group.KubeadmConfigTemplate = nil
	lookup := yamlutil.NewObjectLookupBuilder().Add(group.MachineDeployment).Build()

	yaml.ProcessWorkerGroupObjects(group, lookup)
	g.Expect(group.KubeadmConfigTemplate).To(BeNil())
}

func TestProcessWorkerGroupObjectsNoMachineTemplate(t *testing.T) {
	g := NewWithT(t)
	group := group("workers-1")
	group.ProviderMachineTemplate = nil
	lookup := yamlutil.NewObjectLookupBuilder().Add(group.MachineDeployment).Build()

	yaml.ProcessWorkerGroupObjects(group, lookup)
	g.Expect(group.ProviderMachineTemplate).To(BeNil())
}

func TestProcessWorkerGroupObjects(t *testing.T) {
	g := NewWithT(t)
	group := group("workers-1")
	kct := group.KubeadmConfigTemplate
	mt := group.ProviderMachineTemplate
	group.KubeadmConfigTemplate = nil
	group.ProviderMachineTemplate = nil

	lookup := yamlutil.NewObjectLookupBuilder().Add(
		group.MachineDeployment,
		kct,
		mt,
	).Build()

	yaml.ProcessWorkerGroupObjects(group, lookup)
	g.Expect(group.KubeadmConfigTemplate).To(Equal(kct))
	g.Expect(group.ProviderMachineTemplate).To(Equal(mt))
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

func group(baseName string) *dockerWorkerGroup {
	md := machineDeployment()
	md.Name = baseName
	kct := kubeadmConfigTemplate()
	kct.Name = baseName
	dmt := dockerMachineTemplate(baseName)

	md.Spec.Template.Spec.Bootstrap.ConfigRef = objectReference(kct)
	md.Spec.Template.Spec.InfrastructureRef = *objectReference(dmt)

	return &dockerWorkerGroup{
		MachineDeployment:       md,
		KubeadmConfigTemplate:   kct,
		ProviderMachineTemplate: dmt,
	}
}

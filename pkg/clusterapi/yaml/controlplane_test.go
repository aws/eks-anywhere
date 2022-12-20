package yaml_test

import (
	"testing"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type dockerControlPlane = clusterapi.ControlPlane[*dockerv1.DockerCluster, *dockerv1.DockerMachineTemplate]

func TestNewControlPlaneParserAndBuilderSuccessParsing(t *testing.T) {
	g := NewWithT(t)
	parser, builder, err := yaml.NewControlPlaneParserAndBuilder(
		test.NewNullLogger(),
		yamlutil.NewMapping(
			"DockerCluster",
			func() *dockerv1.DockerCluster {
				return &dockerv1.DockerCluster{}
			},
		),
		yamlutil.NewMapping(
			"DockerMachineTemplate",
			func() *dockerv1.DockerMachineTemplate {
				return &dockerv1.DockerMachineTemplate{}
			},
		),
	)

	g.Expect(err).To(Succeed())

	yaml := []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster
  namespace: eksa-system
spec:
  controlPlaneRef:
    apiVersion: controlplane.clusterapi.k8s/v1beta1
    kind: KubeadmControlPlane
    name: cp
    namespace: eksa-system
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
    kind: DockerCluster
    name: cluster
    namespace: eksa-system
`)

	g.Expect(parser.Parse(yaml, builder)).To(Succeed())
	g.Expect(builder.ControlPlane.Cluster).To(Equal(capiCluster()))
}

func TestNewControlPlaneParserAndBuilderErrorFromMappings(t *testing.T) {
	g := NewWithT(t)
	_, _, err := yaml.NewControlPlaneParserAndBuilder(
		test.NewNullLogger(),
		yamlutil.NewMapping(
			"Cluster",
			func() *dockerv1.DockerCluster {
				return &dockerv1.DockerCluster{}
			},
		),
		yamlutil.NewMapping(
			"DockerMachineTemplate",
			func() *dockerv1.DockerMachineTemplate {
				return &dockerv1.DockerMachineTemplate{}
			},
		),
	)

	g.Expect(err).To(MatchError(ContainSubstring("registering provider control plane mappings")))
}

func TestRegisterControlPlaneMappingsError(t *testing.T) {
	g := NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	g.Expect(parser.RegisterMapping("Cluster", func() yamlutil.APIObject { return nil })).To(Succeed())
	g.Expect(yaml.RegisterControlPlaneMappings(parser)).To(MatchError(ContainSubstring("registering base control plane mappings")))
}

func TestRegisterControlPlaneSuccess(t *testing.T) {
	g := NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	g.Expect(yaml.RegisterControlPlaneMappings(parser)).To(Succeed())
}

func TestProcessControlPlaneObjectsNoCluster(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	lookup := yamlutil.NewObjectLookupBuilder().Add(dockerCluster()).Build()

	yaml.ProcessControlPlaneObjects(cp, lookup)
	g.Expect(cp.Cluster).To(BeNil())
}

func TestProcessControlPlaneObjects(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	cluster := capiCluster()
	providerCluster := dockerCluster()
	kubeadmCP := kubeadmControlPlane()
	cpMachineTemplate := dockerMachineTemplate("cp-mt")
	etcdCluster := etcdCluster()
	cluster.Spec.ManagedExternalEtcdRef = objectReference(etcdCluster)
	etcdMachineTemplate := dockerMachineTemplate("etcd-mt")
	etcdCluster.Spec.InfrastructureTemplate = *objectReference(etcdMachineTemplate)
	lookup := yamlutil.NewObjectLookupBuilder().Add(
		cluster,
		providerCluster,
		kubeadmCP,
		cpMachineTemplate,
		etcdCluster,
		etcdMachineTemplate,
	).Build()

	yaml.ProcessControlPlaneObjects(cp, lookup)
	g.Expect(cp.Cluster).To(Equal(cluster))
	g.Expect(cp.ProviderCluster).To(Equal(providerCluster))
	g.Expect(cp.KubeadmControlPlane).To(Equal(kubeadmCP))
	g.Expect(cp.ControlPlaneMachineTemplate).To(Equal(cpMachineTemplate))
	g.Expect(cp.EtcdCluster).To(Equal(etcdCluster))
	g.Expect(cp.EtcdMachineTemplate).To(Equal(etcdMachineTemplate))
}

func TestProcessClusterNoCluster(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	lookup := yamlutil.NewObjectLookupBuilder().Add(dockerCluster()).Build()

	yaml.ProcessCluster(cp, lookup)
	g.Expect(cp.Cluster).To(BeNil())
}

func TestProcessClusterWithCluster(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	cluster := capiCluster()
	lookup := yamlutil.NewObjectLookupBuilder().Add(cluster).Build()

	yaml.ProcessCluster(cp, lookup)
	g.Expect(cp.Cluster).To(Equal(cluster))
}

func TestProcessProviderClusterWithNoCluster(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	cluster := capiCluster()
	cp.Cluster = cluster
	lookup := yamlutil.NewObjectLookupBuilder().Add(cluster).Build()

	yaml.ProcessProviderCluster(cp, lookup)
	g.Expect(cp.ProviderCluster).To(BeNil())
}

func TestProcessProviderClusterWithCluster(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	cluster := capiCluster()
	cp.Cluster = cluster
	providerCluster := dockerCluster()
	lookup := yamlutil.NewObjectLookupBuilder().Add(providerCluster).Build()

	yaml.ProcessProviderCluster(cp, lookup)
	g.Expect(cp.ProviderCluster).To(Equal(providerCluster))
}

func TestProcessKubeadmControlPlaneNoControlPlane(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	cp.Cluster = capiCluster()
	lookup := yamlutil.NewObjectLookupBuilder().Add(dockerCluster()).Build()

	yaml.ProcessKubeadmControlPlane(cp, lookup)
	g.Expect(cp.KubeadmControlPlane).To(BeNil())
}

func TestProcessKubeadmControlPlaneNoMachineTemplate(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	cp.Cluster = capiCluster()
	kubeadmControlPlane := kubeadmControlPlane()
	lookup := yamlutil.NewObjectLookupBuilder().Add(
		dockerCluster(),
		kubeadmControlPlane,
	).Build()

	yaml.ProcessKubeadmControlPlane(cp, lookup)
	g.Expect(cp.KubeadmControlPlane).To(Equal(kubeadmControlPlane))
	g.Expect(cp.ControlPlaneMachineTemplate).To(BeNil())
}

func TestProcessKubeadmControlPlaneWithControlPlaneAndMachineTemplate(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	cp.Cluster = capiCluster()
	kubeadmControlPlane := kubeadmControlPlane()
	mt := dockerMachineTemplate(kubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef.Name)
	lookup := yamlutil.NewObjectLookupBuilder().Add(
		dockerCluster(),
		kubeadmControlPlane,
		mt,
	).Build()

	yaml.ProcessKubeadmControlPlane(cp, lookup)
	g.Expect(cp.KubeadmControlPlane).To(Equal(kubeadmControlPlane))
	g.Expect(cp.ControlPlaneMachineTemplate).To(Equal(mt))
}

func TestProcessEtcdClusterStackedEtcd(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	cp.Cluster = capiCluster()
	lookup := yamlutil.NewObjectLookupBuilder().Add(dockerCluster()).Build()

	yaml.ProcessEtcdCluster(cp, lookup)
	g.Expect(cp.EtcdCluster).To(BeNil())
}

func TestProcessEtcdClusterNoCluster(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	cp.Cluster = capiCluster()
	cp.Cluster.Spec.ManagedExternalEtcdRef = objectReference(etcdCluster())
	lookup := yamlutil.NewObjectLookupBuilder().Add(dockerCluster()).Build()

	yaml.ProcessEtcdCluster(cp, lookup)
	g.Expect(cp.EtcdCluster).To(BeNil())
}

func TestProcessEtcdClusterNoMachineTemplate(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	cp.Cluster = capiCluster()
	etcdCluster := etcdCluster()
	cp.Cluster.Spec.ManagedExternalEtcdRef = objectReference(etcdCluster)
	lookup := yamlutil.NewObjectLookupBuilder().Add(
		dockerCluster(),
		etcdCluster,
	).Build()

	yaml.ProcessEtcdCluster(cp, lookup)
	g.Expect(cp.EtcdCluster).To(Equal(etcdCluster))
	g.Expect(cp.EtcdMachineTemplate).To(BeNil())
}

func TestProcessEtcdClusterWithClusterAndMachineTemplate(t *testing.T) {
	g := NewWithT(t)
	cp := &dockerControlPlane{}
	cp.Cluster = capiCluster()
	etcdCluster := etcdCluster()
	mt := dockerMachineTemplate("etcd-mt")
	etcdCluster.Spec.InfrastructureTemplate = *objectReference(mt)
	cp.Cluster.Spec.ManagedExternalEtcdRef = objectReference(etcdCluster)
	lookup := yamlutil.NewObjectLookupBuilder().Add(
		dockerCluster(),
		etcdCluster,
		mt,
	).Build()

	yaml.ProcessEtcdCluster(cp, lookup)
	g.Expect(cp.EtcdCluster).To(Equal(etcdCluster))
	g.Expect(cp.EtcdMachineTemplate).To(Equal(mt))
}

func capiCluster() *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: clusterv1.ClusterSpec{
			InfrastructureRef: &corev1.ObjectReference{
				Name:       "cluster",
				Namespace:  constants.EksaSystemNamespace,
				Kind:       "DockerCluster",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			},
			ControlPlaneRef: &corev1.ObjectReference{
				Name:       "cp",
				Namespace:  constants.EksaSystemNamespace,
				Kind:       "KubeadmControlPlane",
				APIVersion: "controlplane.clusterapi.k8s/v1beta1",
			},
		},
	}
}

func kubeadmControlPlane() *controlplanev1.KubeadmControlPlane {
	return &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: "controlplane.clusterapi.k8s/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cp",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: corev1.ObjectReference{
					Name:       "cp-mt",
					Namespace:  constants.EksaSystemNamespace,
					Kind:       "DockerMachineTemplate",
					APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				},
			},
		},
	}
}

func dockerCluster() *dockerv1.DockerCluster {
	return &dockerv1.DockerCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DockerCluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster",
			Namespace: constants.EksaSystemNamespace,
		},
	}
}

func dockerMachineTemplate(name string) *dockerv1.DockerMachineTemplate {
	return &dockerv1.DockerMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DockerMachineTemplate",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaSystemNamespace,
		},
	}
}

func etcdCluster() *etcdv1.EtcdadmCluster {
	return &etcdv1.EtcdadmCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EtcdCluster",
			APIVersion: "etcd.clusterapi.k8s",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd",
			Namespace: constants.EksaSystemNamespace,
		},
	}
}

func objectReference(obj client.Object) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
		APIVersion: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Name:       obj.GetName(),
		Namespace:  obj.GetNamespace(),
	}
}

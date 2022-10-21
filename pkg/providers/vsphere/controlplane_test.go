package vsphere_test

import (
	"testing"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clusterapi"
	yamlcapi "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type baseControlPlane = clusterapi.ControlPlane[*v1beta1.VSphereCluster, *v1beta1.VSphereMachineTemplate]

func TestBuildFromParsed(t *testing.T) {
	g := NewWithT(t)
	cl := capiCluster()
	vsphereCluster := vsphereCluster()
	kcp := kubeadmControlPlane()
	etcd := etcdCluster()
	cpMachineTemplate := vsphereMachineTemplate(kcp.Spec.MachineTemplate.InfrastructureRef.Name)
	etcdMachineTemplate := vsphereMachineTemplate(etcd.Spec.InfrastructureTemplate.Name)
	secret := secret()
	configMap := configMap()

	o := yamlutil.NewObjectLookupBuilder().Add(cl, vsphereCluster, kcp, cpMachineTemplate, etcd, etcdMachineTemplate, secret, configMap).Build()

	cpb := vsphere.ControlPlaneBuilder{
		BaseBuilder:  yamlcapi.NewControlPlaneBuilder[*v1beta1.VSphereCluster, *v1beta1.VSphereMachineTemplate](),
		ControlPlane: &vsphere.ControlPlane{},
	}

	err := cpb.BuildFromParsed(o)

	g.Expect(cpb.ControlPlane.Cluster).To(Equal(cl))
	g.Expect(cpb.ControlPlane.ProviderCluster).To(Equal(vsphereCluster))
	g.Expect(cpb.ControlPlane.KubeadmControlPlane).To(Equal(kcp))
	g.Expect(cpb.ControlPlane.Secrets).To(Equal([]*corev1.Secret{secret}))
	g.Expect(cpb.ControlPlane.ConfigMaps).To(Equal([]*corev1.ConfigMap{configMap}))
	g.Expect(err).To(BeNil())
}

func TestControlPlaneObjects(t *testing.T) {
	tests := []struct {
		name         string
		controlPlane *vsphere.ControlPlane
		want         []clusterapi.Object
	}{
		{
			name: "stacked etcd",
			controlPlane: &vsphere.ControlPlane{
				BaseControlPlane: baseControlPlane{
					Cluster:                     capiCluster(),
					ProviderCluster:             vsphereCluster(),
					KubeadmControlPlane:         kubeadmControlPlane(),
					ControlPlaneMachineTemplate: vsphereMachineTemplate("cp-mt"),
				},
			},
			want: []clusterapi.Object{
				capiCluster(),
				vsphereCluster(),
				kubeadmControlPlane(),
				vsphereMachineTemplate("cp-mt"),
			},
		},
		{
			name: "unstacked etcd",
			controlPlane: &vsphere.ControlPlane{
				BaseControlPlane: baseControlPlane{
					Cluster:                     capiCluster(),
					ProviderCluster:             vsphereCluster(),
					KubeadmControlPlane:         kubeadmControlPlane(),
					ControlPlaneMachineTemplate: vsphereMachineTemplate("cp-mt"),
					EtcdCluster:                 etcdCluster(),
					EtcdMachineTemplate:         vsphereMachineTemplate("etcd-mt"),
				},
				Secrets:    []*corev1.Secret{secret()},
				ConfigMaps: []*corev1.ConfigMap{configMap()},
			},
			want: []clusterapi.Object{
				capiCluster(),
				vsphereCluster(),
				kubeadmControlPlane(),
				vsphereMachineTemplate("cp-mt"),
				etcdCluster(),
				vsphereMachineTemplate("etcd-mt"),
				secret(),
				configMap(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.controlPlane.Objects()).To(ConsistOf(tt.want))
		})
	}
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
				Kind:       "VSphereCluster",
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
					Kind:       "VSphereMachineTemplate",
					APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				},
			},
		},
	}
}

func vsphereCluster() *v1beta1.VSphereCluster {
	return &v1beta1.VSphereCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VSphereCluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster",
			Namespace: constants.EksaSystemNamespace,
		},
	}
}

func vsphereMachineTemplate(name string) *v1beta1.VSphereMachineTemplate {
	return &v1beta1.VSphereMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VSphereMachineTemplate",
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
		Spec: etcdv1.EtcdadmClusterSpec{
			InfrastructureTemplate: corev1.ObjectReference{
				Name: "etcd-1",
			},
		},
	}
}

func secret() *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "eksa-system",
			Name:      "my-secret",
		},
		Data: map[string][]byte{
			"username": []byte("test"),
			"password": []byte("test"),
		},
	}
}

func configMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "eksa-system",
			Name:      "my-configmap",
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}
}

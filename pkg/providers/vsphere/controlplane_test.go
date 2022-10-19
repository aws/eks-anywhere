package vsphere_test

import (
	"context"
	"path"
	"testing"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	addons "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

const (
	testClusterConfigMainFilename = "cluster_main.yaml"
	testDataDir                   = "testdata"
)

type baseControlPlane = clusterapi.ControlPlane[*v1beta1.VSphereCluster, *v1beta1.VSphereMachineTemplate]

func TestControlPlaneObjects(t *testing.T) {
	tests := []struct {
		name         string
		controlPlane *vsphere.ControlPlane
		want         []kubernetes.Object
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
				Secrets:            []*corev1.Secret{secret()},
				ConfigMaps:         []*corev1.ConfigMap{configMap()},
				ClusterResourceSet: clusterResourceSet(),
			},
			want: []kubernetes.Object{
				capiCluster(),
				vsphereCluster(),
				kubeadmControlPlane(),
				vsphereMachineTemplate("cp-mt"),
				secret(),
				configMap(),
				clusterResourceSet(),
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
				Secrets:            []*corev1.Secret{secret()},
				ConfigMaps:         []*corev1.ConfigMap{configMap()},
				ClusterResourceSet: clusterResourceSet(),
			},
			want: []kubernetes.Object{
				capiCluster(),
				vsphereCluster(),
				kubeadmControlPlane(),
				vsphereMachineTemplate("cp-mt"),
				etcdCluster(),
				vsphereMachineTemplate("etcd-mt"),
				secret(),
				configMap(),
				clusterResourceSet(),
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

func TestControlPlaneSpecNewCluster(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	spec := givenClusterSpec(t, testClusterConfigMainFilename)

	cp, err := vsphere.ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
	g.Expect(cp.Cluster).To(Equal(capiCluster()))
	g.Expect(cp.KubeadmControlPlane).To(Equal(wantKCP()))
	g.Expect(cp.EtcdCluster).To(Equal(wantEtcdCluster()))
	g.Expect(cp.ProviderCluster).To(Equal(vsphereCluster()))
	g.Expect(cp.ControlPlaneMachineTemplate.Name).To(Equal("test-control-plane-1"))
	g.Expect(cp.EtcdMachineTemplate.Name).To(Equal("test-etcd-1"))
}

func TestControlPlaneSpecNoKubeVersion(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	spec := givenClusterSpec(t, testClusterConfigMainFilename)
	spec.Cluster.Spec.KubernetesVersion = ""

	_, err := vsphere.ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).To(MatchError(ContainSubstring("generating vsphere control plane yaml spec")))
}

func givenClusterSpec(t *testing.T, fileName string) *cluster.Spec {
	return test.NewFullClusterSpec(t, path.Join(testDataDir, fileName))
}

func kubeadmControlPlane() *controlplanev1.KubeadmControlPlane {
	return &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: corev1.ObjectReference{
					Name:       "test-control-plane-1",
					Namespace:  constants.EksaSystemNamespace,
					Kind:       "VSphereMachineTemplate",
					APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				},
			},
		},
	}
}

func capiCluster() *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "eksa-system",
			Labels: map[string]string{
				"cluster.x-k8s.io/cluster-name": "test",
			},
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: &clusterv1.ClusterNetwork{
				APIServerPort: nil,
				Services: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{"10.96.0.0/12"},
				},
				Pods: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{"192.168.0.0/16"},
				},
			},
			ControlPlaneRef: &corev1.ObjectReference{
				Kind:       "KubeadmControlPlane",
				Name:       "test",
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
			},
			ManagedExternalEtcdRef: &corev1.ObjectReference{
				Kind:       "EtcdadmCluster",
				Name:       "test-etcd",
				APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
			},
			InfrastructureRef: &corev1.ObjectReference{
				Kind:       "VSphereCluster",
				Name:       "test",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
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
			Name:      "test",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: v1beta1.VSphereClusterSpec{
			Server:     "vsphere_server",
			Thumbprint: "ABCDEFG",
			ControlPlaneEndpoint: v1beta1.APIEndpoint{
				Host: "1.2.3.4",
				Port: 6443,
			},
			IdentityRef: &v1beta1.VSphereIdentityReference{
				Kind: "Secret",
				Name: "test-vsphere-credentials",
			},
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
			Kind:       "EtcdadmCluster",
			APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-etcd",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: etcdv1.EtcdadmClusterSpec{
			InfrastructureTemplate: corev1.ObjectReference{
				Name:       "test-etcd-1",
				Kind:       "VSphereMachineTemplate",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
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

func clusterResourceSet() *addons.ClusterResourceSet {
	return &addons.ClusterResourceSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "addons.cluster.x-k8s.io/v1beta1",
			Kind:       "ClusterResourceSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "eksa-system",
			Name:      "my-crs",
		},
	}
}

func wantKCP() *controlplanev1.KubeadmControlPlane {
	var kcp *controlplanev1.KubeadmControlPlane
	b := []byte(`apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: test
  namespace: eksa-system
spec:
  machineTemplate:
    infrastructureRef:
      apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
      kind: VSphereMachineTemplate
      name: test-control-plane-1
  kubeadmConfigSpec:
    clusterConfiguration:
      imageRepository: public.ecr.aws/eks-distro/kubernetes
      etcd:
        external:
          endpoints: []
          caFile: "/etc/kubernetes/pki/etcd/ca.crt"
          certFile: "/etc/kubernetes/pki/apiserver-etcd-client.crt"
          keyFile: "/etc/kubernetes/pki/apiserver-etcd-client.key"
      dns:
        imageRepository: public.ecr.aws/eks-distro/coredns
        imageTag: v1.8.0-eks-1-19-4
      apiServer:
        extraArgs:
          cloud-provider: external
          audit-policy-file: /etc/kubernetes/audit-policy.yaml
          audit-log-path: /var/log/kubernetes/api-audit.log
          audit-log-maxage: "30"
          audit-log-maxbackup: "10"
          audit-log-maxsize: "512"
          profiling: "false"
          tls-cipher-suites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
        extraVolumes:
        - hostPath: /etc/kubernetes/audit-policy.yaml
          mountPath: /etc/kubernetes/audit-policy.yaml
          name: audit-policy
          pathType: File
          readOnly: true
        - hostPath: /var/log/kubernetes
          mountPath: /var/log/kubernetes
          name: audit-log-dir
          pathType: DirectoryOrCreate
          readOnly: false
      controllerManager:
        extraArgs:
          cloud-provider: external
          profiling: "false"
          tls-cipher-suites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
      scheduler:
        extraArgs:
          profiling: "false"
          tls-cipher-suites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
    files:
    - content: |
        apiVersion: v1
        kind: Pod
        metadata:
          creationTimestamp: null
          name: kube-vip
          namespace: kube-system
        spec:
          containers:
          - args:
            - manager
            env:
            - name: vip_arp
              value: "true"
            - name: port
              value: "6443"
            - name: vip_cidr
              value: "32"
            - name: cp_enable
              value: "true"
            - name: cp_namespace
              value: kube-system
            - name: vip_ddns
              value: "false"
            - name: vip_leaderelection
              value: "true"
            - name: vip_leaseduration
              value: "15"
            - name: vip_renewdeadline
              value: "10"
            - name: vip_retryperiod
              value: "2"
            - name: address
              value: 1.2.3.4
            image: public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.2-2093eaeda5a4567f0e516d652e0b25b1d7abc774
            imagePullPolicy: IfNotPresent
            name: kube-vip
            resources: {}
            securityContext:
              capabilities:
                add:
                - NET_ADMIN
                - NET_RAW
            volumeMounts:
            - mountPath: /etc/kubernetes/admin.conf
              name: kubeconfig
          hostNetwork: true
          volumes:
          - hostPath:
              path: /etc/kubernetes/admin.conf
            name: kubeconfig
        status: {}
      owner: root:root
      path: /etc/kubernetes/manifests/kube-vip.yaml
    - content: |
        apiVersion: audit.k8s.io/v1beta1
        kind: Policy
        rules:
        # Log aws-auth configmap changes
        - level: RequestResponse
          namespaces: ["kube-system"]
          verbs: ["update", "patch", "delete"]
          resources:
          - group: "" # core
            resources: ["configmaps"]
            resourceNames: ["aws-auth"]
          omitStages:
          - "RequestReceived"
        # The following requests were manually identified as high-volume and low-risk,
        # so drop them.
        - level: None
          users: ["system:kube-proxy"]
          verbs: ["watch"]
          resources:
          - group: "" # core
            resources: ["endpoints", "services", "services/status"]
        - level: None
          users: ["kubelet"] # legacy kubelet identity
          verbs: ["get"]
          resources:
          - group: "" # core
            resources: ["nodes", "nodes/status"]
        - level: None
          userGroups: ["system:nodes"]
          verbs: ["get"]
          resources:
          - group: "" # core
            resources: ["nodes", "nodes/status"]
        - level: None
          users:
          - system:kube-controller-manager
          - system:kube-scheduler
          - system:serviceaccount:kube-system:endpoint-controller
          verbs: ["get", "update"]
          namespaces: ["kube-system"]
          resources:
          - group: "" # core
            resources: ["endpoints"]
        - level: None
          users: ["system:apiserver"]
          verbs: ["get"]
          resources:
          - group: "" # core
            resources: ["namespaces", "namespaces/status", "namespaces/finalize"]
        # Don't log HPA fetching metrics.
        - level: None
          users:
          - system:kube-controller-manager
          verbs: ["get", "list"]
          resources:
          - group: "metrics.k8s.io"
        # Don't log these read-only URLs.
        - level: None
          nonResourceURLs:
          - /healthz*
          - /version
          - /swagger*
        # Don't log events requests.
        - level: None
          resources:
          - group: "" # core
            resources: ["events"]
        # node and pod status calls from nodes are high-volume and can be large, don't log responses for expected updates from nodes
        - level: Request
          users: ["kubelet", "system:node-problem-detector", "system:serviceaccount:kube-system:node-problem-detector"]
          verbs: ["update","patch"]
          resources:
          - group: "" # core
            resources: ["nodes/status", "pods/status"]
          omitStages:
          - "RequestReceived"
        - level: Request
          userGroups: ["system:nodes"]
          verbs: ["update","patch"]
          resources:
          - group: "" # core
            resources: ["nodes/status", "pods/status"]
          omitStages:
          - "RequestReceived"
        # deletecollection calls can be large, don't log responses for expected namespace deletions
        - level: Request
          users: ["system:serviceaccount:kube-system:namespace-controller"]
          verbs: ["deletecollection"]
          omitStages:
          - "RequestReceived"
        # Secrets, ConfigMaps, and TokenReviews can contain sensitive & binary data,
        # so only log at the Metadata level.
        - level: Metadata
          resources:
          - group: "" # core
            resources: ["secrets", "configmaps"]
          - group: authentication.k8s.io
            resources: ["tokenreviews"]
          omitStages:
            - "RequestReceived"
        - level: Request
          resources:
          - group: ""
            resources: ["serviceaccounts/token"]
        # Get repsonses can be large; skip them.
        - level: Request
          verbs: ["get", "list", "watch"]
          resources:
          - group: "" # core
          - group: "admissionregistration.k8s.io"
          - group: "apiextensions.k8s.io"
          - group: "apiregistration.k8s.io"
          - group: "apps"
          - group: "authentication.k8s.io"
          - group: "authorization.k8s.io"
          - group: "autoscaling"
          - group: "batch"
          - group: "certificates.k8s.io"
          - group: "extensions"
          - group: "metrics.k8s.io"
          - group: "networking.k8s.io"
          - group: "policy"
          - group: "rbac.authorization.k8s.io"
          - group: "scheduling.k8s.io"
          - group: "settings.k8s.io"
          - group: "storage.k8s.io"
          omitStages:
          - "RequestReceived"
        # Default level for known APIs
        - level: RequestResponse
          resources:
          - group: "" # core
          - group: "admissionregistration.k8s.io"
          - group: "apiextensions.k8s.io"
          - group: "apiregistration.k8s.io"
          - group: "apps"
          - group: "authentication.k8s.io"
          - group: "authorization.k8s.io"
          - group: "autoscaling"
          - group: "batch"
          - group: "certificates.k8s.io"
          - group: "extensions"
          - group: "metrics.k8s.io"
          - group: "networking.k8s.io"
          - group: "policy"
          - group: "rbac.authorization.k8s.io"
          - group: "scheduling.k8s.io"
          - group: "settings.k8s.io"
          - group: "storage.k8s.io"
          omitStages:
          - "RequestReceived"
        # Default level for all other requests.
        - level: Metadata
          omitStages:
          - "RequestReceived"
      owner: root:root
      path: /etc/kubernetes/audit-policy.yaml
    initConfiguration:
      nodeRegistration:
        criSocket: /var/run/containerd/containerd.sock
        kubeletExtraArgs:
          cloud-provider: external
          read-only-port: "0"
          anonymous-auth: "false"
          tls-cipher-suites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
        name: '{{ ds.meta_data.hostname }}'
    joinConfiguration:
      nodeRegistration:
        criSocket: /var/run/containerd/containerd.sock
        kubeletExtraArgs:
          cloud-provider: external
          read-only-port: "0"
          anonymous-auth: "false"
          tls-cipher-suites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
        name: '{{ ds.meta_data.hostname }}'
    preKubeadmCommands:
    - hostname "{{ ds.meta_data.hostname }}"
    - echo "::1         ipv6-localhost ipv6-loopback" >/etc/hosts
    - echo "127.0.0.1   localhost" >>/etc/hosts
    - echo "127.0.0.1   {{ ds.meta_data.hostname }}" >>/etc/hosts
    - echo "{{ ds.meta_data.hostname }}" >/etc/hostname
    useExperimentalRetryJoin: true
    users:
    - name: capv
      sshAuthorizedKeys:
      - 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=='
      sudo: ALL=(ALL) NOPASSWD:ALL
    format: cloud-config
  replicas: 3
  version: v1.19.8-eks-1-19-4`)
	if err := yaml.UnmarshalStrict(b, &kcp); err != nil {
		return nil
	}
	return kcp
}

func wantEtcdCluster() *etcdv1.EtcdadmCluster {
	var etcdCluster *etcdv1.EtcdadmCluster
	b := []byte(`kind: EtcdadmCluster
apiVersion: etcdcluster.cluster.x-k8s.io/v1beta1
metadata:
  name: test-etcd
  namespace: eksa-system
spec:
  replicas: 3
  etcdadmConfigSpec:
    etcdadmBuiltin: true
    format: cloud-config
    cloudInitConfig:
      version: 3.4.14
      installDir: "/usr/bin"
    preEtcdadmCommands:
      - hostname "{{ ds.meta_data.hostname }}"
      - echo "::1         ipv6-localhost ipv6-loopback" >/etc/hosts
      - echo "127.0.0.1   localhost" >>/etc/hosts
      - echo "127.0.0.1   {{ ds.meta_data.hostname }}" >>/etc/hosts
      - echo "{{ ds.meta_data.hostname }}" >/etc/hostname
    cipherSuites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
    users:
      - name: capv
        sshAuthorizedKeys:
          - 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=='
        sudo: ALL=(ALL) NOPASSWD:ALL
  infrastructureTemplate:
    apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
    kind: VSphereMachineTemplate
    name: test-etcd-1`)
	if err := yaml.UnmarshalStrict(b, &etcdCluster); err != nil {
		return nil
	}
	return etcdCluster
}

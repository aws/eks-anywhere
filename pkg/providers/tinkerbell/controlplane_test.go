package tinkerbell

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	testClusterConfigFilename = "testdata/cluster_tinkerbell_awsiam.yaml"
)

func TestControlPlaneObjects(t *testing.T) {
	tests := []struct {
		name         string
		controlPlane *ControlPlane
		expected     []kubernetes.Object
	}{
		{
			name: "stacked etcd",
			controlPlane: &ControlPlane{
				BaseControlPlane: BaseControlPlane{
					Cluster:                     capiCluster(),
					ProviderCluster:             tinkerbellCluster(),
					KubeadmControlPlane:         kubeadmControlPlane(),
					ControlPlaneMachineTemplate: tinkerbellMachineTemplate("controlplane-machinetemplate"),
				},
				Secrets: secret(),
			},
			expected: []kubernetes.Object{
				capiCluster(),
				tinkerbellCluster(),
				kubeadmControlPlane(),
				tinkerbellMachineTemplate("controlplane-machinetemplate"),
				secret(),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tc.controlPlane.Objects()).To(ConsistOf(tc.expected))
		})
	}
}

func TestControlPlaneSpecNewCluster(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)

	cp, err := ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp.Cluster).To(Equal(capiCluster()))
	g.Expect(cp.KubeadmControlPlane).To(Equal(kubeadmControlPlane()))
	g.Expect(cp.ProviderCluster).To(Equal(tinkerbellCluster()))
	g.Expect(cp.ControlPlaneMachineTemplate.Name).To(Equal("test-control-plane-1"))
}

func TestControlPlaneSpecNoChangesMachineTemplates(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	originalKCP := kubeadmControlPlane()
	originalCPMachineTemplate := tinkerbellMachineTemplate("test-control-plane-1")

	expectedKCP := originalKCP.DeepCopy()
	expectedCPtemplate := originalCPMachineTemplate.DeepCopy()

	client := test.NewFakeKubeClient(
		originalKCP,
		originalCPMachineTemplate,
	)

	cp, err := ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
	g.Expect(cp.Cluster).To(Equal(capiCluster()))
	g.Expect(cp.KubeadmControlPlane).To(Equal(expectedKCP))
	g.Expect(cp.ProviderCluster).To(Equal(tinkerbellCluster()))
	g.Expect(cp.ControlPlaneMachineTemplate).To(Equal(expectedCPtemplate))
}

func TestControlPlaneSpecUpdateMachineTemplates(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	originalKubeadmControlPlane := kubeadmControlPlane()
	originalCPMachineTemplate := tinkerbellMachineTemplate("test-control-plane")
	expectedKCP := originalKubeadmControlPlane.DeepCopy()
	expectedCPTemplate := originalCPMachineTemplate.DeepCopy()

	client := test.NewFakeKubeClient(
		originalKubeadmControlPlane,
		originalCPMachineTemplate,
	)
	cpTaints := []corev1.Taint{
		{
			Key:    "foo",
			Value:  "bar",
			Effect: "PreferNoSchedule",
		},
	}
	spec.Cluster.Spec.ControlPlaneConfiguration.Taints = cpTaints

	expectedKCP.Spec.KubeadmConfigSpec.InitConfiguration.NodeRegistration.Taints = cpTaints
	expectedKCP.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.Taints = cpTaints
	expectedKCP.Spec.MachineTemplate.InfrastructureRef.Name = "test-control-plane-1"

	expectedCPTemplate.Name = "test-control-plane-1"
	expectedCPTemplate.Spec.Template.Spec.TemplateOverride = "global_timeout: 6000\nid: \"\"\nname: tink-test\ntasks:\n- actions:\n  - environment:\n      COMPRESSED: \"true\"\n      DEST_DISK: /dev/sda\n      IMG_URL: \"\"\n    image: image2disk:v1.0.0\n    name: stream-image\n    timeout: 360\n  - environment:\n      BLOCK_DEVICE: /dev/sda2\n      CHROOT: \"y\"\n      CMD_LINE: apt -y update && apt -y install openssl\n      DEFAULT_INTERPRETER: /bin/sh -c\n      FS_TYPE: ext4\n    image: cexec:v1.0.0\n    name: install-openssl\n    timeout: 90\n  - environment:\n      CONTENTS: |\n        network:\n          version: 2\n          renderer: networkd\n          ethernets:\n              eno1:\n                  dhcp4: true\n              eno2:\n                  dhcp4: true\n              eno3:\n                  dhcp4: true\n              eno4:\n                  dhcp4: true\n      DEST_DISK: /dev/sda2\n      DEST_PATH: /etc/netplan/config.yaml\n      DIRMODE: \"0755\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      MODE: \"0644\"\n      UID: \"0\"\n    image: writefile:v1.0.0\n    name: write-netplan\n    timeout: 90\n  - environment:\n      CONTENTS: |\n        datasource:\n          Ec2:\n            metadata_urls: []\n            strict_id: false\n        system_info:\n          default_user:\n            name: tink\n            groups: [wheel, adm]\n            sudo: [\"ALL=(ALL) NOPASSWD:ALL\"]\n            shell: /bin/bash\n        manage_etc_hosts: localhost\n        warnings:\n          dsid_missing_source: off\n      DEST_DISK: /dev/sda2\n      DEST_PATH: /etc/cloud/cloud.cfg.d/10_tinkerbell.cfg\n      DIRMODE: \"0700\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      MODE: \"0600\"\n    image: writefile:v1.0.0\n    name: add-tink-cloud-init-config\n    timeout: 90\n  - environment:\n      CONTENTS: |\n        datasource: Ec2\n      DEST_DISK: /dev/sda2\n      DEST_PATH: /etc/cloud/ds-identify.cfg\n      DIRMODE: \"0700\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      MODE: \"0600\"\n      UID: \"0\"\n    image: writefile:v1.0.0\n    name: add-tink-cloud-init-ds-config\n    timeout: 90\n  - environment:\n      BLOCK_DEVICE: /dev/sda2\n      FS_TYPE: ext4\n    image: kexec:v1.0.0\n    name: kexec-image\n    pid: host\n    timeout: 90\n  name: tink-test\n  volumes:\n  - /dev:/dev\n  - /dev/console:/dev/console\n  - /lib/firmware:/lib/firmware:ro\n  worker: '{{.device_1}}'\nversion: \"0.1\"\n"
	expectedCPTemplate.Spec.Template.Spec.HardwareAffinity = &tinkerbellv1.HardwareAffinity{
		Required: []tinkerbellv1.HardwareAffinityTerm{
			{
				LabelSelector: metav1.LabelSelector{MatchLabels: map[string]string{"type": "cp"}},
			},
		},
	}

	cp, err := ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
	g.Expect(cp.Cluster).To(Equal(capiCluster()))
	g.Expect(cp.KubeadmControlPlane).To(Equal(expectedKCP))
	g.Expect(cp.ProviderCluster).To(Equal(tinkerbellCluster()))
	g.Expect(cp.ControlPlaneMachineTemplate).To(Equal(expectedCPTemplate))
}

func TestControlPlaneSpecRegistryMirrorAuthentication(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	t.Setenv("REGISTRY_USERNAME", "username")
	t.Setenv("REGISTRY_PASSWORD", "password")
	spec.Cluster.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
		Authenticate: true,
	}
	cp, err := ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp.Cluster).To(Equal(capiCluster()))
	g.Expect(cp.KubeadmControlPlane).To(Equal(kcpWithRegistryCredentials()))
	g.Expect(cp.ProviderCluster).To(Equal(tinkerbellCluster()))
	g.Expect(cp.Secrets).To(Equal(secret()))
	g.Expect(cp.ControlPlaneMachineTemplate.Name).To(Equal("test-control-plane-1"))
}

func TestControlPlaneSpecRegistryMirrorInsecureSkipVerify(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	tests := []struct {
		name         string
		mirrorConfig *anywherev1.RegistryMirrorConfiguration
		files        []bootstrapv1.File
	}{
		{
			name:         "insecure skip verify",
			mirrorConfig: test.RegistryMirrorInsecureSkipVerifyEnabled(),
			files:        test.RegistryMirrorConfigFilesInsecureSkipVerify(),
		},
		{
			name:         "insecure skip verify with cacert",
			mirrorConfig: test.RegistryMirrorInsecureSkipVerifyEnabledAndCACert(),
			files:        test.RegistryMirrorConfigFilesInsecureSkipVerifyAndCACert(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec.Cluster.Spec.RegistryMirrorConfiguration = tt.mirrorConfig
			cp, err := ControlPlaneSpec(ctx, logger, client, spec)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(cp.Cluster).To(Equal(capiCluster()))
			g.Expect(cp.KubeadmControlPlane).To(Equal(kubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Spec.KubeadmConfigSpec.Files = append(kcp.Spec.KubeadmConfigSpec.Files, tt.files...)
				kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands, test.RegistryMirrorSudoPreKubeadmCommands()...)
			})))
			g.Expect(cp.ProviderCluster).To(Equal(tinkerbellCluster()))
			g.Expect(cp.Secrets).To(BeNil())
			g.Expect(cp.ControlPlaneMachineTemplate.Name).To(Equal("test-control-plane-1"))
		})
	}
}

func tinkerbellCluster() *tinkerbellv1.TinkerbellCluster {
	return &tinkerbellv1.TinkerbellCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TinkerbellCluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: tinkerbellv1.TinkerbellClusterSpec{
			ImageLookupFormat:       "--kube-v1.21.2-eks-1-21-4.raw.gz",
			ImageLookupBaseRegistry: "/",
		},
	}
}

func kubeadmControlPlane(opts ...func(*controlplanev1.KubeadmControlPlane)) *controlplanev1.KubeadmControlPlane {
	var kcp *controlplanev1.KubeadmControlPlane
	b := []byte(`apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: test
  namespace: eksa-system
spec:
  kubeadmConfigSpec:
    clusterConfiguration:
      imageRepository: public.ecr.aws/eks-distro/kubernetes
      etcd:
        local:
          imageRepository: public.ecr.aws/eks-distro/etcd-io
          imageTag: v3.4.16-eks-1-21-4
      dns:
        imageRepository: public.ecr.aws/eks-distro/coredns
        imageTag: v1.8.3-eks-1-21-4
      apiServer:
        extraArgs:
          audit-policy-file: /etc/kubernetes/audit-policy.yaml
          audit-log-path: /var/log/kubernetes/api-audit.log
          audit-log-maxage: "30"
          audit-log-maxbackup: "10"
          audit-log-maxsize: "512"
          authentication-token-webhook-config-file: /etc/kubernetes/aws-iam-authenticator/kubeconfig.yaml
          feature-gates: ServiceLoadBalancerClass=true
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
          - hostPath: /var/lib/kubeadm/aws-iam-authenticator/
            mountPath: /etc/kubernetes/aws-iam-authenticator/
            name: authconfig
            readOnly: false
          - hostPath: /var/lib/kubeadm/aws-iam-authenticator/pki/
            mountPath: /var/aws-iam-authenticator/
            name: awsiamcert
            readOnly: false
    initConfiguration:
      nodeRegistration:
        kubeletExtraArgs:
          read-only-port: 0
          provider-id: PROVIDER_ID
          tls-cipher-suites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
          anonymous-auth: false
    joinConfiguration:
      nodeRegistration:
        ignorePreflightErrors:
        - DirAvailable--etc-kubernetes-manifests
        kubeletExtraArgs:
          anonymous-auth: false
          provider-id: PROVIDER_ID
          read-only-port: 0
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
              image: public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.581
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
      - owner: root:root
        path: /etc/kubernetes/audit-policy.yaml
      - content: |
          # clusters refers to the remote service.
          clusters:
            - name: aws-iam-authenticator
              cluster:
                certificate-authority: /var/aws-iam-authenticator/cert.pem
                server: https://localhost:21362/authenticate
          # users refers to the API Server's webhook configuration
          # (we don't need to authenticate the API server).
          users:
            - name: apiserver
          # kubeconfig files require a context. Provide one for the API Server.
          current-context: webhook
          contexts:
          - name: webhook
            context:
              cluster: aws-iam-authenticator
              user: apiserver
        permissions: "0640"
        owner: root:root
        path: /var/lib/kubeadm/aws-iam-authenticator/kubeconfig.yaml
      - contentFrom:
          secret:
            name: test-aws-iam-authenticator-ca
            key: cert.pem
        permissions: "0640"
        owner: root:root
        path: /var/lib/kubeadm/aws-iam-authenticator/pki/cert.pem
      - contentFrom:
          secret:
            name: test-aws-iam-authenticator-ca
            key: key.pem
        permissions: "0640"
        owner: root:root
        path: /var/lib/kubeadm/aws-iam-authenticator/pki/key.pem
    users:
    - name: tink-user
      sshAuthorizedKeys:
      - 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=='
      sudo: ALL=(ALL) NOPASSWD:ALL
    format: cloud-config
  rolloutStrategy:
    rollingUpdate:
      maxSurge: 1
  machineTemplate:
    infrastructureRef:
      apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
      kind: TinkerbellMachineTemplate
      name: test-control-plane-1
  replicas: 1
  version: v1.21.2-eks-1-21-4`)
	if err := yaml.UnmarshalStrict(b, &kcp); err != nil {
		return nil
	}
	for _, opt := range opts {
		opt(kcp)
	}
	return kcp
}

func kcpWithRegistryCredentials() *controlplanev1.KubeadmControlPlane {
	var kcp *controlplanev1.KubeadmControlPlane
	b := []byte(`apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  creationTimestamp: null
  name: test
  namespace: eksa-system
spec:
  kubeadmConfigSpec:
    clusterConfiguration:
      apiServer:
        extraArgs:
          audit-policy-file: /etc/kubernetes/audit-policy.yaml
          audit-log-path: /var/log/kubernetes/api-audit.log
          audit-log-maxage: "30"
          audit-log-maxbackup: "10"
          audit-log-maxsize: "512"
          authentication-token-webhook-config-file: /etc/kubernetes/aws-iam-authenticator/kubeconfig.yaml
          feature-gates: ServiceLoadBalancerClass=true
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
        - hostPath: /var/lib/kubeadm/aws-iam-authenticator/
          mountPath: /etc/kubernetes/aws-iam-authenticator/
          name: authconfig
        - hostPath: /var/lib/kubeadm/aws-iam-authenticator/pki/
          mountPath: /var/aws-iam-authenticator/
          name: awsiamcert
      bottlerocketAdmin: {}
      bottlerocketBootstrap: {}
      bottlerocketControl: {}
      controllerManager: {}
      dns:
        imageRepository: public.ecr.aws/eks-distro/coredns
        imageTag: v1.8.3-eks-1-21-4
      etcd:
        local:
          imageRepository: public.ecr.aws/eks-distro/etcd-io
          imageTag: v3.4.16-eks-1-21-4
      imageRepository: public.ecr.aws/eks-distro/kubernetes
      networking: {}
      pause: {}
      proxy: {}
      registryMirror: {}
      scheduler: {}
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
            image: public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.581
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
    - owner: root:root
      path: /etc/kubernetes/audit-policy.yaml
    - content: |
        # clusters refers to the remote service.
        clusters:
          - name: aws-iam-authenticator
            cluster:
              certificate-authority: /var/aws-iam-authenticator/cert.pem
              server: https://localhost:21362/authenticate
        # users refers to the API Server's webhook configuration
        # (we don't need to authenticate the API server).
        users:
          - name: apiserver
        # kubeconfig files require a context. Provide one for the API Server.
        current-context: webhook
        contexts:
        - name: webhook
          context:
            cluster: aws-iam-authenticator
            user: apiserver
      owner: root:root
      path: /var/lib/kubeadm/aws-iam-authenticator/kubeconfig.yaml
      permissions: "0640"
    - contentFrom:
        secret:
          key: cert.pem
          name: test-aws-iam-authenticator-ca
      owner: root:root
      path: /var/lib/kubeadm/aws-iam-authenticator/pki/cert.pem
      permissions: "0640"
    - contentFrom:
        secret:
          key: key.pem
          name: test-aws-iam-authenticator-ca
      owner: root:root
      path: /var/lib/kubeadm/aws-iam-authenticator/pki/key.pem
      permissions: "0640"
    - content: |
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
          [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
            endpoint = ["https://:"]
          [plugins."io.containerd.grpc.v1.cri".registry.configs.":".auth]
            username = "username"
            password = "password"
      owner: root:root
      path: /etc/containerd/config_append.toml
    format: cloud-config
    initConfiguration:
      localAPIEndpoint: {}
      nodeRegistration:
        kubeletExtraArgs:
          anonymous-auth: "false"
          provider-id: PROVIDER_ID
          read-only-port: "0"
          tls-cipher-suites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
    joinConfiguration:
      bottlerocketAdmin: {}
      bottlerocketBootstrap: {}
      bottlerocketControl: {}
      discovery: {}
      nodeRegistration:
        ignorePreflightErrors:
        - DirAvailable--etc-kubernetes-manifests
        kubeletExtraArgs:
          anonymous-auth: "false"
          provider-id: PROVIDER_ID
          read-only-port: "0"
          tls-cipher-suites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
      pause: {}
      proxy: {}
      registryMirror: {}
    preKubeadmCommands:
    - cat /etc/containerd/config_append.toml >> /etc/containerd/config.toml
    - sudo systemctl daemon-reload
    - sudo systemctl restart containerd
    users:
    - name: tink-user
      sshAuthorizedKeys:
      - 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=='
      sudo: ALL=(ALL) NOPASSWD:ALL
  machineTemplate:
    infrastructureRef:
      apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
      kind: TinkerbellMachineTemplate
      name: test-control-plane-1
    metadata: {}
  replicas: 1
  rolloutStrategy:
    rollingUpdate:
      maxSurge: 1
  version: v1.21.2-eks-1-21-4
status:
  initialized: false
  ready: false
  readyReplicas: 0
  replicas: 0
  unavailableReplicas: 0
  updatedReplicas: 0

`)
	if err := yaml.UnmarshalStrict(b, &kcp); err != nil {
		return nil
	}
	return kcp
}

func tinkerbellMachineTemplate(name string) *tinkerbellv1.TinkerbellMachineTemplate {
	return &tinkerbellv1.TinkerbellMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TinkerbellMachineTemplate",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: tinkerbellv1.TinkerbellMachineTemplateSpec{
			Template: tinkerbellv1.TinkerbellMachineTemplateResource{
				Spec: tinkerbellv1.TinkerbellMachineSpec{
					TemplateOverride: "global_timeout: 6000\nid: \"\"\nname: tink-test\ntasks:\n- actions:\n  - environment:\n      COMPRESSED: \"true\"\n      DEST_DISK: /dev/sda\n      IMG_URL: \"\"\n    image: image2disk:v1.0.0\n    name: stream-image\n    timeout: 360\n  - environment:\n      BLOCK_DEVICE: /dev/sda2\n      CHROOT: \"y\"\n      CMD_LINE: apt -y update && apt -y install openssl\n      DEFAULT_INTERPRETER: /bin/sh -c\n      FS_TYPE: ext4\n    image: cexec:v1.0.0\n    name: install-openssl\n    timeout: 90\n  - environment:\n      CONTENTS: |\n        network:\n          version: 2\n          renderer: networkd\n          ethernets:\n              eno1:\n                  dhcp4: true\n              eno2:\n                  dhcp4: true\n              eno3:\n                  dhcp4: true\n              eno4:\n                  dhcp4: true\n      DEST_DISK: /dev/sda2\n      DEST_PATH: /etc/netplan/config.yaml\n      DIRMODE: \"0755\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      MODE: \"0644\"\n      UID: \"0\"\n    image: writefile:v1.0.0\n    name: write-netplan\n    timeout: 90\n  - environment:\n      CONTENTS: |\n        datasource:\n          Ec2:\n            metadata_urls: []\n            strict_id: false\n        system_info:\n          default_user:\n            name: tink\n            groups: [wheel, adm]\n            sudo: [\"ALL=(ALL) NOPASSWD:ALL\"]\n            shell: /bin/bash\n        manage_etc_hosts: localhost\n        warnings:\n          dsid_missing_source: off\n      DEST_DISK: /dev/sda2\n      DEST_PATH: /etc/cloud/cloud.cfg.d/10_tinkerbell.cfg\n      DIRMODE: \"0700\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      MODE: \"0600\"\n    image: writefile:v1.0.0\n    name: add-tink-cloud-init-config\n    timeout: 90\n  - environment:\n      CONTENTS: |\n        datasource: Ec2\n      DEST_DISK: /dev/sda2\n      DEST_PATH: /etc/cloud/ds-identify.cfg\n      DIRMODE: \"0700\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      MODE: \"0600\"\n      UID: \"0\"\n    image: writefile:v1.0.0\n    name: add-tink-cloud-init-ds-config\n    timeout: 90\n  - environment:\n      BLOCK_DEVICE: /dev/sda2\n      FS_TYPE: ext4\n    image: kexec:v1.0.0\n    name: kexec-image\n    pid: host\n    timeout: 90\n  name: tink-test\n  volumes:\n  - /dev:/dev\n  - /dev/console:/dev/console\n  - /lib/firmware:/lib/firmware:ro\n  worker: '{{.device_1}}'\nversion: \"0.1\"\n",
					HardwareAffinity: &tinkerbellv1.HardwareAffinity{
						Required: []tinkerbellv1.HardwareAffinityTerm{
							{
								LabelSelector: metav1.LabelSelector{MatchLabels: map[string]string{"type": "cp"}},
							},
						},
					},
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
			Namespace: constants.EksaSystemNamespace,
			Labels:    map[string]string{"cluster.x-k8s.io/cluster-name": "test"},
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
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: "1.2.3.4",
				Port: 6443,
			},
			ControlPlaneRef: &corev1.ObjectReference{
				Kind:       "KubeadmControlPlane",
				Name:       "test",
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
			},
			InfrastructureRef: &corev1.ObjectReference{
				Kind:       "TinkerbellCluster",
				Name:       "test",
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
			Name:      "registry-credentials",
			Labels: map[string]string{
				"clusterctl.cluster.x-k8s.io/move": "true",
			},
		},
		StringData: map[string]string{
			"username": "username",
			"password": "password",
		},
	}
}

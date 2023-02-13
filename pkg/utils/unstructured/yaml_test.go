package unstructured_test

import (
	"bytes"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	unstructuredutil "github.com/aws/eks-anywhere/pkg/utils/unstructured"
)

func TestYamlToClientObjects(t *testing.T) {
	tests := []struct {
		name string
		yaml []byte
		want map[string]unstructured.Unstructured
	}{
		{
			name: "two objects",
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: ns-1
spec:
  paused: true
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-2
  namespace: ns-1
spec:
  controlPlaneEndpoint:
    host: 1.1.1.1
    port: 8080`),
			want: map[string]unstructured.Unstructured{
				"cluster-1": {
					Object: map[string]interface{}{
						"apiVersion": "cluster.x-k8s.io/v1beta1",
						"kind":       "Cluster",
						"metadata": map[string]interface{}{
							"name":      "cluster-1",
							"namespace": "ns-1",
						},
						"spec": map[string]interface{}{
							"paused": true,
						},
					},
				},
				"cluster-2": {
					Object: map[string]interface{}{
						"apiVersion": "cluster.x-k8s.io/v1beta1",
						"kind":       "Cluster",
						"metadata": map[string]interface{}{
							"name":      "cluster-2",
							"namespace": "ns-1",
						},
						"spec": map[string]interface{}{
							"controlPlaneEndpoint": map[string]interface{}{
								"host": "1.1.1.1",
								"port": float64(8080),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got, err := unstructuredutil.YamlToUnstructured(tt.yaml)
			g.Expect(err).To(BeNil(), "YamlToClientObjects() returned an error")
			g.Expect(len(got)).To(Equal(len(tt.want)), "Should have got %d objects", len(tt.want))
			for _, obj := range got {
				g.Expect(obj).To(Equal(tt.want[obj.GetName()]))
			}
		})
	}
}

func TestClientObjectsToYaml(t *testing.T) {
	tests := []struct {
		name string
		want []byte
		objs []unstructured.Unstructured
	}{
		{
			name: "two objects",
			want: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: ns-1
spec:
  paused: true
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-2
  namespace: ns-1
spec:
  controlPlaneEndpoint:
    host: 1.1.1.1
    port: 8080`),
			objs: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "cluster.x-k8s.io/v1beta1",
						"kind":       "Cluster",
						"metadata": map[string]interface{}{
							"name":      "cluster-1",
							"namespace": "ns-1",
						},
						"spec": map[string]interface{}{
							"paused": true,
						},
					},
				},
				{
					Object: map[string]interface{}{
						"apiVersion": "cluster.x-k8s.io/v1beta1",
						"kind":       "Cluster",
						"metadata": map[string]interface{}{
							"name":      "cluster-2",
							"namespace": "ns-1",
						},
						"spec": map[string]interface{}{
							"controlPlaneEndpoint": map[string]interface{}{
								"host": "1.1.1.1",
								"port": float64(8080),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got, err := unstructuredutil.UnstructuredToYaml(tt.objs)
			g.Expect(err).To(BeNil(), "ClientObjectsToYaml() returned an error")
			g.Expect(len(got)).To(Equal(len(tt.want)), "Should have got yaml of length", len(tt.want))
			res := bytes.Compare(tt.want, got)
			g.Expect(res).To(Equal(0), "ClientObjectsToYaml() produced erroneous yaml")
		})
	}
}

func TestYamlStripNull(t *testing.T) {
	hardwareYaml := []byte(`apiVersion: tinkerbell.org/v1alpha1
kind: Hardware
metadata:
  creationTimestamp: ~
  labels:
    type: cp
  name: eksa-dev27
  namespace: eksa-system
spec:
  bmcRef:
    apiGroup: 
    kind: Machine
    name: bmc-eksa-dev27
  disks:
  - device: /dev/sda
  interfaces:
  - dhcp:
      arch: x86_64
      hostname: eksa-dev27
      ip:
        address: 10.80.8.38
        family: 4
        gateway: 10.80.8.1
        netmask: 255.255.252.0
      lease_time: 4294967294
      mac: 88:e9:a4:58:5c:ac
      name_servers:
      - 8.8.8.8
      - 8.8.4.4
      uefi: true
    netboot:
      allowPXE: true
      allowWorkflow: true
  metadata:
    facility:
      facility_code: onprem
      plan_slug: c2.medium.x86
    instance:
      allow_pxe: true
      always_pxe: true
      hostname: eksa-dev27
      id: 88:e9:a4:58:5c:ac
      ips:
      - address: 10.80.8.38
        family: 4
        gateway: 10.80.8.1
        netmask: 255.255.252.0
        public: true
      operating_system: {}
status: {}
---
apiVersion: bmc.tinkerbell.org/v1alpha1
kind: Machine
metadata:
  creationTimestamp: null
  name: bmc-eksa-dev27
  namespace: eksa-system
spec:
  connection:
    authSecretRef:
      name: bmc-eksa-dev27-auth
      namespace: eksa-system
    host: 10.80.12.46
    insecureTLS: true
    port: 0
status: {}
---
apiVersion: v1
data:
  password: TTZiaUhFcE0=
  username: QWRtaW5pc3RyYXRvcg==
kind: Secret
metadata:
  creationTimestamp: null
  labels:
    clusterctl.cluster.x-k8s.io/move: "true"
  name: bmc-eksa-dev27-auth
  namespace: eksa-system
type: kubernetes.io/basic-auth`)

	wantHardwareYaml := []byte(`apiVersion: tinkerbell.org/v1alpha1
kind: Hardware
metadata:
  labels:
    type: cp
  name: eksa-dev27
  namespace: eksa-system
spec:
  bmcRef:
    kind: Machine
    name: bmc-eksa-dev27
  disks:
  - device: /dev/sda
  interfaces:
  - dhcp:
      arch: x86_64
      hostname: eksa-dev27
      ip:
        address: 10.80.8.38
        family: 4
        gateway: 10.80.8.1
        netmask: 255.255.252.0
      lease_time: 4294967294
      mac: 88:e9:a4:58:5c:ac
      name_servers:
      - 8.8.8.8
      - 8.8.4.4
      uefi: true
    netboot:
      allowPXE: true
      allowWorkflow: true
  metadata:
    facility:
      facility_code: onprem
      plan_slug: c2.medium.x86
    instance:
      allow_pxe: true
      always_pxe: true
      hostname: eksa-dev27
      id: 88:e9:a4:58:5c:ac
      ips:
      - address: 10.80.8.38
        family: 4
        gateway: 10.80.8.1
        netmask: 255.255.252.0
        public: true
      operating_system: {}
status: {}
---
apiVersion: bmc.tinkerbell.org/v1alpha1
kind: Machine
metadata:
  name: bmc-eksa-dev27
  namespace: eksa-system
spec:
  connection:
    authSecretRef:
      name: bmc-eksa-dev27-auth
      namespace: eksa-system
    host: 10.80.12.46
    insecureTLS: true
    port: 0
status: {}
---
apiVersion: v1
data:
  password: TTZiaUhFcE0=
  username: QWRtaW5pc3RyYXRvcg==
kind: Secret
metadata:
  labels:
    clusterctl.cluster.x-k8s.io/move: "true"
  name: bmc-eksa-dev27-auth
  namespace: eksa-system
type: kubernetes.io/basic-auth`)

	g := NewWithT(t)
	got, err := unstructuredutil.StripNull(hardwareYaml)
	g.Expect(err).To(BeNil())
	g.Expect(got).To(Equal(wantHardwareYaml))
}

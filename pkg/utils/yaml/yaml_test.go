package yaml_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/utils/yaml"
)

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
	got, err := yaml.StripNull(hardwareYaml)
	g.Expect(err).To(BeNil())
	g.Expect(got).To(Equal(wantHardwareYaml))
}

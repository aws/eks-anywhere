package hardware_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

const hardwareManifestsYAML = `
apiVersion: tinkerbell.org/v1alpha1
kind: Hardware
metadata:
  labels:
    clusterctl.cluster.x-k8s.io/move: "true"
  name: worker1
  namespace: eksa-system
spec:
  bmcRef: bmc-worker1
  id: b14d7f5b-8903-4a4c-b38d-55889ba820ba
status: {}
---
apiVersion: tinkerbell.org/v1alpha1
kind: BMC
metadata:
  labels:
    clusterctl.cluster.x-k8s.io/move: "true"
  name: bmc-worker1
  namespace: eksa-system
spec:
  authSecretRef:
    name: bmc-worker1-auth
    namespace: eksa-system
  host: 192.168.0.10
  vendor: supermicro
status: {}
---
apiVersion: v1
data:
  password: QWRtaW4=
  username: YWRtaW4=
kind: Secret
metadata:
  labels:
    clusterctl.cluster.x-k8s.io/move: "true"
  name: bmc-worker1-auth
  namespace: eksa-system
type: kubernetes.io/basic-auth
`

func TestParseYAMLCatalogueWithData(t *testing.T) {
	g := gomega.NewWithT(t)

	buffer := bufio.NewReader(bytes.NewBufferString(hardwareManifestsYAML))
	catalogue := hardware.NewCatalogue()

	err := hardware.ParseYAMLCatalogue(catalogue, buffer)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	g.Expect(catalogue.TotalHardware()).To(gomega.Equal(1))
	g.Expect(catalogue.TotalBMCs()).To(gomega.Equal(1))
	g.Expect(catalogue.TotalSecrets()).To(gomega.Equal(1))
}

func TestParseYAMLCatalogueWithoutData(t *testing.T) {
	g := gomega.NewWithT(t)

	var buf bytes.Buffer
	buffer := bufio.NewReader(&buf)
	catalogue := hardware.NewCatalogue()

	err := hardware.ParseYAMLCatalogue(catalogue, buffer)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	g.Expect(catalogue.TotalHardware()).To(gomega.Equal(0))
	g.Expect(catalogue.TotalBMCs()).To(gomega.Equal(0))
	g.Expect(catalogue.TotalSecrets()).To(gomega.Equal(0))
}

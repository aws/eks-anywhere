package diagnostics_test

import (
	"fmt"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
	"github.com/aws/eks-anywhere/pkg/filewriter"
)

func TestFileCollectors(t *testing.T) {
	g := NewGomegaWithT(t)
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())

	w, err := filewriter.NewWriter(t.TempDir())
	g.Expect(err).To(BeNil())

	logOut, err := w.Write("test.log", []byte("test content"))
	g.Expect(err).To(BeNil())
	g.Expect(logOut).To(BeAnExistingFile())

	collectors := factory.FileCollectors([]string{logOut})
	g.Expect(collectors).To(HaveLen(1), "DefaultCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Data.Data).To(Equal("test content"))
	g.Expect(collectors[0].Data.Name).To(Equal(filepath.Base(logOut)))

	collectors = factory.FileCollectors([]string{"does-not-exist.log"})
	g.Expect(collectors).To(HaveLen(1), "DefaultCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Data.Data).To(ContainSubstring("Failed to retrieve file does-not-exist.log for collection"))
}

func TestVsphereDataCenterConfigCollectors(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &eksav1alpha1.Cluster{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec: eksav1alpha1.ClusterSpec{
				ControlPlaneConfiguration: eksav1alpha1.ControlPlaneConfiguration{
					Endpoint: &eksav1alpha1.Endpoint{
						Host: "1.1.1.1",
					},
					Taints: []v1.Taint{
						{
							Key:    "test-key",
							Value:  "test-value",
							Effect: "NoSchedule",
						},
					},
				},
				DatacenterRef: eksav1alpha1.Ref{
					Kind: eksav1alpha1.VSphereDatacenterKind,
					Name: "testRef",
				},
				ExternalEtcdConfiguration: &eksav1alpha1.ExternalEtcdConfiguration{
					Count: 3,
					MachineGroupRef: &eksav1alpha1.Ref{
						Kind: eksav1alpha1.VSphereMachineConfigKind,
						Name: "testRef",
					},
				},
			},
			Status: eksav1alpha1.ClusterStatus{},
		}
	})
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.VSphereDatacenterKind}
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())
	collectors := factory.DataCenterConfigCollectors(datacenter, spec)
	g.Expect(collectors).To(HaveLen(4), "DataCenterConfigCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Logs.Namespace).To(Equal(constants.CapvSystemNamespace))
	g.Expect(collectors[0].Logs.Name).To(Equal(fmt.Sprintf("logs/%s", constants.CapvSystemNamespace)))
	g.Expect(collectors[1].RunPod.PodSpec.Containers[0].Name).To(Equal("check-host-port"))
	g.Expect(collectors[2].RunPod.PodSpec.Containers[0].Name).To(Equal("ping-host-ip"))
	g.Expect(collectors[3].RunPod.PodSpec.Containers[0].Name).To(Equal("check-cloud-controller"))
}

func TestCloudStackDataCenterConfigCollectors(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := test.NewClusterSpec(func(s *cluster.Spec) {})
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.CloudStackDatacenterKind}
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())
	collectors := factory.DataCenterConfigCollectors(datacenter, spec)
	g.Expect(collectors).To(HaveLen(1), "DataCenterConfigCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Logs.Namespace).To(Equal(constants.CapcSystemNamespace))
	g.Expect(collectors[0].Logs.Name).To(Equal(fmt.Sprintf("logs/%s", constants.CapcSystemNamespace)))
}

func TestTinkerbellDataCenterConfigCollectors(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := test.NewClusterSpec(func(s *cluster.Spec) {})
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.TinkerbellDatacenterKind}
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())
	collectors := factory.DataCenterConfigCollectors(datacenter, spec)
	g.Expect(collectors).To(HaveLen(1), "DataCenterConfigCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Logs.Namespace).To(Equal(constants.CaptSystemNamespace))
	g.Expect(collectors[0].Logs.Name).To(Equal(fmt.Sprintf("logs/%s", constants.CaptSystemNamespace)))
}

func TestSnowCollectors(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := test.NewClusterSpec(func(s *cluster.Spec) {})
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.SnowDatacenterKind}
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())
	collectors := factory.DataCenterConfigCollectors(datacenter, spec)
	g.Expect(collectors).To(HaveLen(1), "DataCenterConfigCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Logs.Namespace).To(Equal(constants.CapasSystemNamespace))
	g.Expect(collectors[0].Logs.Name).To(Equal(fmt.Sprintf("logs/%s", constants.CapasSystemNamespace)))
}

func TestNutanixCollectors(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := test.NewClusterSpec(func(s *cluster.Spec) {})
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.NutanixDatacenterKind}
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())
	collectors := factory.DataCenterConfigCollectors(datacenter, spec)
	g.Expect(collectors).To(HaveLen(1), "DataCenterConfigCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Logs.Namespace).To(Equal(constants.CapxSystemNamespace))
	g.Expect(collectors[0].Logs.Name).To(Equal(fmt.Sprintf("logs/%s", constants.CapxSystemNamespace)))
}

func TestHostCollectors(t *testing.T) {
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())

	tests := []struct {
		name           string
		datacenterKind string
		expectNil      bool
		expectedCount  int
	}{
		{
			name:           "Tinkerbell datacenter",
			datacenterKind: eksav1alpha1.TinkerbellDatacenterKind,
			expectNil:      false,
			expectedCount:  1,
		},
		{
			name:           "VSphere datacenter",
			datacenterKind: eksav1alpha1.VSphereDatacenterKind,
			expectNil:      true,
			expectedCount:  0,
		},
		{
			name:           "Unknown datacenter",
			datacenterKind: "UnknownDatacenterKind",
			expectNil:      true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			datacenter := eksav1alpha1.Ref{Kind: tt.datacenterKind}
			collectors := factory.HostCollectors(datacenter)

			if tt.expectNil {
				g.Expect(collectors).To(BeNil(), "HostCollectors() should return nil for %s", tt.datacenterKind)
			} else {
				g.Expect(collectors).NotTo(BeNil(), "HostCollectors() should not return nil for %s", tt.datacenterKind)
				g.Expect(collectors).To(HaveLen(tt.expectedCount), "HostCollectors() mismatch between number of desired collectors and actual")

				if tt.datacenterKind == eksav1alpha1.TinkerbellDatacenterKind {
					g.Expect(collectors[0].Run.CollectorName).To(Equal("boots-logs"))
					g.Expect(collectors[0].Run.Command).To(Equal("docker"))
					g.Expect(collectors[0].Run.Args).To(Equal([]string{"logs", "boots"}))
					g.Expect(collectors[0].Run.OutputDir).To(Equal("boots-logs"))
				}
			}
		})
	}
}

func TestAuditLogCollectors(t *testing.T) {
	tests := []struct {
		name                     string
		diagnosticCollectorImage string
	}{
		{
			name:                     "audit logs happy case",
			diagnosticCollectorImage: "test-image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			factory := diagnostics.NewCollectorFactory(tt.diagnosticCollectorImage, test.NewFileReader())
			collectors := factory.AuditLogCollectors()

			g.Expect(collectors).To(HaveLen(1), "AuditLogCollectors() should return exactly one collector")

			collector := collectors[0]
			g.Expect(collector.RunDaemonSet).NotTo(BeNil(), "AuditLogCollectors() should return a RunDaemonSet collector")

			podSpec := collector.RunDaemonSet.PodSpec
			g.Expect(podSpec).NotTo(BeNil(), "PodSpec should not be nil")
			g.Expect(podSpec.Containers).To(HaveLen(1), "PodSpec should have exactly one container")
			g.Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(1), "Container should have exactly one volume mount")
			g.Expect(podSpec.Volumes).To(HaveLen(1), "PodSpec should have exactly one volume")
			g.Expect(podSpec.NodeSelector).To(HaveKeyWithValue("node-role.kubernetes.io/control-plane", ""), "NodeSelector should target control-plane nodes")
			g.Expect(podSpec.Tolerations).To(HaveLen(1), "PodSpec should have exactly one toleration")
			g.Expect(podSpec.Tolerations[0].Key).To(Equal("node-role.kubernetes.io/control-plane"), "Toleration key should be 'node-role.kubernetes.io/control-plane'")
		})
	}
}

func TestWebhookConfigCollectors(t *testing.T) {
	g := NewGomegaWithT(t)
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())

	collectors := factory.DefaultCollectors()

	// Verify webhook collectors are included in default collectors
	var webhookCollectors []*diagnostics.Collect
	for _, collector := range collectors {
		if collector.RunPod != nil {
			name := collector.RunPod.Name
			if name == "validatingwebhookconfigurations" || name == "mutatingwebhookconfigurations" {
				webhookCollectors = append(webhookCollectors, collector)
			}
		}
	}

	g.Expect(webhookCollectors).To(HaveLen(2), "DefaultCollectors() should include 2 webhook configuration collectors")

	// Verify validating webhook collector
	var validatingCollector *diagnostics.Collect
	for _, c := range webhookCollectors {
		if c.RunPod.Name == "validatingwebhookconfigurations" {
			validatingCollector = c
			break
		}
	}
	g.Expect(validatingCollector).NotTo(BeNil(), "Should have validating webhook configuration collector")
	g.Expect(validatingCollector.RunPod.CollectorName).To(Equal("validatingwebhookconfigurations"))
	g.Expect(validatingCollector.RunPod.Namespace).To(Equal(constants.EksaDiagnosticsNamespace))
	g.Expect(validatingCollector.RunPod.PodSpec.Containers).To(HaveLen(1))
	g.Expect(validatingCollector.RunPod.PodSpec.Containers[0].Command).To(Equal([]string{"kubectl"}))
	g.Expect(validatingCollector.RunPod.PodSpec.Containers[0].Args).To(Equal([]string{"get", "validatingwebhookconfigurations", "-o", "json"}))
	g.Expect(validatingCollector.RunPod.PodSpec.HostNetwork).To(BeTrue())
	g.Expect(validatingCollector.RunPod.Timeout).To(Equal("30s"))

	// Verify mutating webhook collector
	var mutatingCollector *diagnostics.Collect
	for _, c := range webhookCollectors {
		if c.RunPod.Name == "mutatingwebhookconfigurations" {
			mutatingCollector = c
			break
		}
	}
	g.Expect(mutatingCollector).NotTo(BeNil(), "Should have mutating webhook configuration collector")
	g.Expect(mutatingCollector.RunPod.CollectorName).To(Equal("mutatingwebhookconfigurations"))
	g.Expect(mutatingCollector.RunPod.Namespace).To(Equal(constants.EksaDiagnosticsNamespace))
	g.Expect(mutatingCollector.RunPod.PodSpec.Containers).To(HaveLen(1))
	g.Expect(mutatingCollector.RunPod.PodSpec.Containers[0].Command).To(Equal([]string{"kubectl"}))
	g.Expect(mutatingCollector.RunPod.PodSpec.Containers[0].Args).To(Equal([]string{"get", "mutatingwebhookconfigurations", "-o", "json"}))
	g.Expect(mutatingCollector.RunPod.PodSpec.HostNetwork).To(BeTrue())
	g.Expect(mutatingCollector.RunPod.Timeout).To(Equal("30s"))
}

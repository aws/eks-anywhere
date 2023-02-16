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
	g.Expect(collectors).To(HaveLen(11), "DataCenterConfigCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Logs.Namespace).To(Equal(constants.CapvSystemNamespace))
	g.Expect(collectors[0].Logs.Name).To(Equal(fmt.Sprintf("logs/%s", constants.CapvSystemNamespace)))
	for _, collector := range collectors[1:7] {
		g.Expect(collector.RunPod.PodSpec.Containers[0].Command).To(Equal([]string{"kubectl"}))
		g.Expect(collector.RunPod.Namespace).To(Equal("eksa-diagnostics"))
	}
	g.Expect(collectors[8].RunPod.PodSpec.Containers[0].Name).To(Equal("check-host-port"))
	g.Expect(collectors[9].RunPod.PodSpec.Containers[0].Name).To(Equal("ping-host-ip"))
	g.Expect(collectors[10].RunPod.PodSpec.Containers[0].Name).To(Equal("check-cloud-controller"))
}

func TestCloudStackDataCenterConfigCollectors(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := test.NewClusterSpec(func(s *cluster.Spec) {})
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.CloudStackDatacenterKind}
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())
	collectors := factory.DataCenterConfigCollectors(datacenter, spec)
	g.Expect(collectors).To(HaveLen(10), "DataCenterConfigCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Logs.Namespace).To(Equal(constants.CapcSystemNamespace))
	g.Expect(collectors[0].Logs.Name).To(Equal(fmt.Sprintf("logs/%s", constants.CapcSystemNamespace)))
	for _, collector := range collectors[1:] {
		g.Expect([]string{"kubectl"}).To(Equal(collector.RunPod.PodSpec.Containers[0].Command))
		g.Expect("eksa-diagnostics").To(Equal(collector.RunPod.Namespace))
	}
}

func TestTinkerbellDataCenterConfigCollectors(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := test.NewClusterSpec(func(s *cluster.Spec) {})
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.TinkerbellDatacenterKind}
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())
	collectors := factory.DataCenterConfigCollectors(datacenter, spec)
	g.Expect(collectors).To(HaveLen(13), "DataCenterConfigCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Logs.Namespace).To(Equal(constants.CaptSystemNamespace))
	g.Expect(collectors[0].Logs.Name).To(Equal(fmt.Sprintf("logs/%s", constants.CaptSystemNamespace)))
	for _, collector := range collectors[1:] {
		g.Expect([]string{"kubectl"}).To(Equal(collector.RunPod.PodSpec.Containers[0].Command))
		g.Expect("eksa-diagnostics").To(Equal(collector.RunPod.Namespace))
	}
}

func TestSnowCollectors(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := test.NewClusterSpec(func(s *cluster.Spec) {})
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.SnowDatacenterKind}
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())
	collectors := factory.DataCenterConfigCollectors(datacenter, spec)
	g.Expect(collectors).To(HaveLen(6), "DataCenterConfigCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Logs.Namespace).To(Equal(constants.CapasSystemNamespace))
	g.Expect(collectors[0].Logs.Name).To(Equal(fmt.Sprintf("logs/%s", constants.CapasSystemNamespace)))
	for _, collector := range collectors[1:] {
		g.Expect([]string{"kubectl"}).To(Equal(collector.RunPod.PodSpec.Containers[0].Command))
		g.Expect("eksa-diagnostics").To(Equal(collector.RunPod.Namespace))
	}
}

func TestNutanixCollectors(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := test.NewClusterSpec(func(s *cluster.Spec) {})
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.NutanixDatacenterKind}
	factory := diagnostics.NewDefaultCollectorFactory(test.NewFileReader())
	collectors := factory.DataCenterConfigCollectors(datacenter, spec)
	g.Expect(collectors).To(HaveLen(6), "DataCenterConfigCollectors() mismatch between number of desired collectors and actual")
	g.Expect(collectors[0].Logs.Namespace).To(Equal(constants.CapxSystemNamespace))
	g.Expect(collectors[0].Logs.Name).To(Equal(fmt.Sprintf("logs/%s", constants.CapxSystemNamespace)))
	for _, collector := range collectors[1:] {
		g.Expect([]string{"kubectl"}).To(Equal(collector.RunPod.PodSpec.Containers[0].Command))
		g.Expect("eksa-diagnostics").To(Equal(collector.RunPod.Namespace))
	}
}

package certificates_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/certificates"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func newFakeClientBuilder() *fake.ClientBuilder {
	scheme := runtime.NewScheme()
	_ = anywherev1.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme)
}

const (
	endpoint        = "127.0.0.1"
	invalidEndpoint = "invalid-endpoint.local"
)

func TestNewCertificateScanner(t *testing.T) {
	g := gomega.NewWithT(t)

	client := newFakeClientBuilder().Build()
	logger := testr.New(t)

	scanner := certificates.NewCertificateScanner(client, logger)

	g.Expect(scanner).ToNot(gomega.BeNil())
}

func TestScanner_CheckCertificateExpiry_Success(t *testing.T) {
	tests := []struct {
		name                 string
		cluster              *anywherev1.Cluster
		controlPlaneMachines []clusterv1.Machine
		etcdMachines         []clusterv1.Machine
	}{
		{
			name: "successful certificate check with control plane only",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
			},
			controlPlaneMachines: []clusterv1.Machine{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-control-plane-abc123",
						Namespace: constants.EksaSystemNamespace,
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":  "test-cluster",
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Status: clusterv1.MachineStatus{
						Addresses: []clusterv1.MachineAddress{
							{
								Type:    clusterv1.MachineExternalIP,
								Address: endpoint,
							},
						},
					},
				},
			},
		},
		{
			name: "cluster with external etcd configuration",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
					},
					ExternalEtcdConfiguration: &anywherev1.ExternalEtcdConfiguration{
						Count: 3,
					},
				},
			},
			controlPlaneMachines: []clusterv1.Machine{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-control-plane-abc123",
						Namespace: constants.EksaSystemNamespace,
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":  "test-cluster",
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Status: clusterv1.MachineStatus{
						Addresses: []clusterv1.MachineAddress{
							{
								Type:    clusterv1.MachineExternalIP,
								Address: endpoint,
							},
						},
					},
				},
			},
			etcdMachines: []clusterv1.Machine{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-etcd-xyz789",
						Namespace: constants.EksaSystemNamespace,
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name": "test-cluster",
							"cluster.x-k8s.io/etcd-cluster": "test-cluster-etcd",
						},
					},
					Status: clusterv1.MachineStatus{
						Addresses: []clusterv1.MachineAddress{
							{
								Type:    clusterv1.MachineExternalIP,
								Address: endpoint,
							},
						},
					},
				},
			},
		},
		{
			name: "no control plane machines - fallback to endpoint",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
						Endpoint: &anywherev1.Endpoint{
							Host: "control-plane.example.com",
						},
					},
				},
			},
			controlPlaneMachines: []clusterv1.Machine{},
		},
		{
			name: "machine without external IP address",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
					},
				},
			},
			controlPlaneMachines: []clusterv1.Machine{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-control-plane-abc123",
						Namespace: constants.EksaSystemNamespace,
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":  "test-cluster",
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Status: clusterv1.MachineStatus{
						Addresses: []clusterv1.MachineAddress{
							{
								Type: clusterv1.MachineInternalIP,
							},
						},
					},
				},
			},
		},
		{
			name: "cluster with nil endpoint",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count:    1,
						Endpoint: nil,
					},
				},
			},
			controlPlaneMachines: []clusterv1.Machine{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			ctx := context.Background()

			objs := []runtime.Object{tt.cluster}
			for _, machine := range tt.controlPlaneMachines {
				objs = append(objs, machine.DeepCopy())
			}
			for _, machine := range tt.etcdMachines {
				objs = append(objs, machine.DeepCopy())
			}

			client := newFakeClientBuilder().WithRuntimeObjects(objs...).Build()
			logger := testr.New(t)
			scanner := certificates.NewCertificateScanner(client, logger)

			_, err := scanner.CheckCertificateExpiry(ctx, tt.cluster)
			g.Expect(err).ToNot(gomega.HaveOccurred())
		})
	}
}

func TestScanner_UpdateClusterCertificateStatus_Success(t *testing.T) {
	tests := []struct {
		name     string
		cluster  *anywherev1.Cluster
		machines []clusterv1.Machine
	}{
		{
			name: "successfully updates cluster certificate status",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
					},
				},
			},
			machines: []clusterv1.Machine{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-control-plane-abc123",
						Namespace: constants.EksaSystemNamespace,
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":  "test-cluster",
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Status: clusterv1.MachineStatus{
						Addresses: []clusterv1.MachineAddress{
							{
								Type:    clusterv1.MachineExternalIP,
								Address: endpoint,
							},
						},
					},
				},
			},
		},
		{
			name: "handles cluster with no machines gracefully",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
						Endpoint: &anywherev1.Endpoint{
							Host: "control-plane.example.com",
						},
					},
				},
			},
			machines: []clusterv1.Machine{},
		},
		{
			name: "handles cluster with empty endpoint host",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
						Endpoint: &anywherev1.Endpoint{
							Host: "",
						},
					},
				},
			},
			machines: []clusterv1.Machine{},
		},
		{
			name: "UpdateClusterCertificateStatus no error", // UpdateClusterCertificateStatus will never return an error even if there is any error.
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
						Endpoint: &anywherev1.Endpoint{
							Host: invalidEndpoint,
						},
					},
				},
			},
			machines: []clusterv1.Machine{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			ctx := context.Background()

			objs := []runtime.Object{tt.cluster}
			for _, machine := range tt.machines {
				objs = append(objs, machine.DeepCopy())
			}

			client := newFakeClientBuilder().WithRuntimeObjects(objs...).Build()
			logger := testr.New(t)
			scanner := certificates.NewCertificateScanner(client, logger)

			err := scanner.UpdateClusterCertificateStatus(ctx, tt.cluster)
			g.Expect(err).ToNot(gomega.HaveOccurred())
		})
	}
}

// MockClient that fails on List operations.
type MockClient struct {
	client.Client
}

func (m *MockClient) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return fmt.Errorf("simulated client error during list operation")
}

func TestScanner_CheckCertificateExpiry_Error(t *testing.T) {
	tests := []struct {
		name          string
		cluster       *anywherev1.Cluster
		expectedError string
	}{
		{
			name: "list control plane machines error",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
					},
				},
			},
			expectedError: "listing control plane machines",
		},
		{
			name: "list etcd machines error",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
					},
					ExternalEtcdConfiguration: &anywherev1.ExternalEtcdConfiguration{
						Count: 3,
					},
				},
			},
			expectedError: "listing etcd machines",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			ctx := context.Background()

			baseClient := newFakeClientBuilder().WithRuntimeObjects(tt.cluster).Build()
			failingClient := &MockClient{Client: baseClient}

			logger := testr.New(t)
			scanner := certificates.NewCertificateScanner(failingClient, logger)

			_, err := scanner.CheckCertificateExpiry(ctx, tt.cluster)

			g.Expect(err).To(gomega.HaveOccurred())
			g.Expect(err.Error()).To(gomega.ContainSubstring(tt.expectedError))
		})
	}
}

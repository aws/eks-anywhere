package cli_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cli"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/defaulting"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestNewCreateClusterDefaulter(t *testing.T) {
	g := NewWithT(t)

	skipIPCheck := cluster.NewControlPlaneIPCheckAnnotationDefaulter(false)

	r := defaulting.NewRunner[*cluster.Spec]()
	r.Register(
		skipIPCheck.ControlPlaneIPCheckDefault,
	)

	got := cli.NewCreateClusterDefaulter(skipIPCheck)

	g.Expect(got).NotTo(BeNil())
}

func TestRunWithoutSkipIPAnnotation(t *testing.T) {
	g := NewWithT(t)

	clusterSpec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: baseCluster(),
		},
	}
	skipIPCheck := cluster.NewControlPlaneIPCheckAnnotationDefaulter(false)

	createClusterDefaulter := cli.NewCreateClusterDefaulter(skipIPCheck)
	clusterSpec, err := createClusterDefaulter.Run(context.Background(), clusterSpec)

	skipIPClusterAnnotation := clusterSpec.Config.Cluster.ControlPlaneIPCheckDisabled()

	g.Expect(err).To(BeNil())
	g.Expect(skipIPClusterAnnotation).To(BeFalse())
}

func TestRunWithSkipIPAnnotation(t *testing.T) {
	g := NewWithT(t)

	clusterSpec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: baseCluster(),
		},
	}
	skipIPCheck := cluster.NewControlPlaneIPCheckAnnotationDefaulter(true)

	createClusterDefaulter := cli.NewCreateClusterDefaulter(skipIPCheck)
	clusterSpec, err := createClusterDefaulter.Run(context.Background(), clusterSpec)

	skipIPClusterAnnotation := clusterSpec.Config.Cluster.ControlPlaneIPCheckDisabled()

	g.Expect(err).To(BeNil())
	g.Expect(skipIPClusterAnnotation).To(BeTrue())
}

type clusterOpt func(c *anywherev1.Cluster)

func baseCluster(opts ...clusterOpt) *anywherev1.Cluster {
	c := &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "mgmt",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube121,
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Count: 3,
				Endpoint: &anywherev1.Endpoint{
					Host: "1.1.1.1",
				},
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			},
			BundlesRef: &anywherev1.BundlesRef{
				Name:       "bundles-1",
				Namespace:  constants.EksaSystemNamespace,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{{
				Name:  "md-0",
				Count: ptr.Int(1),
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			}},
			ClusterNetwork: anywherev1.ClusterNetwork{
				CNIConfig: &anywherev1.CNIConfig{Cilium: &anywherev1.CiliumConfig{}},
				Pods: anywherev1.Pods{
					CidrBlocks: []string{"192.168.0.0/16"},
				},
				Services: anywherev1.Services{
					CidrBlocks: []string{"10.96.0.0/12"},
				},
			},
			DatacenterRef: anywherev1.Ref{
				Kind: anywherev1.VSphereDatacenterKind,
				Name: "eksa-unit-test",
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

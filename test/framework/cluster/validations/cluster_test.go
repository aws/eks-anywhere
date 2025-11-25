package validations_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
	"github.com/aws/eks-anywhere/test/framework/cluster/validations"
)

const (
	clusterName      = "test-cluster"
	clusterNamespace = "test-namespace"
)

type clusterValidationTest struct {
	t testing.TB
	*WithT
	config          clusterf.StateValidationConfig
	clusterSpec     *cluster.Spec
	eksaSupportObjs []client.Object
}

func newStateValidatorTest(t testing.TB, clusterSpec *cluster.Spec) *clusterValidationTest {
	tt := &clusterValidationTest{
		t:           t,
		WithT:       NewWithT(t),
		clusterSpec: clusterSpec,
		config: clusterf.StateValidationConfig{
			ClusterClient:           fake.NewClientBuilder().Build(),
			ManagementClusterClient: fake.NewClientBuilder().Build(),
			ClusterSpec:             clusterSpec,
		},
		eksaSupportObjs: []client.Object{
			test.Namespace(clusterNamespace),
			test.Namespace(constants.EksaSystemNamespace),
			test.Namespace(constants.KubeSystemNamespace),
		},
	}
	return tt
}

func (tt *clusterValidationTest) actualObjects(excludedObjs ...client.Object) []client.Object {
	objs := tt.allObjs()
	actual := make([]client.Object, 0, len(objs)-len(excludedObjs))
	for _, obj := range objs {
		isExcluded := false
		for _, excluded := range excludedObjs {
			isExcluded = equality.Semantic.DeepEqual(obj, excluded)
			if isExcluded {
				break
			}
		}
		if !isExcluded {
			actual = append(actual, obj)
		}
	}
	return actual
}

func (tt *clusterValidationTest) createTestObjects(ctx context.Context, excludedObjects ...client.Object) {
	tt.createManagementClusterObjects(ctx, tt.actualObjects(excludedObjects...)...)
}

func (tt *clusterValidationTest) createClusterObjects(ctx context.Context, objs ...client.Object) {
	if err := createClientObjects(ctx, tt.config.ClusterClient, objs...); err != nil {
		tt.t.Fatalf("failed to create cluster objects: %v", err)
	}
}

func (tt *clusterValidationTest) createManagementClusterObjects(ctx context.Context, objs ...client.Object) {
	if err := createClientObjects(ctx, tt.config.ManagementClusterClient, objs...); err != nil {
		tt.t.Fatalf("failed to create management cluster objects: %v", err)
	}
}

func createClientObjects(ctx context.Context, clusterClient client.Client, objs ...client.Object) error {
	for _, obj := range objs {
		if err := clusterClient.Create(ctx, obj); err != nil {
			return err
		}
	}
	return nil
}

func (tt *clusterValidationTest) allObjs() []client.Object {
	childObjects := tt.clusterSpec.ChildObjects()
	objs := make([]client.Object, 0, len(tt.eksaSupportObjs)+len(childObjects)+1)
	objs = append(objs, tt.eksaSupportObjs...)
	objs = append(objs, tt.clusterSpec.Cluster)
	for _, o := range childObjects {
		objs = append(objs, o)
	}
	return objs
}

func TestValidateClusterReady(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name       string
		conditions v1beta1.Conditions
		cluster    *v1alpha1.Cluster
		wantErr    string
	}{
		{
			name: "CAPI cluster ready",
			conditions: v1beta1.Conditions{
				{
					Type:   v1beta1.ConditionType(corev1.NodeReady),
					Status: "True",
					Reason: "",
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				},
			},
			cluster: testCluster(),
			wantErr: "",
		},
		{
			name: "CAPI cluster does not ready",
			conditions: v1beta1.Conditions{
				{
					Type:   v1beta1.ConditionType(corev1.NodeReady),
					Status: "False",
					Reason: "Never ready for testing",
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				},
			},
			cluster: testCluster(),
			wantErr: fmt.Sprintf("CAPI cluster %s not ready yet.", clusterName),
		},
		{
			name:       "CAPI cluster does not exist",
			conditions: v1beta1.Conditions{},
			cluster:    testCluster(),
			wantErr:    fmt.Sprintf("cluster %s does not exist", clusterName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			spec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = tt.cluster
			})
			vt := newStateValidatorTest(t, spec)
			vt.createTestObjects(ctx)
			if len(tt.conditions) != 0 {
				capiCluster := test.CAPICluster(func(c *v1beta1.Cluster) {
					c.Name = tt.cluster.Name
					c.SetConditions(tt.conditions)
				})
				vt.createManagementClusterObjects(ctx, capiCluster)
			}
			err := validations.ValidateClusterReady(ctx, vt.config)
			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func TestValidateEKSAObjects(t *testing.T) {
	clusterDatacenter := dataCenter()
	tests := []struct {
		name         string
		spec         *cluster.Spec
		excludedObjs []client.Object
		wantErr      string
	}{
		{
			name: "EKSA objects exists",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = testCluster()
				s.Cluster.Spec.DatacenterRef = v1alpha1.Ref{
					Kind: v1alpha1.DockerDatacenterKind,
					Name: clusterDatacenter.Name,
				}
				s.DockerDatacenter = clusterDatacenter
			}),
			excludedObjs: []client.Object{},
			wantErr:      "",
		},
		{
			name: "EKSA objects missing",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = testCluster()
				s.Cluster.Spec.DatacenterRef = v1alpha1.Ref{
					Kind: v1alpha1.DockerDatacenterKind,
					Name: clusterDatacenter.Name,
				}
				s.DockerDatacenter = clusterDatacenter
			}),
			excludedObjs: []client.Object{
				clusterDatacenter,
			},
			wantErr: "dockerdatacenterconfigs.anywhere.eks.amazonaws.com \"datacenter\" not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			g := NewWithT(t)
			vt := newStateValidatorTest(t, tt.spec)
			vt.createTestObjects(ctx, tt.excludedObjs...)
			err := validations.ValidateEKSAObjects(ctx, vt.config)
			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func TestValidateControlPlanes(t *testing.T) {
	tests := []struct {
		name    string
		spec    *cluster.Spec
		nodes   []*corev1.Node
		wantErr string
	}{
		{
			name: "control plane nodes valid",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = testCluster()
				s.Cluster.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
					Count: 2,
				}
			}),
			nodes: []*corev1.Node{
				controlPlaneNode(func(node *corev1.Node) {
					node.Name = "test-node-1"
					node.Status.Conditions = []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					}
				}),
				controlPlaneNode(func(node *corev1.Node) {
					node.Name = "test-node-2"
					node.Status.Conditions = []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					}
				}),
			},
			wantErr: "",
		},
		{
			name: "control planes nodes count mismatch",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = testCluster()
				s.Cluster.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
					Count: 2,
				}
			}),
			nodes: []*corev1.Node{
				controlPlaneNode(func(node *corev1.Node) {
					node.Name = "test-node-1"
					node.Status.Conditions = []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					}
				}),
			},
			wantErr: "control plane node count does not match",
		},
		{
			name: "control planes nodes not ready",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = testCluster()
				s.Cluster.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
					Count: 2,
				}
			}),
			nodes: []*corev1.Node{
				controlPlaneNode(func(node *corev1.Node) {
					node.Name = "test-node-1"
					node.Status.Conditions = []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					}
				}),
				controlPlaneNode(func(node *corev1.Node) {
					node.Name = "test-node-2"
					node.Status.Conditions = []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionFalse,
						},
					}
				}),
			},

			wantErr: "node test-node-2 not ready yet.",
		},
		{
			name: "control plane node with taint match ",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = testCluster()
				s.Cluster.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
					Count:  2,
					Taints: []corev1.Taint{api.ControlPlaneTaint()},
				}
			}),
			nodes: []*corev1.Node{
				controlPlaneNode(func(node *corev1.Node) {
					node.Name = "test-node-1"
					node.Status.Conditions = []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					}
				}),
				controlPlaneNode(func(node *corev1.Node) {
					node.Name = "test-node-2"
					node.Status.Conditions = []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					}
				}),
			},
			wantErr: "",
		},
		{
			name: "control plane single node with taints ",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = testCluster()
				s.Cluster.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
					Count:  1,
					Taints: []corev1.Taint{api.ControlPlaneTaint()},
				}
			}),
			nodes: []*corev1.Node{
				controlPlaneNode(func(node *corev1.Node) {
					node.Name = "test-node-1"
					node.Status.Conditions = []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					}
					node.Spec.Taints = append(node.Spec.Taints, api.ControlPlaneTaint())
				}),
			},
			wantErr: "taints on control plane node test-node-1 or corresponding control plane configuration found",
		},
		{
			name: "control plane node with taint does not match ",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = testCluster()
				s.Cluster.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
					Count: 1,
					Taints: []corev1.Taint{
						{
							Key:    "key1",
							Value:  "value1",
							Effect: corev1.TaintEffectNoExecute,
						},
					},
				}
			}),
			nodes: []*corev1.Node{
				controlPlaneNode(func(node *corev1.Node) {
					node.Name = "test-node-1"
					node.Status.Conditions = []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					}
				}),
			},
			wantErr: "failed to validate controlplane node taints: taints on control plane node test-node-1 or corresponding control plane configuration found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			g := NewWithT(t)
			vt := newStateValidatorTest(t, tt.spec)
			vt.createTestObjects(ctx)
			vt.createClusterObjects(ctx, clientutil.ObjectsToClientObjects(tt.nodes)...)
			err := validations.ValidateControlPlaneNodes(ctx, vt.config)
			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func TestValidateCilium(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	tests := []struct {
		name           string
		cilumPolicy    v1alpha1.CiliumPolicyEnforcementMode
		cilumConfigMap *corev1.ConfigMap
		wantErr        string
	}{
		{
			name:        "cilium policy enforcement empty",
			cilumPolicy: "",
			cilumConfigMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.KubeSystemNamespace,
					Name:      "cilium-config",
				},
				Data: map[string]string{
					"enable-policy": "default",
				},
			},
			wantErr: "",
		},
		{
			name:        "matching cilium enforcement policy",
			cilumPolicy: v1alpha1.CiliumPolicyModeAlways,
			cilumConfigMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.KubeSystemNamespace,
					Name:      "cilium-config",
				},
				Data: map[string]string{
					"enable-policy": "always",
				},
			},
			wantErr: "",
		},
		{
			name:           "no cilium config map",
			cilumPolicy:    v1alpha1.CiliumPolicyModeAlways,
			cilumConfigMap: nil,
			wantErr:        "failed to retrieve configmap: configmaps \"cilium-config\" not found",
		},
		{
			name:        "mismatched cilium enforcement policy ",
			cilumPolicy: v1alpha1.CiliumPolicyModeNever,
			cilumConfigMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.KubeSystemNamespace,
					Name:      "cilium-config",
				},
				Data: map[string]string{
					"enable-policy": "always",
				},
			},
			wantErr: "cilium policy does not match. ConfigMap: always, YAML: never",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vt := newStateValidatorTest(t, test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = testCluster()
				s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
					CNIConfig: &v1alpha1.CNIConfig{
						Cilium: &v1alpha1.CiliumConfig{
							PolicyEnforcementMode: tt.cilumPolicy,
						},
					},
					Pods: v1alpha1.Pods{
						CidrBlocks: []string{"192.168.0.0/16"},
					},
					Services: v1alpha1.Services{
						CidrBlocks: []string{"10.96.0.0/12"},
					},
				}
			}))
			vt.createTestObjects(ctx)
			if tt.cilumConfigMap != nil {
				vt.createClusterObjects(ctx, tt.cilumConfigMap)
			}
			err := validations.ValidateCilium(ctx, vt.config)
			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func testCluster() *v1alpha1.Cluster {
	return &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: clusterNamespace,
		},
	}
}

func dataCenter() *v1alpha1.DockerDatacenterConfig {
	return &v1alpha1.DockerDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.DockerDatacenterKind,
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: clusterNamespace,
		},
	}
}

type controlPlaneNodeOpt = func(node *corev1.Node)

func controlPlaneNode(opts ...controlPlaneNodeOpt) *corev1.Node {
	n := &corev1.Node{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Node",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-node",
			Namespace: clusterNamespace,
			Labels: map[string]string{
				"node-role.kubernetes.io/control-plane": "",
			},
		},
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{
				api.ControlPlaneTaint(),
			},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

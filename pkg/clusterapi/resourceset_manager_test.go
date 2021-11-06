package clusterapi_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	addons "sigs.k8s.io/cluster-api/exp/addons/api/v1alpha3"

	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clusterapi/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

type resourceSetManagerTest struct {
	*WithT
	manager                            *clusterapi.ResourceSetManager
	client                             *mocks.MockClient
	ctx                                context.Context
	managementCluster, workloadCluster *types.Cluster
	kubeconfig                         string
	resourceSetName                    string
	namespace                          string
	resourceSet                        *addons.ClusterResourceSet
	configMapName                      string
	configMap                          *corev1.ConfigMap
	secretName                         string
	secret                             *corev1.Secret
	secretResources                    [][]byte
}

func newResourceSetManagerTest(t *testing.T) *resourceSetManagerTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	configMapName := "vsphere-csi-controller-role"
	secretName := "vsphere-csi-controller"
	kubeconfig := "kubeconfig.kubeconfig"
	return &resourceSetManagerTest{
		WithT:      NewWithT(t),
		ctx:        context.Background(),
		client:     client,
		manager:    clusterapi.NewResourceSetManager(client),
		kubeconfig: kubeconfig,
		managementCluster: &types.Cluster{
			KubeconfigFile: kubeconfig,
		},
		workloadCluster: &types.Cluster{
			KubeconfigFile: "kubeconfig",
		},
		resourceSetName: "resourceset",
		namespace:       "eksa-system",
		resourceSet: &addons.ClusterResourceSet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "addons.cluster.x-k8s.io/v1alpha3",
				Kind:       "ClusterResourceSet",
			},
			Spec: addons.ClusterResourceSetSpec{
				ClusterSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"cluster.x-k8s.io/cluster-name": "cluster-1",
					},
				},
				Strategy: "ApplyOnce",
				Resources: []addons.ResourceRef{
					{
						Kind: "Secret",
						Name: secretName,
					},
					{
						Kind: "ConfigMap",
						Name: configMapName,
					},
				},
			},
		},
		configMapName: configMapName,
		configMap: &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapName,
				Namespace: "eksa-system",
			},
			Data: map[string]string{
				"data1": `apiVersion: storage.k8s.io/v1
    kind: CSIDriver
    metadata:
      name: csi.vsphere.vmware.com
    spec:
      attachRequired: true`,
				"data2": `apiVersion: storage.k8s.io/v1
    kind: CSIDriver
    metadata:
      name: csi2.vsphere.vmware.com
    spec:
      attachRequired: false`,
			},
		},
		secretName: secretName,
		secret: &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			Data: map[string][]byte{
				"data": []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: vsphere-csi-controller
  namespace: kube-system
`),
			},
		},
		secretResources: [][]byte{
			[]byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: vsphere-csi-controller
  namespace: kube-system
`),
		},
	}
}

func TestResourceSetManagerForceUpdateSuccess(t *testing.T) {
	tt := newResourceSetManagerTest(t)

	tt.client.EXPECT().GetClusterResourceSet(tt.ctx, tt.kubeconfig, tt.resourceSetName, tt.namespace).Return(tt.resourceSet, nil)
	tt.client.EXPECT().GetConfigMap(tt.ctx, tt.kubeconfig, tt.configMapName, tt.namespace).Return(tt.configMap, nil)
	for _, o := range tt.configMap.Data {
		tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.workloadCluster, []byte(o))
	}
	tt.client.EXPECT().GetSecretFromNamespace(tt.ctx, tt.kubeconfig, tt.secretName, tt.namespace).Return(tt.secret, nil)
	for _, o := range tt.secretResources {
		tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.workloadCluster, o)
	}

	tt.Expect(tt.manager.ForceUpdate(tt.ctx, tt.resourceSetName, tt.namespace, tt.managementCluster, tt.workloadCluster)).To(Succeed())
}

func TestResourceSetManagerForceUpdateInvalidResourceType(t *testing.T) {
	tt := newResourceSetManagerTest(t)
	tt.resourceSet.Spec.Resources[0].Kind = "FakeKind"

	tt.client.EXPECT().GetClusterResourceSet(tt.ctx, tt.kubeconfig, tt.resourceSetName, tt.namespace).Return(tt.resourceSet, nil)

	tt.Expect(tt.manager.ForceUpdate(tt.ctx, tt.resourceSetName, tt.namespace, tt.managementCluster, tt.workloadCluster)).To(MatchError("invalid type [FakeKind] for resource in ClusterResourceSet"))
}

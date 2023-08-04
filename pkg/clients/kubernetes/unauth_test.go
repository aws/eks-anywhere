package kubernetes_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterapiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
)

func TestUnAuthClientGetSuccess(t *testing.T) {
	tests := []struct {
		name             string
		namespace        string
		obj              runtime.Object
		wantResourceType string
	}{
		{
			name:             "eksa cluster",
			namespace:        "eksa-system",
			obj:              &anywherev1.Cluster{},
			wantResourceType: "Cluster.v1alpha1.anywhere.eks.amazonaws.com",
		},
		{
			name:             "capi cluster",
			namespace:        "eksa-system",
			obj:              &clusterapiv1.Cluster{},
			wantResourceType: "Cluster.v1beta1.cluster.x-k8s.io",
		},
		{
			name:             "capi kubeadm controlplane",
			namespace:        "eksa-system",
			obj:              &controlplanev1.KubeadmControlPlane{},
			wantResourceType: "KubeadmControlPlane.v1beta1.controlplane.cluster.x-k8s.io",
		},
		{
			name:             "my-node",
			namespace:        "",
			obj:              &corev1.NodeList{},
			wantResourceType: "Node",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			kubectl := mocks.NewMockKubectl(ctrl)
			kubeconfig := "k.kubeconfig"

			o := &kubernetes.KubectlGetOptions{
				Name:      tt.name,
				Namespace: tt.namespace,
			}
			kubectl.EXPECT().Get(ctx, tt.wantResourceType, kubeconfig, tt.obj, o)

			c := kubernetes.NewUnAuthClient(kubectl)
			g.Expect(c.Init()).To(Succeed())

			g.Expect(c.Get(ctx, tt.name, tt.namespace, kubeconfig, tt.obj)).To(Succeed())
		})
	}
}

func TestUnAuthClientGetUnknownObjType(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())

	g.Expect(c.Get(ctx, "name", "namespace", "kubeconfig", &metav1.Status{})).Error()
}

func TestUnAuthClientDeleteSuccess(t *testing.T) {
	tests := []struct {
		name             string
		obj              client.Object
		wantResourceType string
		wantOpts         []interface{}
	}{
		{
			name: "eksa cluster",
			obj: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eksa-cluster",
					Namespace: "eksa-system",
				},
			},
			wantOpts: []interface{}{
				&kubernetes.KubectlDeleteOptions{
					Name:      "eksa-cluster",
					Namespace: "eksa-system",
				},
			},
			wantResourceType: "Cluster.v1alpha1.anywhere.eks.amazonaws.com",
		},
		{
			name: "capi cluster",
			obj: &clusterapiv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "capi-cluster",
					Namespace: "eksa-system",
				},
			},
			wantOpts: []interface{}{
				&kubernetes.KubectlDeleteOptions{
					Name:      "capi-cluster",
					Namespace: "eksa-system",
				},
			},
			wantResourceType: "Cluster.v1beta1.cluster.x-k8s.io",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			kubectl := mocks.NewMockKubectl(ctrl)
			kubeconfig := "k.kubeconfig"

			kubectl.EXPECT().Delete(ctx, tt.wantResourceType, kubeconfig, tt.wantOpts...)

			c := kubernetes.NewUnAuthClient(kubectl)
			g.Expect(c.Init()).To(Succeed())

			g.Expect(c.Delete(ctx, kubeconfig, tt.obj)).To(Succeed())
		})
	}
}

func TestUnAuthClientApplySuccess(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		obj       client.Object
	}{
		{
			name:      "eksa cluster",
			namespace: "eksa-system",
			obj:       &anywherev1.Cluster{},
		},
		{
			name:      "capi cluster",
			namespace: "eksa-system",
			obj:       &clusterapiv1.Cluster{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			kubectl := mocks.NewMockKubectl(ctrl)
			kubeconfig := "k.kubeconfig"

			kubectl.EXPECT().Apply(ctx, kubeconfig, tt.obj)

			c := kubernetes.NewUnAuthClient(kubectl)
			g.Expect(c.Init()).To(Succeed())

			g.Expect(c.Apply(ctx, kubeconfig, tt.obj)).To(Succeed())
		})
	}
}

func TestUnAuthClientDeleteUnknownObjType(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())

	g.Expect(c.Delete(ctx, "kubeconfig", &unknownType{})).Error()
}

type unknownType struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (*unknownType) DeepCopyObject() runtime.Object {
	return nil
}

func TestUnauthClientList(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)
	kubeconfig := "k.kubeconfig"
	list := &corev1.NodeList{}

	kubectl.EXPECT().Get(ctx, "Node", kubeconfig, list)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())

	g.Expect(c.List(ctx, kubeconfig, list)).To(Succeed())
}

func TestUnauthClientCreate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)
	kubeconfig := "k.kubeconfig"
	obj := &corev1.Pod{}

	kubectl.EXPECT().Create(ctx, kubeconfig, obj)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())

	g.Expect(c.Create(ctx, kubeconfig, obj)).To(Succeed())
}

func TestUnauthClientUpdate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)
	kubeconfig := "k.kubeconfig"
	obj := &corev1.Pod{}

	kubectl.EXPECT().Replace(ctx, kubeconfig, obj)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())

	g.Expect(c.Update(ctx, kubeconfig, obj)).To(Succeed())
}

func TestUnauthClientDelete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)
	kubeconfig := "k.kubeconfig"
	obj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "my-ns",
		},
	}

	opts := &kubernetes.KubectlDeleteOptions{
		Name:      "my-pod",
		Namespace: "my-ns",
	}
	kubectl.EXPECT().Delete(ctx, "Pod", kubeconfig, opts)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())

	g.Expect(c.Delete(ctx, kubeconfig, obj)).To(Succeed())
}

func TestUnauthClientDeleteAllOf(t *testing.T) {
	tests := []struct {
		name           string
		opts           []kubernetes.DeleteAllOfOption
		wantKubectlOpt *kubernetes.KubectlDeleteOptions
	}{
		{
			name:           "no options",
			wantKubectlOpt: &kubernetes.KubectlDeleteOptions{},
		},
		{
			name: "delete all in namespace",
			opts: []kubernetes.DeleteAllOfOption{
				&kubernetes.DeleteAllOfOptions{
					Namespace: "my-ns",
				},
			},
			wantKubectlOpt: &kubernetes.KubectlDeleteOptions{
				Namespace: "my-ns",
			},
		},
		{
			name: "delete all in namespace with label selector",
			opts: []kubernetes.DeleteAllOfOption{
				&kubernetes.DeleteAllOfOptions{
					Namespace: "my-ns",
					HasLabels: map[string]string{
						"label": "value",
					},
				},
			},
			wantKubectlOpt: &kubernetes.KubectlDeleteOptions{
				Namespace: "my-ns",
				HasLabels: map[string]string{
					"label": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			kubectl := mocks.NewMockKubectl(ctrl)
			kubeconfig := "k.kubeconfig"
			obj := &corev1.Pod{}

			kubectl.EXPECT().Delete(ctx, "Pod", kubeconfig, tt.wantKubectlOpt)

			c := kubernetes.NewUnAuthClient(kubectl)
			g.Expect(c.Init()).To(Succeed())

			g.Expect(c.DeleteAllOf(ctx, kubeconfig, obj, tt.opts...)).To(Succeed())
		})
	}
}

func TestUnauthClientApplyServerSide(t *testing.T) {
	fieldManager := "my-manager"
	tests := []struct {
		name           string
		opts           []kubernetes.ApplyServerSideOption
		wantKubectlOpt kubernetes.KubectlApplyOptions
	}{
		{
			name: "no options",
			wantKubectlOpt: kubernetes.KubectlApplyOptions{
				ServerSide:   true,
				FieldManager: fieldManager,
			},
		},
		{
			name: "force ownership",
			opts: []kubernetes.ApplyServerSideOption{
				&kubernetes.ApplyServerSideOptions{
					ForceOwnership: true,
				},
			},
			wantKubectlOpt: kubernetes.KubectlApplyOptions{
				ServerSide:     true,
				FieldManager:   fieldManager,
				ForceOwnership: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			kubectl := mocks.NewMockKubectl(ctrl)
			kubeconfig := "k.kubeconfig"
			obj := &corev1.Pod{}

			kubectl.EXPECT().Apply(ctx, kubeconfig, obj, tt.wantKubectlOpt)

			c := kubernetes.NewUnAuthClient(kubectl)
			g.Expect(c.Init()).To(Succeed())

			g.Expect(c.ApplyServerSide(ctx, kubeconfig, fieldManager, obj, tt.opts...)).To(Succeed())
		})
	}
}

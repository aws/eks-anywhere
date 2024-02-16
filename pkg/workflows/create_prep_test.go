package workflows_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	clientmocks "github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
	"github.com/aws/eks-anywhere/pkg/workflows"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
)

type createPrepTestSetup struct {
	t             *testing.T
	ctx           context.Context
	client        *clientmocks.MockClient
	clientFactory *mocks.MockClientFactory
}

func newCreatePrepTest(t *testing.T) *createPrepTestSetup {
	mockCtrl := gomock.NewController(t)
	client := clientmocks.NewMockClient(mockCtrl)
	clientFactory := mocks.NewMockClientFactory(mockCtrl)

	return &createPrepTestSetup{
		t:             t,
		ctx:           context.Background(),
		client:        client,
		clientFactory: clientFactory,
	}
}

func newNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func TestCreateNamespaceNotExistsSuccess(t *testing.T) {
	test := newCreatePrepTest(t)
	kubeconfig := "testpath"
	namespace := "test-ns"

	test.clientFactory.EXPECT().BuildClientFromKubeconfig(kubeconfig).Return(test.client, nil)
	test.client.EXPECT().Get(test.ctx, namespace, "", &corev1.Namespace{}).Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))
	test.client.EXPECT().Create(test.ctx, newNamespace(namespace)).Return(nil)

	err := workflows.CreateNamespaceIfNotPresent(test.ctx, namespace, kubeconfig, test.clientFactory)
	if err != nil {
		t.Fatalf("Expected nil, but got %v", err)
	}
}

func TestCreateNamespaceAlreadyExistsSuccess(t *testing.T) {
	test := newCreatePrepTest(t)
	kubeconfig := "testpath"
	namespace := "default"

	test.clientFactory.EXPECT().BuildClientFromKubeconfig(kubeconfig).Return(test.client, nil)
	test.client.EXPECT().Get(test.ctx, namespace, "", &corev1.Namespace{}).Return(nil)

	err := workflows.CreateNamespaceIfNotPresent(test.ctx, namespace, kubeconfig, test.clientFactory)
	if err != nil {
		t.Fatalf("Expected nil, but got %v", err)
	}
}

func TestCreateNamespaceBuildClientFail(t *testing.T) {
	test := newCreatePrepTest(t)
	kubeconfig := "testpath"
	namespace := "test-ns"

	test.clientFactory.EXPECT().BuildClientFromKubeconfig(kubeconfig).Return(test.client, fmt.Errorf(""))

	err := workflows.CreateNamespaceIfNotPresent(test.ctx, namespace, kubeconfig, test.clientFactory)

	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}
}

func TestCreateNamespaceGetNamespaceFail(t *testing.T) {
	test := newCreatePrepTest(t)
	kubeconfig := "testpath"
	namespace := "test-ns"

	test.clientFactory.EXPECT().BuildClientFromKubeconfig(kubeconfig).Return(test.client, nil)
	test.client.EXPECT().Get(test.ctx, namespace, "", &corev1.Namespace{}).Return(fmt.Errorf(""))

	err := workflows.CreateNamespaceIfNotPresent(test.ctx, namespace, kubeconfig, test.clientFactory)

	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}
}

func TestCreateNamespaceFail(t *testing.T) {
	test := newCreatePrepTest(t)
	kubeconfig := "testpath"
	namespace := "test-ns"

	test.clientFactory.EXPECT().BuildClientFromKubeconfig(kubeconfig).Return(test.client, nil)
	test.client.EXPECT().Get(test.ctx, namespace, "", &corev1.Namespace{}).Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))
	test.client.EXPECT().Create(test.ctx, newNamespace(namespace)).Return(fmt.Errorf(""))

	err := workflows.CreateNamespaceIfNotPresent(test.ctx, namespace, kubeconfig, test.clientFactory)

	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}
}

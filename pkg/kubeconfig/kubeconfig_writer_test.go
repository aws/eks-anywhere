package kubeconfig_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
)

type KubeconfigWriterTest struct {
	*WithT
	ctx           context.Context
	client        kubernetes.Client
	clientFactory *mocks.MockClientFactory
	mgmtCluster   *types.Cluster
}

func newKubeconfigWriter(t *testing.T) (kubeconfig.Writer, *KubeconfigWriterTest) {
	ctrl := gomock.NewController(t)
	tt := &KubeconfigWriterTest{
		WithT:         NewWithT(t),
		ctx:           context.Background(),
		clientFactory: mocks.NewMockClientFactory(ctrl),
		mgmtCluster: &types.Cluster{
			KubeconfigFile: "my-config",
		},
	}

	writer := kubeconfig.NewClusterAPIKubeconfigSecretWriter(tt.clientFactory, kubeconfig.WithTimeout(time.Microsecond), kubeconfig.WithBackoff(time.Microsecond))

	return writer, tt
}

func (k *KubeconfigWriterTest) buildClient(err error, objs ...kubernetes.Object) {
	k.client = test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(objs)...)
	k.clientFactory.EXPECT().BuildClientFromKubeconfig(k.mgmtCluster.KubeconfigFile).Return(k.client, err)
}

func TestWriteKubeconfig(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		buildErr    error
		secret      *corev1.Secret
		expectErr   bool
	}{
		{
			name:        "success",
			buildErr:    nil,
			clusterName: "test",
			secret: &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      clusterapi.ClusterKubeconfigSecretName("test"),
					Namespace: constants.EksaSystemNamespace,
				},
			},
		},
		{
			name:      "build client fail",
			buildErr:  fmt.Errorf("failed to build client"),
			secret:    &corev1.Secret{},
			expectErr: true,
		},
		{
			name:        "secret not found",
			buildErr:    nil,
			clusterName: "test",
			secret:      &corev1.Secret{},
			expectErr:   true,
		},
	}
	for _, tc := range tests {
		writer, tt := newKubeconfigWriter(t)
		buf := bytes.NewBuffer(make([]byte, tc.secret.Size()))
		tt.buildClient(tc.buildErr, tc.secret)
		err := writer.WriteKubeconfig(tt.ctx, tc.clusterName, tt.mgmtCluster.KubeconfigFile, buf)
		if !tc.expectErr {
			tt.Expect(err).To(BeNil())
		} else {
			tt.Expect(err).ToNot(BeNil())
		}
	}
}

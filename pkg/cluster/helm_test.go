package cluster_test

import (
	"context"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/cluster/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
)

type helmClientTest struct {
	*WithT
	ctx     context.Context
	h       *mocks.MockHelm
	cluster *anywherev1.Cluster
}

func newHelmClientTest(t *testing.T) *helmClientTest {
	ctrl := gomock.NewController(t)
	h := mocks.NewMockHelm(ctrl)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Namespace = constants.EksaSystemNamespace
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: c.Name,
		}
	})
	f := &helmClientTest{
		WithT:   NewGomegaWithT(t),
		ctx:     context.Background(),
		cluster: cluster,
		h:       h,
	}

	return f
}

func TestHelmClient_RegistryLoginIfNeededNoRegistryMirror(t *testing.T) {
	tt := newHelmClientTest(t)

	hc := cluster.NewHelmClient(tt.h, tt.cluster, "", "")
	err := hc.RegistryLoginIfNeeded(tt.ctx)

	tt.Expect(err).To(BeNil())
}

func TestHelmClient_RegistryLoginIfNeededAuthenticatedRegistryMirror(t *testing.T) {
	tt := newHelmClientTest(t)

	tt.cluster.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
		Authenticate: true,
		Endpoint:     "1.2.3.4",
		Port:         "65536",
	}

	username := "username"
	password := "password"
	hc := cluster.NewHelmClient(tt.h, tt.cluster, username, password)
	registry := net.JoinHostPort(tt.cluster.Spec.RegistryMirrorConfiguration.Endpoint, tt.cluster.Spec.RegistryMirrorConfiguration.Port)

	tt.h.EXPECT().RegistryLogin(tt.ctx, registry, username, password)

	err := hc.RegistryLoginIfNeeded(tt.ctx)

	tt.Expect(err).To(BeNil())
}

func TestHelmClient_Template(t *testing.T) {
	tt := newHelmClientTest(t)

	tt.cluster.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
		Authenticate: true,
		Endpoint:     "1.2.3.4",
		Port:         "65536",
	}

	username := "username"
	password := "password"
	ociURI := "oci://public.ecr.aws/account/charts"
	version := "1.1.1"
	namespace := "kube-system"
	values := map[string]string{
		"key1": "values1",
		"key2": "values2",
	}
	kubeVersion := "1.22"
	wantTemplateContent := []byte("template-content")

	hc := cluster.NewHelmClient(tt.h, tt.cluster, username, password)
	tt.h.EXPECT().Template(tt.ctx, ociURI, version, namespace, values, kubeVersion).Return(wantTemplateContent, nil)
	got, err := hc.Template(tt.ctx, ociURI, version, namespace, values, kubeVersion)

	tt.Expect(err).To(BeNil())
	tt.Expect(got).To(Equal(wantTemplateContent))
}

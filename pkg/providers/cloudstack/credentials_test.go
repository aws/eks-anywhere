package cloudstack

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

var (
	name      = "test-cluster"
	namespace = "eksa-system"
)

func TestGetCloudstackExecConfigMultipleProfile(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	dcConfig := createCloudstackDatacenterConfig()
	dcConfig.Spec.AvailabilityZones = append(dcConfig.Spec.AvailabilityZones, anywherev1.CloudStackAvailabilityZone{
		Name:           "testAz-2",
		CredentialsRef: "testCred2",
		Zone: anywherev1.CloudStackZone{
			Name: "zone1",
			Network: anywherev1.CloudStackResourceIdentifier{
				Name: "SharedNet1",
			},
		},
		Domain:                "testDomain",
		Account:               "testAccount",
		ManagementApiEndpoint: "testApiEndpoint",
	})
	secret1 := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testCred",
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string][]byte{
			decoder.APIKeyKey:    []byte("test-key1"),
			decoder.APIUrlKey:    []byte("http://1.1.1.1:8080/client/api"),
			decoder.SecretKeyKey: []byte("test-secret1"),
		},
	}
	secret2 := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testCred2",
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string][]byte{
			decoder.APIKeyKey:    []byte("test-key2"),
			decoder.APIUrlKey:    []byte("http://1.1.1.1:8081/client/api"),
			decoder.SecretKeyKey: []byte("test-secret2"),
		},
	}
	objs := []runtime.Object{dcConfig, secret1, secret2}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	expectedProfile1 := decoder.CloudStackProfileConfig{
		Name:          "testCred",
		ApiKey:        "test-key1",
		SecretKey:     "test-secret1",
		ManagementUrl: "http://1.1.1.1:8080/client/api",
	}
	expectedProfile2 := decoder.CloudStackProfileConfig{
		Name:          "testCred2",
		ApiKey:        "test-key2",
		SecretKey:     "test-secret2",
		ManagementUrl: "http://1.1.1.1:8081/client/api",
	}

	expectedExecConfig := &decoder.CloudStackExecConfig{
		Profiles: []decoder.CloudStackProfileConfig{expectedProfile1, expectedProfile2},
	}
	gotExecConfig, err := GetCloudstackExecConfig(ctx, client, dcConfig)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(len(gotExecConfig.Profiles)).To(Equal(len(expectedExecConfig.Profiles)))
	g.Expect(gotExecConfig.Profiles).To(ContainElements(expectedProfile1))
	g.Expect(gotExecConfig.Profiles).To(ContainElements(expectedProfile2))
}

func TestGetCloudstackExecConfigFail(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	dcConfig := createCloudstackDatacenterConfig()
	objs := []runtime.Object{dcConfig}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	gotExecConfig, err := GetCloudstackExecConfig(ctx, client, dcConfig)
	g.Expect(err).To(MatchError(ContainSubstring("secrets \"testCred\" not found")))
	g.Expect(gotExecConfig).To(BeNil())
}

func createCloudstackDatacenterConfig() *anywherev1.CloudStackDatacenterConfig {
	return &anywherev1.CloudStackDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.CloudStackDatacenterKind,
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: anywherev1.CloudStackDatacenterConfigSpec{
			AvailabilityZones: []anywherev1.CloudStackAvailabilityZone{
				{
					Name:           "testAz",
					CredentialsRef: "testCred",
					Zone: anywherev1.CloudStackZone{
						Name: "zone1",
						Network: anywherev1.CloudStackResourceIdentifier{
							Name: "SharedNet1",
						},
					},
					Domain:                "testDomain",
					Account:               "testAccount",
					ManagementApiEndpoint: "testApiEndpoint",
				},
			},
		},
	}
}

package controllers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

func TestCloudStackDatacenterReconcilerSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewCloudStackDatacenterReconciler(client, nil)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}

func TestCloudStackDatacenterReconcilerSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	dcConfig := createCloudstackDatacenterConfig()
	secrets := &apiv1.Secret{
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
	objs := []runtime.Object{dcConfig, secrets}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	ctrl := gomock.NewController(t)
	validatorRegistry := cloudstack.NewMockValidatorRegistry(ctrl)
	execConfig := &decoder.CloudStackExecConfig{
		Profiles: []decoder.CloudStackProfileConfig{
			{
				Name:          "testCred",
				ApiKey:        "test-key1",
				SecretKey:     "test-secret1",
				ManagementUrl: "http://1.1.1.1:8080/client/api",
			},
		},
	}
	validator := cloudstack.NewMockProviderValidator(ctrl)
	validatorRegistry.EXPECT().Get(execConfig).Return(validator, nil).Times(1)
	validator.EXPECT().ValidateCloudStackDatacenterConfig(ctx, dcConfig).Times(1)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	r := controllers.NewCloudStackDatacenterReconciler(client, validatorRegistry)

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestCloudStackDatacenterReconcilerSetDefaultSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	dcConfig := createCloudstackDatacenterConfig()
	dcConfig.Spec.AvailabilityZones = nil
	dcConfig.Spec.Zones = []anywherev1.CloudStackZone{
		{
			Id:      "",
			Name:    "",
			Network: anywherev1.CloudStackResourceIdentifier{},
		},
	}
	secrets := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "global",
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string][]byte{
			decoder.APIKeyKey:    []byte("test-key1"),
			decoder.APIUrlKey:    []byte("http://1.1.1.1:8080/client/api"),
			decoder.SecretKeyKey: []byte("test-secret1"),
		},
	}
	objs := []runtime.Object{dcConfig, secrets}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	ctrl := gomock.NewController(t)
	validatorRegistry := cloudstack.NewMockValidatorRegistry(ctrl)
	execConfig := &decoder.CloudStackExecConfig{
		Profiles: []decoder.CloudStackProfileConfig{
			{
				Name:          "global",
				ApiKey:        "test-key1",
				SecretKey:     "test-secret1",
				ManagementUrl: "http://1.1.1.1:8080/client/api",
			},
		},
	}
	validator := cloudstack.NewMockProviderValidator(ctrl)
	validatorRegistry.EXPECT().Get(execConfig).Return(validator, nil).Times(1)
	az := anywherev1.CloudStackAvailabilityZone{
		Name:           anywherev1.DefaultCloudStackAZPrefix + "-0",
		CredentialsRef: "global",
		Zone:           dcConfig.Spec.Zones[0],
	}
	dcConfig.Spec.AvailabilityZones = append(dcConfig.Spec.AvailabilityZones, az)
	dcConfig.Spec.Zones = nil
	validator.EXPECT().ValidateCloudStackDatacenterConfig(ctx, dcConfig).Times(1)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	r := controllers.NewCloudStackDatacenterReconciler(client, validatorRegistry)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	getDcConfig := &anywherev1.CloudStackDatacenterConfig{}
	err = client.Get(ctx, req.NamespacedName, getDcConfig)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(len(getDcConfig.Spec.AvailabilityZones)).ToNot(Equal(0))
	g.Expect(getDcConfig.Spec.AvailabilityZones[0].Name).To(Equal(anywherev1.DefaultCloudStackAZPrefix + "-0"))
}

func TestCloudstackDatacenterConfigReconcilerDelete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	dcConfig := createCloudstackDatacenterConfig()
	dcConfig.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	objs := []runtime.Object{dcConfig}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	ctrl := gomock.NewController(t)
	validatorRegistry := cloudstack.NewMockValidatorRegistry(ctrl)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	r := controllers.NewCloudStackDatacenterReconciler(client, validatorRegistry)

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestCloudstackDatacenterConfigGetValidatorFailure(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	dcConfig := createCloudstackDatacenterConfig()
	dcConfig.Spec.AvailabilityZones = nil
	objs := []runtime.Object{dcConfig}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	ctrl := gomock.NewController(t)
	validatorRegistry := cloudstack.NewMockValidatorRegistry(ctrl)
	execConfig := &decoder.CloudStackExecConfig{}
	errMsg := "building cmk executable: nil exec config for CloudMonkey, unable to proceed"
	validatorRegistry.EXPECT().Get(execConfig).Return(nil, errors.New(errMsg)).Times(1)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	r := controllers.NewCloudStackDatacenterReconciler(client, validatorRegistry)

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError(ContainSubstring(errMsg)))
}

func TestCloudstackDatacenterConfigGetDatacenterFailure(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	client := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()
	ctrl := gomock.NewController(t)
	validatorRegistry := cloudstack.NewMockValidatorRegistry(ctrl)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	r := controllers.NewCloudStackDatacenterReconciler(client, validatorRegistry)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError(ContainSubstring("failed getting cloudstack datacenter config")))
}

func TestCloudstackDatacenterConfigGetExecConfigFailure(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	dcConfig := createCloudstackDatacenterConfig()
	objs := []runtime.Object{dcConfig}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	r := controllers.NewCloudStackDatacenterReconciler(client, nil)

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError(ContainSubstring("secrets \"testCred\" not found")))

	gotDatacenterConfig := &anywherev1.CloudStackDatacenterConfig{}
	err = client.Get(ctx, req.NamespacedName, gotDatacenterConfig)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotDatacenterConfig.Status.SpecValid).To(BeFalse())
}

func TestCloudstackDatacenterConfigAccountNotPresentFailure(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	dcConfig := createCloudstackDatacenterConfig()
	secrets := &apiv1.Secret{
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
	objs := []runtime.Object{dcConfig, secrets}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	ctrl := gomock.NewController(t)
	validatorRegistry := cloudstack.NewMockValidatorRegistry(ctrl)
	execConfig := &decoder.CloudStackExecConfig{
		Profiles: []decoder.CloudStackProfileConfig{
			{
				Name:          "testCred",
				ApiKey:        "test-key1",
				SecretKey:     "test-secret1",
				ManagementUrl: "http://1.1.1.1:8080/client/api",
			},
		},
	}
	validator := cloudstack.NewMockProviderValidator(ctrl)
	validatorRegistry.EXPECT().Get(execConfig).Return(validator, nil).Times(1)
	validator.EXPECT().ValidateCloudStackDatacenterConfig(ctx, dcConfig).Return(errors.New("test error")).Times(1)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	r := controllers.NewCloudStackDatacenterReconciler(client, validatorRegistry)

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError(ContainSubstring("test error")))

	gotDatacenterConfig := &anywherev1.CloudStackDatacenterConfig{}
	err = client.Get(ctx, req.NamespacedName, gotDatacenterConfig)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotDatacenterConfig.Status.SpecValid).To(BeFalse())
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

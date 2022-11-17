package controllers_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/controllers/mocks"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

var (
	name      = "test-cluster"
	namespace = "eksa-system"
)

func TestSnowMachineConfigReconcilerSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewSnowMachineConfigReconciler(client, nil)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}

func TestSnowMachineConfigReconcilerSuccess(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	config := createSnowMachineConfig()
	validator := mocks.NewMockValidator(ctrl)
	validator.EXPECT().ValidateEC2ImageExistsOnDevice(ctx, config).Return(nil)
	validator.EXPECT().ValidateEC2SshKeyNameExists(ctx, config).Return(nil)

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	r := controllers.NewSnowMachineConfigReconciler(cl, validator)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	snowMachineConfig := &anywherev1.SnowMachineConfig{}
	err = cl.Get(ctx, req.NamespacedName, snowMachineConfig)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(snowMachineConfig.Status.FailureMessage).To(BeNil())
	g.Expect(snowMachineConfig.Status.SpecValid).To(BeTrue())
}

func TestSnowMachineConfigReconcilerFailureIncorrectObject(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	config := &anywherev1.SnowDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.SnowDatacenterKind,
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
	}

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	r := controllers.NewSnowMachineConfigReconciler(cl, nil)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(HaveOccurred())
}

func TestSnowMachineConfigReconcilerDelete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	config := createSnowMachineConfig()
	config.DeletionTimestamp = &metav1.Time{Time: time.Now()}

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	r := controllers.NewSnowMachineConfigReconciler(cl, nil)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestSnowMachineConfigReconcilerFailureImageExists(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	config := createSnowMachineConfig()
	validator := mocks.NewMockValidator(ctrl)
	validator.EXPECT().ValidateEC2SshKeyNameExists(ctx, config).Return(nil)
	validator.EXPECT().ValidateEC2ImageExistsOnDevice(ctx, config).Return(errors.New("test error"))

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	r := controllers.NewSnowMachineConfigReconciler(cl, validator)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(HaveOccurred())
	snowMachineConfig := &anywherev1.SnowMachineConfig{}
	err = cl.Get(ctx, req.NamespacedName, snowMachineConfig)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(snowMachineConfig.Status.FailureMessage).NotTo(BeNil())
	g.Expect(snowMachineConfig.Status.SpecValid).To(BeFalse())
}

func TestSnowMachineConfigReconcilerFailureKeyNameExists(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	config := createSnowMachineConfig()
	validator := mocks.NewMockValidator(ctrl)
	validator.EXPECT().ValidateEC2ImageExistsOnDevice(ctx, config).Return(nil)
	validator.EXPECT().ValidateEC2SshKeyNameExists(ctx, config).Return(errors.New("test error"))

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	r := controllers.NewSnowMachineConfigReconciler(cl, validator)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err := r.Reconcile(ctx, req)
	fmt.Println("test")
	fmt.Println(err.Error())
	g.Expect(err).To(HaveOccurred())
	snowMachineConfig := &anywherev1.SnowMachineConfig{}
	err = cl.Get(ctx, req.NamespacedName, snowMachineConfig)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(snowMachineConfig.Status.FailureMessage).NotTo(BeNil())
	g.Expect(snowMachineConfig.Status.SpecValid).To(BeFalse())
}

func TestSnowMachineConfigReconcilerFailureAggregate(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	config := createSnowMachineConfig()
	validator := mocks.NewMockValidator(ctrl)
	validator.EXPECT().ValidateEC2ImageExistsOnDevice(ctx, config).Return(errors.New("test error1"))
	validator.EXPECT().ValidateEC2SshKeyNameExists(ctx, config).Return(errors.New("test error2"))

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	r := controllers.NewSnowMachineConfigReconciler(cl, validator)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(HaveOccurred())

	// Now check to make sure error returned contains substring and right number of aggregate of errors
	errorPrefix := "reconciling snowmachineconfig: "
	g.Expect(err.Error()).To(ContainSubstring(errorPrefix))
	result := strings.TrimPrefix(err.Error(), errorPrefix)
	errors := strings.Split(result, ", ")
	g.Expect(len(errors)).To(BeIdenticalTo(2))

	snowMachineConfig := &anywherev1.SnowMachineConfig{}
	err = cl.Get(ctx, req.NamespacedName, snowMachineConfig)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(snowMachineConfig.Status.SpecValid).To(BeFalse())

	// Now check to make sure failure message status contains substring and right number of aggregate of errors
	g.Expect(snowMachineConfig.Status.FailureMessage).NotTo(BeNil())
	errors = strings.Split(*snowMachineConfig.Status.FailureMessage, ", ")
	g.Expect(len(errors)).To(BeIdenticalTo(2))
}

func createSnowMachineConfig() *anywherev1.SnowMachineConfig {
	return &anywherev1.SnowMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.SnowMachineConfigKind,
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		Spec: anywherev1.SnowMachineConfigSpec{
			Devices:    []string{"test-ip-1", "test-ip-2"},
			AMIID:      "test-ami",
			SshKeyName: "test-key",
		},
	}
}

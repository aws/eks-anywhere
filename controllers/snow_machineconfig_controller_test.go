package controllers_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	controllerMock "github.com/aws/eks-anywhere/controllers/mocks"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/aws"
	awsMock "github.com/aws/eks-anywhere/pkg/aws/mocks"
)

var (
	name      = "test-cluster"
	namespace = "eksa-system"
)

func TestSnowMachineConfigReconcilerSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewSnowMachineConfigReconciler(client, logf.Log, nil)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}

func TestSnowMachineConfigReconcilerSuccess(t *testing.T) {
	g := NewWithT(t)
	mockClients := make(map[string]*aws.Client)
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	config := createSnowMachineConfig()
	for _, ip := range config.Spec.Devices {
		client := awsMock.NewMockEC2Client(ctrl)
		client.EXPECT().DescribeKeyPairs(ctx, gomock.Any(), gomock.Any()).Return(&ec2.DescribeKeyPairsOutput{
			KeyPairs: []ec2Types.KeyPairInfo{{
				KeyName: &config.Spec.SshKeyName,
			}},
		}, nil)
		client.EXPECT().DescribeImages(ctx, gomock.Any()).Return(nil, nil)
		mockClients[ip] = aws.NewClientFromEC2(client)
	}

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	clientBuilder := controllerMock.NewMockClientBuilder(ctrl)
	clientBuilder.EXPECT().BuildSnowAwsClientMap(ctx).Return(mockClients, nil)

	r := controllers.NewSnowMachineConfigReconciler(cl, logf.Log, clientBuilder)

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
	ctrl := gomock.NewController(t)
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

	clientBuilder := controllerMock.NewMockClientBuilder(ctrl)

	r := controllers.NewSnowMachineConfigReconciler(cl, logf.Log, clientBuilder)

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
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	config := createSnowMachineConfig()
	config.DeletionTimestamp = &metav1.Time{Time: time.Now()}

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	clientBuilder := controllerMock.NewMockClientBuilder(ctrl)

	r := controllers.NewSnowMachineConfigReconciler(cl, logf.Log, clientBuilder)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestSnowMachineConfigReconcilerFailureBuildAwsClient(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	config := createSnowMachineConfig()

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	clientBuilder := controllerMock.NewMockClientBuilder(ctrl)
	clientBuilder.EXPECT().BuildSnowAwsClientMap(ctx).Return(nil, errors.New("test error"))

	r := controllers.NewSnowMachineConfigReconciler(cl, logf.Log, clientBuilder)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(HaveOccurred())
}

func TestSnowMachineConfigReconcilerFailureMachineDeviceIps(t *testing.T) {
	g := NewWithT(t)
	mockClients := make(map[string]*aws.Client)
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	config := createSnowMachineConfig()
	for _, ip := range config.Spec.Devices {
		client := awsMock.NewMockEC2Client(ctrl)
		client.EXPECT().DescribeKeyPairs(ctx, gomock.Any(), gomock.Any()).Return(&ec2.DescribeKeyPairsOutput{
			KeyPairs: []ec2Types.KeyPairInfo{{
				KeyName: &config.Spec.SshKeyName,
			}},
		}, nil)
		client.EXPECT().DescribeImages(ctx, gomock.Any()).Return(nil, nil)
		mockClients[ip] = aws.NewClientFromEC2(client)
	}
	config.Spec.Devices = append(config.Spec.Devices, "another-one")

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	clientBuilder := controllerMock.NewMockClientBuilder(ctrl)
	clientBuilder.EXPECT().BuildSnowAwsClientMap(ctx).Return(mockClients, nil)

	r := controllers.NewSnowMachineConfigReconciler(cl, logf.Log, clientBuilder)

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

func TestSnowMachineConfigReconcilerFailureImageExists(t *testing.T) {
	g := NewWithT(t)
	mockClients := make(map[string]*aws.Client)
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	config := createSnowMachineConfig()
	// Save only the first element for testing purposes
	config.Spec.Devices = config.Spec.Devices[0:1]
	client := awsMock.NewMockEC2Client(ctrl)
	client.EXPECT().DescribeImages(ctx, gomock.Any()).Return(nil, errors.New("test error"))
	client.EXPECT().DescribeKeyPairs(ctx, gomock.Any(), gomock.Any()).Return(&ec2.DescribeKeyPairsOutput{
		KeyPairs: []ec2Types.KeyPairInfo{{
			KeyName: &config.Spec.SshKeyName,
		}},
	}, nil)
	mockClients[config.Spec.Devices[0]] = aws.NewClientFromEC2(client)

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	clientBuilder := controllerMock.NewMockClientBuilder(ctrl)
	clientBuilder.EXPECT().BuildSnowAwsClientMap(ctx).Return(mockClients, nil)

	r := controllers.NewSnowMachineConfigReconciler(cl, logf.Log, clientBuilder)

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
	mockClients := make(map[string]*aws.Client)
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	config := createSnowMachineConfig()
	// Save only the first element for testing purposes
	config.Spec.Devices = config.Spec.Devices[0:1]
	client := awsMock.NewMockEC2Client(ctrl)
	client.EXPECT().DescribeImages(ctx, gomock.Any()).Return(nil, nil)
	client.EXPECT().DescribeKeyPairs(ctx, gomock.Any(), gomock.Any()).Return(nil, errors.New("test error"))
	mockClients[config.Spec.Devices[0]] = aws.NewClientFromEC2(client)

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	clientBuilder := controllerMock.NewMockClientBuilder(ctrl)
	clientBuilder.EXPECT().BuildSnowAwsClientMap(ctx).Return(mockClients, nil)

	r := controllers.NewSnowMachineConfigReconciler(cl, logf.Log, clientBuilder)

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
	mockClients := make(map[string]*aws.Client)
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	config := createSnowMachineConfig()
	// Save only the first element for testing purposes
	config.Spec.Devices = config.Spec.Devices[0:1]
	client := awsMock.NewMockEC2Client(ctrl)
	client.EXPECT().DescribeKeyPairs(ctx, gomock.Any(), gomock.Any()).Return(nil, errors.New("test error"))
	client.EXPECT().DescribeImages(ctx, gomock.Any()).Return(nil, errors.New("test error"))
	mockClients[config.Spec.Devices[0]] = aws.NewClientFromEC2(client)
	config.Spec.Devices = append(config.Spec.Devices, "another-one")

	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	clientBuilder := controllerMock.NewMockClientBuilder(ctrl)
	clientBuilder.EXPECT().BuildSnowAwsClientMap(ctx).Return(mockClients, nil)

	r := controllers.NewSnowMachineConfigReconciler(cl, logf.Log, clientBuilder)

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
	g.Expect(len(errors)).To(BeIdenticalTo(3))

	snowMachineConfig := &anywherev1.SnowMachineConfig{}
	err = cl.Get(ctx, req.NamespacedName, snowMachineConfig)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(snowMachineConfig.Status.SpecValid).To(BeFalse())

	// Now check to make sure failure message status contains substring and right number of aggregate of errors
	g.Expect(snowMachineConfig.Status.FailureMessage).NotTo(BeNil())
	errors = strings.Split(*snowMachineConfig.Status.FailureMessage, ", ")
	g.Expect(len(errors)).To(BeIdenticalTo(3))
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

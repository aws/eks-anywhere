package executables_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	addons "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	secretObjectType = "addons.cluster.x-k8s.io/resource-set"
	secretObjectName = "cpi-vsphere-config"
)

//go:embed testdata/nutanix/machineConfig.yaml
var nutanixMachineConfigSpec string

//go:embed testdata/nutanix/machineConfig.json
var nutanixMachineConfigSpecJSON string

//go:embed testdata/nutanix/datacenterConfig.json
var nutanixDatacenterConfigSpecJSON string

//go:embed testdata/nutanix/machineConfigs.json
var nutanixMachineConfigsJSON string

//go:embed testdata/nutanix/datacenterConfigs.json
var nutanixDatacenterConfigsJSON string

var capiClustersResourceType = fmt.Sprintf("clusters.%s", clusterv1.GroupVersion.Group)

func newKubectl(t *testing.T) (*executables.Kubectl, context.Context, *types.Cluster, *mockexecutables.MockExecutable) {
	kubeconfigFile := "c.kubeconfig"
	cluster := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
		Name:           "test-cluster",
	}

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)

	return executables.NewKubectl(executable), ctx, cluster, executable
}

type kubectlTest struct {
	t *testing.T
	*WithT
	k          *executables.Kubectl
	ctx        context.Context
	cluster    *types.Cluster
	e          *mockexecutables.MockExecutable
	namespace  string
	kubeconfig string
}

func newKubectlTest(t *testing.T) *kubectlTest {
	k, ctx, cluster, e := newKubectl(t)
	return &kubectlTest{
		t:          t,
		k:          k,
		ctx:        ctx,
		cluster:    cluster,
		e:          e,
		WithT:      NewWithT(t),
		namespace:  "namespace",
		kubeconfig: cluster.KubeconfigFile,
	}
}

func TestKubectlGetCAPIMachines(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	k, ctx, cluster, e := newKubectl(t)

	machinesResponseBuffer := bytes.Buffer{}
	machinesResponseBuffer.WriteString(test.ReadFile(t, "testdata/kubectl_machines_no_conditions.json"))

	tests := []struct {
		name          string
		buffer        bytes.Buffer
		machineLength int
		execErr       error
		expectedErr   error
	}{
		{
			name:          "GetCAPIMachines_Success",
			buffer:        machinesResponseBuffer,
			machineLength: 2,
			execErr:       nil,
			expectedErr:   nil,
		},
		{
			name:          "GetCAPIMachines_Success",
			buffer:        bytes.Buffer{},
			machineLength: 0,
			execErr:       nil,
			expectedErr:   errors.New("parsing get machines response: unexpected end of JSON input"),
		},
		{
			name:          "GetCAPIMachines_Success",
			buffer:        bytes.Buffer{},
			machineLength: 0,
			execErr:       errors.New("exec error"),
			expectedErr:   errors.New("getting machines: exec error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e.EXPECT().Execute(ctx,
				"get", "machines.cluster.x-k8s.io",
				"-o", "json",
				"--kubeconfig", cluster.KubeconfigFile,
				"--selector=cluster.x-k8s.io/cluster-name="+cluster.Name,
				"--namespace", constants.EksaSystemNamespace,
			).Return(test.buffer, test.execErr)

			machines, err := k.GetCAPIMachines(ctx, cluster, cluster.Name)
			if test.expectedErr != nil {
				g.Expect(err).To(MatchError(test.expectedErr))
			} else {
				g.Expect(err).To(Not(HaveOccurred()))
			}
			g.Expect(len(machines)).To(Equal(test.machineLength))
		})
	}
}

func TestKubectlApplyManifestSuccess(t *testing.T) {
	t.Parallel()
	spec := "specfile"

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", spec, "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.ApplyManifest(ctx, cluster.KubeconfigFile, spec); err != nil {
		t.Errorf("Kubectl.ApplyManifest() error = %v, want nil", err)
	}
}

func TestKubectlApplyManifestError(t *testing.T) {
	t.Parallel()
	spec := "specfile"

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", spec, "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.ApplyManifest(ctx, cluster.KubeconfigFile, spec); err == nil {
		t.Errorf("Kubectl.ApplyManifest() error = nil, want not nil")
	}
}

func TestKubectlApplyKubeSpecFromBytesSuccess(t *testing.T) {
	t.Parallel()
	var data []byte

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", "-", "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.ApplyKubeSpecFromBytes(ctx, cluster, data); err != nil {
		t.Errorf("Kubectl.ApplyKubeSpecFromBytes() error = %v, want nil", err)
	}
}

func TestKubectlApplyKubeSpecFromBytesError(t *testing.T) {
	t.Parallel()
	var data []byte

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", "-", "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.ApplyKubeSpecFromBytes(ctx, cluster, data); err == nil {
		t.Errorf("Kubectl.ApplyKubeSpecFromBytes() error = nil, want not nil")
	}
}

func TestKubectlDeleteManifestSuccess(t *testing.T) {
	t.Parallel()
	spec := "specfile"

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"delete", "-f", spec, "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.DeleteManifest(ctx, cluster.KubeconfigFile, spec); err != nil {
		t.Errorf("Kubectl.DeleteManifest() error = %v, want nil", err)
	}
}

func TestKubectlDeleteManifestError(t *testing.T) {
	t.Parallel()
	spec := "specfile"

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"delete", "-f", spec, "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.DeleteManifest(ctx, cluster.KubeconfigFile, spec); err == nil {
		t.Errorf("Kubectl.DeleteManifest() error = nil, want not nil")
	}
}

func TestKubectlDeleteKubeSpecFromBytesSuccess(t *testing.T) {
	t.Parallel()
	var data []byte

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"delete", "-f", "-", "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.DeleteKubeSpecFromBytes(ctx, cluster, data); err != nil {
		t.Errorf("Kubectl.DeleteKubeSpecFromBytes() error = %v, want nil", err)
	}
}

func TestKubectlDeleteSpecFromBytesError(t *testing.T) {
	t.Parallel()
	var data []byte

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"delete", "-f", "-", "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.DeleteKubeSpecFromBytes(ctx, cluster, data); err == nil {
		t.Errorf("Kubectl.DeleteKubeSpecFromBytes() error = nil, want not nil")
	}
}

func TestKubectlApplyKubeSpecFromBytesWithNamespaceSuccess(t *testing.T) {
	t.Parallel()
	var data []byte = []byte("someData")
	var namespace string

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", "-", "--namespace", namespace, "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, data, namespace); err != nil {
		t.Errorf("Kubectl.ApplyKubeSpecFromBytesWithNamespace() error = %v, want nil", err)
	}
}

func TestKubectlApplyKubeSpecFromBytesWithNamespaceSuccessWithEmptyInput(t *testing.T) {
	t.Parallel()
	var data []byte
	var namespace string

	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Any()).Times(0)
	if err := k.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, data, namespace); err != nil {
		t.Errorf("Kubectl.ApplyKubeSpecFromBytesWithNamespace() error = %v, want nil", err)
	}
}

func TestKubectlApplyKubeSpecFromBytesWithNamespaceError(t *testing.T) {
	t.Parallel()
	var data []byte = []byte("someData")
	var namespace string

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"apply", "-f", "-", "--namespace", namespace, "--kubeconfig", cluster.KubeconfigFile}
	e.EXPECT().ExecuteWithStdin(ctx, data, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, data, namespace); err == nil {
		t.Errorf("Kubectl.ApplyKubeSpecFromBytes() error = nil, want not nil")
	}
}

func TestKubectlCreateNamespaceSuccess(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"create", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.CreateNamespace(ctx, kubeconfig, namespace); err != nil {
		t.Errorf("Kubectl.CreateNamespace() error = %v, want nil", err)
	}
}

func TestKubectlCreateNamespaceError(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"create", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.CreateNamespace(ctx, kubeconfig, namespace); err == nil {
		t.Errorf("Kubectl.CreateNamespace() error = nil, want not nil")
	}
}

func TestKubectlCreateNamespaceIfNotPresentSuccessOnNamespacePresent(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParamForGetNamespace := []string{"get", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParamForGetNamespace)).Return(bytes.Buffer{}, nil)
	if err := k.CreateNamespaceIfNotPresent(ctx, kubeconfig, namespace); err != nil {
		t.Errorf("Kubectl.CreateNamespaceIfNotPresent() error = %v, want nil", err)
	}
}

func TestKubectlCreateNamespaceIfNotPresentSuccessOnNamespaceNotPresent(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParamForGetNamespace := []string{"get", "namespace", namespace, "--kubeconfig", kubeconfig}
	expectedParamForCreateNamespace := []string{"create", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParamForGetNamespace)).Return(bytes.Buffer{}, errors.New("not found"))
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParamForCreateNamespace)).Return(bytes.Buffer{}, nil)
	if err := k.CreateNamespaceIfNotPresent(ctx, kubeconfig, namespace); err != nil {
		t.Errorf("Kubectl.CreateNamespaceIfNotPresent() error = %v, want nil", err)
	}
}

func TestKubectlCreateNamespaceIfNotPresentFailureOnNamespaceCreationFailure(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParamForGetNamespace := []string{"get", "namespace", namespace, "--kubeconfig", kubeconfig}
	expectedParamForCreateNamespace := []string{"create", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParamForGetNamespace)).Return(bytes.Buffer{}, errors.New("not found"))
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParamForCreateNamespace)).Return(bytes.Buffer{}, errors.New("exception"))
	if err := k.CreateNamespaceIfNotPresent(ctx, kubeconfig, namespace); err == nil {
		t.Errorf("Kubectl.CreateNamespaceIfNotPresent() error = nil, want not nil")
	}
}

func TestKubectlDeleteNamespaceSuccess(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"delete", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.DeleteNamespace(ctx, kubeconfig, namespace); err != nil {
		t.Errorf("Kubectl.DeleteNamespace() error = %v, want nil", err)
	}
}

func TestKubectlDeleteNamespaceError(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"delete", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.DeleteNamespace(ctx, kubeconfig, namespace); err == nil {
		t.Errorf("Kubectl.DeleteNamespace() error = nil, want not nil")
	}
}

func TestKubectlDeleteSecretSuccess(t *testing.T) {
	t.Parallel()
	var secretName, namespace string

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"delete", "secret", secretName, "--kubeconfig", cluster.KubeconfigFile, "--namespace", namespace}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.DeleteSecret(ctx, cluster, secretName, namespace); err != nil {
		t.Errorf("Kubectl.DeleteNamespace() error = %v, want nil", err)
	}
}

func TestKubectlGetNamespaceSuccess(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"get", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.GetNamespace(ctx, kubeconfig, namespace); err != nil {
		t.Errorf("Kubectl.GetNamespace() error = %v, want nil", err)
	}
}

func TestKubectlGetNamespaceError(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace string

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"get", "namespace", namespace, "--kubeconfig", kubeconfig}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.GetNamespace(ctx, kubeconfig, namespace); err == nil {
		t.Errorf("Kubectl.GetNamespace() error = nil, want not nil")
	}
}

func TestKubectlWaitSuccess(t *testing.T) {
	t.Parallel()
	var timeout, kubeconfig, forCondition, property, namespace string

	k, ctx, _, e := newKubectl(t)

	// KubeCtl wait does not tolerate blank timeout values.
	// It also converts timeouts provided to seconds before actually invoking kubectl wait.
	timeout = "1m"
	expectedTimeout := "60.00s"

	expectedParam := []string{"wait", "--timeout", expectedTimeout, "--for=condition=" + forCondition, property, "--kubeconfig", kubeconfig, "-n", namespace}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.Wait(ctx, kubeconfig, timeout, forCondition, property, namespace); err != nil {
		t.Errorf("Kubectl.Wait() error = %v, want nil", err)
	}
}

func TestKubectlWaitBadTimeout(t *testing.T) {
	t.Parallel()
	var timeout, kubeconfig, forCondition, property, namespace string

	k, ctx, _, _ := newKubectl(t)

	timeout = "1y"
	if err := k.Wait(ctx, kubeconfig, timeout, forCondition, property, namespace); err == nil {
		t.Errorf("Kubectl.Wait() error = nil, want duration parse error")
	}

	timeout = "-1s"
	if err := k.Wait(ctx, kubeconfig, timeout, forCondition, property, namespace); err == nil {
		t.Errorf("Kubectl.Wait() error = nil, want duration parse error")
	}
}

func TestKubectlWaitRetryPolicy(t *testing.T) {
	t.Parallel()
	connectionRefusedError := fmt.Errorf("The connection to the server 127.0.0.1:56789 was refused")
	ioTimeoutError := fmt.Errorf("Unable to connect to the server 127.0.0.1:56789, i/o timeout\n")
	tlsHandshakeError := fmt.Errorf("Unable to connect to the server: net/http: TLS handshake timeout")
	miscellaneousError := fmt.Errorf("Some other random miscellaneous error")

	k := executables.NewKubectl(nil)

	_, wait := executables.KubectlWaitRetryPolicy(k, 1, connectionRefusedError)
	if wait != 10*time.Second {
		t.Errorf("kubectlWaitRetryPolicy didn't correctly calculate first retry wait for connection refused")
	}

	_, wait = executables.KubectlWaitRetryPolicy(k, -1, connectionRefusedError)
	if wait != 10*time.Second {
		t.Errorf("kubectlWaitRetryPolicy didn't correctly protect for total retries < 0")
	}

	_, wait = executables.KubectlWaitRetryPolicy(k, 2, connectionRefusedError)
	if wait != 15*time.Second {
		t.Errorf("kubectlWaitRetryPolicy didn't correctly protect for second retry wait")
	}

	_, wait = executables.KubectlWaitRetryPolicy(k, 1, ioTimeoutError)
	if wait != 10*time.Second {
		t.Errorf("kubectlWaitRetryPolicy didn't correctly calculate first retry wait for ioTimeout")
	}

	_, wait = executables.KubectlWaitRetryPolicy(k, 1, tlsHandshakeError)
	if wait != 10*time.Second {
		t.Errorf("kubectlWaitRetryPolicy didn't correctly calculate first retry wait for tls handshake error")
	}

	retry, _ := executables.KubectlWaitRetryPolicy(k, 1, miscellaneousError)
	if retry != false {
		t.Errorf("kubectlWaitRetryPolicy didn't not-retry on non-network error")
	}
}

func TestWaitForTimeout(t *testing.T) {
	t.Parallel()
	k := executables.Kubectl{}
	timeoutTime := time.Now()
	err := executables.CallKubectlPrivateWait(&k, nil, "", timeoutTime, "myCondition", "myProperty", "")
	if err == nil || err.Error() != "error: timed out waiting for condition myCondition on myProperty" {
		t.Errorf("kubectl private wait didn't timeout")
	}
}

func TestKubectlWaitForService(t *testing.T) {
	t.Parallel()
	testSvc := &corev1.Service{
		Spec: corev1.ServiceSpec{
			ClusterIP: "192.168.1.2",
		},
	}
	respJSON, err := json.Marshal(testSvc)
	if err != nil {
		t.Errorf("marshaling test service: %s", err)
	}
	ret := bytes.NewBuffer(respJSON)
	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"get", "--ignore-not-found", "-o", "json", "--kubeconfig", "kubeconfig", "service", "--namespace", "eksa-packages", "test"}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(*ret, nil).AnyTimes()
	if err := k.WaitForService(ctx, "kubeconfig", "5m", "test", "eksa-packages"); err != nil {
		t.Errorf("Kubectl.WaitForService() error = %v, want nil", err)
	}
}

func TestKubectlWaitForServiceWithLoadBalancer(t *testing.T) {
	t.Parallel()
	testSvc := &corev1.Service{
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{
						IP: "192.168.1.1",
					},
				},
			},
		},
	}
	respJSON, err := json.Marshal(testSvc)
	if err != nil {
		t.Errorf("marshaling test service: %s", err)
	}
	ret := bytes.NewBuffer(respJSON)
	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"get", "--ignore-not-found", "-o", "json", "--kubeconfig", "kubeconfig", "service", "--namespace", "eksa-packages", "test"}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(*ret, nil).AnyTimes()
	if err := k.WaitForService(ctx, "kubeconfig", "5m", "test", "eksa-packages"); err != nil {
		t.Errorf("Kubectl.WaitForService() error = %v, want nil", err)
	}
}

func TestKubectlWaitForServiceTimedOut(t *testing.T) {
	t.Parallel()
	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"get", "--ignore-not-found", "-o", "json", "--kubeconfig", "kubeconfig", "service", "--namespace", "eksa-packages", "test"}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil).AnyTimes()
	if err := k.WaitForService(ctx, "kubeconfig", "2s", "test", "eksa-packages"); err == nil {
		t.Errorf("Kubectl.WaitForService() error = nil, want %v", context.Canceled)
	}
}

func TestKubectlWaitForServiceBadTimeout(t *testing.T) {
	t.Parallel()
	k, ctx, _, _ := newKubectl(t)
	if err := k.WaitForService(ctx, "kubeconfig", "abc", "test", "eksa-packages"); err == nil {
		t.Errorf("Kubectl.WaitForService() error = nil, want parsing duration error")
	}
}

func TestKubectlWaitError(t *testing.T) {
	t.Parallel()
	var timeout, kubeconfig, forCondition, property, namespace string

	k, ctx, _, e := newKubectl(t)

	timeout = "1m"
	expectedTimeout := "60.00s"

	expectedParam := []string{"wait", "--timeout", expectedTimeout, "--for=condition=" + forCondition, property, "--kubeconfig", kubeconfig, "-n", namespace}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.Wait(ctx, kubeconfig, timeout, forCondition, property, namespace); err == nil {
		t.Errorf("Kubectl.Wait() error = nil, want not nil")
	}
}

func TestKubectlWaitNetworkErrorWithRetries(t *testing.T) {
	t.Parallel()
	var timeout, kubeconfig, forCondition, property, namespace string

	k, ctx, _, e := newKubectl(t)
	// Rebuild the kubectl executable with custom wait params so we avoid extremely long tests.
	k = executables.NewKubectl(e,
		executables.WithKubectlNetworkFaultBaseRetryTime(1*time.Millisecond),
		executables.WithNetworkFaultBackoffFactor(1),
	)

	timeout = "1m"
	expectedTimeout := "60.00s"

	expectedParam := []string{"wait", "--timeout", expectedTimeout, "--for=condition=" + forCondition, property, "--kubeconfig", kubeconfig, "-n", namespace}
	firstTry := e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("The connection to the server 127.0.0.1:56789 was refused")) //nolint:revive // The format of the message it's important here since the policy code depends on it

	// Kubectl Wait is intelligently adjusting the timeout param on retries. This is hard to predict from within the test
	// so I'm not having the mock validate params on the retried calls.

	secondTry := e.EXPECT().Execute(ctx, gomock.Any()).Return(bytes.Buffer{}, errors.New("Unable to connect to the server: 127.0.0.1: 56789, i/o timeout.\n")) //nolint:revive // The format of the message it's important here since the policy code depends on it

	thirdTry := e.EXPECT().Execute(ctx, gomock.Any()).Return(bytes.Buffer{}, nil)

	gomock.InOrder(
		firstTry,
		secondTry,
		thirdTry,
	)

	if err := k.Wait(ctx, kubeconfig, timeout, forCondition, property, namespace); err != nil {
		t.Errorf("Kubectl.Wait() error = %v, want nil", err)
	}
}

func TestKubectlSearchCloudStackMachineConfigs(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace, name string
	buffer := bytes.Buffer{}
	buffer.WriteString(test.ReadFile(t, "testdata/kubectl_no_cs_machineconfigs.json"))
	k, ctx, _, e := newKubectl(t)

	expectedParam := []string{
		"get", fmt.Sprintf("cloudstackmachineconfigs.%s", v1alpha1.GroupVersion.Group), "-o", "json", "--kubeconfig",
		kubeconfig, "--namespace", namespace, "--field-selector=metadata.name=" + name,
	}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(buffer, nil)
	mc, err := k.SearchCloudStackMachineConfig(ctx, name, kubeconfig, namespace)
	if err != nil {
		t.Errorf("Kubectl.SearchCloudStackMachineConfig() error = %v, want nil", err)
	}
	if len(mc) > 0 {
		t.Errorf("expected 0 machine configs, got %d", len(mc))
	}
}

func TestKubectlSearchCloudStackDatacenterConfigs(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace, name string
	buffer := bytes.Buffer{}
	buffer.WriteString(test.ReadFile(t, "testdata/kubectl_no_cs_datacenterconfigs.json"))
	k, ctx, _, e := newKubectl(t)

	expectedParam := []string{
		"get", fmt.Sprintf("cloudstackdatacenterconfigs.%s", v1alpha1.GroupVersion.Group), "-o", "json", "--kubeconfig",
		kubeconfig, "--namespace", namespace, "--field-selector=metadata.name=" + name,
	}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(buffer, nil)
	mc, err := k.SearchCloudStackDatacenterConfig(ctx, name, kubeconfig, namespace)
	if err != nil {
		t.Errorf("Kubectl.SearchCloudStackDatacenterConfig() error = %v, want nil", err)
	}
	if len(mc) > 0 {
		t.Errorf("expected 0 datacenter configs, got %d", len(mc))
	}
}

func TestCloudStackWorkerNodesMachineTemplate(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace, clusterName, machineTemplateName string
	machineTemplateNameBuffer := bytes.NewBufferString(machineTemplateName)
	machineTemplatesBuffer := bytes.NewBufferString(test.ReadFile(t, "testdata/kubectl_no_cs_machineconfigs.json"))
	k, ctx, _, e := newKubectl(t)
	expectedParam1 := []string{
		"get", "machinedeployments.cluster.x-k8s.io", fmt.Sprintf("%s-md-0", clusterName), "-o", "go-template",
		"--template", "{{.spec.template.spec.infrastructureRef.name}}", "--kubeconfig", kubeconfig, "--namespace", namespace,
	}
	expectedParam2 := []string{
		"get", "cloudstackmachinetemplates.infrastructure.cluster.x-k8s.io", machineTemplateName, "-o", "go-template", "--template",
		"{{.spec.template.spec}}", "-o", "yaml", "--kubeconfig", kubeconfig, "--namespace", namespace,
	}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam1)).Return(*machineTemplateNameBuffer, nil)
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam2)).Return(*machineTemplatesBuffer, nil)
	_, err := k.CloudstackWorkerNodesMachineTemplate(ctx, clusterName, kubeconfig, namespace)
	if err != nil {
		t.Errorf("Kubectl.GetNamespace() error = %v, want nil", err)
	}
}

func TestVsphereWorkerNodesMachineTemplate(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace, clusterName, machineTemplateName string
	machineTemplateNameBuffer := bytes.NewBufferString(machineTemplateName)
	machineTemplatesBuffer := bytes.NewBufferString(test.ReadFile(t, "testdata/kubectl_no_cs_machineconfigs.json"))
	k, ctx, _, e := newKubectl(t)
	expectedParam1 := []string{
		"get", "machinedeployments.cluster.x-k8s.io", fmt.Sprintf("%s-md-0", clusterName), "-o", "go-template",
		"--template", "{{.spec.template.spec.infrastructureRef.name}}", "--kubeconfig", kubeconfig, "--namespace", namespace,
	}
	expectedParam2 := []string{
		"get", "vspheremachinetemplates.infrastructure.cluster.x-k8s.io", machineTemplateName, "-o", "go-template", "--template",
		"{{.spec.template.spec}}", "-o", "yaml", "--kubeconfig", kubeconfig, "--namespace", namespace,
	}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam1)).Return(*machineTemplateNameBuffer, nil)
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam2)).Return(*machineTemplatesBuffer, nil)
	_, err := k.VsphereWorkerNodesMachineTemplate(ctx, clusterName, kubeconfig, namespace)
	if err != nil {
		t.Errorf("Kubectl.GetNamespace() error = %v, want nil", err)
	}
}

func TestKubectlSaveLogSuccess(t *testing.T) {
	t.Parallel()
	filename := "testfile"
	_, writer := test.NewWriter(t)

	deployment := &types.Deployment{
		Namespace: "namespace",
		Name:      "testname",
		Container: "container",
	}

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"--kubeconfig", cluster.KubeconfigFile, "logs", fmt.Sprintf("deployment/%s", deployment.Name), "-n", deployment.Namespace, "-c", deployment.Container}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.SaveLog(ctx, cluster, deployment, filename, writer); err != nil {
		t.Errorf("Kubectl.SaveLog() error = %v, want nil", err)
	}
}

func TestKubectlSaveLogError(t *testing.T) {
	t.Parallel()
	filename := "testfile"
	_, writer := test.NewWriter(t)

	deployment := &types.Deployment{
		Namespace: "namespace",
		Name:      "testname",
		Container: "container",
	}

	k, ctx, cluster, e := newKubectl(t)
	expectedParam := []string{"--kubeconfig", cluster.KubeconfigFile, "logs", fmt.Sprintf("deployment/%s", deployment.Name), "-n", deployment.Namespace, "-c", deployment.Container}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.SaveLog(ctx, cluster, deployment, filename, writer); err == nil {
		t.Errorf("Kubectl.SaveLog() error = nil, want not nil")
	}
}

func TestKubectlDeleteClusterSuccess(t *testing.T) {
	t.Parallel()
	kubeconfigFile := "c.kubeconfig"
	managementCluster := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
	}
	clusterToDelete := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
	}

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"delete", capiClustersResourceType, clusterToDelete.Name, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)
	if err := k.DeleteCluster(ctx, managementCluster, clusterToDelete); err != nil {
		t.Errorf("Kubectl.DeleteCluster() error = %v, want nil", err)
	}
}

func TestKubectlDeleteClusterError(t *testing.T) {
	t.Parallel()
	kubeconfigFile := "c.kubeconfig"
	managementCluster := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
	}
	clusterToDelete := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
	}

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"delete", capiClustersResourceType, clusterToDelete.Name, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, errors.New("error from execute"))
	if err := k.DeleteCluster(ctx, managementCluster, clusterToDelete); err == nil {
		t.Errorf("Kubectl.DeleteCluster() error = nil, want not nil")
	}
}

func TestKubectlGetMachines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantMachines     []types.Machine
	}{
		{
			testName:         "no machines",
			jsonResponseFile: "testdata/kubectl_no_machines.json",
			wantMachines:     []types.Machine{},
		},
		{
			testName:         "machines with no node ref",
			jsonResponseFile: "testdata/kubectl_machines_no_node_ref_no_labels.json",
			wantMachines: []types.Machine{
				{
					Metadata: types.MachineMetadata{
						Name: "eksa-test-capd-control-plane-5nfdg",
					},
					Status: types.MachineStatus{
						Conditions: types.Conditions{
							{
								Status: "True",
								Type:   "Ready",
							},
							{
								Status: "True",
								Type:   "APIServerPodHealthy",
							},
							{
								Status: "True",
								Type:   "BootstrapReady",
							},
							{
								Status: "True",
								Type:   "ControllerManagerPodHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdMemberHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdPodHealthy",
							},
							{
								Status: "True",
								Type:   "InfrastructureReady",
							},
							{
								Status: "True",
								Type:   "NodeHealthy",
							},
							{
								Status: "True",
								Type:   "SchedulerPodHealthy",
							},
						},
					},
				},
				{
					Metadata: types.MachineMetadata{
						Name: "eksa-test-capd-md-0-bb7885f6f-gkb85",
					},
					Status: types.MachineStatus{
						Conditions: types.Conditions{
							{
								Status: "True",
								Type:   "Ready",
							},
							{
								Status: "True",
								Type:   "BootstrapReady",
							},
							{
								Status: "True",
								Type:   "InfrastructureReady",
							},
							{
								Status: "True",
								Type:   "NodeHealthy",
							},
						},
					},
				},
			},
		},
		{
			testName:         "machines with no conditions",
			jsonResponseFile: "testdata/kubectl_machines_no_conditions.json",
			wantMachines: []types.Machine{
				{
					Metadata: types.MachineMetadata{
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":  "eksa-test-capd",
							"cluster.x-k8s.io/control-plane": "",
						},
						Name: "eksa-test-capd-control-plane-5nfdg",
					},
					Status: types.MachineStatus{
						NodeRef: &types.ResourceRef{
							APIVersion: "v1",
							Kind:       "Node",
							Name:       "eksa-test-capd-control-plane-5nfdg",
						},
					},
				},
				{
					Metadata: types.MachineMetadata{
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":    "eksa-test-capd",
							"cluster.x-k8s.io/deployment-name": "eksa-test-capd-md-0",
							"machine-template-hash":            "663441929",
						},
						Name: "eksa-test-capd-md-0-bb7885f6f-gkb85",
					},
					Status: types.MachineStatus{
						NodeRef: &types.ResourceRef{
							APIVersion: "v1",
							Kind:       "Node",
							Name:       "eksa-test-capd-md-0-bb7885f6f-gkb85",
						},
					},
				},
			},
		},
		{
			testName:         "machines with node ref",
			jsonResponseFile: "testdata/kubectl_machines_with_node_ref.json",
			wantMachines: []types.Machine{
				{
					Metadata: types.MachineMetadata{
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":  "eksa-test-capd",
							"cluster.x-k8s.io/control-plane": "",
						},
						Name: "eksa-test-capd-control-plane-5nfdg",
					},
					Status: types.MachineStatus{
						NodeRef: &types.ResourceRef{
							APIVersion: "v1",
							Kind:       "Node",
							Name:       "eksa-test-capd-control-plane-5nfdg",
						},
						Conditions: types.Conditions{
							{
								Status: "True",
								Type:   "Ready",
							},
							{
								Status: "True",
								Type:   "APIServerPodHealthy",
							},
							{
								Status: "True",
								Type:   "BootstrapReady",
							},
							{
								Status: "True",
								Type:   "ControllerManagerPodHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdMemberHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdPodHealthy",
							},
							{
								Status: "True",
								Type:   "InfrastructureReady",
							},
							{
								Status: "True",
								Type:   "NodeHealthy",
							},
							{
								Status: "True",
								Type:   "SchedulerPodHealthy",
							},
						},
					},
				},
				{
					Metadata: types.MachineMetadata{
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":    "eksa-test-capd",
							"cluster.x-k8s.io/deployment-name": "eksa-test-capd-md-0",
							"machine-template-hash":            "663441929",
						},
						Name: "eksa-test-capd-md-0-bb7885f6f-gkb85",
					},
					Status: types.MachineStatus{
						NodeRef: &types.ResourceRef{
							APIVersion: "v1",
							Kind:       "Node",
							Name:       "eksa-test-capd-md-0-bb7885f6f-gkb85",
						},
						Conditions: types.Conditions{
							{
								Status: "True",
								Type:   "Ready",
							},
							{
								Status: "True",
								Type:   "BootstrapReady",
							},
							{
								Status: "True",
								Type:   "InfrastructureReady",
							},
							{
								Status: "True",
								Type:   "NodeHealthy",
							},
						},
					},
				},
			},
		},
		{
			testName:         "etcd machines",
			jsonResponseFile: "testdata/kubectl_etcd_machines_no_node_ref.json",
			wantMachines: []types.Machine{
				{
					Metadata: types.MachineMetadata{
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name": "eksa-test-capd",
							"cluster.x-k8s.io/etcd-cluster": "",
						},
						Name: "eksa-test-capd-control-plane-5nfdg",
					},
					Status: types.MachineStatus{
						Conditions: types.Conditions{
							{
								Status: "True",
								Type:   "Ready",
							},
							{
								Status: "True",
								Type:   "APIServerPodHealthy",
							},
							{
								Status: "True",
								Type:   "BootstrapReady",
							},
							{
								Status: "True",
								Type:   "ControllerManagerPodHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdMemberHealthy",
							},
							{
								Status: "True",
								Type:   "EtcdPodHealthy",
							},
							{
								Status: "True",
								Type:   "InfrastructureReady",
							},
							{
								Status: "True",
								Type:   "NodeHealthy",
							},
							{
								Status: "True",
								Type:   "SchedulerPodHealthy",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			k, ctx, cluster, e := newKubectl(t)
			e.EXPECT().Execute(ctx, []string{
				"get", "machines.cluster.x-k8s.io", "-o", "json", "--kubeconfig", cluster.KubeconfigFile,
				"--selector=cluster.x-k8s.io/cluster-name=" + cluster.Name,
				"--namespace", constants.EksaSystemNamespace,
			}).Return(*bytes.NewBufferString(fileContent), nil)

			gotMachines, err := k.GetMachines(ctx, cluster, cluster.Name)
			if err != nil {
				t.Fatalf("Kubectl.GetMachines() error = %v, want nil", err)
			}

			if !reflect.DeepEqual(gotMachines, tt.wantMachines) {
				t.Fatalf("Kubectl.GetMachines() machines = %+v, want %+v", gotMachines, tt.wantMachines)
			}
		})
	}
}

func TestKubectlGetEksaCloudStackMachineConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantMachines     *v1alpha1.CloudStackMachineConfig
	}{
		{
			testName:         "no machines",
			jsonResponseFile: "testdata/kubectl_no_cs_machineconfigs.json",
			wantMachines: &v1alpha1.CloudStackMachineConfig{
				TypeMeta: metav1.TypeMeta{APIVersion: "v1"},
			},
		},
		{
			testName:         "one machineconfig",
			jsonResponseFile: "testdata/kubectl_eksa_cs_machineconfig.json",
			wantMachines: &v1alpha1.CloudStackMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CloudStackMachineConfig",
					APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "test-etcd"},
				Spec: v1alpha1.CloudStackMachineConfigSpec{
					Template: v1alpha1.CloudStackResourceIdentifier{
						Name: "testTemplate",
					},
					ComputeOffering: v1alpha1.CloudStackResourceIdentifier{
						Name: "testOffering",
					},
					DiskOffering: &v1alpha1.CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
							Name: "testOffering",
						},
						MountPath:  "/data",
						Device:     "/dev/vdb",
						Filesystem: "ext4",
						Label:      "data_disk",
					},
					Users: []v1alpha1.UserConfiguration{
						{
							Name:              "maxdrib",
							SshAuthorizedKeys: []string{"ssh-rsa test123 hi"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			k, ctx, cluster, e := newKubectl(t)
			machineConfigName := "testMachineConfig"
			e.EXPECT().Execute(ctx, []string{
				"get", "--ignore-not-found",
				"-o", "json", "--kubeconfig", cluster.KubeconfigFile,
				"cloudstackmachineconfigs.anywhere.eks.amazonaws.com",
				"--namespace", constants.EksaSystemNamespace,
				machineConfigName,
			}).Return(*bytes.NewBufferString(fileContent), nil)

			gotMachines, err := k.GetEksaCloudStackMachineConfig(ctx, machineConfigName, cluster.KubeconfigFile, constants.EksaSystemNamespace)
			if err != nil {
				t.Fatalf("Kubectl.GetEksaCloudStackMachineConfig() error = %v, want nil", err)
			}

			if !reflect.DeepEqual(gotMachines, tt.wantMachines) {
				t.Fatalf("Kubectl.GetEksaCloudStackMachineConfig() machines = %+v, want %+v", gotMachines, tt.wantMachines)
			}
		})
	}
}

func TestKubectlGetEksaCloudStackDatacenterConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantDatacenter   *v1alpha1.CloudStackDatacenterConfig
	}{
		{
			testName:         "no datacenter",
			jsonResponseFile: "testdata/kubectl_no_cs_datacenterconfigs.json",
			wantDatacenter: &v1alpha1.CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{APIVersion: "v1"},
			},
		},
		{
			testName:         "one datacenter availability zones",
			jsonResponseFile: "testdata/kubectl_eksa_cs_datacenterconfig_az.json",
			wantDatacenter: &v1alpha1.CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CloudStackDatacenterConfig",
					APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: v1alpha1.CloudStackDatacenterConfigSpec{
					AvailabilityZones: []v1alpha1.CloudStackAvailabilityZone{{
						Name: "default-az-0",
						Zone: v1alpha1.CloudStackZone{
							Name: "testZone",
							Network: v1alpha1.CloudStackResourceIdentifier{
								Name: "testNetwork",
							},
						},
						CredentialsRef: "global",
						Domain:         "testDomain",
						Account:        "testAccount",
					}},
				},
			},
		},
		{
			testName:         "one datacenter legacy zones",
			jsonResponseFile: "testdata/kubectl_eksa_cs_datacenterconfig.json",
			wantDatacenter: &v1alpha1.CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CloudStackDatacenterConfig",
					APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: v1alpha1.CloudStackDatacenterConfigSpec{
					Domain:  "testDomain",
					Account: "testAccount",
					Zones: []v1alpha1.CloudStackZone{
						{
							Name: "testZone",
							Network: v1alpha1.CloudStackResourceIdentifier{
								Name: "testNetwork",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			k, ctx, cluster, e := newKubectl(t)
			datacenterConfigName := "testDatacenterConfig"
			e.EXPECT().Execute(ctx, []string{
				"get", "--ignore-not-found",
				"-o", "json", "--kubeconfig", cluster.KubeconfigFile,
				"cloudstackdatacenterconfigs.anywhere.eks.amazonaws.com",
				"--namespace", constants.EksaSystemNamespace,
				datacenterConfigName,
			}).Return(*bytes.NewBufferString(fileContent), nil)

			gotDatacenter, err := k.GetEksaCloudStackDatacenterConfig(ctx, datacenterConfigName, cluster.KubeconfigFile, constants.EksaSystemNamespace)
			if err != nil {
				t.Fatalf("Kubectl.GetEksaCloudStackDatacenterConfig() error = %v, want nil", err)
			}

			if !gotDatacenter.Spec.Equal(&tt.wantDatacenter.Spec) {
				t.Fatalf("Kubectl.GetEksaCloudStackDatacenterConfig() machines = %+v, want %+v", gotDatacenter, tt.wantDatacenter)
			}
		})
	}
}

func TestKubectlLoadSecret(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName string
		params   []string
		wantErr  error
	}{
		{
			testName: "SuccessScenario",
			params:   []string{"create", "secret", "generic", secretObjectName, "--type", secretObjectType, "--from-literal", "test_cluster", "--kubeconfig", "test_cluster.kind.kubeconfig", "--namespace", constants.EksaSystemNamespace},
			wantErr:  nil,
		},
		{
			testName: "ErrorScenario",
			params:   []string{"create", "secret", "generic", secretObjectName, "--type", secretObjectType, "--from-literal", "test_cluster", "--kubeconfig", "test_cluster.kind.kubeconfig", "--namespace", constants.EksaSystemNamespace},
			wantErr:  errors.New("error loading secret: "),
		},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(tt *testing.T) {
			k, ctx, _, e := newKubectl(t)
			e.EXPECT().Execute(ctx, tc.params).Return(bytes.Buffer{}, tc.wantErr)

			err := k.LoadSecret(ctx, "test_cluster", secretObjectType, secretObjectName, "test_cluster.kind.kubeconfig")

			if (tc.wantErr != nil && err == nil) && !reflect.DeepEqual(tc.wantErr, err) {
				t.Errorf("%v got = %v, want %v", tc.testName, err, tc.wantErr)
			}
		})
	}
}

func TestKubectlGetSecretFromNamespaceSuccess(t *testing.T) {
	t.Parallel()
	newKubectlGetterTest(t).withResourceType(
		"secret",
	).withGetter(func(tt *kubectlGetterTest) (client.Object, error) {
		return tt.k.GetSecretFromNamespace(tt.ctx, tt.kubeconfig, tt.name, tt.namespace)
	}).withJsonFromFile(
		"testdata/kubectl_secret.json",
	).andWant(
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vsphere-cloud-controller-manager",
				Namespace: "eksa-system",
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			Data: map[string][]byte{
				"data": []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloud-controller-manager
  namespace: kube-system
`),
			},
			Type: corev1.SecretType("addons.cluster.x-k8s.io/resource-set"),
		},
	).testSuccess()
}

func TestKubectlGetSecretFromNamespaceError(t *testing.T) {
	t.Parallel()
	newKubectlGetterTest(t).withResourceType(
		"secret",
	).withGetter(func(tt *kubectlGetterTest) (client.Object, error) {
		return tt.k.GetSecretFromNamespace(tt.ctx, tt.kubeconfig, tt.name, tt.namespace)
	}).testError()
}

func TestKubectlGetSecret(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName     string
		responseFile string
		wantSecret   *corev1.Secret
		params       []string
		wantErr      error
	}{
		{
			testName:     "SuccessScenario",
			responseFile: "testdata/kubectl_secret.json",
			wantSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vsphere-cloud-controller-manager",
					Namespace: "eksa-system",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				Data: map[string][]byte{
					"data": []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloud-controller-manager
  namespace: kube-system
`),
				},
				Type: corev1.SecretType("addons.cluster.x-k8s.io/resource-set"),
			},
			params:  []string{"get", "secret", secretObjectName, "-o", "json", "--namespace", constants.EksaSystemNamespace, "--kubeconfig", "c.kubeconfig"},
			wantErr: nil,
		},
		{
			testName:     "ErrorScenario",
			responseFile: "testdata/kubectl_secret.json",
			wantSecret:   nil,
			params:       []string{"get", "secret", secretObjectName, "-o", "json", "--namespace", constants.EksaSystemNamespace, "--kubeconfig", "c.kubeconfig"},
			wantErr:      errors.New("error from kubectl client"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(tt *testing.T) {
			response := test.ReadFile(t, tc.responseFile)
			k, ctx, cluster, e := newKubectl(t)
			e.EXPECT().Execute(ctx, tc.params).Return(*bytes.NewBufferString(response), tc.wantErr)

			secret, err := k.GetSecret(ctx, secretObjectName, executables.WithNamespace(constants.EksaSystemNamespace), executables.WithCluster(cluster))

			g := NewWithT(t)
			if tc.wantErr != nil {
				g.Expect(err.Error()).To(HaveSuffix(tc.wantErr.Error()))
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(secret).To(Equal(tc.wantSecret))
			}
		})
	}
}

func TestKubectlGetClusters(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantClusters     []types.CAPICluster
	}{
		{
			testName:         "no clusters",
			jsonResponseFile: "testdata/kubectl_no_clusters.json",
			wantClusters:     []types.CAPICluster{},
		},
		{
			testName:         "machines with node ref",
			jsonResponseFile: "testdata/kubectl_clusters_one.json",
			wantClusters: []types.CAPICluster{
				{
					Metadata: types.Metadata{
						Name: "eksa-test-capd",
					},
					Status: types.ClusterStatus{
						Phase: "Provisioned",
						Conditions: []types.Condition{
							{Type: "Ready", Status: "True"},
							{Type: "ControlPlaneReady", Status: "True"},
							{Type: "InfrastructureReady", Status: "True"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			k, ctx, cluster, e := newKubectl(t)
			e.EXPECT().Execute(ctx, []string{"get", "clusters.cluster.x-k8s.io", "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}).Return(*bytes.NewBufferString(fileContent), nil)

			gotClusters, err := k.GetClusters(ctx, cluster)
			if err != nil {
				t.Fatalf("Kubectl.GetClusters() error = %v, want nil", err)
			}

			if !reflect.DeepEqual(gotClusters, tt.wantClusters) {
				t.Fatalf("Kubectl.GetClusters() clusters = %+v, want %+v", gotClusters, tt.wantClusters)
			}
		})
	}
}

func TestKubectlGetEKSAClusters(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName         string
		jsonResponseFile string
		expectedSpec     v1alpha1.ClusterSpec
		clusterName      string
	}{
		{
			testName:         "EKS-A cluster found",
			jsonResponseFile: "testdata/kubectl_eksa_cluster.json",
			expectedSpec: v1alpha1.ClusterSpec{
				KubernetesVersion:             "1.19",
				ControlPlaneConfiguration:     v1alpha1.ControlPlaneConfiguration{Count: 3},
				WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3)}},
				DatacenterRef: v1alpha1.Ref{
					Kind: v1alpha1.VSphereDatacenterKind,
					Name: "test-cluster",
				},
			},
			clusterName: "test-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			k, ctx, cluster, e := newKubectl(t)
			e.EXPECT().Execute(ctx, []string{"get", "clusters.anywhere.eks.amazonaws.com", "-A", "-o", "jsonpath={.items[0]}", "--kubeconfig", cluster.KubeconfigFile, "--field-selector=metadata.name=" + tt.clusterName}).Return(*bytes.NewBufferString(fileContent), nil)

			gotCluster, err := k.GetEksaCluster(ctx, cluster, tt.clusterName)
			if err != nil {
				t.Fatalf("Kubectl.GetEKSAClusters() error = %v, want nil", err)
			}

			if !reflect.DeepEqual(gotCluster.Spec, tt.expectedSpec) {
				t.Fatalf("Kubectl.GetEKSAClusters() clusters = %+v, want %+v", gotCluster.Spec, tt.expectedSpec)
			}
		})
	}
}

func TestKubectlGetGetApiServerUrlSuccess(t *testing.T) {
	t.Parallel()
	wantUrl := "https://127.0.0.1:37479"
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(
		ctx,
		[]string{"config", "view", "--kubeconfig", cluster.KubeconfigFile, "--minify", "--raw", "-o", "jsonpath={.clusters[0].cluster.server}"},
	).Return(*bytes.NewBufferString(wantUrl), nil)

	gotUrl, err := k.GetApiServerUrl(ctx, cluster)
	if err != nil {
		t.Fatalf("Kubectl.GetApiServerUrl() error = %v, want nil", err)
	}

	if gotUrl != wantUrl {
		t.Fatalf("Kubectl.GetApiServerUrl() url = %s, want %s", gotUrl, wantUrl)
	}
}

func TestKubectlSetControllerEnvVarSuccess(t *testing.T) {
	t.Parallel()
	envVar := "TEST_VAR"
	envVarValue := "TEST_VALUE"
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(
		ctx,
		[]string{
			"set", "env", "deployment/eksa-controller-manager", fmt.Sprintf("%s=%s", envVar, envVarValue),
			"--kubeconfig", cluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace,
		},
	).Return(bytes.Buffer{}, nil)

	err := k.SetEksaControllerEnvVar(ctx, envVar, envVarValue, cluster.KubeconfigFile)
	if err != nil {
		t.Fatalf("Kubectl.RolloutRestartDaemonSet() error = %v, want nil", err)
	}
}

func TestKubectlRolloutRestartDaemonSetSuccess(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(
		ctx,
		[]string{
			"rollout", "restart", "daemonset", "cilium",
			"--kubeconfig", cluster.KubeconfigFile, "--namespace", constants.KubeSystemNamespace,
		},
	).Return(bytes.Buffer{}, nil)

	err := k.RolloutRestartDaemonSet(ctx, "cilium", constants.KubeSystemNamespace, cluster.KubeconfigFile)
	if err != nil {
		t.Fatalf("Kubectl.RolloutRestartDaemonSet() error = %v, want nil", err)
	}
}

func TestKubectlRolloutRestartDaemonSetError(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(
		ctx,
		[]string{
			"rollout", "restart", "daemonset", "cilium",
			"--kubeconfig", cluster.KubeconfigFile, "--namespace", constants.KubeSystemNamespace,
		},
	).Return(bytes.Buffer{}, fmt.Errorf("error"))

	err := k.RolloutRestartDaemonSet(ctx, "cilium", constants.KubeSystemNamespace, cluster.KubeconfigFile)
	if err == nil {
		t.Fatalf("Kubectl.RolloutRestartDaemonSet() expected error, but was nil")
	}
}

func TestKubectlGetGetApiServerUrlError(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(
		ctx,
		[]string{"config", "view", "--kubeconfig", cluster.KubeconfigFile, "--minify", "--raw", "-o", "jsonpath={.clusters[0].cluster.server}"},
	).Return(bytes.Buffer{}, errors.New("error in command"))

	_, err := k.GetApiServerUrl(ctx, cluster)
	if err == nil {
		t.Fatal("Kubectl.GetApiServerUrl() error = nil, want not nil")
	}
}

func TestKubectlGetPodsWithAllNamespaces(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantPodNames     []string
	}{
		{
			testName:         "no pods",
			jsonResponseFile: "testdata/kubectl_no_pods.json",
			wantPodNames:     []string{},
		},
		{
			testName:         "multiple pods",
			jsonResponseFile: "testdata/kubectl_pods.json",
			wantPodNames: []string{
				"coredns-74ff55c5b-cnbwh",
				"coredns-74ff55c5b-zlbph",
				"etcd-lol-control-plane",
				"kindnet-xzddb",
				"kube-apiserver-lol-control-plane",
				"kube-controller-manager-lol-control-plane",
				"kube-proxy-27v6c",
				"kube-scheduler-lol-control-plane",
				"local-path-provisioner-78776bfc44-s9ggt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "pods", "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "-A"}).Return(*bytes.NewBufferString(fileContent), nil)

			gotPods, err := k.GetPods(ctx, executables.WithCluster(cluster), executables.WithAllNamespaces())
			if err != nil {
				t.Fatalf("Kubectl.GetPods() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotPods))
			for _, p := range gotPods {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantPodNames) {
				t.Fatalf("Kubectl.GetPods() pod names = %+v, want %+v", gotNames, tt.wantPodNames)
			}
		})
	}
}

func TestKubectlGetPodsWithNamespace(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantPodNames     []string
		namespace        string
	}{
		{
			testName:         "no pods",
			jsonResponseFile: "testdata/kubectl_no_pods.json",
			wantPodNames:     []string{},
			namespace:        "kube-system",
		},
		{
			testName:         "multiple pods",
			jsonResponseFile: "testdata/kubectl_pods.json",
			wantPodNames: []string{
				"coredns-74ff55c5b-cnbwh",
				"coredns-74ff55c5b-zlbph",
				"etcd-lol-control-plane",
				"kindnet-xzddb",
				"kube-apiserver-lol-control-plane",
				"kube-controller-manager-lol-control-plane",
				"kube-proxy-27v6c",
				"kube-scheduler-lol-control-plane",
				"local-path-provisioner-78776bfc44-s9ggt",
			},
			namespace: "kube-system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "pods", "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", tt.namespace}).Return(*bytes.NewBufferString(fileContent), nil)

			gotPods, err := k.GetPods(ctx, executables.WithCluster(cluster), executables.WithNamespace(tt.namespace))
			if err != nil {
				t.Fatalf("Kubectl.GetPods() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotPods))
			for _, p := range gotPods {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantPodNames) {
				t.Fatalf("Kubectl.GetPods() pod names = %+v, want %+v", gotNames, tt.wantPodNames)
			}
		})
	}
}

func TestKubectlGetPodsWithServerSkipTLSAndToken(t *testing.T) {
	t.Parallel()
	k, ctx, _, e := newKubectl(t)
	server := "https://127.0.0.1:37479"
	token := "token"
	fileContent := test.ReadFile(t, "testdata/kubectl_no_pods.json")
	e.EXPECT().Execute(
		ctx, []string{"get", "pods", "-o", "json", "--server", server, "--token", token, "--insecure-skip-tls-verify=true", "-A"},
	).Return(*bytes.NewBufferString(fileContent), nil)

	gotPods, err := k.GetPods(ctx,
		executables.WithServer(server), executables.WithToken(token), executables.WithSkipTLSVerify(), executables.WithAllNamespaces(),
	)
	if err != nil {
		t.Fatalf("Kubectl.GetPods() error = %v, want nil", err)
	}

	if len(gotPods) != 0 {
		t.Fatalf("Kubectl.GetPods() num pod  = %d, want 0", len(gotPods))
	}
}

func TestKubectlGetDeploymentsWithAllNamespaces(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName            string
		jsonResponseFile    string
		wantDeploymentNames []string
	}{
		{
			testName:            "no deployments",
			jsonResponseFile:    "testdata/kubectl_no_deployments.json",
			wantDeploymentNames: []string{},
		},
		{
			testName:         "multiple deployments",
			jsonResponseFile: "testdata/kubectl_deployments.json",
			wantDeploymentNames: []string{
				"coredns",
				"local-path-provisioner",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "deployments", "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "-A"}).Return(*bytes.NewBufferString(fileContent), nil)

			gotDeployments, err := k.GetDeployments(ctx, executables.WithCluster(cluster), executables.WithAllNamespaces())
			if err != nil {
				t.Fatalf("Kubectl.GetDeployments() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotDeployments))
			for _, p := range gotDeployments {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantDeploymentNames) {
				t.Fatalf("Kubectl.GetDeployments() deployments = %+v, want %+v", gotNames, tt.wantDeploymentNames)
			}
		})
	}
}

func TestKubectlGetDeploymentsWithServerSkipTLSAndToken(t *testing.T) {
	t.Parallel()
	server := "https://127.0.0.1:37479"
	token := "token"
	k, ctx, _, e := newKubectl(t)
	fileContent := test.ReadFile(t, "testdata/kubectl_no_deployments.json")
	e.EXPECT().Execute(
		ctx,
		[]string{"get", "deployments", "-o", "json", "--server", server, "--token", token, "--insecure-skip-tls-verify=true", "-A"}).Return(*bytes.NewBufferString(fileContent), nil)

	gotDeployments, err := k.GetDeployments(
		ctx,
		executables.WithServer(server), executables.WithToken(token), executables.WithSkipTLSVerify(), executables.WithAllNamespaces(),
	)
	if err != nil {
		t.Fatalf("Kubectl.GetDeployments() error = %v, want nil", err)
	}

	if len(gotDeployments) != 0 {
		t.Fatalf("Kubectl.GetDeployments() num deployments = %d, want 0", len(gotDeployments))
	}
}

func TestKubectlGetDeploymentsWithNamespace(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName            string
		jsonResponseFile    string
		wantDeploymentNames []string
		deploymentNamespace string
	}{
		{
			testName:            "no deployments",
			jsonResponseFile:    "testdata/kubectl_no_deployments.json",
			wantDeploymentNames: []string{},
			deploymentNamespace: "kube-system",
		},
		{
			testName:         "multiple deployments",
			jsonResponseFile: "testdata/kubectl_deployments.json",
			wantDeploymentNames: []string{
				"coredns",
				"local-path-provisioner",
			},
			deploymentNamespace: "kube-system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "deployments", "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", tt.deploymentNamespace}).Return(*bytes.NewBufferString(fileContent), nil)

			gotDeployments, err := k.GetDeployments(ctx, executables.WithCluster(cluster), executables.WithNamespace(tt.deploymentNamespace))
			if err != nil {
				t.Fatalf("Kubectl.GetDeployments() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotDeployments))
			for _, p := range gotDeployments {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantDeploymentNames) {
				t.Fatalf("Kubectl.GetDeployments() deployments = %+v, want %+v", gotNames, tt.wantDeploymentNames)
			}
		})
	}
}

func TestKubectlGetMachineDeployments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName                   string
		jsonResponseFile           string
		wantMachineDeploymentNames []string
	}{
		{
			testName:                   "no machine deployments",
			jsonResponseFile:           "testdata/kubectl_no_machine_deployments.json",
			wantMachineDeploymentNames: []string{},
		},
		{
			testName:         "multiple machine deployments",
			jsonResponseFile: "testdata/kubectl_machine_deployments.json",
			wantMachineDeploymentNames: []string{
				"test0-md-0",
				"test1-md-0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "machinedeployments.cluster.x-k8s.io", "-o", "json", "--kubeconfig", cluster.KubeconfigFile}).Return(*bytes.NewBufferString(fileContent), nil)

			gotDeployments, err := k.GetMachineDeployments(ctx, executables.WithCluster(cluster))
			if err != nil {
				t.Fatalf("Kubectl.GetMachineDeployments() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotDeployments))
			for _, p := range gotDeployments {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantMachineDeploymentNames) {
				t.Fatalf("Kubectl.GetMachineDeployments() deployments = %+v, want %+v", gotNames, tt.wantMachineDeploymentNames)
			}
		})
	}
}

func TestKubectlCountMachineDeploymentReplicasReady(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantError        bool
		wantTotal        int
		wantReady        int
		returnError      bool
	}{
		{
			testName:         "no machine deployments",
			jsonResponseFile: "testdata/kubectl_no_machine_deployments.json",
			wantError:        false,
			wantReady:        0,
			wantTotal:        0,
			returnError:      false,
		},
		{
			testName:         "multiple machine deployments",
			jsonResponseFile: "testdata/kubectl_machine_deployments.json",
			wantError:        false,
			wantReady:        2,
			wantTotal:        2,
			returnError:      false,
		},
		{
			testName:         "multiple machine deployments with unready replicas",
			jsonResponseFile: "testdata/kubectl_machine_deployments_unready.json",
			wantError:        false,
			wantReady:        2,
			wantTotal:        3,
			returnError:      false,
		},
		{
			testName:         "non-running machine deployments",
			jsonResponseFile: "testdata/kubectl_machine_deployments_provisioned.json",
			wantError:        true,
			wantReady:        0,
			wantTotal:        0,
			returnError:      false,
		},
		{
			testName:         "unavailable replicas",
			jsonResponseFile: "testdata/kubectl_machine_deployments_unavailable.json",
			wantError:        true,
			wantReady:        0,
			wantTotal:        0,
		},
		{
			testName:         "error response",
			jsonResponseFile: "",
			wantError:        true,
			wantReady:        0,
			wantTotal:        0,
			returnError:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			tt := newKubectlTest(t)
			if tc.returnError {
				e.EXPECT().
					Execute(ctx, []string{
						"get", "machinedeployments.cluster.x-k8s.io",
						"-o", "json",
						"--kubeconfig", cluster.KubeconfigFile,
						"--namespace", "eksa-system",
						"--selector=cluster.x-k8s.io/cluster-name=test-cluster",
					}).
					Return(*bytes.NewBufferString(""), errors.New(""))
			} else {
				fileContent := test.ReadFile(t, tc.jsonResponseFile)
				e.EXPECT().
					Execute(ctx, []string{
						"get", "machinedeployments.cluster.x-k8s.io",
						"-o", "json",
						"--kubeconfig", cluster.KubeconfigFile,
						"--namespace", "eksa-system",
						"--selector=cluster.x-k8s.io/cluster-name=test-cluster",
					}).
					Return(*bytes.NewBufferString(fileContent), nil)
			}

			ready, total, err := k.CountMachineDeploymentReplicasReady(ctx, cluster.Name, cluster.KubeconfigFile)
			if tc.wantError {
				tt.Expect(err).NotTo(BeNil())
			} else {
				tt.Expect(err).To(BeNil())
			}
			tt.Expect(ready).To(Equal(tc.wantReady))
			tt.Expect(total).To(Equal(tc.wantTotal))
		})
	}
}

func TestKubectlValidateWorkerNodes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantError        bool
	}{
		{
			testName:         "no machine deployments",
			jsonResponseFile: "testdata/kubectl_no_machine_deployments.json",
			wantError:        false,
		},
		{
			testName:         "multiple machine deployments",
			jsonResponseFile: "testdata/kubectl_machine_deployments.json",
			wantError:        false,
		},
		{
			testName:         "multiple machine deployments with unready replicas",
			jsonResponseFile: "testdata/kubectl_machine_deployments_unready.json",
			wantError:        true,
		},
		{
			testName:         "non-running machine deployments",
			jsonResponseFile: "testdata/kubectl_machine_deployments_provisioned.json",
			wantError:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			tt := newKubectlTest(t)
			fileContent := test.ReadFile(t, tc.jsonResponseFile)
			e.EXPECT().
				Execute(ctx, []string{
					"get", "machinedeployments.cluster.x-k8s.io",
					"-o", "json",
					"--kubeconfig", cluster.KubeconfigFile,
					"--namespace", "eksa-system",
					"--selector=cluster.x-k8s.io/cluster-name=test-cluster",
				}).
				Return(*bytes.NewBufferString(fileContent), nil)

			err := k.ValidateWorkerNodes(ctx, cluster.Name, cluster.KubeconfigFile)
			if tc.wantError {
				tt.Expect(err).NotTo(BeNil())
			} else {
				tt.Expect(err).To(BeNil())
			}
		})
	}
}

func TestKubectlGetKubeAdmControlPlanes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantCpNames      []string
	}{
		{
			testName:         "no control planes",
			jsonResponseFile: "testdata/kubectl_no_kubeadmcps.json",
			wantCpNames:      []string{},
		},
		{
			testName:         "multiple control planes",
			jsonResponseFile: "testdata/kubectl_kubeadmcps.json",
			wantCpNames: []string{
				"test0-control-plane",
				"test1-control-plane",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			k, ctx, cluster, e := newKubectl(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			e.EXPECT().Execute(ctx, []string{"get", "kubeadmcontrolplanes.controlplane.cluster.x-k8s.io", "-o", "json", "--kubeconfig", cluster.KubeconfigFile}).Return(*bytes.NewBufferString(fileContent), nil)

			gotCps, err := k.GetKubeadmControlPlanes(ctx, executables.WithCluster(cluster))
			if err != nil {
				t.Fatalf("Kubectl.GetKubeadmControlPlanes() error = %v, want nil", err)
			}

			gotNames := make([]string, 0, len(gotCps))
			for _, p := range gotCps {
				gotNames = append(gotNames, p.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantCpNames) {
				t.Fatalf("Kubectl.GetKubeadmControlPlanes() controlPlanes = %+v, want %+v", gotNames, tt.wantCpNames)
			}
		})
	}
}

func TestKubectlVersion(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	fileContent := test.ReadFile(t, "testdata/kubectl_version.json")
	e.EXPECT().Execute(ctx, []string{"version", "-o", "json", "--kubeconfig", cluster.KubeconfigFile}).Return(*bytes.NewBufferString(fileContent), nil)
	gotVersions, err := k.Version(ctx, cluster)
	if err != nil {
		t.Fatalf("Kubectl.Version() error = %v, want nil", err)
	}
	wantVersions := &executables.VersionResponse{
		ClientVersion: version.Info{
			Major:        "1",
			Minor:        "21",
			GitVersion:   "v1.21.2",
			GitCommit:    "092fbfbf53427de67cac1e9fa54aaa09a28371d7",
			GitTreeState: "clean",
			BuildDate:    "2021-06-16T12:59:11Z",
			GoVersion:    "go1.16.5",
			Compiler:     "gc",
			Platform:     "darwin/amd64",
		},
		ServerVersion: version.Info{
			Major:        "1",
			Minor:        "18+",
			GitVersion:   "v1.18.16-eks-1-18-4",
			GitCommit:    "3cdb4c9ab835e2964c8eaeb3ee77d088c7fa36aa",
			GitTreeState: "clean",
			BuildDate:    "2021-05-05T13:09:23Z",
			GoVersion:    "go1.13.15",
			Compiler:     "gc",
			Platform:     "linux/amd64",
		},
	}
	if !reflect.DeepEqual(gotVersions, wantVersions) {
		t.Fatalf("Kubectl.Version() versionResponse = %+v, want %+v", gotVersions, wantVersions)
	}
}

func TestKubectlValidateClustersCRDSuccess(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{"get", "customresourcedefinition", "clusters.cluster.x-k8s.io", "--kubeconfig", cluster.KubeconfigFile}).Return(bytes.Buffer{}, nil)
	err := k.ValidateClustersCRD(ctx, cluster)
	if err != nil {
		t.Fatalf("Kubectl.ValidateClustersCRD() error = %v, want nil", err)
	}
}

func TestKubectlValidateClustersCRDNotFound(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{"get", "customresourcedefinition", "clusters.cluster.x-k8s.io", "--kubeconfig", cluster.KubeconfigFile}).Return(bytes.Buffer{}, errors.New("CRD not found"))
	err := k.ValidateClustersCRD(ctx, cluster)
	if err == nil {
		t.Fatalf("Kubectl.ValidateClustersCRD() error == nil, want CRD not found")
	}
}

func TestKubectlValidateEKSAClustersCRDSuccess(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{"get", "customresourcedefinition", "clusters.anywhere.eks.amazonaws.com", "--kubeconfig", cluster.KubeconfigFile}).Return(bytes.Buffer{}, nil)
	err := k.ValidateEKSAClustersCRD(ctx, cluster)
	if err != nil {
		t.Fatalf("Kubectl.ValidateEKSAClustersCRD() error = %v, want nil", err)
	}
}

func TestKubectlValidateEKSAClustersCRDNotFound(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{"get", "customresourcedefinition", "clusters.anywhere.eks.amazonaws.com", "--kubeconfig", cluster.KubeconfigFile}).Return(bytes.Buffer{}, errors.New("CRD not found"))
	err := k.ValidateEKSAClustersCRD(ctx, cluster)
	if err == nil {
		t.Fatalf("Kubectl.ValidateEKSAClustersCRD() error == nil, want CRD not found")
	}
}

func TestKubectlUpdateEnvironmentVariablesInNamespace(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	envMap := map[string]string{
		"key": "val",
	}
	e.EXPECT().Execute(ctx, []string{
		"set", "env", "deployment",
		"eksa-controller-manager", "key=val",
		"--kubeconfig", cluster.KubeconfigFile,
		"--namespace", "eksa-system",
	})

	err := k.UpdateEnvironmentVariablesInNamespace(ctx, "deployment", "eksa-controller-manager", envMap, cluster, "eksa-system")
	if err != nil {
		t.Fatalf("Kubectl.UpdateEnvironmentVariablesInNamespace() error = %v, want nil", err)
	}
}

func TestKubectlUpdateAnnotation(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{
		"annotate", "gitrepositories", "flux-system",
		"key1=val1", "--overwrite",
		"--kubeconfig", cluster.KubeconfigFile,
	})

	a := map[string]string{
		"key1": "val1",
	}
	err := k.UpdateAnnotation(ctx, "gitrepositories", "flux-system", a, executables.WithOverwrite(), executables.WithCluster(cluster))
	if err != nil {
		t.Fatalf("Kubectl.UpdateAnnotation() error = %v, want nil", err)
	}
}

func TestKubectlRemoveAnnotation(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{
		"annotate", "cluster", "test-cluster", "key1-", "--kubeconfig", cluster.KubeconfigFile,
	})

	err := k.RemoveAnnotation(ctx, "cluster", "test-cluster", "key1", executables.WithCluster(cluster))
	if err != nil {
		t.Fatalf("Kubectl.RemoveAnnotation() error = %v, want nil", err)
	}
}

func TestKubectlRemoveAnnotationInNamespace(t *testing.T) {
	t.Parallel()
	k, ctx, cluster, e := newKubectl(t)
	e.EXPECT().Execute(ctx, []string{
		"annotate", "cluster", "test-cluster", "key1-", "--kubeconfig", cluster.KubeconfigFile, "--namespace", "",
	})

	err := k.RemoveAnnotationInNamespace(ctx, "cluster", "test-cluster", "key1", cluster, "")
	if err != nil {
		t.Fatalf("Kubectl.RemoveAnnotationInNamespace() error = %v, want nil", err)
	}
}

func TestKubectlGetBundles(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	wantBundles := test.Bundles(t)
	bundleName := "Bundle-name"
	bundlesJson, err := json.Marshal(wantBundles)
	if err != nil {
		t.Fatalf("Failed marshalling Bundles: %s", err)
	}

	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "bundles.anywhere.eks.amazonaws.com", bundleName, "-o", "json", "--kubeconfig", tt.cluster.KubeconfigFile, "--namespace", tt.namespace,
	).Return(*bytes.NewBuffer(bundlesJson), nil)

	gotBundles, err := tt.k.GetBundles(tt.ctx, tt.cluster.KubeconfigFile, bundleName, tt.namespace)
	tt.Expect(err).To(BeNil())
	tt.Expect(gotBundles).To(Equal(wantBundles))
}

func TestKubectlGetClusterResourceSet(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	resourceSetJson := test.ReadFile(t, "testdata/kubectl_clusterresourceset.json")
	resourceSetName := "Bundle-name"
	wantResourceSet := &addons.ClusterResourceSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "addons.cluster.x-k8s.io/v1beta1",
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
					Name: "vsphere-cloud-controller-manager",
				},
				{
					Kind: "ConfigMap",
					Name: "vsphere-cloud-controller-manager-role",
				},
			},
		},
	}

	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "--ignore-not-found", "-o", "json", "--kubeconfig", tt.cluster.KubeconfigFile, "clusterresourcesets.addons.cluster.x-k8s.io", "--namespace", tt.namespace, resourceSetName,
	).Return(*bytes.NewBufferString(resourceSetJson), nil)

	gotResourceSet, err := tt.k.GetClusterResourceSet(tt.ctx, tt.cluster.KubeconfigFile, resourceSetName, tt.namespace)
	tt.Expect(err).To(BeNil())
	tt.Expect(gotResourceSet).To(Equal(wantResourceSet))
}

func TestKubectlPauseCAPICluster(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	patch := "{\"spec\":{\"paused\":true}}"

	tt.e.EXPECT().Execute(
		tt.ctx,
		"patch", capiClustersResourceType, "test-cluster", "--type=merge", "-p", patch,
		"--kubeconfig", tt.cluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace,
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.PauseCAPICluster(tt.ctx, "test-cluster", tt.cluster.KubeconfigFile)).To(Succeed())
}

func TestKubectlResumeCAPICluster(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	patch := "{\"spec\":{\"paused\":null}}"

	tt.e.EXPECT().Execute(
		tt.ctx,
		"patch", capiClustersResourceType, "test-cluster", "--type=merge", "-p", patch,
		"--kubeconfig", tt.cluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace,
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.ResumeCAPICluster(tt.ctx, "test-cluster", tt.cluster.KubeconfigFile)).To(Succeed())
}

func TestKubectlMergePatchResource(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	patch := "{\"spec\":{\"paused\":false}}"

	tt.e.EXPECT().Execute(
		tt.ctx,
		"patch", capiClustersResourceType, "test-cluster", "--type=merge", "-p", patch,
		"--kubeconfig", tt.cluster.KubeconfigFile, "--namespace", tt.namespace,
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.MergePatchResource(tt.ctx, capiClustersResourceType, "test-cluster", patch,
		tt.cluster.KubeconfigFile, tt.namespace)).To(Succeed())
}

func TestKubectlMergePatchResourceError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	patch := "{\"spec\":{\"paused\":false}}"

	tt.e.EXPECT().Execute(
		tt.ctx,
		"patch", capiClustersResourceType, "test-cluster", "--type=merge", "-p", patch,
		"--kubeconfig", tt.cluster.KubeconfigFile, "--namespace", tt.namespace,
	).Return(bytes.Buffer{}, errors.New("Error with kubectl merge patch"))

	err := tt.k.MergePatchResource(tt.ctx, capiClustersResourceType, "test-cluster", patch,
		tt.cluster.KubeconfigFile, tt.namespace)
	tt.Expect(err).To(HaveOccurred())
}

func TestKubectlGetConfigMap(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	configmapJson := test.ReadFile(t, "testdata/kubectl_configmap.json")
	configmapName := "foo"
	wantConfigmap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: "eksa-system",
		},
		Data: map[string]string{
			"data": "foo",
		},
	}

	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "configmap", configmapName, "-o", "json", "--kubeconfig", tt.cluster.KubeconfigFile, "--namespace", tt.namespace,
	).Return(*bytes.NewBufferString(configmapJson), nil)

	gotConfigmap, err := tt.k.GetConfigMap(tt.ctx, tt.cluster.KubeconfigFile, configmapName, tt.namespace)
	tt.Expect(err).To(BeNil())
	tt.Expect(gotConfigmap).To(Equal(wantConfigmap))
}

func TestKubectlSetDaemonSetImage(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	daemonSetName := "ds-1"
	container := "cont1"
	image := "public.ecr.aws/image2"

	tt.e.EXPECT().Execute(
		tt.ctx,
		"set", "image", "daemonset/ds-1", "cont1=public.ecr.aws/image2", "--namespace", tt.namespace, "--kubeconfig", tt.cluster.KubeconfigFile,
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.SetDaemonSetImage(tt.ctx, tt.cluster.KubeconfigFile, daemonSetName, tt.namespace, container, image)).To(Succeed())
}

func TestKubectlCheckCAPIProviderExistsNotInstalled(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	providerName := "providerName"
	providerNs := "providerNs"

	tt.e.EXPECT().Execute(tt.ctx,
		[]string{"get", "namespace", fmt.Sprintf("--field-selector=metadata.name=%s", providerNs), "--kubeconfig", tt.cluster.KubeconfigFile}).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.CheckProviderExists(tt.ctx, tt.cluster.KubeconfigFile, providerName, providerNs))
}

func TestKubectlCheckCAPIProviderExistsInstalled(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	providerName := "providerName"
	providerNs := "providerNs"

	tt.e.EXPECT().Execute(tt.ctx,
		[]string{"get", "namespace", fmt.Sprintf("--field-selector=metadata.name=%s", providerNs), "--kubeconfig", tt.cluster.KubeconfigFile}).Return(*bytes.NewBufferString("namespace"), nil)

	tt.e.EXPECT().Execute(tt.ctx,
		[]string{"get", "providers.clusterctl.cluster.x-k8s.io", "--namespace", providerNs, fmt.Sprintf("--field-selector=metadata.name=%s", providerName), "--kubeconfig", tt.cluster.KubeconfigFile})

	tt.Expect(tt.k.CheckProviderExists(tt.ctx, tt.cluster.KubeconfigFile, providerName, providerNs))
}

func TestKubectlGetDeploymentSuccess(t *testing.T) {
	t.Parallel()
	var replicas int32 = 2
	newKubectlGetterTest(t).withResourceType(
		"deployment",
	).withGetter(func(tt *kubectlGetterTest) (client.Object, error) {
		return tt.k.GetDeployment(tt.ctx, tt.name, tt.namespace, tt.kubeconfig)
	}).withJsonFromFile(
		"testdata/kubectl_deployment.json",
	).andWant(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "coredns",
				Namespace: "kube-system",
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "registry.k8s/coredns:1.7.0",
								Name:  "coredns",
							},
						},
					},
				},
			},
		},
	).testSuccess()
}

func TestKubectlGetDeploymentError(t *testing.T) {
	t.Parallel()
	newKubectlGetterTest(t).withResourceType(
		"deployment",
	).withGetter(func(tt *kubectlGetterTest) (client.Object, error) {
		return tt.k.GetDeployment(tt.ctx, tt.name, tt.namespace, tt.kubeconfig)
	}).testError()
}

func TestKubectlGetDaemonSetSuccess(t *testing.T) {
	t.Parallel()
	newKubectlGetterTest(t).withResourceType(
		"daemonset",
	).withGetter(func(tt *kubectlGetterTest) (client.Object, error) {
		return tt.k.GetDaemonSet(tt.ctx, tt.name, tt.namespace, tt.kubeconfig)
	}).withJsonFromFile(
		"testdata/kubectl_daemonset.json",
	).andWant(
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cilium",
				Namespace: "kube-system",
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Command: []string{"cilium-agent"},
								Image:   "public.ecr.aws/isovalent/cilium:v1.9.11-eksa.1",
								Name:    "cilium-agent",
							},
						},
					},
				},
			},
		},
	).testSuccess()
}

func TestKubectlGetDaemonSetError(t *testing.T) {
	t.Parallel()
	newKubectlGetterTest(t).withResourceType(
		"daemonset",
	).withGetter(func(tt *kubectlGetterTest) (client.Object, error) {
		return tt.k.GetDaemonSet(tt.ctx, tt.name, tt.namespace, tt.kubeconfig)
	}).testError()
}

func TestApplyTolerationsFromTaints(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	params := []string{
		"get", "ds", "test",
		"-o", "jsonpath={range .spec.template.spec}{.tolerations} {end}",
		"-n", "testNs", "--kubeconfig", tt.cluster.KubeconfigFile,
	}
	tt.e.EXPECT().Execute(
		tt.ctx, gomock.Eq(params)).Return(bytes.Buffer{}, nil)
	var taints []corev1.Taint
	tt.Expect(tt.k.ApplyTolerationsFromTaints(tt.ctx, taints, taints, "ds", "test", tt.cluster.KubeconfigFile, "testNs", "/test")).To(Succeed())
}

func TestKubectlGetObjectNotFound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		resourceType string
	}{
		{
			name:         "simple resource type",
			resourceType: "cluster",
		},
		{
			name:         "resource type with resource and group",
			resourceType: "cluster.x-k8s.io",
		},
		{
			name:         "resource type with resource, version and group",
			resourceType: "cluster.v1beta1.x-k8s.io",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tt := newKubectlTest(t)
			name := "my-cluster"
			tt.e.EXPECT().Execute(
				tt.ctx,
				"get", "--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, tc.resourceType, "--namespace", tt.namespace, name,
			).Return(bytes.Buffer{}, nil)

			err := tt.k.GetObject(tt.ctx, tc.resourceType, name, tt.namespace, tt.kubeconfig, &clusterv1.Cluster{})
			tt.Expect(err).To(HaveOccurred())
			tt.Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})
	}
}

func TestKubectlGetObjectWithEMptyNamespace(t *testing.T) {
	tt := newKubectlTest(t)
	name := "my-cluster"
	emptyNamespace := ""
	tt.e.EXPECT().Execute(
		tt.ctx,
		// Here we expect the command to have the default namespace set explicitly
		"get", "--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, "cluster", "--namespace", "default", name,
	).Return(bytes.Buffer{}, nil)

	err := tt.k.GetObject(tt.ctx, "cluster", name, emptyNamespace, tt.kubeconfig, &clusterv1.Cluster{})
	tt.Expect(err).To(HaveOccurred())
	tt.Expect(apierrors.IsNotFound(err)).To(BeTrue())
}

func TestKubectlGetAllObjects(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	list := &v1alpha1.ClusterList{}
	b, err := json.Marshal(list)
	tt.Expect(err).To(Succeed())
	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, "clusters", "--all-namespaces",
	).Return(*bytes.NewBuffer(b), nil)

	got := &v1alpha1.ClusterList{}
	tt.Expect(tt.k.Get(tt.ctx, "clusters", tt.kubeconfig, got)).To(Succeed())
	tt.Expect(got).To(BeComparableTo(list))
}

func TestKubectlGetAllObjectsInNamespace(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	list := &v1alpha1.ClusterList{}
	b, err := json.Marshal(list)
	tt.Expect(err).To(Succeed())
	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, "clusters", "--namespace", tt.namespace,
	).Return(*bytes.NewBuffer(b), nil)

	opts := &kubernetes.KubectlGetOptions{
		Namespace: tt.namespace,
	}
	got := &v1alpha1.ClusterList{}
	tt.Expect(tt.k.Get(tt.ctx, "clusters", tt.kubeconfig, got, opts)).To(Succeed())
	tt.Expect(got).To(BeComparableTo(list))
}

func TestKubectlGetSingleObject(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	clusterName := "my-cluster"
	cluster := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
	}
	b, err := json.Marshal(cluster)
	tt.Expect(err).To(Succeed())
	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, "clusters", "--namespace", tt.namespace, clusterName,
	).Return(*bytes.NewBuffer(b), nil)

	opts := &kubernetes.KubectlGetOptions{
		Namespace: tt.namespace,
		Name:      clusterName,
	}
	got := &v1alpha1.Cluster{}
	tt.Expect(tt.k.Get(tt.ctx, "clusters", tt.kubeconfig, got, opts)).To(Succeed())
	tt.Expect(got).To(BeComparableTo(cluster))
}

func TestKubectlGetWithNameAndWithoutNamespace(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	clusterName := "my-cluster"

	opts := &kubernetes.KubectlGetOptions{
		Name: clusterName,
	}

	tt.Expect(tt.k.Get(tt.ctx, "clusters", tt.kubeconfig, &v1alpha1.Cluster{}, opts)).To(
		MatchError(ContainSubstring("if Name is specified, Namespace is required")),
	)
}

func TestKubectlCreateSuccess(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	secret := &corev1.Secret{}
	b, err := yaml.Marshal(secret)
	tt.Expect(err).To(Succeed())

	tt.e.EXPECT().ExecuteWithStdin(
		tt.ctx,
		b,
		"create", "-f", "-", "--kubeconfig", tt.kubeconfig,
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.Create(tt.ctx, tt.kubeconfig, secret)).To(Succeed())
}

func TestKubectlCreateAlreadyExistsError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	secret := &corev1.Secret{}
	b, err := yaml.Marshal(secret)
	tt.Expect(err).To(Succeed())

	tt.e.EXPECT().ExecuteWithStdin(
		tt.ctx,
		b,
		"create", "-f", "-", "--kubeconfig", tt.kubeconfig,
	).Return(
		bytes.Buffer{},
		errors.New("Error from server (AlreadyExists): error when creating \"STDIN\": secret \"my-secret\" already exists\n"), //nolint:revive // The format of the message it's important here since the code checks for its content
	)

	err = tt.k.Create(tt.ctx, tt.kubeconfig, secret)
	tt.Expect(err).To(HaveOccurred())
	tt.Expect(apierrors.IsAlreadyExists(err)).To(BeTrue(), "error should be an AlreadyExists apierror")
}

func TestKubectlReplaceSuccess(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	secret := &corev1.Secret{}
	b, err := yaml.Marshal(secret)
	tt.Expect(err).To(Succeed())

	tt.e.EXPECT().ExecuteWithStdin(
		tt.ctx,
		b,
		"replace", "-f", "-", "--kubeconfig", tt.kubeconfig,
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.Replace(tt.ctx, tt.kubeconfig, secret)).To(Succeed())
}

func TestKubectlReplaceSuccessWithLastAppliedAnnotation(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"kubectl.kubernetes.io/last-applied-configuration": "fake value",
				"my-other-annotation":                              "true",
			},
			Name:      "my-secret",
			Namespace: "my-ns",
		},
	}
	cleanedUpSecret := secret.DeepCopy()
	cleanedUpSecret.Annotations = map[string]string{
		"my-other-annotation": "true",
	}
	b, err := yaml.Marshal(cleanedUpSecret)
	tt.Expect(err).To(Succeed())

	tt.e.EXPECT().ExecuteWithStdin(
		tt.ctx,
		b,
		"replace", "-f", "-", "--kubeconfig", tt.kubeconfig,
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.Replace(tt.ctx, tt.kubeconfig, secret)).To(Succeed())
}

func TestKubectlGetClusterObjectNotFound(t *testing.T) {
	t.Parallel()
	test := newKubectlTest(t)
	name := "my-cluster"
	test.e.EXPECT().Execute(
		test.ctx,
		"get", "--ignore-not-found", "-o", "json", "--kubeconfig", test.kubeconfig, "testresource", name,
	).Return(bytes.Buffer{}, nil)

	err := test.k.GetClusterObject(test.ctx, "testresource", name, test.kubeconfig, &clusterv1.Cluster{})
	test.Expect(err).To(HaveOccurred())
	test.Expect(apierrors.IsNotFound(err)).To(BeTrue())
}

func TestKubectlGetEksaFluxConfig(t *testing.T) {
	t.Parallel()
	kubeconfig := "/my/kubeconfig"
	namespace := "eksa-system"
	eksaFluxConfigResourceType := fmt.Sprintf("fluxconfigs.%s", v1alpha1.GroupVersion.Group)

	returnConfig := &v1alpha1.FluxConfig{}
	returnConfigBytes, err := json.Marshal(returnConfig)
	if err != nil {
		t.Errorf("failed to create output object for test")
	}

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"get", eksaFluxConfigResourceType, "testFluxConfig", "-o", "json", "--kubeconfig", kubeconfig, "--namespace", namespace}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(*bytes.NewBuffer(returnConfigBytes), nil)
	_, err = k.GetEksaFluxConfig(ctx, "testFluxConfig", kubeconfig, namespace)
	if err != nil {
		t.Errorf("Kubectl.GetEksaFluxConfig() error = %v, want error = nil", err)
	}
}

func TestKubectlDeleteFluxConfig(t *testing.T) {
	t.Parallel()
	namespace := "eksa-system"
	kubeconfig := "/my/kubeconfig"
	eksaFluxConfigResourceType := fmt.Sprintf("fluxconfigs.%s", v1alpha1.GroupVersion.Group)

	returnConfig := &v1alpha1.FluxConfig{}
	returnConfigBytes, err := json.Marshal(returnConfig)
	if err != nil {
		t.Errorf("failed to create output object for test")
	}

	mgmtCluster := &types.Cluster{KubeconfigFile: kubeconfig}

	k, ctx, _, e := newKubectl(t)
	expectedParam := []string{"delete", eksaFluxConfigResourceType, "testFluxConfig", "--kubeconfig", mgmtCluster.KubeconfigFile, "--namespace", namespace, "--ignore-not-found=true"}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(*bytes.NewBuffer(returnConfigBytes), nil)
	err = k.DeleteFluxConfig(ctx, mgmtCluster, "testFluxConfig", namespace)
	if err != nil {
		t.Errorf("Kubectl.DeleteFluxConfig() error = %v, want error = nil", err)
	}
}

func TestGetTinkerbellDatacenterConfig(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	datacenterJson := test.ReadFile(t, "testdata/kubectl_tinkerbelldatacenter.json")
	wantDatacenter := &v1alpha1.TinkerbellDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mycluster",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "TinkerbellDatacenterConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		Spec: v1alpha1.TinkerbellDatacenterConfigSpec{
			TinkerbellIP: "1.2.3.4",
		},
	}

	params := []string{
		"get", "tinkerbelldatacenterconfigs.anywhere.eks.amazonaws.com", "mycluster", "-o", "json", "--kubeconfig",
		tt.cluster.KubeconfigFile, "--namespace", tt.namespace,
	}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(*bytes.NewBufferString(datacenterJson), nil)

	got, err := tt.k.GetEksaTinkerbellDatacenterConfig(tt.ctx, "mycluster", tt.cluster.KubeconfigFile, tt.namespace)
	tt.Expect(err).To(BeNil())
	tt.Expect(got).To(Equal(wantDatacenter))
}

func TestGetTinkerbellMachineConfig(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	machineconfigJson := test.ReadFile(t, "testdata/kubectl_tinkerbellmachineconfig.json")
	wantMachineConfig := &v1alpha1.TinkerbellMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mycluster",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "TinkerbellMachineConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		Spec: v1alpha1.TinkerbellMachineConfigSpec{
			OSFamily: "ubuntu",
			TemplateRef: v1alpha1.Ref{
				Name: "mycluster",
				Kind: "TinkerbellTemplateConfig",
			},
		},
	}

	params := []string{
		"get", "tinkerbellmachineconfigs.anywhere.eks.amazonaws.com", "mycluster", "-o", "json", "--kubeconfig",
		tt.cluster.KubeconfigFile, "--namespace", tt.namespace,
	}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(*bytes.NewBufferString(machineconfigJson), nil)

	got, err := tt.k.GetEksaTinkerbellMachineConfig(tt.ctx, "mycluster", tt.cluster.KubeconfigFile, tt.namespace)
	tt.Expect(err).To(BeNil())
	tt.Expect(got).To(Equal(wantMachineConfig))
}

func TestGetTinkerbellMachineConfigInvalid(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	machineconfigJson := test.ReadFile(t, "testdata/kubectl_tinkerbellmachineconfig_invalid.json")

	params := []string{
		"get", "tinkerbellmachineconfigs.anywhere.eks.amazonaws.com", "mycluster", "-o", "json", "--kubeconfig",
		tt.cluster.KubeconfigFile, "--namespace", tt.namespace,
	}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(*bytes.NewBufferString(machineconfigJson), nil)

	_, err := tt.k.GetEksaTinkerbellMachineConfig(tt.ctx, "mycluster", tt.cluster.KubeconfigFile, tt.namespace)
	tt.Expect(err).NotTo(BeNil())
}

func TestGetTinkerbellDatacenterConfigInvalid(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	datacenterconfigJson := test.ReadFile(t, "testdata/kubectl_tinkerbelldatacenter_invalid.json")

	params := []string{
		"get", "tinkerbelldatacenterconfigs.anywhere.eks.amazonaws.com", "mycluster", "-o", "json", "--kubeconfig",
		tt.cluster.KubeconfigFile, "--namespace", tt.namespace,
	}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(*bytes.NewBufferString(datacenterconfigJson), nil)

	_, err := tt.k.GetEksaTinkerbellDatacenterConfig(tt.ctx, "mycluster", tt.cluster.KubeconfigFile, tt.namespace)
	tt.Expect(err).NotTo(BeNil())
}

func TestGetTinkerbellMachineConfigNotFound(t *testing.T) {
	t.Parallel()
	var kubeconfigfile string
	tt := newKubectlTest(t)

	params := []string{
		"get", "tinkerbellmachineconfigs.anywhere.eks.amazonaws.com", "test", "-o", "json", "--kubeconfig",
		kubeconfigfile, "--namespace", tt.namespace,
	}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(*bytes.NewBufferString(""), errors.New("machineconfig not found"))

	_, err := tt.k.GetEksaTinkerbellMachineConfig(tt.ctx, "test", kubeconfigfile, tt.namespace)
	tt.Expect(err).NotTo(BeNil())
}

func TestGetTinkerbellDatacenterConfigNotFound(t *testing.T) {
	t.Parallel()
	var kubeconfigfile string
	tt := newKubectlTest(t)

	params := []string{
		"get", "tinkerbelldatacenterconfigs.anywhere.eks.amazonaws.com", "test", "-o", "json", "--kubeconfig",
		kubeconfigfile, "--namespace", tt.namespace,
	}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(*bytes.NewBufferString(""), errors.New("datacenterconfig not found"))

	_, err := tt.k.GetEksaTinkerbellDatacenterConfig(tt.ctx, "test", kubeconfigfile, tt.namespace)
	tt.Expect(err).NotTo(BeNil())
}

func TestGetUnprovisionedTinkerbellHardware(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	hardwareJSON := test.ReadFile(t, "testdata/kubectl_tinkerbellhardware.json")
	kubeconfig := "foo/bar"

	var expect []tinkv1alpha1.Hardware
	for _, h := range []string{"hw1", "hw2"} {
		expect = append(expect, tinkv1alpha1.Hardware{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Hardware",
				APIVersion: "tinkerbell.org/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: h,
			},
		})
	}

	params := []string{
		"get", executables.TinkerbellHardwareResourceType,
		"-l", "!v1alpha1.tinkerbell.org/ownerName",
		"--kubeconfig", kubeconfig,
		"-o", "json",
		"--namespace", tt.namespace,
	}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(*bytes.NewBufferString(hardwareJSON), nil)

	hardware, err := tt.k.GetUnprovisionedTinkerbellHardware(tt.ctx, kubeconfig, tt.namespace)
	tt.Expect(err).To(Succeed())
	tt.Expect(hardware).To(Equal(expect))
}

func TestGetUnprovisionedTinkerbellHardware_MarshallingError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	kubeconfig := "foo/bar"
	var buf bytes.Buffer

	params := []string{
		"get", executables.TinkerbellHardwareResourceType,
		"-l", "!v1alpha1.tinkerbell.org/ownerName",
		"--kubeconfig", kubeconfig,
		"-o", "json",
		"--namespace", tt.namespace,
	}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(buf, nil)

	_, err := tt.k.GetUnprovisionedTinkerbellHardware(tt.ctx, kubeconfig, tt.namespace)
	tt.Expect(err).NotTo(BeNil())
}

func TestGetVsphereMachine(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	vsphereMachineJSON := test.ReadFile(t, "testdata/kubectl_vspheremachine.json")
	expect := []vspherev1.VSphereMachine{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "VSphereMachine",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testcluster-control-plane-template-1234-1234",
				Namespace: constants.EksaSystemNamespace,
				Labels: map[string]string{
					"cluster.x-k8s.io/control-plane": "",
				},
			},
			Spec: vspherev1.VSphereMachineSpec{
				VirtualMachineCloneSpec: vspherev1.VirtualMachineCloneSpec{
					Template: "test-template",
				},
			},
		},
	}

	selector := framework.ControlPlaneMachineLabel
	params := []string{"get", "vspheremachines", "-o", "json", "--namespace", constants.EksaSystemNamespace, "--kubeconfig", tt.cluster.KubeconfigFile, "--selector=" + selector}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(*bytes.NewBufferString(vsphereMachineJSON), nil)

	got, err := tt.k.GetVsphereMachine(tt.ctx, tt.cluster.KubeconfigFile, selector)
	tt.Expect(err).To(Succeed())
	tt.Expect(got).To(Equal(expect))
}

func TestGetVsphereMachineFail(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)

	selector := framework.ControlPlaneMachineLabel
	params := []string{"get", "vspheremachines", "-o", "json", "--namespace", constants.EksaSystemNamespace, "--kubeconfig", tt.cluster.KubeconfigFile, "--selector=" + selector}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(bytes.Buffer{}, errors.New("error"))

	_, err := tt.k.GetVsphereMachine(tt.ctx, tt.cluster.KubeconfigFile, selector)
	tt.Expect(err).To(MatchError(ContainSubstring("getting VSphere machine: error")))
}

func TestGetUnprovisionedTinkerbellHardware_ExecutableErrors(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	kubeconfig := "foo/bar"
	var buf bytes.Buffer
	expect := errors.New("foo bar")

	params := []string{
		"get", executables.TinkerbellHardwareResourceType,
		"-l", "!v1alpha1.tinkerbell.org/ownerName",
		"--kubeconfig", kubeconfig,
		"-o", "json",
		"--namespace", tt.namespace,
	}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(buf, expect)

	_, err := tt.k.GetUnprovisionedTinkerbellHardware(tt.ctx, kubeconfig, tt.namespace)
	tt.Expect(err).NotTo(BeNil())
}

func TestKubectlDeleteSingleObject(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	name := "my-cluster"
	resourceType := "cluster.x-k8s.io"
	tt.e.EXPECT().Execute(
		tt.ctx,
		"delete", "--kubeconfig", tt.kubeconfig, resourceType, name, "--namespace", tt.namespace,
	).Return(bytes.Buffer{}, nil)

	opts := &kubernetes.KubectlDeleteOptions{
		Name:      name,
		Namespace: tt.namespace,
	}
	tt.Expect(tt.k.Delete(tt.ctx, resourceType, tt.kubeconfig, opts)).To(Succeed())
}

func TestKubectlDeleteAllObjectsInNamespace(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	resourceType := "cluster.x-k8s.io"
	tt.e.EXPECT().Execute(
		tt.ctx,
		"delete", "--kubeconfig", tt.kubeconfig, resourceType, "--all", "--namespace", tt.namespace,
	).Return(bytes.Buffer{}, nil)

	opts := &kubernetes.KubectlDeleteOptions{
		Namespace: tt.namespace,
	}
	tt.Expect(tt.k.Delete(tt.ctx, resourceType, tt.kubeconfig, opts)).To(Succeed())
}

func TestKubectlDeleteAllObjectsInNamespaceWithLabels(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	resourceType := "cluster.x-k8s.io"
	tt.e.EXPECT().Execute(
		tt.ctx,
		"delete",
		"--kubeconfig", tt.kubeconfig,
		resourceType,
		"--namespace", tt.namespace,
		"--selector", "label1=val1,label2=val2",
	).Return(bytes.Buffer{}, nil)

	opts := &kubernetes.KubectlDeleteOptions{
		Namespace: tt.namespace,
		HasLabels: map[string]string{
			"label2": "val2",
			"label1": "val1",
		},
	}
	tt.Expect(tt.k.Delete(tt.ctx, resourceType, tt.kubeconfig, opts)).To(Succeed())
}

func TestKubectlDeleteAllObjectsInAllNamespaces(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	resourceType := "cluster.x-k8s.io"
	tt.e.EXPECT().Execute(
		tt.ctx,
		"delete", "--kubeconfig", tt.kubeconfig, resourceType, "--all", "--all-namespaces",
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.Delete(tt.ctx, resourceType, tt.kubeconfig)).To(Succeed())
}

func TestKubectlDeleteObjectWithNoNamespace(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	resourceType := "cluster.x-k8s.io"
	clusterName := "my-cluster"

	opts := &kubernetes.KubectlDeleteOptions{
		Name: clusterName,
	}

	tt.Expect(tt.k.Delete(tt.ctx, resourceType, tt.kubeconfig, opts)).To(
		MatchError(ContainSubstring("if Name is specified, Namespace is required")),
	)
}

func TestKubectlDeleteObjectWithHasLabels(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	resourceType := "cluster.x-k8s.io"
	clusterName := "my-cluster"

	opts := &kubernetes.KubectlDeleteOptions{
		Name:      clusterName,
		Namespace: tt.namespace,
		HasLabels: map[string]string{
			"mylabel": "value",
		},
	}

	tt.Expect(tt.k.Delete(tt.ctx, resourceType, tt.kubeconfig, opts)).To(
		MatchError(ContainSubstring("options for HasLabels and Name are mutually exclusive")),
	)
}

func TestKubectlDeleteObjectNotFoundError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	name := "my-cluster"
	resourceType := "cluster.x-k8s.io"
	tt.e.EXPECT().Execute(
		tt.ctx,
		"delete", "--kubeconfig", tt.kubeconfig, resourceType, name, "--namespace", tt.namespace,
	).Return(
		bytes.Buffer{},
		errors.New("Error from server (NotFound): cluster \"my-cluster\" not found\n"), //nolint:revive // The format of the message it's important here since the code checks for its content
	)

	opts := &kubernetes.KubectlDeleteOptions{
		Name:      name,
		Namespace: tt.namespace,
	}

	err := tt.k.Delete(tt.ctx, resourceType, tt.kubeconfig, opts)
	tt.Expect(err).To(HaveOccurred())
	tt.Expect(apierrors.IsNotFound(err)).To(BeTrue(), "error should be a NotFound apierror")
}

func TestKubectlDeleteClusterObject(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	name := "my-storageclass"
	resourceType := "storageclass"
	tt.e.EXPECT().Execute(
		tt.ctx,
		"delete", resourceType, name, "--kubeconfig", tt.kubeconfig,
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.DeleteClusterObject(tt.ctx, resourceType, name, tt.kubeconfig)).To(Succeed())
}

func TestKubectlDeleteClusterObjectError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	name := "my-storageclass"
	resourceType := "storageclass"
	tt.e.EXPECT().Execute(
		tt.ctx,
		"delete", resourceType, name, "--kubeconfig", tt.kubeconfig,
	).Return(bytes.Buffer{}, errors.New("test error"))

	tt.Expect(tt.k.DeleteClusterObject(tt.ctx, resourceType, name, tt.kubeconfig)).NotTo(Succeed())
}

func TestKubectlWaitForManagedExternalEtcdNotReady(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	timeout := "5m"
	expectedTimeout := "300.00s"

	tt.e.EXPECT().Execute(
		tt.ctx,
		"wait", "--timeout", expectedTimeout, "--for=condition=ManagedEtcdReady=false", "clusters.cluster.x-k8s.io/test", "--kubeconfig", tt.cluster.KubeconfigFile, "-n", "eksa-system",
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.WaitForManagedExternalEtcdNotReady(tt.ctx, tt.cluster, timeout, "test")).To(Succeed())
}

func TestKubectlWaitForMachineDeploymentReady(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	timeout := "5m"
	expectedTimeout := "300.00s"

	tt.e.EXPECT().Execute(
		tt.ctx,
		"wait", "--timeout", expectedTimeout, "--for=condition=Ready=true", "machinedeployments.cluster.x-k8s.io/test", "--kubeconfig", tt.cluster.KubeconfigFile, "-n", "eksa-system",
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.WaitForMachineDeploymentReady(tt.ctx, tt.cluster, timeout, "test")).To(Succeed())
}

func TestKubectlWaitForClusterReady(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)

	timeout := "5m"
	expectedTimeout := "300.00s"

	tt.e.EXPECT().Execute(
		tt.ctx,
		"wait", "--timeout", expectedTimeout, "--for=condition=Ready", "clusters.cluster.x-k8s.io/test", "--kubeconfig", tt.cluster.KubeconfigFile, "-n", "eksa-system",
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.WaitForClusterReady(tt.ctx, tt.cluster, timeout, "test")).To(Succeed())
}

func TestWaitForRufioMachines(t *testing.T) {
	t.Parallel()
	kt := newKubectlTest(t)

	timeout := "5m"
	expectedTimeout := "300.00s"

	kt.e.EXPECT().Execute(
		kt.ctx,
		"wait", "--timeout", expectedTimeout, "--for=condition=Contactable", "machines.bmc.tinkerbell.org", "--kubeconfig", kt.cluster.KubeconfigFile, "-n", "eksa-system", "--all",
	).Return(bytes.Buffer{}, nil)

	kt.Expect(kt.k.WaitForRufioMachines(kt.ctx, kt.cluster, timeout, "Contactable", "eksa-system")).To(Succeed())
}

func TestKubectlApply(t *testing.T) {
	kubeconfig := "myconfig.kubeconfig"
	testCases := []struct {
		name                  string
		opts                  []kubernetes.KubectlApplyOption
		expectedCommandParams []string
	}{
		{
			name:                  "no opts",
			expectedCommandParams: []string{"apply", "-f", "-", "--kubeconfig", kubeconfig},
		},
		{
			name: "with field manager",
			opts: []kubernetes.KubectlApplyOption{
				kubernetes.KubectlApplyOptions{
					FieldManager: "my-manager",
				},
			},
			expectedCommandParams: []string{"apply", "-f", "-", "--kubeconfig", kubeconfig, "--field-manager", "my-manager"},
		},
		{
			name: "with server side and force ownership",
			opts: []kubernetes.KubectlApplyOption{
				kubernetes.KubectlApplyOptions{
					ServerSide:     true,
					ForceOwnership: true,
				},
			},
			expectedCommandParams: []string{"apply", "-f", "-", "--kubeconfig", kubeconfig, "--server-side", "--force-conflicts"},
		},
		{
			name: "with server side and field manager",
			opts: []kubernetes.KubectlApplyOption{
				kubernetes.KubectlApplyOptions{
					ServerSide:   true,
					FieldManager: "my-manager",
				},
			},
			expectedCommandParams: []string{"apply", "-f", "-", "--kubeconfig", kubeconfig, "--field-manager", "my-manager", "--server-side"},
		},
	}
	for _, tCase := range testCases {
		tc := tCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tt := newKubectlTest(t)
			tt.kubeconfig = kubeconfig
			secret := &corev1.Secret{}
			b, err := yaml.Marshal(secret)
			tt.Expect(err).To(Succeed())

			tt.e.EXPECT().ExecuteWithStdin(
				tt.ctx,
				b,
				tc.expectedCommandParams,
			).Return(bytes.Buffer{}, nil)

			tt.Expect(tt.k.Apply(tt.ctx, tt.kubeconfig, secret, tc.opts...)).To(Succeed())
		})
	}
}

func TestKubectlListObjects(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	list := &v1alpha1.ClusterList{}
	b, err := json.Marshal(list)
	tt.Expect(err).To(Succeed())
	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, "clusters", "--namespace", tt.namespace,
	).Return(*bytes.NewBuffer(b), nil)

	tt.Expect(tt.k.ListObjects(tt.ctx, "clusters", tt.namespace, tt.kubeconfig, &v1alpha1.ClusterList{})).To(Succeed())
}

func TestKubectlListObjectsExecError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, "clusters", "--namespace", tt.namespace,
	).Return(bytes.Buffer{}, errors.New("error"))

	tt.Expect(tt.k.ListObjects(tt.ctx, "clusters", tt.namespace, tt.kubeconfig, &v1alpha1.ClusterList{})).To(MatchError(ContainSubstring("getting clusters with kubectl: error")))
}

func TestKubectlListObjectsMarshalError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, "clusters", "--namespace", tt.namespace,
	).Return(*bytes.NewBufferString("//"), nil)

	tt.Expect(tt.k.ListObjects(tt.ctx, "clusters", tt.namespace, tt.kubeconfig, &v1alpha1.ClusterList{})).To(MatchError(ContainSubstring("parsing get clusters response")))
}

func TestKubectlHasResource(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	pbc := &packagesv1.PackageBundleController{
		TypeMeta: metav1.TypeMeta{Kind: "packageBundleController"},
		Spec: packagesv1.PackageBundleControllerSpec{
			ActiveBundle: "some bundle",
		},
	}
	b, err := json.Marshal(pbc)
	tt.Expect(err).To(Succeed())
	tt.e.EXPECT().Execute(tt.ctx,
		"get", "--ignore-not-found", "-o", "json",
		"--kubeconfig", tt.kubeconfig, "packageBundleController", "--namespace", tt.namespace, "testResourceName",
	).Return(*bytes.NewBuffer(b), nil)

	has, err := tt.k.HasResource(tt.ctx, "packageBundleController", "testResourceName", tt.kubeconfig, tt.namespace)
	tt.Expect(err).To(Succeed())
	tt.Expect(has).To(BeTrue())
}

func TestKubectlHasResourceWithGetError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(tt.ctx,
		"get", "--ignore-not-found", "-o", "json",
		"--kubeconfig", tt.kubeconfig, "packageBundleController", "--namespace", tt.namespace, "testResourceName",
	).Return(bytes.Buffer{}, fmt.Errorf("test error"))

	has, err := tt.k.HasResource(tt.ctx, "packageBundleController", "testResourceName", tt.kubeconfig, tt.namespace)
	tt.Expect(err).To(MatchError(ContainSubstring("test error")))
	tt.Expect(has).To(BeFalse())
}

func TestKubectlDeletePackageResources(t *testing.T) {
	t.Parallel()

	t.Run("golden path", func(t *testing.T) {
		tt := newKubectlTest(t)
		tt.e.EXPECT().Execute(
			tt.ctx,
			"delete", "packagebundlecontroller.packages.eks.amazonaws.com", "clusterName", "--kubeconfig", tt.kubeconfig, "--namespace", "eksa-packages", "--ignore-not-found=true",
		).Return(*bytes.NewBufferString("//"), nil)
		tt.e.EXPECT().Execute(
			tt.ctx,
			"delete", "namespace", "eksa-packages-clusterName", "--kubeconfig", tt.kubeconfig, "--ignore-not-found=true",
		).Return(*bytes.NewBufferString("//"), nil)

		tt.Expect(tt.k.DeletePackageResources(tt.ctx, tt.cluster, "clusterName")).To(Succeed())
	})

	t.Run("pbc failure", func(t *testing.T) {
		tt := newKubectlTest(t)
		tt.e.EXPECT().Execute(
			tt.ctx,
			"delete", "packagebundlecontroller.packages.eks.amazonaws.com", "clusterName", "--kubeconfig", tt.kubeconfig, "--namespace", "eksa-packages", "--ignore-not-found=true",
		).Return(*bytes.NewBufferString("//"), fmt.Errorf("bam"))

		tt.Expect(tt.k.DeletePackageResources(tt.ctx, tt.cluster, "clusterName")).To(MatchError(ContainSubstring("bam")))
	})

	t.Run("namespace failure", func(t *testing.T) {
		tt := newKubectlTest(t)
		tt.e.EXPECT().Execute(
			tt.ctx,
			"delete", "packagebundlecontroller.packages.eks.amazonaws.com", "clusterName", "--kubeconfig", tt.kubeconfig, "--namespace", "eksa-packages", "--ignore-not-found=true",
		).Return(*bytes.NewBufferString("//"), nil)
		tt.e.EXPECT().Execute(
			tt.ctx,
			"delete", "namespace", "eksa-packages-clusterName", "--kubeconfig", tt.kubeconfig, "--ignore-not-found=true",
		).Return(*bytes.NewBufferString("//"), fmt.Errorf("boom"))

		tt.Expect(tt.k.DeletePackageResources(tt.ctx, tt.cluster, "clusterName")).To(MatchError(ContainSubstring("boom")))
	})
}

func TestKubectlExecuteFromYaml(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().ExecuteWithStdin(
		tt.ctx, []byte(nutanixMachineConfigSpec), "apply", "--kubeconfig", tt.kubeconfig, "--namespace", tt.namespace,
	).Return(bytes.Buffer{}, nil)
	_, err := tt.k.ExecuteFromYaml(tt.ctx, []byte(nutanixMachineConfigSpec), "apply", "--kubeconfig", tt.kubeconfig, "--namespace", tt.namespace)
	tt.Expect(err).ToNot(HaveOccurred())
}

func TestKubectlSearchNutanixMachineConfig(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "nutanixmachineconfigs.anywhere.eks.amazonaws.com", "-o", "json", "--kubeconfig", tt.kubeconfig, "--namespace", tt.namespace, "--field-selector=metadata.name=eksa-unit-test",
	).Return(*bytes.NewBufferString(nutanixMachineConfigsJSON), nil)
	items, err := tt.k.SearchNutanixMachineConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)
	tt.Expect(err).ToNot(HaveOccurred())
	tt.Expect(items).To(HaveLen(1))
}

func TestKubectlSearchNutanixMachineConfigError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "nutanixmachineconfigs.anywhere.eks.amazonaws.com", "-o", "json", "--kubeconfig", tt.kubeconfig, "--namespace", tt.namespace, "--field-selector=metadata.name=eksa-unit-test",
	).Return(bytes.Buffer{}, fmt.Errorf("error"))
	items, err := tt.k.SearchNutanixMachineConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)
	tt.Expect(err).To(HaveOccurred())
	tt.Expect(items).To(BeNil())
}

func TestKubectlSearchNutanixDatacenterConfig(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", "-o", "json", "--kubeconfig", tt.kubeconfig, "--namespace", tt.namespace, "--field-selector=metadata.name=eksa-unit-test",
	).Return(*bytes.NewBufferString(nutanixDatacenterConfigsJSON), nil)
	items, err := tt.k.SearchNutanixDatacenterConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)
	tt.Expect(err).ToNot(HaveOccurred())
	tt.Expect(items).To(HaveLen(1))
}

func TestKubectlSearchNutanixDatacenterConfigError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(
		tt.ctx,
		"get", "nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", "-o", "json", "--kubeconfig", tt.kubeconfig, "--namespace", tt.namespace, "--field-selector=metadata.name=eksa-unit-test",
	).Return(bytes.Buffer{}, fmt.Errorf("error"))
	items, err := tt.k.SearchNutanixDatacenterConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)
	tt.Expect(err).To(HaveOccurred())
	tt.Expect(items).To(BeNil())
}

func TestKubectlGetEksaNutanixMachineConfig(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(gomock.Any(), "get",
		[]string{"--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, "nutanixmachineconfigs.anywhere.eks.amazonaws.com", "--namespace", tt.namespace, "eksa-unit-test"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(*bytes.NewBufferString(nutanixMachineConfigSpecJSON), nil)
	item, err := tt.k.GetEksaNutanixMachineConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)
	tt.Expect(err).ToNot(HaveOccurred())
	tt.Expect(item).ToNot(BeNil())
}

func TestKubectlGetEksaNutanixMachineConfigError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(gomock.Any(), "get",
		[]string{"--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, "nutanixmachineconfigs.anywhere.eks.amazonaws.com", "--namespace", tt.namespace, "eksa-unit-test"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(bytes.Buffer{}, fmt.Errorf("error"))
	item, err := tt.k.GetEksaNutanixMachineConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)
	tt.Expect(err).To(HaveOccurred())
	tt.Expect(item).To(BeNil())
}

func TestKubectlGetEksaNutanixDatacenterConfig(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(gomock.Any(), "get",
		[]string{"--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, "nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", "--namespace", tt.namespace, "eksa-unit-test"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(*bytes.NewBufferString(nutanixDatacenterConfigSpecJSON), nil)
	item, err := tt.k.GetEksaNutanixDatacenterConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)
	tt.Expect(err).ToNot(HaveOccurred())
	tt.Expect(item).ToNot(BeNil())
}

func TestKubectlGetEksaNutanixDatacenterConfigError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(gomock.Any(), "get",
		[]string{"--ignore-not-found", "-o", "json", "--kubeconfig", tt.kubeconfig, "nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", "--namespace", tt.namespace, "eksa-unit-test"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(bytes.Buffer{}, fmt.Errorf("error"))
	item, err := tt.k.GetEksaNutanixDatacenterConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)
	tt.Expect(err).To(HaveOccurred())
	tt.Expect(item).To(BeNil())
}

func TestKubectlDeleteEksaNutanixDatacenterConfig(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(gomock.Any(), "delete",
		[]string{"nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", "eksa-unit-test", "--kubeconfig", tt.kubeconfig, "--namespace", tt.namespace, "--ignore-not-found=true"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(bytes.Buffer{}, nil)
	tt.Expect(tt.k.DeleteEksaNutanixDatacenterConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)).To(Succeed())
}

func TestKubectlDeleteEksaNutanixDatacenterConfigError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(gomock.Any(), "delete",
		[]string{"nutanixdatacenterconfigs.anywhere.eks.amazonaws.com", "eksa-unit-test", "--kubeconfig", tt.kubeconfig, "--namespace", tt.namespace, "--ignore-not-found=true"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(bytes.Buffer{}, fmt.Errorf("error"))
	tt.Expect(tt.k.DeleteEksaNutanixDatacenterConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)).NotTo(Succeed())
}

func TestKubectlDeleteEksaNutanixMachineConfig(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(gomock.Any(), "delete",
		[]string{"nutanixmachineconfigs.anywhere.eks.amazonaws.com", "eksa-unit-test", "--kubeconfig", tt.kubeconfig, "--namespace", tt.namespace, "--ignore-not-found=true"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(bytes.Buffer{}, nil)
	tt.Expect(tt.k.DeleteEksaNutanixMachineConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)).To(Succeed())
}

func TestKubectlDeleteEksaNutanixMachineConfigError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	tt.e.EXPECT().Execute(gomock.Any(), "delete",
		[]string{"nutanixmachineconfigs.anywhere.eks.amazonaws.com", "eksa-unit-test", "--kubeconfig", tt.kubeconfig, "--namespace", tt.namespace, "--ignore-not-found=true"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(bytes.Buffer{}, fmt.Errorf("error"))
	tt.Expect(tt.k.DeleteEksaNutanixMachineConfig(tt.ctx, "eksa-unit-test", tt.kubeconfig, tt.namespace)).NotTo(Succeed())
}

func TestWaitForResourceRolledout(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	timeout := "2m"
	target := "testdaemonset"
	var b bytes.Buffer
	expectedParam := []string{"rollout", "status", "daemonset", target, "--kubeconfig", tt.kubeconfig, "--namespace", "eksa-system", "--timeout", timeout}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	tt.Expect(tt.k.WaitForResourceRolledout(tt.ctx, tt.cluster, timeout, target, "eksa-system", "daemonset")).To(Succeed())
}

func TestWaitForDaemonsetRolledoutError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	timeout := "2m"
	target := "testdaemonset"
	var b bytes.Buffer
	expectedParam := []string{"rollout", "status", "daemonset", target, "--kubeconfig", tt.kubeconfig, "--namespace", "eksa-system", "--timeout", timeout}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, fmt.Errorf("error")).AnyTimes()
	tt.Expect(tt.k.WaitForResourceRolledout(tt.ctx, tt.cluster, timeout, target, "eksa-system", "daemonset")).NotTo(Succeed())
}

func TestWaitForPod(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	timeout := "2m"
	expectedTimeout := "120.00s"
	condition := "status.containerStatuses[0].state.terminated.reason"
	target := "testpod"
	tt.e.EXPECT().Execute(gomock.Any(), "wait", "--timeout",
		[]string{expectedTimeout, fmt.Sprintf("%s=%s", "--for=condition", condition), fmt.Sprintf("%s/%s", "pod", target), "--kubeconfig", tt.kubeconfig, "-n", "eksa-system"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.WaitForPod(tt.ctx, tt.cluster, timeout, condition, target, "eksa-system")).To(Succeed())
}

func TestWaitForJob(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	timeout := "2m"
	expectedTimeout := "120.00s"
	condition := "status.containerStatuses[0].state.terminated.reason"
	target := "testpod"
	tt.e.EXPECT().Execute(gomock.Any(), "wait", "--timeout",
		[]string{expectedTimeout, fmt.Sprintf("%s=%s", "--for=condition", condition), fmt.Sprintf("%s/%s", "job", target), "--kubeconfig", tt.kubeconfig, "-n", "eksa-system"},
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(bytes.Buffer{}, nil)

	tt.Expect(tt.k.WaitForJobCompleted(tt.ctx, tt.cluster.KubeconfigFile, timeout, condition, target, "eksa-system")).To(Succeed())
}

func TestWaitForPodCompleted(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.WriteString("'Completed'")
	tt := newKubectlTest(t)
	expectedParam := []string{"get", "pod/testpod", "-o", fmt.Sprintf("%s=%s", "jsonpath", "'{.status.containerStatuses[0].state.terminated.reason}'"), "--kubeconfig", "c.kubeconfig", "-n", "eksa-system"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()

	tt.Expect(tt.k.WaitForPodCompleted(tt.ctx, tt.cluster, "testpod", "2m", "eksa-system")).To(Succeed())
}

func TestWaitForPackagesInstalled(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.WriteString("'installed'")
	tt := newKubectlTest(t)
	expectedParam := []string{"get", "packages.packages.eks.amazonaws.com/testpackage", "-o", fmt.Sprintf("%s=%s", "jsonpath", "'{.status.state}'"), "--kubeconfig", "c.kubeconfig", "-n", "eksa-system"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	err := tt.k.WaitForPackagesInstalled(tt.ctx, tt.cluster, "testpackage", "2m", "eksa-system")
	tt.Expect(err).ToNot(HaveOccurred())
}

func TestWaitJSONPathLoop(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.WriteString("'installed'")
	tt := newKubectlTest(t)
	expectedParam := []string{"get", "packages.packages.eks.amazonaws.com/testpackage", "-o", fmt.Sprintf("%s=%s", "jsonpath", "'{.status.state}'"), "--kubeconfig", "c.kubeconfig", "-n", "eksa-system"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	err := tt.k.WaitJSONPathLoop(tt.ctx, tt.cluster.KubeconfigFile, "2m", "status.state", "installed", "packages.packages.eks.amazonaws.com/testpackage", "eksa-system")
	tt.Expect(err).ToNot(HaveOccurred())
	tt.Expect(b.String()).ToNot(ContainSubstring("waiting 5 seconds...."))
}

func TestWaitJSONPathLoopTimeParseError(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.WriteString("'installed'")
	tt := newKubectlTest(t)
	expectedParam := []string{"get", "packages.packages.eks.amazonaws.com/testpackage", "-o", fmt.Sprintf("%s=%s", "jsonpath", "'{.status.state}'"), "--kubeconfig", "c.kubeconfig", "-n", "eksa-system"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	tt.Expect(tt.k.WaitJSONPathLoop(tt.ctx, tt.cluster.KubeconfigFile, "", "status.state", "installed", "packages.packages.eks.amazonaws.com/testpackage", "eksa-system")).To(MatchError(ContainSubstring("unparsable timeout specified")))
}

func TestWaitJSONPathLoopNegativeTimeError(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.WriteString("'installed'")
	tt := newKubectlTest(t)
	expectedParam := []string{"get", "packages.packages.eks.amazonaws.com/testpackage", "-o", fmt.Sprintf("%s=%s", "jsonpath", "'{.status.state}'"), "--kubeconfig", "c.kubeconfig", "-n", "eksa-system"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	tt.Expect(tt.k.WaitJSONPathLoop(tt.ctx, tt.cluster.KubeconfigFile, "-1m", "status.state", "installed", "packages.packages.eks.amazonaws.com/testpackage", "eksa-system")).To(MatchError(ContainSubstring("negative timeout specified")))
}

func TestWaitJSONPathLoopRetrierError(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.WriteString("")
	tt := newKubectlTest(t)
	expectedParam := []string{"get", "packages.packages.eks.amazonaws.com/testpackage", "-o", fmt.Sprintf("%s=%s", "jsonpath", "'{.}'"), "--kubeconfig", "c.kubeconfig", "-n", "eksa-system"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	tt.Expect(tt.k.WaitJSONPathLoop(tt.ctx, tt.cluster.KubeconfigFile, "2m", "", "", "packages.packages.eks.amazonaws.com/testpackage", "eksa-system")).To(MatchError(ContainSubstring("executing wait")))
}

func TestWaitJSONPath(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.WriteString("'installed'")
	tt := newKubectlTest(t)
	expectedParam := []string{"wait", "--timeout", "2m", fmt.Sprintf("--for=jsonpath='{.%s}'=%s", "status.state", "installed"), "packages.packages.eks.amazonaws.com/testpackage", "--kubeconfig", "c.kubeconfig", "-n", "eksa-system"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	err := tt.k.WaitJSONPath(tt.ctx, tt.cluster.KubeconfigFile, "2m", "status.state", "installed", "packages.packages.eks.amazonaws.com/testpackage", "eksa-system")
	tt.Expect(err).To(BeNil())
}

func TestWaitJSONPathTimeParseError(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.WriteString("'installed'")
	tt := newKubectlTest(t)
	expectedParam := []string{"wait", "--timeout", "2m", fmt.Sprintf("--for=jsonpath='{.%s}'=%s", "status.state", "installed"), "packages.packages.eks.amazonaws.com/testpackage", "--kubeconfig", "c.kubeconfig", "-n", "eksa-system"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	tt.Expect(tt.k.WaitJSONPath(tt.ctx, tt.cluster.KubeconfigFile, "", "status.state", "installed", "packages.packages.eks.amazonaws.com/testpackage", "eksa-system")).To(MatchError(ContainSubstring("unparsable timeout specified")))
}

func TestWaitJSONPathNegativeTimeError(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.WriteString("'installed'")
	tt := newKubectlTest(t)
	expectedParam := []string{"wait", "--timeout", "2m", fmt.Sprintf("--for=jsonpath='{.%s}'=%s", "status.state", "installed"), "packages.packages.eks.amazonaws.com/testpackage", "--kubeconfig", "c.kubeconfig", "-n", "eksa-system"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	tt.Expect(tt.k.WaitJSONPath(tt.ctx, tt.cluster.KubeconfigFile, "-1m", "status.state", "installed", "packages.packages.eks.amazonaws.com/testpackage", "eksa-system")).To(MatchError(ContainSubstring("negative timeout specified")))
}

func TestWaitJSONPathRetrierError(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.WriteString("'installed'")
	tt := newKubectlTest(t)
	expectedParam := []string{"wait", "--timeout", "2m", fmt.Sprintf("--for=jsonpath='{.%s}'=%s", "", ""), "packages.packages.eks.amazonaws.com/testpackage", "--kubeconfig", "c.kubeconfig", "-n", "eksa-system"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	tt.Expect(tt.k.WaitJSONPath(tt.ctx, tt.cluster.KubeconfigFile, "2m", "", "", "packages.packages.eks.amazonaws.com/testpackage", "eksa-system")).To(MatchError(ContainSubstring("executing wait")))
}

func TestGetPackageBundleController(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	testPbc := &packagesv1.PackageBundleController{}
	respJSON, err := json.Marshal(testPbc)
	if err != nil {
		t.Errorf("marshaling test service: %s", err)
	}
	ret := bytes.NewBuffer(respJSON)
	expectedParam := []string{"get", "packagebundlecontroller.packages.eks.amazonaws.com", "testcluster", "-o", "json", "--kubeconfig", "c.kubeconfig", "--namespace", "eksa-packages", "--ignore-not-found=true"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(*ret, nil).AnyTimes()
	if _, err := tt.k.GetPackageBundleController(tt.ctx, tt.cluster.KubeconfigFile, "testcluster"); err != nil {
		t.Errorf("Kubectl.GetPackageBundleController() error = %v, want nil", err)
	}
}

func TestGetPackageBundleList(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	testPbc := &packagesv1.PackageBundleList{
		TypeMeta: metav1.TypeMeta{
			Kind: "PackageBundle",
		},
	}
	respJSON, err := json.Marshal(testPbc)
	if err != nil {
		t.Errorf("marshaling test service: %s", err)
	}
	ret := bytes.NewBuffer(respJSON)
	expectedParam := []string{"get", "packagebundles.packages.eks.amazonaws.com", "-o", "jsonpath='{.items}'", "--kubeconfig", "c.kubeconfig", "-n", "eksa-packages"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(*ret, nil).AnyTimes()
	expectedParam = []string{"get", "packagebundles.packages.eks.amazonaws.com", "-o", "json", "--kubeconfig", "c.kubeconfig", "--namespace", "eksa-packages", "--ignore-not-found=true"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(*ret, nil).AnyTimes()
	if _, err := tt.k.GetPackageBundleList(tt.ctx, tt.cluster.KubeconfigFile); err != nil {
		t.Errorf("Kubectl.GetPackageBundleList() error = %v, want nil", err)
	}
}

func TestRunCurlPod(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	var b bytes.Buffer
	expectedParam := []string{"run", "testpod-123", "--image=public.ecr.aws/eks-anywhere/diagnostic-collector:v0.16.2-eks-a-41", "-o", "json", "--kubeconfig", "c.kubeconfig", "--namespace", "eksa-packages", "--restart=Never", "--", "pwd"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	if _, err := tt.k.RunCurlPod(tt.ctx, "eksa-packages", "testpod-123", tt.cluster.KubeconfigFile, []string{"pwd"}); err != nil {
		t.Errorf("Kubectl.RunCurlPod() error = %v, want nil", err)
	}
}

func TestGetPodNameByLabel(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	var b bytes.Buffer
	expectedParam := []string{"get", "pod", "-l=app.kubernetes.io/name=aws-otel-collector", "-o=jsonpath='{.items[0].metadata.name}'", "--kubeconfig", "c.kubeconfig", "--namespace", "observability"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	if _, err := tt.k.GetPodNameByLabel(tt.ctx, "observability", "app.kubernetes.io/name=aws-otel-collector", tt.cluster.KubeconfigFile); err != nil {
		t.Errorf("Kubectl.GetPodNameByLabel() error = %v, want nil", err)
	}
}

func TestGetPodIP(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	var b bytes.Buffer
	expectedParam := []string{"get", "pod", "generated-adot-75f769bc7-f7pmv", "-o=jsonpath='{.status.podIP}'", "--kubeconfig", "c.kubeconfig", "--namespace", "observability"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	if _, err := tt.k.GetPodIP(tt.ctx, "observability", "generated-adot-75f769bc7-f7pmv", tt.cluster.KubeconfigFile); err != nil {
		t.Errorf("Kubectl.GetPodIP() error = %v, want nil", err)
	}
}

func TestGetPodLogs(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	var b bytes.Buffer
	expectedParam := []string{"logs", "testpod", "testcontainer", "--kubeconfig", "c.kubeconfig", "--namespace", "eksa-packages"}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	if _, err := tt.k.GetPodLogs(tt.ctx, "eksa-packages", "testpod", "testcontainer", tt.cluster.KubeconfigFile); err != nil {
		t.Errorf("Kubectl.GetPodLogs() error = %v, want nil", err)
	}
}

func TestGetPodLogsSince(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	var b bytes.Buffer
	now := time.Now()
	expectedParam := []string{"logs", "testpod", "testcontainer", "--kubeconfig", "c.kubeconfig", "--namespace", "eksa-packages", "--since-time", now.Format(time.RFC3339)}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	if _, err := tt.k.GetPodLogsSince(tt.ctx, "eksa-packages", "testpod", "testcontainer", tt.cluster.KubeconfigFile, now); err != nil {
		t.Errorf("Kubectl.GetPodLogsSince() error = %v, want nil", err)
	}
}

func TestGetPodLogsSinceInternalError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	now := time.Now()
	var b bytes.Buffer
	b.WriteString("Internal Error")
	expectedParam := []string{"logs", "testpod", "testcontainer", "--kubeconfig", "c.kubeconfig", "--namespace", "eksa-packages", "--since-time", now.Format(time.RFC3339)}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, nil).AnyTimes()
	_, err := tt.k.GetPodLogsSince(tt.ctx, "eksa-packages", "testpod", "testcontainer", tt.cluster.KubeconfigFile, now)
	tt.Expect(err).To(MatchError(ContainSubstring("Internal Error")))
}

func TestGetPodLogsSinceExecuteError(t *testing.T) {
	t.Parallel()
	tt := newKubectlTest(t)
	now := time.Now()
	var b bytes.Buffer
	b.WriteString("execute error")
	expectedParam := []string{"logs", "testpod", "testcontainer", "--kubeconfig", "c.kubeconfig", "--namespace", "eksa-packages", "--since-time", now.Format(time.RFC3339)}
	tt.e.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(b, fmt.Errorf("execute error")).AnyTimes()
	_, err := tt.k.GetPodLogsSince(tt.ctx, "eksa-packages", "testpod", "testcontainer", tt.cluster.KubeconfigFile, now)
	tt.Expect(err).To(MatchError(ContainSubstring("execute error")))
}

func TestKubectlSearchTinkerbellMachineConfig(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace, name string
	buffer := bytes.Buffer{}
	buffer.WriteString(test.ReadFile(t, "testdata/kubectl_no_cs_machineconfigs.json"))
	k, ctx, _, e := newKubectl(t)

	expectedParam := []string{
		"get", fmt.Sprintf("tinkerbellmachineconfigs.%s", v1alpha1.GroupVersion.Group), "-o", "json", "--kubeconfig",
		kubeconfig, "--namespace", namespace, "--field-selector=metadata.name=" + name,
	}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(buffer, nil)
	mc, err := k.SearchTinkerbellMachineConfig(ctx, name, kubeconfig, namespace)
	if err != nil {
		t.Errorf("Kubectl.SearchTinkerbellMachineConfig() error = %v, want nil", err)
	}
	if len(mc) > 0 {
		t.Errorf("expected 0 machine configs, got %d", len(mc))
	}
}

func TestKubectlSearchTinkerbellMachineConfigNotFound(t *testing.T) {
	t.Parallel()
	var kubeconfigfile string
	tt := newKubectlTest(t)

	params := []string{
		"get", fmt.Sprintf("tinkerbellmachineconfigs.%s", v1alpha1.GroupVersion.Group), "-o", "json", "--kubeconfig",
		kubeconfigfile, "--namespace", tt.namespace, "--field-selector=metadata.name=test",
	}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(*bytes.NewBufferString(""), errors.New("machineconfig not found"))

	_, err := tt.k.SearchTinkerbellMachineConfig(tt.ctx, "test", kubeconfigfile, tt.namespace)
	tt.Expect(err).NotTo(BeNil())
}

func TestKubectlSearchTinkerbellDatacenterConfigs(t *testing.T) {
	t.Parallel()
	var kubeconfig, namespace, name string
	buffer := bytes.Buffer{}
	buffer.WriteString(test.ReadFile(t, "testdata/kubectl_no_cs_datacenterconfigs.json"))
	k, ctx, _, e := newKubectl(t)

	expectedParam := []string{
		"get", fmt.Sprintf("tinkerbelldatacenterconfigs.%s", v1alpha1.GroupVersion.Group), "-o", "json", "--kubeconfig",
		kubeconfig, "--namespace", namespace, "--field-selector=metadata.name=" + name,
	}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(buffer, nil)
	mc, err := k.SearchTinkerbellDatacenterConfig(ctx, name, kubeconfig, namespace)
	if err != nil {
		t.Errorf("Kubectl.SearchTinkerbellDatacenterConfig() error = %v, want nil", err)
	}
	if len(mc) > 0 {
		t.Errorf("expected 0 datacenter configs, got %d", len(mc))
	}
}

func TestKubectlSearchTinkerbellDatacenterConfigNotFound(t *testing.T) {
	t.Parallel()
	var kubeconfigfile string
	tt := newKubectlTest(t)

	params := []string{
		"get", fmt.Sprintf("tinkerbelldatacenterconfigs.%s", v1alpha1.GroupVersion.Group), "-o", "json", "--kubeconfig",
		kubeconfigfile, "--namespace", tt.namespace, "--field-selector=metadata.name=test",
	}
	tt.e.EXPECT().Execute(tt.ctx, gomock.Eq(params)).Return(*bytes.NewBufferString(""), errors.New("datacenterconfig not found"))

	_, err := tt.k.SearchTinkerbellDatacenterConfig(tt.ctx, "test", kubeconfigfile, tt.namespace)
	tt.Expect(err).NotTo(BeNil())
}

func TestGetMachineDeploymentsForCluster(t *testing.T) {
	t.Parallel()
	k, ctx, _, e := newKubectl(t)
	clusterName := "test0"
	fileContent := test.ReadFile(t, "testdata/kubectl_machine_deployments.json")
	wantMachineDeploymentNames := []string{
		"test0-md-0",
		"test1-md-0",
	}

	params := []string{
		"get", fmt.Sprintf("machinedeployments.%s", clusterv1.GroupVersion.Group), "-o", "json",
		fmt.Sprintf("--selector=cluster.x-k8s.io/cluster-name=%s", clusterName),
	}
	e.EXPECT().Execute(ctx, gomock.Eq(params)).Return(*bytes.NewBufferString(""), nil).Return(*bytes.NewBufferString(fileContent), nil)

	gotMachineDeployments, err := k.GetMachineDeploymentsForCluster(ctx, clusterName)
	if err != nil {
		t.Fatalf("Kubectl.GetMachineDeploymentsForCluster() error = %v, want nil", err)
	}

	gotMachineDeploymentNames := make([]string, 0, len(gotMachineDeployments))
	for _, p := range gotMachineDeployments {
		gotMachineDeploymentNames = append(gotMachineDeploymentNames, p.Name)
	}

	if !reflect.DeepEqual(gotMachineDeploymentNames, wantMachineDeploymentNames) {
		t.Fatalf("Kubectl.GetMachineDeployments() deployments = %+v, want %+v", gotMachineDeploymentNames, wantMachineDeploymentNames)
	}
}

func TestKubectlHasCRD(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		Name      string
		Error     error
		ExpectCRD bool
		ExpectErr bool
	}{
		{
			Name:      "CRDPresent",
			ExpectCRD: true,
		},
		{
			Name:  "CRDNotPresent",
			Error: errors.New("NotFound"),
		},
		{
			Name:      "Error",
			Error:     errors.New("some error"),
			ExpectErr: true,
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			k, ctx, _, e := newKubectl(t)
			const crd = "foo"
			const kubeconfig = "kubeconfig"
			var b bytes.Buffer

			params := []string{"get", "customresourcedefinition", crd, "--kubeconfig", kubeconfig}
			e.EXPECT().Execute(ctx, gomock.Eq(params)).Return(b, tt.Error)

			r, err := k.HasCRD(context.Background(), crd, kubeconfig)

			if tt.ExpectErr && !strings.Contains(err.Error(), tt.Error.Error()) {
				t.Fatalf("Expected error: %v; Received: %v", tt.ExpectErr, err)
			}

			if r != tt.ExpectCRD {
				t.Fatalf("Expected: %v; Received: %v", tt.ExpectCRD, r)
			}
		})
	}
}

func TestKubectlDeleteCRD(t *testing.T) {
	t.Parallel()
	//"delete", "crd", crd, "--kubeconfig", kubeconfig

	for _, tt := range []struct {
		Name      string
		Error     error
		ExpectErr bool
	}{
		{
			Name: "CRDPresent",
		},
		{
			Name:  "CRDNotPresent",
			Error: errors.New("NotFound"),
		},
		{
			Name:      "Error",
			Error:     errors.New("some error"),
			ExpectErr: true,
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			k, ctx, _, e := newKubectl(t)
			const crd = "foo"
			const kubeconfig = "kubeconfig"
			var b bytes.Buffer

			params := []string{"delete", "customresourcedefinition", crd, "--kubeconfig", kubeconfig}
			e.EXPECT().Execute(ctx, gomock.Eq(params)).Return(b, tt.Error)

			err := k.DeleteCRD(context.Background(), crd, kubeconfig)

			if tt.ExpectErr && !strings.Contains(err.Error(), tt.Error.Error()) {
				t.Fatalf("Expected error: %v; Received: %v", tt.ExpectErr, err)
			}
		})
	}
}

package reconciler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/go-logr/logr"
	"github.com/nutanix-cloud-native/prism-go-client/converged"
	"github.com/nutanix-cloud-native/prism-go-client/environment/credentials"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

type clusterSpecValidatorFunc func(context.Context, *cluster.Spec, credentials.BasicAuthCredential) error

func (f clusterSpecValidatorFunc) ValidateClusterSpec(ctx context.Context, spec *cluster.Spec, creds credentials.BasicAuthCredential) error {
	return f(ctx, spec, creds)
}

func TestValidationRecoversWithoutSpecChange(t *testing.T) {
	apiErr := fmt.Errorf("list clusters: %w", &url.Error{
		Op: "Get", URL: "https://pc.example.test:9440/api/clustermgmt/v4.2/config/clusters",
		Err: errors.New("GET request giving up after 6 attempt(s)"),
	})
	validationCalls := 0
	validator := clusterSpecValidatorFunc(func(_ context.Context, _ *cluster.Spec, _ credentials.BasicAuthCredential) error {
		validationCalls++
		if validationCalls == 1 {
			return apiErr
		}
		return nil
	})
	reconciler, spec := validationRecoveryFixture(t, validator)
	cniCalls := 0
	cni := func(_ context.Context, _ logr.Logger, spec *cluster.Spec) (controller.Result, error) {
		cniCalls++
		spec.Cluster.SetConditions(clusterv1.Conditions{{
			Type: anywherev1.DefaultCNIConfiguredCondition, Status: corev1.ConditionTrue,
		}})
		return controller.Result{}, nil
	}
	runner := controller.NewPhaseRunner[*cluster.Spec]().Register(
		reconciler.ValidateClusterSpec, clusters.CleanupStatusAfterValidate, cni,
	)
	original := spec.Cluster.DeepCopy()
	result, err := runner.Run(context.Background(), test.NewNullLogger(), spec)
	require.ErrorIs(t, err, apiErr)
	require.False(t, result.Return())
	require.False(t, spec.Cluster.HasFailure())
	require.Zero(t, cniCalls)
	require.Equal(t, original.Spec, spec.Cluster.Spec)

	result, err = runner.Run(context.Background(), test.NewNullLogger(), spec)
	require.NoError(t, err)
	require.False(t, result.Return())
	require.Equal(t, 2, validationCalls)
	require.Equal(t, 1, cniCalls)
	require.Equal(t, corev1.ConditionTrue, spec.Cluster.GetConditions()[0].Status)
	require.Equal(t, original.Spec, spec.Cluster.Spec)
	require.Equal(t, original.Generation, spec.Cluster.Generation)
}

func TestPermanentValidationStillStopsReconciliation(t *testing.T) {
	for _, validationErr := range []error{
		errors.New("invalid memory size"),
		fmt.Errorf("missing configured image: %w", converged.ErrNotFound),
	} {
		t.Run(validationErr.Error(), func(t *testing.T) {
			validator := clusterSpecValidatorFunc(func(context.Context, *cluster.Spec, credentials.BasicAuthCredential) error {
				return validationErr
			})
			reconciler, spec := validationRecoveryFixture(t, validator)
			cniCalled := false
			runner := controller.NewPhaseRunner[*cluster.Spec]().Register(
				reconciler.ValidateClusterSpec, clusters.CleanupStatusAfterValidate,
				func(context.Context, logr.Logger, *cluster.Spec) (controller.Result, error) {
					cniCalled = true
					return controller.Result{}, nil
				},
			)
			result, err := runner.Run(context.Background(), test.NewNullLogger(), spec)
			require.NoError(t, err)
			require.True(t, result.Return())
			require.False(t, cniCalled)
			require.Equal(t, anywherev1.ClusterInvalidReason, *spec.Cluster.Status.FailureReason)
			require.Equal(t, validationErr.Error(), *spec.Cluster.Status.FailureMessage)
		})
	}
}

func TestSuccessfulValidationClearsPreviousFailure(t *testing.T) {
	validator := clusterSpecValidatorFunc(func(context.Context, *cluster.Spec, credentials.BasicAuthCredential) error { return nil })
	reconciler, spec := validationRecoveryFixture(t, validator)
	spec.Cluster.SetFailure(anywherev1.ClusterInvalidReason, "old Prism availability error")
	runner := controller.NewPhaseRunner[*cluster.Spec]().Register(
		reconciler.ValidateClusterSpec, clusters.CleanupStatusAfterValidate,
	)
	result, err := runner.Run(context.Background(), test.NewNullLogger(), spec)
	require.NoError(t, err)
	require.False(t, result.Return())
	require.False(t, spec.Cluster.HasFailure())
}

func validationRecoveryFixture(t *testing.T, validator ClusterSpecValidator) (*Reconciler, *cluster.Spec) {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "test-creds", Namespace: "eksa-system"},
		Data: map[string][]byte{
			"credentials": []byte(`[{"type":"basic_auth","data":{"prismCentral":{"username":"test-user","password":"test-password"}}}]`),
		},
	}
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()
	spec := &cluster.Spec{Config: &cluster.Config{
		Cluster: &anywherev1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "test-cluster", Generation: 1}},
		NutanixDatacenter: &anywherev1.NutanixDatacenterConfig{
			Spec: anywherev1.NutanixDatacenterConfigSpec{
				CredentialRef: &anywherev1.Ref{Name: "test-creds", Kind: "Secret"},
			},
		},
	}}
	return New(client, validator, nil, nil, nil), spec
}

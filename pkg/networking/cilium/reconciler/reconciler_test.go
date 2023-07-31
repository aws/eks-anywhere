package reconciler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/networking/cilium/reconciler"
	"github.com/aws/eks-anywhere/pkg/networking/cilium/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestReconcilerReconcileInstall(t *testing.T) {
	tt := newReconcileTest(t)
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	manifest := buildManifest(tt.WithT, ds, operator)
	tt.templater.EXPECT().GenerateManifest(tt.ctx, tt.spec).Return(manifest, nil)

	tt.Expect(
		tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec),
	).To(Equal(controller.Result{}))
	tt.expectDaemonSetSemanticallyEqual(ds)
	tt.expectOperatorSemanticallyEqual(operator)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("True", "", "", ""))
}

func TestReconcilerReconcileInstallErrorGeneratingManifest(t *testing.T) {
	tt := newReconcileTest(t)
	tt.templater.EXPECT().GenerateManifest(tt.ctx, tt.spec).Return(nil, errors.New("generating manifest"))

	result, err := tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.Expect(err).To(MatchError(ContainSubstring("generating manifest")))
}

func TestReconcilerReconcileErrorYamlReconcile(t *testing.T) {
	tt := newReconcileTest(t)
	tt.templater.EXPECT().GenerateManifest(tt.ctx, tt.spec).Return([]byte("invalid yaml"), nil)

	result, err := tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.Expect(err).To(MatchError(ContainSubstring("error unmarshaling JSON")))
}

func TestReconcilerReconcileAlreadyUpToDate(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	cm := ciliumConfigMap()
	tt := newReconcileTest(t).withObjects(ds, operator, cm)

	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.Result{}),
	)
	tt.expectDaemonSetSemanticallyEqual(ds)
	tt.expectOperatorSemanticallyEqual(operator)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("True", "", "", ""))
}

func TestReconcilerReconcileAlreadyInDesiredVersionWithPreflight(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	preflightDS := ciliumPreflightDaemonSet()
	cm := ciliumConfigMap()
	preflightDeployment := ciliumPreflightDeployment()
	tt := newReconcileTest(t)

	// for deleting the preflight
	preflightManifest := tt.buildManifest(preflightDS, preflightDeployment)
	tt.templater.EXPECT().GenerateUpgradePreflightManifest(tt.ctx, tt.spec).Return(preflightManifest, nil)

	tt.withObjects(ds, operator, preflightDS, preflightDeployment, cm)

	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.Result{}),
	)
	tt.expectDaemonSetSemanticallyEqual(ds)
	tt.expectOperatorSemanticallyEqual(operator)
	tt.expectDSToNotExist(preflightDS.Name, preflightDS.Namespace)
	tt.expectDeploymentToNotExist(preflightDeployment.Name, preflightDeployment.Namespace)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("True", "", "", ""))
}

func TestReconcilerReconcileAlreadyInDesiredVersionWithPreflightErrorFromTemplater(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	cm := ciliumConfigMap()
	preflightDS := ciliumPreflightDaemonSet()
	preflightDeployment := ciliumPreflightDeployment()
	tt := newReconcileTest(t)

	// for deleting the preflight
	tt.templater.EXPECT().GenerateUpgradePreflightManifest(tt.ctx, tt.spec).Return(nil, errors.New("generating preflight"))

	tt.withObjects(ds, operator, cm, preflightDS, preflightDeployment)

	result, err := tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.Expect(err).To(MatchError(ContainSubstring("generating preflight")))
}

func TestReconcilerReconcileAlreadyInDesiredVersionWithPreflightErrorDeletingYaml(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	cm := ciliumConfigMap()
	preflightDS := ciliumPreflightDaemonSet()
	preflightDeployment := ciliumPreflightDeployment()
	tt := newReconcileTest(t)

	// for deleting the preflight
	tt.templater.EXPECT().GenerateUpgradePreflightManifest(tt.ctx, tt.spec).Return([]byte("invalid yaml"), nil)

	tt.withObjects(ds, operator, cm, preflightDS, preflightDeployment)

	result, err := tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.Expect(err).To(MatchError(ContainSubstring("error unmarshaling JSON")))
}

func TestReconcilerReconcileUpgradeButCiliumDaemonSetNotReady(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	tt := newReconcileTest(t).withObjects(ds, operator)
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:1.11.1-eksa-1"
		s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:1.11.1-eksa-1"
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
	})

	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.ResultWithRequeue(10 * time.Second)),
	)

	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("False", anywherev1.DefaultCNIUpgradeInProgressReason, v1beta1.ConditionSeverityInfo, "Cilium version upgrade needed"))
}

func TestReconcilerReconcileUpgradeNeedsPreflightAndPreflightDaemonSetNotAvailable(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	tt := newReconcileTest(t).withObjects(ds, operator)
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:1.11.1-eksa-1"
		s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:1.11.1-eksa-1"
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
	})

	tt.templater.EXPECT().GenerateUpgradePreflightManifest(tt.ctx, tt.spec)

	tt.makeCiliumDaemonSetReady()
	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.ResultWithRequeue(10 * time.Second)),
	)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("False", anywherev1.DefaultCNIUpgradeInProgressReason, v1beta1.ConditionSeverityInfo, "Cilium version upgrade needed"))
}

func TestReconcilerReconcileUpgradeErrorGeneratingPreflight(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	tt := newReconcileTest(t).withObjects(ds, operator)
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:1.11.1-eksa-1"
		s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:1.11.1-eksa-1"
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
	})

	tt.templater.EXPECT().GenerateUpgradePreflightManifest(tt.ctx, tt.spec).Return(nil, errors.New("generating preflight"))

	tt.makeCiliumDaemonSetReady()
	result, err := tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.Expect(err).To(MatchError(ContainSubstring("generating preflight")))
}

func TestReconcilerReconcileUpgradeNeedsPreflightAndPreflightDeploymentNotAvailable(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	tt := newReconcileTest(t).withObjects(ds, operator)
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:1.11.1-eksa-1"
		s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:1.11.1-eksa-1"
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
	})

	preflightManifest := tt.buildManifest(ciliumPreflightDaemonSet())
	tt.templater.EXPECT().GenerateUpgradePreflightManifest(tt.ctx, tt.spec).Return(preflightManifest, nil)

	tt.makeCiliumDaemonSetReady()
	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.ResultWithRequeue(10 * time.Second)),
	)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("False", anywherev1.DefaultCNIUpgradeInProgressReason, v1beta1.ConditionSeverityInfo, "Cilium version upgrade needed"))
}

func TestReconcilerReconcileUpgradeNeedsPreflightAndPreflightNotReady(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	tt := newReconcileTest(t).withObjects(ds, operator)
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:1.11.1-eksa-1"
		s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:1.11.1-eksa-1"
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
	})

	preflightManifest := tt.buildManifest(ciliumPreflightDaemonSet(), ciliumPreflightDeployment())
	tt.templater.EXPECT().GenerateUpgradePreflightManifest(tt.ctx, tt.spec).Return(preflightManifest, nil)

	tt.makeCiliumDaemonSetReady()
	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.ResultWithRequeue(10 * time.Second)),
	)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("False", anywherev1.DefaultCNIUpgradeInProgressReason, v1beta1.ConditionSeverityInfo, "Cilium version upgrade needed"))
}

func TestReconcilerReconcileUpgradePreflightDaemonSetNotReady(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	tt := newReconcileTest(t).withObjects(ds, operator, ciliumPreflightDaemonSet(), ciliumPreflightDeployment())
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:1.11.1-eksa-1"
		s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:1.11.1-eksa-1"
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
	})

	tt.makeCiliumDaemonSetReady()
	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.ResultWithRequeue(10 * time.Second)),
	)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("False", anywherev1.DefaultCNIUpgradeInProgressReason, v1beta1.ConditionSeverityInfo, "Cilium version upgrade needed"))
}

func TestReconcilerReconcileUpgradePreflightDeploymentSetNotReady(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	preflight := ciliumPreflightDaemonSet()
	tt := newReconcileTest(t).withObjects(ds, operator, preflight, ciliumPreflightDeployment())
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:1.11.1-eksa-1"
		s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:1.11.1-eksa-1"
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
	})

	tt.makeCiliumDaemonSetReady()
	tt.makePreflightDaemonSetReady()
	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.ResultWithRequeue(10 * time.Second)),
	)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("False", anywherev1.DefaultCNIUpgradeInProgressReason, v1beta1.ConditionSeverityInfo, "Cilium version upgrade needed"))
}

func TestReconcilerReconcileUpgradeInvalidCiliumInstalledVersion(t *testing.T) {
	ds := ciliumDaemonSet()
	ds.Spec.Template.Spec.Containers[0].Image = "cilium:eksa-invalid-version"
	operator := ciliumOperator()
	preflight := ciliumPreflightDaemonSet()
	newDSImage := "cilium:1.11.1-eksa-1"
	newOperatorImage := "cilium-operator:1.11.1-eksa-1"

	tt := newReconcileTest(t).withObjects(ds, operator, preflight, ciliumPreflightDeployment())
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = newDSImage
		s.VersionsBundles["1.19"].Cilium.Operator.URI = newOperatorImage
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
	})

	tt.makeCiliumDaemonSetReady()
	tt.makePreflightDaemonSetReady()
	tt.makePreflightDeploymentReady()

	result, err := tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.Expect(err).To(MatchError(ContainSubstring("installed cilium DS has an invalid version tag")))
}

func TestReconcilerReconcileUpgradeErrorGeneratingManifest(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	preflight := ciliumPreflightDaemonSet()
	newDSImage := "cilium:1.11.1-eksa-1"
	newOperatorImage := "cilium-operator:1.11.1-eksa-1"

	tt := newReconcileTest(t).withObjects(ds, operator, preflight, ciliumPreflightDeployment())
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = newDSImage
		s.VersionsBundles["1.19"].Cilium.Operator.URI = newOperatorImage
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
	})

	tt.templater.EXPECT().GenerateManifest(tt.ctx, tt.spec, gomock.Not(gomock.Nil())).Return(nil, errors.New("generating manifest"))

	tt.makeCiliumDaemonSetReady()
	tt.makePreflightDaemonSetReady()
	tt.makePreflightDeploymentReady()

	result, err := tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.Expect(err).To(MatchError(ContainSubstring("generating manifest")))
	tt.expectCiliumInstalledAnnotation()
}

func TestReconcilerReconcileUpgradePreflightErrorYamlReconcile(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	preflight := ciliumPreflightDaemonSet()
	newDSImage := "cilium:1.11.1-eksa-1"
	newOperatorImage := "cilium-operator:1.11.1-eksa-1"

	tt := newReconcileTest(t).withObjects(ds, operator, preflight, ciliumPreflightDeployment())
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = newDSImage
		s.VersionsBundles["1.19"].Cilium.Operator.URI = newOperatorImage
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
	})

	tt.templater.EXPECT().GenerateManifest(tt.ctx, tt.spec, gomock.Not(gomock.Nil())).Return([]byte("invalid yaml"), nil)

	tt.makeCiliumDaemonSetReady()
	tt.makePreflightDaemonSetReady()
	tt.makePreflightDeploymentReady()
	result, err := tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.Expect(err).To(MatchError(ContainSubstring("error unmarshaling JSON")))
	tt.expectCiliumInstalledAnnotation()
}

func TestReconcilerReconcileUpgradePreflightReady(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	preflight := ciliumPreflightDaemonSet()
	newDSImage := "cilium:1.11.1-eksa-1"
	newOperatorImage := "cilium-operator:1.11.1-eksa-1"
	wantDS := ds.DeepCopy()
	wantDS.Spec.Template.Spec.Containers[0].Image = newDSImage
	wantOperator := operator.DeepCopy()
	wantOperator.Spec.Template.Spec.Containers[0].Image = newOperatorImage

	tt := newReconcileTest(t).withObjects(ds, operator, preflight, ciliumPreflightDeployment())
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = newDSImage
		s.VersionsBundles["1.19"].Cilium.Operator.URI = newOperatorImage
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
	})

	upgradeManifest := tt.buildManifest(wantDS, wantOperator)
	tt.templater.EXPECT().GenerateManifest(tt.ctx, tt.spec, gomock.Not(gomock.Nil())).Return(upgradeManifest, nil)

	// for deleting the preflight
	preflightManifest := tt.buildManifest(ciliumPreflightDaemonSet(), ciliumPreflightDeployment())
	tt.templater.EXPECT().GenerateUpgradePreflightManifest(tt.ctx, tt.spec).Return(preflightManifest, nil)

	tt.makeCiliumDaemonSetReady()
	tt.makePreflightDaemonSetReady()
	tt.makePreflightDeploymentReady()
	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.Result{}),
	)
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("True", "", "", ""))
}

func TestReconcilerReconcileUpdateConfigConfigMapEnablePolicyChange(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	cm := ciliumConfigMap()
	tt := newReconcileTest(t).withObjects(ds, operator, cm)

	newDSImage := "cilium:1.10.1-eksa-1"
	newOperatorImage := "cilium-operator:1.10.1-eksa-1"

	upgradeManifest := tt.buildManifest(ds, operator, cm)

	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].Cilium.Cilium.URI = newDSImage
		s.VersionsBundles["1.19"].Cilium.Operator.URI = newOperatorImage
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{
				PolicyEnforcementMode: "always",
			},
		}
	})

	tt.templater.EXPECT().GenerateManifest(tt.ctx, tt.spec, gomock.Not(gomock.Nil())).Return(upgradeManifest, nil)

	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.Result{}),
	)
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("True", "", "", ""))
}

func TestReconcilerReconcileSkipUpgradeWithoutCiliumInstalled(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	cm := ciliumConfigMap()
	tt := newReconcileTest(t)

	upgradeManifest := tt.buildManifest(ds, operator, cm)

	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{
				SkipUpgrade: ptr.Bool(true),
			},
		}
	})

	tt.templater.EXPECT().GenerateManifest(tt.ctx, tt.spec).Return(upgradeManifest, nil)

	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.Result{}),
	)
	tt.expectDaemonSetSemanticallyEqual(ds)
	tt.expectOperatorSemanticallyEqual(operator)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("True", "", "", ""))
}

func TestReconcilerReconcileSkipUpgradeWithCiliumInstalled(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	cm := ciliumConfigMap()
	tt := newReconcileTest(t).withObjects(ds, operator, cm)

	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{
				SkipUpgrade: ptr.Bool(true),
			},
		}
	})

	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.Result{}),
	)
	tt.expectDaemonSetSemanticallyEqual(ds)
	tt.expectOperatorSemanticallyEqual(operator)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("False", anywherev1.SkipUpgradesForDefaultCNIConfiguredReason, v1beta1.ConditionSeverityWarning, "Configured to skip default Cilium CNI upgrades"))
}

func TestReconcilerReconcileSkipUpgradeWithAnnotationWithoutCilium(t *testing.T) {
	tt := newReconcileTest(t)
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{
				SkipUpgrade: ptr.Bool(true),
			},
		}
		s.Cluster.Annotations = map[string]string{
			reconciler.EKSACiliumInstalledAnnotation: "true",
		}
	})

	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.Result{}),
	)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("False", anywherev1.SkipUpgradesForDefaultCNIConfiguredReason, v1beta1.ConditionSeverityWarning, "Configured to skip default Cilium CNI upgrades"))
}

func TestReconcilerReconcileSkipUpgradeWithAnnotationWithCilium(t *testing.T) {
	ds := ciliumDaemonSet()
	operator := ciliumOperator()
	cm := ciliumConfigMap()
	tt := newReconcileTest(t).withObjects(ds, operator, cm)

	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{
				SkipUpgrade: ptr.Bool(true),
			},
		}
		s.Cluster.Name = rand.String(10)
		s.Cluster.Namespace = "default"
		s.Cluster.Annotations = map[string]string{
			reconciler.EKSACiliumInstalledAnnotation: "true",
		}
	})

	if err := tt.client.Create(context.Background(), tt.spec.Cluster); err != nil {
		t.Fatal(err)
	}

	tt.Expect(tt.reconciler.Reconcile(tt.ctx, test.NewNullLogger(), tt.client, tt.spec)).To(
		Equal(controller.Result{}),
	)
	tt.expectDaemonSetSemanticallyEqual(ds)
	tt.expectOperatorSemanticallyEqual(operator)
	tt.expectCiliumInstalledAnnotation()
	tt.expectDefaultCNIConfigured(defaultCNIConfiguredCondition("False", anywherev1.SkipUpgradesForDefaultCNIConfiguredReason, v1beta1.ConditionSeverityWarning, "Configured to skip default Cilium CNI upgrades"))
}

type reconcileTest struct {
	*WithT
	t          *testing.T
	ctx        context.Context
	env        *envtest.Environment
	spec       *cluster.Spec
	client     client.Client
	templater  *mocks.MockTemplater
	reconciler *reconciler.Reconciler
}

func newReconcileTest(t *testing.T) *reconcileTest {
	ctrl := gomock.NewController(t)
	templater := mocks.NewMockTemplater(ctrl)

	tt := &reconcileTest{
		WithT: NewWithT(t),
		t:     t,
		ctx:   context.Background(),
		spec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.VersionsBundles["1.19"].Cilium.Cilium.URI = "cilium:1.10.1-eksa-1"
			s.VersionsBundles["1.19"].Cilium.Operator.URI = "cilium-operator:1.10.1-eksa-1"
			s.Cluster.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
				Cilium: &anywherev1.CiliumConfig{
					PolicyEnforcementMode: "default",
				},
			}
		}),
		client:     env.Client(),
		env:        env,
		templater:  templater,
		reconciler: reconciler.New(templater),
	}

	t.Cleanup(tt.cleanup)

	return tt
}

func (tt *reconcileTest) cleanup() {
	tt.Expect(tt.client.DeleteAllOf(tt.ctx, &appsv1.DaemonSet{}, client.InNamespace("kube-system")))
	tt.Expect(tt.client.DeleteAllOf(tt.ctx, &appsv1.Deployment{}, client.InNamespace("kube-system")))
	tt.Expect(tt.client.DeleteAllOf(tt.ctx, &corev1.ConfigMap{}, client.InNamespace("kube-system")))
	tt.Expect(tt.client.DeleteAllOf(tt.ctx, &anywherev1.Cluster{}))
}

func (tt *reconcileTest) withObjects(objs ...client.Object) *reconcileTest {
	tt.t.Helper()
	envtest.CreateObjs(tt.ctx, tt.t, tt.client, objs...)
	return tt
}

func (tt *reconcileTest) expectDSToNotExist(name, namespace string) {
	tt.t.Helper()
	err := tt.env.APIReader().Get(tt.ctx, types.NamespacedName{Name: name, Namespace: namespace}, &appsv1.DaemonSet{})
	tt.Expect(apierrors.IsNotFound(err)).To(BeTrue(), "DaemonSet %s should not exist", name)
}

func (tt *reconcileTest) expectDeploymentToNotExist(name, namespace string) {
	tt.t.Helper()
	err := tt.env.APIReader().Get(tt.ctx, types.NamespacedName{Name: name, Namespace: namespace}, &appsv1.Deployment{})
	tt.Expect(apierrors.IsNotFound(err)).To(BeTrue(), "Deployment %s should not exist", name)
}

func (tt *reconcileTest) getDaemonSet(name, namespace string) *appsv1.DaemonSet {
	tt.t.Helper()
	ds := &appsv1.DaemonSet{}
	tt.Expect(tt.env.APIReader().Get(tt.ctx, types.NamespacedName{Name: name, Namespace: namespace}, ds)).To(Succeed())

	return ds
}

func (tt *reconcileTest) getDeployment(name, namespace string) *appsv1.Deployment {
	tt.t.Helper()
	deployment := &appsv1.Deployment{}
	tt.Expect(tt.env.APIReader().Get(tt.ctx, types.NamespacedName{Name: name, Namespace: namespace}, deployment)).To(Succeed())

	return deployment
}

func (tt *reconcileTest) getCiliumOperator() *appsv1.Deployment {
	tt.t.Helper()
	return tt.getDeployment(cilium.DeploymentName, "kube-system")
}

func (tt *reconcileTest) getCiliumDaemonSet() *appsv1.DaemonSet {
	tt.t.Helper()
	return tt.getDaemonSet(cilium.DaemonSetName, "kube-system")
}

func (tt *reconcileTest) makeCiliumDaemonSetReady() {
	tt.t.Helper()
	tt.makeDaemonSetReady(cilium.DaemonSetName, "kube-system")
}

func (tt *reconcileTest) makePreflightDaemonSetReady() {
	tt.t.Helper()
	tt.makeDaemonSetReady(cilium.PreflightDaemonSetName, "kube-system")
}

func (tt *reconcileTest) makeDaemonSetReady(name, namespace string) {
	tt.t.Helper()
	ds := tt.getDaemonSet(name, namespace)
	ds.Status.ObservedGeneration = ds.Generation
	tt.Expect(tt.client.Status().Update(tt.ctx, ds)).To(Succeed())

	// wait for cache to refresh
	r := retrier.New(1*time.Second, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		return true, 50 * time.Millisecond
	}))
	tt.Expect(
		r.Retry(func() error {
			ds := &appsv1.DaemonSet{}
			tt.Expect(tt.client.Get(tt.ctx, types.NamespacedName{Name: name, Namespace: namespace}, ds)).To(Succeed())

			if ds.Status.ObservedGeneration != ds.Generation {
				return errors.New("ds cache not updated yet")
			}
			return nil
		}),
	).To(Succeed())
}

func (tt *reconcileTest) makePreflightDeploymentReady() {
	tt.t.Helper()
	tt.makeDeploymentReady(cilium.PreflightDeploymentName, "kube-system")
}

func (tt *reconcileTest) makeDeploymentReady(name, namespace string) {
	tt.t.Helper()
	deployment := tt.getDeployment(name, namespace)
	deployment.Status.ObservedGeneration = deployment.Generation
	tt.Expect(tt.client.Status().Update(tt.ctx, deployment)).To(Succeed())

	// wait for cache to refresh
	r := retrier.New(1*time.Second, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		return true, 50 * time.Millisecond
	}))
	tt.Expect(
		r.Retry(func() error {
			deployment := &appsv1.Deployment{}
			tt.Expect(tt.client.Get(tt.ctx, types.NamespacedName{Name: name, Namespace: namespace}, deployment)).To(Succeed())

			if deployment.Status.ObservedGeneration != deployment.Generation {
				return errors.New("deployment cache not updated yet")
			}
			return nil
		}),
	).To(Succeed())
}

func (tt *reconcileTest) expectDaemonSetSemanticallyEqual(wantDS *appsv1.DaemonSet) {
	tt.t.Helper()
	gotDS := tt.getCiliumDaemonSet()
	tt.Expect(equality.Semantic.DeepDerivative(wantDS.Spec, gotDS.Spec)).To(
		BeTrue(), "Cilium DaemonSet should be semantically equivalent",
	)
}

func (tt *reconcileTest) expectOperatorSemanticallyEqual(wantOperator *appsv1.Deployment) {
	tt.t.Helper()
	gotOperator := tt.getCiliumOperator()
	tt.Expect(equality.Semantic.DeepDerivative(wantOperator.Spec, gotOperator.Spec)).To(
		BeTrue(), "Cilium Operator should be semantically equivalent",
	)
}

func (tt *reconcileTest) expectDefaultCNIConfigured(wantCondition *anywherev1.Condition) {
	tt.t.Helper()
	condition := conditions.Get(tt.spec.Cluster, anywherev1.DefaultCNIConfiguredCondition)
	tt.Expect(condition).ToNot(BeNil(), "missing defaultcniconfigured condition")
	tt.Expect(condition).To(conditions.HaveSameStateOf(wantCondition))
}

func (tt *reconcileTest) expectCiliumInstalledAnnotation() {
	tt.t.Helper()

	if tt.spec.Cluster.Annotations == nil {
		tt.t.Fatal("missing cilium installed annotation")
	}

	if _, ok := tt.spec.Cluster.Annotations[reconciler.EKSACiliumInstalledAnnotation]; !ok {
		tt.t.Fatal("missing cilium installed annotation")
	}
}

func (tt *reconcileTest) buildManifest(objs ...client.Object) []byte {
	tt.t.Helper()
	return buildManifest(tt.WithT, objs...)
}

func buildManifest(g *WithT, objs ...client.Object) []byte {
	manifests := [][]byte{}
	for _, obj := range objs {
		o, err := yaml.Marshal(obj)
		g.Expect(err).ToNot(HaveOccurred(), "Marshall obj for manifest should succeed")
		manifests = append(manifests, o)
	}

	return templater.AppendYamlResources(manifests...)
}

func ciliumDaemonSet() *appsv1.DaemonSet {
	return simpleDaemonSet(cilium.DaemonSetName, "cilium:1.10.1-eksa-1")
}

func ciliumOperator() *appsv1.Deployment {
	return simpleDeployment(cilium.DeploymentName, "cilium-operator:1.10.1-eksa-1")
}

func ciliumConfigMap() *corev1.ConfigMap {
	return simpleConfigMap(cilium.ConfigMapName, "default")
}

func ciliumPreflightDaemonSet() *appsv1.DaemonSet {
	return simpleDaemonSet(cilium.PreflightDaemonSetName, "cilium-pre-flight-check:1.10.1-eksa-1")
}

func ciliumPreflightDeployment() *appsv1.Deployment {
	return simpleDeployment(cilium.PreflightDeploymentName, "cilium-pre-flight-check:1.10.1-eksa-1")
}

func simpleDeployment(name, image string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kube-system",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "cilium",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "cilium",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
						},
					},
				},
			},
		},
	}
}

func simpleDaemonSet(name, image string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kube-system",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "cilium",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "cilium",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
						},
					},
				},
			},
		},
	}
}

func simpleConfigMap(name, enablePolicy string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"enable-policy": enablePolicy,
		},
	}
}

func defaultCNIConfiguredCondition(status corev1.ConditionStatus, reason string, severity v1beta1.ConditionSeverity, message string) *anywherev1.Condition {
	return &anywherev1.Condition{
		Type:     anywherev1.DefaultCNIConfiguredCondition,
		Status:   status,
		Severity: severity,
		Reason:   reason,
		Message:  message,
	}
}

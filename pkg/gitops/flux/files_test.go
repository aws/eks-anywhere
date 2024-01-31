package flux_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	writerMocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/gitops/flux"
	fluxMocks "github.com/aws/eks-anywhere/pkg/gitops/flux/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
)

var wantConfig = `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: test-cluster
  namespace: default
spec:
  clusterNetwork:
    cniConfig: {}
    pods: {}
    services: {}
  controlPlaneConfiguration: {}
  datacenterRef: {}
  gitOpsRef:
    kind: FluxConfig
    name: test-gitops
  kubernetesVersion: "1.19"
  managementCluster:
    name: test-cluster

---
kind: VSphereDatacenterConfig
metadata:
  name: test-cluster
  namespace: default
spec:
  datacenter: SDDC-Datacenter
  insecure: false
  network: ""
  server: ""
  thumbprint: ""

---
kind: VSphereMachineConfig
metadata:
  name: test-cluster
  namespace: default
spec:
  datastore: ""
  folder: ""
  memoryMiB: 0
  numCPUs: 0
  osFamily: ""
  resourcePool: ""
  template: /SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: FluxConfig
metadata:
  name: test-gitops
  namespace: default
spec:
  branch: testBranch
  clusterConfigPath: clusters/test-cluster
  github:
    owner: mFolwer
    personal: true
    repository: testRepo
  systemNamespace: flux-system

---
`

var wantEksaKustomization = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- {{.ConfigFileName}}`

var wantFluxKustomization = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{.Namespace}}
resources:
- gotk-components.yaml
- gotk-sync.yaml
patches:
- patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: helm-controller
      namespace: {{.Namespace}}
    spec:
      template:
        spec:
          containers:
          - image: {{.HelmControllerImage}}
            name: manager
- patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: kustomize-controller
      namespace: {{.Namespace}}
    spec:
      template:
        spec:
          containers:
          - image: {{.KustomizeControllerImage}}
            name: manager
- patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: notification-controller
      namespace: {{.Namespace}}
    spec:
      template:
        spec:
          containers:
          - image: {{.NotificationControllerImage}}
            name: manager
- patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: source-controller
      namespace: {{.Namespace}}
    spec:
      template:
        spec:
          containers:
          - image: {{.SourceControllerImage}}
            name: manager
`

var wantKustomizationValues = map[string]string{
	"Namespace":                   "flux-system",
	"SourceControllerImage":       "public.ecr.aws/l0g8r8j6/fluxcd/source-controller:v0.12.1-8539f509df046a4f567d2182dde824b957136599",
	"KustomizeControllerImage":    "public.ecr.aws/l0g8r8j6/fluxcd/kustomize-controller:v0.11.1-d82011942ec8a447ba89a70ff9a84bf7b9579492",
	"HelmControllerImage":         "public.ecr.aws/l0g8r8j6/fluxcd/helm-controller:v0.10.0-d82011942ec8a447ba89a70ff9a84bf7b9579492",
	"NotificationControllerImage": "public.ecr.aws/l0g8r8j6/fluxcd/notification-controller:v0.13.0-d82011942ec8a447ba89a70ff9a84bf7b9579492",
}

type fileGeneratorTest struct {
	*WithT
	ctx                  context.Context
	g                    *flux.FileGenerator
	w                    *writerMocks.MockFileWriter
	t                    *fluxMocks.MockTemplater
	managementComponents *cluster.ManagementComponents
	clusterSpec          *cluster.Spec
	datacenterConfig     providers.DatacenterConfig
	machineConfigs       []providers.MachineConfig
}

func newFileGeneratorTest(t *testing.T) *fileGeneratorTest {
	ctrl := gomock.NewController(t)
	writer := writerMocks.NewMockFileWriter(ctrl)
	templater := fluxMocks.NewMockTemplater(ctrl)
	clusterName := "test-cluster"

	clusterSpec := newClusterSpec(t, NewCluster(clusterName), "")

	return &fileGeneratorTest{
		WithT:                NewWithT(t),
		ctx:                  context.Background(),
		g:                    flux.NewFileGeneratorWithWriterTemplater(writer, writer, templater, templater),
		w:                    writer,
		t:                    templater,
		managementComponents: cluster.ManagementComponentsFromBundles(clusterSpec.Bundles),
		clusterSpec:          clusterSpec,
		datacenterConfig:     datacenterConfig(clusterName),
		machineConfigs:       []providers.MachineConfig{machineConfig(clusterName)},
	}
}

func TestFileGeneratorInitSuccess(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.w.EXPECT().WithDir("dir1").Return(tt.w, nil)
	tt.w.EXPECT().WithDir("dir2").Return(tt.w, nil)
	tt.w.EXPECT().CleanUpTemp().Times(2)

	tt.Expect(tt.g.Init(tt.w, "dir1", "dir2")).To(Succeed())
}

func TestFileGeneratorInitEksaWriterError(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.w.EXPECT().WithDir("dir1").Return(nil, errors.New("error in writer dir1"))

	tt.Expect(tt.g.Init(tt.w, "dir1", "dir2")).To(MatchError(ContainSubstring("error in writer dir1")))
}

func TestFileGeneratorInitFluxWriterError(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.w.EXPECT().WithDir("dir1").Return(tt.w, nil)
	tt.w.EXPECT().CleanUpTemp()
	tt.w.EXPECT().WithDir("dir2").Return(nil, errors.New("error in writer dir2"))

	tt.Expect(tt.g.Init(tt.w, "dir1", "dir2")).To(MatchError(ContainSubstring("error in writer dir2")))
}

func TestFileGeneratorWriteEksaFilesSuccess(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.w.EXPECT().Write("eksa-cluster.yaml", []byte(wantConfig), gomock.Any()).Return("", nil)
	tt.t.EXPECT().WriteToFile(wantEksaKustomization, map[string]string{"ConfigFileName": "eksa-cluster.yaml"}, "kustomization.yaml", gomock.Any()).Return("", nil)

	tt.Expect(tt.g.WriteEksaFiles(tt.clusterSpec, tt.datacenterConfig, tt.machineConfigs)).To(Succeed())
}

func TestFileGeneratorWriteEksaFilesSkip(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.Expect(tt.g.WriteEksaFiles(tt.clusterSpec, nil, nil)).To(Succeed())
}

func TestFileGeneratorWriteEksaFilesWriteError(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.w.EXPECT().Write("eksa-cluster.yaml", []byte(wantConfig), gomock.Any()).Return("", errors.New("error in write"))

	tt.Expect(tt.g.WriteEksaFiles(tt.clusterSpec, tt.datacenterConfig, tt.machineConfigs)).To(MatchError(ContainSubstring("error in write")))
}

func TestFileGeneratorWriteEksaFilesWriteToFileError(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.w.EXPECT().Write("eksa-cluster.yaml", []byte(wantConfig), gomock.Any()).Return("", nil)
	tt.t.EXPECT().WriteToFile(wantEksaKustomization, map[string]string{"ConfigFileName": "eksa-cluster.yaml"}, "kustomization.yaml", gomock.Any()).Return("", errors.New("error in write to file"))

	tt.Expect(tt.g.WriteEksaFiles(tt.clusterSpec, tt.datacenterConfig, tt.machineConfigs)).To(MatchError(ContainSubstring("error in write to file")))
}

func TestFileGeneratorWriteFluxSystemFilesSuccess(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.t.EXPECT().WriteToFile(wantFluxKustomization, wantKustomizationValues, "kustomization.yaml", gomock.Any()).Return("", nil)
	tt.t.EXPECT().WriteToFile("", nil, "gotk-sync.yaml", gomock.Any()).Return("", nil)

	tt.Expect(tt.g.WriteFluxSystemFiles(tt.managementComponents, tt.clusterSpec)).To(Succeed())
}

func TestFileGeneratorWriteFluxSystemFilesWriteFluxKustomizationError(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.t.EXPECT().WriteToFile(wantFluxKustomization, wantKustomizationValues, "kustomization.yaml", gomock.Any()).Return("", errors.New("error in write kustomization"))

	tt.Expect(tt.g.WriteFluxSystemFiles(tt.managementComponents, tt.clusterSpec)).To(MatchError(ContainSubstring("error in write kustomization")))
}

func TestFileGeneratorWriteFluxSystemFilesWriteFluxSyncError(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.t.EXPECT().WriteToFile(wantFluxKustomization, wantKustomizationValues, "kustomization.yaml", gomock.Any()).Return("", nil)
	tt.t.EXPECT().WriteToFile("", nil, "gotk-sync.yaml", gomock.Any()).Return("", errors.New("error in write sync"))

	tt.Expect(tt.g.WriteFluxSystemFiles(tt.managementComponents, tt.clusterSpec)).To(MatchError(ContainSubstring("error in write sync")))
}

func NewCluster(clusterName string) *anywherev1.Cluster {
	c := &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube119,
		},
		Status: anywherev1.ClusterStatus{},
	}
	c.SetSelfManaged()

	return c
}

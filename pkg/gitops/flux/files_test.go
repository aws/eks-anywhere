package flux_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
- {{.ConfigFileName}}
{{- if .HardwareFileName }}
- {{.HardwareFileName}}
{{- end }}`

var wantFluxKustomization = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{.Namespace}}
resources:
  - gotk-components.yaml
  - gotk-sync.yaml
patchesStrategicMerge:
  - gotk-patches.yaml`

var wantFluxPatches = `apiVersion: apps/v1
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
---
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
---
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
---
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
        name: manager`

var wantPatchesValues = map[string]string{
	"Namespace":                   "flux-system",
	"SourceControllerImage":       "public.ecr.aws/l0g8r8j6/fluxcd/source-controller:v0.12.1-8539f509df046a4f567d2182dde824b957136599",
	"KustomizeControllerImage":    "public.ecr.aws/l0g8r8j6/fluxcd/kustomize-controller:v0.11.1-d82011942ec8a447ba89a70ff9a84bf7b9579492",
	"HelmControllerImage":         "public.ecr.aws/l0g8r8j6/fluxcd/helm-controller:v0.10.0-d82011942ec8a447ba89a70ff9a84bf7b9579492",
	"NotificationControllerImage": "public.ecr.aws/l0g8r8j6/fluxcd/notification-controller:v0.13.0-d82011942ec8a447ba89a70ff9a84bf7b9579492",
}

var clusterName = "test-cluster"

type fileGeneratorTest struct {
	*WithT
	ctx              context.Context
	g                *flux.FileGenerator
	w                *writerMocks.MockFileWriter
	t                *fluxMocks.MockTemplater
	clusterSpec      *cluster.Spec
	datacenterConfig providers.DatacenterConfig
	machineConfigs   []providers.MachineConfig
}

func newFileGeneratorTest(t *testing.T) *fileGeneratorTest {
	ctrl := gomock.NewController(t)
	writer := writerMocks.NewMockFileWriter(ctrl)
	templater := fluxMocks.NewMockTemplater(ctrl)
	return &fileGeneratorTest{
		WithT:            NewWithT(t),
		ctx:              context.Background(),
		g:                flux.NewFileGeneratorWithWriterTemplater(writer, writer, templater, templater),
		w:                writer,
		t:                templater,
		clusterSpec:      newClusterSpec(t, v1alpha1.NewCluster(clusterName), ""),
		datacenterConfig: datacenterConfig(clusterName),
		machineConfigs:   []providers.MachineConfig{machineConfig(clusterName)},
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

	tt.Expect(tt.g.WriteEksaFiles(tt.clusterSpec, tt.datacenterConfig, tt.machineConfigs, "")).To(Succeed())
}

func TestFileGeneratorWriteEksaFilesWithHardwareSuccess(t *testing.T) {
	tt := newFileGeneratorTest(t)
	clus := v1alpha1.NewCluster(clusterName)
	clus.Spec.DatacenterRef = v1alpha1.Ref{
		Kind: v1alpha1.TinkerbellDatacenterKind,
		Name: clusterName,
	}
	tt.clusterSpec = newClusterSpec(t, clus, "")
	tinkDatacenterConfig := &v1alpha1.TinkerbellDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind: v1alpha1.TinkerbellDatacenterKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.TinkerbellDatacenterConfigSpec{
			TinkerbellIP: "",
		},
	}
	tinkMachineConfig := &v1alpha1.TinkerbellMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind: v1alpha1.TinkerbellMachineConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.TinkerbellMachineConfigSpec{
			OSFamily: "bottlerocket",
		},
	}
	tt.w.EXPECT().Write("eksa-cluster.yaml", []byte(wantTinkerbellConfig()), gomock.Any()).Return("", nil)
	tt.w.EXPECT().Write("hardware.yaml", []byte(wantHardwareConfig()), gomock.Any()).Return("", nil)
	tt.t.EXPECT().WriteToFile(wantEksaKustomization, map[string]string{"ConfigFileName": "eksa-cluster.yaml", "HardwareFileName": "hardware.yaml"}, "kustomization.yaml", gomock.Any()).Return("", nil)

	tt.Expect(tt.g.WriteEksaFiles(tt.clusterSpec, tinkDatacenterConfig, []providers.MachineConfig{tinkMachineConfig}, "./testdata/hardware.csv")).To(Succeed())
}

func TestFileGeneratorWriteEksaFilesSkip(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.Expect(tt.g.WriteEksaFiles(tt.clusterSpec, nil, nil, "")).To(Succeed())
}

func TestFileGeneratorWriteEksaFilesWriteError(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.w.EXPECT().Write("eksa-cluster.yaml", []byte(wantConfig), gomock.Any()).Return("", errors.New("error in write"))

	tt.Expect(tt.g.WriteEksaFiles(tt.clusterSpec, tt.datacenterConfig, tt.machineConfigs, "")).To(MatchError(ContainSubstring("error in write")))
}

func TestFileGeneratorWriteEksaFilesWriteToFileError(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.w.EXPECT().Write("eksa-cluster.yaml", []byte(wantConfig), gomock.Any()).Return("", nil)
	tt.t.EXPECT().WriteToFile(wantEksaKustomization, map[string]string{"ConfigFileName": "eksa-cluster.yaml"}, "kustomization.yaml", gomock.Any()).Return("", errors.New("error in write to file"))

	tt.Expect(tt.g.WriteEksaFiles(tt.clusterSpec, tt.datacenterConfig, tt.machineConfigs, "")).To(MatchError(ContainSubstring("error in write to file")))
}

func TestFileGeneratorWriteFluxSystemFilesSuccess(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.t.EXPECT().WriteToFile(wantFluxKustomization, map[string]string{"Namespace": "flux-system"}, "kustomization.yaml", gomock.Any()).Return("", nil)
	tt.t.EXPECT().WriteToFile("", nil, "gotk-sync.yaml", gomock.Any()).Return("", nil)
	tt.t.EXPECT().WriteToFile(wantFluxPatches, wantPatchesValues, "gotk-patches.yaml", gomock.Any()).Return("", nil)

	tt.Expect(tt.g.WriteFluxSystemFiles(tt.clusterSpec)).To(Succeed())
}

func TestFileGeneratorWriteFluxSystemFilesWriteFluxKustomizationError(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.t.EXPECT().WriteToFile(wantFluxKustomization, map[string]string{"Namespace": "flux-system"}, "kustomization.yaml", gomock.Any()).Return("", errors.New("error in write kustomization"))

	tt.Expect(tt.g.WriteFluxSystemFiles(tt.clusterSpec)).To(MatchError(ContainSubstring("error in write kustomization")))
}

func TestFileGeneratorWriteFluxSystemFilesWriteFluxSyncError(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.t.EXPECT().WriteToFile(wantFluxKustomization, map[string]string{"Namespace": "flux-system"}, "kustomization.yaml", gomock.Any()).Return("", nil)
	tt.t.EXPECT().WriteToFile("", nil, "gotk-sync.yaml", gomock.Any()).Return("", errors.New("error in write sync"))

	tt.Expect(tt.g.WriteFluxSystemFiles(tt.clusterSpec)).To(MatchError(ContainSubstring("error in write sync")))
}

func TestFileGeneratorWriteFluxSystemFilesWriteFluxPatchesError(t *testing.T) {
	tt := newFileGeneratorTest(t)

	tt.t.EXPECT().WriteToFile(wantFluxKustomization, map[string]string{"Namespace": "flux-system"}, "kustomization.yaml", gomock.Any()).Return("", nil)
	tt.t.EXPECT().WriteToFile("", nil, "gotk-sync.yaml", gomock.Any()).Return("", nil)
	tt.t.EXPECT().WriteToFile(wantFluxPatches, wantPatchesValues, "gotk-patches.yaml", gomock.Any()).Return("", errors.New("error in write patches"))

	tt.Expect(tt.g.WriteFluxSystemFiles(tt.clusterSpec)).To(MatchError(ContainSubstring("error in write patches")))
}

func wantTinkerbellConfig() string {
	return `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
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
  datacenterRef:
    kind: TinkerbellDatacenterConfig
    name: test-cluster
  gitOpsRef:
    kind: FluxConfig
    name: test-gitops
  kubernetesVersion: "1.19"
  managementCluster:
    name: test-cluster

---
kind: TinkerbellDatacenterConfig
metadata:
  name: test-cluster
  namespace: default
spec:
  tinkerbellIP: ""

---
kind: TinkerbellMachineConfig
metadata:
  name: test-cluster
  namespace: default
spec:
  hardwareSelector: null
  osFamily: bottlerocket
  templateRef: {}

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
}

func wantHardwareConfig() string {
	return `apiVersion: tinkerbell.org/v1alpha1
kind: Hardware
metadata:
  creationTimestamp: null
  labels:
    type: cp
  name: worker1
  namespace: eksa-system
spec:
  bmcRef:
    apiGroup: null
    kind: Machine
    name: bmc-worker1
  disks:
  - device: /dev/sda
  interfaces:
  - dhcp:
      arch: x86_64
      hostname: worker1
      ip:
        address: 10.10.10.10
        family: 4
        gateway: 10.10.10.1
        netmask: 255.255.255.0
      lease_time: 4294967294
      mac: "00:00:00:00:00:01"
      name_servers:
      - 1.1.1.1
      uefi: true
    netboot:
      allowPXE: true
      allowWorkflow: true
  metadata:
    facility:
      facility_code: onprem
      plan_slug: c2.medium.x86
    instance:
      allow_pxe: true
      always_pxe: true
      hostname: worker1
      id: "00:00:00:00:00:01"
      ips:
      - address: 10.10.10.10
        family: 4
        gateway: 10.10.10.1
        netmask: 255.255.255.0
        public: true
      operating_system: {}
status: {}
---
apiVersion: bmc.tinkerbell.org/v1alpha1
kind: Machine
metadata:
  creationTimestamp: null
  name: bmc-worker1
  namespace: eksa-system
spec:
  connection:
    authSecretRef:
      name: bmc-worker1-auth
      namespace: eksa-system
    host: 192.168.0.10
    insecureTLS: true
    port: 0
status: {}
---
apiVersion: v1
data:
  password: YWRtaW4=
  username: QWRtaW4=
kind: Secret
metadata:
  creationTimestamp: null
  labels:
    clusterctl.cluster.x-k8s.io/move: "true"
  name: bmc-worker1-auth
  namespace: eksa-system
type: kubernetes.io/basic-auth
---
`
}

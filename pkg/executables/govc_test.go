package executables_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

const (
	govcUsername    = "GOVC_USERNAME"
	govcPassword    = "GOVC_PASSWORD"
	govcURL         = "GOVC_URL"
	govcInsecure    = "GOVC_INSECURE"
	vSphereUsername = "EKSA_VSPHERE_USERNAME"
	vSpherePassword = "EKSA_VSPHERE_PASSWORD"
	vSphereServer   = "VSPHERE_SERVER"
	templateLibrary = "eks-a-templates"
)

var govcEnvironment = map[string]string{
	govcUsername: "vsphere_username",
	govcPassword: "vsphere_password",
	govcURL:      "vsphere_server",
	govcInsecure: "false",
}

type testContext struct {
	oldUsername   string
	isUsernameSet bool
	oldPassword   string
	isPasswordSet bool
	oldServer     string
	isServerSet   bool
}

func (tctx *testContext) SaveContext() {
	tctx.oldUsername, tctx.isUsernameSet = os.LookupEnv(vSphereUsername)
	tctx.oldPassword, tctx.isPasswordSet = os.LookupEnv(vSpherePassword)
	tctx.oldServer, tctx.isServerSet = os.LookupEnv(vSphereServer)
	os.Setenv(vSphereUsername, "vsphere_username")
	os.Setenv(vSpherePassword, "vsphere_password")
	os.Setenv(vSphereServer, "vsphere_server")
	os.Setenv(govcUsername, os.Getenv(vSphereUsername))
	os.Setenv(govcPassword, os.Getenv(vSpherePassword))
	os.Setenv(govcURL, os.Getenv(vSphereServer))
	os.Setenv(govcInsecure, "false")
}

func (tctx *testContext) RestoreContext() {
	if tctx.isUsernameSet {
		os.Setenv(vSphereUsername, tctx.oldUsername)
	} else {
		os.Unsetenv(vSphereUsername)
	}
	if tctx.isPasswordSet {
		os.Setenv(vSpherePassword, tctx.oldPassword)
	} else {
		os.Unsetenv(vSpherePassword)
	}
	if tctx.isServerSet {
		os.Setenv(vSphereServer, tctx.oldServer)
	} else {
		os.Unsetenv(vSphereServer)
	}
}

func setupContext(t *testing.T) {
	var tctx testContext
	tctx.SaveContext()
	t.Cleanup(func() {
		tctx.RestoreContext()
	})
}

func setup(t *testing.T) (govc *executables.Govc, mockExecutable *mockexecutables.MockExecutable, env map[string]string) {
	setupContext(t)
	_, writer := test.NewWriter(t)
	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	g := executables.NewGovc(executable, writer)

	return g, executable, govcEnvironment
}

type deployTemplateTest struct {
	govc                     *executables.Govc
	mockExecutable           *mockexecutables.MockExecutable
	env                      map[string]string
	datacenter               string
	datastore                string
	resourcePool             string
	templatePath             string
	ovaURL                   string
	deployFolder             string
	templateInLibraryPathAbs string
	templateName             string
	resizeDisk2              bool
	ctx                      context.Context
	fakeExecResponse         *bytes.Buffer
	expectations             []*gomock.Call
}

func newDeployTemplateTest(t *testing.T) *deployTemplateTest {
	g, exec, env := setup(t)
	return &deployTemplateTest{
		govc:                     g,
		mockExecutable:           exec,
		env:                      env,
		datacenter:               "SDDC-Datacenter",
		datastore:                "/SDDC-Datacenter/datastore/WorkloadDatastore",
		resourcePool:             "*/Resources/Compute-ResourcePool",
		templatePath:             "/SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6",
		ovaURL:                   "https://aws.com/ova",
		deployFolder:             "/SDDC-Datacenter/vm/Templates",
		templateInLibraryPathAbs: "/eks-a-templates/ubuntu-2004-kube-v1.19.6",
		templateName:             "ubuntu-2004-kube-v1.19.6",
		resizeDisk2:              false,
		ctx:                      context.Background(),
		fakeExecResponse:         bytes.NewBufferString("dummy"),
		expectations:             make([]*gomock.Call, 0),
	}
}

func newMachineConfig(t *testing.T) *v1alpha1.VSphereMachineConfig {
	return &v1alpha1.VSphereMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind: v1alpha1.VSphereMachineConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "eksa-unit-test",
		},
		Spec: v1alpha1.VSphereMachineConfigSpec{
			Template: "/SDDC-Datacenter/vm/Templates/ubuntu-v1.19.8-eks-d-1-19-4-eks-a-0.0.1.build.38-amd64",
		},
	}
}

func (dt *deployTemplateTest) expectFolderInfoToReturn(err error) {
	dt.expectations = append(
		dt.expectations,
		dt.mockExecutable.EXPECT().ExecuteWithEnv(dt.ctx, dt.env, "folder.info", dt.deployFolder).Return(*dt.fakeExecResponse, err),
	)
}

func (dt *deployTemplateTest) expectDeployToReturn(err error) {
	dt.expectations = append(
		dt.expectations,
		dt.mockExecutable.EXPECT().ExecuteWithEnv(dt.ctx, dt.env, "library.deploy", "-dc", dt.datacenter, "-ds", dt.datastore, "-pool", dt.resourcePool, "-folder", dt.deployFolder, "-options", test.OfType("string"), dt.templateInLibraryPathAbs, dt.templateName).Return(*dt.fakeExecResponse, err),
	)
}

func (dt *deployTemplateTest) expectCreateSnapshotToReturn(err error) {
	dt.expectations = append(
		dt.expectations,
		dt.mockExecutable.EXPECT().ExecuteWithEnv(dt.ctx, dt.env, "snapshot.create", "-dc", dt.datacenter, "-m=false", "-vm", dt.templatePath, "root").Return(*dt.fakeExecResponse, err),
	)
}

func (dt *deployTemplateTest) expectMarkAsTemplateToReturn(err error) {
	dt.expectations = append(
		dt.expectations,
		dt.mockExecutable.EXPECT().ExecuteWithEnv(dt.ctx, dt.env, "vm.markastemplate", "-dc", dt.datacenter, dt.templatePath).Return(*dt.fakeExecResponse, err),
	)
}

func (dt *deployTemplateTest) DeployTemplateFromLibrary() error {
	gomock.InOrder(dt.expectations...)
	return dt.govc.DeployTemplateFromLibrary(dt.ctx, dt.deployFolder, dt.templateName, templateLibrary, dt.datacenter, dt.datastore, dt.resourcePool, dt.resizeDisk2)
}

func (dt *deployTemplateTest) assertDeployTemplateSuccess(t *testing.T) {
	if err := dt.DeployTemplateFromLibrary(); err != nil {
		t.Fatalf("govc.DeployTemplateFromLibrary() err = %v, want err = nil", err)
	}
}

func (dt *deployTemplateTest) assertDeployTemplateError(t *testing.T) {
	if err := dt.DeployTemplateFromLibrary(); err == nil {
		t.Fatal("govc.DeployTemplateFromLibrary() err = nil, want err not nil")
	}
}

func TestSearchTemplateItExists(t *testing.T) {
	ctx := context.Background()
	datacenter := "SDDC-Datacenter"

	g, executable, env := setup(t)
	machineConfig := newMachineConfig(t)
	machineConfig.Spec.Template = "/SDDC Datacenter/vm/Templates/ubuntu 2004-kube-v1.19.6"
	executable.EXPECT().ExecuteWithEnv(ctx, env, "find", "-json", "/"+datacenter, "-type", "VirtualMachine", "-name", filepath.Base(machineConfig.Spec.Template)).Return(*bytes.NewBufferString("[\"/SDDC Datacenter/vm/Templates/ubuntu 2004-kube-v1.19.6\"]"), nil)

	_, err := g.SearchTemplate(ctx, datacenter, machineConfig)
	if err != nil {
		t.Fatalf("Govc.SearchTemplate() exists = false, want true %v", err)
	}
}

func TestSearchTemplateItDoesNotExists(t *testing.T) {
	machineConfig := newMachineConfig(t)
	ctx := context.Background()
	datacenter := "SDDC-Datacenter"

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "find", "-json", "/"+datacenter, "-type", "VirtualMachine", "-name", filepath.Base(machineConfig.Spec.Template)).Return(*bytes.NewBufferString(""), nil)

	templateFullPath, err := g.SearchTemplate(ctx, datacenter, machineConfig)
	if err == nil && len(templateFullPath) > 0 {
		t.Fatalf("Govc.SearchTemplate() exists = true, want false %v", err)
	}
}

func TestSearchTemplateError(t *testing.T) {
	machineConfig := newMachineConfig(t)
	ctx := context.Background()
	datacenter := "SDDC-Datacenter"

	g, executable, env := setup(t)
	g.Retrier = retrier.NewWithMaxRetries(5, 0)
	executable.EXPECT().ExecuteWithEnv(ctx, env, gomock.Any()).Return(bytes.Buffer{}, errors.New("error from execute with env")).Times(5)

	_, err := g.SearchTemplate(ctx, datacenter, machineConfig)
	if err == nil {
		t.Fatal("Govc.SearchTemplate() err = nil, want err not nil")
	}
}

func TestLibraryElementExistsItExists(t *testing.T) {
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.ls", templateLibrary).Return(*bytes.NewBufferString("testing"), nil)

	exists, err := g.LibraryElementExists(ctx, templateLibrary)
	if err != nil {
		t.Fatalf("Govc.LibraryElementExists() err = %v, want err nil", err)
	}
	if !exists {
		t.Fatalf("Govc.LibraryElementExists() exists = false, want true")
	}
}

func TestLibraryElementExistsItDoesNotExists(t *testing.T) {
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.ls", templateLibrary).Return(*bytes.NewBufferString(""), nil)

	exists, err := g.LibraryElementExists(ctx, templateLibrary)
	if err != nil {
		t.Fatalf("Govc.LibraryElementExists() err = %v, want err nil", err)
	}
	if exists {
		t.Fatalf("Govc.LibraryElementExists() exists = true, want false")
	}
}

func TestLibraryElementExistsError(t *testing.T) {
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.ls", templateLibrary).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	_, err := g.LibraryElementExists(ctx, templateLibrary)
	if err == nil {
		t.Fatal("Govc.LibraryElementExists() err = nil, want err not nil")
	}
}

func TestGetLibraryElementContentVersionSuccess(t *testing.T) {
	ctx := context.Background()
	response := `[
		{
			"content_version": "1"
		},
	]`
	libraryElement := "/eks-a-templates/ubuntu-2004-kube-v1.19.6"

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.info", "-json", libraryElement).Return(*bytes.NewBufferString(response), nil)

	_, err := g.GetLibraryElementContentVersion(ctx, libraryElement)
	if err != nil {
		t.Fatalf("Govc.GetLibraryElementContentVersion() err = %v, want err nil", err)
	}
}

func TestGetLibraryElementContentVersionError(t *testing.T) {
	ctx := context.Background()
	libraryElement := "/eks-a-templates/ubuntu-2004-kube-v1.19.6"

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.info", "-json", libraryElement).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	_, err := g.GetLibraryElementContentVersion(ctx, libraryElement)
	if err == nil {
		t.Fatal("Govc.GetLibraryElementContentVersion() err = nil, want err not nil")
	}
}

func TestDeleteLibraryElementSuccess(t *testing.T) {
	ctx := context.Background()
	libraryElement := "/eks-a-templates/ubuntu-2004-kube-v1.19.6"

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.rm", libraryElement).Return(*bytes.NewBufferString(""), nil)

	err := g.DeleteLibraryElement(ctx, libraryElement)
	if err != nil {
		t.Fatalf("Govc.DeleteLibraryElement() err = %v, want err nil", err)
	}
}

func TestDeleteLibraryElementError(t *testing.T) {
	ctx := context.Background()
	libraryElement := "/eks-a-templates/ubuntu-2004-kube-v1.19.6"

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.rm", libraryElement).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	err := g.DeleteLibraryElement(ctx, libraryElement)
	if err == nil {
		t.Fatal("Govc.DeleteLibraryElement() err = nil, want err not nil")
	}
}

func TestGovcTemplateHasSnapshot(t *testing.T) {
	_, writer := test.NewWriter(t)
	template := "/SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6"

	env := govcEnvironment

	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	params := []string{"snapshot.tree", "-vm", template}
	executable.EXPECT().ExecuteWithEnv(ctx, env, params).Return(*bytes.NewBufferString("testing"), nil)
	g := executables.NewGovc(executable, writer)
	snap, err := g.TemplateHasSnapshot(ctx, template)
	if err != nil {
		t.Fatalf("error getting template snapshot: %v", err)
	}
	if !snap {
		t.Fatalf("Govc.TemplateHasSnapshot() error got = %+v, want %+v", snap, true)
	}
}

func TestGovcGetWorkloadAvailableSpace(t *testing.T) {
	tests := []struct {
		testName         string
		jsonResponseFile string
		wantValue        float64
	}{
		{
			testName:         "success",
			jsonResponseFile: "testdata/govc_no_datastore.json",
			wantValue:        1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, writer := test.NewWriter(t)
			fileContent := test.ReadFile(t, tt.jsonResponseFile)
			env := govcEnvironment
			datastore := "/SDDC-Datacenter/datastore/WorkloadDatastore"

			ctx := context.Background()
			mockCtrl := gomock.NewController(t)

			var tctx testContext
			tctx.SaveContext()
			defer tctx.RestoreContext()

			executable := mockexecutables.NewMockExecutable(mockCtrl)
			params := []string{"datastore.info", "-json=true", datastore}
			executable.EXPECT().ExecuteWithEnv(ctx, env, params).Return(*bytes.NewBufferString(fileContent), nil)
			g := executables.NewGovc(executable, writer)
			freeSpace, err := g.GetWorkloadAvailableSpace(ctx, datastore)
			if err != nil {
				t.Fatalf("Govc.GetWorkloadAvailableSpace() error: %v", err)
			}

			if freeSpace != tt.wantValue {
				t.Fatalf("Govc.GetWorkloadAvailableSpace() freeSpace = %+v, want %+v", freeSpace, tt.wantValue)
			}
		})
	}
}

func TestDeployTemplateFromLibrarySuccess(t *testing.T) {
	tt := newDeployTemplateTest(t)
	tt.expectFolderInfoToReturn(nil)
	tt.expectDeployToReturn(nil)
	tt.expectCreateSnapshotToReturn(nil)
	tt.expectMarkAsTemplateToReturn(nil)

	tt.assertDeployTemplateSuccess(t)
}

func TestDeployTemplateFromLibraryErrorDeploy(t *testing.T) {
	tt := newDeployTemplateTest(t)
	tt.expectFolderInfoToReturn(nil)
	tt.expectDeployToReturn(errors.New("error exec"))
	tt.assertDeployTemplateError(t)
}

func TestDeployTemplateFromLibraryErrorCreateSnapshot(t *testing.T) {
	tt := newDeployTemplateTest(t)
	tt.expectFolderInfoToReturn(nil)
	tt.expectDeployToReturn(nil)
	tt.expectCreateSnapshotToReturn(errors.New("error exec"))
	tt.assertDeployTemplateError(t)
}

func TestDeployTemplateFromLibraryErrorMarkAsTemplate(t *testing.T) {
	tt := newDeployTemplateTest(t)
	tt.expectFolderInfoToReturn(nil)
	tt.expectDeployToReturn(nil)
	tt.expectCreateSnapshotToReturn(nil)
	tt.expectMarkAsTemplateToReturn(errors.New("error exec"))

	tt.assertDeployTemplateError(t)
}

func TestGovcValidateVCenterSetupMachineConfig(t *testing.T) {
	ctx := context.Background()
	ts := newHTTPSServer(t)
	datacenterConfig := v1alpha1.VSphereDatacenterConfig{
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Datacenter: "SDDC Datacenter",
			Network:    "/SDDC Datacenter/network/test network",
			Server:     strings.TrimPrefix(ts.URL, "https://"),
			Insecure:   true,
		},
	}
	machineConfig := v1alpha1.VSphereMachineConfig{
		Spec: v1alpha1.VSphereMachineConfigSpec{
			Datastore:    "/SDDC Datacenter/datastore/testDatastore",
			Folder:       "/SDDC Datacenter/vm/test",
			ResourcePool: "*/Resources/Compute ResourcePool",
		},
	}
	env := govcEnvironment
	mockCtrl := gomock.NewController(t)
	_, writer := test.NewWriter(t)
	selfSigned := true

	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	executable := mockexecutables.NewMockExecutable(mockCtrl)

	var params []string

	params = []string{"datastore.info", machineConfig.Spec.Datastore}
	executable.EXPECT().ExecuteWithEnv(ctx, env, params).Return(bytes.Buffer{}, nil)

	params = []string{"folder.info", machineConfig.Spec.Folder}
	executable.EXPECT().ExecuteWithEnv(ctx, env, params).Return(bytes.Buffer{}, nil)

	datacenter := "/" + datacenterConfig.Spec.Datacenter
	resourcePoolName := "Compute ResourcePool"
	params = []string{"find", "-json", datacenter, "-type", "p", "-name", resourcePoolName}
	executable.EXPECT().ExecuteWithEnv(ctx, env, params).Return(*bytes.NewBufferString("[\"/SDDC Datacenter/host/Cluster-1/Resources/Compute ResourcePool\"]"), nil)

	g := executables.NewGovc(executable, writer)

	err := g.ValidateVCenterSetupMachineConfig(ctx, &datacenterConfig, &machineConfig, &selfSigned)
	if err != nil {
		t.Fatalf("Govc.ValidateVCenterSetup() error: %v", err)
	}
}

func newHTTPSServer(t *testing.T) *httptest.Server {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("ready")); err != nil {
			t.Errorf("Failed writing response to http request: %s", err)
		}
	}))
	t.Cleanup(func() { ts.Close() })
	return ts
}

func TestGovcCleanupVms(t *testing.T) {
	ctx := context.Background()
	clusterName := "cluster"
	vmName := clusterName

	var dryRun bool

	env := govcEnvironment
	mockCtrl := gomock.NewController(t)
	_, writer := test.NewWriter(t)

	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	executable := mockexecutables.NewMockExecutable(mockCtrl)

	var params []string
	params = []string{"find", "-type", "VirtualMachine", "-name", clusterName + "*"}
	executable.EXPECT().ExecuteWithEnv(ctx, env, params).Return(*bytes.NewBufferString(clusterName), nil)

	params = []string{"vm.power", "-off", "-force", vmName}
	executable.EXPECT().ExecuteWithEnv(ctx, env, params).Return(bytes.Buffer{}, nil)

	params = []string{"object.destroy", vmName}
	executable.EXPECT().ExecuteWithEnv(ctx, env, params).Return(bytes.Buffer{}, nil)

	g := executables.NewGovc(executable, writer)

	err := g.CleanupVms(ctx, clusterName, dryRun)
	if err != nil {
		t.Fatalf("Govc.CleanupVms() error: %v", err)
	}
}

func TestCreateLibrarySuccess(t *testing.T) {
	datastore := "/SDDC-Datacenter/datastore/WorkloadDatastore"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.create", "-ds", datastore, templateLibrary).Return(*bytes.NewBufferString("testing"), nil)

	err := g.CreateLibrary(ctx, datastore, templateLibrary)
	if err != nil {
		t.Fatalf("Govc.CreateLibrary() err = %v, want err nil", err)
	}
}

func TestCreateLibraryError(t *testing.T) {
	datastore := "/SDDC-Datacenter/datastore/WorkloadDatastore"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.create", "-ds", datastore, templateLibrary).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	err := g.CreateLibrary(ctx, datastore, templateLibrary)
	if err == nil {
		t.Fatal("Govc.CreateLibrary() err = nil, want err not nil")
	}
}

func TestGetTagsSuccessNoTags(t *testing.T) {
	path := "/SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.attached.ls", "-json", "-r", path).Return(*bytes.NewBufferString("null"), nil)

	tags, err := g.GetTags(ctx, path)
	if err != nil {
		t.Fatalf("Govc.GetTags() err = %v, want err nil", err)
	}

	if len(tags) != 0 {
		t.Fatalf("Govc.GetTags() tags size = %d, want 0", len(tags))
	}
}

func TestGetTagsSuccessHasTags(t *testing.T) {
	path := "/SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6"
	ctx := context.Background()

	tagsReponse := `[
		 "kubernetesChannel:1.19",
		 "eksd:1.19-4"
	]`
	wantTags := []string{"kubernetesChannel:1.19", "eksd:1.19-4"}

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.attached.ls", "-json", "-r", path).Return(*bytes.NewBufferString(tagsReponse), nil)

	gotTags, err := g.GetTags(ctx, path)
	if err != nil {
		t.Fatalf("Govc.GetTags() err = %v, want err nil", err)
	}

	if !reflect.DeepEqual(gotTags, wantTags) {
		t.Fatalf("Govc.GetTags() tags %v, want %v", gotTags, wantTags)
	}
}

func TestGetTagsErrorGovc(t *testing.T) {
	path := "/SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6"
	ctx := context.Background()

	g, executable, env := setup(t)
	g.Retrier = retrier.NewWithMaxRetries(5, 0)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.attached.ls", "-json", "-r", path).Return(bytes.Buffer{}, errors.New("error from exec")).Times(5)

	_, err := g.GetTags(ctx, path)
	if err == nil {
		t.Fatal("Govc.GetTags() err = nil, want err not nil")
	}
}

func TestGetTagsErrorUnmarshalling(t *testing.T) {
	path := "/SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.attached.ls", "-json", "-r", path).Return(*bytes.NewBufferString("invalid"), nil)

	_, err := g.GetTags(ctx, path)
	if err == nil {
		t.Fatal("Govc.GetTags() err = nil, want err not nil")
	}
}

func TestListTagsSuccessNoTags(t *testing.T) {
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.ls", "-json").Return(*bytes.NewBufferString("null"), nil)

	tags, err := g.ListTags(ctx)
	if err != nil {
		t.Fatalf("Govc.ListTags() err = %v, want err nil", err)
	}

	if len(tags) != 0 {
		t.Fatalf("Govc.ListTags() tags size = %d, want 0", len(tags))
	}
}

func TestListTagsSuccessHasTags(t *testing.T) {
	ctx := context.Background()

	tagsReponse := `[
		{
			"id": "urn:vmomi:InventoryServiceTag:5555:GLOBAL",
			"name": "eksd:1.19-4",
			"category_id": "eksd"
		},
		{
			"id": "urn:vmomi:InventoryServiceTag:5555:GLOBAL",
			"name": "kubernetesChannel:1.19",
			"category_id": "kubernetesChannel"
		}
	]`
	wantTags := []string{"eksd:1.19-4", "kubernetesChannel:1.19"}

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.ls", "-json").Return(*bytes.NewBufferString(tagsReponse), nil)

	gotTags, err := g.ListTags(ctx)
	if err != nil {
		t.Fatalf("Govc.ListTags() err = %v, want err nil", err)
	}

	if !reflect.DeepEqual(gotTags, wantTags) {
		t.Fatalf("Govc.ListTags() tags = %v, want %v", gotTags, wantTags)
	}
}

func TestListTagsErrorGovc(t *testing.T) {
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.ls", "-json").Return(bytes.Buffer{}, errors.New("error from exec"))

	_, err := g.ListTags(ctx)
	if err == nil {
		t.Fatal("Govc.ListTags() err = nil, want err not nil")
	}
}

func TestListTagsErrorUnmarshalling(t *testing.T) {
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.ls", "-json").Return(*bytes.NewBufferString("invalid"), nil)

	_, err := g.ListTags(ctx)
	if err == nil {
		t.Fatal("Govc.ListTags() err = nil, want err not nil")
	}
}

func TestAddTagSuccess(t *testing.T) {
	tag := "tag"
	path := "/SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.attach", tag, path).Return(*bytes.NewBufferString(""), nil)

	err := g.AddTag(ctx, path, tag)
	if err != nil {
		t.Fatalf("Govc.AddTag() err = %v, want err nil", err)
	}
}

func TestAddTagError(t *testing.T) {
	tag := "tag"
	path := "/SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.attach", tag, path).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	err := g.AddTag(ctx, path, tag)
	if err == nil {
		t.Fatal("Govc.AddTag() err = nil, want err not nil")
	}
}

func TestCreateTagSuccess(t *testing.T) {
	category := "category"
	tag := "tag"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.create", "-c", category, tag).Return(*bytes.NewBufferString(""), nil)

	err := g.CreateTag(ctx, tag, category)
	if err != nil {
		t.Fatalf("Govc.CreateTag() err = %v, want err nil", err)
	}
}

func TestCreateTagError(t *testing.T) {
	category := "category"
	tag := "tag"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.create", "-c", category, tag).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	err := g.CreateTag(ctx, tag, category)
	if err == nil {
		t.Fatal("Govc.CreateTag() err = nil, want err not nil")
	}
}

func TestListCategoriesSuccessNoCategories(t *testing.T) {
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.category.ls", "-json").Return(*bytes.NewBufferString("null"), nil)

	gotCategories, err := g.ListCategories(ctx)
	if err != nil {
		t.Fatalf("Govc.ListCategories() err = %v, want err nil", err)
	}

	if len(gotCategories) != 0 {
		t.Fatalf("Govc.ListCategories() tags size = %d, want 0", len(gotCategories))
	}
}

func TestListCategoriesSuccessHasCategories(t *testing.T) {
	ctx := context.Background()

	catsResponse := `[
		{
			"id": "urn:vmomi:InventoryServiceCategory:78484:GLOBAL",
			"name": "eksd",
			"cardinality": "MULTIPLE",
			"associable_types": [
			"com.vmware.content.library.Item",
				"VirtualMachine"
			]
		},
		{
			"id": "urn:vmomi:InventoryServiceCategory:78484:GLOBAL",
			"name": "kubernetesChannel",
			"cardinality": "SINGLE",
			"associable_types": [
				"VirtualMachine"
			]
		}
	]`
	wantCats := []string{"eksd", "kubernetesChannel"}

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.category.ls", "-json").Return(*bytes.NewBufferString(catsResponse), nil)

	gotCats, err := g.ListCategories(ctx)
	if err != nil {
		t.Fatalf("Govc.ListCategories() err = %v, want err nil", err)
	}

	if !reflect.DeepEqual(gotCats, wantCats) {
		t.Fatalf("Govc.ListCategories() tags = %v, want %v", gotCats, wantCats)
	}
}

func TestListCategoriesErrorGovc(t *testing.T) {
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.category.ls", "-json").Return(bytes.Buffer{}, errors.New("error from exec"))

	_, err := g.ListCategories(ctx)
	if err == nil {
		t.Fatal("Govc.ListCategories() err = nil, want err not nil")
	}
}

func TestListCategoriesErrorUnmarshalling(t *testing.T) {
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.category.ls", "-json").Return(*bytes.NewBufferString("invalid"), nil)

	_, err := g.ListCategories(ctx)
	if err == nil {
		t.Fatal("Govc.ListCategories() err = nil, want err not nil")
	}
}

func TestCreateCategoryForVMSuccess(t *testing.T) {
	category := "category"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.category.create", "-t", "VirtualMachine", category).Return(*bytes.NewBufferString(""), nil)

	err := g.CreateCategoryForVM(ctx, category)
	if err != nil {
		t.Fatalf("Govc.CreateCategoryForVM() err = %v, want err nil", err)
	}
}

func TestCreateCategoryForVMError(t *testing.T) {
	category := "category"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.category.create", "-t", "VirtualMachine", category).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	err := g.CreateCategoryForVM(ctx, category)
	if err == nil {
		t.Fatal("Govc.CreateCategoryForVM() err = nil, want err not nil")
	}
}

func TestImportTemplateSuccess(t *testing.T) {
	ovaURL := "ovaURL"
	name := "name"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.import", "-k", "-pull", "-n", name, templateLibrary, ovaURL).Return(*bytes.NewBufferString(""), nil)

	if err := g.ImportTemplate(ctx, templateLibrary, ovaURL, name); err != nil {
		t.Fatalf("Govc.ImportTemplate() err = %v, want err nil", err)
	}
}

func TestImportTemplateError(t *testing.T) {
	ovaURL := "ovaURL"
	name := "name"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.import", "-k", "-pull", "-n", name, templateLibrary, ovaURL).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	if err := g.ImportTemplate(ctx, templateLibrary, ovaURL, name); err == nil {
		t.Fatal("Govc.ImportTemplate() err = nil, want err not nil")
	}
}

func TestDeleteTemplateSuccess(t *testing.T) {
	template := "template"
	resourcePool := "resourcePool"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "vm.markasvm", "-pool", resourcePool, template).Return(*bytes.NewBufferString(""), nil)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "snapshot.remove", "-vm", template, "*").Return(*bytes.NewBufferString(""), nil)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "vm.destroy", template).Return(*bytes.NewBufferString(""), nil)

	if err := g.DeleteTemplate(ctx, resourcePool, template); err != nil {
		t.Fatalf("Govc.DeleteTemplate() err = %v, want err nil", err)
	}
}

func TestDeleteTemplateMarkAsVMError(t *testing.T) {
	template := "template"
	resourcePool := "resourcePool"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "vm.markasvm", "-pool", resourcePool, template).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	if err := g.DeleteTemplate(ctx, resourcePool, template); err == nil {
		t.Fatal("Govc.DeleteTemplate() err = nil, want err not nil")
	}
}

func TestDeleteTemplateRemoveSnapshotError(t *testing.T) {
	template := "template"
	resourcePool := "resourcePool"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "vm.markasvm", "-pool", resourcePool, template).Return(*bytes.NewBufferString(""), nil)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "snapshot.remove", "-vm", template, "*").Return(bytes.Buffer{}, errors.New("error from execute with env"))

	if err := g.DeleteTemplate(ctx, resourcePool, template); err == nil {
		t.Fatal("Govc.DeleteTemplate() err = nil, want err not nil")
	}
}

func TestDeleteTemplateDeleteVMError(t *testing.T) {
	template := "template"
	resourcePool := "resourcePool"
	ctx := context.Background()

	g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "vm.markasvm", "-pool", resourcePool, template).Return(*bytes.NewBufferString(""), nil)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "snapshot.remove", "-vm", template, "*").Return(*bytes.NewBufferString(""), nil)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "vm.destroy", template).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	if err := g.DeleteTemplate(ctx, resourcePool, template); err == nil {
		t.Fatal("Govc.DeleteTemplate() err = nil, want err not nil")
	}
}

func TestGovcLogoutSuccess(t *testing.T) {
	ctx := context.Background()
	g, executable, env := setup(t)

	executable.EXPECT().ExecuteWithEnv(ctx, env, "session.logout").Return(*bytes.NewBufferString(""), nil)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "session.logout", "-k").Return(*bytes.NewBufferString(""), nil)

	if err := g.Logout(ctx); err != nil {
		t.Fatalf("Govc.Logout() err = %v, want err nil", err)
	}
}

func TestGovcValidateVCenterConnectionSuccess(t *testing.T) {
	ctx := context.Background()
	ts := newHTTPSServer(t)
	g, _, _ := setup(t)

	if err := g.ValidateVCenterConnection(ctx, strings.TrimPrefix(ts.URL, "https://")); err != nil {
		t.Fatalf("Govc.ValidateVCenterConnection() err = %v, want err nil", err)
	}
}

func TestGovcValidateVCenterAuthenticationSuccess(t *testing.T) {
	ctx := context.Background()
	g, executable, env := setup(t)

	executable.EXPECT().ExecuteWithEnv(ctx, env, "about", "-k").Return(*bytes.NewBufferString(""), nil)

	if err := g.ValidateVCenterAuthentication(ctx); err != nil {
		t.Fatalf("Govc.ValidateVCenterAuthentication() err = %v, want err nil", err)
	}
}

func TestGovcIsCertSelfSignedTrue(t *testing.T) {
	ctx := context.Background()
	g, executable, env := setup(t)

	executable.EXPECT().ExecuteWithEnv(ctx, env, "about").Return(*bytes.NewBufferString(""), errors.New(""))

	if !g.IsCertSelfSigned(ctx) {
		t.Fatalf("Govc.IsCertSelfSigned) = false, want true")
	}
}

func TestGovcIsCertSelfSignedFalse(t *testing.T) {
	ctx := context.Background()
	g, executable, env := setup(t)

	executable.EXPECT().ExecuteWithEnv(ctx, env, "about").Return(*bytes.NewBufferString(""), nil)

	if g.IsCertSelfSigned(ctx) {
		t.Fatalf("Govc.IsCertSelfSigned) = true, want false")
	}
}

func TestGovcGetCertThumbprintSuccess(t *testing.T) {
	ctx := context.Background()
	g, executable, env := setup(t)
	wantThumbprint := "AB:AB:AB"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "about.cert", "-thumbprint", "-k").Return(*bytes.NewBufferString("server.com AB:AB:AB"), nil)

	gotThumbprint, err := g.GetCertThumbprint(ctx)
	if err != nil {
		t.Fatalf("Govc.GetCertThumbprint() err = %v, want err nil", err)
	}

	if gotThumbprint != wantThumbprint {
		t.Fatalf("Govc.GetCertThumbprint() thumbprint = %s, want %s", gotThumbprint, wantThumbprint)
	}
}

func TestGovcGetCertThumbprintBadOutput(t *testing.T) {
	ctx := context.Background()
	g, executable, env := setup(t)
	wantErr := "invalid thumbprint format"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "about.cert", "-thumbprint", "-k").Return(*bytes.NewBufferString("server.comAB:AB:AB"), nil)

	if _, err := g.GetCertThumbprint(ctx); err == nil || err.Error() != wantErr {
		t.Fatalf("Govc.GetCertThumbprint() err = %s, want err %s", err, wantErr)
	}
}

func TestGovcConfigureCertThumbprint(t *testing.T) {
	ctx := context.Background()
	g, _, _ := setup(t)
	server := "server.com"
	thumbprint := "AB:AB:AB"
	wantKnownHostsContent := "server.com AB:AB:AB"

	if err := g.ConfigureCertThumbprint(ctx, server, thumbprint); err != nil {
		t.Fatalf("Govc.ConfigureCertThumbprint() err = %v, want err nil", err)
	}

	path, ok := os.LookupEnv("GOVC_TLS_KNOWN_HOSTS")
	if !ok {
		t.Fatal("GOVC_TLS_KNOWN_HOSTS is not set")
	}

	gotKnownHostsContent := test.ReadFile(t, path)
	if gotKnownHostsContent != wantKnownHostsContent {
		t.Fatalf("GOVC_TLS_KNOWN_HOSTS file content = %s, want %s", gotKnownHostsContent, wantKnownHostsContent)
	}
}

func TestGovcDatacenterExistsTrue(t *testing.T) {
	ctx := context.Background()
	g, executable, env := setup(t)
	datacenter := "datacenter_1"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "datacenter.info", datacenter).Return(*bytes.NewBufferString(""), nil)

	exists, err := g.DatacenterExists(ctx, datacenter)
	if err != nil {
		t.Fatalf("Govc.DatacenterExists() err = %v, want err nil", err)
	}

	if !exists {
		t.Fatalf("Govc.DatacenterExists() = false, want true")
	}
}

func TestGovcDatacenterExistsFalse(t *testing.T) {
	ctx := context.Background()
	g, executable, env := setup(t)
	datacenter := "datacenter_1"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "datacenter.info", datacenter).Return(*bytes.NewBufferString("datacenter_1 not found"), errors.New("exit code 1"))

	exists, err := g.DatacenterExists(ctx, datacenter)
	if err != nil {
		t.Fatalf("Govc.DatacenterExists() err = %v, want err nil", err)
	}

	if exists {
		t.Fatalf("Govc.DatacenterExists() = true, want false")
	}
}

func TestGovcNetworkExistsTrue(t *testing.T) {
	ctx := context.Background()
	g, executable, env := setup(t)
	network := "/Networks/network_1"
	networkName := "network_1"
	networkDir := "/Networks"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "find", "-maxdepth=1", networkDir, "-type", "n", "-name", networkName).Return(*bytes.NewBufferString(network), nil)

	exists, err := g.NetworkExists(ctx, network)
	if err != nil {
		t.Fatalf("Govc.NetworkExists() err = %v, want err nil", err)
	}

	if !exists {
		t.Fatalf("Govc.NetworkExists() = false, want true")
	}
}

func TestGovcNetworkExistsFalse(t *testing.T) {
	ctx := context.Background()
	g, executable, env := setup(t)
	network := "/Networks/network_1"
	networkName := "network_1"
	networkDir := "/Networks"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "find", "-maxdepth=1", networkDir, "-type", "n", "-name", networkName).Return(*bytes.NewBufferString(""), nil)

	exists, err := g.NetworkExists(ctx, network)
	if err != nil {
		t.Fatalf("Govc.NetworkExistsNetworkExists() err = %v, want err nil", err)
	}

	if exists {
		t.Fatalf("Govc.NetworkExists() = true, want false")
	}
}

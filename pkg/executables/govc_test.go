package executables_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

const (
	govcUsername       = "GOVC_USERNAME"
	govcPassword       = "GOVC_PASSWORD"
	govcURL            = "GOVC_URL"
	govcDatacenter     = "GOVC_DATACENTER"
	govcInsecure       = "GOVC_INSECURE"
	vSphereUsername    = "EKSA_VSPHERE_USERNAME"
	vSpherePassword    = "EKSA_VSPHERE_PASSWORD"
	vSphereServer      = "VSPHERE_SERVER"
	templateLibrary    = "eks-a-templates"
	expectedDeployOpts = `{"DiskProvisioning":"thin","NetworkMapping":[{"Name":"nic0","Network":"/SDDC-Datacenter/network/sddc-cgw-network-1"},{"Name":"VM Network","Network":"/SDDC-Datacenter/network/sddc-cgw-network-1"}]}`
)

var govcEnvironment = map[string]string{
	govcUsername:   "vsphere_username",
	govcPassword:   "vsphere_password",
	govcURL:        "vsphere_server",
	govcDatacenter: "vsphere_datacenter",
	govcInsecure:   "false",
}

func setupContext(t *testing.T) {
	t.Setenv(vSphereUsername, "vsphere_username")
	t.Setenv(vSpherePassword, "vsphere_password")
	t.Setenv(vSphereServer, "vsphere_server")
	t.Setenv(govcUsername, os.Getenv(vSphereUsername))
	t.Setenv(govcPassword, os.Getenv(vSpherePassword))
	t.Setenv(govcURL, os.Getenv(vSphereServer))
	t.Setenv(govcInsecure, "false")
	t.Setenv(govcDatacenter, "vsphere_datacenter")
}

func setup(t *testing.T, opts ...executables.GovcOpt) (dir string, govc *executables.Govc, mockExecutable *mockexecutables.MockExecutable, env map[string]string) {
	setupContext(t)
	dir, writer := test.NewWriter(t)
	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	g := executables.NewGovc(executable, writer, opts...)

	return dir, g, executable, govcEnvironment
}

type deployTemplateTest struct {
	govc                     *executables.Govc
	mockExecutable           *mockexecutables.MockExecutable
	env                      map[string]string
	datacenter               string
	datastore                string
	dir                      string
	network                  string
	resourcePool             string
	templatePath             string
	ovaURL                   string
	deployFolder             string
	templateInLibraryPathAbs string
	templateName             string
	diskName                 string
	diskSize                 int
	resizeDisk2              bool
	ctx                      context.Context
	fakeExecResponse         *bytes.Buffer
	expectations             []*gomock.Call
}

func newDeployTemplateTest(t *testing.T) *deployTemplateTest {
	dir, g, exec, env := setup(t)
	return &deployTemplateTest{
		govc:                     g,
		mockExecutable:           exec,
		env:                      env,
		datacenter:               "SDDC-Datacenter",
		datastore:                "/SDDC-Datacenter/datastore/WorkloadDatastore",
		dir:                      dir,
		network:                  "/SDDC-Datacenter/network/sddc-cgw-network-1",
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
		diskName:                 "disk-31000-1",
		diskSize:                 20,
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

func (dt *deployTemplateTest) expectDevicesInfoToReturn(err error) {
	dt.expectations = append(
		dt.expectations,
		dt.mockExecutable.EXPECT().ExecuteWithEnv(dt.ctx, dt.env, "device.info", "-dc", dt.datacenter, "-vm", dt.templateName, "-json").Return(*dt.fakeExecResponse, err),
	)
}

func (dt *deployTemplateTest) expectResizeDiskToReturn(err error) {
	dt.expectations = append(
		dt.expectations,
		dt.mockExecutable.EXPECT().ExecuteWithEnv(dt.ctx, dt.env, "vm.disk.change", "-dc", dt.datacenter, "-vm", dt.templateName, "-disk.name", dt.diskName, "-size", strconv.Itoa(dt.diskSize)+"G").Return(*dt.fakeExecResponse, err),
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
	return dt.govc.DeployTemplateFromLibrary(dt.ctx, dt.deployFolder, dt.templateName, templateLibrary, dt.datacenter, dt.datastore, dt.network, dt.resourcePool, dt.resizeDisk2)
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

func (dt *deployTemplateTest) assertDeployOptsMatches(t *testing.T) {
	g := NewWithT(t)

	actual, err := os.ReadFile(filepath.Join(dt.dir, executables.DeployOptsFile))
	if err != nil {
		t.Fatalf("failed to read deploy options file: %v", err)
	}

	g.Expect(string(actual)).To(Equal(expectedDeployOpts))
}

func TestSearchTemplateItExists(t *testing.T) {
	ctx := context.Background()
	template := "my-template"
	datacenter := "SDDC-Datacenter"

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "find", "-json", "/"+datacenter, "-type", "VirtualMachine", "-name", filepath.Base(template)).Return(*bytes.NewBufferString("[\"/SDDC Datacenter/vm/Templates/ubuntu 2004-kube-v1.19.6\"]"), nil)

	_, err := g.SearchTemplate(ctx, datacenter, template)
	if err != nil {
		t.Fatalf("Govc.SearchTemplate() exists = false, want true %v", err)
	}
}

func TestSearchTemplateItDoesNotExists(t *testing.T) {
	template := "my-template"
	ctx := context.Background()
	datacenter := "SDDC-Datacenter"

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "find", "-json", "/"+datacenter, "-type", "VirtualMachine", "-name", filepath.Base(template)).Return(*bytes.NewBufferString(""), nil)

	templateFullPath, err := g.SearchTemplate(ctx, datacenter, template)
	if err == nil && len(templateFullPath) > 0 {
		t.Fatalf("Govc.SearchTemplate() exists = true, want false %v", err)
	}
}

func TestSearchTemplateError(t *testing.T) {
	template := "my-template"
	ctx := context.Background()
	datacenter := "SDDC-Datacenter"

	_, g, executable, env := setup(t)
	g.Retrier = retrier.NewWithMaxRetries(5, 0)
	executable.EXPECT().ExecuteWithEnv(ctx, env, gomock.Any()).Return(bytes.Buffer{}, errors.New("error from execute with env")).Times(5)

	_, err := g.SearchTemplate(ctx, datacenter, template)
	if err == nil {
		t.Fatal("Govc.SearchTemplate() err = nil, want err not nil")
	}
}

func TestLibraryElementExistsItExists(t *testing.T) {
	ctx := context.Background()

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.info", "-json", libraryElement).Return(*bytes.NewBufferString(response), nil)

	_, err := g.GetLibraryElementContentVersion(ctx, libraryElement)
	if err != nil {
		t.Fatalf("Govc.GetLibraryElementContentVersion() err = %v, want err nil", err)
	}
}

func TestGetLibraryElementContentVersionError(t *testing.T) {
	ctx := context.Background()
	libraryElement := "/eks-a-templates/ubuntu-2004-kube-v1.19.6"

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.info", "-json", libraryElement).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	_, err := g.GetLibraryElementContentVersion(ctx, libraryElement)
	if err == nil {
		t.Fatal("Govc.GetLibraryElementContentVersion() err = nil, want err not nil")
	}
}

func TestDeleteLibraryElementSuccess(t *testing.T) {
	ctx := context.Background()
	libraryElement := "/eks-a-templates/ubuntu-2004-kube-v1.19.6"

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.rm", libraryElement).Return(*bytes.NewBufferString(""), nil)

	err := g.DeleteLibraryElement(ctx, libraryElement)
	if err != nil {
		t.Fatalf("Govc.DeleteLibraryElement() err = %v, want err nil", err)
	}
}

func TestDeleteLibraryElementError(t *testing.T) {
	ctx := context.Background()
	libraryElement := "/eks-a-templates/ubuntu-2004-kube-v1.19.6"

	_, g, executable, env := setup(t)
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

	setupContext(t)

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

			setupContext(t)

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
	tt.assertDeployOptsMatches(t)
}

func TestDeployTemplateFromLibraryResizeBRSuccess(t *testing.T) {
	tt := newDeployTemplateTest(t)
	tt.resizeDisk2 = true
	response := map[string][]interface{}{
		"Devices": {
			map[string]interface{}{
				"Name": "disk-31000-0",
				"DeviceInfo": map[string]string{
					"Label": "Hard disk 1",
				},
			},
			map[string]interface{}{
				"Name": "disk-31000-1",
				"DeviceInfo": map[string]string{
					"Label": "Hard disk 2",
				},
			},
		},
	}
	mashaledResponse, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	responseBytes := bytes.NewBuffer(mashaledResponse)
	tt.fakeExecResponse = responseBytes
	tt.expectFolderInfoToReturn(nil)
	tt.expectDeployToReturn(nil)
	tt.expectDevicesInfoToReturn(nil)
	tt.expectResizeDiskToReturn(nil)
	tt.expectCreateSnapshotToReturn(nil)
	tt.expectMarkAsTemplateToReturn(nil)

	tt.assertDeployTemplateSuccess(t)
	tt.assertDeployOptsMatches(t)
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

	setupContext(t)

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

	setupContext(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)

	var params []string
	params = []string{"find", "/" + env[govcDatacenter], "-type", "VirtualMachine", "-name", clusterName + "*"}
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

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.create", "-ds", datastore, templateLibrary).Return(*bytes.NewBufferString("testing"), nil)

	err := g.CreateLibrary(ctx, datastore, templateLibrary)
	if err != nil {
		t.Fatalf("Govc.CreateLibrary() err = %v, want err nil", err)
	}
}

func TestCreateLibraryError(t *testing.T) {
	datastore := "/SDDC-Datacenter/datastore/WorkloadDatastore"
	ctx := context.Background()

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.create", "-ds", datastore, templateLibrary).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	err := g.CreateLibrary(ctx, datastore, templateLibrary)
	if err == nil {
		t.Fatal("Govc.CreateLibrary() err = nil, want err not nil")
	}
}

func TestGetTagsSuccessNoTags(t *testing.T) {
	path := "/SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6"
	ctx := context.Background()

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.attached.ls", "-json", "-r", path).Return(*bytes.NewBufferString("invalid"), nil)

	_, err := g.GetTags(ctx, path)
	if err == nil {
		t.Fatal("Govc.GetTags() err = nil, want err not nil")
	}
}

func TestListTagsSuccessNoTags(t *testing.T) {
	ctx := context.Background()

	_, g, executable, env := setup(t)
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

	wantTags := []executables.Tag{
		{
			Name:       "eksd:1.19-4",
			Id:         "urn:vmomi:InventoryServiceTag:5555:GLOBAL",
			CategoryId: "eksd",
		},
		{
			Name:       "kubernetesChannel:1.19",
			Id:         "urn:vmomi:InventoryServiceTag:5555:GLOBAL",
			CategoryId: "kubernetesChannel",
		},
	}

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.ls", "-json").Return(bytes.Buffer{}, errors.New("error from exec"))

	_, err := g.ListTags(ctx)
	if err == nil {
		t.Fatal("Govc.ListTags() err = nil, want err not nil")
	}
}

func TestListTagsErrorUnmarshalling(t *testing.T) {
	ctx := context.Background()

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.attach", tag, path).Return(*bytes.NewBufferString(""), nil)

	err := g.AddTag(ctx, path, tag)
	if err != nil {
		t.Fatalf("Govc.AddTag() err = %v, want err nil", err)
	}
}

func TestEnvMapOverride(t *testing.T) {
	category := "category"
	tag := "tag"
	ctx := context.Background()

	envOverride := map[string]string{
		govcUsername:   "override_vsphere_username",
		govcPassword:   "override_vsphere_password",
		govcURL:        "override_vsphere_server",
		govcDatacenter: "override_vsphere_datacenter",
		govcInsecure:   "false",
	}

	_, g, executable, _ := setup(t, executables.WithGovcEnvMap(envOverride))
	executable.EXPECT().ExecuteWithEnv(ctx, envOverride, "tags.create", "-c", category, tag).Return(*bytes.NewBufferString(""), nil)

	err := g.CreateTag(ctx, tag, category)
	if err != nil {
		t.Fatalf("Govc.CreateTag() with envMap override err = %v, want err nil", err)
	}
}

func TestAddTagError(t *testing.T) {
	tag := "tag"
	path := "/SDDC-Datacenter/vm/Templates/ubuntu-2004-kube-v1.19.6"
	ctx := context.Background()

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.create", "-c", category, tag).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	err := g.CreateTag(ctx, tag, category)
	if err == nil {
		t.Fatal("Govc.CreateTag() err = nil, want err not nil")
	}
}

func TestListCategoriesSuccessNoCategories(t *testing.T) {
	ctx := context.Background()

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.category.ls", "-json").Return(bytes.Buffer{}, errors.New("error from exec"))

	_, err := g.ListCategories(ctx)
	if err == nil {
		t.Fatal("Govc.ListCategories() err = nil, want err not nil")
	}
}

func TestListCategoriesErrorUnmarshalling(t *testing.T) {
	ctx := context.Background()

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.category.ls", "-json").Return(*bytes.NewBufferString("invalid"), nil)

	_, err := g.ListCategories(ctx)
	if err == nil {
		t.Fatal("Govc.ListCategories() err = nil, want err not nil")
	}
}

func TestCreateCategoryForVMSuccess(t *testing.T) {
	category := "category"
	ctx := context.Background()

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "tags.category.create", "-t", "VirtualMachine", category).Return(*bytes.NewBufferString(""), nil)

	err := g.CreateCategoryForVM(ctx, category)
	if err != nil {
		t.Fatalf("Govc.CreateCategoryForVM() err = %v, want err nil", err)
	}
}

func TestCreateCategoryForVMError(t *testing.T) {
	category := "category"
	ctx := context.Background()

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.import", "-k", "-pull", "-n", name, templateLibrary, ovaURL).Return(*bytes.NewBufferString(""), nil)

	if err := g.ImportTemplate(ctx, templateLibrary, ovaURL, name); err != nil {
		t.Fatalf("Govc.ImportTemplate() err = %v, want err nil", err)
	}
}

func TestImportTemplateError(t *testing.T) {
	ovaURL := "ovaURL"
	name := "name"
	ctx := context.Background()

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "library.import", "-k", "-pull", "-n", name, templateLibrary, ovaURL).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	if err := g.ImportTemplate(ctx, templateLibrary, ovaURL, name); err == nil {
		t.Fatal("Govc.ImportTemplate() err = nil, want err not nil")
	}
}

func TestDeleteTemplateSuccess(t *testing.T) {
	template := "template"
	resourcePool := "resourcePool"
	ctx := context.Background()

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "vm.markasvm", "-pool", resourcePool, template).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	if err := g.DeleteTemplate(ctx, resourcePool, template); err == nil {
		t.Fatal("Govc.DeleteTemplate() err = nil, want err not nil")
	}
}

func TestDeleteTemplateRemoveSnapshotError(t *testing.T) {
	template := "template"
	resourcePool := "resourcePool"
	ctx := context.Background()

	_, g, executable, env := setup(t)
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

	_, g, executable, env := setup(t)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "vm.markasvm", "-pool", resourcePool, template).Return(*bytes.NewBufferString(""), nil)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "snapshot.remove", "-vm", template, "*").Return(*bytes.NewBufferString(""), nil)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "vm.destroy", template).Return(bytes.Buffer{}, errors.New("error from execute with env"))

	if err := g.DeleteTemplate(ctx, resourcePool, template); err == nil {
		t.Fatal("Govc.DeleteTemplate() err = nil, want err not nil")
	}
}

func TestGovcLogoutSuccess(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)

	executable.EXPECT().ExecuteWithEnv(ctx, env, "session.logout").Return(*bytes.NewBufferString(""), nil)
	executable.EXPECT().ExecuteWithEnv(ctx, env, "session.logout", "-k").Return(*bytes.NewBufferString(""), nil)

	if err := g.Logout(ctx); err != nil {
		t.Fatalf("Govc.Logout() err = %v, want err nil", err)
	}
}

func TestGovcValidateVCenterConnectionSuccess(t *testing.T) {
	ctx := context.Background()
	ts := newHTTPSServer(t)
	_, g, _, _ := setup(t)

	if err := g.ValidateVCenterConnection(ctx, strings.TrimPrefix(ts.URL, "https://")); err != nil {
		t.Fatalf("Govc.ValidateVCenterConnection() err = %v, want err nil", err)
	}
}

func TestGovcValidateVCenterAuthenticationSuccess(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)

	executable.EXPECT().ExecuteWithEnv(ctx, env, "about", "-k").Return(*bytes.NewBufferString(""), nil)

	if err := g.ValidateVCenterAuthentication(ctx); err != nil {
		t.Fatalf("Govc.ValidateVCenterAuthentication() err = %v, want err nil", err)
	}
}

func TestGovcValidateVCenterAuthenticationErrorNoDatacenter(t *testing.T) {
	ctx := context.Background()
	_, g, _, _ := setup(t)

	t.Setenv(govcDatacenter, "")

	if err := g.ValidateVCenterAuthentication(ctx); err == nil {
		t.Fatal("Govc.ValidateVCenterAuthentication() err = nil, want err not nil")
	}
}

func TestGovcIsCertSelfSignedTrue(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)

	executable.EXPECT().ExecuteWithEnv(ctx, env, "about").Return(*bytes.NewBufferString(""), errors.New(""))

	if !g.IsCertSelfSigned(ctx) {
		t.Fatalf("Govc.IsCertSelfSigned) = false, want true")
	}
}

func TestGovcIsCertSelfSignedFalse(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)

	executable.EXPECT().ExecuteWithEnv(ctx, env, "about").Return(*bytes.NewBufferString(""), nil)

	if g.IsCertSelfSigned(ctx) {
		t.Fatalf("Govc.IsCertSelfSigned) = true, want false")
	}
}

func TestGovcGetCertThumbprintSuccess(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
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
	_, g, executable, env := setup(t)
	wantErr := "invalid thumbprint format"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "about.cert", "-thumbprint", "-k").Return(*bytes.NewBufferString("server.comAB:AB:AB"), nil)

	if _, err := g.GetCertThumbprint(ctx); err == nil || err.Error() != wantErr {
		t.Fatalf("Govc.GetCertThumbprint() err = %s, want err %s", err, wantErr)
	}
}

func TestGovcConfigureCertThumbprint(t *testing.T) {
	ctx := context.Background()
	_, g, _, _ := setup(t)
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
	_, g, executable, env := setup(t)
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
	_, g, executable, env := setup(t)
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
	_, g, executable, env := setup(t)
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
	_, g, executable, env := setup(t)
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

func TestGovcCreateUser(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	username := "ralph"
	password := "verysecret"

	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "test CreateGroup success",
			wantErr: nil,
		},
		{
			name:    "test CreateGroup error",
			wantErr: errors.New("operation failed"),
		},
	}

	for _, tt := range tests {

		executable.EXPECT().ExecuteWithEnv(ctx, env, "sso.user.create", "-p", password, username).Return(*bytes.NewBufferString(""), tt.wantErr)

		err := g.CreateUser(ctx, username, password)
		gt := NewWithT(t)

		if tt.wantErr != nil {
			gt.Expect(err).ToNot(BeNil())
		} else {
			gt.Expect(err).To(BeNil())
		}
	}
}

func TestGovcCreateGroup(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	group := "EKSA"

	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "test CreateGroup success",
			wantErr: nil,
		},
		{
			name:    "test CreateGroup error",
			wantErr: errors.New("operation failed"),
		},
	}

	for _, tt := range tests {

		executable.EXPECT().ExecuteWithEnv(ctx, env, "sso.group.create", group).Return(*bytes.NewBufferString(""), tt.wantErr)

		err := g.CreateGroup(ctx, group)
		gt := NewWithT(t)
		if tt.wantErr != nil {
			gt.Expect(err).ToNot(BeNil())
		} else {
			gt.Expect(err).To(BeNil())
		}

	}
}

func TestGovcUserExistsFalse(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	username := "eksa"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "sso.user.ls", username).Return(*bytes.NewBufferString(""), nil)

	exists, err := g.UserExists(ctx, username)
	gt := NewWithT(t)
	gt.Expect(err).To(BeNil())
	gt.Expect(exists).To(BeFalse())
}

func TestGovcUserExistsTrue(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	username := "eksa"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "sso.user.ls", username).Return(*bytes.NewBufferString(username), nil)

	exists, err := g.UserExists(ctx, username)
	gt := NewWithT(t)
	gt.Expect(err).To(BeNil())
	gt.Expect(exists).To(BeTrue())
}

func TestGovcUserExistsError(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	username := "eksa"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "sso.user.ls", username).Return(*bytes.NewBufferString(""), errors.New("operation failed"))

	_, err := g.UserExists(ctx, username)
	gt := NewWithT(t)
	gt.Expect(err).ToNot(BeNil())
}

func TestGovcCreateRole(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	role := "EKSACloudAdmin"
	privileges := []string{"vSphereDataProtection.Recovery", "vSphereDataProtection.Protection"}

	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "test CreateRole success",
			wantErr: nil,
		},
		{
			name:    "test CreateRole error",
			wantErr: errors.New("operation failed"),
		},
	}

	for _, tt := range tests {

		targetArgs := append([]string{"role.create", role}, privileges...)
		executable.EXPECT().ExecuteWithEnv(ctx, env, targetArgs).Return(*bytes.NewBufferString(""), tt.wantErr)

		err := g.CreateRole(ctx, role, privileges)
		gt := NewWithT(t)
		if tt.wantErr != nil {
			gt.Expect(err).ToNot(BeNil())
		} else {
			gt.Expect(err).To(BeNil())
		}
	}
}

func TestGovcGroupExistsFalse(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	group := "EKSA"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "sso.group.ls", group).Return(*bytes.NewBufferString(""), nil)

	exists, err := g.GroupExists(ctx, group)
	gt := NewWithT(t)
	gt.Expect(err).To(BeNil())
	gt.Expect(exists).To(BeFalse())
}

func TestGovcGroupExistsTrue(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	group := "EKSA"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "sso.group.ls", group).Return(*bytes.NewBufferString(group), nil)

	exists, err := g.GroupExists(ctx, group)
	gt := NewWithT(t)
	gt.Expect(err).To(BeNil())
	gt.Expect(exists).To(BeTrue())
}

func TestGovcGroupExistsError(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	group := "EKSA"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "sso.group.ls", group).Return(*bytes.NewBufferString(""), errors.New("operation failed"))

	_, err := g.GroupExists(ctx, group)
	gt := NewWithT(t)
	gt.Expect(err).ToNot(BeNil())
}

func TestGovcRoleExistsTrue(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	role := "EKSACloudAdmin"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "role.ls", role).Return(*bytes.NewBufferString(role), nil)

	exists, err := g.RoleExists(ctx, role)
	gt := NewWithT(t)
	gt.Expect(err).To(BeNil())
	gt.Expect(exists).To(BeTrue())
}

func TestGovcRoleExistsFalse(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	role := "EKSACloudAdmin"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "role.ls", role).Return(*bytes.NewBufferString(""), fmt.Errorf("role \"%s\" not found", role))

	exists, err := g.RoleExists(ctx, role)
	gt := NewWithT(t)
	gt.Expect(err).To(BeNil())
	gt.Expect(exists).To(BeFalse())
}

func TestGovcRoleExistsError(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	role := "EKSACloudAdmin"

	executable.EXPECT().ExecuteWithEnv(ctx, env, "role.ls", role).Return(*bytes.NewBufferString(""), errors.New("operation failed"))

	_, err := g.RoleExists(ctx, role)
	gt := NewWithT(t)
	gt.Expect(err).ToNot(BeNil())
}

func TestGovcAddUserToGroup(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	group := "EKSA"
	username := "ralph"

	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "test AddUserToGroup success",
			wantErr: nil,
		},
		{
			name:    "test AddUserToGroup error",
			wantErr: errors.New("operation failed"),
		},
	}

	for _, tt := range tests {

		executable.EXPECT().ExecuteWithEnv(ctx, env, "sso.group.update", "-a", username, group).Return(*bytes.NewBufferString(""), tt.wantErr)

		err := g.AddUserToGroup(ctx, group, username)
		gt := NewWithT(t)
		if tt.wantErr != nil {
			gt.Expect(err).ToNot(BeNil())
		} else {
			gt.Expect(err).To(BeNil())
		}

	}
}

func TestGovcSetGroupRoleOnObject(t *testing.T) {
	ctx := context.Background()
	_, g, executable, env := setup(t)
	principal := "EKSAGroup"
	domain := "vsphere.local"
	role := "EKSACloudAdmin"
	object := "/Datacenter/vm/MyVirtualMachines"

	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "test SetGroupRoleOnObject success",
			wantErr: nil,
		},
		{
			name:    "test SetGroupRoleOnObject error",
			wantErr: errors.New("operation failed"),
		},
	}

	for _, tt := range tests {
		executable.EXPECT().ExecuteWithEnv(
			ctx,
			env,
			"permissions.set",
			"-group=true",
			"-principal",
			principal+"@"+domain,
			"-role",
			role,
			object,
		).Return(*bytes.NewBufferString(""), tt.wantErr)

		err := g.SetGroupRoleOnObject(ctx, principal, role, object, domain)
		gt := NewWithT(t)
		if tt.wantErr != nil {
			gt.Expect(err).ToNot(BeNil())
		} else {
			gt.Expect(err).To(BeNil())
		}
	}
}

func TestGovcGetVMDiskSizeInGB(t *testing.T) {
	datacenter := "SDDC-Datacenter"
	template := "bottlerocket-kube-v1.24.6"
	ctx := context.Background()
	_, g, executable, env := setup(t)
	gt := NewWithT(t)

	response := map[string][]interface{}{
		"Devices": {
			map[string]interface{}{
				"Name": "disk-31000-0",
				"DeviceInfo": map[string]string{
					"Label": "Hard disk 1",
				},
				"CapacityInKB": 25 * 1024 * 1024, // 25GB in KB
			},
		},
	}

	mashaledResponse, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	responseBytes := bytes.NewBuffer(mashaledResponse)

	executable.EXPECT().ExecuteWithEnv(ctx, env, "device.info", "-dc", datacenter, "-vm", template, "-json", "disk-*").Return(*responseBytes, nil)

	size, err := g.GetVMDiskSizeInGB(ctx, template, datacenter)
	gt.Expect(err).To(BeNil())
	gt.Expect(size).To(Equal(25))
}

func TestGovcGetVMDiskSizeInGBError(t *testing.T) {
	datacenter := "SDDC-Datacenter"
	template := "bottlerocket-kube-v1.24.6"
	ctx := context.Background()
	_, g, executable, env := setup(t)
	govcErr := errors.New("error DevicesInfo()")

	tests := []struct {
		testName string
		response map[string][]interface{}
		govcErr  error
		wantErr  error
	}{
		{
			testName: "devices_info_govc_error",
			response: nil,
			govcErr:  govcErr,
			wantErr:  fmt.Errorf("getting disk size for vm %s: getting template device information: %v", template, govcErr),
		},
		{
			testName: "devices_info_no_devices",
			response: map[string][]interface{}{
				"Devices": {},
			},
			govcErr: nil,
			wantErr: fmt.Errorf("no disks found for vm %s", template),
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			gt := NewWithT(t)
			mashaledResponse, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("failed to marshal response: %v", err)
			}
			responseBytes := bytes.NewBuffer(mashaledResponse)

			executable.EXPECT().ExecuteWithEnv(ctx, env, "device.info", "-dc", datacenter, "-vm", template, "-json", "disk-*").Return(*responseBytes, tt.govcErr)

			_, err = g.GetVMDiskSizeInGB(ctx, template, datacenter)
			gt.Expect(err.Error()).To(Equal(tt.wantErr.Error()))
		})
	}
}

func TestGovcGetHardDiskSize(t *testing.T) {
	datacenter := "SDDC-Datacenter"
	template := "bottlerocket-kube-v1-21"
	ctx := context.Background()
	wantDiskMap := map[string]float64{
		"Hard disk 1": 2097152,
		"Hard disk 2": 20971520,
	}
	_, g, executable, env := setup(t)
	gt := NewWithT(t)

	response := map[string][]interface{}{
		"Devices": {
			map[string]interface{}{
				"Name": "disk-31000-0",
				"DeviceInfo": map[string]string{
					"Label": "Hard disk 1",
				},
				"CapacityInKB": 2097152,
			}, map[string]interface{}{
				"Name": "disk-31000-1",
				"DeviceInfo": map[string]string{
					"Label": "Hard disk 2",
				},
				"CapacityInKB": 20971520,
			},
		},
	}

	marshaledResponse, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	responseBytes := bytes.NewBuffer(marshaledResponse)

	executable.EXPECT().ExecuteWithEnv(ctx, env, "device.info", "-dc", datacenter, "-vm", template, "-json", "disk-*").Return(*responseBytes, nil)

	diskSizeMap, err := g.GetHardDiskSize(ctx, template, datacenter)
	gt.Expect(err).To(BeNil())
	gt.Expect(diskSizeMap).To(Equal(wantDiskMap))
}

func TestGovcGetHardDiskSizeError(t *testing.T) {
	datacenter := "SDDC-Datacenter"
	template := "bottlerocket-kube-v1-21"
	ctx := context.Background()
	_, g, executable, env := setup(t)
	govcErr := errors.New("error DevicesInfo()")

	tests := []struct {
		testName string
		response map[string][]interface{}
		govcErr  error
		wantErr  error
	}{
		{
			testName: "devices_info_govc_error",
			response: nil,
			govcErr:  govcErr,
			wantErr:  fmt.Errorf("getting hard disk sizes for vm %s: getting template device information: %v", template, govcErr),
		},
		{
			testName: "devices_info_no_devices",
			response: map[string][]interface{}{
				"Devices": {},
			},
			govcErr: nil,
			wantErr: fmt.Errorf("no hard disks found for vm %s", template),
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			gt := NewWithT(t)
			marshaledResponse, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("failed to marshal response: %v", err)
			}
			responseBytes := bytes.NewBuffer(marshaledResponse)
			executable.EXPECT().ExecuteWithEnv(ctx, env, "device.info", "-dc", datacenter, "-vm", template, "-json", "disk-*").Return(*responseBytes, tt.govcErr)
			_, err = g.GetHardDiskSize(ctx, template, datacenter)
			gt.Expect(err.Error()).To(Equal(tt.wantErr.Error()))
		})
	}
}

func TestGovcGetResourcePoolInfo(t *testing.T) {
	datacenter := "SDDC-Datacenter"
	resourcePool := "*/Resources/Test-ResourcePool"
	govcErr := errors.New("error PoolInfo()")
	ctx := context.Background()
	_, g, executable, env := setup(t)

	tests := []struct {
		testName    string
		response    string
		govcErr     error
		wantErr     error
		wantMemInfo map[string]int
	}{
		{
			testName: "pool_info_memory_limit_set",
			response: `Name: Test-ResourcePool
					Path: /SDDC-Datacenter/host/Cluster-1/Resources/Test-ResourcePool
					Mem Usage: 100MB (11.3%)
					Mem Shares: normal
					Mem Reservation: 0MB (expandable=true)
					Mem Limit: 1000MB`,
			govcErr:     nil,
			wantErr:     nil,
			wantMemInfo: map[string]int{executables.MemoryAvailable: 900},
		},
		{
			testName: "pool_info_memory_limit_unset",
			response: `Name: Test-ResourcePool
					Path: /SDDC-Datacenter/host/Cluster-1/Resources/Test-ResourcePool
					Mem Usage: 100MB (11.3%)
					Mem Shares: normal
					Mem Reservation: 0MB (expandable=true)
					Mem Limit: -1MB`,
			govcErr:     nil,
			wantErr:     nil,
			wantMemInfo: map[string]int{executables.MemoryAvailable: -1},
		},
		{
			testName: "pool_info_memory_usage_corrupt",
			response: `Name: Test-ResourcePool
					Mem Usage:corrupt-val
					Mem Limit:-1MB`,
			govcErr:     nil,
			wantErr:     fmt.Errorf("unable to obtain memory usage for resource pool corrupt-val: strconv.Atoi: parsing \"-\": invalid syntax"),
			wantMemInfo: nil,
		},
		{
			testName: "pool_info_memory_limit_corrupt",
			response: `Name: Test-ResourcePool
					Mem Usage:100
					Mem Limit:corrupt-val`,
			govcErr:     nil,
			wantErr:     fmt.Errorf("unable to obtain memory limit for resource pool corrupt-val: strconv.Atoi: parsing \"-\": invalid syntax"),
			wantMemInfo: nil,
		},
		{
			testName:    "pool_info_error",
			response:    "",
			govcErr:     govcErr,
			wantErr:     fmt.Errorf("getting resource pool information: %v", govcErr),
			wantMemInfo: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			gt := NewWithT(t)
			responseBytes := bytes.NewBuffer([]byte(tt.response))
			executable.EXPECT().ExecuteWithEnv(ctx, env, "pool.info", "-dc", datacenter, resourcePool).Return(*responseBytes, tt.govcErr)
			poolMemInfo, err := g.GetResourcePoolInfo(ctx, datacenter, resourcePool)
			if tt.wantErr != nil {
				gt.Expect(err.Error()).To(Equal(tt.wantErr.Error()))
			}
			gt.Expect(poolMemInfo).To(Equal(tt.wantMemInfo))
		})
	}
}

package templates_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/internal/templates"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/internal/templates/mocks"
)

type test struct {
	t               *testing.T
	datastore       string
	resourcePool    string
	templateLibrary string
	govc            *mocks.MockGovcClient
	factory         *templates.Factory
	ctx             context.Context
	dummyError      error
}

type createTest struct {
	*test
	datacenter        string
	machineConfig     *v1alpha1.VSphereMachineConfig
	templatePath      string
	templateName      string
	templateDir       string
	templateInLibrary string
	ovaURL            string
	tagsByCategory    map[string][]string
}

func newTest(t *testing.T) *test {
	ctrl := gomock.NewController(t)
	test := &test{
		t:               t,
		datastore:       "datastore",
		resourcePool:    "*/pool/",
		templateLibrary: "library",
		govc:            mocks.NewMockGovcClient(ctrl),
		ctx:             context.Background(),
		dummyError:      errors.New("error from govc"),
	}
	f := templates.NewFactory(
		test.govc,
		test.datastore,
		test.resourcePool,
		test.templateLibrary,
	)
	test.factory = f
	return test
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

func newCreateTest(t *testing.T) *createTest {
	test := newTest(t)
	return &createTest{
		test:              test,
		datacenter:        "SDDC-Datacenter",
		machineConfig:     newMachineConfig(t),
		templatePath:      "/SDDC-Datacenter/vm/Templates/ubuntu-v1.19.8-eks-d-1-19-4-eks-a-0.0.1.build.38-amd64",
		templateDir:       "/SDDC-Datacenter/vm/Templates",
		templateName:      "ubuntu-v1.19.8-eks-d-1-19-4-eks-a-0.0.1.build.38-amd64",
		templateInLibrary: "library/ubuntu-v1.19.8-eks-d-1-19-4-eks-a-0.0.1.build.38-amd64",
		ovaURL:            "https://amazonaws.com/artifacts/0.0.1/eks-distro/ova/1-19/1-19-4/ubuntu-v1.19.8-eks-d-1-19-4-eks-a-0.0.1.build.38-amd64.ova",
		tagsByCategory:    map[string][]string{},
	}
}

func (ct *createTest) createIfMissing() error {
	return ct.factory.CreateIfMissing(ct.ctx, ct.datacenter, ct.machineConfig, ct.ovaURL, ct.tagsByCategory)
}

func (ct *createTest) assertErrorFromCreateIfMissing() {
	if err := ct.createIfMissing(); err == nil {
		ct.t.Fatal("factory.CreateIfMissing() err = nil, want err not nil")
	}
}

func (ct *createTest) assertSuccessFromCreateIfMissing() {
	if err := ct.createIfMissing(); err != nil {
		ct.t.Fatalf("factory.CreateIfMissing() err = %v, want err = nil", err)
	}
}

func TestFactoryCreateIfMissingSearchTemplate(t *testing.T) {
	ct := newCreateTest(t)
	ct.govc.EXPECT().SearchTemplate(ct.ctx, ct.datacenter, ct.machineConfig).Return(ct.machineConfig.Spec.Template, nil)
	ct.assertSuccessFromCreateIfMissing()
}

func TestFactoryCreateIfMissingErrorSearchTemplate(t *testing.T) {
	ct := newCreateTest(t)
	ct.govc.EXPECT().SearchTemplate(ct.ctx, ct.datacenter, ct.machineConfig).Return("", ct.dummyError) // error getting template

	ct.assertErrorFromCreateIfMissing()
}

func TestFactoryCreateIfMissingErrorLibraryElementExists(t *testing.T) {
	ct := newCreateTest(t)
	ct.govc.EXPECT().SearchTemplate(ct.ctx, ct.datacenter, ct.machineConfig).Return("", nil) // template not present
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateLibrary).Return(false, ct.dummyError)

	ct.assertErrorFromCreateIfMissing()
}

func TestFactoryCreateIfMissingErrorCreateLibrary(t *testing.T) {
	ct := newCreateTest(t)
	ct.govc.EXPECT().SearchTemplate(ct.ctx, ct.datacenter, ct.machineConfig).Return("", nil) // template not present
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateLibrary).Return(false, nil)
	ct.govc.EXPECT().CreateLibrary(ct.ctx, ct.datastore, ct.templateLibrary).Return(ct.dummyError)

	ct.assertErrorFromCreateIfMissing()
}

func TestFactoryCreateIfMissingErrorTemplateExistsInLibrary(t *testing.T) {
	ct := newCreateTest(t)
	ct.govc.EXPECT().SearchTemplate(ct.ctx, ct.datacenter, ct.machineConfig).Return("", nil) // template not present
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateLibrary).Return(false, nil)
	ct.govc.EXPECT().CreateLibrary(ct.ctx, ct.datastore, ct.templateLibrary).Return(nil)
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateInLibrary).Return(false, ct.dummyError)

	ct.assertErrorFromCreateIfMissing()
}

func TestFactoryCreateIfMissingErrorImport(t *testing.T) {
	ct := newCreateTest(t)
	ct.govc.EXPECT().SearchTemplate(ct.ctx, ct.datacenter, ct.machineConfig).Return("", nil) // template not present
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateLibrary).Return(false, nil)
	ct.govc.EXPECT().CreateLibrary(ct.ctx, ct.datastore, ct.templateLibrary).Return(nil)
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateInLibrary).Return(false, nil)
	ct.govc.EXPECT().ImportTemplate(ct.ctx, ct.templateLibrary, ct.ovaURL, ct.templateName).Return(ct.dummyError)

	ct.assertErrorFromCreateIfMissing()
}

func TestFactoryCreateIfMissingErrorDeploy(t *testing.T) {
	ct := newCreateTest(t)
	ct.govc.EXPECT().SearchTemplate(ct.ctx, ct.datacenter, ct.machineConfig).Return("", nil) // template not present
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateLibrary).Return(false, nil)
	ct.govc.EXPECT().CreateLibrary(ct.ctx, ct.datastore, ct.templateLibrary).Return(nil)
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateInLibrary).Return(false, nil)
	ct.govc.EXPECT().ImportTemplate(ct.ctx, ct.templateLibrary, ct.ovaURL, ct.templateName).Return(nil)
	ct.govc.EXPECT().DeployTemplateFromLibrary(
		ct.ctx, ct.templateDir, ct.templateName, ct.templateLibrary, ct.resourcePool,
	).Return(ct.dummyError)

	ct.assertErrorFromCreateIfMissing()
}

func TestFactoryCreateIfMissingErrorFromTagFactory(t *testing.T) {
	ct := newCreateTest(t)
	ct.govc.EXPECT().SearchTemplate(ct.ctx, ct.datacenter, ct.machineConfig).Return("", nil) // template not present
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateLibrary).Return(false, nil)
	ct.govc.EXPECT().CreateLibrary(ct.ctx, ct.datastore, ct.templateLibrary).Return(nil)
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateInLibrary).Return(false, nil)
	ct.govc.EXPECT().ImportTemplate(ct.ctx, ct.templateLibrary, ct.ovaURL, ct.templateName).Return(nil)
	ct.govc.EXPECT().DeployTemplateFromLibrary(
		ct.ctx, ct.templateDir, ct.templateName, ct.templateLibrary, ct.resourcePool,
	).Return(nil)

	// expects for tagging
	ct.govc.EXPECT().ListCategories(ct.ctx).Return(nil, ct.dummyError)

	ct.assertErrorFromCreateIfMissing()
}

func TestFactoryCreateIfMissingSuccessLibraryDoesNotExist(t *testing.T) {
	ct := newCreateTest(t)
	ct.govc.EXPECT().SearchTemplate(ct.ctx, ct.datacenter, ct.machineConfig).Return("", nil) // template not present
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateLibrary).Return(false, nil)
	ct.govc.EXPECT().CreateLibrary(ct.ctx, ct.datastore, ct.templateLibrary).Return(nil)
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateInLibrary).Return(false, nil)
	ct.govc.EXPECT().ImportTemplate(ct.ctx, ct.templateLibrary, ct.ovaURL, ct.templateName).Return(nil)
	ct.govc.EXPECT().DeployTemplateFromLibrary(
		ct.ctx, ct.templateDir, ct.templateName, ct.templateLibrary, ct.resourcePool,
	).Return(nil)

	// expects for tagging
	ct.govc.EXPECT().ListCategories(ct.ctx).Return(nil, nil)
	ct.govc.EXPECT().ListTags(ct.ctx).Return(nil, nil)

	ct.assertSuccessFromCreateIfMissing()
}

func TestFactoryCreateIfMissingSuccessLibraryExists(t *testing.T) {
	ct := newCreateTest(t)
	ct.govc.EXPECT().SearchTemplate(ct.ctx, ct.datacenter, ct.machineConfig).Return("", nil) // template not present
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateLibrary).Return(true, nil)
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateInLibrary).Return(false, nil)
	ct.govc.EXPECT().ImportTemplate(ct.ctx, ct.templateLibrary, ct.ovaURL, ct.templateName).Return(nil)
	ct.govc.EXPECT().DeployTemplateFromLibrary(
		ct.ctx, ct.templateDir, ct.templateName, ct.templateLibrary, ct.resourcePool,
	).Return(nil)

	// expects for tagging
	ct.govc.EXPECT().ListCategories(ct.ctx).Return(nil, nil)
	ct.govc.EXPECT().ListTags(ct.ctx).Return(nil, nil)

	ct.assertSuccessFromCreateIfMissing()
}

func TestFactoryCreateIfMissingSuccessTemplateInLibrarytExists(t *testing.T) {
	ct := newCreateTest(t)
	ct.govc.EXPECT().SearchTemplate(ct.ctx, ct.datacenter, ct.machineConfig).Return("", nil) // template not present
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateLibrary).Return(true, nil)
	ct.govc.EXPECT().LibraryElementExists(ct.ctx, ct.templateInLibrary).Return(true, nil)
	ct.govc.EXPECT().DeployTemplateFromLibrary(
		ct.ctx, ct.templateDir, ct.templateName, ct.templateLibrary, ct.resourcePool,
	).Return(nil)

	// expects for tagging
	ct.govc.EXPECT().ListCategories(ct.ctx).Return(nil, nil)
	ct.govc.EXPECT().ListTags(ct.ctx).Return(nil, nil)

	ct.assertSuccessFromCreateIfMissing()
}

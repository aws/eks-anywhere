package tags_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/internal/tags"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/internal/tags/mocks"
)

type test struct {
	t          *testing.T
	govc       *mocks.MockGovcClient
	factory    *tags.Factory
	ctx        context.Context
	dummyError error
}

type tagTest struct {
	*test
	templatePath   string
	tagsByCategory map[string][]string
}

func newTest(t *testing.T) *test {
	ctrl := gomock.NewController(t)
	test := &test{
		t:          t,
		govc:       mocks.NewMockGovcClient(ctrl),
		ctx:        context.Background(),
		dummyError: errors.New("error from govc"),
	}
	f := tags.NewFactory(test.govc)
	test.factory = f
	return test
}

func newTagTest(t *testing.T) *tagTest {
	test := newTest(t)
	return &tagTest{
		test:         test,
		templatePath: "/SDDC-Datacenter/vm/Templates/ubuntu-v1.19.8-eks-d-1-19-4-eks-a-0.0.1.build.38-amd64",
		tagsByCategory: map[string][]string{
			"kubernetesChannel": {"kubernetesChannel:1.19"},
			"eksd":              {"eksd:1.19", "eksd:1.19.4"},
		},
	}
}

func (tt *tagTest) tagTemplate() error {
	return tt.factory.TagTemplate(tt.ctx, tt.templatePath, tt.tagsByCategory)
}

func (tt *tagTest) assertErrorFromTagTemplate() {
	if err := tt.tagTemplate(); err == nil {
		tt.t.Fatal("factory.TagTemplate() err = nil, want err not nil")
	}
}

func (tt *tagTest) assertSuccessFromTagTemplate() {
	if err := tt.tagTemplate(); err != nil {
		tt.t.Fatalf("factory.TagTemplate() err = %v, want err = nil", err)
	}
}

func TestFactoryTagTemplateErrorListCategories(t *testing.T) {
	tt := newTagTest(t)
	tt.govc.EXPECT().ListCategories(tt.ctx).Return(nil, tt.dummyError)

	tt.assertErrorFromTagTemplate()
}

func TestFactoryTagTemplateErrorListTags(t *testing.T) {
	tt := newTagTest(t)
	tt.govc.EXPECT().ListCategories(tt.ctx).Return(nil, nil)
	tt.govc.EXPECT().ListTags(tt.ctx).Return(nil, tt.dummyError)

	tt.assertErrorFromTagTemplate()
}

func TestFactoryTagTemplateErrorCreateCategoryForVM(t *testing.T) {
	tt := newTagTest(t)
	tt.govc.EXPECT().ListCategories(tt.ctx).Return(nil, nil)
	tt.govc.EXPECT().ListTags(tt.ctx).Return(nil, nil)
	tt.govc.EXPECT().CreateCategoryForVM(tt.ctx, gomock.Any()).Return(tt.dummyError)

	tt.assertErrorFromTagTemplate()
}

func TestFactoryTagTemplateErrorCreateTag(t *testing.T) {
	tt := newTagTest(t)
	tt.govc.EXPECT().ListCategories(tt.ctx).Return(nil, nil)
	tt.govc.EXPECT().ListTags(tt.ctx).Return(nil, nil)
	tt.govc.EXPECT().CreateCategoryForVM(tt.ctx, gomock.Any()).Return(nil)
	tt.govc.EXPECT().CreateTag(tt.ctx, gomock.Any(), gomock.Any()).Return(tt.dummyError)

	tt.assertErrorFromTagTemplate()
}

func TestFactoryTagTemplateErrorAddTag(t *testing.T) {
	tt := newTagTest(t)
	tt.govc.EXPECT().ListCategories(tt.ctx).Return(nil, nil)
	tt.govc.EXPECT().ListTags(tt.ctx).Return(nil, nil)
	tt.govc.EXPECT().CreateCategoryForVM(tt.ctx, gomock.Any()).Return(nil)
	tt.govc.EXPECT().CreateTag(tt.ctx, gomock.Any(), gomock.Any()).Return(nil)
	tt.govc.EXPECT().AddTag(tt.ctx, tt.templatePath, gomock.Any()).Return(tt.dummyError)

	tt.assertErrorFromTagTemplate()
}

func TestFactoryTagTemplateSuccess(t *testing.T) {
	tt := newTagTest(t)
	tt.govc.EXPECT().ListCategories(tt.ctx).Return([]string{"kubernetesChannel"}, nil)
	tags := []executables.Tag{
		{
			Name:       "eksd:1.19",
			Id:         "urn:vmomi:InventoryServiceTag:5555:GLOBAL",
			CategoryId: "eksd",
		},
	}
	tt.govc.EXPECT().ListTags(tt.ctx).Return(tags, nil)
	tt.govc.EXPECT().CreateTag(tt.ctx, "kubernetesChannel:1.19", "kubernetesChannel").Return(nil)
	tt.govc.EXPECT().AddTag(tt.ctx, tt.templatePath, "kubernetesChannel:1.19").Return(nil)

	tt.govc.EXPECT().CreateCategoryForVM(tt.ctx, "eksd").Return(nil)
	tt.govc.EXPECT().AddTag(tt.ctx, tt.templatePath, "eksd:1.19").Return(nil)
	tt.govc.EXPECT().CreateTag(tt.ctx, "eksd:1.19.4", "eksd").Return(nil)
	tt.govc.EXPECT().AddTag(tt.ctx, tt.templatePath, "eksd:1.19.4").Return(nil)

	tt.assertSuccessFromTagTemplate()
}

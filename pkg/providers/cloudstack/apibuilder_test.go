package cloudstack_test

import (
	"fmt"
	"testing"

	"github.com/aws/eks-anywhere/pkg/constants"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta2"
)

type apiBuilderTest struct {
	*WithT
	machineConfig *v1alpha1.CloudStackMachineConfig
}

func newApiBuilderTest(t *testing.T) apiBuilderTest {
	return apiBuilderTest{
		WithT:         NewWithT(t),
		machineConfig: givenMachineConfig(),
	}
}

const (
	testMountPath       = "testMountPath"
	testDevice          = "testDevice"
	testFilesystem      = "testFilesystem"
	testLabel           = "testLabel"
	testDiskSize        = 5
	computeOfferingId   = "computeOfferingId"
	computeOfferingName = "computeOfferingName"
	diskOfferingId      = "diskOfferingId"
	diskOfferingName    = "diskOfferingName"
	templateId          = "templateId"
	templateName        = "templateName"
	proAffinity         = "pro"
)

var (
	affinityGroupIds = []string{"ag1", "ag2"}
	testSymLinks     = map[string]string{
		"sym":  "link",
		"sym2": "link2",
	}
)
var (
	testSymLinksString = "sym:link,sym2:link2"
	testDetails        = map[string]string{
		"user": "details",
	}
)

func givenMachineConfig() *v1alpha1.CloudStackMachineConfig {
	return &v1alpha1.CloudStackMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind: "CloudStackMachineConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cp",
			Namespace: "test-namespace",
		},
		Spec: v1alpha1.CloudStackMachineConfigSpec{
			ComputeOffering: v1alpha1.CloudStackResourceIdentifier{
				Id:   computeOfferingId,
				Name: computeOfferingName,
			},
			DiskOffering: &v1alpha1.CloudStackResourceDiskOffering{
				CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
					Name: diskOfferingName,
					Id:   diskOfferingId,
				},
				CustomSize: testDiskSize,
				MountPath:  testMountPath,
				Device:     testDevice,
				Filesystem: testFilesystem,
				Label:      testLabel,
			},
			Template: v1alpha1.CloudStackResourceIdentifier{
				Id:   templateId,
				Name: templateName,
			},
			Symlinks:          testSymLinks,
			UserCustomDetails: testDetails,
			AffinityGroupIds:  affinityGroupIds,
			Affinity:          proAffinity,
		},
	}
}

func fullCloudStackMachineTemplate() *cloudstackv1.CloudStackMachineTemplate {
	return &cloudstackv1.CloudStackMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			Kind:       "CloudStackMachineTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cloudstack-test-md-0-1",
			Namespace: "eksa-system",
			Annotations: map[string]string{
				fmt.Sprintf("mountpath.diskoffering.%s", constants.CloudstackAnnotationSuffix):  testMountPath,
				fmt.Sprintf("device.diskoffering.%s", constants.CloudstackAnnotationSuffix):     testDevice,
				fmt.Sprintf("filesystem.diskoffering.%s", constants.CloudstackAnnotationSuffix): testFilesystem,
				fmt.Sprintf("label.diskoffering.%s", constants.CloudstackAnnotationSuffix):      testLabel,
				fmt.Sprintf("symlinks.%s", constants.CloudstackAnnotationSuffix):                testSymLinksString,
			},
		},
		Spec: cloudstackv1.CloudStackMachineTemplateSpec{
			Spec: cloudstackv1.CloudStackMachineTemplateResource{
				Spec: cloudstackv1.CloudStackMachineSpec{
					Details: testDetails,
					Offering: cloudstackv1.CloudStackResourceIdentifier{
						ID:   computeOfferingId,
						Name: computeOfferingName,
					},
					Template: cloudstackv1.CloudStackResourceIdentifier{
						ID:   templateId,
						Name: templateName,
					},
					AffinityGroupIDs: affinityGroupIds,
					Affinity:         proAffinity,
					DiskOffering: cloudstackv1.CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: cloudstackv1.CloudStackResourceIdentifier{
							ID:   diskOfferingId,
							Name: diskOfferingName,
						},
						CustomSize: testDiskSize,
						MountPath:  testMountPath,
						Device:     testDevice,
						Filesystem: testFilesystem,
						Label:      testLabel,
					},
				},
			},
		},
	}
}

func TestFullCloudStackMachineTemplate(t *testing.T) {
	tt := newApiBuilderTest(t)
	got := cloudstack.CloudStackMachineTemplate("cloudstack-test-control-plane-1", tt.machineConfig)
	want := fullCloudStackMachineTemplate()
	tt.Expect(got.Spec.Spec.Spec).To(Equal(want.Spec.Spec.Spec))
	tt.Expect(got.Annotations).To(Equal(want.Annotations))
}

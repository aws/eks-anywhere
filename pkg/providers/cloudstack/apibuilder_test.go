package cloudstack_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
)

type apiBuilderTest struct {
	*WithT
	machineConfig *v1alpha1.CloudStackMachineConfigSpec
}

func newAPIBuilderTest(t *testing.T) apiBuilderTest {
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
	computeOfferingID   = "computeOfferingID"
	computeOfferingName = "computeOfferingName"
	diskOfferingID      = "diskOfferingID"
	diskOfferingName    = "diskOfferingName"
	templateID          = "templateID"
	templateName        = "templateName"
	proAffinity         = "pro"
)

var (
	affinityGroupIds = []string{"ag1", "ag2"}
	testSymLinks     = map[string]string{
		"sym": "link",
	}
	testSymLinksString = "sym:link"
	testDetails        = map[string]string{
		"user": "details",
	}
)

func givenMachineConfig() *v1alpha1.CloudStackMachineConfigSpec {
	return &v1alpha1.CloudStackMachineConfigSpec{
		ComputeOffering: v1alpha1.CloudStackResourceIdentifier{
			Id:   computeOfferingID,
			Name: computeOfferingName,
		},
		DiskOffering: &v1alpha1.CloudStackResourceDiskOffering{
			CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
				Name: diskOfferingName,
				Id:   diskOfferingID,
			},
			CustomSize: testDiskSize,
			MountPath:  testMountPath,
			Device:     testDevice,
			Filesystem: testFilesystem,
			Label:      testLabel,
		},
		Template: v1alpha1.CloudStackResourceIdentifier{
			Id:   templateID,
			Name: templateName,
		},
		Symlinks:          testSymLinks,
		UserCustomDetails: testDetails,
		AffinityGroupIds:  affinityGroupIds,
		Affinity:          proAffinity,
	}
}

func fullCloudStackMachineTemplate() *cloudstackv1.CloudStackMachineTemplate {
	return &cloudstackv1.CloudStackMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: cloudstackv1.GroupVersion.String(),
			Kind:       cloudstack.CloudStackMachineTemplateKind,
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
			Template: cloudstackv1.CloudStackMachineTemplateResource{
				Spec: cloudstackv1.CloudStackMachineSpec{
					Details: testDetails,
					Offering: cloudstackv1.CloudStackResourceIdentifier{
						ID:   computeOfferingID,
						Name: computeOfferingName,
					},
					Template: cloudstackv1.CloudStackResourceIdentifier{
						ID:   templateID,
						Name: templateName,
					},
					AffinityGroupIDs: affinityGroupIds,
					Affinity:         proAffinity,
					DiskOffering: cloudstackv1.CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: cloudstackv1.CloudStackResourceIdentifier{
							ID:   diskOfferingID,
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
	tt := newAPIBuilderTest(t)
	got := cloudstack.MachineTemplate("cloudstack-test-control-plane-1", tt.machineConfig)
	want := fullCloudStackMachineTemplate()
	tt.Expect(got.Spec.Template.Spec).To(Equal(want.Spec.Template.Spec))
	tt.Expect(got.Annotations).To(Equal(want.Annotations))
}

func TestBasicCloudStackMachineDeployment(t *testing.T) {
	tt := newAPIBuilderTest(t)
	count := 1
	workerNodeGroupConfig := v1alpha1.WorkerNodeGroupConfiguration{
		Name:  "test-worker-node-group",
		Count: &count,
	}
	kubeadmConfigTemplates := map[string]*bootstrapv1.KubeadmConfigTemplate{
		workerNodeGroupConfig.Name: {
			TypeMeta: metav1.TypeMeta{
				Kind:       "KubeadmConfigTemplate",
				APIVersion: bootstrapv1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "kubeadmConfigTemplate",
			},
		},
	}
	fullMatchineTemplate := fullCloudStackMachineTemplate()
	matchineTemplates := map[string]*cloudstackv1.CloudStackMachineTemplate{
		workerNodeGroupConfig.Name: fullMatchineTemplate,
	}
	spec := &cluster.Spec{
		VersionsBundles: map[v1alpha1.KubernetesVersion]*cluster.VersionsBundle{
			v1alpha1.Kube119: {
				KubeDistro: &cluster.KubeDistro{
					Kubernetes: cluster.VersionedRepository{
						Tag: "eksd-tag",
					},
				},
			},
		},
		Config: &cluster.Config{
			Cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{
						workerNodeGroupConfig,
					},
					KubernetesVersion: v1alpha1.Kube119,
				},
			},
		},
	}
	got := cloudstack.MachineDeployments(spec, kubeadmConfigTemplates, matchineTemplates)
	tt.Expect(len(got)).To(Equal(*workerNodeGroupConfig.Count))
	tt.Expect(int(*got[workerNodeGroupConfig.Name].Spec.Replicas)).To(Equal(*workerNodeGroupConfig.Count))
	tt.Expect(got[workerNodeGroupConfig.Name].Spec.Template.Spec.InfrastructureRef.Name).To(Equal(fullMatchineTemplate.Name))
	tt.Expect(got[workerNodeGroupConfig.Name].Spec.Template.Spec.InfrastructureRef.Kind).To(Equal(cloudstack.CloudStackMachineTemplateKind))
	tt.Expect(got[workerNodeGroupConfig.Name].Spec.Template.Spec.InfrastructureRef.APIVersion).To(Equal(cloudstackv1.GroupVersion.String()))
}

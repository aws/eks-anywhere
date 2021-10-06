package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestValidateUpdateAWSIamConfigAWSRegion(t *testing.T) {
	aiOld := awsIamConfig()
	aiOld.Spec.AWSRegion = "oldRegion"
	aiNew := aiOld.DeepCopy()

	aiNew.Spec.AWSRegion = "newRegion"
	g := NewWithT(t)
	g.Expect(aiNew.ValidateUpdate(&aiOld)).NotTo(Succeed())
}

func TestValidateUpdateAWSIamConfigClusterID(t *testing.T) {
	aiOld := awsIamConfig()
	aiOld.Spec.ClusterID = "oldClusterID"
	aiNew := aiOld.DeepCopy()

	aiNew.Spec.ClusterID = "newClusterID"
	g := NewWithT(t)
	g.Expect(aiNew.ValidateUpdate(&aiOld)).NotTo(Succeed())
}

func TestValidateUpdateAWSIamConfigPartition(t *testing.T) {
	aiOld := awsIamConfig()
	aiOld.Spec.Partition = "oldPartition"
	aiNew := aiOld.DeepCopy()

	aiNew.Spec.Partition = "newPartition"
	g := NewWithT(t)
	g.Expect(aiNew.ValidateUpdate(&aiOld)).NotTo(Succeed())
}

func TestValidateUpdateAWSIamConfigBackendMode(t *testing.T) {
	aiOld := awsIamConfig()
	aiOld.Spec.BackendMode = []string{"mode1", "mode2"}
	aiNew := aiOld.DeepCopy()

	aiNew.Spec.BackendMode = []string{"mode1"}
	g := NewWithT(t)
	g.Expect(aiNew.ValidateUpdate(&aiOld)).NotTo(Succeed())
}

func awsIamConfig() v1alpha1.AWSIamConfig {
	return v1alpha1.AWSIamConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec:       v1alpha1.AWSIamConfigSpec{},
		Status:     v1alpha1.AWSIamConfigStatus{},
	}
}

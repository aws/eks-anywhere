package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestValidateUpdateAWSIamConfigFail(t *testing.T) {
	aiOld := awsIamConfig()
	aiOld.Spec.BackendMode = []string{"mode1", "mode2"}
	aiNew := aiOld.DeepCopy()

	aiNew.Spec.BackendMode = []string{"mode1"}
	g := NewWithT(t)
	g.Expect(aiNew.ValidateUpdate(&aiOld)).NotTo(Succeed())
}

func TestValidateUpdateAWSIamConfigSuccess(t *testing.T) {
	aiOld := awsIamConfig()
	aiOld.Spec.MapRoles = []v1alpha1.MapRoles{}
	aiNew := aiOld.DeepCopy()

	aiNew.Spec.MapRoles = []v1alpha1.MapRoles{
		{
			RoleARN:  "test-role-arn",
			Username: "test-user",
			Groups:   []string{"group1", "group2"},
		},
	}
	g := NewWithT(t)
	g.Expect(aiNew.ValidateUpdate(&aiOld)).To(Succeed())
}

func awsIamConfig() v1alpha1.AWSIamConfig {
	return v1alpha1.AWSIamConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec:       v1alpha1.AWSIamConfigSpec{},
		Status:     v1alpha1.AWSIamConfigStatus{},
	}
}

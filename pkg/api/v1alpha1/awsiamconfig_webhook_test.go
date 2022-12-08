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
	g.Expect(aiNew.ValidateUpdate(&aiOld)).To(MatchError(ContainSubstring("config is immutable")))
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

func TestValidateCreateAWSIamConfigSuccess(t *testing.T) {
	aiNew := awsIamConfig()

	g := NewWithT(t)
	g.Expect(aiNew.ValidateCreate()).To(Succeed())
}

func TestValidateCreateAWSIamConfigFail(t *testing.T) {
	aiNew := awsIamConfig()
	aiNew.Spec.AWSRegion = ""

	g := NewWithT(t)
	g.Expect(aiNew.ValidateCreate()).To(MatchError(ContainSubstring("AWSRegion is a required field")))
}

func TestValidateUpdateAWSIamConfigFailCausedByMutableFieldChange(t *testing.T) {
	aiOld := awsIamConfig()
	aiOld.Spec.MapRoles = []v1alpha1.MapRoles{}
	aiNew := aiOld.DeepCopy()

	aiNew.Spec.MapRoles = []v1alpha1.MapRoles{
		{
			RoleARN:  "test-role-arn",
			Username: "",
			Groups:   []string{"group1", "group2"},
		},
	}
	g := NewWithT(t)
	g.Expect(aiNew.ValidateUpdate(&aiOld)).To(MatchError(ContainSubstring("MapRoles Username is required")))
}

func TestAWSIamConfigSetDefaults(t *testing.T) {
	g := NewWithT(t)

	sOld := awsIamConfig()
	sOld.Default()

	g.Expect(sOld.Spec.Partition).To(Equal(v1alpha1.DefaultAWSIamConfigPartition))
}

func awsIamConfig() v1alpha1.AWSIamConfig {
	return v1alpha1.AWSIamConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec: v1alpha1.AWSIamConfigSpec{
			AWSRegion:   "us-east-1",
			BackendMode: []string{"mode1"},
		},
		Status: v1alpha1.AWSIamConfigStatus{},
	}
}

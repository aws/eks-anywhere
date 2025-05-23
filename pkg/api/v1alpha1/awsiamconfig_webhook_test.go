package v1alpha1_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestValidateUpdateAWSIamConfigFail(t *testing.T) {
	ctx := context.Background()
	aiOld := awsIamConfig()
	aiOld.Spec.BackendMode = []string{"mode1", "mode2"}
	aiNew := aiOld.DeepCopy()

	aiNew.Spec.BackendMode = []string{"mode1"}
	g := NewWithT(t)
	g.Expect(aiNew.ValidateUpdate(ctx, aiNew, &aiOld)).Error().To(MatchError(ContainSubstring("config is immutable")))
}

func TestValidateUpdateAWSIamConfigSuccess(t *testing.T) {
	ctx := context.Background()
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
	g.Expect(aiNew.ValidateUpdate(ctx, aiNew, &aiOld)).Error().To(Succeed())
}

func TestValidateCreateAWSIamConfigSuccess(t *testing.T) {
	ctx := context.Background()
	aiNew := awsIamConfig()

	g := NewWithT(t)
	g.Expect(aiNew.ValidateCreate(ctx, &aiNew)).Error().To(Succeed())
}

func TestValidateCreateAWSIamConfigFail(t *testing.T) {
	ctx := context.Background()
	aiNew := awsIamConfig()
	aiNew.Spec.AWSRegion = ""

	g := NewWithT(t)
	g.Expect(aiNew.ValidateCreate(ctx, &aiNew)).Error().To(MatchError(ContainSubstring("AWSRegion is a required field")))
}

func TestValidateUpdateAWSIamConfigFailCausedByMutableFieldChange(t *testing.T) {
	ctx := context.Background()
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
	g.Expect(aiNew.ValidateUpdate(ctx, aiNew, &aiOld)).Error().To(MatchError(ContainSubstring("MapRoles Username is required")))
}

func TestAWSIamConfigSetDefaults(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	sOld := awsIamConfig()
	err := sOld.Default(ctx, &sOld)
	g.Expect(err).To(BeNil())

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

func TestAWSIamConfigDefaultCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomDefaulter
	config := &v1alpha1.AWSIamConfig{}

	// Call Default with the wrong type
	err := config.Default(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected an AWSIamConfig"))
}

func TestAWSIamConfigValidateCreateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.AWSIamConfig{}

	// Call ValidateCreate with the wrong type
	warnings, err := config.ValidateCreate(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected an AWSIamConfig"))
}

func TestAWSIamConfigValidateUpdateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.AWSIamConfig{}

	// Call ValidateUpdate with the wrong type
	warnings, err := config.ValidateUpdate(context.TODO(), wrongType, &v1alpha1.AWSIamConfig{})

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected an AWSIamConfig"))
}

func TestAWSIamConfigValidateDeleteCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.AWSIamConfig{}

	// Call ValidateDelete with the wrong type
	warnings, err := config.ValidateDelete(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected an AWSIamConfig"))
}

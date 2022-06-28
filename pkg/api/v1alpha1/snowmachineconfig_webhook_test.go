package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestSnowMachineConfigSetDefaults(t *testing.T) {
	g := NewWithT(t)

	sOld := snowMachineConfig()
	sOld.Default()

	g.Expect(sOld.Spec.InstanceType).To(Equal(v1alpha1.DefaultSnowInstanceType))
	g.Expect(sOld.Spec.PhysicalNetworkConnector).To(Equal(v1alpha1.DefaultSnowPhysicalNetworkConnectorType))
}

func snowMachineConfig() v1alpha1.SnowMachineConfig {
	return v1alpha1.SnowMachineConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 2)},
		Spec:       v1alpha1.SnowMachineConfigSpec{},
		Status:     v1alpha1.SnowMachineConfigStatus{},
	}
}

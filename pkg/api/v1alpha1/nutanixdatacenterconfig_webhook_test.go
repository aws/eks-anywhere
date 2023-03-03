package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNutanixDatacenterConfigWebhookSetsDefaults(t *testing.T) {
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	dcConf.SetDefaults()
	g.Expect(dcConf.ValidateCreate()).To(Succeed())
}

func nutanixDatacenterConfig() *NutanixDatacenterConfig {
	return &NutanixDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nutanix-datacenter-config",
		},
		Spec: NutanixDatacenterConfigSpec{
			Endpoint: "prism.nutanix.com",
			Port:     9440,
		},
	}
}

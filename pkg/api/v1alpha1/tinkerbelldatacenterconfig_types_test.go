package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestTinkerbellDatacenterConfigValidateFail(t *testing.T) {
	tests := []struct {
		name    string
		tinkDC  *v1alpha1.TinkerbellDatacenterConfig
		wantErr string
	}{
		{
			name: "Empty Tink IP",
			tinkDC: newTinkerbellDatacenterConfig(func(dc *v1alpha1.TinkerbellDatacenterConfig) {
				dc.Spec.TinkerbellIP = ""
			}),
			wantErr: "missing spec.tinkerbellIP field",
		},
		{
			name: "Invalid Tink IP",
			tinkDC: newTinkerbellDatacenterConfig(func(dc *v1alpha1.TinkerbellDatacenterConfig) {
				dc.Spec.TinkerbellIP = "10"
			}),
			wantErr: "invalid tinkerbell ip: ",
		},
		{
			name: "Invalid OS Image URL",
			tinkDC: newTinkerbellDatacenterConfig(func(dc *v1alpha1.TinkerbellDatacenterConfig) {
				dc.Spec.OSImageURL = "test"
			}),
			wantErr: "parsing osImageOverride: parse \"test\": invalid URI for request",
		},
		{
			name: "Invalid hook Image URL",
			tinkDC: newTinkerbellDatacenterConfig(func(dc *v1alpha1.TinkerbellDatacenterConfig) {
				dc.Spec.HookImagesURLPath = "test"
			}),
			wantErr: "parsing hookOverride: parse \"test\": invalid URI for request",
		},
		{
			name: "invalid object data",
			tinkDC: newTinkerbellDatacenterConfig(func(dc *v1alpha1.TinkerbellDatacenterConfig) {
				dc.ObjectMeta.Name = ""
			}),
			wantErr: "TinkerbellDatacenterConfig: missing name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.tinkDC.Validate()).To(MatchError(ContainSubstring(tt.wantErr)))
		})
	}
}

func TestTinkerbellDatacenterConfigValidateSuccess(t *testing.T) {
	tinkDC := createTinkerbellDatacenterConfig()

	g := NewWithT(t)
	g.Expect(tinkDC.Validate()).To(Succeed())
}

func newTinkerbellDatacenterConfig(opts ...func(*v1alpha1.TinkerbellDatacenterConfig)) *v1alpha1.TinkerbellDatacenterConfig {
	c := createTinkerbellDatacenterConfig()
	for _, o := range opts {
		o(c)
	}

	return c
}

type tinkerbellDatacenterOpt func(dc *v1alpha1.TinkerbellDatacenterConfig)

func createTinkerbellDatacenterConfig(opts ...tinkerbellDatacenterOpt) *v1alpha1.TinkerbellDatacenterConfig {
	dc := &v1alpha1.TinkerbellDatacenterConfig{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: make(map[string]string, 1),
			Name:        "tinkerbelldatacenterconfig",
		},
		Spec: v1alpha1.TinkerbellDatacenterConfigSpec{
			TinkerbellIP: "1.1.1.1",
		},
		Status: v1alpha1.TinkerbellDatacenterConfigStatus{},
	}

	for _, opt := range opts {
		opt(dc)
	}

	return dc
}

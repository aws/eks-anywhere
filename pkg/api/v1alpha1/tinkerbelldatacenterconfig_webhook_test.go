package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

func TestTinkerbellDatacenterValidateCreate(t *testing.T) {
	dataCenterConfig := tinkerbellDatacenterConfig()

	g := NewWithT(t)
	g.Expect(dataCenterConfig.ValidateCreate()).To(Succeed())
}

func TestTinkerbellDatacenterValidateCreateFail(t *testing.T) {
	dataCenterConfig := tinkerbellDatacenterConfig()
	dataCenterConfig.Spec.TinkerbellIP = ""

	g := NewWithT(t)
	g.Expect(dataCenterConfig.ValidateCreate()).NotTo(Succeed())
}

func TestTinkerbellDatacenterValidateUpdateSucceed(t *testing.T) {
	tOld := tinkerbellDatacenterConfig()
	tOld.Spec.TinkerbellIP = "1.1.1.1"
	tNew := tOld.DeepCopy()

	tNew.Spec.TinkerbellIP = "1.1.1.1"
	g := NewWithT(t)
	g.Expect(tNew.ValidateUpdate(&tOld)).To(Succeed())
}

func TestTinkerbellDatacenterValidateUpdateSucceedOSImageURL(t *testing.T) {
	tOld := tinkerbellDatacenterConfig()
	tNew := tOld.DeepCopy()

	tNew.Spec.OSImageURL = "https://os-image-url"
	g := NewWithT(t)
	g.Expect(tNew.ValidateUpdate(&tOld)).To(Succeed())
}

func TestTinkerbellDatacenterValidateUpdateFailBadReq(t *testing.T) {
	cOld := &v1alpha1.Cluster{}
	c := &v1alpha1.TinkerbellDatacenterConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("expected a TinkerbellDatacenterConfig but got a *v1alpha1.Cluster")))
}

func TestTinkerbellDatacenterValidateUpdateImmutable(t *testing.T) {
	tests := []struct {
		name    string
		old     v1alpha1.TinkerbellDatacenterConfig
		new     v1alpha1.TinkerbellDatacenterConfig
		wantErr string
	}{
		{
			name: "updated tinkerbellIP",
			old: tinkerbellDatacenterConfig(func(d *v1alpha1.TinkerbellDatacenterConfig) {
				d.Spec.TinkerbellIP = "1.1.1.1"
			}),
			new: tinkerbellDatacenterConfig(func(d *v1alpha1.TinkerbellDatacenterConfig) {
				d.Spec.TinkerbellIP = "1.1.1.2"
			}),
			wantErr: "spec.tinkerbellIP: Forbidden: field is immutable",
		},
		{
			name: "updated hookImagesURLPath",
			old: tinkerbellDatacenterConfig(func(d *v1alpha1.TinkerbellDatacenterConfig) {
				d.Spec.HookImagesURLPath = "https://oldpath"
			}),
			new: tinkerbellDatacenterConfig(func(d *v1alpha1.TinkerbellDatacenterConfig) {
				d.Spec.HookImagesURLPath = "https://newpath"
			}),
			wantErr: "spec.hookImagesURLPath: Forbidden: field is immutable",
		},
		{
			name: "updated hookImagesURLPath, managed by cli",
			old: tinkerbellDatacenterConfig(func(d *v1alpha1.TinkerbellDatacenterConfig) {
				d.Spec.HookImagesURLPath = "https://oldpath"
			}),
			new: tinkerbellDatacenterConfig(func(d *v1alpha1.TinkerbellDatacenterConfig) {
				d.Spec.HookImagesURLPath = "https://newpath"
				clientutil.AddAnnotation(d, v1alpha1.ManagedByCLIAnnotation, "true")
			}),
			wantErr: "",
		},
		{
			name: "updated skipLoadBalancerDeployment",
			old: tinkerbellDatacenterConfig(func(d *v1alpha1.TinkerbellDatacenterConfig) {
				d.Spec.SkipLoadBalancerDeployment = false
			}),
			new: tinkerbellDatacenterConfig(func(d *v1alpha1.TinkerbellDatacenterConfig) {
				d.Spec.SkipLoadBalancerDeployment = true
			}),
			wantErr: "spec.skipLoadBalancerDeployment: Forbidden: field is immutable",
		},
		{
			name: "updated skipLoadBalancerDeployment, managed by cli",
			old: tinkerbellDatacenterConfig(func(d *v1alpha1.TinkerbellDatacenterConfig) {
				d.Spec.SkipLoadBalancerDeployment = false
			}),
			new: tinkerbellDatacenterConfig(func(d *v1alpha1.TinkerbellDatacenterConfig) {
				d.Spec.SkipLoadBalancerDeployment = true
				clientutil.AddAnnotation(d, v1alpha1.ManagedByCLIAnnotation, "true")
			}),
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			err := tt.new.ValidateUpdate(&tt.old)
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}

}

func TestTinkerbellDatacenterValidateDelete(t *testing.T) {
	tOld := tinkerbellDatacenterConfig()

	g := NewWithT(t)
	g.Expect(tOld.ValidateDelete()).To(Succeed())
}

type tinkerbellDatacenterConfigOpt func(*v1alpha1.TinkerbellDatacenterConfig)

func tinkerbellDatacenterConfig(opts ...tinkerbellDatacenterConfigOpt) v1alpha1.TinkerbellDatacenterConfig {
	d := v1alpha1.TinkerbellDatacenterConfig{
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
		opt(&d)
	}

	return d
}

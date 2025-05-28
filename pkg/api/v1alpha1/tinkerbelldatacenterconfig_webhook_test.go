package v1alpha1_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

func TestTinkerbellDatacenterValidateCreate(t *testing.T) {
	ctx := context.Background()
	dataCenterConfig := tinkerbellDatacenterConfig()

	g := NewWithT(t)
	g.Expect(dataCenterConfig.ValidateCreate(ctx, &dataCenterConfig)).Error().To(Succeed())
}

func TestTinkerbellDatacenterValidateCreateFail(t *testing.T) {
	ctx := context.Background()
	dataCenterConfig := tinkerbellDatacenterConfig()
	dataCenterConfig.Spec.TinkerbellIP = ""

	g := NewWithT(t)
	g.Expect(dataCenterConfig.ValidateCreate(ctx, &dataCenterConfig)).Error().To(HaveOccurred())
}

func TestTinkerbellDatacenterValidateUpdateSucceed(t *testing.T) {
	ctx := context.Background()
	tOld := tinkerbellDatacenterConfig()
	tOld.Spec.TinkerbellIP = "1.1.1.1"
	tNew := tOld.DeepCopy()

	tNew.Spec.TinkerbellIP = "1.1.1.1"
	g := NewWithT(t)
	g.Expect(tNew.ValidateUpdate(ctx, tNew, &tOld)).Error().To(Succeed())
}

func TestTinkerbellDatacenterValidateUpdateSucceedOSImageURL(t *testing.T) {
	ctx := context.Background()
	tOld := tinkerbellDatacenterConfig()
	tNew := tOld.DeepCopy()

	tNew.Spec.OSImageURL = "https://os-image-url"
	g := NewWithT(t)
	g.Expect(tNew.ValidateUpdate(ctx, tNew, &tOld)).Error().To(Succeed())
}

func TestTinkerbellDatacenterValidateUpdateFailBadReq(t *testing.T) {
	ctx := context.Background()
	cOld := &v1alpha1.Cluster{}
	c := &v1alpha1.TinkerbellDatacenterConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(ctx, cOld, c)).Error().To(MatchError(ContainSubstring("expected a TinkerbellDatacenterConfig but got a *v1alpha1.Cluster")))
}

func TestTinkerbellDatacenterValidateUpdateImmutable(t *testing.T) {
	ctx := context.Background()
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

			_, err := tt.new.ValidateUpdate(ctx, &tt.old, &tt.new)
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestTinkerbellDatacenterValidateDelete(t *testing.T) {
	ctx := context.Background()
	tOld := tinkerbellDatacenterConfig()

	g := NewWithT(t)
	g.Expect(tOld.ValidateDelete(ctx, &tOld)).Error().To(Succeed())
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

func TestTinkerbellDatacenterConfigValidateCreateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.TinkerbellDatacenterConfig{}

	// Call ValidateCreate with the wrong type
	warnings, err := config.ValidateCreate(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a TinkerbellDatacenterConfig"))
}

func TestTinkerbellDatacenterConfigValidateUpdateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.TinkerbellDatacenterConfig{}

	// Call ValidateUpdate with the wrong type
	warnings, err := config.ValidateUpdate(context.TODO(), wrongType, &v1alpha1.TinkerbellDatacenterConfig{})

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a TinkerbellDatacenterConfig"))
}

func TestTinkerbellDatacenterConfigValidateDeleteCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.TinkerbellDatacenterConfig{}

	// Call ValidateDelete with the wrong type
	warnings, err := config.ValidateDelete(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a TinkerbellDatacenterConfig"))
}

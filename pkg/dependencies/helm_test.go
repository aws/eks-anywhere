package dependencies_test

import (
	"context"
	"testing"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	. "github.com/onsi/gomega"
)

func TestHelmFactory_GetClient(t *testing.T) {
	tests := map[string]struct {
		wantErr error
	}{
		"Success": {
			wantErr: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			g := NewWithT(t)
			deps, err := dependencies.NewFactory().
				WithLocalExecutables().
				WithHelmFactory().
				Build(context.Background())

			g.Expect(err).To(BeNil())

			helm, err := deps.HelmFactory.GetClientForCluster(ctx, "")

			if tt.wantErr != nil {
				g.Expect(err).To(Equal(tt.wantErr))
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(deps.HelmFactory).NotTo(BeNil())
				g.Expect(helm).NotTo(BeNil())
			}
		})
	}
}

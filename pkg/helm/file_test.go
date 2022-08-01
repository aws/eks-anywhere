package helm_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/helm"
)

func TestChartFileName(t *testing.T) {
	tests := []struct {
		name  string
		chart string
		want  string
	}{
		{
			name:  "full path",
			chart: "ecr.com/folder/folder2/chart-name:1.0.0",
			want:  "chart-name-1.0.0.tgz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(helm.ChartFileName(tt.chart)).To(Equal(tt.want))
		})
	}
}

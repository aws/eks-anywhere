package oci_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/utils/oci"
)

func TestURL(t *testing.T) {
	tests := []struct {
		name         string
		artifactPath string
		want         string
	}{
		{
			name:         "normal artifact",
			artifactPath: "public.ecr.aws/folder/folder2/name",
			want:         "oci://public.ecr.aws/folder/folder2/name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(oci.URL(tt.artifactPath)).To(Equal(tt.want))
		})
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name              string
		artifact          string
		wantPath, wantTag string
	}{
		{
			name:     "normal chart",
			artifact: "public.ecr.aws/folder/folder2/name:v1.0.0",
			wantPath: "public.ecr.aws/folder/folder2/name",
			wantTag:  "v1.0.0",
		},
		{
			name:     "no version",
			artifact: "public.ecr.aws/folder/folder2/name",
			wantPath: "public.ecr.aws/folder/folder2/name",
			wantTag:  "",
		},
		{
			name:     "no version with colon",
			artifact: "public.ecr.aws/folder/folder2/name:",
			wantPath: "public.ecr.aws/folder/folder2/name",
			wantTag:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			gotPath, gotTag := oci.Split(tt.artifact)
			g.Expect(gotPath).To(Equal(tt.wantPath))
			g.Expect(gotTag).To(Equal(tt.wantTag))
		})
	}
}

func TestChartURLAndVersion(t *testing.T) {
	tests := []struct {
		name                 string
		chart                string
		wantURL, wantVersion string
	}{
		{
			name:        "normal chart",
			chart:       "public.ecr.aws/folder/folder2/name:v1.0.0",
			wantURL:     "oci://public.ecr.aws/folder/folder2/name",
			wantVersion: "v1.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			gotURL, gotVersion := oci.ChartURLAndVersion(tt.chart)
			g.Expect(gotURL).To(Equal(tt.wantURL))
			g.Expect(gotVersion).To(Equal(tt.wantVersion))
		})
	}
}

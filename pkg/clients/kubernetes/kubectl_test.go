package kubernetes_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestKubectlGetOptionsApplyToGet(t *testing.T) {
	tests := []struct {
		name             string
		option, in, want *kubernetes.KubectlGetOptions
	}{
		{
			name:   "empty",
			option: &kubernetes.KubectlGetOptions{},
			in: &kubernetes.KubectlGetOptions{
				Name:      "my-name",
				Namespace: "ns",
			},
			want: &kubernetes.KubectlGetOptions{
				Name:      "my-name",
				Namespace: "ns",
			},
		},
		{
			name: "only Namespace",
			option: &kubernetes.KubectlGetOptions{
				Namespace: "other-ns",
			},
			in: &kubernetes.KubectlGetOptions{
				Name:      "my-name",
				Namespace: "ns",
			},
			want: &kubernetes.KubectlGetOptions{
				Name:      "my-name",
				Namespace: "other-ns",
			},
		},
		{
			name: "Namespace and Name",
			option: &kubernetes.KubectlGetOptions{
				Name:      "my-other-name",
				Namespace: "other-ns",
			},
			in: &kubernetes.KubectlGetOptions{
				Name:      "my-name",
				Namespace: "ns",
			},
			want: &kubernetes.KubectlGetOptions{
				Name:      "my-other-name",
				Namespace: "other-ns",
			},
		},
		{
			name: "Name and ClusterScope",
			option: &kubernetes.KubectlGetOptions{
				Name:          "my-other-name",
				ClusterScoped: ptr.Bool(true),
			},
			in: &kubernetes.KubectlGetOptions{
				Name: "my-name",
			},
			want: &kubernetes.KubectlGetOptions{
				Name:          "my-other-name",
				ClusterScoped: ptr.Bool(true),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			tt.option.ApplyToGet(tt.in)
			g.Expect(tt.in).To(BeComparableTo(tt.want))
		})
	}
}

func TestKubectlDeleteOptionsApplyToDelete(t *testing.T) {
	tests := []struct {
		name             string
		option, in, want *kubernetes.KubectlDeleteOptions
	}{
		{
			name:   "empty",
			option: &kubernetes.KubectlDeleteOptions{},
			in: &kubernetes.KubectlDeleteOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "ns",
			},
			want: &kubernetes.KubectlDeleteOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "ns",
			},
		},
		{
			name: "only Namespace",
			option: &kubernetes.KubectlDeleteOptions{
				Namespace: "other-ns",
			},
			in: &kubernetes.KubectlDeleteOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "ns",
			},
			want: &kubernetes.KubectlDeleteOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "other-ns",
			},
		},
		{
			name: "Namespace and labels",
			option: &kubernetes.KubectlDeleteOptions{
				Namespace: "other-ns",
				HasLabels: map[string]string{
					"label2": "value2",
				},
			},
			in: &kubernetes.KubectlDeleteOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "ns",
			},
			want: &kubernetes.KubectlDeleteOptions{
				HasLabels: map[string]string{
					"label2": "value2",
				},
				Namespace: "other-ns",
			},
		},
		{
			name: "empty not nil labels",
			option: &kubernetes.KubectlDeleteOptions{
				HasLabels: map[string]string{},
			},
			in: &kubernetes.KubectlDeleteOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "ns",
			},
			want: &kubernetes.KubectlDeleteOptions{
				HasLabels: map[string]string{},
				Namespace: "ns",
			},
		},
		{
			name: "Namespace and Name",
			option: &kubernetes.KubectlDeleteOptions{
				Name:      "my-other-name",
				Namespace: "other-ns",
			},
			in: &kubernetes.KubectlDeleteOptions{
				Name:      "my-name",
				Namespace: "ns",
			},
			want: &kubernetes.KubectlDeleteOptions{
				Name:      "my-other-name",
				Namespace: "other-ns",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			tt.option.ApplyToDelete(tt.in)
			g.Expect(tt.in).To(BeComparableTo(tt.want))
		})
	}
}

func TestKubectlApplyOptionsApplyToApply(t *testing.T) {
	tests := []struct {
		name             string
		option, in, want *kubernetes.KubectlApplyOptions
	}{
		{
			name:   "empty",
			option: &kubernetes.KubectlApplyOptions{},
			in:     &kubernetes.KubectlApplyOptions{},
			want:   &kubernetes.KubectlApplyOptions{},
		},
		{
			name: "serverside",
			option: &kubernetes.KubectlApplyOptions{
				ServerSide: true,
			},
			in: &kubernetes.KubectlApplyOptions{
				FieldManager: "a",
			},
			want: &kubernetes.KubectlApplyOptions{
				ServerSide:   true,
				FieldManager: "a",
			},
		},
		{
			name: "force ownership",
			option: &kubernetes.KubectlApplyOptions{
				ForceOwnership: true,
			},
			in: &kubernetes.KubectlApplyOptions{
				FieldManager: "a",
			},
			want: &kubernetes.KubectlApplyOptions{
				ForceOwnership: true,
				FieldManager:   "a",
			},
		},
		{
			name: "field manager",
			option: &kubernetes.KubectlApplyOptions{
				FieldManager: "a",
			},
			in: &kubernetes.KubectlApplyOptions{
				FieldManager:   "b",
				ServerSide:     true,
				ForceOwnership: true,
			},
			want: &kubernetes.KubectlApplyOptions{
				FieldManager:   "a",
				ServerSide:     true,
				ForceOwnership: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			tt.option.ApplyToApply(tt.in)
			g.Expect(tt.in).To(BeComparableTo(tt.want))
		})
	}
}

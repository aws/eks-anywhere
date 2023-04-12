package kubernetes_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

func TestDeleteAllOfOptionsApplyToDeleteAllOf(t *testing.T) {
	tests := []struct {
		name             string
		option, in, want *kubernetes.DeleteAllOfOptions
	}{
		{
			name:   "empty",
			option: &kubernetes.DeleteAllOfOptions{},
			in: &kubernetes.DeleteAllOfOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "ns",
			},
			want: &kubernetes.DeleteAllOfOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "ns",
			},
		},
		{
			name: "only Namespace",
			option: &kubernetes.DeleteAllOfOptions{
				Namespace: "other-ns",
			},
			in: &kubernetes.DeleteAllOfOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "ns",
			},
			want: &kubernetes.DeleteAllOfOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "other-ns",
			},
		},
		{
			name: "Namespace and labels",
			option: &kubernetes.DeleteAllOfOptions{
				Namespace: "other-ns",
				HasLabels: map[string]string{
					"label2": "value2",
				},
			},
			in: &kubernetes.DeleteAllOfOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "ns",
			},
			want: &kubernetes.DeleteAllOfOptions{
				HasLabels: map[string]string{
					"label2": "value2",
				},
				Namespace: "other-ns",
			},
		},
		{
			name: "empty not nil labels",
			option: &kubernetes.DeleteAllOfOptions{
				HasLabels: map[string]string{},
			},
			in: &kubernetes.DeleteAllOfOptions{
				HasLabels: map[string]string{
					"label": "value",
				},
				Namespace: "ns",
			},
			want: &kubernetes.DeleteAllOfOptions{
				HasLabels: map[string]string{},
				Namespace: "ns",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			tt.option.ApplyToDeleteAllOf(tt.in)
			g.Expect(tt.in).To(BeComparableTo(tt.want))
		})
	}
}

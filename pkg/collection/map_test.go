package collection_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/collection"
)

func TestMapDiff(t *testing.T) {
	tests := []struct {
		name string
		map1 map[string]string
		map2 map[string]string
		want map[string]string
	}{
		{
			name: "Both maps empty",
			map1: map[string]string{},
			map2: map[string]string{},
			want: map[string]string{},
		},
		{
			name: "No diff in maps",
			map1: map[string]string{
				"key": "val",
			},
			map2: map[string]string{
				"key": "val",
			},
			want: map[string]string{},
		},
		{
			name: "item in map1 not in map2 (by key)",
			map1: map[string]string{
				"key": "val",
			},
			map2: map[string]string{},
			want: map[string]string{
				"key": "val",
			},
		},
		{
			name: "item in map1 not in map2 (by value)",
			map1: map[string]string{
				"key": "val1",
			},
			map2: map[string]string{
				"key": "val",
			},
			want: map[string]string{
				"key": "val1",
			},
		},
		{
			name: "item in map2 not in map1",
			map1: map[string]string{
				"key": "val",
			},
			map2: map[string]string{
				"key":  "val",
				"key2": "val",
			},
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			diff := collection.MapDiff(tt.map1, tt.map2)
			g.Expect(diff).To(Equal(tt.want))
		})
	}
}

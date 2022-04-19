package types_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/types"
)

func TestLookupIsPresent(t *testing.T) {
	tests := []struct {
		testName    string
		value       string
		slice       []string
		wantPresent bool
	}{
		{
			testName:    "empty slice",
			slice:       []string{},
			value:       "v",
			wantPresent: false,
		},
		{
			testName:    "nil slice",
			slice:       nil,
			value:       "v",
			wantPresent: false,
		},
		{
			testName:    "value present",
			slice:       []string{"v2", "v1"},
			value:       "v",
			wantPresent: false,
		},
		{
			testName:    "value present",
			slice:       []string{"v2", "v"},
			value:       "v",
			wantPresent: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			l := types.SliceToLookup(tt.slice)
			if got := l.IsPresent(tt.value); got != tt.wantPresent {
				t.Errorf("Lookup.IsPresent() = %v, want %v", got, tt.wantPresent)
			}
		})
	}
}

func TestLookupToSlice(t *testing.T) {
	tests := []struct {
		name string
		l    types.Lookup
		want []string
	}{
		{
			name: "empty",
			l:    types.Lookup{},
			want: []string{},
		},
		{
			name: "not empty",
			l:    types.SliceToLookup([]string{"a", "a", "a", "a"}),
			want: []string{"a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.l.ToSlice()).To(Equal(tt.want))
		})
	}
}

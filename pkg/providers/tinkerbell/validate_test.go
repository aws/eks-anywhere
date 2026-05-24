package tinkerbell

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPermuatations(t *testing.T) {
	tests := []struct {
		initial    string
		separators []string
		want       string
	}{
		{"125", []string{"-", "_", ""}, "125"},
		{"1.26", []string{"-", "_", ""}, "1.26, 1-26, 1_26 or 126"},
		{"1.27", []string{"."}, "1.27"},
		{"1.28", []string{".", "."}, "1.28"},
		{"1.29", []string{"-", "-", ""}, "1.29, 1-29 or 129"},
		{"1.29.1", []string{"-", "_", "_", ""}, "1.29.1, 1-29-1, 1_29_1 or 1291"},
	}

	for _, tt := range tests {
		t.Run(tt.initial, func(t *testing.T) {
			got := permutations(tt.initial, tt.separators)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("permutations() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

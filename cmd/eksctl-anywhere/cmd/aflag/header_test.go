package aflag

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHeaderM(t *testing.T) {
	tests := map[string]struct {
		shouldErr bool
		want      http.Header
		input     string
	}{
		"success":            {want: map[string][]string{"A": {"1", "2", "3"}, "B": {"2"}, "C": {"4"}}, input: `A=1;2;3,B=2,C=4`},
		"bad input format 1": {shouldErr: true, input: `abc`, want: http.Header{}},
		"bad input format 2": {shouldErr: true, input: `abc=123;abd=345,`, want: http.Header{}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			h := http.Header{}
			got := NewHeader(&h)
			if err := got.Set(tc.input); err != nil {
				if !tc.shouldErr {
					t.Fatal(err)
				}
			}
			if diff := cmp.Diff(tc.want, h); diff != "" {
				t.Fatalf("diff: %s", diff)
			}
		})
	}
}

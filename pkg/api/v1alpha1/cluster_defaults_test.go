package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestSetClusterDefaults(t *testing.T) {
	tests := []struct {
		name            string
		in, wantCluster *Cluster
		wantErr         string
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			gotErr := setClusterDefaults(tt.in)
			if tt.wantErr == "" {
				g.Expect(gotErr).To(BeNil())
			} else {
				g.Expect(gotErr).To(MatchError(ContainSubstring(tt.wantErr)))
			}

			g.Expect(tt.in).To(Equal(tt.wantCluster))
		})
	}
}

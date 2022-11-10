package ec2

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestMakeChunks(t *testing.T) {
	tests := []struct {
		name      string
		elements  []*int
		chunkSize int
		want      [][]*int
	}{
		{
			name:      "empty elements",
			elements:  []*int{},
			chunkSize: 2,
			want:      [][]*int{},
		},
		{
			name:      "nil elements",
			elements:  nil,
			chunkSize: 2,
			want:      [][]*int{},
		},
		{
			name: "even length, exact division",
			elements: []*int{
				ptr.Int(1), ptr.Int(2),
				ptr.Int(3), ptr.Int(4),
				ptr.Int(5), ptr.Int(6),
				ptr.Int(7), ptr.Int(8),
			},
			chunkSize: 2,
			want: [][]*int{
				{ptr.Int(1), ptr.Int(2)},
				{ptr.Int(3), ptr.Int(4)},
				{ptr.Int(5), ptr.Int(6)},
				{ptr.Int(7), ptr.Int(8)},
			},
		},
		{
			name: "odd length, exact division",
			elements: []*int{
				ptr.Int(1), ptr.Int(2), ptr.Int(3),
				ptr.Int(4), ptr.Int(5), ptr.Int(6),
				ptr.Int(7), ptr.Int(8), ptr.Int(9),
			},
			chunkSize: 3,
			want: [][]*int{
				{ptr.Int(1), ptr.Int(2), ptr.Int(3)},
				{ptr.Int(4), ptr.Int(5), ptr.Int(6)},
				{ptr.Int(7), ptr.Int(8), ptr.Int(9)},
			},
		},
		{
			name: "even length, non exact division",
			elements: []*int{
				ptr.Int(1), ptr.Int(2),
				ptr.Int(3), ptr.Int(4),
				ptr.Int(5), ptr.Int(6),
				ptr.Int(7),
			},
			chunkSize: 2,
			want: [][]*int{
				{ptr.Int(1), ptr.Int(2)},
				{ptr.Int(3), ptr.Int(4)},
				{ptr.Int(5), ptr.Int(6)},
				{ptr.Int(7)},
			},
		},
		{
			name: "odd length, non exact division",
			elements: []*int{
				ptr.Int(1), ptr.Int(2), ptr.Int(3),
				ptr.Int(4), ptr.Int(5), ptr.Int(6),
				ptr.Int(7),
			},
			chunkSize: 3,
			want: [][]*int{
				{ptr.Int(1), ptr.Int(2), ptr.Int(3)},
				{ptr.Int(4), ptr.Int(5), ptr.Int(6)},
				{ptr.Int(7)},
			},
		},
		{
			name: "length same as chunk size",
			elements: []*int{
				ptr.Int(1), ptr.Int(2),
				ptr.Int(3), ptr.Int(4),
				ptr.Int(5), ptr.Int(6),
				ptr.Int(7),
			},
			chunkSize: 7,
			want: [][]*int{
				{
					ptr.Int(1), ptr.Int(2),
					ptr.Int(3), ptr.Int(4),
					ptr.Int(5), ptr.Int(6),
					ptr.Int(7),
				},
			},
		},
		{
			name: "length same smaller than chunk size",
			elements: []*int{
				ptr.Int(1), ptr.Int(2),
				ptr.Int(3), ptr.Int(4),
				ptr.Int(5), ptr.Int(6),
				ptr.Int(7),
			},
			chunkSize: 17,
			want: [][]*int{
				{
					ptr.Int(1), ptr.Int(2),
					ptr.Int(3), ptr.Int(4),
					ptr.Int(5), ptr.Int(6),
					ptr.Int(7),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(makeChunks(tt.elements, tt.chunkSize)).To(Equal(tt.want))
		})
	}
}

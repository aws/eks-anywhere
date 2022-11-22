package collection_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/collection"
)

func TestSetContains(t *testing.T) {
	testCases := []struct {
		testName string
		set      collection.Set[string]
		element  string
		want     bool
	}{
		{
			testName: "empty set",
			set:      collection.NewSet[string](),
			element:  "a",
			want:     false,
		},
		{
			testName: "contained in non empty set",
			set:      collection.NewSetFrom("b", "a", "c"),
			element:  "a",
			want:     true,
		},
		{
			testName: "not contained in non empty set",
			set:      collection.NewSetFrom("b", "a", "c"),
			element:  "d",
			want:     false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.set.Contains(tt.element)).To(Equal(tt.want))
		})
	}
}

func TestSetDelete(t *testing.T) {
	g := NewWithT(t)
	s := collection.NewSetFrom("b", "a", "c")
	g.Expect(s.Contains("c")).To(BeTrue())
	s.Delete("c")
	g.Expect(s.Contains("c")).To(BeFalse())

	g.Expect(s.ToSlice()).To(ConsistOf("a", "b"))
}

func TestSetToSlice(t *testing.T) {
	testCases := []struct {
		testName string
		set      collection.Set[string]
		want     []string
	}{
		{
			testName: "empty set",
			set:      collection.NewSet[string](),
			want:     []string{},
		},
		{
			testName: "non empty set",
			set:      collection.NewSetFrom("b", "a", "c", "d", "a", "b"),
			want: []string{
				"a", "b", "c", "d",
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.set.ToSlice()).To(ConsistOf(tt.want))
		})
	}
}

func TestMapSet(t *testing.T) {
	g := NewWithT(t)
	elements := []myStruct{
		{
			name: "a",
		},
		{
			name: "b",
		},
		{
			name: "b",
		},
	}

	s := collection.MapSet(elements, func(e myStruct) string {
		return e.name
	})

	g.Expect(s.ToSlice()).To(ConsistOf("a", "b"))
}

type myStruct struct {
	name string
}

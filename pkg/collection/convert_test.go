package collection_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/collection"
)

func TestToMap(t *testing.T) {
	type Elem struct {
		Name  string
		Value int
	}
	elems := []Elem{
		{Name: "4", Value: 4},
		{Name: "3", Value: 3},
		{Name: "1", Value: 1},
	}

	m := collection.ToMap(elems, func(e Elem) string { return e.Name })

	for _, e := range elems {
		if v, present := m[e.Name]; !present || v.Value != e.Value {
			t.Fatalf("Missing elements: %#v", e)
		}
	}
}

func TestToSlice(t *testing.T) {
	type Elem struct {
		Name  string
		Value int
	}
	elems := map[string]Elem{
		"4": {Name: "4", Value: 4},
		"5": {Name: "5", Value: 5},
		"2": {Name: "2", Value: 2},
	}

	s := collection.ToSlice(elems)

	contains := func(s []Elem, e Elem) bool {
		for _, v := range s {
			if v == e {
				return true
			}
		}
		return false
	}

	for _, e := range elems {
		if !contains(s, e) {
			t.Fatalf("Missing elements: %#v", e)
		}
	}
}

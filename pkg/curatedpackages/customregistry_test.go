package curatedpackages

import (
	"context"
	"fmt"
)

func ExampleCustomRegistry_GetRegistryBaseRef() {
	r := NewCustomRegistry("foo")
	ref, _ := r.GetRegistryBaseRef(context.Background())
	fmt.Println(ref)
	// Output: foo/eks-anywhere-packages-bundles
}

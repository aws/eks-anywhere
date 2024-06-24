package e2e

import (
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

var kubeVersions = []v1alpha1.KubernetesVersion{v1alpha1.Kube126, v1alpha1.Kube127, v1alpha1.Kube128, v1alpha1.Kube129, v1alpha1.Kube130}

// Subtest is an interface to represent a test case.
type Subtest interface {
	// Return the name of the subtest case.
	GetName() string
	// Return the name suffix of the parent test, which includes this subtest. It should be "*Suite".
	GetSuiteSuffix() string
}

// Suites contain all test suites. The key is the suite suffix.
var Suites = map[string][]Subtest{
	simpleFlowSubtest.GetSuiteSuffix(): {},
}

// SimpleFlowSubtest is a struct to represent a simple flow test.
type SimpleFlowSubtest struct {
	KubeVersion v1alpha1.KubernetesVersion
}

// GetName returns the subtest name.
func (st *SimpleFlowSubtest) GetName() string {
	return strings.ReplaceAll(string(st.KubeVersion), ".", "")
}

// GetSuiteSuffix returns the Suite suffix.
func (st *SimpleFlowSubtest) GetSuiteSuffix() string {
	return "SimpleFlowSuite"
}

// Make sure the SimpleFlowSubtest implements the Subtest interface.
var simpleFlowSubtest Subtest = (*SimpleFlowSubtest)(nil)

// Init SimpleFlowSuite.
func init() {
	s := simpleFlowSubtest.GetSuiteSuffix()
	for _, k := range kubeVersions {
		Suites[s] = append(Suites[s], &SimpleFlowSubtest{KubeVersion: k})
	}
}

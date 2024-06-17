//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

func init() {
	// This is necessary because the test framework builds controller-runtime
	// k8s client, and the library requires SetLogger to be called before
	// it's used. Otherwise it prints a confusing warning and it hides any
	// other client log.
	// There might a better way for this, but this will do for now.
	ctrl.SetLogger(klog.Background())
}

var kubeVersions = []v1alpha1.KubernetesVersion{v1alpha1.Kube128, v1alpha1.Kube129, v1alpha1.Kube130}

// Subtest is an interface to represent a test case.
type Subtest interface {
	// Return the name of the subtest case.
	GetName() string
	// Return the name suffix of the test case, which includes this subtest.
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

// GetName returns the subtest cname.
func (st *SimpleFlowSubtest) GetName() string {
	return string(st.KubeVersion)
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

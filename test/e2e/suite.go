//go:build e2e
// +build e2e

package e2e

import (
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

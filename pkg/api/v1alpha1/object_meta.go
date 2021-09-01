package v1alpha1

// Need to override metav1.ObjectMeta as a hack due to the following issue of creationTimestamp being printed:
// https://github.com/kubernetes/kubernetes/issues/86811
// Add more fields based on https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/types.go#L114-L288
// and https://github.com/kubernetes-sigs/cluster-api/blob/bf790fc2a53614ff5d3405c83c0de0dd3303bb1f/api/v1alpha2/common_types.go#L67-L128
// as needed.
type ObjectMeta struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

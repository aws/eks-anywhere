package clusterapi

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	addonsalpha3 "sigs.k8s.io/cluster-api/exp/addons/api/v1alpha3"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/templater"
)

type ClusterResourceSet struct {
	resources   map[string][]byte
	clusterName string
	namespace   string
}

func NewClusterResourceSet(clusterName string) *ClusterResourceSet {
	return &ClusterResourceSet{
		clusterName: clusterName,
		namespace:   "default",
		resources:   make(map[string][]byte),
	}
}

func (c ClusterResourceSet) AddResource(name string, content []byte) {
	c.resources[name] = content
}

func (c ClusterResourceSet) ToYaml() ([]byte, error) {
	if len(c.resources) == 0 {
		return nil, nil
	}

	return marshall(append(c.buildResourceConfigMaps(), c.buildSet())...)
}

func (c ClusterResourceSet) buildSet() *addonsalpha3.ClusterResourceSet {
	return &addonsalpha3.ClusterResourceSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: addonsalpha3.GroupVersion.Identifier(),
			Kind:       "ClusterResourceSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-crs", c.clusterName),
			Labels: map[string]string{
				v1alpha3.ClusterLabelName: c.clusterName,
			},
			Namespace: c.namespace,
		},
		Spec: addonsalpha3.ClusterResourceSetSpec{
			ClusterSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					v1alpha3.ClusterLabelName: c.clusterName,
				},
			},
			Resources: c.resourceRefs(),
		},
	}
}

func (c ClusterResourceSet) resourceRefs() []addonsalpha3.ResourceRef {
	refs := make([]addonsalpha3.ResourceRef, 0, len(c.resources))

	for name := range c.resources {
		refs = append(refs, addonsalpha3.ResourceRef{Name: name, Kind: string(addonsalpha3.ConfigMapClusterResourceSetResourceKind)})
	}

	return refs
}

func (c ClusterResourceSet) buildResourceConfigMaps() []interface{} {
	cms := make([]interface{}, 0, len(c.resources))

	for name, content := range c.resources {
		cm := corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: c.namespace,
			},
			Data: map[string]string{
				"data": string(content),
			},
		}
		cms = append(cms, cm)
	}

	return cms
}

func marshall(objects ...interface{}) ([]byte, error) {
	bytes := make([][]byte, 0, len(objects))
	for _, o := range objects {
		b, err := yaml.Marshal(o)
		if err != nil {
			return nil, fmt.Errorf("error marshalling object for cluster resource set: %v", err)
		}

		bytes = append(bytes, b)
	}

	return templater.AppendYamlResources(bytes...), nil
}

package kindnetd

import (
	"errors"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/templater"
)

func generateManifest(reader manifests.FileReader, clusterSpec *cluster.Spec) ([]byte, error) {
	versionsBundle := clusterSpec.RootVersionsBundle()
	kindnetdManifest, err := bundles.ReadManifest(reader, versionsBundle.Kindnetd.Manifest)
	if err != nil {
		return nil, fmt.Errorf("can't load kindnetd manifest: %v", err)
	}
	templates := strings.Split(string(kindnetdManifest.Content), "---")
	finalTemplates := make([][]byte, 0, len(templates))
	for _, template := range templates {
		u := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(template), u); err != nil {
			return nil, fmt.Errorf("unmarshaling kindnetd type [%s]: %v", template, err)
		}
		if u.GetKind() == "DaemonSet" {
			updated, err := updatePodSubnet(clusterSpec, u)
			if err != nil {
				return nil, fmt.Errorf("updating kindnetd pod subnet [%s]: %v", template, err)
			}
			finalTemplates = append(finalTemplates, updated)
		} else {
			finalTemplates = append(finalTemplates, []byte(template))
		}
	}
	return templater.AppendYamlResources(finalTemplates...), nil
}

func updatePodSubnet(clusterSpec *cluster.Spec, unstructured *unstructured.Unstructured) ([]byte, error) {
	var daemonSet appsv1.DaemonSet
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.UnstructuredContent(), &daemonSet); err != nil {
		return nil, fmt.Errorf("unmarshaling kindnetd daemonset: %v", err)
	}
	if len(daemonSet.Spec.Template.Spec.Containers) == 0 {
		return nil, errors.New("missing container in kindnetd daemonset")
	}
	for idx, env := range daemonSet.Spec.Template.Spec.Containers[0].Env {
		if env.Name == "POD_SUBNET" {
			daemonSet.Spec.Template.Spec.Containers[0].Env[idx].Value = clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks[0]
		}
	}
	return yaml.Marshal(daemonSet)
}

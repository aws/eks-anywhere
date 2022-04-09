package stack

import (
	_ "embed"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/templater"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

//go:embed manifests/pbnj.yaml
var pbnjManifest string

//go:embed manifests/tink.yaml
var tinkManifest string

const defaultEksaNamespace = "eksa-system"

func GeneratePbnjManifest(image string) ([]byte, error) {
	pbnjDeployment := appsv1.Deployment{}

	if err := yaml.UnmarshalStrict([]byte(pbnjManifest), &pbnjDeployment); err != nil {
		return nil, fmt.Errorf("unmarshalling PBNJ deployment: %v", err)
	}

	pbnjDeployment.Spec.Template.Spec.Containers[0].Image = image

	return yaml.Marshal(pbnjDeployment)
}

func GenerateTinkManifest(image, tinkerbellIp string) ([]byte, error) {
	values := map[string]string{
		"tinkServerImage":  image,
		"tinkerbellHostIp": tinkerbellIp,
		"namespace":        defaultEksaNamespace,
		"grpcPort":         "42113",
		"certPort":         "42114",
	}

	return templater.Execute(tinkManifest, values)
}

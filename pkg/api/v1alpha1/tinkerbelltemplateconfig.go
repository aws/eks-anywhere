package v1alpha1

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const TinkerbellTemplateConfigKind = "TinkerbellTemplateConfig"

// +kubebuilder:object:generate=false
type ActionOpt func(action *[]tinkerbell.Action)

// NewDefaultTinkerbellTemplateConfigCreate returns a default TinkerbellTemplateConfig with the
// required Tasks and Actions.
func NewDefaultTinkerbellTemplateConfigCreate(clusterSpec *Cluster, versionBundle v1alpha1.VersionsBundle, osImageOverride, tinkerbellLocalIP, tinkerbellLBIP string, osFamily OSFamily) *TinkerbellTemplateConfig {
	config := &TinkerbellTemplateConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       TinkerbellTemplateConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterSpec.Name,
		},
		Spec: TinkerbellTemplateConfigSpec{
			Template: tinkerbell.Workflow{
				Version:       "0.1",
				Name:          clusterSpec.Name,
				GlobalTimeout: 6000,
				Tasks: []tinkerbell.Task{{
					Name:       clusterSpec.Name,
					WorkerAddr: "{{.device_1}}",
					Volumes: []string{
						"/dev:/dev",
						"/dev/console:/dev/console",
						"/lib/firmware:/lib/firmware:ro",
					},
				}},
			},
		},
	}

	defaultActions := GetDefaultActionsFromBundle(clusterSpec, versionBundle, osImageOverride, tinkerbellLocalIP, tinkerbellLBIP, osFamily)
	for _, action := range defaultActions {
		action(&config.Spec.Template.Tasks[0].Actions)
	}

	return config
}

func (c *TinkerbellTemplateConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *TinkerbellTemplateConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *TinkerbellTemplateConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetTinkerbellTemplateConfig(fileName string) (map[string]*TinkerbellTemplateConfig, error) {
	templates := make(map[string]*TinkerbellTemplateConfig)
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}

	r := yamlutil.NewYAMLReader(bufio.NewReader(bytes.NewReader(content)))
	for {
		d, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		var template TinkerbellTemplateConfig
		if err := yaml.Unmarshal(d, &template); err != nil {
			return nil, fmt.Errorf("unable to unmarshall content from file due to: %v", err)
		}

		if template.Kind() == template.ExpectedKind() {
			if err := yamlutil.UnmarshalStrict(d, &template); err != nil {
				return nil, fmt.Errorf("invalid template config content: %v", err)
			}
			templates[template.Name] = &template
		}

	}
	return templates, nil
}

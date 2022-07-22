package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const TinkerbellTemplateConfigKind = "TinkerbellTemplateConfig"

// +kubebuilder:object:generate=false
type ActionOpt func(action *[]tinkerbell.Action)

// NewDefaultTinkerbellTemplateConfigCreate returns a default TinkerbellTemplateConfig with the required Tasks and Actions
func NewDefaultTinkerbellTemplateConfigCreate(name string, versionBundle v1alpha1.VersionsBundle, disk string, osImageOverride, tinkerbellLocalIp, tinkerbellLBIp string, osFamily OSFamily) *TinkerbellTemplateConfig {
	config := &TinkerbellTemplateConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       TinkerbellTemplateConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: TinkerbellTemplateConfigSpec{
			Template: tinkerbell.Workflow{
				Version:       "0.1",
				Name:          name,
				GlobalTimeout: 6000,
				Tasks: []tinkerbell.Task{{
					Name:       name,
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

	defaultActions := GetDefaultActionsFromBundle(versionBundle, disk, osImageOverride, tinkerbellLocalIp, tinkerbellLBIp, osFamily)
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
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), YamlSeparator) {
		var template TinkerbellTemplateConfig
		if err := yaml.Unmarshal([]byte(c), &template); err != nil {
			return nil, fmt.Errorf("unable to unmarshall content from file due to: %v", err)
		}

		if template.Kind() == template.ExpectedKind() {
			if err = yaml.UnmarshalStrict([]byte(c), &template); err != nil {
				return nil, fmt.Errorf("invalid template config content: %v", err)
			}
			templates[template.Name] = &template
		}
	}
	return templates, nil
}

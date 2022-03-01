package cluster

import (
	"errors"
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type parsed struct {
	objects        objects
	cluster        *anywherev1.Cluster
	datacenter     apiObject
	machineConfigs []runtime.Object
}

type basicAPIObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func (k *basicAPIObject) empty() bool {
	return k.APIVersion == "" && k.Kind == ""
}

type apiObject interface {
	runtime.Object
	GetName() string
}

func keyForObject(o apiObject) string {
	return key(o.GetObjectKind().GroupVersionKind().GroupVersion().String(), o.GetObjectKind().GroupVersionKind().Kind, o.GetName())
}

type objects map[string]apiObject

func (o objects) add(obj apiObject) {
	o[keyForObject(obj)] = obj
}

func (o objects) getFromRef(apiVersion string, ref anywherev1.Ref) apiObject {
	return o[keyForRef(apiVersion, ref)]
}

func key(apiVersion, kind, name string) string {
	// this assumes we don't allow to have objects in multiple namespaces
	return fmt.Sprintf("%s%s%s", apiVersion, kind, name)
}

func keyForRef(apiVersion string, ref anywherev1.Ref) string {
	return key(apiVersion, ref.Kind, ref.Name)
}

func ParseConfigFromFile(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading cluster config file: %v", err)
	}

	return ParseConfig(content)
}

func ParseConfig(yamlManifest []byte) (*Config, error) {
	parsed := &parsed{
		objects: objects{},
	}
	yamlObjs := strings.Split(string(yamlManifest), "---")

	for _, yamlObj := range yamlObjs {
		k := &basicAPIObject{}
		err := yaml.Unmarshal([]byte(yamlObj), k)
		if err != nil {
			return nil, err
		}

		// Ignore empty objects.
		// Empty objects are generated if there are weird things in manifest files like e.g. two --- in a row without a yaml doc in the middle
		if k.empty() {
			continue
		}

		var obj apiObject

		switch k.Kind {
		case anywherev1.ClusterKind:
			if parsed.cluster != nil {
				return nil, errors.New("only one Cluster per yaml manifest is allowed")
			}
			parsed.cluster = &anywherev1.Cluster{}
			obj = parsed.cluster
		case anywherev1.VSphereDatacenterKind:
			obj = &anywherev1.VSphereDatacenterConfig{}
		case anywherev1.VSphereMachineConfigKind:
			obj = &anywherev1.VSphereMachineConfig{}
		case anywherev1.DockerDatacenterKind:
			obj = &anywherev1.DockerDatacenterConfig{}
		case anywherev1.AWSIamConfigKind:
			obj = &anywherev1.AWSIamConfig{}
		case anywherev1.OIDCConfigKind:
			obj = &anywherev1.OIDCConfig{}
		case anywherev1.GitOpsConfigKind:
			obj = &anywherev1.GitOpsConfig{}
		default:
			return nil, fmt.Errorf("invalid object with kind %s found on manifest", k.Kind)
		}

		if err := yaml.Unmarshal([]byte(yamlObj), obj); err != nil {
			return nil, err
		}

		parsed.objects.add(obj)
	}

	return buildConfigFromParsed(parsed)
}

func buildConfigFromParsed(p *parsed) (*Config, error) {
	if p.cluster == nil {
		return nil, errors.New("no Cluster found in manifest")
	}

	c := &Config{
		Cluster:               p.cluster,
		VSphereMachineConfigs: map[string]*anywherev1.VSphereMachineConfig{},
		OIDCConfigs:           map[string]*anywherev1.OIDCConfig{},
		AWSIAMConfigs:         map[string]*anywherev1.AWSIamConfig{},
	}

	// Process datacenter
	p.datacenter = p.objects.getFromRef(p.cluster.APIVersion, p.cluster.Spec.DatacenterRef)
	switch p.cluster.Spec.DatacenterRef.Kind {
	case anywherev1.VSphereDatacenterKind:
		c.VSphereDatacenter = p.datacenter.(*anywherev1.VSphereDatacenterConfig)
	case anywherev1.DockerDatacenterKind:
		c.DockerDatacenter = p.datacenter.(*anywherev1.DockerDatacenterConfig)
	}

	// Process machine configs
	processMachineConfig(p, c, p.cluster.Spec.ControlPlaneConfiguration.MachineGroupRef)
	if p.cluster.Spec.ExternalEtcdConfiguration != nil {
		processMachineConfig(p, c, c.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef)
	}

	for _, w := range p.cluster.Spec.WorkerNodeGroupConfigurations {
		processMachineConfig(p, c, w.MachineGroupRef)
	}

	// Process IDP
	for _, idr := range p.cluster.Spec.IdentityProviderRefs {
		processIdentityProvider(p, c, idr)
	}

	// Process GitOps
	processGitOps(p, c)

	return c, nil
}

func processMachineConfig(p *parsed, c *Config, machineRef *anywherev1.Ref) {
	if machineRef == nil {
		return
	}

	m := p.objects.getFromRef(c.Cluster.APIVersion, *machineRef)
	if m == nil {
		return
	}

	p.machineConfigs = append(p.machineConfigs, m)
	switch machineRef.Kind {
	case anywherev1.VSphereMachineConfigKind:
		c.VSphereMachineConfigs[m.GetName()] = m.(*anywherev1.VSphereMachineConfig)
	}
}

func processIdentityProvider(p *parsed, c *Config, idpRef anywherev1.Ref) {
	idp := p.objects.getFromRef(c.Cluster.APIVersion, idpRef)
	if idp == nil {
		return
	}

	switch idpRef.Kind {
	case anywherev1.OIDCConfigKind:
		c.OIDCConfigs[idp.GetName()] = idp.(*anywherev1.OIDCConfig)
	case anywherev1.AWSIamConfigKind:
		c.AWSIAMConfigs[idp.GetName()] = idp.(*anywherev1.AWSIamConfig)
	}
}

func processGitOps(p *parsed, c *Config) {
	if c.Cluster.Spec.GitOpsRef == nil {
		return
	}

	gitOps := p.objects.getFromRef(p.cluster.APIVersion, *p.cluster.Spec.GitOpsRef)
	if gitOps == nil {
		return
	}

	switch p.cluster.Spec.GitOpsRef.Kind {
	case anywherev1.GitOpsConfigKind:
		c.GitOpsConfig = gitOps.(*anywherev1.GitOpsConfig)
	}
}

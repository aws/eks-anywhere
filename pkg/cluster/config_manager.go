package cluster

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
)

// ConfigManager allows to parse from yaml, set defaults and validate a Cluster struct
// It allows to dynamically register configuration for all those operations
type ConfigManager struct { // TODO: find a better name
	entry *ConfigManagerEntry
}

// NewConfigManager builds a ConfigManager with empty configuration
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		entry: NewConfigManagerEntry(),
	}
}

// Register records the configuration defined in a ConfigManagerEntry into the ConfigManager
// This is equivalent to the individual register methods
func (c *ConfigManager) Register(entries ...*ConfigManagerEntry) error {
	return c.entry.Merge(entries...)
}

// RegisterMapping records the mapping between a kubernetes Kind and an API concrete type
func (c *ConfigManager) RegisterMapping(kind string, generator APIObjectGenerator) error {
	return c.entry.RegisterMapping(kind, generator)
}

// RegisterProcessors records setters to fill the Config struct from the parsed API objects
func (c *ConfigManager) RegisterProcessors(processors ...ParsedProcessor) {
	c.entry.RegisterProcessors(processors...)
}

// RegisterValidations records validations for a Config struct
func (c *ConfigManager) RegisterValidations(validations ...Validation) {
	c.entry.RegisterValidations(validations...)
}

// RegisterDefaulters records defaults for a Config struct
func (c *ConfigManager) RegisterDefaulters(defaulters ...Defaulter) {
	c.entry.RegisterDefaulters(defaulters...)
}

// Parse reads yaml manifest with at least one cluster object and generates the corresponding Config
func (c *ConfigManager) Parse(yamlManifest []byte) (*Config, error) {
	parsed, err := c.unmarshal(yamlManifest)
	if err != nil {
		return nil, err
	}

	return c.buildConfigFromParsed(parsed)
}

// Parse set the registered defaults in a Config struct
func (c *ConfigManager) SetDefaults(config *Config) error {
	var allErrs []error

	for _, d := range c.entry.Defaulters {
		if err := d(config); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if len(allErrs) > 0 {
		aggregate := utilerrors.NewAggregate(allErrs)
		return fmt.Errorf("error setting defaults on cluster config: %v", aggregate)
	}

	return nil
}

// Validate performs the registered validations in a Config struct
func (c *ConfigManager) Validate(config *Config) error {
	var allErrs []error

	for _, v := range c.entry.Validations {
		if err := v(config); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if len(allErrs) > 0 {
		aggregate := utilerrors.NewAggregate(allErrs)
		return fmt.Errorf("invalid cluster config: %v", aggregate)
	}

	return nil
}

type basicAPIObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func (k *basicAPIObject) empty() bool {
	return k.APIVersion == "" && k.Kind == ""
}

type parsed struct {
	objects ObjectLookup
	cluster *anywherev1.Cluster
}

var separatorRegex = regexp.MustCompile(`(?m)^---$`)

func (c *ConfigManager) unmarshal(yamlManifest []byte) (*parsed, error) {
	parsed := &parsed{
		objects: ObjectLookup{},
	}
	yamlObjs := separatorRegex.Split(string(yamlManifest), -1)

	for _, yamlObj := range yamlObjs {
		trimmedYamlObj := strings.TrimSuffix(yamlObj, "\n")
		k := &basicAPIObject{}
		err := yaml.Unmarshal([]byte(trimmedYamlObj), k)
		if err != nil {
			return nil, err
		}

		// Ignore empty objects.
		// Empty objects are generated if there are weird things in manifest files like e.g. two --- in a row without a yaml doc in the middle
		if k.empty() {
			continue
		}

		var obj APIObject

		if k.Kind == anywherev1.ClusterKind {
			if parsed.cluster != nil {
				return nil, errors.New("only one Cluster per yaml manifest is allowed")
			}
			parsed.cluster = &anywherev1.Cluster{}
			obj = parsed.cluster
		} else if generateApiObj, ok := c.entry.APIObjectMapping[k.Kind]; ok {
			obj = generateApiObj()
		} else {
			logger.V(2).Info("Ignoring object in yaml of unknown type when parsing cluster Config", "kind", k.Kind)
			continue
		}

		if err := yaml.Unmarshal([]byte(trimmedYamlObj), obj); err != nil {
			return nil, err
		}
		parsed.objects.add(obj)
	}

	return parsed, nil
}

func (c *ConfigManager) buildConfigFromParsed(p *parsed) (*Config, error) {
	if p.cluster == nil {
		return nil, errors.New("no Cluster found in manifest")
	}

	config := &Config{
		Cluster: p.cluster,
	}

	for _, processor := range c.entry.Processors {
		processor(config, p.objects)
	}

	return config, nil
}

// machineConfigsProcessor is a helper to generate a ParsedProcessor for all machine configs in a Cluster
func machineConfigsProcessor(processMachineRef func(c *Config, o ObjectLookup, machineRef *anywherev1.Ref)) ParsedProcessor {
	return func(c *Config, o ObjectLookup) {
		for _, m := range c.Cluster.MachineConfigRefs() {
			processMachineRef(c, o, &m)
		}
	}
}

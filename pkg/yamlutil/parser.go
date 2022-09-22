package yamlutil

import (
	"bufio"
	"bytes"
	"io"
	"regexp"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

type (
	// APIObjectGenerator returns an implementor of the APIObject interface
	APIObjectGenerator func() APIObject
	// ParsedProcessor fills the Config struct from the parsed API objects in ObjectLookup
	ParsedProcessor[T any] func(*T, ObjectLookup)

	// ConfigManager allows to parse from yaml, set defaults and validate a Cluster struct
	// It allows to dynamically register configuration for all those operations
	Parser[T any] struct {
		apiObjectMapping map[string]APIObjectGenerator
		processors       []ParsedProcessor[T]
		logger           logr.Logger
	}
)

func NewParser[T any](logger logr.Logger) *Parser[T] {
	return &Parser[T]{
		apiObjectMapping: make(map[string]APIObjectGenerator),
		logger:           logger,
	}
}

// RegisterMapping records the mapping between a kubernetes Kind and an API concrete type
func (c *Parser[T]) RegisterMapping(kind string, generator APIObjectGenerator) error {
	if _, ok := c.apiObjectMapping[kind]; ok {
		return errors.Errorf("mapping for api object %s already registered", kind)
	}

	c.apiObjectMapping[kind] = generator
	return nil
}

type Mapping[T APIObject] struct {
	New  func() T
	Kind string
}

func NewMapping[T APIObject](kind string, new func() T) Mapping[T] {
	return Mapping[T]{
		Kind: kind,
		New:  new,
	}
}

func (m Mapping[T]) ToGenericMapping() Mapping[APIObject] {
	return Mapping[APIObject]{
		Kind: m.Kind,
		New: func() APIObject {
			return m.New()
		},
	}
}

func (c *Parser[T]) RegisterMappings(mappings ...Mapping[APIObject]) error {
	for _, m := range mappings {
		if err := c.RegisterMapping(m.Kind, m.New); err != nil {
			return err
		}
	}

	return nil
}

type MapperRegistrar interface {
	RegisterMapping(kind string, generator APIObjectGenerator) error
}

func RegisterMapping[K APIObject](registrar MapperRegistrar, kind string) error {
	return registrar.RegisterMapping(kind, func() APIObject {
		return *new(K)
	})
}

// RegisterProcessors records setters to fill the Config struct from the parsed API objects
func (c *Parser[T]) RegisterProcessors(processors ...ParsedProcessor[T]) {
	c.processors = append(c.processors, processors...)
}

// Parse reads yaml manifest with at least one cluster object and generates the corresponding Config
func (c *Parser[T]) Parse(yamlManifest []byte) (*T, error) {
	return c.Read(bytes.NewReader(yamlManifest))
}

func (c *Parser[T]) Read(reader io.Reader) (*T, error) {
	parsed, err := c.unmarshal(reader)
	if err != nil {
		return nil, err
	}

	return c.buildConfigFromParsed(parsed)
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
}

var separatorRegex = regexp.MustCompile(`(?m)^---$`)

func (c Parser[T]) unmarshal(reader io.Reader) (*parsed, error) {
	parsed := &parsed{
		objects: ObjectLookup{},
	}

	yamlReader := apiyaml.NewYAMLReader(bufio.NewReader(reader))
	for {
		// Read one YAML document at a time, until io.EOF is returned
		b, err := yamlReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrapf(err, "failed to read yaml")
		}
		if len(b) == 0 {
			break
		}

		k := &basicAPIObject{}
		if err = yaml.Unmarshal(b, k); err != nil {
			return nil, err
		}

		// Ignore empty objects.
		// Empty objects are generated if there are weird things in manifest files like e.g. two --- in a row without a yaml doc in the middle
		if k.empty() {
			continue
		}

		var obj APIObject
		if generateApiObj, ok := c.apiObjectMapping[k.Kind]; ok {
			obj = generateApiObj()
		} else {
			c.logger.V(2).Info("Ignoring object in yaml of unknown type during parsing", "kind", k.Kind)
			continue
		}

		if err := yaml.Unmarshal(b, obj); err != nil {
			return nil, err
		}
		parsed.objects.add(obj)
	}

	return parsed, nil
}

func (c Parser[T]) buildConfigFromParsed(p *parsed) (*T, error) {
	t := new(T)
	for _, processor := range c.processors {
		processor(t, p.objects)
	}

	return t, nil
}

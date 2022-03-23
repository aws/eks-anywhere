package cluster

import "fmt"

type (
	// APIObjectGenerator returns an implementor of the APIObject interface
	APIObjectGenerator func() APIObject
	// ParsedProcessor fills the Config struct from the parsed API objects in ObjectLookup
	ParsedProcessor func(*Config, ObjectLookup)
	// Validation performs a validation over the Config object
	Validation func(*Config) error
	// Defaulter sets defaults in a Config object
	Defaulter func(*Config) error
)

// ConfigManagerEntry allows to declare the necessary configuration to parse
// from yaml, set defaults and validate a Cluster struct for one or more types.
// It is semantically equivalent to use the individual register methods and its
// only purpose is convenience.
type ConfigManagerEntry struct {
	APIObjectMapping map[string]APIObjectGenerator
	Processors       []ParsedProcessor
	Validations      []Validation
	Defaulters       []Defaulter
}

// NewConfigManagerEntry builds a ConfigManagerEntry with empty configuration
func NewConfigManagerEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{},
	}
}

// Merge combines the configuration declared in multiple ConfigManagerEntry
func (c *ConfigManagerEntry) Merge(entries ...*ConfigManagerEntry) error {
	for _, config := range entries {
		for k, v := range config.APIObjectMapping {
			if err := c.RegisterMapping(k, v); err != nil {
				return err
			}
		}

		c.RegisterProcessors(config.Processors...)
		c.RegisterDefaulters(config.Defaulters...)
		c.RegisterValidations(config.Validations...)
	}

	return nil
}

// RegisterMapping records the mapping between a kubernetes Kind and an API concrete type
func (c *ConfigManagerEntry) RegisterMapping(kind string, generator APIObjectGenerator) error {
	if _, ok := c.APIObjectMapping[kind]; ok {
		return fmt.Errorf("mapping for api object %s already registered", kind)
	}

	c.APIObjectMapping[kind] = generator
	return nil
}

// RegisterProcessors records setters to fill the Config struct from the parsed API objects
func (c *ConfigManagerEntry) RegisterProcessors(processors ...ParsedProcessor) {
	c.Processors = append(c.Processors, processors...)
}

// RegisterValidations records validations for a Config struct
func (c *ConfigManagerEntry) RegisterValidations(validations ...Validation) {
	c.Validations = append(c.Validations, validations...)
}

// RegisterDefaulters records defaults for a Config struct
func (c *ConfigManagerEntry) RegisterDefaulters(defaulters ...Defaulter) {
	c.Defaulters = append(c.Defaulters, defaulters...)
}

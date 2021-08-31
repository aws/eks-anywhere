package templater

import (
	"reflect"
	"strings"

	"sigs.k8s.io/yaml"
)

type PartialYaml map[string]interface{}

func (p PartialYaml) AddIfNotZero(k string, v interface{}) {
	if !isZeroVal(v) {
		p[k] = v
	}
}

func isZeroVal(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func (p PartialYaml) ToYaml() (string, error) {
	b, err := yaml.Marshal(p)
	if err != nil {
		return "", err
	}
	s := string(b)
	s = strings.TrimSuffix(s, "\n")

	return s, nil
}

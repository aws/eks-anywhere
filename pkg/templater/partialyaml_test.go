package templater_test

import (
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/templater"
)

func TestPartialYamlAddIfNotZero(t *testing.T) {
	tests := []struct {
		testName  string
		p         templater.PartialYaml
		k         string
		v         interface{}
		wantAdded bool
		wantV     interface{}
	}{
		{
			testName:  "add string",
			p:         templater.PartialYaml{},
			k:         "key",
			v:         "value",
			wantAdded: true,
			wantV:     "value",
		},
		{
			testName:  "add nil",
			p:         templater.PartialYaml{},
			k:         "key",
			v:         nil,
			wantAdded: false,
			wantV:     nil,
		},
		{
			testName:  "add empty string",
			p:         templater.PartialYaml{},
			k:         "key",
			v:         "",
			wantAdded: false,
			wantV:     nil,
		},
		{
			testName: "add present string",
			p: templater.PartialYaml{
				"key": "value_old",
			},
			k:         "key",
			v:         "value_new",
			wantAdded: true,
			wantV:     "value_new",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			tt.p.AddIfNotZero(tt.k, tt.v)

			gotV, gotAdded := tt.p[tt.k]
			if tt.wantAdded != gotAdded {
				t.Errorf("PartialYaml.AddIfNotZero() wasAdded = %v, wantAdded %v", gotAdded, tt.wantAdded)
			}

			if !reflect.DeepEqual(gotV, tt.wantV) {
				t.Errorf("PartialYaml.AddIfNotZero() gotValue = %v, wantValue %v", gotV, tt.wantV)
			}
		})
	}
}

func TestPartialYamlToYaml(t *testing.T) {
	tests := []struct {
		testName string
		p        templater.PartialYaml
		wantFile string
		wantErr  bool
	}{
		{
			testName: "simple object",
			p: templater.PartialYaml{
				"key1": "value 1",
				"key2": 2,
				"key3": "value3",
			},
			wantFile: "testdata/partial_yaml_object_expected.yaml",
			wantErr:  false,
		},
		{
			testName: "map",
			p: templater.PartialYaml{
				"key1": "value 1",
				"key2": 2,
				"key3": map[string]string{
					"key_nest1": "value nest",
					"key_nest2": "value nest 2",
				},
				"key4": map[string]interface{}{
					"key_nest1": "value nest",
					"key_nest2": 22,
				},
			},
			wantFile: "testdata/partial_yaml_map_expected.yaml",
			wantErr:  false,
		},
		{
			testName: "array",
			p: templater.PartialYaml{
				"key1": "value 1",
				"key2": 2,
				"key3": []string{"value array 1", "value array 2"},
				"key4": []interface{}{
					map[string]interface{}{
						"key_in_nest_array":   "value",
						"key_in_nest_array_2": 22,
					},
				},
			},
			wantFile: "testdata/partial_yaml_array_expected.yaml",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := tt.p.ToYaml()
			if (err != nil) != tt.wantErr {
				t.Fatalf("PartialYaml.ToYaml() error = %v, wantErr %v", err, tt.wantErr)
			}
			test.AssertContentToFile(t, got, tt.wantFile)
		})
	}
}

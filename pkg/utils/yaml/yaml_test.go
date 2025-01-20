package yaml_test

import (
	"testing"

	yamlutil "github.com/aws/eks-anywhere/pkg/utils/yaml"
)

func TestJoin(t *testing.T) {
	tests := []struct {
		name   string
		input  [][]byte
		output []byte
	}{
		{
			name:   "Empty input",
			input:  [][]byte{},
			output: []byte{},
		},
		{
			name: "Single document",
			input: [][]byte{
				[]byte(`apiVersion: v1
kind: Pod
metadata:
  name: pod-1
`),
			},
			output: []byte(`apiVersion: v1
kind: Pod
metadata:
  name: pod-1
`),
		},
		{
			name: "Multiple documents",
			input: [][]byte{
				[]byte(`apiVersion: v1
kind: Pod
metadata:
  name: pod-1
`),
				[]byte(`apiVersion: v1
kind: Service
metadata:
  name: service-1
`),
			},
			output: []byte(`apiVersion: v1
kind: Pod
metadata:
  name: pod-1

---
apiVersion: v1
kind: Service
metadata:
  name: service-1
`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			joinedDoc := yamlutil.Join(test.input)
			if string(joinedDoc) != string(test.output) {
				t.Errorf("Document mismatch.\nExpected:\n%s\nGot:\n%s", string(test.output), string(joinedDoc))
			}
		})
	}
}

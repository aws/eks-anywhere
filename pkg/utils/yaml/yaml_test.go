package yaml_test

import (
	"bufio"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	yamlutil "github.com/aws/eks-anywhere/pkg/utils/yaml"
)

func TestSplitDocuments(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedDocs [][]byte
		expectedErr  error
	}{
		{
			name:         "Empty input",
			input:        "",
			expectedDocs: [][]byte{},
			expectedErr:  nil,
		},
		{
			name: "Single document",
			input: `apiVersion: v1
kind: Pod
metadata:
  name: pod-1
`,
			expectedDocs: [][]byte{
				[]byte(`apiVersion: v1
kind: Pod
metadata:
  name: pod-1
`),
			},
			expectedErr: nil,
		},
		{
			name: "Multiple documents",
			input: `apiVersion: v1
kind: Pod
metadata:
  name: pod-1
---
apiVersion: v1
kind: Service
metadata:
  name: service-1
`,
			expectedDocs: [][]byte{
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
			expectedErr: nil,
		},
		{
			name:         "Error reading input 2",
			input:        `---\nkey: value\ninvalid_separator\n`,
			expectedDocs: nil,
			expectedErr:  errors.New("invalid Yaml document separator: \\nkey: value\\ninvalid_separator\\n"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := strings.NewReader(test.input)

			docs, err := yamlutil.SplitDocuments(bufio.NewReader(r))
			if test.expectedErr != nil {
				assert.Equal(t, test.expectedErr.Error(), err.Error())
				assert.Equal(t, len(test.expectedDocs), len(docs))
			} else {
				require.NoError(t, err)
				if len(docs) != len(test.expectedDocs) {
					t.Errorf("Expected %d documents, but got %d", len(test.expectedDocs), len(docs))
				}

				for i, doc := range docs {
					if string(doc) != string(test.expectedDocs[i]) {
						t.Errorf("Document %d mismatch.\nExpected:\n%s\nGot:\n%s", i+1, string(test.expectedDocs[i]), string(doc))
					}
				}
			}
		})
	}
}

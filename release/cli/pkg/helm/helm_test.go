// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryHost(t *testing.T) {
	tests := []struct {
		name     string
		remote   string
		expected string
	}{
		{
			name:     "private ECR registry",
			remote:   "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			expected: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
		},
		{
			name:     "public ECR alias",
			remote:   "public.ecr.aws/l0g8r8j6",
			expected: "public.ecr.aws",
		},
		{
			name:     "registry with port and repository path",
			remote:   "localhost:5000/charts",
			expected: "localhost:5000",
		},
		{
			name:     "OCI registry URI",
			remote:   "oci://public.ecr.aws/l0g8r8j6/charts",
			expected: "public.ecr.aws",
		},
		{
			name:     "IPv6 registry with port",
			remote:   "[2001:db8::1]:5000/charts",
			expected: "[2001:db8::1]:5000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			host, err := registryHost(test.remote)
			require.NoError(t, err)
			assert.Equal(t, test.expected, host)
		})
	}
}

func TestRegistryHostInvalidRemote(t *testing.T) {
	tests := []string{
		"",
		"oci:///charts",
		"public.ecr.aws:invalid-port/charts",
	}

	for _, remote := range tests {
		t.Run(remote, func(t *testing.T) {
			_, err := registryHost(remote)
			assert.Error(t, err)
		})
	}
}

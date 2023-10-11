package validations_test

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/pkg/validations"
)

func TestOldClusterConfigExists(t *testing.T) {
	tests := map[string]struct {
		Filename string
		Expect   bool
	}{
		"Non existence should return false": {
			Filename: "nonexistence",
			Expect:   false,
		},
		"Empty file should return false": {
			Filename: "empty",
			Expect:   false,
		},
		"Non empty file should return true": {
			Filename: "nonempty",
			Expect:   true,
		},
	}

	for tn, td := range tests {
		t.Run(tn, func(t *testing.T) {
			filename := filepath.Join("testdata", td.Filename)
			got := validations.FileExistsAndIsNotEmpty(filename)
			if td.Expect != got {
				t.Errorf("FileExistsAndIsNotEmpty(%v): want = %v; got = %v", filename, td.Expect, got)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		exists   bool
	}{
		{
			name:     "ExistingFile",
			filename: "testdata/testfile",
			exists:   true,
		},
		{
			name:     "NonExistenceFile",
			filename: "testdata/testfileNonExisting",
			exists:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			got := validations.FileExists(tc.filename)
			if tc.exists != got {
				t.Errorf("%v got = %v, want %v", tc.name, got, tc.exists)
			}
		})
	}
}

func TestValidateClusterNameArg(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError error
		expectedArg   string
	}{
		{
			name:          "Failure Empty Arguments",
			args:          []string{},
			expectedError: errors.New("please specify a cluster name"),
			expectedArg:   "",
		},
		{
			name:          "Success Non-empty Arguments",
			args:          []string{"test-cluster"},
			expectedError: nil,
			expectedArg:   "test-cluster",
		},
		{
			name:          "Failure Cluster Name",
			args:          []string{"test-cluster@123"},
			expectedError: errors.New("test-cluster@123 is not a valid cluster name, cluster names must start with lowercase/uppercase letters and can include numbers and dashes. For instance 'testCluster-123' is a valid name but '123testCluster' is not. "),
			expectedArg:   "test-cluster@123",
		},
		{
			name:          "Failure Cluster Length",
			args:          []string{"qwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnm12345"},
			expectedError: errors.New("number of characters in qwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnm12345 should be less than 81"),
			expectedArg:   "qwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnm12345",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			gotArgs, gotError := validations.ValidateClusterNameArg(tc.args)
			if !reflect.DeepEqual(tc.expectedError, gotError) || !reflect.DeepEqual(tc.expectedArg, gotArgs) {
				t.Errorf("\n%v got Error = %v, want Error %v", tc.name, gotError, tc.expectedError)
				t.Errorf("\n%v got Arguments = %v, want Arguments %v", tc.name, gotArgs, tc.expectedArg)

			}
		})
	}
}

func TestValidateClusterNameFromCommandAndConfig(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		clusterNameConfig string
		expectedError     error
	}{
		{
			name:              "Success cluster name match",
			args:              []string{"test-cluster"},
			clusterNameConfig: "test-cluster",
			expectedError:     nil,
		},
		{
			name:              "Success empty Arguments",
			args:              []string{},
			clusterNameConfig: "test-cluster",
			expectedError:     nil,
		},
		{
			name:              "Failure invalid cluster name",
			args:              []string{"123test-Cluster"},
			clusterNameConfig: "test-cluster",
			expectedError:     errors.New("please provide a valid <cluster-name>"),
		},
		{
			name:              "Failure cluster name not match",
			args:              []string{"test-cluster-1"},
			clusterNameConfig: "test-cluster",
			expectedError:     errors.New("please make sure cluster name provided in command matches with cluster name in config file"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			gotError := validations.ValidateClusterNameFromCommandAndConfig(tc.args, tc.clusterNameConfig)
			if !reflect.DeepEqual(tc.expectedError, gotError) {
				t.Errorf("\n%v got Error = %v, want Error %v", tc.name, gotError, tc.expectedError)
			}
		})
	}
}

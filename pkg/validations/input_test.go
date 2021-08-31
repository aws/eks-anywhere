package validations_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/pkg/validations"
)

func TestValidateClusterName(t *testing.T) {
	tests := []struct {
		clusterName, name string
		wantErr           error
	}{
		{
			name:        "FailureSpecialChars",
			clusterName: "test-cluster@123_",
			wantErr:     errors.New("test-cluster@123_ is not a valid cluster name, cluster names must start with lowercase/uppercase letters and can include numbers and dashes. For instance 'testCluster-123' is a valid name but '123testCluster' is not. "),
		},
		{
			name:        "FailureFirstCharNumeric",
			clusterName: "123test-Cluster",
			wantErr:     errors.New("123test-Cluster is not a valid cluster name, cluster names must start with lowercase/uppercase letters and can include numbers and dashes. For instance 'testCluster-123' is a valid name but '123testCluster' is not. "),
		},
		{
			name:        "SuccessUpperCaseChars",
			clusterName: "test-Cluster",
			wantErr:     nil,
		},
		{
			name:        "SuccessLowerCase",
			clusterName: "test-cluster",
			wantErr:     nil,
		},
		{
			name:        "SuccessLowerCaseDashNumeric",
			clusterName: "test-cluster123",
			wantErr:     nil,
		},
		{
			name:        "SuccessLowerCaseNumeric",
			clusterName: "test123cluster",
			wantErr:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			got := validations.ClusterName(tc.clusterName)
			if !reflect.DeepEqual(tc.wantErr, got) {
				t.Errorf("%v got = %v, want %v", tc.name, got, tc.wantErr)
			}
		})
	}
}

func TestClusterNameLength(t *testing.T) {
	tests := []struct {
		clusterName, name string
		wantErr           error
	}{

		{
			name:        "SuccessClusterNameLength",
			clusterName: "qwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnm",
			wantErr:     nil,
		},
		{
			name:        "FailureClusterNameLength",
			clusterName: "qwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnm12345",
			wantErr:     errors.New("number of characters in qwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnm12345 should be less than 81"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			got := validations.ClusterNameLength(tc.clusterName)
			if !reflect.DeepEqual(tc.wantErr, got) {
				t.Errorf("%v got = %v, want %v", tc.name, got, tc.wantErr)
			}
		})
	}
}

func TestOldClusterConfigExists(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		expect      bool
	}{
		{
			name:        "Non existence should return false",
			clusterName: "nonexistence",
			expect:      false,
		},
		{
			name:        "Empty file should return false",
			clusterName: "empty",
			expect:      false,
		},
		{
			name:        "Non Empty file should return true",
			clusterName: "nonempty",
			expect:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			got := validations.KubeConfigExists("testdata", tc.clusterName, "", "%s-eks-a-cluster.kubeconfig")
			if tc.expect != got {
				t.Errorf("%v got = %v, want %v", tc.name, got, tc.expect)
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

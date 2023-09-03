package framework

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetVsphereMachine(t *testing.T) {
	g := NewWithT(t)
	for _, tt := range []struct {
		testName        string
		machineName     string
		clusterName     string
		expectedWngName string
		err             error
	}{
		{
			testName:        "successfully get worker node group name",
			machineName:     "test-cluster-md-0-12345-12345",
			clusterName:     "test-cluster",
			expectedWngName: "md-0",
		},
		{
			testName:    "invalid machine name",
			machineName: "test-cluster-md-12345",
			clusterName: "test-cluster",
			err:         errors.New("invalid machine name test-cluster-md-12345"),
		},
	} {
		t.Run(tt.testName, func(t *testing.T) {
			gotWngName, err := getWngNameFromMachine(tt.machineName, tt.clusterName)
			if err != nil {
				g.Expect(err).To(MatchError(tt.err))
			} else {
				g.Expect(gotWngName).To(Equal(tt.expectedWngName))
			}
		})
	}
}

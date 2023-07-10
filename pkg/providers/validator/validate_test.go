package validator_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/networkutils/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/validator"
)

func TestValidateControlPlaneIPUniqueness(t *testing.T) {
	g := NewWithT(t)
	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
			},
		},
	}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockNetClient(ctrl)
	client.EXPECT().DialTimeout(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("no connection"))
	ipValidator := validator.NewIPValidator(validator.CustomNetClient(client))

	g.Expect(ipValidator.ValidateControlPlaneIPUniqueness(cluster)).To(Succeed())
}

func TestSkipValidateControlPlaneIPUniqueness(t *testing.T) {
	g := NewWithT(t)
	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
			},
		},
	}
	cluster.DisableControlPlaneIPCheck()
	ctrl := gomock.NewController(t)
	client := mocks.NewMockNetClient(ctrl)
	ipValidator := validator.NewIPValidator(validator.CustomNetClient(client))

	g.Expect(ipValidator.ValidateControlPlaneIPUniqueness(cluster)).To(BeNil())
}

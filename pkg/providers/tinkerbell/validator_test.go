package tinkerbell_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	networkutilsmocks "github.com/aws/eks-anywhere/pkg/networkutils/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	tinkerbellmocks "github.com/aws/eks-anywhere/pkg/providers/tinkerbell/mocks"
)

func TestValidateMinHardwareAvailableForCreate_SufficientHardware(t *testing.T) {
	for name, tt := range map[string]struct {
		AvailableHardware int
		ClusterSpec       v1alpha1.ClusterSpec
	}{
		"SufficientHardware":  {3, newValidClusterSpec(1, 1, 1)},
		"SuperfluousHardware": {5, newValidClusterSpec(1, 1, 1)},
	} {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			catalogue := newCatalogueWithHardware(tt.AvailableHardware)

			var validator tinkerbell.Validator

			assert.NoError(t, validator.ValidateMinHardwareAvailableForCreate(tt.ClusterSpec, catalogue))
		})
	}
}

func TestValidateMinHardwareAvailableForCreate_InsufficientHardware(t *testing.T) {
	clusterSpec := newValidClusterSpec(1, 1, 1)

	catalogue := newCatalogueWithHardware(2)

	var validator tinkerbell.Validator

	assert.Error(t, validator.ValidateMinHardwareAvailableForCreate(clusterSpec, catalogue))
}

func TestValidateMinHardware_EtcdUnspecified(t *testing.T) {
	clusterSpec := newValidClusterSpec(1, 0, 1)
	clusterSpec.ExternalEtcdConfiguration = nil

	catalogue := newCatalogueWithHardware(3)

	var validator tinkerbell.Validator

	assert.NoError(t, validator.ValidateMinHardwareAvailableForCreate(clusterSpec, catalogue))
}

func TestValidateTinkerbellConfig_ValidAuthorities(t *testing.T) {
	ctrl := gomock.NewController(t)
	pbnj := tinkerbellmocks.NewMockProviderPbnjClient(ctrl)
	tink := tinkerbellmocks.NewMockProviderTinkClient(ctrl)
	net := networkutilsmocks.NewMockNetClient(ctrl)

	datacenter := newValidTinkerbellDatacenterConfig()

	validator := tinkerbell.NewValidator(tink, net, pbnj)
	err := validator.ValidateTinkerbellConfig(context.Background(), datacenter)

	assert.NoError(t, err)
}

func TestValidateTinkerbellConfig_InvalidGrpcAuthority(t *testing.T) {
	cases := map[string]string{
		"Missing port":     "1.1.1.1",
		"Missing hostname": ":44",
		"Port is alpha":    "1.1.1.1:foo",
		"Port too large":   "1.1.1.1:99999",
	}

	for name, address := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			pbnj := tinkerbellmocks.NewMockProviderPbnjClient(ctrl)
			tink := tinkerbellmocks.NewMockProviderTinkClient(ctrl)
			net := networkutilsmocks.NewMockNetClient(ctrl)

			datacenter := newValidTinkerbellDatacenterConfig()
			datacenter.Spec.TinkerbellGRPCAuth = address

			validator := tinkerbell.NewValidator(tink, net, pbnj)
			err := validator.ValidateTinkerbellConfig(context.Background(), datacenter)

			assert.Error(t, err)
		})
	}
}

func TestValidateTinkerbellConfig_InvalidPbnjAuthority(t *testing.T) {
	cases := map[string]string{
		"Missing port":     "1.1.1.1",
		"Missing hostname": ":44",
		"Port is alpha":    "1.1.1.1:foo",
		"Port too large":   "1.1.1.1:99999",
	}

	for name, address := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			pbnj := tinkerbellmocks.NewMockProviderPbnjClient(ctrl)
			tink := tinkerbellmocks.NewMockProviderTinkClient(ctrl)
			net := networkutilsmocks.NewMockNetClient(ctrl)

			datacenter := newValidTinkerbellDatacenterConfig()
			datacenter.Spec.TinkerbellPBnJGRPCAuth = address

			validator := tinkerbell.NewValidator(tink, net, pbnj)
			err := validator.ValidateTinkerbellConfig(context.Background(), datacenter)

			assert.Error(t, err)
		})
	}
}

func newValidClusterSpec(cp, etcd, worker int) v1alpha1.ClusterSpec {
	return v1alpha1.ClusterSpec{
		ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
			Count: cp,
		},
		ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{
			Count: etcd,
		},
		WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{
			{Count: worker},
		},
	}
}

func newCatalogueWithHardware(hardwareCount int) *hardware.Catalogue {
	catalogue := hardware.NewCatalogue()
	for i := 0; i < hardwareCount; i++ {
		if err := catalogue.InsertHardware(&tinkv1alpha1.Hardware{}); err != nil {
			panic(err)
		}
	}
	return catalogue
}

func newValidTinkerbellDatacenterConfig() *v1alpha1.TinkerbellDatacenterConfig {
	return &v1alpha1.TinkerbellDatacenterConfig{
		Status: v1alpha1.TinkerbellDatacenterConfigStatus{},
		Spec: v1alpha1.TinkerbellDatacenterConfigSpec{
			TinkerbellIP:           "1.1.1.1",
			TinkerbellCertURL:      "http://1.1.1.1:444/path",
			TinkerbellGRPCAuth:     "1.1.1.1:444",
			TinkerbellPBnJGRPCAuth: "1.1.1.1:444",
			TinkerbellHegelURL:     "http://1.1.1.1:444",
		},
	}
}

package cloudstack

import (
	"context"
	"fmt"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/types"
)

// ValidatorRegistry exposes a single method for retrieving the CloudStack validator, and abstracts away how they are injected.
type ValidatorRegistry interface {
	Get(execConfig *decoder.CloudStackExecConfig) (ProviderValidator, error)
}

// CmkBuilder defines the interface to be consumed by the ValidatorFactory which enables it to build a new CloudStackClient.
type CmkBuilder interface {
	BuildCloudstackClient(writer filewriter.FileWriter, config *decoder.CloudStackExecConfig) (ProviderCmkClient, error)
}

// ValidatorFactory implements the ValidatorRegistry interface and holds the necessary structs for building fresh Validator objects.
type ValidatorFactory struct {
	builder     CmkBuilder
	writer      filewriter.FileWriter
	skipIPCheck bool
}

// ProviderValidator exposes a common interface to avoid coupling on implementation details and to support mocking.
type ProviderValidator interface {
	ValidateCloudStackDatacenterConfig(ctx context.Context, datacenterConfig *anywherev1.CloudStackDatacenterConfig) error
	ValidateClusterMachineConfigs(ctx context.Context, clusterSpec *cluster.Spec) error
	ValidateControlPlaneEndpointUniqueness(endpoint string) error
	ValidateSecretsUnchanged(ctx context.Context, cluster *types.Cluster, execConfig *decoder.CloudStackExecConfig, client ProviderKubectlClient) error
}

// NewValidatorFactory initializes a factory for the CloudStack provider validator.
func NewValidatorFactory(builder CmkBuilder, writer filewriter.FileWriter, skipIPCheck bool) ValidatorFactory {
	return ValidatorFactory{
		builder:     builder,
		writer:      writer,
		skipIPCheck: skipIPCheck,
	}
}

// Get returns a validator for a particular cloudstack exec config.
func (r ValidatorFactory) Get(execConfig *decoder.CloudStackExecConfig) (ProviderValidator, error) {
	cmk, err := r.builder.BuildCloudstackClient(r.writer, execConfig)
	if err != nil {
		return nil, fmt.Errorf("building cmk executable: %v", err)
	}

	return NewValidator(cmk, &networkutils.DefaultNetClient{}, r.skipIPCheck), nil
}

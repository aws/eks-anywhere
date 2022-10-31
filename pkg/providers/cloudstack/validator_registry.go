package cloudstack

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

// ValidatorRegistry exposes a single method for retrieving the CloudStack validator, and abstracts away how they are injected.
type ValidatorRegistry interface {
	Get(execConfig *decoder.CloudStackExecConfig) (*Validator, error)
}

// ValidatorFactory implements the ValidatorRegistry interface and holds the necessary structs for building fresh Validator objects.
type ValidatorFactory struct {
	builder     *executables.ExecutablesBuilder
	writer      filewriter.FileWriter
	skipIPCheck bool
}

// NewValidatorFactory initializes a factory for the CloudStack provider validator.
func NewValidatorFactory(builder *executables.ExecutablesBuilder, writer filewriter.FileWriter, skipIPCheck bool) ValidatorFactory {
	return ValidatorFactory{
		builder:     builder,
		writer:      writer,
		skipIPCheck: skipIPCheck,
	}
}

// Get returns a validator for a particular cloudstack exec config.
func (r ValidatorFactory) Get(execConfig *decoder.CloudStackExecConfig) (*Validator, error) {
	cmk, err := r.builder.BuildCmkExecutable(r.writer, execConfig)
	if err != nil {
		return nil, fmt.Errorf("building cmk executable: %v", err)
	}

	return NewValidator(cmk, &networkutils.DefaultNetClient{}, r.skipIPCheck), nil
}

package cmk

import (
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

// Builder serves as an interface wrapper to wrap the executablesBuilder without coupling consumers with its logic.
type Builder struct {
	builder *executables.ExecutablesBuilder
}

// NewCmkBuilder initializes the cmk executable builder.
func NewCmkBuilder(builder *executables.ExecutablesBuilder) *Builder {
	return &Builder{builder: builder}
}

// BuildCloudstackClient exposes a single method to consumers to abstract away executableBuilder's other operations and business logic.
func (b *Builder) BuildCloudstackClient(writer filewriter.FileWriter, config *decoder.CloudStackExecConfig) (cloudstack.ProviderCmkClient, error) {
	return b.builder.BuildCmkExecutable(writer, config)
}

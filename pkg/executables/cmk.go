package executables

import (
	"bytes"
	"context"
	_ "embed"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

const (
	cmkPath             = "cmk"
)

type Cmk struct {
	writer filewriter.FileWriter
	Executable
	retrier      *retrier.Retrier
	requiredEnvs *syncSlice
}

func (c *Cmk) ValidateCloudStackConnection(ctx context.Context) error {
	logger.V(6).Info("validating cloudstack connection")
	return nil
}

func (c *Cmk) ValidateServiceOfferingPresent(ctx context.Context, domain string, zone v1alpha1.CloudStackResourceRef, account string, serviceOffering v1alpha1.CloudStackResourceRef) error {
	logger.V(6).Info("validating service offering")
	return nil}

func (c *Cmk) ValidateTemplatePresent(ctx context.Context, domain string, zone v1alpha1.CloudStackResourceRef, account string, template v1alpha1.CloudStackResourceRef) error {
	logger.V(6).Info("validating template presence")
	return nil
}

func (c *Cmk) ValidateAffinityGroupsPresent(ctx context.Context, domain string, zone v1alpha1.CloudStackResourceRef, account string, affinityGroupIds []string) error {
	logger.V(6).Info("validating affinitygroup presence")
	return nil
}

func (c *Cmk) ValidateZonePresent(ctx context.Context, zone v1alpha1.CloudStackResourceRef) error {
	logger.V(6).Info("validating zone presence")
	return nil
}

func (c *Cmk) ValidateAccountPresent(ctx context.Context, account string) error {
	logger.V(6).Info("validating account presence")
	return nil
}

func NewCmk(executable Executable, writer filewriter.FileWriter) *Cmk {

	return &Cmk{
		writer:       writer,
		Executable:   executable,
		retrier:      retrier.NewWithMaxRetries(maxRetries, backOffPeriod),
		requiredEnvs: nil,
	}
}

func (g *Cmk) exec(ctx context.Context, args ...string) (stdout bytes.Buffer, err error) {
	return bytes.Buffer{}, nil
}

func (g *Cmk) Close(ctx context.Context) error {
	return nil
}

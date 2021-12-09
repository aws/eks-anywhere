package e2e

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

const (
	vsphereHostVar        = "T_VSPHERE_HOST"
	cidrVar               = "T_VSPHERE_CIDR"
	privateNetworkCidrVar = "T_VSPHERE_PRIVATE_NETWORK_CIDR"
	vsphereRegex          = `^.*VSphere.*$`
)

func (e *E2ESession) setupVSphereEnv(testRegex string) error {
	re := regexp.MustCompile(vsphereRegex)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running VSphere tests, skipping Env variable setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredVsphereEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}
	if e.controlPlaneIP != "" {
		e.testEnvVars[vsphereHostVar] = e.controlPlaneIP
	}

	return nil
}

func vsphereRmVms(ctx context.Context, clusterName string) error {
	logger.V(1).Info("Deleting vsphere vcenter vms")
	executableBuilder, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	tmpWriter, _ := filewriter.NewWriter("rmvms")
	govc := executableBuilder.BuildGovcExecutable(tmpWriter)
	defer govc.Close(ctx)

	return govc.CleanupVms(ctx, clusterName, false)
}

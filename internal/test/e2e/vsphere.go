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

func (e *E2ESession) setupVSphereEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*VSphere.*$`)
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

	return nil
}

func vsphereRmVms(ctx context.Context, clusterName string) error {
	logger.V(1).Info("Deleting vsphere vcenter vms")
	executableBuilder, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	tmpWriter, _ := filewriter.NewWriter("rmvms")
	return executableBuilder.BuildGovcExecutable(tmpWriter).CleanupVms(ctx, clusterName, false)
}

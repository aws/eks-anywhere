package analyzer

import (
	"context"
	"fmt"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func (a *Analyzer) analyzeControlPlane(ctx context.Context, cluster *anywherev1.Cluster) ([]Finding, error) {
	kcpName := clusterapi.KubeadmControlPlaneName(cluster)
	finding, err := run(ctx, a.readers, newKCP(kcpName, constants.EksaSystemNamespace))
	if err != nil {
		return nil, err
	}

	if finding == nil {
		return nil, nil
	}

	controlPlaneCondition := condition(cluster, clusterv1.ControlPlaneReadyCondition)

	return []Finding{
		{
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("The %s is %s: %s", bold("Control Plane"), yellow("not ready"), controlPlaneCondition.Message),
			Findings: []Finding{*finding},
		},
	}, nil
}

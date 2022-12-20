package snow

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

func CredentialsSecretName(clusterSpec *cluster.Spec) string {
	return fmt.Sprintf("%s-snow-credentials", clusterSpec.Cluster.GetName())
}

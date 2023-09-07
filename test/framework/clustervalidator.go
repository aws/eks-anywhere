package framework

import (
	"context"
	"fmt"
	"time"

	machinerytypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/retrier"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
	"github.com/aws/eks-anywhere/test/framework/cluster/validations"
)

func validationsForExpectedObjects() []clusterf.StateValidation {
	mediumRetier := retrier.New(10 * time.Minute)
	longRetier := retrier.New(30 * time.Minute)
	return []clusterf.StateValidation{
		clusterf.RetriableStateValidation(mediumRetier, validations.ValidateEKSAObjects),
		clusterf.RetriableStateValidation(longRetier, validations.ValidateControlPlaneNodes),
		clusterf.RetriableStateValidation(longRetier, validations.ValidateWorkerNodes),
		clusterf.RetriableStateValidation(mediumRetier, validations.ValidateCilium),
		// This should be checked last as the Cluster should only be ready after all the other validations pass.
		clusterf.RetriableStateValidation(mediumRetier, validations.ValidateClusterReady),
	}
}

func validationsForClusterDoesNotExist() []clusterf.StateValidation {
	return []clusterf.StateValidation{
		clusterf.RetriableStateValidation(retrier.NewWithMaxRetries(120, 5*time.Second), validations.ValidateClusterDoesNotExist),
	}
}

func (e *ClusterE2ETest) buildClusterStateValidationConfig(ctx context.Context) {
	managementClusterClient, err := buildClusterClient(e.managementKubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("failed to create management cluster client: %s", err)
	}
	clusterClient := managementClusterClient
	if e.managementKubeconfigFilePath() != e.KubeconfigFilePath() {
		clusterClient, err = buildClusterClient(e.KubeconfigFilePath())
	}
	if err != nil {
		e.T.Fatalf("failed to create cluster client: %s", err)
	}
	spec, err := buildClusterSpec(ctx, managementClusterClient, e.ClusterConfig)
	if err != nil {
		e.T.Fatalf("failed to build cluster spec with kubeconfig %s: %v", e.KubeconfigFilePath(), err)
	}

	e.clusterStateValidationConfig = &clusterf.StateValidationConfig{
		ClusterClient:           clusterClient,
		ManagementClusterClient: managementClusterClient,
		ClusterSpec:             spec,
	}
}

func newClusterStateValidator(config *clusterf.StateValidationConfig) *clusterf.StateValidator {
	return clusterf.NewStateValidator(*config)
}

func buildClusterClient(kubeconfigFileName string) (client.Client, error) {
	var clusterClient client.Client
	// Adding the retry logic here because the connection to the client does not always
	// succedd on the first try due to connection failure after the kubeconfig becomes
	// available in the cluster.
	err := retrier.Retry(12, 5*time.Second, func() error {
		c, err := kubernetes.NewRuntimeClientFromFileName(kubeconfigFileName)
		if err != nil {
			return fmt.Errorf("failed to build cluster client: %v", err)
		}
		clusterClient = c
		return nil
	})

	return clusterClient, err
}

func buildClusterSpec(ctx context.Context, client client.Client, config *cluster.Config) (*cluster.Spec, error) {
	clusterConfig := config.DeepCopy()
	// The cluster config built by the test does not have certain defaults like the bundle reference,
	// so fetch that information from the cluster if missing. This is needed inorder to build the cluster spec.
	if clusterConfig.Cluster.Spec.BundlesRef == nil {
		clus := &v1alpha1.Cluster{}
		key := machinerytypes.NamespacedName{Namespace: clusterConfig.Cluster.Namespace, Name: clusterConfig.Cluster.Name}
		if err := client.Get(ctx, key, clus); err != nil {
			return nil, fmt.Errorf("failed to get cluster to build spec: %s", err)
		}
		clusterConfig.Cluster.Spec.BundlesRef = clus.Spec.BundlesRef
		if clusterConfig.Cluster.Spec.BundlesRef == nil {
			clusterConfig.Cluster.Spec.EksaVersion = clus.Spec.EksaVersion
		}
	}
	spec, err := cluster.BuildSpecFromConfig(ctx, clientutil.NewKubeClient(client), clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build cluster spec from config: %s", err)
	}
	return spec, nil
}

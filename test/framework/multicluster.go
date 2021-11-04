package framework

import (
	"fmt"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

type MulticlusterE2ETest struct {
	T                 *testing.T
	ManagementCluster *ClusterE2ETest
	WorkloadClusters  WorkloadClusters
}

func NewMulticlusterE2ETest(t *testing.T, managementCluster *ClusterE2ETest, workloadClusters ...*ClusterE2ETest) *MulticlusterE2ETest {
	m := &MulticlusterE2ETest{
		T:                 t,
		ManagementCluster: managementCluster,
	}

	m.WorkloadClusters = make(WorkloadClusters, len(workloadClusters))
	for i, c := range workloadClusters {
		c.clusterFillers = append(c.clusterFillers, api.WithManagementCluster(managementCluster.ClusterName))
		c.ClusterName = fmt.Sprintf("%s-w-%d", managementCluster.ClusterName, i)
		m.WorkloadClusters[c.ClusterName] = &WorkloadCluster{
			ClusterE2ETest:                  c,
			managementClusterKubeconfigFile: managementCluster.kubeconfigFilePath,
		}
	}

	return m
}

func (m *MulticlusterE2ETest) RunInWorkloadClusters(flow func(*WorkloadCluster)) {
	for name, w := range m.WorkloadClusters {
		m.T.Logf("Running test flow in workload cluster %s", name)
		flow(w)
	}
}

func (m *MulticlusterE2ETest) CreateManagementCluster() {
	m.ManagementCluster.GenerateClusterConfig()
	m.ManagementCluster.CreateCluster()
}

func (m *MulticlusterE2ETest) DeleteManagementCluster() {
	m.ManagementCluster.DeleteCluster()
}

package nutanix

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/nutanix-cloud-native/cluster-api-provider-nutanix/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	capiyaml "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type (
	// Workers represents the nutanix specific CAPI spec for worker nodes.
	Workers        = clusterapi.Workers[*v1beta1.NutanixMachineTemplate]
	workersBuilder = capiyaml.WorkersBuilder[*v1beta1.NutanixMachineTemplate]
)

// WorkersSpec generates a nutanix specific CAPI spec for an eks-a cluster worker nodes.
// It talks to the cluster with a client to detect changes in immutable objects and generates new
// names for them.
func WorkersSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, spec *cluster.Spec) (*Workers, error) {
	ndcs := spec.NutanixDatacenter.Spec
	machineConfigs := spec.NutanixMachineConfigs

	wnmcs := make(map[string]v1alpha1.NutanixMachineConfigSpec, len(spec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, machineConfig := range machineConfigs {
		machineConfig.SetDefaults()
	}

	creds := GetCredsFromEnv()
	templateBuilder := NewNutanixTemplateBuilder(&ndcs, nil, nil, wnmcs, creds, time.Now)
	workloadTemplateNames, kubeadmconfigTemplateNames := getTemplateNames(spec, templateBuilder, machineConfigs)

	workersYaml, err := templateBuilder.GenerateCAPISpecWorkers(spec, workloadTemplateNames, kubeadmconfigTemplateNames)
	if err != nil {
		return nil, err
	}

	workers, err := parseWorkersYaml(logger, workersYaml)
	if err != nil {
		return nil, fmt.Errorf("parsing nutanix CAPI workers yaml: %w", err)
	}

	if err = workers.UpdateImmutableObjectNames(ctx, client, getMachineTemplate, machineTemplateEquals); err != nil {
		return nil, fmt.Errorf("updating nutanix worker immutable object names: %w", err)
	}

	return workers, nil
}

func getTemplateNames(spec *cluster.Spec, templateBuilder *TemplateBuilder, machineConfigs map[string]*v1alpha1.NutanixMachineConfig) (map[string]string, map[string]string) {
	workloadTemplateNames := make(map[string]string, len(spec.Cluster.Spec.WorkerNodeGroupConfigurations))
	kubeadmconfigTemplateNames := make(map[string]string, len(spec.Cluster.Spec.WorkerNodeGroupConfigurations))

	for _, workerNodeGroupConfiguration := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workloadTemplateNames[workerNodeGroupConfiguration.Name] = common.WorkerMachineTemplateName(spec.Cluster.Name, workerNodeGroupConfiguration.Name, templateBuilder.now)
		kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = common.KubeadmConfigTemplateName(spec.Cluster.Name, workerNodeGroupConfiguration.Name, templateBuilder.now)
		templateBuilder.workerNodeGroupMachineSpecs[workerNodeGroupConfiguration.MachineGroupRef.Name] = machineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec
	}

	return workloadTemplateNames, kubeadmconfigTemplateNames
}

func parseWorkersYaml(logger logr.Logger, workersYaml []byte) (*Workers, error) {
	parser, builder, err := newWorkersParserAndBuilder(logger)
	if err != nil {
		return nil, err
	}

	if err = parser.Parse(workersYaml, builder); err != nil {
		return nil, fmt.Errorf("parsing nutanix CAPI workers yaml: %w", err)
	}

	return builder.Workers, nil
}

func newWorkersParserAndBuilder(logger logr.Logger) (*yamlutil.Parser, *workersBuilder, error) {
	parser, builder, err := capiyaml.NewWorkersParserAndBuilder(
		logger,
		machineTemplateMapping(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("building nutanix workers parser and builder: %w", err)
	}

	return parser, builder, nil
}

func machineTemplateMapping() yamlutil.Mapping[*v1beta1.NutanixMachineTemplate] {
	return yamlutil.NewMapping(
		"NutanixMachineTemplate",
		func() *v1beta1.NutanixMachineTemplate {
			return &v1beta1.NutanixMachineTemplate{}
		},
	)
}

func getMachineTemplate(ctx context.Context, client kubernetes.Client, name, namespace string) (*v1beta1.NutanixMachineTemplate, error) {
	m := &v1beta1.NutanixMachineTemplate{}
	if err := client.Get(ctx, name, namespace, m); err != nil {
		return nil, fmt.Errorf("reading nutanixMachineTemplate: %w", err)
	}

	return m, nil
}

func machineTemplateEquals(new, old *v1beta1.NutanixMachineTemplate) bool {
	return equality.Semantic.DeepDerivative(new.Spec, old.Spec)
}

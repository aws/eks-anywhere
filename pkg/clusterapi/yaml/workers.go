package yaml

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

const machineDeploymentKind = "MachineDeployment"

// WorkersBuilder implements yamlutil.Builder
// It's a wrapper around Workers to provide yaml parsing functionality.
type WorkersBuilder[M clusterapi.Object[M]] struct {
	Workers *clusterapi.Workers[M]
}

// NewWorkersBuilder builds a WorkersBuilder.
func NewWorkersBuilder[M clusterapi.Object[M]]() *WorkersBuilder[M] {
	return &WorkersBuilder[M]{
		Workers: new(clusterapi.Workers[M]),
	}
}

// BuildFromParsed reads parsed objects in ObjectLookup and sets them in the Workers.
func (cp *WorkersBuilder[M]) BuildFromParsed(lookup yamlutil.ObjectLookup) error {
	ProcessWorkerObjects(cp.Workers, lookup)
	return nil
}

// NewWorkersParserAndBuilder builds a Parser and a Builder for a particular provider Workers
// It registers the basic shared mappings plus another one for the provider machine template
// For worker specs that need to include more objects, wrap around the provider builder and
// implement BuildFromParsed.
// Any extra mappings will need to be registered manually in the Parser.
func NewWorkersParserAndBuilder[M clusterapi.Object[M]](
	logger logr.Logger,
	machineTemplateMapping yamlutil.Mapping[M],
) (*yamlutil.Parser, *WorkersBuilder[M], error) {
	parser := yamlutil.NewParser(logger)
	if err := RegisterWorkerMappings(parser); err != nil {
		return nil, nil, errors.Wrap(err, "building capi worker parser")
	}

	err := parser.RegisterMappings(
		machineTemplateMapping.ToAPIObjectMapping(),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "registering provider worker mappings")
	}

	return parser, NewWorkersBuilder[M](), nil
}

// RegisterWorkerMappings records the basic mappings for CAPI MachineDeployment
// and KubeadmConfigTemplate in a Parser.
func RegisterWorkerMappings(parser *yamlutil.Parser) error {
	err := parser.RegisterMappings(
		yamlutil.NewMapping(
			machineDeploymentKind, func() yamlutil.APIObject {
				return &clusterv1.MachineDeployment{}
			},
		),
		yamlutil.NewMapping(
			"KubeadmConfigTemplate", func() yamlutil.APIObject {
				return &kubeadmv1.KubeadmConfigTemplate{}
			},
		),
	)
	if err != nil {
		return errors.Wrap(err, "registering base worker mappings")
	}

	return nil
}

// ProcessWorkerObjects finds all necessary objects in the parsed objects and sets them in Workers.
func ProcessWorkerObjects[M clusterapi.Object[M]](w *clusterapi.Workers[M], lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == machineDeploymentKind {
			g := new(clusterapi.WorkerGroup[M])
			g.MachineDeployment = obj.(*clusterv1.MachineDeployment)
			ProcessWorkerGroupObjects(g, lookup)
			w.Groups = append(w.Groups, *g)
		}
	}
}

// ProcessWorkerGroupObjects looks in the parsed objects for the KubeadmConfigTemplate and
// the provider machine template referenced in the MachineDeployment and sets them in the WorkerGroup.
// MachineDeployment needs to be already set in the WorkerGroup.
func ProcessWorkerGroupObjects[M clusterapi.Object[M]](g *clusterapi.WorkerGroup[M], lookup yamlutil.ObjectLookup) {
	kubeadmConfigTemplate := lookup.GetFromRef(*g.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef)
	if kubeadmConfigTemplate != nil {
		g.KubeadmConfigTemplate = kubeadmConfigTemplate.(*kubeadmv1.KubeadmConfigTemplate)
	}

	machineTemplate := lookup.GetFromRef(g.MachineDeployment.Spec.Template.Spec.InfrastructureRef)
	if machineTemplate != nil {
		g.ProviderMachineTemplate = machineTemplate.(M)
	}
}

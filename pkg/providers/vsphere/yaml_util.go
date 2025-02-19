package vsphere

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"

	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

const (
	// VSphereFailureDomainKind is kind for capv failure domain.
	VSphereFailureDomainKind = "VSphereFailureDomain"
	// VSphereDeploymentZoneKind is kind for capv vsphere deployment zone.
	VSphereDeploymentZoneKind = "VSphereDeploymentZone"
)

// FailureDomainsYamlProcessor handles parsing and transformation of failure domains YAML into FailureDomains Objects.
type FailureDomainsYamlProcessor struct {
	parser                *yamlutil.Parser
	failureDomainsBuilder *FailureDomainsBuilder
}

// FailureDomainsBuilder implements yamlutil.Builder
// It's a wrapper around FailureDomains to provide yaml parsing functionality.
type FailureDomainsBuilder struct {
	FailureDomains *FailureDomains
}

// NewFailureDomainsYamlProcessor initializes and returns a new FailureDomainsYamlProcessor.
func NewFailureDomainsYamlProcessor(logger logr.Logger) *FailureDomainsYamlProcessor {
	parser := yamlutil.NewParser(logger)
	return &FailureDomainsYamlProcessor{
		parser: parser,
		failureDomainsBuilder: &FailureDomainsBuilder{
			FailureDomains: new(FailureDomains),
		},
	}
}

// ProcessYAML processes the YAML content and returns the failure domains.
func (fbp *FailureDomainsYamlProcessor) ProcessYAML(failureDomainYaml []byte) (*FailureDomains, error) {
	if err := fbp.registerFailureDomainMappings(); err != nil {
		return nil, errors.Wrap(err, "registering failure domains mappings in the yaml processor")
	}

	if err := fbp.parser.Parse(failureDomainYaml, fbp.failureDomainsBuilder); err != nil {
		return nil, errors.Wrap(err, "parsing vsphere failure domains yaml")
	}

	return fbp.failureDomainsBuilder.FailureDomains, nil
}

func (fbp *FailureDomainsYamlProcessor) registerFailureDomainMappings() error {
	err := fbp.parser.RegisterMappings(
		yamlutil.NewMapping(
			VSphereDeploymentZoneKind, func() yamlutil.APIObject {
				return &vspherev1.VSphereDeploymentZone{}
			},
		),
		yamlutil.NewMapping(
			VSphereFailureDomainKind, func() yamlutil.APIObject {
				return &vspherev1.VSphereFailureDomain{}
			},
		),
	)
	if err != nil {
		return errors.Wrap(err, "registering failure domain mappings")
	}

	return nil
}

// BuildFromParsed reads parsed objects in ObjectLookup and sets them in the FailureDomains.
func (fb *FailureDomainsBuilder) BuildFromParsed(lookup yamlutil.ObjectLookup) error {
	fb.ProcessFailureDomainObjects(fb.FailureDomains, lookup)
	return nil
}

// ProcessFailureDomainObjects finds all necessary objects in the parsed objects and sets them in FailureDomains.
func (fb *FailureDomainsBuilder) ProcessFailureDomainObjects(f *FailureDomains, lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == VSphereDeploymentZoneKind {
			g := new(FailureDomainGroup)
			g.VsphereDeploymentZone = obj.(*vspherev1.VSphereDeploymentZone)
			fb.ProcessFailureDomainGroupObjects(g, lookup)
			f.Groups = append(f.Groups, *g)
		}
	}
}

// ProcessFailureDomainGroupObjects looks in the parsed objects for the VsphereFailureDomain by the kind, apiversion
// and VsphereFailureDomain name. Once it is found, it sets in the FailureDomainGroup.
// VsphereDeploymentZone needs to be already set in the FailureDomainGroup.
func (fb *FailureDomainsBuilder) ProcessFailureDomainGroupObjects(g *FailureDomainGroup, lookup yamlutil.ObjectLookup) {
	vsphereFailureDomainName := g.VsphereDeploymentZone.Spec.FailureDomain
	vsphereFailureDomainKey := yamlutil.Key(g.VsphereDeploymentZone.APIVersion, VSphereFailureDomainKind, vsphereFailureDomainName)
	obj := lookup[vsphereFailureDomainKey]
	if obj != nil {
		g.VsphereFailureDomain = obj.(*vspherev1.VSphereFailureDomain)
	}
}

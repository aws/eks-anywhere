package vsphere

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"

	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

const VSphereFailureDomainKind = "VSphereFailureDomain"
const VSphereDeploymentZoneKind = "VSphereDeploymentZone"

// FailureDomainsBuilder implements yamlutil.Builder
// It's a wrapper around FailureDomains to provide yaml parsing functionality.
type FailureDomainsBuilder struct {
	FailureDomains *FailureDomains
}

// NewFailureDomainsBuilder builds a NewFailureDomainsBuilder.
func NewFailureDomainsBuilder() *FailureDomainsBuilder {
	return &FailureDomainsBuilder{
		FailureDomains: new(FailureDomains),
	}
}

// BuildFromParsed reads parsed objects in ObjectLookup and sets them in the FailureDomains.
func (fb *FailureDomainsBuilder) BuildFromParsed(lookup yamlutil.ObjectLookup) error {
	ProcessFailureDomainObjects(fb.FailureDomains, lookup)
	return nil
}

// NewFailureDomainsParserAndBuilder builds a Parser and a Builder for a Failure Domains
func NewFailureDomainsParserAndBuilder(logger logr.Logger) (*yamlutil.Parser, *FailureDomainsBuilder, error) {
	parser := yamlutil.NewParser(logger)
	if err := RegisterFailureDomainMappings(parser); err != nil {
		return nil, nil, errors.Wrap(err, "building vSphere failure domains parser")
	}

	return parser, NewFailureDomainsBuilder(), nil
}

// RegisterFailureDomainMappings records the basic mappings for VSphereDeploymentZone
// and VSphereFailureDomain in a Parser.
func RegisterFailureDomainMappings(parser *yamlutil.Parser) error {
	err := parser.RegisterMappings(
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

// ProcessFailureDomainObjects finds all necessary objects in the parsed objects and sets them in FailureDomains.
func ProcessFailureDomainObjects(f *FailureDomains, lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == VSphereDeploymentZoneKind {
			g := new(FailureDomainGroup)
			g.VsphereDeploymentZone = obj.(*vspherev1.VSphereDeploymentZone)
			ProcessFailureDomainGroupObjects(g, lookup)
			f.Groups = append(f.Groups, *g)
		}
	}
}

// ProcessFailureDomainGroupObjects looks in the parsed objects for the VsphereFailureDomain by the kind, apiversion 
// and VsphereFailureDomain name. Once it is found, it sets in the FailureDomainGroup.
// VsphereDeploymentZone needs to be already set in the FailureDomainGroup.
func ProcessFailureDomainGroupObjects(g *FailureDomainGroup, lookup yamlutil.ObjectLookup) {
	vsphereFailureDomainName := g.VsphereDeploymentZone.Spec.FailureDomain
	vsphereFailureDomainKey := yamlutil.Key(g.VsphereDeploymentZone.APIVersion, VSphereFailureDomainKind, vsphereFailureDomainName)
	obj := lookup[vsphereFailureDomainKey]
	if obj != nil {
		g.VsphereFailureDomain = obj.(*vspherev1.VSphereFailureDomain)
	}
}

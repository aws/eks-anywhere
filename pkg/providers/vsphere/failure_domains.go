package vsphere

import (
	"fmt"
	"time"

	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"

	"github.com/go-logr/logr"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

// FailureDomains represents the list of failure domain groups.
type FailureDomains struct {
	Groups []FailureDomainGroup
}

// FailureDomainGroup represents the Vsphere failure domains objects group.
type FailureDomainGroup struct {
	VsphereDeploymentZone   *vspherev1.VSphereDeploymentZone
	VsphereFailureDomain    *vspherev1.VSphereFailureDomain
}

// objects returns a list of API objects for a collection of failure domain groups.
func (f *FailureDomains) Objects() []client.Object {
	objs := make([]client.Object, 0, len(f.Groups)*2)
	for _, g := range f.Groups {

		objs = append(objs, g.objects()...)
	}
	return objs
}

func (fg *FailureDomainGroup) objects() []client.Object {
	return []client.Object{fg.VsphereFailureDomain, fg.VsphereDeploymentZone}
}

func templateNamesForFailureDomains(spec *cluster.Spec) map[string]string {
	failureDomainsLen := len(spec.VSphereDatacenter.Spec.FailureDomains)
	templateNames := make(map[string]string, failureDomainsLen)
	for _, failureDomain := range spec.VSphereDatacenter.Spec.FailureDomains {
		failureDomainTemplateName := BuildFailureDomainTemplateName(spec, failureDomain.Name)
		templateNames[failureDomain.Name] = failureDomainTemplateName
	}
	return templateNames
}

func BuildFailureDomainTemplateName(spec *cluster.Spec, failureDomainName string) string {
	return fmt.Sprintf("%s-%s-%s", spec.Cluster.Name, spec.VSphereDatacenter.Name, failureDomainName)
}

// FailureDomainsSpec generates a vSphere Failure domains spec for the cluster.
func FailureDomainsSpec(logger logr.Logger, spec *cluster.Spec) (*FailureDomains, error){
	templateBuilder := NewVsphereTemplateBuilder(time.Now)
	templateNames := templateNamesForFailureDomains(spec)

	failureDomainYaml, err := templateBuilder.GenerateVsphereFailureDomainsSpec(spec, templateNames)
	if err != nil {
		return nil, err
	}
	parser, builder, err := newFailureDomainsParserAndBuilder(logger)
	if err != nil {
		return nil, err
	}

	if err = parser.Parse(failureDomainYaml, builder); err != nil {
		return nil, errors.Wrap(err, "parsing vSphere Failure Domaines yaml")
	}

	failureDomains := builder.FailureDomains

	return failureDomains, nil
}



func newFailureDomainsParserAndBuilder(logger logr.Logger) (*yamlutil.Parser, *FailureDomainsBuilder, error) {
	parser, builder, err := NewFailureDomainsParserAndBuilder(logger)
	if err != nil {
		return nil, nil, errors.Wrap(err, "building vSphere workers parser and builder")
	}

	return parser, builder, nil
}
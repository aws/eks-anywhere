package vsphere

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

// FailureDomains represents the list of failure domain groups.
type FailureDomains struct {
	Groups []FailureDomainGroup
}

// FailureDomainGroup represents the Vsphere failure domains objects group.
type FailureDomainGroup struct {
	VsphereDeploymentZone *vspherev1.VSphereDeploymentZone
	VsphereFailureDomain  *vspherev1.VSphereFailureDomain
}

const (
	// VsphereDataCenterConfigNameLabel is label for VsphereDataCenter name in Cluster.Spec.VsphereDataCenter.Name.
	VsphereDataCenterConfigNameLabel = "infrastructure.cluster.x-k8s.io/vsphere-datacenter-config-name"
	// ClusterNameLabel is label for cluster name.
	ClusterNameLabel = "infrastructure.cluster.x-k8s.io/cluster-name"
)

// Objects returns a list of API objects for a collection of failure domain groups.
func (f *FailureDomains) Objects() []client.Object {
	objs := make([]client.Object, 0, len(f.Groups)*2)
	for _, g := range f.Groups {
		failureDomainGroupObjects := []client.Object{g.VsphereFailureDomain, g.VsphereDeploymentZone}
		objs = append(objs, failureDomainGroupObjects...)
	}
	return objs
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

// BuildFailureDomainTemplateName generates the template name for failure domain.
func BuildFailureDomainTemplateName(spec *cluster.Spec, failureDomainName string) string {
	return fmt.Sprintf("%s-%s-%s", spec.Cluster.Name, spec.VSphereDatacenter.Name, failureDomainName)
}

// FailureDomainsSpec generates a vSphere Failure domains spec for the cluster.
func FailureDomainsSpec(logger logr.Logger, spec *cluster.Spec) (*FailureDomains, error) {
	templateBuilder := NewVsphereTemplateBuilder(time.Now)
	templateNames := templateNamesForFailureDomains(spec)
	failureDomainYaml, err := templateBuilder.GenerateVsphereFailureDomainsSpec(spec, templateNames)
	if err != nil {
		return nil, err
	}

	yamlProcessor := NewFailureDomainsYamlProcessor(logger)

	failureDomains, err := yamlProcessor.ProcessYAML(failureDomainYaml)
	if err != nil {
		return nil, errors.Wrap(err, "processing vsphere failure domains yaml")
	}

	return failureDomains, nil
}

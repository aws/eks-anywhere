package supportbundle

import (
	"fmt"

	"github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type analyzerFactory struct{}

func NewAnalyzerFactory() *analyzerFactory {
	return &analyzerFactory{}
}

func (a *analyzerFactory) DefaultAnalyzers() []*v1beta2.Analyze {
	return append(a.defaultDeploymentAnalyzers(), a.defaultCrdAnalyzers()...)
}

func (a *analyzerFactory) defaultDeploymentAnalyzers() []*v1beta2.Analyze {
	d := []eksaDeployment{
		{
			Name:             "capv-controller-manager",
			Namespace:        "capi-webhook-system",
			ExpectedReplicas: 1,
		}, {
			Name:             "capv-controller-manager",
			Namespace:        "capv-system",
			ExpectedReplicas: 1,
		}, {
			Name:             "coredns",
			Namespace:        "kube-system",
			ExpectedReplicas: 2,
		}, {
			Name:             "cert-manager-webhook",
			Namespace:        "cert-manager",
			ExpectedReplicas: 1,
		}, {
			Name:             "cert-manager-cainjector",
			Namespace:        "cert-manager",
			ExpectedReplicas: 1,
		}, {
			Name:             "cert-manager",
			Namespace:        "cert-manager",
			ExpectedReplicas: 1,
		}, {
			Name:             "capi-kubeadm-control-plane-controller-manager",
			Namespace:        "capi-webhook-system",
			ExpectedReplicas: 1,
		}, {
			Name:             "capi-kubeadm-bootstrap-controller-manager",
			Namespace:        "capi-webhook-system",
			ExpectedReplicas: 1,
		}, {
			Name:             "capi-controller-manager",
			Namespace:        "capi-webhook-system",
			ExpectedReplicas: 1,
		}, {
			Name:             "capi-controller-manager",
			Namespace:        "capi-system",
			ExpectedReplicas: 1,
		}, {
			Name:             "capi-kubeadm-control-plane-controller-manager",
			Namespace:        "capi-kubeadm-control-plane-system",
			ExpectedReplicas: 1,
		}, {
			Name:             "capi-kubeadm-control-plane-controller-manager",
			Namespace:        "capi-kubeadm-control-plane-system",
			ExpectedReplicas: 1,
		}, {
			Name:             "capi-kubeadm-bootstrap-controller-manager",
			Namespace:        "capi-kubeadm-bootstrap-system",
			ExpectedReplicas: 1,
		},
	}
	return a.generateDeploymentAnalyzers(d)
}

func (a *analyzerFactory) defaultCrdAnalyzers() []*v1beta2.Analyze {
	crds := []string{
		fmt.Sprintf("clusters.%s", v1alpha1.GroupVersion.Group),
		fmt.Sprintf("bundles.%s", v1alpha1.GroupVersion.Group),
	}
	return a.generateCrdAnalyzers(crds)
}

func (a *analyzerFactory) EksaGitopsAnalyzers() []*v1beta2.Analyze {
	crds := []string{
		fmt.Sprintf("gitopsconfigs.%s", v1alpha1.GroupVersion.Group),
	}
	return a.generateCrdAnalyzers(crds)
}

func (a *analyzerFactory) EksaOidcAnalyzers() []*v1beta2.Analyze {
	crds := []string{
		fmt.Sprintf("oidcconfigs.%s", v1alpha1.GroupVersion.Group),
	}
	return a.generateCrdAnalyzers(crds)
}

func (a *analyzerFactory) EksaExternalEtcdAnalyzers() []*v1beta2.Analyze {
	deployments := []eksaDeployment{
		{
			Name:             "etcdadm-controller-controller-manager",
			Namespace:        "etcdadm-controller-system",
			ExpectedReplicas: 1,
		}, {
			Name:             "etcdadm-bootstrap-provider-controller-manager",
			Namespace:        "etcdadm-bootstrap-provider-system",
			ExpectedReplicas: 1,
		},
	}
	return a.generateDeploymentAnalyzers(deployments)
}

func (a *analyzerFactory) DataCenterConfigAnalyzers(datacenter v1alpha1.Ref) []*v1beta2.Analyze {
	switch datacenter.Kind {
	case v1alpha1.VSphereDatacenterKind:
		return a.eksaVsphereAnalyzers()
	case v1alpha1.DockerDatacenterKind:
		return a.eksaDockerAnalyzers()
	default:
		return nil
	}
}

func (a *analyzerFactory) eksaVsphereAnalyzers() []*v1beta2.Analyze {
	crds := []string{
		fmt.Sprintf("vspheredatacenterconfigs.%s", v1alpha1.GroupVersion.Group),
		fmt.Sprintf("vspheremachineconfigs.%s", v1alpha1.GroupVersion.Group),
	}
	return a.generateCrdAnalyzers(crds)
}

func (a *analyzerFactory) eksaDockerAnalyzers() []*v1beta2.Analyze {
	var analyazers []*v1beta2.Analyze

	crds := []string{
		fmt.Sprintf("dockerdatacenterconfigs.%s", v1alpha1.GroupVersion.Group),
	}

	deployments := []eksaDeployment{
		{
			Name:             "local-path-provisioner",
			Namespace:        "local-path-storage",
			ExpectedReplicas: 1,
		},
	}

	analyazers = append(analyazers, a.generateCrdAnalyzers(crds)...)
	return append(analyazers, a.generateDeploymentAnalyzers(deployments)...)
}

type eksaDeployment struct {
	Name             string
	Namespace        string
	ExpectedReplicas int
}

func (a *analyzerFactory) generateDeploymentAnalyzers(deployments []eksaDeployment) []*v1beta2.Analyze {
	var deploymentAnalyzers []*v1beta2.Analyze
	for _, d := range deployments {
		deploymentAnalyzers = append(deploymentAnalyzers, a.deploymentAnalyzer(d))
	}
	return deploymentAnalyzers
}

func (a *analyzerFactory) deploymentAnalyzer(deployment eksaDeployment) *v1beta2.Analyze {
	return &v1beta2.Analyze{
		DeploymentStatus: &v1beta2.DeploymentStatus{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			Outcomes: []*v1beta2.Outcome{
				{
					Fail: &v1beta2.SingleOutcome{
						When:    fmt.Sprintf("< %d", deployment.ExpectedReplicas),
						Message: fmt.Sprintf("%s is not ready.", deployment.Name),
					},
				}, {
					Pass: &v1beta2.SingleOutcome{
						Message: fmt.Sprintf("%s is running.", deployment.Name),
					},
				},
			},
		},
	}
}

func (a *analyzerFactory) generateCrdAnalyzers(crds []string) []*v1beta2.Analyze {
	var crdAnalyzers []*v1beta2.Analyze
	for _, crd := range crds {
		crdAnalyzers = append(crdAnalyzers, a.crdAnalyzer(crd))
	}
	return crdAnalyzers
}

func (a *analyzerFactory) crdAnalyzer(crdName string) *v1beta2.Analyze {
	return &v1beta2.Analyze{
		CustomResourceDefinition: &v1beta2.CustomResourceDefinition{
			AnalyzeMeta: v1beta2.AnalyzeMeta{
				CheckName: crdName,
			},
			Outcomes: []*v1beta2.Outcome{
				{
					Fail: &v1beta2.SingleOutcome{
						When:    "< 1",
						Message: fmt.Sprintf("%s is not present on cluster", crdName),
					},
				},
				{
					Pass: &v1beta2.SingleOutcome{
						Message: fmt.Sprintf("%s is present on the cluster", crdName),
					},
				},
			},
			CustomResourceDefinitionName: crdName,
		},
	}
}

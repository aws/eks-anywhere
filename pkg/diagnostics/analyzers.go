package diagnostics

import (
	"fmt"
	"path"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	logAnalysisAnalyzerPrefix = "log analysis:"
)

type analyzerFactory struct{}

func NewAnalyzerFactory() *analyzerFactory {
	return &analyzerFactory{}
}

func (a *analyzerFactory) DefaultAnalyzers() []*Analyze {
	var analyzers []*Analyze
	return append(analyzers, a.defaultDeploymentAnalyzers()...)
}

func (a *analyzerFactory) defaultDeploymentAnalyzers() []*Analyze {
	d := []eksaDeployment{
		{
			Name:             "coredns",
			Namespace:        constants.KubeSystemNamespace,
			ExpectedReplicas: 2,
		},
	}
	return a.generateDeploymentAnalyzers(d)
}

func (a *analyzerFactory) ManagementClusterAnalyzers() []*Analyze {
	var analyzers []*Analyze
	analyzers = append(analyzers, a.managementClusterDeploymentAnalyzers()...)
	return append(analyzers, a.managementClusterCrdAnalyzers()...)
}

func (a *analyzerFactory) managementClusterCrdAnalyzers() []*Analyze {
	crds := []string{
		fmt.Sprintf("clusters.%s", v1alpha1.GroupVersion.Group),
		fmt.Sprintf("bundles.%s", v1alpha1.GroupVersion.Group),
	}
	return a.generateCrdAnalyzers(crds)
}

func (a *analyzerFactory) PackageAnalyzers() []*Analyze {
	var analyzers []*Analyze
	analyzers = append(analyzers, a.packageDeploymentAnalyzers()...)
	return append(analyzers, a.packageCrdAnalyzers()...)
}

func (a *analyzerFactory) packageCrdAnalyzers() []*Analyze {
	crds := []string{
		"packagebundlecontrollers.packages.eks.amazonaws.com",
		"packagebundles.packages.eks.amazonaws.com",
		"packagecontrollers.packages.eks.amazonaws.com",
		"packages.packages.eks.amazonaws.com",
	}
	return a.generateCrdAnalyzers(crds)
}

func (a *analyzerFactory) packageDeploymentAnalyzers() []*Analyze {
	d := []eksaDeployment{
		{
			Name:             "eks-anywhere-packages",
			Namespace:        constants.EksaPackagesName,
			ExpectedReplicas: 1,
		},
	}
	return a.generateDeploymentAnalyzers(d)
}

func (a *analyzerFactory) managementClusterDeploymentAnalyzers() []*Analyze {
	d := []eksaDeployment{
		{
			Name:             "capt-controller-manager",
			Namespace:        constants.CaptSystemNamespace,
			ExpectedReplicas: 1,
		}, {
			Name:             "capv-controller-manager",
			Namespace:        constants.CapvSystemNamespace,
			ExpectedReplicas: 1,
		}, {
			Name:             "capc-controller-manager",
			Namespace:        constants.CapcSystemNamespace,
			ExpectedReplicas: 1,
		}, {
			Name:             "capx-controller-manager",
			Namespace:        constants.CapxSystemNamespace,
			ExpectedReplicas: 1,
		}, {
			Name:             "cert-manager-webhook",
			Namespace:        constants.CertManagerNamespace,
			ExpectedReplicas: 1,
		}, {
			Name:             "cert-manager-cainjector",
			Namespace:        constants.CertManagerNamespace,
			ExpectedReplicas: 1,
		}, {
			Name:             "cert-manager",
			Namespace:        constants.CertManagerNamespace,
			ExpectedReplicas: 1,
		}, {
			Name:             "capi-controller-manager",
			Namespace:        constants.CapiSystemNamespace,
			ExpectedReplicas: 1,
		}, {
			Name:             "capi-kubeadm-control-plane-controller-manager",
			Namespace:        constants.CapiKubeadmControlPlaneSystemNamespace,
			ExpectedReplicas: 1,
		}, {
			Name:             "capi-kubeadm-control-plane-controller-manager",
			Namespace:        constants.CapiKubeadmControlPlaneSystemNamespace,
			ExpectedReplicas: 1,
		}, {
			Name:             "capi-kubeadm-bootstrap-controller-manager",
			Namespace:        constants.CapiKubeadmBootstrapSystemNamespace,
			ExpectedReplicas: 1,
		},
	}
	return a.generateDeploymentAnalyzers(d)
}

func (a *analyzerFactory) EksaGitopsAnalyzers() []*Analyze {
	crds := []string{
		fmt.Sprintf("gitopsconfigs.%s", v1alpha1.GroupVersion.Group),
	}
	return a.generateCrdAnalyzers(crds)
}

func (a *analyzerFactory) EksaOidcAnalyzers() []*Analyze {
	crds := []string{
		fmt.Sprintf("oidcconfigs.%s", v1alpha1.GroupVersion.Group),
	}
	return a.generateCrdAnalyzers(crds)
}

func (a *analyzerFactory) EksaExternalEtcdAnalyzers() []*Analyze {
	deployments := []eksaDeployment{
		{
			Name:             "etcdadm-controller-controller-manager",
			Namespace:        constants.EtcdAdmControllerSystemNamespace,
			ExpectedReplicas: 1,
		}, {
			Name:             "etcdadm-bootstrap-provider-controller-manager",
			Namespace:        constants.EtcdAdmBootstrapProviderSystemNamespace,
			ExpectedReplicas: 1,
		},
	}
	return a.generateDeploymentAnalyzers(deployments)
}

func (a *analyzerFactory) DataCenterConfigAnalyzers(datacenter v1alpha1.Ref) []*Analyze {
	switch datacenter.Kind {
	case v1alpha1.VSphereDatacenterKind:
		return a.eksaVsphereAnalyzers()
	case v1alpha1.DockerDatacenterKind:
		return a.eksaDockerAnalyzers()
	case v1alpha1.CloudStackDatacenterKind:
		return a.eksaCloudstackAnalyzers()
	case v1alpha1.SnowDatacenterKind:
		return a.eksaSnowAnalyzers()
	case v1alpha1.NutanixDatacenterKind:
		return a.eksaNutanixAnalyzers()
	default:
		return nil
	}
}

func (a *analyzerFactory) eksaVsphereAnalyzers() []*Analyze {
	var analyzers []*Analyze
	crds := []string{
		fmt.Sprintf("vspheredatacenterconfigs.%s", v1alpha1.GroupVersion.Group),
		fmt.Sprintf("vspheremachineconfigs.%s", v1alpha1.GroupVersion.Group),
	}
	analyzers = append(analyzers, a.generateCrdAnalyzers(crds)...)
	analyzers = append(analyzers, a.vsphereDiagnosticAnalyzers()...)
	return analyzers
}

func (a *analyzerFactory) eksaCloudstackAnalyzers() []*Analyze {
	crds := []string{
		fmt.Sprintf("cloudstackdatacenterconfigs.%s", v1alpha1.GroupVersion.Group),
		fmt.Sprintf("cloudstackmachineconfigs.%s", v1alpha1.GroupVersion.Group),
	}
	analyzers := a.generateCrdAnalyzers(crds)
	return append(analyzers, a.validControlPlaneIPAnalyzer())
}

func (a *analyzerFactory) eksaSnowAnalyzers() []*Analyze {
	crds := []string{
		fmt.Sprintf("snowdatacenterconfigs.%s", v1alpha1.GroupVersion.Group),
		fmt.Sprintf("snowmachineconfigs.%s", v1alpha1.GroupVersion.Group),
	}
	analyzers := a.generateCrdAnalyzers(crds)
	return append(analyzers, a.validControlPlaneIPAnalyzer())
}

func (a *analyzerFactory) eksaDockerAnalyzers() []*Analyze {
	var analyazers []*Analyze

	crds := []string{
		fmt.Sprintf("dockerdatacenterconfigs.%s", v1alpha1.GroupVersion.Group),
	}

	deployments := []eksaDeployment{
		{
			Name:             "local-path-provisioner",
			Namespace:        constants.LocalPathStorageNamespace,
			ExpectedReplicas: 1,
		},
	}

	analyazers = append(analyazers, a.generateCrdAnalyzers(crds)...)
	return append(analyazers, a.generateDeploymentAnalyzers(deployments)...)
}

func (a *analyzerFactory) eksaNutanixAnalyzers() []*Analyze {
	crds := []string{
		fmt.Sprintf("nutanixdatacenterconfigs.%s", v1alpha1.GroupVersion.Group),
		fmt.Sprintf("nutanixmachineconfigs.%s", v1alpha1.GroupVersion.Group),
	}
	analyzers := a.generateCrdAnalyzers(crds)
	return append(analyzers, a.validControlPlaneIPAnalyzer())
}

// EksaLogTextAnalyzers given a slice of Collectors will check which namespaced log collectors are present
// and return the log analyzers associated with the namespace in the namespaceLogTextAnalyzersMap.
func (a *analyzerFactory) EksaLogTextAnalyzers(collectors []*Collect) []*Analyze {
	var analyzers []*Analyze
	analyzersMap := a.namespaceLogTextAnalyzersMap()
	for _, collector := range collectors {
		if collector.Logs != nil {
			analyzer, ok := analyzersMap[collector.Logs.Namespace]
			if ok {
				analyzers = append(analyzers, analyzer...)
			}
		}
	}
	return analyzers
}

// namespaceLogTextAnalyzersMap is used to associated log text analyzers with the logs collected from a specific namespace.
// the key of the analyzers map is the namespace name, and the value are the associated log text analyzers.
func (a *analyzerFactory) namespaceLogTextAnalyzersMap() map[string][]*Analyze {
	return map[string][]*Analyze{
		constants.CapiKubeadmControlPlaneSystemNamespace: a.capiKubeadmControlPlaneSystemLogAnalyzers(),
	}
}

func (a *analyzerFactory) capiKubeadmControlPlaneSystemLogAnalyzers() []*Analyze {
	capiCpManagerPod := "capi-kubeadm-control-plane-controller-manager-*"
	capiCpManagerContainerLogFile := capiCpManagerPod + ".log"
	fullManagerPodLogPath := path.Join(logpath(constants.CapiKubeadmControlPlaneSystemNamespace), capiCpManagerContainerLogFile)
	return []*Analyze{
		{
			TextAnalyze: &textAnalyze{
				analyzeMeta: analyzeMeta{
					CheckName: fmt.Sprintf("%s: API server pod missing. Log: %s", logAnalysisAnalyzerPrefix, fullManagerPodLogPath),
				},
				FileName:     fullManagerPodLogPath,
				RegexPattern: `machine (.*?) reports APIServerPodHealthy condition is false \(Error, Pod kube-apiserver-(.*?) is missing\)`,
				Outcomes: []*outcome{
					{
						Fail: &singleOutcome{
							When:    "true",
							Message: fmt.Sprintf("Node failed to launch correctly; API server pod is missing. See %s", fullManagerPodLogPath),
						},
					},
					{
						Pass: &singleOutcome{
							When:    "false",
							Message: "API server pods launched correctly",
						},
					},
				},
			},
		},
	}
}

type eksaDeployment struct {
	Name             string
	Namespace        string
	ExpectedReplicas int
}

func (a *analyzerFactory) generateDeploymentAnalyzers(deployments []eksaDeployment) []*Analyze {
	var deploymentAnalyzers []*Analyze
	for _, d := range deployments {
		deploymentAnalyzers = append(deploymentAnalyzers, a.deploymentAnalyzer(d))
	}
	return deploymentAnalyzers
}

func (a *analyzerFactory) deploymentAnalyzer(deployment eksaDeployment) *Analyze {
	return &Analyze{
		DeploymentStatus: &deploymentStatus{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			Outcomes: []*outcome{
				{
					Fail: &singleOutcome{
						When:    fmt.Sprintf("< %d", deployment.ExpectedReplicas),
						Message: fmt.Sprintf("%s is not ready.", deployment.Name),
					},
				}, {
					Pass: &singleOutcome{
						Message: fmt.Sprintf("%s is running.", deployment.Name),
					},
				},
			},
		},
	}
}

func (a *analyzerFactory) generateCrdAnalyzers(crds []string) []*Analyze {
	var crdAnalyzers []*Analyze
	for _, crd := range crds {
		crdAnalyzers = append(crdAnalyzers, a.crdAnalyzer(crd))
	}
	return crdAnalyzers
}

func (a *analyzerFactory) crdAnalyzer(crdName string) *Analyze {
	return &Analyze{
		CustomResourceDefinition: &customResourceDefinition{
			analyzeMeta: analyzeMeta{
				CheckName: crdName,
			},
			Outcomes: []*outcome{
				{
					Fail: &singleOutcome{
						When:    "< 1",
						Message: fmt.Sprintf("%s is not present on cluster", crdName),
					},
				},
				{
					Pass: &singleOutcome{
						Message: fmt.Sprintf("%s is present on the cluster", crdName),
					},
				},
			},
			CustomResourceDefinitionName: crdName,
		},
	}
}

// vsphereDiagnosticAnalyzers will return diagnostic analyzers to analyze the condition of vSphere cluster.
func (a *analyzerFactory) vsphereDiagnosticAnalyzers() []*Analyze {
	return []*Analyze{a.validControlPlaneIPAnalyzer(), a.vcenterSessionValidatePermissionAnalyzer()}
}

// validControlPlaneIPAnalyzer analyzes whether a valid control plane IP is used to connect
// to API server.
func (a *analyzerFactory) validControlPlaneIPAnalyzer() *Analyze {
	runPingPod := "ping-host-ip"
	runPingPodLog := fmt.Sprintf("%s.log", runPingPod)
	fullRunPingPodLogPath := path.Join(runPingPod, runPingPodLog)
	return &Analyze{
		TextAnalyze: &textAnalyze{
			analyzeMeta: analyzeMeta{
				CheckName: fmt.Sprintf("%s:  Destination Host Unreachable. Log: %s", logAnalysisAnalyzerPrefix, fullRunPingPodLogPath),
			},
			FileName:     fullRunPingPodLogPath,
			RegexPattern: `exit code: 0`,
			Outcomes: []*outcome{
				{
					Fail: &singleOutcome{
						When:    "false",
						Message: fmt.Sprintf("The control plane endpoint host is unavailable. See %s", fullRunPingPodLogPath),
					},
				},

				{
					Pass: &singleOutcome{
						When:    "true",
						Message: "Control plane IP verified.",
					},
				},
			},
		},
	}
}

// vcenterSessionValidateAnalyzer analyzes whether the vcenter user has Session validate permissions for CAPV
// to be able to look up existing valid sessions to reuse them instead of having to create new ones.
func (a *analyzerFactory) vcenterSessionValidatePermissionAnalyzer() *Analyze {
	capvManagerPod := "capv-controller-manager-*"
	capvManagerContainerLogFile := capvManagerPod + ".log"
	fullManagerPodLogPath := path.Join(logpath(constants.CapvSystemNamespace), capvManagerContainerLogFile)
	return &Analyze{
		TextAnalyze: &textAnalyze{
			analyzeMeta: analyzeMeta{
				CheckName: fmt.Sprintf("%s:  Session Validate permission missing. Log: %s", logAnalysisAnalyzerPrefix, fullManagerPodLogPath),
			},
			FileName:     fullManagerPodLogPath,
			RegexPattern: `session "msg"="error checking if session is active" "error"="ServerFaultCode: Permission to perform this operation was denied."`,
			Outcomes: []*outcome{
				{
					Fail: &singleOutcome{
						When:    "true",
						Message: fmt.Sprintf("VCenter user doesn't have Sessions.ValidateSession permission. See %s", fullManagerPodLogPath),
					},
				},

				{
					Pass: &singleOutcome{
						When:    "false",
						Message: "VCenter user has Sessions.ValidateSession permission.",
					},
				},
			},
		},
	}
}

// vmsAccessAnalyzer will analyze if vms have access to the API server of vSphere cluster
// not used yet but it will once the workflows are updated to support this usecase.
func (a *analyzerFactory) vmsAccessAnalyzer() *Analyze { //nolint:unused
	runBashPod := "check-cloud-controller"
	runBashPodLog := fmt.Sprintf("%s.log", runBashPod)
	vSphereCloudControllerPodLogPath := path.Join(runBashPod, runBashPodLog)
	return &Analyze{
		TextAnalyze: &textAnalyze{
			analyzeMeta: analyzeMeta{
				CheckName: fmt.Sprintf("%s: Virtual Machine has no access to vSphere API server. Logs: %s", logAnalysisAnalyzerPrefix, vSphereCloudControllerPodLogPath),
			},
			FileName:     vSphereCloudControllerPodLogPath,
			RegexPattern: `Failed to create new client. err: Post (.*) dial tcp (.*) connect: connection timed out\n(.*)Failed to create govmomi client. err: Post (.*) dial tcp (.*) connect: connection timed out`,
			Outcomes: []*outcome{
				{
					Fail: &singleOutcome{
						When:    "true",
						Message: fmt.Sprintf("Failed to create client, Virtural Machines have no access to vSphere API server. See the cloud controller log in control plane node: %s", vSphereCloudControllerPodLogPath),
					},
				},
				{
					Pass: &singleOutcome{
						When:    "false",
						Message: fmt.Sprintf("Virtual Machines have access to vSphere API server. See %s \nPlease ignore the result when this analyzer is running on bootstrap cluster", vSphereCloudControllerPodLogPath),
					},
				},
			},
		},
	}
}

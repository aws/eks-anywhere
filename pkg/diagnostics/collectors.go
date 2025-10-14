package diagnostics

import (
	"fmt"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers"
)

// FileReader reads files from local disk or http urls.
type FileReader interface {
	ReadFile(url string) ([]byte, error)
}

// EKSACollectorFactory generates support-bundle collectors for eks-a clusters.
type EKSACollectorFactory struct {
	DiagnosticCollectorImage string
	reader                   FileReader
}

// NewCollectorFactory builds a collector factory.
func NewCollectorFactory(diagnosticCollectorImage string, reader FileReader) *EKSACollectorFactory {
	return &EKSACollectorFactory{
		DiagnosticCollectorImage: diagnosticCollectorImage,
		reader:                   reader,
	}
}

// NewDefaultCollectorFactory builds a collector factory that will use the default
// diagnostic collector image.
func NewDefaultCollectorFactory(reader FileReader) *EKSACollectorFactory {
	return NewCollectorFactory("", reader)
}

// DefaultCollectors returns the collectors that apply to all clusters.
func (c *EKSACollectorFactory) DefaultCollectors() []*Collect {
	collectors := []*Collect{
		{
			ClusterInfo: &clusterInfo{},
		},
		{
			ClusterResources: &clusterResources{},
		},
		{
			Secret: &secret{
				Namespace:    "eksa-system",
				SecretName:   "eksa-license",
				IncludeValue: true,
				Key:          "license",
			},
		},
	}
	collectors = append(collectors, c.defaultLogCollectors()...)
	collectors = append(collectors, c.webhookConfigCollectors()...)
	return collectors
}

// EksaHostCollectors returns the collectors that interact with the kubernetes node machine hosts.
func (c *EKSACollectorFactory) EksaHostCollectors(machineConfigs []providers.MachineConfig) []*Collect {
	var collectors []*Collect
	collectorsMap := c.getCollectorsMap()

	// we don't want to duplicate the collectors if multiple machine configs have the same OS family
	osFamiliesSeen := map[v1alpha1.OSFamily]bool{}
	for _, config := range machineConfigs {
		if _, seen := osFamiliesSeen[config.OSFamily()]; !seen {
			collectors = append(collectors, collectorsMap[config.OSFamily()]...)
			osFamiliesSeen[config.OSFamily()] = true
		}
	}
	return collectors
}

// HostCollectors returns the collectors that run on host machines.
func (c *EKSACollectorFactory) HostCollectors(datacenter v1alpha1.Ref) []*Collect {
	// Only Tinkerbell needs this right now to collect boots/smee logs in docker container.
	switch datacenter.Kind {
	case v1alpha1.TinkerbellDatacenterKind:
		return c.hostTinkerbellCollectors()
	default:
		return nil
	}
}

// AuditLogCollectors returns the audit log collectors that run on all control plane nodes.
func (c *EKSACollectorFactory) AuditLogCollectors() []*Collect {
	return []*Collect{
		{
			RunDaemonSet: &RunDaemonSet{
				Name:      "audit-logs",
				Namespace: constants.EksaDiagnosticsNamespace,
				PodSpec: &v1.PodSpec{
					Containers: []v1.Container{{
						Name:    "audit-logs-collector",
						Image:   c.DiagnosticCollectorImage,
						Command: []string{"/bin/sh", "-c"},
						Args: []string{
							"cat /hostlogs/kubernetes/api-audit.log",
						},
						VolumeMounts: []v1.VolumeMount{{
							Name:      "host-var-log",
							MountPath: "/hostlogs",
							ReadOnly:  true,
						}},
					}},
					Volumes: []v1.Volume{{
						Name: "host-var-log",
						VolumeSource: v1.VolumeSource{
							HostPath: &v1.HostPathVolumeSource{
								Path: "/var/log",
							},
						},
					}},
					NodeSelector: map[string]string{
						"node-role.kubernetes.io/control-plane": "",
					},
					Tolerations: []v1.Toleration{{
						Key:      "node-role.kubernetes.io/control-plane",
						Operator: "Exists",
						Effect:   "NoSchedule",
					}},
				},
				Timeout: "60s",
			},
		},
	}
}

// DataCenterConfigCollectors returns the collectors for the provider datacenter config in the cluster spec.
func (c *EKSACollectorFactory) DataCenterConfigCollectors(datacenter v1alpha1.Ref, spec *cluster.Spec) []*Collect {
	switch datacenter.Kind {
	case v1alpha1.VSphereDatacenterKind:
		return c.eksaVsphereCollectors(spec)
	case v1alpha1.DockerDatacenterKind:
		return c.eksaDockerCollectors()
	case v1alpha1.CloudStackDatacenterKind:
		return c.eksaCloudstackCollectors()
	case v1alpha1.TinkerbellDatacenterKind:
		return c.eksaTinkerbellCollectors()
	case v1alpha1.SnowDatacenterKind:
		return c.eksaSnowCollectors()
	case v1alpha1.NutanixDatacenterKind:
		return c.eksaNutanixCollectors()
	default:
		return nil
	}
}

func (c *EKSACollectorFactory) eksaNutanixCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.CapxSystemNamespace,
				Name:      logpath(constants.CapxSystemNamespace),
			},
		},
	}
}

func (c *EKSACollectorFactory) eksaSnowCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.CapasSystemNamespace,
				Name:      logpath(constants.CapasSystemNamespace),
			},
		},
	}
}

func (c *EKSACollectorFactory) eksaTinkerbellCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.CaptSystemNamespace,
				Name:      logpath(constants.CaptSystemNamespace),
			},
		},
	}
}

func (c *EKSACollectorFactory) eksaVsphereCollectors(spec *cluster.Spec) []*Collect {
	var collectors []*Collect
	vsphereLogs := []*Collect{
		{
			Logs: &logs{
				Namespace: constants.CapvSystemNamespace,
				Name:      logpath(constants.CapvSystemNamespace),
			},
		},
	}
	collectors = append(collectors, vsphereLogs...)
	collectors = append(collectors, c.apiServerCollectors(spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)...)
	collectors = append(collectors, c.vmsAccessCollector(spec.Cluster.Spec.ControlPlaneConfiguration))
	return collectors
}

func (c *EKSACollectorFactory) eksaCloudstackCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.CapcSystemNamespace,
				Name:      logpath(constants.CapcSystemNamespace),
			},
		},
	}
}

func (c *EKSACollectorFactory) eksaDockerCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.CapdSystemNamespace,
				Name:      logpath(constants.CapdSystemNamespace),
			},
		},
	}
}

func (c *EKSACollectorFactory) hostTinkerbellCollectors() []*Collect {
	collectors := []*Collect{
		{
			Run: &Run{
				CollectorName: "boots-logs",
				Command:       "docker",
				Args:          []string{"logs", "boots"},
				OutputDir:     "boots-logs",
			},
		},
	}

	return collectors
}

// ManagementClusterCollectors returns the collectors that only apply to management clusters.
func (c *EKSACollectorFactory) ManagementClusterCollectors() []*Collect {
	return c.managementClusterLogCollectors()
}

// PackagesCollectors returns the collectors that read information for curated packages.
func (c *EKSACollectorFactory) PackagesCollectors() []*Collect {
	return c.packagesLogCollectors()
}

// FileCollectors returns the collectors that interact with files.
func (c *EKSACollectorFactory) FileCollectors(paths []string) []*Collect {
	collectors := []*Collect{}

	for _, path := range paths {
		content, err := c.reader.ReadFile(path)
		if err != nil {
			content = []byte(fmt.Sprintf("Failed to retrieve file %s for collection: %s", path, err))
		}

		collectors = append(collectors, &Collect{
			Data: &data{
				Data: string(content),
				Name: filepath.Base(path),
			},
		})
	}

	return collectors
}

func (c *EKSACollectorFactory) getCollectorsMap() map[v1alpha1.OSFamily][]*Collect {
	return map[v1alpha1.OSFamily][]*Collect{
		v1alpha1.Ubuntu:       c.ubuntuHostCollectors(),
		v1alpha1.Bottlerocket: c.bottleRocketHostCollectors(),
	}
}

func (c *EKSACollectorFactory) bottleRocketHostCollectors() []*Collect {
	return []*Collect{}
}

func (c *EKSACollectorFactory) ubuntuHostCollectors() []*Collect {
	return []*Collect{
		{
			CopyFromHost: &copyFromHost{
				Name:      hostlogPath("cloud-init"),
				Namespace: constants.EksaDiagnosticsNamespace,
				Image:     c.DiagnosticCollectorImage,
				HostPath:  "/var/log/cloud-init.log",
			},
		},
		{
			CopyFromHost: &copyFromHost{
				Name:      hostlogPath("cloud-init-output"),
				Namespace: constants.EksaDiagnosticsNamespace,
				Image:     c.DiagnosticCollectorImage,
				HostPath:  "/var/log/cloud-init-output.log",
			},
		},
		{
			CopyFromHost: &copyFromHost{
				Name:      hostlogPath("syslog"),
				Namespace: constants.EksaDiagnosticsNamespace,
				Image:     c.DiagnosticCollectorImage,
				HostPath:  "/var/log/syslog",
				Timeout:   time.Minute.String(),
			},
		},
	}
}

func (c *EKSACollectorFactory) defaultLogCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.EksaSystemNamespace,
				Name:      logpath(constants.EksaSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.DefaultNamespace,
				Name:      logpath(constants.DefaultNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.KubeNodeLeaseNamespace,
				Name:      logpath(constants.KubeNodeLeaseNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.KubePublicNamespace,
				Name:      logpath(constants.KubePublicNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.KubeSystemNamespace,
				Name:      logpath(constants.KubeSystemNamespace),
			},
		},
	}
}

func (c *EKSACollectorFactory) packagesLogCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.EksaPackagesName,
				Name:      logpath(constants.EksaPackagesName),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.AdotPrometheusNamespace,
				Name:      logpath(constants.AdotPrometheusNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.MetallbNamespace,
				Name:      logpath(constants.MetallbNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.EmissaryNamespace,
				Name:      logpath(constants.EmissaryNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.CertManagerNamespace,
				Name:      logpath(constants.CertManagerNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.HarborNamespace,
				Name:      logpath(constants.HarborNamespace),
			},
		},
	}
}

func (c *EKSACollectorFactory) managementClusterLogCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.CapiKubeadmBootstrapSystemNamespace,
				Name:      logpath(constants.CapiKubeadmBootstrapSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.CapiKubeadmControlPlaneSystemNamespace,
				Name:      logpath(constants.CapiKubeadmControlPlaneSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.CapiSystemNamespace,
				Name:      logpath(constants.CapiSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.CapiWebhookSystemNamespace,
				Name:      logpath(constants.CapiWebhookSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.CertManagerNamespace,
				Name:      logpath(constants.CertManagerNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.EtcdAdmBootstrapProviderSystemNamespace,
				Name:      logpath(constants.EtcdAdmBootstrapProviderSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.EtcdAdmControllerSystemNamespace,
				Name:      logpath(constants.EtcdAdmControllerSystemNamespace),
			},
		},
	}
}

// apiServerCollectors collect connection info when running a pod on an existing cluster.
func (c *EKSACollectorFactory) apiServerCollectors(controlPlaneIP string) []*Collect {
	var collectors []*Collect
	collectors = append(collectors, c.controlPlaneNetworkPathCollector(controlPlaneIP)...)
	return collectors
}

func (c *EKSACollectorFactory) controlPlaneNetworkPathCollector(controlPlaneIP string) []*Collect {
	ports := []string{"6443", "22"}
	var collectors []*Collect
	collectors = append(collectors, c.hostPortCollector(ports, controlPlaneIP))
	collectors = append(collectors, c.pingHostCollector(controlPlaneIP))
	return collectors
}

func (c *EKSACollectorFactory) hostPortCollector(ports []string, hostIP string) *Collect {
	apiServerPort := ports[0]
	port := ports[1]
	tempIPRequest := fmt.Sprintf("for port in %s %s; do nc -z -v -w5 %s $port; done", apiServerPort, port, hostIP)
	argsIP := []string{tempIPRequest}
	return &Collect{
		RunPod: &runPod{
			Name:      "check-host-port",
			Namespace: constants.EksaDiagnosticsNamespace,
			PodSpec: &v1.PodSpec{
				Containers: []v1.Container{{
					Name:    "check-host-port",
					Image:   c.DiagnosticCollectorImage,
					Command: []string{"/bin/sh", "-c"},
					Args:    argsIP,
				}},
			},
			Timeout: "30s",
		},
	}
}

func (c *EKSACollectorFactory) pingHostCollector(hostIP string) *Collect {
	tempPingRequest := fmt.Sprintf("ping -w10 -c5 %s; echo exit code: $?", hostIP)
	argsPing := []string{tempPingRequest}
	return &Collect{
		RunPod: &runPod{
			Name:      "ping-host-ip",
			Namespace: constants.EksaDiagnosticsNamespace,
			PodSpec: &v1.PodSpec{
				Containers: []v1.Container{{
					Name:    "ping-host-ip",
					Image:   c.DiagnosticCollectorImage,
					Command: []string{"/bin/sh", "-c"},
					Args:    argsPing,
				}},
			},
			Timeout: "30s",
		},
	}
}

// vmsAccessCollector will connect to API server first, then collect vsphere-cloud-controller-manager logs
// on control plane node.
func (c *EKSACollectorFactory) vmsAccessCollector(controlPlaneConfiguration v1alpha1.ControlPlaneConfiguration) *Collect {
	controlPlaneEndpointHost := controlPlaneConfiguration.Endpoint.Host
	taints := controlPlaneConfiguration.Taints
	tolerations := makeTolerations(taints)

	makeConnection := fmt.Sprintf("curl --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt -H \"Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)\" https://%s:6443/", controlPlaneEndpointHost)
	getLogs := "kubectl logs -n kube-system -l k8s-app=vsphere-cloud-controller-manager"
	args := []string{fmt.Sprintf("%s && %s", makeConnection, getLogs)}
	return &Collect{
		RunPod: &runPod{
			Name:      "check-cloud-controller",
			Namespace: constants.EksaDiagnosticsNamespace,
			PodSpec: &v1.PodSpec{
				Containers: []v1.Container{{
					Name:    "check-cloud-controller",
					Image:   c.DiagnosticCollectorImage,
					Command: []string{"/bin/sh", "-c"},
					Args:    args,
				}},
				ServiceAccountName: "default",
				HostNetwork:        true,
				Tolerations:        tolerations,
				Affinity: &v1.Affinity{
					NodeAffinity: &v1.NodeAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
							{
								Weight: 10,
								Preference: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{{
										Key:      "node-role.kubernetes.io/control-plane",
										Operator: "Exists",
									}},
								},
							}, {
								Weight: 10,
								Preference: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{{
										Key:      "node-role.kubernetes.io/master",
										Operator: "Exists",
									}},
								},
							},
						},
					},
				},
			},
			Timeout: "20s",
		},
	}
}

func makeTolerations(taints []v1.Taint) []v1.Toleration {
	tolerations := []v1.Toleration{
		{
			Key:    "node.cloudprovider.kubernetes.io/uninitialized",
			Value:  "true",
			Effect: "NoSchedule",
		},
		{
			Key:    "node.kubernetes.io/not-ready",
			Effect: "NoSchedule",
		},
	}
	if taints == nil {
		toleration := v1.Toleration{
			Effect: "NoSchedule",
			Key:    "node-role.kubernetes.io/master",
		}
		tolerations = append(tolerations, toleration)
	} else {
		for _, taint := range taints {
			var toleration v1.Toleration
			if taint.Key != "" {
				toleration.Key = taint.Key
			}
			if taint.Value != "" {
				toleration.Value = taint.Value
			}
			if taint.Effect != "" {
				toleration.Effect = taint.Effect
			}
			tolerations = append(tolerations, toleration)
		}
	}
	return tolerations
}

func logpath(namespace string) string {
	return fmt.Sprintf("logs/%s", namespace)
}

func hostlogPath(logType string) string {
	return fmt.Sprintf("hostLogs/%s", logType)
}

// webhookConfigCollectors returns collectors for admission webhook configurations.
func (c *EKSACollectorFactory) webhookConfigCollectors() []*Collect {
	return []*Collect{
		c.webhookConfigCollector("validatingwebhookconfigurations"),
		c.webhookConfigCollector("mutatingwebhookconfigurations"),
	}
}

func (c *EKSACollectorFactory) webhookConfigCollector(resourceType string) *Collect {
	command := []string{"kubectl"}
	args := []string{"get", resourceType, "-o", "json"}
	return &Collect{
		RunPod: &runPod{
			collectorMeta: collectorMeta{
				CollectorName: resourceType,
			},
			Name:      resourceType,
			Namespace: constants.EksaDiagnosticsNamespace,
			PodSpec: &v1.PodSpec{
				Containers: []v1.Container{{
					Name:    resourceType,
					Image:   c.DiagnosticCollectorImage,
					Command: command,
					Args:    args,
				}},
				// It's possible for networking to not be working on the cluster or the nodes
				// not being ready, so adding tolerations and running the pod on host networking
				// to be able to pull the resources from the cluster
				HostNetwork: true,
				Tolerations: []v1.Toleration{{
					Key:    "node.kubernetes.io",
					Value:  "not-ready",
					Effect: "NoSchedule",
				}},
			},
			Timeout: "30s",
		},
	}
}

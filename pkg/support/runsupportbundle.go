package supportbundle

import (
	_ "embed"
	"fmt"
	"time"

	analyzerunner "github.com/replicatedhq/troubleshoot/pkg/analyze"
	"github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	"github.com/replicatedhq/troubleshoot/pkg/k8sutil"
	"github.com/replicatedhq/troubleshoot/pkg/supportbundle"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const resultsSeparator = "\n------------\n"

func ParseBundleFromDoc(bundleConfig string) (*v1beta2.SupportBundle, error) {
	if bundleConfig == "" {
		// user did not provide any bundle-config to the support-bundle command, generate one using the default collectors & analyzers
		return newBundleConfig(), nil
	}

	// parse bundle-config provided by the user
	collectorContent, err := supportbundle.LoadSupportBundleSpec(bundleConfig)
	if err != nil {
		return nil, err
	}
	return supportbundle.ParseSupportBundleFromDoc(collectorContent)
}

func CollectBundleFromSpec(sinceTimeValue *time.Time, spec *v1beta2.SupportBundleSpec) (string, error) {
	k8sConfig, err := k8sutil.GetRESTConfig()
	if err != nil {
		return "", fmt.Errorf("failed to convert kube flags to rest config: %v", err)
	}

	progressChan := make(chan interface{})
	go func() {
		var lastMsg string
		for {
			msg := <-progressChan
			switch msg := msg.(type) {
			case error:
				logger.Info(fmt.Sprintf("\r * %v", msg))
			case string:
				if lastMsg != msg {
					logger.Info(fmt.Sprintf("\r \033[36mCollecting support bundle\033[m %v", msg))
					lastMsg = msg
				}
			}
		}
	}()

	collectorCB := func(c chan interface{}, msg string) {
		c <- msg
	}
	additionalRedactors := &v1beta2.Redactor{}
	createOpts := supportbundle.SupportBundleCreateOpts{
		CollectorProgressCallback: collectorCB,
		KubernetesRestConfig:      k8sConfig,
		ProgressChan:              progressChan,
		SinceTime:                 sinceTimeValue,
	}

	archivePath, err := supportbundle.CollectSupportBundleFromSpec(spec, additionalRedactors, createOpts)
	if err != nil {
		return "", err
	}
	return archivePath, nil
}

func AnalyzeBundle(spec *v1beta2.SupportBundleSpec, archivePath string) error {
	analyzeResults, err := supportbundle.AnalyzeAndExtractSupportBundle(spec, archivePath)
	if err != nil {
		return err
	}
	if len(analyzeResults) > 0 {
		showAnalyzeResults(analyzeResults)
	}
	return nil
}

func ParseTimeOptions(since string, sinceTime string) (*time.Time, error) {
	var sinceTimeValue time.Time
	var err error
	if sinceTime == "" && since == "" {
		return &sinceTimeValue, nil
	} else if sinceTime != "" && since != "" {
		return nil, fmt.Errorf("at most one of `sinceTime` or `since` could be specified")
	} else if sinceTime != "" {
		sinceTimeValue, err = time.Parse(time.RFC3339, sinceTime)
		if err != nil {
			return nil, fmt.Errorf("unable to parse --since-time option: %v", err)
		}
	} else if since != "" {
		duration, err := time.ParseDuration(since)
		if err != nil {
			return nil, fmt.Errorf("unable to parse --since option: %v", err)
		}
		now := time.Now()
		sinceTimeValue = now.Add(0 - duration)
	}
	return &sinceTimeValue, nil
}

func showAnalyzeResults(analyzeResults []*analyzerunner.AnalyzeResult) {
	results := "\n Analyze Results" + resultsSeparator
	for _, analyzeResult := range analyzeResults {
		result := ""
		if analyzeResult.IsPass {
			result = "Check PASS\n"
		} else if analyzeResult.IsFail {
			result = "Check FAIL\n"
		}
		result = result + fmt.Sprintf("Title: %s\n", analyzeResult.Title)
		result = result + fmt.Sprintf("Message: %s\n", analyzeResult.Message)
		if analyzeResult.URI != "" {
			result = result + fmt.Sprintf("URI: %s\n", analyzeResult.URI)
		}
		result = result + resultsSeparator
		results = results + result
	}
	logger.Info(results)
}

func GenerateBundleConfig() error {
	config := newBundleConfig()
	clusterYaml, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error outputting yaml: %v", err)
	}
	fmt.Println(string(clusterYaml))
	return nil
}

func newBundleConfig() *v1beta2.SupportBundle {
	return &v1beta2.SupportBundle{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SupportBundle",
			APIVersion: "troubleshoot.sh/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "example",
		},
		Spec: v1beta2.SupportBundleSpec{
			Collectors: getDefaultCollectors(),
			Analyzers:  getDefaultAnalyzers(),
		},
	}
}

func getDefaultCollectors() []*v1beta2.Collect {
	return []*v1beta2.Collect{
		{
			ClusterInfo: &v1beta2.ClusterInfo{},
		},
		{
			ClusterResources: &v1beta2.ClusterResources{},
		},
		{
			Secret: &v1beta2.Secret{
				Namespace:    "eksa-system",
				SecretName:   "eksa-license",
				IncludeValue: true,
				Key:          "license",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "capd-system",
				Name:      "logs/capd-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "capi-kubeadm-bootstrap-system",
				Name:      "logs/capi-kubeadm-bootstrap-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "capi-kubeadm-control-plane-system",
				Name:      "logs/capi-kubeadm-control-plane-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "capi-system",
				Name:      "logs/capi-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "capi-webhook-system",
				Name:      "logs/capi-webhook-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "cert-manager",
				Name:      "logs/cert-manager",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "eksa-system",
				Name:      "logs/eksa-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "default",
				Name:      "logs/default",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "etcdadm-bootstrap-provider-system",
				Name:      "logs/etcdadm-bootstrap-provider-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "etcdadm-controller-system",
				Name:      "logs/etcdadm-controller-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "kube-node-lease",
				Name:      "logs/kube-node-lease",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "kube-public",
				Name:      "logs/kube-public",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "kube-system",
				Name:      "logs/kube-system",
			},
		},
	}
}

func getDefaultAnalyzers() []*v1beta2.Analyze {
	return []*v1beta2.Analyze{
		{
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "capi-kubeadm-bootstrap-controller-manager",
				Namespace: "capi-kubeadm-bootstrap-system",
				Outcomes: []*v1beta2.Outcome{
					{
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "capi-kubeadm-bootstrap-controller-manager is not ready.",
						},
					}, {
						Pass: &v1beta2.SingleOutcome{
							Message: "capi-kubeadm-bootstrap-controller-manager is running.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "capi-kubeadm-control-plane-controller-manager",
				Namespace: "capi-kubeadm-control-plane-system",
				Outcomes: []*v1beta2.Outcome{
					{
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "capi-kubeadm-control-plane-controller-manager is not ready.",
						},
					}, {
						Pass: &v1beta2.SingleOutcome{
							Message: "capi-kubeadm-control-plane-controller-manager is running.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "capi-controller-manager",
				Namespace: "capi-system",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "capi-controller-manager is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "capi-controller-manager is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "capi-controller-manager",
				Namespace: "capi-webhook-system",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "capi webhook controller manager is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "capi webhook controller manager is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "capi-kubeadm-bootstrap-controller-manager",
				Namespace: "capi-webhook-system",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "capi-kubeadm-bootstrap-controller-manager is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "capi-kubeadm-bootstrap-controller-manager is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "capi-kubeadm-control-plane-controller-manager",
				Namespace: "capi-webhook-system",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "capi-kubeadm-control-plane-controller-manager is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "capi-kubeadm-control-plane-controller-manager is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "cert-manager",
				Namespace: "cert-manager",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "cert-manager is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "cert-manager is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "cert-manager-cainjector",
				Namespace: "cert-manager",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "cert-manager-cainjector is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "cert-manager-cainjector is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "cert-manager-webhook",
				Namespace: "cert-manager",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "cert-manager-webhook is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "cert-manager-webhook is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "coredns",
				Namespace: "kube-system",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "coredns is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "coredns is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "local-path-provisioner",
				Namespace: "local-path-storage",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "local-path-provisioner is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "local-path-provisioner is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "capv-controller-manager",
				Namespace: "capv-system",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "capv-controller-manager is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "capv-controller-manager is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "capv-controller-manager",
				Namespace: "capi-webhook-system",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "capv-controller-manager is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "capv-controller-manager is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "etcdadm-bootstrap-provider-controller-manager",
				Namespace: "etcdadm-bootstrap-provider-system",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "etcdadm-bootstrap-provider-controller-manager is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "etcdadm-bootstrap-provider-controller-manager is not ready.",
						},
					},
				},
			},
		}, {
			DeploymentStatus: &v1beta2.DeploymentStatus{
				Name:      "etcdadm-controller-controller-manager",
				Namespace: "etcdadm-controller-system",
				Outcomes: []*v1beta2.Outcome{
					{
						Pass: &v1beta2.SingleOutcome{
							Message: "etcdadm-controller-controller-manager is running.",
						},
					}, {
						Fail: &v1beta2.SingleOutcome{
							When:    "< 1",
							Message: "etcdadm-controller-controller-manager is not ready.",
						},
					},
				},
			},
		},
	}
}

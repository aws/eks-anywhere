package diagnostics

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Collect struct {
	ClusterInfo      *clusterInfo      `json:"clusterInfo,omitempty"`
	ClusterResources *clusterResources `json:"clusterResources,omitempty"`
	Secret           *secret           `json:"secret,omitempty"`
	Logs             *logs             `json:"logs,omitempty"`
	Data             *data             `json:"data,omitempty"`
	CopyFromHost     *copyFromHost     `json:"copyFromHost,omitempty"`
	Exec             *exec             `json:"exec,omitempty"`
	RunPod           *runPod           `json:"runPod,omitempty"`
}

type clusterResources struct {
	collectorMeta `json:",inline"`
}

type secret struct {
	collectorMeta `json:",inline"`
	SecretName    string `json:"name"`
	Namespace     string `json:"namespace,omitempty"`
	Key           string `json:"key,omitempty"`
	IncludeValue  bool   `json:"includeValue,omitempty"`
}

type clusterInfo struct {
	collectorMeta `json:",inline"`
}

type logLimits struct {
	MaxAge    string `json:"maxAge,omitempty"`
	MaxLines  int64  `json:"maxLines,omitempty"`
	SinceTime metav1.Time
}

type logs struct {
	collectorMeta  `json:",inline" yaml:",inline"`
	Name           string     `json:"name,omitempty"`
	Selector       []string   `json:"selector"`
	Namespace      string     `json:"namespace,omitempty"`
	ContainerNames []string   `json:"containerNames,omitempty"`
	Limits         *logLimits `json:"limits,omitempty"`
}

type data struct {
	Name string `json:"name,omitempty"`
	Data string `json:"data,omitempty"`
}

type copyFromHost struct {
	collectorMeta   `json:",inline"`
	Name            string            `json:"name,omitempty"`
	Namespace       string            `json:"namespace"`
	Image           string            `json:"image"`
	ImagePullPolicy string            `json:"imagePullPolicy,omitempty"`
	ImagePullSecret *imagePullSecrets `json:"imagePullSecret,omitempty"`
	Timeout         string            `json:"timeout,omitempty"`
	HostPath        string            `json:"hostPath"`
	ExtractArchive  bool              `json:"extractArchive,omitempty"`
}

type exec struct {
	collectorMeta `json:",inline"`
	Name          string   `json:"name,omitempty"`
	Selector      []string `json:"selector"`
	Namespace     string   `json:"namespace"`
	ContainerName string   `json:"containerName,omitempty"`
	Command       []string `json:"command,omitempty"`
	Args          []string `json:"args,omitempty"`
	Timeout       string   `json:"timeout,omitempty"`
}

type imagePullSecrets struct {
	Name       string            `json:"name,omitempty"`
	Data       map[string]string `json:"data,omitempty"`
	SecretType string            `json:"type,omitempty"`
}

type collectorMeta struct {
	CollectorName string `json:"collectorName,omitempty"`
	Exclude       bool   `json:"exclude,omitempty"`
}

type runPod struct {
	collectorMeta    `json:",inline"`
	Name             string      `json:"name,omitempty"`
	Namespace        string      `json:"namespace"`
	PodSpec          *v1.PodSpec `json:"podSpec,omitempty"`
	Timeout          string      `json:"timeout,omitempty"`
	imagePullSecrets `json:",inline"`
}

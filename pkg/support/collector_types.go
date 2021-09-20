package supportbundle

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Collect struct {
	ClusterInfo      *clusterInfo      `json:"clusterInfo,omitempty"`
	ClusterResources *clusterResources `json:"clusterResources,omitempty"`
	Secret           *secret           `json:"secret,omitempty"`
	Logs             *logs             `json:"logs,omitempty"`
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

type collectorMeta struct {
	CollectorName string `json:"collectorName,omitempty"`
	Exclude       bool   `json:"exclude,omitempty"`
}

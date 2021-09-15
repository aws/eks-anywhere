package supportbundle

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Collect struct {
	ClusterInfo      *clusterInfo      `json:"clusterInfo,omitempty" yaml:"clusterInfo,omitempty"`
	ClusterResources *clusterResources `json:"clusterResources,omitempty" yaml:"clusterResources,omitempty"`
	Secret           *secret           `json:"secret,omitempty" yaml:"secret,omitempty"`
	Logs             *logs             `json:"logs,omitempty" yaml:"logs,omitempty"`
}

type clusterResources struct {
	collectorMeta `json:",inline" yaml:",inline"`
}

type secret struct {
	collectorMeta `json:",inline" yaml:",inline"`
	SecretName    string `json:"name" yaml:"name"`
	Namespace     string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Key           string `json:"key,omitempty" yaml:"key,omitempty"`
	IncludeValue  bool   `json:"includeValue,omitempty" yaml:"includeValue,omitempty"`
}

type clusterInfo struct {
	collectorMeta `json:",inline" yaml:",inline"`
}

type logLimits struct {
	MaxAge    string `json:"maxAge,omitempty" yaml:"maxAge,omitempty"`
	MaxLines  int64  `json:"maxLines,omitempty" yaml:"maxLines,omitempty"`
	SinceTime metav1.Time
}

type logs struct {
	collectorMeta  `json:",inline" yaml:",inline"`
	Name           string     `json:"name,omitempty" yaml:"name,omitempty"`
	Selector       []string   `json:"selector" yaml:"selector"`
	Namespace      string     `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	ContainerNames []string   `json:"containerNames,omitempty" yaml:"containerNames,omitempty"`
	Limits         *logLimits `json:"limits,omitempty" yaml:"omitempty"`
}

type collectorMeta struct {
	CollectorName string `json:"collectorName,omitempty" yaml:"collectorName,omitempty"`
	Exclude       bool   `json:"exclude,omitempty" yaml:"exclude,omitempty"`
}

// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Requires struct {
	Kind     string            `json:"kind,omitempty"`
	Metadata metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec     RequiresSpec      `json:"spec,omitempty"`
}

type RequiresSpec struct {
	Images         []Image         `json:"images,omitempty"`
	Configurations []Configuration `json:"configurations,omitempty"`
	Schema         string          `json:"schema,omitempty"`
}

type Configuration struct {
	Name     string `json:"name,omitempty"`
	Required bool   `json:"required,omitempty"`
	Default  string `json:"default,omitempty"`
}

type Image struct {
	Repository string `json:"repository,omitempty"`
	Tag        string `json:"tag,omitempty"`
	Digest     string `json:"digest,omitempty"`
}

type DockerAuth struct {
	Auths map[string]DockerAuthRegistry `json:"auths,omitempty"`
}

type DockerAuthRegistry struct {
	Auth string `json:"auth"`
}

type DockerAuthFile struct {
	Authfile string `json:"authfile"`
}

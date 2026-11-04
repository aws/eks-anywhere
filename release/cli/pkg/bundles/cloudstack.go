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

package bundles

import (
	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// GetDeprecatedCloudStackBundle returns a CloudStackBundle with placeholder values.
// CloudStack has been deprecated and is no longer built or released.
// Customers can override these placeholders via bundle override.
func GetDeprecatedCloudStackBundle() anywherev1alpha1.CloudStackBundle {
	return anywherev1alpha1.CloudStackBundle{
		Version: "deprecated",
		ClusterAPIController: anywherev1alpha1.Image{
			Name: "cluster-api-provider-cloudstack",
			URI:  "<placeholder>",
		},
		KubeRbacProxy: anywherev1alpha1.Image{
			Name: "kube-rbac-proxy",
			URI:  "<placeholder>",
		},
		KubeVip: anywherev1alpha1.Image{
			Name: "kube-vip",
			URI:  "<placeholder>",
		},
		Components: anywherev1alpha1.Manifest{
			URI: "<placeholder>",
		},
		Metadata: anywherev1alpha1.Manifest{
			URI: "<placeholder>",
		},
	}
}

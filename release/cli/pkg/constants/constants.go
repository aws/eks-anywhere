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

package constants

const (
	// Artifacts-related constants.
	ReleaseKind              = "Release"
	BundlesKind              = "Bundles"
	HexAlphabet              = "0123456789abcdef"
	SuccessIcon              = "✅"
	FakeComponentChecksum    = "abcdef1"
	FakeGitCommit            = "0123456789abcdef0123456789abcdef01234567"
	ReleaseFolderName        = "release"
	EksDReleaseComponentsUrl = "https://distro.eks.amazonaws.com/crds/releases.distro.eks.amazonaws.com-v1alpha1.yaml"
	YamlSeparator            = "\n---\n"

	// Project paths.
	CapasProjectPath                    = "projects/aws/cluster-api-provider-aws-snow"
	CapcProjectPath                     = "projects/kubernetes-sigs/cluster-api-provider-cloudstack"
	CapiProjectPath                     = "projects/kubernetes-sigs/cluster-api"
	CaptProjectPath                     = "projects/tinkerbell/cluster-api-provider-tinkerbell"
	CapvProjectPath                     = "projects/kubernetes-sigs/cluster-api-provider-vsphere"
	CapxProjectPath                     = "projects/nutanix-cloud-native/cluster-api-provider-nutanix"
	CertManagerProjectPath              = "projects/cert-manager/cert-manager"
	CiliumProjectPath                   = "projects/cilium/cilium"
	EtcdadmBootstrapProviderProjectPath = "projects/aws/etcdadm-bootstrap-provider"
	EtcdadmControllerProjectPath        = "projects/aws/etcdadm-controller"
	FluxcdRootPath                      = "projects/fluxcd"
	Flux2ProjectPath                    = "projects/fluxcd/flux2"
	HookProjectPath                     = "projects/tinkerbell/hook"
	ImageBuilderProjectPath             = "projects/kubernetes-sigs/image-builder"
	KindProjectPath                     = "projects/kubernetes-sigs/kind"
	KubeRbacProxyProjectPath            = "projects/brancz/kube-rbac-proxy"
	PackagesProjectPath                 = "projects/aws/eks-anywhere-packages"
	UpgraderProjectPath                 = "projects/aws/upgrader"

	// Date format with standard reference time values
	// The reference time used is the specific time stamp:
	//
	//	01/02 03:04:05PM '06 -0700
	//
	// (January 2, 15:04:05, 2006, in time zone seven hours west of GMT).
	YYYYMMDD = "2006-01-02"
)

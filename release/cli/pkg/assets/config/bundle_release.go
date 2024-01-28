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

package config

import (
	"github.com/aws/eks-anywhere/release/cli/pkg/assets/archives"
	"github.com/aws/eks-anywhere/release/cli/pkg/assets/tagger"
	assettypes "github.com/aws/eks-anywhere/release/cli/pkg/assets/types"
)

var bundleReleaseAssetsConfigMap = []assettypes.AssetConfig{
	// Boots artifacts
	{
		ProjectName: "boots",
		ProjectPath: "projects/tinkerbell/boots",
		Images: []*assettypes.Image{
			{
				RepoName: "boots",
			},
		},
		ImageRepoPrefix: "tinkerbell",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Bottlerocket-bootstrap artifacts
	{
		ProjectName:    "bottlerocket-bootstrap",
		ProjectPath:    "projects/aws/bottlerocket-bootstrap",
		GitTagAssigner: tagger.NonExistentTagAssigner,
		Images: []*assettypes.Image{
			{
				RepoName: "bottlerocket-bootstrap",
				ImageTagConfiguration: assettypes.ImageTagConfiguration{
					NonProdSourceImageTagFormat: "v<eksDReleaseChannel>-<eksDReleaseNumber>",
					ProdSourceImageTagFormat:    "v<eksDReleaseChannel>-<eksDReleaseNumber>",
					ReleaseImageTagFormat:       "v<eksDReleaseChannel>-<eksDReleaseNumber>",
				},
			},
			{
				RepoName: "bottlerocket-bootstrap-snow",
				ImageTagConfiguration: assettypes.ImageTagConfiguration{
					NonProdSourceImageTagFormat: "v<eksDReleaseChannel>-<eksDReleaseNumber>",
					ProdSourceImageTagFormat:    "v<eksDReleaseChannel>-<eksDReleaseNumber>",
					ReleaseImageTagFormat:       "v<eksDReleaseChannel>-<eksDReleaseNumber>",
				},
			},
		},
		ImageTagOptions: []string{
			"eksDReleaseChannel",
			"eksDReleaseNumber",
			"gitTag",
		},
		HasReleaseBranches: true,
	},
	// Cert-manager artifacts
	{
		ProjectName: "cert-manager",
		ProjectPath: "projects/cert-manager/cert-manager",
		Images: []*assettypes.Image{
			{
				RepoName: "cert-manager-acmesolver",
			},
			{
				RepoName: "cert-manager-cainjector",
			},
			{
				RepoName: "cert-manager-controller",
			},
			{
				RepoName: "cert-manager-ctl",
			},
			{
				RepoName: "cert-manager-webhook",
			},
		},
		ImageRepoPrefix: "cert-manager",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				ManifestFiles: []string{"cert-manager.yaml"},
			},
		},
	},
	// Cilium artifacts
	{
		ProjectName: "cilium",
		ProjectPath: "projects/cilium/cilium",
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:          "cilium",
				ManifestFiles: []string{"cilium.yaml"},
			},
		},
	},
	// Cloud-provider-nutanix artifacts
	{
		ProjectName: "cloud-provider-nutanix",
		ProjectPath: "projects/nutanix-cloud-native/cloud-provider-nutanix",
		Images: []*assettypes.Image{
			{
				RepoName:  "controller",
				AssetName: "cloud-provider-nutanix",
			},
		},
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		ImageRepoPrefix: "nutanix-cloud-native/cloud-provider-nutanix",
	},
	// Cloud-provider-vsphere artifacts
	{
		ProjectName: "cloud-provider-vsphere",
		ProjectPath: "projects/kubernetes/cloud-provider-vsphere",
		Images: []*assettypes.Image{
			{
				RepoName:  "manager",
				AssetName: "cloud-provider-vsphere",
				ImageTagConfiguration: assettypes.ImageTagConfiguration{
					NonProdSourceImageTagFormat: "<gitTag>",
					ProdSourceImageTagFormat:    "<gitTag>-eks-d-<eksDReleaseChannel>",
					ReleaseImageTagFormat:       "<gitTag>-eks-d-<eksDReleaseChannel>",
				},
			},
		},
		ImageRepoPrefix: "kubernetes/cloud-provider-vsphere/cpi",
		ImageTagOptions: []string{
			"eksDReleaseChannel",
			"gitTag",
			"projectPath",
		},
		HasReleaseBranches:             true,
		HasSeparateTagPerReleaseBranch: true,
	},
	// Cluster-api artifacts
	{
		ProjectName: "cluster-api",
		ProjectPath: "projects/kubernetes-sigs/cluster-api",
		Images: []*assettypes.Image{
			{
				RepoName: "cluster-api-controller",
			},
			{
				RepoName: "kubeadm-bootstrap-controller",
			},
			{
				RepoName: "kubeadm-control-plane-controller",
			},
		},
		ImageRepoPrefix: "kubernetes-sigs/cluster-api",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:          "cluster-api",
				ManifestFiles: []string{"core-components.yaml", "metadata.yaml"},
			},
			{
				Name:          "bootstrap-kubeadm",
				ManifestFiles: []string{"bootstrap-components.yaml", "metadata.yaml"},
			},
			{
				Name:          "control-plane-kubeadm",
				ManifestFiles: []string{"control-plane-components.yaml", "metadata.yaml"},
			},
		},
	},
	// Cluster-api-provider-aws-snow artifacts
	{
		ProjectName: "cluster-api-provider-aws-snow",
		ProjectPath: "projects/aws/cluster-api-provider-aws-snow",
		Images: []*assettypes.Image{
			{
				RepoName:  "manager",
				AssetName: "cluster-api-snow-controller",
			},
		},
		ImageRepoPrefix: "aws/cluster-api-provider-aws-snow",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:          "infrastructure-snow",
				ManifestFiles: []string{"infrastructure-components.yaml", "metadata.yaml"},
			},
		},
	},
	// Cluster-api-provider-cloudstack artifacts
	{
		ProjectName: "cluster-api-provider-cloudstack",
		ProjectPath: "projects/kubernetes-sigs/cluster-api-provider-cloudstack",
		Images: []*assettypes.Image{
			{
				RepoName:  "manager",
				AssetName: "cluster-api-provider-cloudstack",
			},
		},
		ImageRepoPrefix: "kubernetes-sigs/cluster-api-provider-cloudstack/release",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:          "infrastructure-cloudstack",
				ManifestFiles: []string{"infrastructure-components.yaml", "metadata.yaml"},
			},
		},
		UsesKubeRbacProxy: true,
	},
	// Cluster-api-provider-docker artifacts
	{
		ProjectName: "cluster-api-provider-docker",
		ProjectPath: "projects/kubernetes-sigs/cluster-api",
		Images: []*assettypes.Image{
			{
				RepoName:  "capd-manager",
				AssetName: "cluster-api-provider-docker",
			},
		},
		ImageRepoPrefix: "kubernetes-sigs/cluster-api",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:                  "infrastructure-docker",
				ManifestFiles:         []string{"infrastructure-components-development.yaml", "cluster-template-development.yaml", "metadata.yaml"},
				ReleaseManifestPrefix: "cluster-api",
			},
		},
	},
	// Cluster-api-provider-nutanix artifacts
	{
		ProjectName: "cluster-api-provider-nutanix",
		ProjectPath: "projects/nutanix-cloud-native/cluster-api-provider-nutanix",
		Images: []*assettypes.Image{
			{
				RepoName: "cluster-api-provider-nutanix",
			},
		},
		ImageRepoPrefix: "nutanix-cloud-native",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:          "infrastructure-nutanix",
				ManifestFiles: []string{"infrastructure-components.yaml", "cluster-template.yaml", "metadata.yaml"},
			},
		},
		UsesKubeRbacProxy: true,
	},
	// Cluster-api-provider-tinkerbell artifacts
	{
		ProjectName: "cluster-api-provider-tinkerbell",
		ProjectPath: "projects/tinkerbell/cluster-api-provider-tinkerbell",
		Images: []*assettypes.Image{
			{
				RepoName: "cluster-api-provider-tinkerbell",
			},
		},
		ImageRepoPrefix: "tinkerbell",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:          "infrastructure-tinkerbell",
				ManifestFiles: []string{"infrastructure-components.yaml", "cluster-template.yaml", "metadata.yaml"},
			},
		},
	},
	// Cluster-api-provider-vsphere artifacts
	{
		ProjectName: "cluster-api-provider-vsphere",
		ProjectPath: "projects/kubernetes-sigs/cluster-api-provider-vsphere",
		Images: []*assettypes.Image{
			{
				RepoName:  "manager",
				AssetName: "cluster-api-provider-vsphere",
			},
		},
		ImageRepoPrefix: "kubernetes-sigs/cluster-api-provider-vsphere/release",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:          "infrastructure-vsphere",
				ManifestFiles: []string{"infrastructure-components.yaml", "cluster-template.yaml", "metadata.yaml"},
			},
		},
	},
	// Containerd artifacts
	{
		ProjectName: "containerd",
		ProjectPath: "projects/containerd/containerd",
		Archives: []*assettypes.Archive{
			{
				Name:   "containerd",
				Format: "tarball",
			},
		},
	},
	// Image-builder cli artifacts
	{
		ProjectName: "image-builder",
		ProjectPath: "projects/aws/image-builder",
		Archives: []*assettypes.Archive{
			{
				Name:   "image-builder",
				Format: "tarball",
			},
		},
	},
	// Cri-tools artifacts
	{
		ProjectName: "cri-tools",
		ProjectPath: "projects/kubernetes-sigs/cri-tools",
		Archives: []*assettypes.Archive{
			{
				Name:   "cri-tools",
				Format: "tarball",
			},
		},
	},
	// EKS-A CLI tools artifacts
	{
		ProjectName:    "eks-anywhere-cli-tools",
		ProjectPath:    "projects/aws/eks-anywhere-build-tooling",
		GitTagAssigner: tagger.CliGitTagAssigner,
		Images: []*assettypes.Image{
			{
				RepoName:       "eks-anywhere-cli-tools",
				TrimEksAPrefix: true,
			},
		},
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// EKS-A cluster-controller artifacts
	{
		ProjectName:    "eks-anywhere-cluster-controller",
		ProjectPath:    "projects/aws/eks-anywhere-cluster-controller",
		GitTagAssigner: tagger.CliGitTagAssigner,
		Images: []*assettypes.Image{
			{
				RepoName:       "eks-anywhere-cluster-controller",
				TrimEksAPrefix: true,
			},
		},
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:                  "cluster-controller",
				ManifestFiles:         []string{"eksa-components.yaml"},
				ReleaseManifestPrefix: "eks-anywhere",
				NoVersionSuffix:       true,
			},
		},
	},
	// EKS-A diagnostic collector artifacts
	{
		ProjectName:    "eks-anywhere-diagnostic-collector",
		ProjectPath:    "projects/aws/eks-anywhere",
		GitTagAssigner: tagger.CliGitTagAssigner,
		Images: []*assettypes.Image{
			{
				RepoName:       "eks-anywhere-diagnostic-collector",
				TrimEksAPrefix: true,
			},
		},
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// EKS-A package controller artifacts
	{
		ProjectName: "eks-anywhere-packages",
		ProjectPath: "projects/aws/eks-anywhere-packages",
		Images: []*assettypes.Image{
			{
				RepoName: "eks-anywhere-packages",
			},
			{
				RepoName: "ecr-token-refresher",
			},
			{
				RepoName: "credential-provider-package",
			},
			{
				AssetName:            "eks-anywhere-packages-helm",
				RepoName:             "eks-anywhere-packages",
				TrimVersionSignifier: true,
				ImageTagConfiguration: assettypes.ImageTagConfiguration{
					NonProdSourceImageTagFormat: "<gitTag>",
				},
			},
		},
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Etcdadm artifacts
	{
		ProjectName: "etcdadm",
		ProjectPath: "projects/kubernetes-sigs/etcdadm",
		Archives: []*assettypes.Archive{
			{
				Name:   "etcdadm",
				Format: "tarball",
			},
		},
	},
	// Etcdadm-bootstrap-provider artifacts
	{
		ProjectName: "etcdadm-bootstrap-provider",
		ProjectPath: "projects/aws/etcdadm-bootstrap-provider",
		Images: []*assettypes.Image{
			{
				RepoName: "etcdadm-bootstrap-provider",
			},
		},
		ImageRepoPrefix: "aws",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:          "bootstrap-etcdadm-bootstrap",
				ManifestFiles: []string{"bootstrap-components.yaml", "metadata.yaml"},
			},
		},
	},
	// Etcdadm-controller artifacts
	{
		ProjectName: "etcdadm-controller",
		ProjectPath: "projects/aws/etcdadm-controller",
		Images: []*assettypes.Image{
			{
				RepoName: "etcdadm-controller",
			},
		},
		ImageRepoPrefix: "aws",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:          "bootstrap-etcdadm-controller",
				ManifestFiles: []string{"bootstrap-components.yaml", "metadata.yaml"},
			},
		},
	},
	// HAProxy artifacts
	{
		ProjectName: "haproxy",
		ProjectPath: "projects/kubernetes-sigs/kind",
		Images: []*assettypes.Image{
			{
				RepoName: "haproxy",
			},
		},
		ImageRepoPrefix: "kubernetes-sigs/kind",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Hegel artifacts
	{
		ProjectName: "hegel",
		ProjectPath: "projects/tinkerbell/hegel",
		Images: []*assettypes.Image{
			{
				RepoName: "hegel",
			},
		},
		ImageRepoPrefix: "tinkerbell",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Helm-controller artifacts
	{
		ProjectName: "helm-controller",
		ProjectPath: "projects/fluxcd/helm-controller",
		Images: []*assettypes.Image{
			{
				RepoName: "helm-controller",
			},
		},
		ImageRepoPrefix: "fluxcd",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Hook artifacts
	{
		ProjectName: "hook",
		ProjectPath: "projects/tinkerbell/hook",
		Images: []*assettypes.Image{
			{
				RepoName: "hook-bootkit",
			},
			{
				RepoName: "hook-docker",
			},
			{
				RepoName: "hook-kernel",
			},
		},
		ImageRepoPrefix: "tinkerbell",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Archives: []*assettypes.Archive{
			{
				Name:                 "initramfs-aarch64",
				Format:               "kernel",
				ArchitectureOverride: "arm64",
				ArchiveS3PathGetter:  archives.KernelArtifactPathGetter,
			},
			{
				Name:                "initramfs-x86_64",
				Format:              "kernel",
				ArchiveS3PathGetter: archives.KernelArtifactPathGetter,
			},
			{
				Name:                 "vmlinuz-aarch64",
				Format:               "kernel",
				ArchitectureOverride: "arm64",
				ArchiveS3PathGetter:  archives.KernelArtifactPathGetter,
			},
			{
				Name:                "vmlinuz-x86_64",
				Format:              "kernel",
				ArchiveS3PathGetter: archives.KernelArtifactPathGetter,
			},
		},
	},
	// Hub artifacts
	{
		ProjectName: "hub",
		ProjectPath: "projects/tinkerbell/hub",
		Images: []*assettypes.Image{
			{
				RepoName: "cexec",
			},
			{
				RepoName: "image2disk",
			},
			{
				RepoName: "kexec",
			},
			{
				RepoName: "oci2disk",
			},
			{
				RepoName: "reboot",
			},
			{
				RepoName: "writefile",
			},
		},
		ImageRepoPrefix: "tinkerbell/hub",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Image-builder artifacts
	{
		ProjectName: "image-builder",
		ProjectPath: "projects/kubernetes-sigs/image-builder",
		Archives: []*assettypes.Archive{
			{
				Name:                "eks-distro",
				OSName:              "bottlerocket",
				OSVersion:           "1",
				Format:              "ami",
				ArchiveS3PathGetter: archives.EksDistroArtifactPathGetter,
			},
			{
				Name:                "eks-distro",
				OSName:              "bottlerocket",
				OSVersion:           "1",
				Format:              "ova",
				ArchiveS3PathGetter: archives.EksDistroArtifactPathGetter,
			},
			{
				Name:                "eks-distro",
				OSName:              "bottlerocket",
				OSVersion:           "1",
				Format:              "raw",
				ArchiveS3PathGetter: archives.EksDistroArtifactPathGetter,
			},
		},
		HasReleaseBranches: true,
	},
	// Kind artifacts
	{
		ProjectName: "kind",
		ProjectPath: "projects/kubernetes-sigs/kind",
		Images: []*assettypes.Image{
			{
				RepoName:  "node",
				AssetName: "kind-node",
				ImageTagConfiguration: assettypes.ImageTagConfiguration{
					NonProdSourceImageTagFormat: "<kubeVersion>-eks-<eksDReleaseChannel>-<eksDReleaseNumber>",
					ProdSourceImageTagFormat:    "<kubeVersion>-eks-d-<eksDReleaseChannel>-<eksDReleaseNumber>",
					ReleaseImageTagFormat:       "<kubeVersion>-eks-d-<eksDReleaseChannel>-<eksDReleaseNumber>",
				},
			},
		},
		ImageRepoPrefix: "kubernetes-sigs/kind",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
			"eksDReleaseChannel",
			"eksDReleaseNumber",
			"kubeVersion",
			"projectPath",
		},
		HasReleaseBranches: true,
	},
	// Kindnetd artifacts
	{
		ProjectName: "kindnetd",
		ProjectPath: "projects/kubernetes-sigs/kind",
		Images: []*assettypes.Image{
			{
				RepoName: "kindnetd",
			},
		},
		ImageRepoPrefix: "kubernetes-sigs/kind",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
		Manifests: []*assettypes.ManifestComponent{
			{
				Name:                  "kindnetd",
				ManifestFiles:         []string{"kindnetd.yaml"},
				ReleaseManifestPrefix: "kind",
			},
		},
	},
	// Kube-rbac-proxy artifacts
	{
		ProjectName: "kube-rbac-proxy",
		ProjectPath: "projects/brancz/kube-rbac-proxy",
		Images: []*assettypes.Image{
			{
				RepoName: "kube-rbac-proxy",
			},
		},
		ImageRepoPrefix: "brancz",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Kube-vip artifacts
	{
		ProjectName: "kube-vip",
		ProjectPath: "projects/kube-vip/kube-vip",
		Images: []*assettypes.Image{
			{
				RepoName: "kube-vip",
			},
		},
		ImageRepoPrefix: "kube-vip",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Envoy artifacts
	{
		ProjectName: "envoy",
		ProjectPath: "projects/envoyproxy/envoy",
		Images: []*assettypes.Image{
			{
				RepoName: "envoy",
			},
		},
		ImageRepoPrefix: "envoyproxy",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Kustomize-controller artifacts
	{
		ProjectName: "kustomize-controller",
		ProjectPath: "projects/fluxcd/kustomize-controller",
		Images: []*assettypes.Image{
			{
				RepoName: "kustomize-controller",
			},
		},
		ImageRepoPrefix: "fluxcd",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Local-path-provisioner artifacts
	{
		ProjectName: "local-path-provisioner",
		ProjectPath: "projects/rancher/local-path-provisioner",
		Images: []*assettypes.Image{
			{
				RepoName: "local-path-provisioner",
			},
		},
		ImageRepoPrefix: "rancher",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Notification-controller artifacts
	{
		ProjectName: "notification-controller",
		ProjectPath: "projects/fluxcd/notification-controller",
		Images: []*assettypes.Image{
			{
				RepoName: "notification-controller",
			},
		},
		ImageRepoPrefix: "fluxcd",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Rufio artifacts
	{
		ProjectName: "rufio",
		ProjectPath: "projects/tinkerbell/rufio",
		Images: []*assettypes.Image{
			{
				RepoName: "rufio",
			},
		},
		ImageRepoPrefix: "tinkerbell",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Source-controller artifacts
	{
		ProjectName: "source-controller",
		ProjectPath: "projects/fluxcd/source-controller",
		Images: []*assettypes.Image{
			{
				RepoName: "source-controller",
			},
		},
		ImageRepoPrefix: "fluxcd",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Tink artifacts
	{
		ProjectName: "tink",
		ProjectPath: "projects/tinkerbell/tink",
		Images: []*assettypes.Image{
			{
				RepoName: "tink-controller",
			},
			{
				RepoName: "tink-server",
			},
			{
				RepoName: "tink-worker",
			},
		},
		ImageRepoPrefix: "tinkerbell/tink",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Tinkerbell chart artifacts
	{
		ProjectName: "tinkerbell-chart",
		ProjectPath: "projects/tinkerbell/tinkerbell-chart",
		Images: []*assettypes.Image{
			{
				RepoName:             "tinkerbell-chart",
				TrimVersionSignifier: true,
				ImageTagConfiguration: assettypes.ImageTagConfiguration{
					NonProdSourceImageTagFormat: "<gitTag>",
				},
			},
		},
		ImageRepoPrefix: "tinkerbell",
		ImageTagOptions: []string{
			"gitTag",
			"projectPath",
		},
	},
	// Upgrader artifacts
	{
		ProjectName:    "upgrader",
		ProjectPath:    "projects/aws/upgrader",
		GitTagAssigner: tagger.NonExistentTagAssigner,
		Images: []*assettypes.Image{
			{
				RepoName: "upgrader",
				ImageTagConfiguration: assettypes.ImageTagConfiguration{
					NonProdSourceImageTagFormat: "v<eksDReleaseChannel>-<eksDReleaseNumber>",
					ProdSourceImageTagFormat:    "v<eksDReleaseChannel>-<eksDReleaseNumber>",
					ReleaseImageTagFormat:       "v<eksDReleaseChannel>-<eksDReleaseNumber>",
				},
			},
		},
		ImageRepoPrefix: "aws",
		ImageTagOptions: []string{
			"eksDReleaseChannel",
			"eksDReleaseNumber",
			"gitTag",
		},
		HasReleaseBranches: true,
	},
}

func GetBundleReleaseAssetsConfigMap() []assettypes.AssetConfig {
	return bundleReleaseAssetsConfigMap
}

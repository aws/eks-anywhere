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

package pkg

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/pkg/errors"
)

var imageBuilderProjectSource = "projects/kubernetes-sigs/image-builder"

// ReleaseConfig contains metadata fields for a release
type ReleaseConfig struct {
	ReleaseVersion           string
	DevReleaseUriVersion     string
	BundleNumber             int
	CliMinVersion            string
	CliMaxVersion            string
	CliRepoSource            string
	CliRepoHead              string
	BuildRepoSource          string
	BuildRepoHead            string
	ArtifactDir              string
	SourceBucket             string
	ReleaseBucket            string
	SourceContainerRegistry  string
	ReleaseContainerRegistry string
	CDN                      string
	ReleaseNumber            int
	ReleaseDate              time.Time
	DevRelease               bool
	ReleaseEnvironment       string
}

// GetArtifactsData will get asset information for each component
// This information will be used to download them (in case of dev release)
// Rename them, create the manifest and to upload the artifacts to the
// proper location in S3 or ECR.
func (r *ReleaseConfig) GetVersionsBundles(imageDigests map[string]string) ([]anywherev1alpha1.VersionsBundle, error) {
	versionsBundles := []anywherev1alpha1.VersionsBundle{}

	certManagerBundle, err := r.GetCertManagerBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for cert-manager")
	}

	coreClusterApiBundle, err := r.GetCoreClusterAPIBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for core cluster-api")
	}

	kubeadmBootstrapBundle, err := r.GetKubeadmBootstrapBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for cluster-api kubeadm-bootstrap")
	}

	kubeadmControlPlaneBundle, err := r.GetKubeadmControlPlaneBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for cluster-api kubeadm-control-plane")
	}

	awsBundle, err := r.GetAwsBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for AWS infrastructure provider")
	}

	dockerBundle, err := r.GetDockerBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Docker infrastructure provider")
	}

	eksaBundle, err := r.GetEksaBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for eks-a tools component")
	}

	ciliumBundle, err := r.GetCiliumBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Cilium")
	}

	fluxBundle, err := r.GetFluxBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Flux controllers")
	}

	etcdadmBootstrapBundle, err := r.GetEtcdadmBootstrapBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for external Etcdadm bootstrap")
	}

	etcdadmControllerBundle, err := r.GetEtcdadmControllerBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for external Etcdadm controller")
	}

	eksDReleaseMap, err := readEksDReleases(r)
	if err != nil {
		return nil, err
	}

	bottlerocketSupportedK8sVersions, err := getBottlerocketSupportedK8sVersions(r, imageBuilderProjectSource)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting supported Kubernetes versions for bottlerocket")
	}

	for channel, release := range eksDReleaseMap {
		if channel == "latest" || !existsInList(channel, bottlerocketSupportedK8sVersions) {
			continue
		}
		releaseNumber := release.(map[interface{}]interface{})["number"]
		releaseNumberInt := releaseNumber.(int)
		releaseNumberStr := strconv.Itoa(releaseNumberInt)

		kubeVersion := release.(map[interface{}]interface{})["kubeVersion"]
		kubeVersionStr := kubeVersion.(string)
		shortKubeVersion := kubeVersionStr[1:strings.LastIndex(kubeVersionStr, ".")]

		eksDReleaseBundle, err := r.GetEksDReleaseBundle(channel, kubeVersionStr, releaseNumberStr, imageDigests)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for eks-d %s-%s release bundle", channel, releaseNumberStr)
		}

		vsphereBundle, err := r.GetVsphereBundle(channel, imageDigests)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for vSphere infrastructure provider")
		}

		bottlerocketBootstrapBundle, err := r.GetBottlerocketBootstrapBundle(channel, releaseNumberStr, imageDigests)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for bottlerocket bootstrap")
		}

		versionsBundle := anywherev1alpha1.VersionsBundle{
			KubeVersion:            shortKubeVersion,
			EksD:                   eksDReleaseBundle,
			CertManager:            certManagerBundle,
			ClusterAPI:             coreClusterApiBundle,
			Bootstrap:              kubeadmBootstrapBundle,
			ControlPlane:           kubeadmControlPlaneBundle,
			Aws:                    awsBundle,
			VSphere:                vsphereBundle,
			Docker:                 dockerBundle,
			Eksa:                   eksaBundle,
			Cilium:                 ciliumBundle,
			Flux:                   fluxBundle,
			ExternalEtcdBootstrap:  etcdadmBootstrapBundle,
			ExternalEtcdController: etcdadmControllerBundle,
			BottleRocketBootstrap:  bottlerocketBootstrapBundle,
		}
		versionsBundles = append(versionsBundles, versionsBundle)
	}
	return versionsBundles, nil
}

func (r *ReleaseConfig) GenerateBundleSpec(bundles *anywherev1alpha1.Bundles, imageDigests map[string]string) error {
	fmt.Println("Generating versions bundles")
	versionsBundles, err := r.GetVersionsBundles(imageDigests)
	fmt.Println(versionsBundles)
	if err != nil {
		return err
	}

	bundles.Spec.VersionsBundles = versionsBundles
	return nil
}

// GetArtifactsData will get asset information for each component
// This information will be used to download them (in case of dev release)
// Rename them, create the manifest and to upload the artifacts to the
// proper location in S3 or ECR.
func (r *ReleaseConfig) GetBundleArtifactsData() (map[string][]Artifact, error) {
	artifactsTable := map[string][]Artifact{}
	eksAArtifactsFuncs := map[string]func() ([]Artifact, error){
		"eks-a-tools":                  r.GetEksAToolsAssets,
		"cluster-api":                  r.GetCapiAssets,
		"cluster-api-provider-aws":     r.GetCapaAssets,
		"cluster-api-provider-docker":  r.GetDockerAssets,
		"cluster-api-provider-vsphere": r.GetCapvAssets,
		"vsphere-csi-driver":           r.GetVsphereCsiAssets,
		"cert-manager":                 r.GetCertManagerAssets,
		"cilium":                       r.GetCiliumAssets,
		"local-path-provisioner":       r.GetLocalPathProvisionerAssets,
		"kube-rbac-proxy":              r.GetKubeRbacProxyAssets,
		"kube-vip":                     r.GetKubeVipAssets,
		"flux":                         r.GetFluxAssets,
		"etcdadm-bootstrap-provider":   r.GetEtcdadmBootstrapAssets,
		"etcdadm-controller":           r.GetEtcdadmControllerAssets,
		"cluster-controller":           r.GetClusterControllerAssets,
		"kindnetd":                     r.GetKindnetdAssets,
		"etcdadm":                      r.GetEtcdadmAssets,
		"cri-tools":                    r.GetCriToolsAssets,
		"diagnostic-collector":         r.GetDiagnosticCollectorAssets,
	}

	for componentName, artifactFunc := range eksAArtifactsFuncs {
		artifacts, err := artifactFunc()
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting artifact information for %s", componentName)
		}

		artifactsTable[componentName] = artifacts
	}

	eksDReleaseMap, err := readEksDReleases(r)
	if err != nil {
		return nil, err
	}

	bottlerocketSupportedK8sVersions, err := getBottlerocketSupportedK8sVersions(r, imageBuilderProjectSource)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting supported Kubernetes versions for bottlerocket")
	}
	for channel, release := range eksDReleaseMap {
		if channel == "latest" || !existsInList(channel, bottlerocketSupportedK8sVersions) {
			continue
		}
		releaseNumber := release.(map[interface{}]interface{})["number"]
		releaseNumberInt := releaseNumber.(int)
		releaseNumberStr := strconv.Itoa(releaseNumberInt)

		kubeVersion := release.(map[interface{}]interface{})["kubeVersion"]
		kubeVersionStr := kubeVersion.(string)

		artifacts, err := r.GetEksDChannelAssets(channel, kubeVersionStr, releaseNumberStr)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting artifact information for %s", channel)
		}

		vSphereCloudProviderArtifacts, err := r.GetVsphereCloudProviderAssets(channel)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting artifact information for %s", channel)
		}

		bottlerocketBootstrapArtifacts, err := r.GetBottlerocketBootstrapAssets(channel, releaseNumberStr)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting artifact information for %s", channel)
		}

		eksDComponentName := fmt.Sprintf("eks-d-%s", channel)
		artifactsTable[eksDComponentName] = artifacts

		vSphereCloudProviderComponentName := fmt.Sprintf("cloud-provider-vsphere-%s", channel)
		artifactsTable[vSphereCloudProviderComponentName] = vSphereCloudProviderArtifacts

		bottlerocketBootstrapComponentName := fmt.Sprintf("bottlerocket-bootstrap-%s-%s", channel, releaseNumberStr)
		artifactsTable[bottlerocketBootstrapComponentName] = bottlerocketBootstrapArtifacts
	}

	return artifactsTable, nil
}

func (r *ReleaseConfig) GetEksAArtifactsData() (map[string][]Artifact, error) {
	artifactsTable := map[string][]Artifact{}
	artifacts, err := r.GetEksACliArtifacts()
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting artifact information for EKS-A CLI")
	}

	artifactsTable["eks-a-cli"] = artifacts

	return artifactsTable, nil
}

// GetURI returns an full URL for the given path
func (r *ReleaseConfig) GetURI(path string) (string, error) {
	uri, err := url.Parse(r.CDN)
	if err != nil {
		return "", err
	}
	uri.Path = path
	return uri.String(), nil
}

func (r *ReleaseConfig) GetSourceImageURI(name, repoName string, tagOptions map[string]string) string {
	var sourceImageUri string
	if r.DevRelease || r.ReleaseEnvironment == "development" {
		if name == "bottlerocket-bootstrap" {
			sourceImageUri = fmt.Sprintf("%s/%s:v%s-%s-latest",
				r.SourceContainerRegistry,
				repoName,
				tagOptions["eksDReleaseChannel"],
				tagOptions["eksDReleaseNumber"],
			)
		} else if name == "cloud-provider-vsphere" {
			sourceImageUri = fmt.Sprintf("%s/%s:%s-latest",
				r.SourceContainerRegistry,
				repoName,
				tagOptions["gitTag"],
			)
		} else if name == "kind-node" {
			sourceImageUri = fmt.Sprintf("%s/%s:%s-eks-%s-%s-latest",
				r.SourceContainerRegistry,
				repoName,
				tagOptions["kubeVersion"],
				tagOptions["eksDReleaseChannel"],
				tagOptions["eksDReleaseNumber"],
			)
		} else {
			sourceImageUri = fmt.Sprintf("%s/%s:latest",
				r.SourceContainerRegistry,
				repoName,
			)
		}
	} else if r.ReleaseEnvironment == "production" {
		if name == "bottlerocket-bootstrap" {
			sourceImageUri = fmt.Sprintf("%s/%s:v%s-%s-eks-a-%d",
				r.SourceContainerRegistry,
				repoName,
				tagOptions["eksDReleaseChannel"],
				tagOptions["eksDReleaseNumber"],
				r.BundleNumber,
			)
		} else if name == "cloud-provider-vsphere" {
			sourceImageUri = fmt.Sprintf("%s/%s:%s-eks-d-%s-eks-a-%d",
				r.SourceContainerRegistry,
				repoName,
				tagOptions["gitTag"],
				tagOptions["eksDReleaseChannel"],
				r.BundleNumber,
			)
		} else if name == "eks-anywhere-cluster-controller" {
			sourceImageUri = fmt.Sprintf("%s/%s:%s-eks-a-%d",
				r.SourceContainerRegistry,
				repoName,
				r.ReleaseVersion,
				r.BundleNumber,
			)
		} else if name == "kind-node" {
			sourceImageUri = fmt.Sprintf("%s/%s:%s-eks-d-%s-%s-eks-a-%d",
				r.SourceContainerRegistry,
				repoName,
				tagOptions["kubeVersion"],
				tagOptions["eksDReleaseChannel"],
				tagOptions["eksDReleaseNumber"],
				r.BundleNumber,
			)
		} else {
			sourceImageUri = fmt.Sprintf("%s/%s:%s-eks-a-%d",
				r.SourceContainerRegistry,
				repoName,
				tagOptions["gitTag"],
				r.BundleNumber,
			)
		}
	}

	return sourceImageUri
}

func (r *ReleaseConfig) GetReleaseImageURI(name, repoName string, tagOptions map[string]string) string {
	var releaseImageUri string
	var semVer string
	if r.DevRelease {
		semVer = r.DevReleaseUriVersion
	} else {
		semVer = fmt.Sprintf("%d", r.BundleNumber)
	}

	if name == "bottlerocket-bootstrap" {
		releaseImageUri = fmt.Sprintf("%s/%s:v%s-%s-eks-a-%s",
			r.ReleaseContainerRegistry,
			repoName,
			tagOptions["eksDReleaseChannel"],
			tagOptions["eksDReleaseNumber"],
			semVer,
		)
	} else if name == "cloud-provider-vsphere" {
		releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-d-%s-eks-a-%s",
			r.ReleaseContainerRegistry,
			repoName,
			tagOptions["gitTag"],
			tagOptions["eksDReleaseChannel"],
			semVer,
		)
	} else if name == "eks-anywhere-cluster-controller" {
		if r.DevRelease {
			releaseImageUri = fmt.Sprintf("%s/%s:v0.0.0-eks-a-%s",
				r.ReleaseContainerRegistry,
				repoName,
				semVer,
			)
		} else {
			releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-a-%s",
				r.ReleaseContainerRegistry,
				repoName,
				r.ReleaseVersion,
				semVer,
			)
		}
	} else if name == "kind-node" {
		releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-d-%s-%s-eks-a-%s",
			r.ReleaseContainerRegistry,
			repoName,
			tagOptions["kubeVersion"],
			tagOptions["eksDReleaseChannel"],
			tagOptions["eksDReleaseNumber"],
			semVer,
		)
	} else {
		releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-a-%s",
			r.ReleaseContainerRegistry,
			repoName,
			tagOptions["gitTag"],
			semVer,
		)
	}

	return releaseImageUri
}

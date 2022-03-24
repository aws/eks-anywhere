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

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/pkg/aws/ecr"
	"github.com/aws/eks-anywhere/release/pkg/aws/ecrpublic"
	"github.com/aws/eks-anywhere/release/pkg/aws/s3"
	"github.com/aws/eks-anywhere/release/pkg/clients"
	"github.com/aws/eks-anywhere/release/pkg/images"
	"github.com/aws/eks-anywhere/release/pkg/utils"
)

const SuccessIcon = "✅"

// ReleaseConfig contains metadata fields for a release
type ReleaseConfig struct {
	ReleaseVersion           string
	DevReleaseUriVersion     string
	BundleNumber             int
	CliMinVersion            string
	CliMaxVersion            string
	CliRepoUrl               string
	CliRepoSource            string
	CliRepoHead              string
	CliRepoBranchName        string
	BuildRepoUrl             string
	BuildRepoSource          string
	BuildRepoHead            string
	BuildRepoBranchName      string
	ArtifactDir              string
	SourceBucket             string
	ReleaseBucket            string
	SourceContainerRegistry  string
	ReleaseContainerRegistry string
	CDN                      string
	ReleaseNumber            int
	ReleaseDate              time.Time
	DevRelease               bool
	DryRun                   bool
	ReleaseEnvironment       string
	SourceClients            *clients.SourceClients
	ReleaseClients           *clients.ReleaseClients
	BundleArtifactsTable     map[string][]Artifact
	EksAArtifactsTable       map[string][]Artifact
}

type projectVersioner interface {
	patchVersion() (string, error)
}

// GetVersionsBundles will build the entire bundle manifest from the
// individual component bundles
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

	ciliumBundle, err := r.GetCiliumBundle()
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Cilium")
	}

	kindnetdBundle, err := r.GetKindnetdBundle()
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Kindnetd")
	}

	haproxyBundle, err := r.GetHaproxyBundle(imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Haproxy")
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

	bottlerocketAdminBundle, err := r.GetBottlerocketAdminBundle()
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Bottlerocket admin container")
	}

	var packageBundle anywherev1alpha1.PackageBundle
	var tinkerbellBundle anywherev1alpha1.TinkerbellBundle
	var snowBundle anywherev1alpha1.SnowBundle
	if r.DevRelease && r.BuildRepoBranchName == "main" {
		tinkerbellBundle, err = r.GetTinkerbellBundle(imageDigests)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for Tinkerbell infrastructure provider")
		}

		snowBundle, err = r.GetSnowBundle(imageDigests)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for Snow infrastructure provider")
		}

		packageBundle, err = r.GetPackagesBundle(imageDigests)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for Package controllers")
		}
	}

	var cloudStackBundle anywherev1alpha1.CloudStackBundle
	if r.DevRelease && r.BuildRepoBranchName == "main" {
		cloudStackBundle, err = r.GetCloudStackBundle(imageDigests)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for CloudStack infrastructure provider")
		}
	}

	eksDReleaseMap, err := readEksDReleases(r)
	if err != nil {
		return nil, err
	}

	supportedK8sVersions, err := getSupportedK8sVersions(r)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting supported Kubernetes versions for bottlerocket")
	}

	for _, release := range eksDReleaseMap.Releases {
		channel := release.Branch
		number := strconv.Itoa(release.Number)
		dev := release.Dev
		kubeVersion := release.KubeVersion
		shortKubeVersion := kubeVersion[1:strings.LastIndex(kubeVersion, ".")]

		if !utils.SliceContains(supportedK8sVersions, channel) {
			continue
		}

		eksDReleaseBundle, err := r.GetEksDReleaseBundle(channel, kubeVersion, number, imageDigests, dev)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for eks-d %s-%s release bundle", channel, number)
		}

		vsphereBundle, err := r.GetVsphereBundle(channel, imageDigests)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for vSphere infrastructure provider")
		}

		bottlerocketBootstrapBundle, err := r.GetBottlerocketBootstrapBundle(channel, number, imageDigests)
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
			CloudStack:             cloudStackBundle,
			Docker:                 dockerBundle,
			Eksa:                   eksaBundle,
			Cilium:                 ciliumBundle,
			Kindnetd:               kindnetdBundle,
			Flux:                   fluxBundle,
			PackageController:      packageBundle,
			ExternalEtcdBootstrap:  etcdadmBootstrapBundle,
			ExternalEtcdController: etcdadmControllerBundle,
			BottleRocketBootstrap:  bottlerocketBootstrapBundle,
			BottleRocketAdmin:      bottlerocketAdminBundle,
			Tinkerbell:             tinkerbellBundle,
			Haproxy:                haproxyBundle,
			Snow:                   snowBundle,
		}
		versionsBundles = append(versionsBundles, versionsBundle)
	}
	return versionsBundles, nil
}

func (r *ReleaseConfig) GenerateBundleSpec(bundles *anywherev1alpha1.Bundles, imageDigests map[string]string) error {
	fmt.Println("\n==========================================================")
	fmt.Println("               Bundles Manifest Spec Generation")
	fmt.Println("==========================================================")
	versionsBundles, err := r.GetVersionsBundles(imageDigests)
	if err != nil {
		return err
	}

	bundles.Spec.VersionsBundles = versionsBundles

	fmt.Printf("%s Successfully generated bundle manifest spec\n", SuccessIcon)
	return nil
}

func (r *ReleaseConfig) GenerateBundleArtifactsTable() (map[string][]Artifact, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("              Bundle Artifacts Table Generation")
	fmt.Println("==========================================================")

	artifactsTable := map[string][]Artifact{}
	eksAArtifactsFuncs := map[string]func() ([]Artifact, error){
		"eks-a-tools":                     r.GetEksAToolsAssets,
		"cluster-api":                     r.GetCAPIAssets,
		"cluster-api-provider-aws":        r.GetCapaAssets,
		"cluster-api-provider-docker":     r.GetDockerAssets,
		"cluster-api-provider-vsphere":    r.GetCapvAssets,
		"vsphere-csi-driver":              r.GetVsphereCsiAssets,
		"cert-manager":                    r.GetCertManagerAssets,
		"cilium":                          r.GetCiliumAssets,
		"local-path-provisioner":          r.GetLocalPathProvisionerAssets,
		"kube-rbac-proxy":                 r.GetKubeRbacProxyAssets,
		"kube-vip":                        r.GetKubeVipAssets,
		"flux":                            r.GetFluxAssets,
		"etcdadm-bootstrap-provider":      r.GetEtcdadmBootstrapAssets,
		"etcdadm-controller":              r.GetEtcdadmControllerAssets,
		"cluster-controller":              r.GetClusterControllerAssets,
		"kindnetd":                        r.GetKindnetdAssets,
		"etcdadm":                         r.GetEtcdadmAssets,
		"cri-tools":                       r.GetCriToolsAssets,
		"diagnostic-collector":            r.GetDiagnosticCollectorAssets,
		"haproxy":                         r.GetHaproxyAssets,
	}

	if r.DevRelease && r.BuildRepoBranchName == "main" {
		eksAArtifactsFuncs["cluster-api-provider-tinkerbell"] = r.GetCaptAssets
		eksAArtifactsFuncs["cluster-api-provider-cloudstack"] = r.GetCapcAssets
		eksAArtifactsFuncs["tink"] = r.GetTinkAssets
		eksAArtifactsFuncs["hegel"] = r.GetHegelAssets
		eksAArtifactsFuncs["cfssl"] = r.GetCfsslAssets
		eksAArtifactsFuncs["pbnj"] = r.GetPbnjAssets
		eksAArtifactsFuncs["boots"] = r.GetBootsAssets
		eksAArtifactsFuncs["hub"] = r.GetHubAssets
		eksAArtifactsFuncs["cluster-api-provider-aws-snow"] = r.GetCapasAssets
		eksAArtifactsFuncs["hook"] = r.GetHookAssets
		eksAArtifactsFuncs["eks-anywhere-packages"] = r.GetPackagesAssets
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

	supportedK8sVersions, err := getSupportedK8sVersions(r)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting supported Kubernetes versions for bottlerocket")
	}

	for _, release := range eksDReleaseMap.Releases {
		channel := release.Branch
		number := strconv.Itoa(release.Number)
		kubeVersion := release.KubeVersion

		if !utils.SliceContains(supportedK8sVersions, channel) {
			continue
		}

		eksDChannelArtifacts, err := r.GetEksDChannelAssets(channel, kubeVersion, number)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting artifact information for %s", channel)
		}

		vSphereCloudProviderArtifacts, err := r.GetVsphereCloudProviderAssets(channel)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting artifact information for %s", channel)
		}

		bottlerocketBootstrapArtifacts, err := r.GetBottlerocketBootstrapAssets(channel, number)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting artifact information for %s", channel)
		}

		eksDComponentName := fmt.Sprintf("eks-d-%s", channel)
		artifactsTable[eksDComponentName] = eksDChannelArtifacts

		vSphereCloudProviderComponentName := fmt.Sprintf("cloud-provider-vsphere-%s", channel)
		artifactsTable[vSphereCloudProviderComponentName] = vSphereCloudProviderArtifacts

		bottlerocketBootstrapComponentName := fmt.Sprintf("bottlerocket-bootstrap-%s-%s", channel, number)
		artifactsTable[bottlerocketBootstrapComponentName] = bottlerocketBootstrapArtifacts
	}

	fmt.Printf("%s Successfully generated bundle artifacts table\n", SuccessIcon)

	return artifactsTable, nil
}

func (r *ReleaseConfig) GenerateEksAArtifactsTable() (map[string][]Artifact, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("                 EKS-A Artifacts Table Generation")
	fmt.Println("==========================================================")

	artifactsTable := map[string][]Artifact{}
	artifacts, err := r.GetEksACliArtifacts()
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting artifact information for EKS-A CLI")
	}

	artifactsTable["eks-a-cli"] = artifacts

	fmt.Printf("%s Successfully generated EKS-A artifacts table\n", SuccessIcon)

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

func (r *ReleaseConfig) GetSourceImageURI(name, repoName string, tagOptions map[string]string) (string, string, error) {
	var sourceImageUri string
	sourcedFromBranch := r.BuildRepoBranchName
	if r.DevRelease || r.ReleaseEnvironment == "development" {
		latestTag := getLatestUploadDestination(r.BuildRepoBranchName)
		if name == "bottlerocket-bootstrap" {
			sourceImageUri = fmt.Sprintf("%s/%s:v%s-%s-%s",
				r.SourceContainerRegistry,
				repoName,
				tagOptions["eksDReleaseChannel"],
				tagOptions["eksDReleaseNumber"],
				latestTag,
			)
		} else if name == "cloud-provider-vsphere" {
			sourceImageUri = fmt.Sprintf("%s/%s:%s-%s",
				r.SourceContainerRegistry,
				repoName,
				tagOptions["gitTag"],
				latestTag,
			)
		} else if name == "kind-node" {
			sourceImageUri = fmt.Sprintf("%s/%s:%s-eks-%s-%s-%s",
				r.SourceContainerRegistry,
				repoName,
				tagOptions["kubeVersion"],
				tagOptions["eksDReleaseChannel"],
				tagOptions["eksDReleaseNumber"],
				latestTag,
			)
		} else {
			sourceImageUri = fmt.Sprintf("%s/%s:%s",
				r.SourceContainerRegistry,
				repoName,
				latestTag,
			)
		}
		if !r.DryRun {
			sourceEcrAuthConfig := r.SourceClients.ECR.AuthConfig
			err := images.PollForExistence(r.DevRelease, sourceEcrAuthConfig, sourceImageUri, r.SourceContainerRegistry, r.ReleaseEnvironment, r.BuildRepoBranchName)
			if err != nil {
				if r.BuildRepoBranchName != "main" {
					fmt.Printf("Tag corresponding to %s branch not found for %s image. Using image artifact from main\n", r.BuildRepoBranchName, repoName)
					var gitTagFromMain string
					if name == "bottlerocket-bootstrap" {
						gitTagFromMain = "non-existent"
					} else {
						gitTagFromMain, err = r.readGitTag(tagOptions["projectPath"], "main")
						if err != nil {
							return "", "", errors.Cause(err)
						}
					}
					sourceImageUri = strings.NewReplacer(r.BuildRepoBranchName, "latest", tagOptions["gitTag"], gitTagFromMain).Replace(sourceImageUri)
					sourcedFromBranch = "main"
				} else {
					return "", "", errors.Cause(err)
				}
			}
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
		} else if name == "eks-anywhere-diagnostic-collector" {
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

	return sourceImageUri, sourcedFromBranch, nil
}

func (r *ReleaseConfig) GetReleaseImageURI(name, repoName string, tagOptions map[string]string) (string, error) {
	var releaseImageUri string

	if name == "bottlerocket-bootstrap" {
		releaseImageUri = fmt.Sprintf("%s/%s:v%s-%s-eks-a",
			r.ReleaseContainerRegistry,
			repoName,
			tagOptions["eksDReleaseChannel"],
			tagOptions["eksDReleaseNumber"],
		)
	} else if name == "cloud-provider-vsphere" {
		releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-d-%s-eks-a",
			r.ReleaseContainerRegistry,
			repoName,
			tagOptions["gitTag"],
			tagOptions["eksDReleaseChannel"],
		)
	} else if name == "eks-anywhere-cluster-controller" {
		if r.DevRelease {
			releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-a",
				r.ReleaseContainerRegistry,
				repoName,
				tagOptions["gitTag"],
			)
		} else {
			releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-a",
				r.ReleaseContainerRegistry,
				repoName,
				r.ReleaseVersion,
			)
		}
	} else if name == "eks-anywhere-diagnostic-collector" {
		if r.DevRelease {
			releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-a",
				r.ReleaseContainerRegistry,
				repoName,
				tagOptions["gitTag"],
			)
		} else {
			releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-a",
				r.ReleaseContainerRegistry,
				repoName,
				r.ReleaseVersion,
			)
		}
	} else if name == "kind-node" {
		releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-d-%s-%s-eks-a",
			r.ReleaseContainerRegistry,
			repoName,
			tagOptions["kubeVersion"],
			tagOptions["eksDReleaseChannel"],
			tagOptions["eksDReleaseNumber"],
		)
	} else {
		releaseImageUri = fmt.Sprintf("%s/%s:%s-eks-a",
			r.ReleaseContainerRegistry,
			repoName,
			tagOptions["gitTag"],
		)
	}

	var semver string
	if r.DevRelease {
		currentSourceImageUri, _, err := r.GetSourceImageURI(name, repoName, tagOptions)
		if err != nil {
			return "", errors.Cause(err)
		}

		previousReleaseImageSemver, err := r.GetPreviousReleaseImageSemver(releaseImageUri)
		if err != nil {
			return "", errors.Cause(err)
		}
		if previousReleaseImageSemver == "" {
			semver = r.DevReleaseUriVersion
		} else {
			fmt.Printf("Previous release image semver for %s image: %s\n", repoName, previousReleaseImageSemver)
			previousReleaseImageUri := fmt.Sprintf("%s-%s", releaseImageUri, previousReleaseImageSemver)

			sameDigest, err := r.CompareHashWithPreviousBundle(currentSourceImageUri, previousReleaseImageUri)
			if err != nil {
				return "", errors.Cause(err)
			}
			if sameDigest {
				semver = previousReleaseImageSemver
				fmt.Printf("Image digest for %s image has not changed, tagging with previous dev release semver: %s\n", repoName, semver)
			} else {
				newSemver, err := generateNewDevReleaseVersion(previousReleaseImageSemver, "vDev", r.BuildRepoBranchName)
				if err != nil {
					return "", errors.Cause(err)
				}
				semver = strings.ReplaceAll(newSemver, "+", "-")
				fmt.Printf("Image digest for %s image has changed, tagging with new dev release semver: %s\n", repoName, semver)
			}
		}
	} else {
		semver = fmt.Sprintf("%d", r.BundleNumber)
	}

	releaseImageUri = fmt.Sprintf("%s-%s", releaseImageUri, semver)

	return releaseImageUri, nil
}

func (r *ReleaseConfig) CompareHashWithPreviousBundle(currentSourceImageUri, previousReleaseImageUri string) (bool, error) {
	if r.DryRun {
		return false, nil
	}
	fmt.Printf("Comparing digests for [%s] and [%s]\n", currentSourceImageUri, previousReleaseImageUri)
	currentSourceImageUriDigest, err := ecr.GetImageDigest(currentSourceImageUri, r.SourceContainerRegistry, r.SourceClients.ECR.EcrClient)
	if err != nil {
		return false, errors.Cause(err)
	}

	previousReleaseImageUriDigest, err := ecrpublic.GetImageDigest(previousReleaseImageUri, r.ReleaseContainerRegistry, r.ReleaseClients.ECRPublic.Client)
	if err != nil {
		return false, errors.Cause(err)
	}

	return currentSourceImageUriDigest == previousReleaseImageUriDigest, nil
}

func (r *ReleaseConfig) GetPreviousReleaseImageSemver(releaseImageUri string) (string, error) {
	var semver string
	if r.DryRun {
		semver = "v0.0.0-dev-build.0"
	} else {
		bundles := &anywherev1alpha1.Bundles{}
		bundleReleaseManifestKey := utils.GetManifestFilepaths(r.DevRelease, r.BundleNumber, anywherev1alpha1.BundlesKind, r.BuildRepoBranchName)
		bundleManifestUrl := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", r.ReleaseBucket, bundleReleaseManifestKey)
		if s3.KeyExists(r.ReleaseBucket, bundleReleaseManifestKey) {
			contents, err := ReadHttpFile(bundleManifestUrl)
			if err != nil {
				return "", fmt.Errorf("Error reading bundle manifest from S3: %v", err)
			}

			if err = yaml.Unmarshal(contents, bundles); err != nil {
				return "", fmt.Errorf("Error unmarshaling bundles manifest from [%s]: %v", bundleManifestUrl, err)
			}

			for _, versionedBundle := range bundles.Spec.VersionsBundles {
				vbImages := versionedBundle.Images()
				for _, image := range vbImages {
					if strings.Contains(image.URI, releaseImageUri) {
						imageUri := image.URI
						var differential int
						if r.BuildRepoBranchName == "main" {
							differential = 1
						} else {
							differential = 2
						}
						numDashes := strings.Count(imageUri, "-")
						splitIndex := numDashes - strings.Count(r.BuildRepoBranchName, "-") - differential
						imageUriSplit := strings.SplitAfterN(imageUri, "-", splitIndex)
						semver = imageUriSplit[len(imageUriSplit)-1]
					}
				}
			}
		}
	}
	return semver, nil
}

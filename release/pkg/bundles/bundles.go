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
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/pkg/constants"
	"github.com/aws/eks-anywhere/release/pkg/filereader"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
	sliceutils "github.com/aws/eks-anywhere/release/pkg/util/slices"
)

func NewBundlesName(r *releasetypes.ReleaseConfig) string {
	return fmt.Sprintf("bundles-%d", r.BundleNumber)
}

func NewBaseBundles(r *releasetypes.ReleaseConfig) *anywherev1alpha1.Bundles {
	return &anywherev1alpha1.Bundles{
		TypeMeta: metav1.TypeMeta{
			APIVersion: anywherev1alpha1.GroupVersion.String(),
			Kind:       constants.BundlesKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              NewBundlesName(r),
			CreationTimestamp: metav1.Time{Time: r.ReleaseDate},
		},
		Spec: anywherev1alpha1.BundlesSpec{
			Number: r.BundleNumber,
		},
	}
}

// GetVersionsBundles will build the entire bundle manifest from the
// individual component bundles
func GetVersionsBundles(r *releasetypes.ReleaseConfig, imageDigests map[string]string) ([]anywherev1alpha1.VersionsBundle, error) {
	versionsBundles := []anywherev1alpha1.VersionsBundle{}

	certManagerBundle, err := GetCertManagerBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for cert-manager")
	}

	coreClusterApiBundle, err := GetCoreClusterAPIBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for core cluster-api")
	}

	kubeadmBootstrapBundle, err := GetKubeadmBootstrapBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for cluster-api kubeadm-bootstrap")
	}

	kubeadmControlPlaneBundle, err := GetKubeadmControlPlaneBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for cluster-api kubeadm-control-plane")
	}

	awsBundle, err := GetAwsBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for AWS infrastructure provider")
	}

	dockerBundle, err := GetDockerBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Docker infrastructure provider")
	}

	eksaBundle, err := GetEksaBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for eks-a tools component")
	}

	ciliumBundle, err := GetCiliumBundle(r)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Cilium")
	}

	kindnetdBundle, err := GetKindnetdBundle(r)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Kindnetd")
	}

	haproxyBundle, err := GetHaproxyBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Haproxy")
	}

	fluxBundle, err := GetFluxBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Flux controllers")
	}

	etcdadmBootstrapBundle, err := GetEtcdadmBootstrapBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for external Etcdadm bootstrap")
	}

	etcdadmControllerBundle, err := GetEtcdadmControllerBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for external Etcdadm controller")
	}

	bottlerocketAdminBundle, err := GetBottlerocketAdminBundle(r)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Bottlerocket admin container")
	}

	packageBundle, err := GetPackagesBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Package controllers")
	}

	tinkerbellBundle, err := GetTinkerbellBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for Tinkerbell infrastructure provider")
	}

	cloudStackBundle, err := GetCloudStackBundle(r, imageDigests)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting bundle for CloudStack infrastructure provider")
	}

	var snowBundle anywherev1alpha1.SnowBundle
	if r.DevRelease && r.BuildRepoBranchName == "main" {
		snowBundle, err = GetSnowBundle(r, imageDigests)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for Snow infrastructure provider")
		}
	}

	eksDReleaseMap, err := filereader.ReadEksDReleases(r)
	if err != nil {
		return nil, err
	}

	supportedK8sVersions, err := filereader.GetSupportedK8sVersions(r)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting supported Kubernetes versions for bottlerocket")
	}

	for _, release := range eksDReleaseMap.Releases {
		channel := release.Branch
		number := strconv.Itoa(release.Number)
		dev := release.Dev
		kubeVersion := release.KubeVersion
		shortKubeVersion := kubeVersion[1:strings.LastIndex(kubeVersion, ".")]

		if !sliceutils.SliceContains(supportedK8sVersions, channel) {
			continue
		}

		eksDReleaseBundle, err := GetEksDReleaseBundle(r, channel, kubeVersion, number, imageDigests, dev)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for eks-d %s-%s release bundle", channel, number)
		}

		vsphereBundle, err := GetVsphereBundle(r, channel, imageDigests)
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting bundle for vSphere infrastructure provider")
		}

		bottlerocketBootstrapBundle, err := GetBottlerocketBootstrapBundle(r, channel, number, imageDigests)
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

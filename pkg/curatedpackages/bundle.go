package curatedpackages

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	ImageRepositoryName = "eks-anywhere-packages-bundles"
)

type Reader interface {
	ReadBundlesForVersion(eksaVersion string) (*releasev1.Bundles, error)
}

type BundleRegistry interface {
	GetRegistryBaseRef(ctx context.Context) (string, error)
}

func GetPackageBundleRef(vb releasev1.VersionsBundle) (string, error) {
	packageController := vb.PackageController
	// Use package controller registry to fetch packageBundles.
	// Format of controller image is: <uri>/<env_type>/<repository_name>
	controllerImage := strings.Split(packageController.Controller.Image(), "/")
	major, minor, err := parseKubeVersion(vb.KubeVersion)
	if err != nil {
		logger.MarkFail("unable to parse kubeversion", "error", err)
		return "", fmt.Errorf("unable to parse kubeversion %s %v", vb.KubeVersion, err)
	}
	latestBundle := fmt.Sprintf("v%s-%s-%s", major, minor, "latest")
	registryBaseRef := fmt.Sprintf("%s/%s/%s:%s", controllerImage[0], controllerImage[1], "eks-anywhere-packages-bundles", latestBundle)
	return registryBaseRef, nil
}

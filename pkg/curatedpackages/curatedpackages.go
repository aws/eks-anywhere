package curatedpackages

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"

	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	license = `The EKS Anywhere package controller and the EKS Anywhere Curated Packages
(referred to as “features”) are provided as “preview features” subject to the AWS Service Terms,
(including Section 2 (Betas and Previews)) of the same. During the EKS Anywhere Curated Packages Public Preview,
the AWS Service Terms are extended to provide customers access to these features free of charge.
These features will be subject to a service charge and fee structure at ”General Availability“ of the features.`
	width     = 112
	mediaType = "application/vnd.oci.image.manifest.v1+json"
)

func NewRegistry(deps *dependencies.Dependencies, registryName, kubeVersion, username, password string) (BundleRegistry, error) {
	if registryName != "" {
		registry := NewCustomRegistry(
			deps.Helm,
			registryName,
			username,
			password,
		)
		return registry, nil
	}
	defaultRegistry := NewDefaultRegistry(
		deps.ManifestReader,
		kubeVersion,
		version.Get(),
	)
	return defaultRegistry, nil
}

func CreateBundleManager(kubeVersion string) bundle.Manager {
	major, minor, err := parseKubeVersion(kubeVersion)
	if err != nil {
		return nil
	}
	log := logr.Discard()
	k := NewKubeVersion(major, minor)
	discovery := NewDiscovery(k)
	puller := artifacts.NewRegistryPuller()
	return bundle.NewBundleManager(log, discovery, puller)
}

func parseKubeVersion(kubeVersion string) (string, string, error) {
	versionSplit := strings.Split(kubeVersion, ".")
	if len(versionSplit) != 2 {
		return "", "", fmt.Errorf("invalid kubeversion %s", kubeVersion)
	}
	major, minor := versionSplit[0], versionSplit[1]
	return major, minor, nil
}

func GetVersionBundle(reader Reader, eksaVersion, kubeVersion string) (*releasev1.VersionsBundle, error) {
	b, err := reader.ReadBundlesForVersion(eksaVersion)
	if err != nil {
		return nil, err
	}
	versionsBundle := bundles.VersionsBundleForKubernetesVersion(b, kubeVersion)
	if versionsBundle == nil {
		return nil, fmt.Errorf("kubernetes version %s is not supported by bundles manifest %d", kubeVersion, b.Spec.Number)
	}
	return versionsBundle, nil
}

func NewDependenciesForPackages(ctx context.Context, paths ...string) (*dependencies.Dependencies, error) {
	return dependencies.NewFactory().
		WithExecutableMountDirs(paths...).
		WithExecutableBuilder().
		WithManifestReader().
		WithKubectl().
		WithHelm().
		Build(ctx)
}

func PrintLicense() {
	// Currently, use the width of the longest line to repeat the dashes
	// Sample Output
	//----------------------------------------------------------------------------------------------------------------
	//The EKS Anywhere package controller and the EKS Anywhere Curated Packages
	//(referred to as “features”) are provided as “preview features” subject to the AWS Service Terms,
	//(including Section 2 (Betas and Previews)) of the same. During the EKS Anywhere Curated Packages Public Preview,
	//the AWS Service Terms are extended to provide customers access to these features free of charge.
	//These features will be subject to a service charge and fee structure at ”General Availability“ of the features.
	//----------------------------------------------------------------------------------------------------------------
	fmt.Println(strings.Repeat("-", width))
	fmt.Println(license)
	fmt.Println(strings.Repeat("-", width))
}

func Pull(ctx context.Context, art string) ([]byte, error) {
	puller := artifacts.NewRegistryPuller()

	data, err := puller.Pull(ctx, art)
	if err != nil {
		return nil, fmt.Errorf("unable to pull artifacts %v", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("latest package bundle artifact is empty")
	}

	return data, nil
}

func Push(ctx context.Context, ref, fileName string, fileContent []byte) error {
	registry, err := content.NewRegistry(content.RegistryOptions{})
	if err != nil {
		return fmt.Errorf("creating registry: %w", err)
	}
	memoryStore := content.NewMemory()
	desc, err := memoryStore.Add(fileName, mediaType, fileContent)
	if err != nil {
		return err
	}

	manifest, manifestDesc, config, configDesc, err := content.GenerateManifestAndConfig(nil, nil, desc)
	if err != nil {
		return err
	}

	memoryStore.Set(configDesc, config)
	err = memoryStore.StoreManifest(ref, manifestDesc, manifest)
	if err != nil {
		return err
	}
	fmt.Printf("Pushing %s to %s...\n", fileName, ref)
	desc, err = oras.Copy(ctx, memoryStore, ref, registry, "")
	if err != nil {
		return err
	}
	fmt.Printf("Pushed to %s with digest %s\n", ref, desc.Digest)
	return nil
}

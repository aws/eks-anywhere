package curatedpackages

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"

	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	license = `The EKS Anywhere package controller and the EKS Anywhere Curated Packages
(referred to as “features”) are provided as “preview features” subject to the AWS Service Terms,
(including Section 2 (Betas and Previews)) of the same. During the EKS Anywhere Curated Packages Public Preview,
the AWS Service Terms are extended to provide customers access to these features free of charge.
These features will be subject to a service charge and fee structure at ”General Availability“ of the features.`
	certManagerDoesNotExistMsg = `Curated packages cannot be installed as cert-manager is not present in the cluster. This is most likely caused
by an action to install Curated packages at a workload cluster. Refer to
https://anywhere.eks.amazonaws.com/docs/tasks/packages/ for how to resolve this issue.`
	width = 112
)

var userMsgSeparator = strings.Repeat("-", width)

func CreateBundleManager(kubeVersion string) bundle.RegistryClient {
	puller := artifacts.NewRegistryPuller()
	return bundle.NewRegistryClient(puller)
}

func parseKubeVersion(kubeVersion string) (string, string, error) {
	versionSplit := strings.Split(kubeVersion, ".")
	if len(versionSplit) != 2 {
		return "", "", fmt.Errorf("invalid kubeversion %s", kubeVersion)
	}
	major, minor := versionSplit[0], versionSplit[1]
	return major, minor, nil
}

func GetVersionBundle(reader Reader, eksaVersion string, spec *v1alpha1.Cluster) (*releasev1.VersionsBundle, error) {
	b, err := reader.ReadBundlesForVersion(eksaVersion)
	if err != nil {
		return nil, err
	}
	versionsBundle, err := cluster.GetVersionsBundle(spec, b)
	if err != nil {
		return nil, err
	}
	return versionsBundle, nil
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
	fmt.Println(userMsgSeparator)
	fmt.Println(license)
	fmt.Println(userMsgSeparator)
}

func PrintCertManagerDoesNotExistMsg() {
	// Currently, use the width of the longest line to repeat the dashes
	// Sample Output
	//----------------------------------------------------------------------------------------------------------------
	//Curated packages cannot be installed as cert-manager is not present in the cluster. This is most likely caused
	//by an action to install Curated packages at a workload cluster. Refer to
	//https://anywhere.eks.amazonaws.com/docs/tasks/packages/ for how to resolve this issue.
	//----------------------------------------------------------------------------------------------------------------
	fmt.Println(userMsgSeparator)
	fmt.Println(certManagerDoesNotExistMsg)
	fmt.Println(userMsgSeparator)
}

func VerifyCertManagerExists(ctx context.Context, kubectl KubectlRunner, kubeConfig string) error {
	// Note although we passed in a namespace parameter in the kubectl command, the GetResource command will be
	// performed in all namespaces since CRDs are not bounded by namespaces.
	certManagerExists, err := kubectl.GetResource(ctx, "crd", "certificates.cert-manager.io", kubeConfig,
		constants.CertManagerNamespace)
	if err != nil {
		return err
	}

	// If cert-manager does not exist, instruct users to follow instructions in
	// PrintCertManagerDoesNotExistMsg to install packages manually.
	if !certManagerExists {
		PrintCertManagerDoesNotExistMsg()
		return errors.New("cert-manager is not present in the cluster")
	}

	return nil
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
	desc, err := memoryStore.Add("bundle.yaml", "", fileContent)
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
	logger.Info(fmt.Sprintf("Pushing %s to %s...", fileName, ref))
	desc, err = oras.Copy(ctx, memoryStore, ref, registry, "")
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Pushed to %s with digest %s", ref, desc.Digest))
	return nil
}

func GetRegistry(uri string) string {
	lastInd := strings.LastIndex(uri, "/")
	if lastInd == -1 {
		return uri
	}
	return uri[:lastInd]
}

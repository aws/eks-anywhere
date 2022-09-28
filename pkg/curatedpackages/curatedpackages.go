package curatedpackages

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	license = `The Amazon EKS Anywhere Curated Packages are only available to customers with the 
Amazon EKS Anywhere Enterprise Subscription`
	certManagerDoesNotExistMsg = `Curated packages cannot be installed as cert-manager is not present in the cluster.
This is most likely caused by an action to install curated packages at a workload 
cluster. Refer to https://anywhere.eks.amazonaws.com/docs/tasks/troubleshoot/packages/ 
for how to resolve this issue.`
	width = 86
)

var userMsgSeparator = strings.Repeat("-", width)

func CreateBundleManager() bundle.RegistryClient {
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
	//-------------------------------------------------------------------------------------
	//The Amazon EKS Anywhere Curated Packages are only available to customers with the
	//Amazon EKS Anywhere Enterprise Subscription
	//-------------------------------------------------------------------------------------
	fmt.Println(userMsgSeparator)
	fmt.Println(license)
	fmt.Println(userMsgSeparator)
}

func PrintCertManagerDoesNotExistMsg() {
	// Currently, use the width of the longest line to repeat the dashes
	// Sample Output
	//--------------------------------------------------------------------------------------
	//Curated packages cannot be installed as cert-manager is not present in the cluster.
	//This is most likely caused by an action to install curated packages at a workload
	//cluster. Refer to https://anywhere.eks.amazonaws.com/docs/tasks/troubleshoot/packages/
	//for how to resolve this issue.
	//--------------------------------------------------------------------------------------
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

func GetRegistry(uri string) string {
	lastInd := strings.LastIndex(uri, "/")
	if lastInd == -1 {
		return uri
	}
	return uri[:lastInd]
}

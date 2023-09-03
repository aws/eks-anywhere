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
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	license = `The Amazon EKS Anywhere Curated Packages are only available to customers with the 
Amazon EKS Anywhere Enterprise Subscription`
	width = 86
)

var userMsgSeparator = strings.Repeat("-", width)

// CreateBundleManager builds a new bundle Manager.
func CreateBundleManager(log logr.Logger) bundle.RegistryClient {
	puller := artifacts.NewRegistryPuller(log)
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
	versionsBundle, err := cluster.GetVersionsBundle(spec.Spec.KubernetesVersion, b)
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

// PullLatestBundle reads the contents of the artifact using the latest bundle.
func PullLatestBundle(ctx context.Context, log logr.Logger, artifact string) ([]byte, error) {
	puller := artifacts.NewRegistryPuller(log)

	data, err := puller.Pull(ctx, artifact, "")
	if err != nil {
		return nil, fmt.Errorf("unable to pull artifacts %v", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("latest package bundle artifact is empty")
	}

	return data, nil
}

func PushBundle(ctx context.Context, ref, fileName string, fileContent []byte) error {
	registry, err := content.NewRegistry(content.RegistryOptions{Insecure: ctx.Value(types.InsecureRegistry).(bool)})
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

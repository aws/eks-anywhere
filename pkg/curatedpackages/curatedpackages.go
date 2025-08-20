package curatedpackages

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"

	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
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

	var data []byte
	err := retrier.Retry(5, 200*time.Millisecond, func() error {
		d, err := puller.Pull(ctx, artifact, "")
		if err != nil {
			return err
		}
		data = d
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to pull artifacts %v", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("latest package bundle artifact is empty")
	}

	return data, nil
}

func PushBundle(ctx context.Context, ref, fileName string, fileContent []byte) error {
	// Parse the reference to get repository
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return fmt.Errorf("parsing reference: %w", err)
	}

	// Configure repository with insecure option if needed
	if insecure, ok := ctx.Value(types.InsecureRegistry).(bool); ok && insecure {
		repo.PlainHTTP = true
	}

	// Create memory store
	memoryStore := memory.New()

	// Create a descriptor with proper digest calculation
	mediaType := "application/vnd.oci.image.config.v1+json"

	// Calculate the digest
	dgst := digest.FromBytes(fileContent)

	// Create descriptor using OCI spec
	desc := ocispec.Descriptor{
		MediaType: mediaType,
		Size:      int64(len(fileContent)),
		Digest:    dgst,
		Annotations: map[string]string{
			"org.opencontainers.image.title": "bundle.yaml",
		},
	}

	// Store the content in memory
	err = memoryStore.Push(ctx, desc, bytes.NewReader(fileContent))
	if err != nil {
		return fmt.Errorf("storing content: %w", err)
	}

	// Push the content to the registry
	logger.Info(fmt.Sprintf("Pushing %s to %s...", fileName, ref))
	err = repo.Push(ctx, desc, bytes.NewReader(fileContent))
	if err != nil {
		return fmt.Errorf("pushing to registry: %w", err)
	}

	logger.Info(fmt.Sprintf("Pushed to %s with digest %s", ref, desc.Digest.String()))
	return nil
}

func GetRegistry(uri string) string {
	lastInd := strings.LastIndex(uri, "/")
	if lastInd == -1 {
		return uri
	}
	return uri[:lastInd]
}

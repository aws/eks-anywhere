package cluster

import (
	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// FileSpecBuilder allows to build [Spec] by reading from files.
type FileSpecBuilder struct {
	// TODO(g-gaston): this is very much a CLI thing. Move to `pkg/cli` when available.
	reader              manifests.FileReader
	cliVersion          version.Info
	releasesManifestURL string
	bundlesManifestURL  string
}

// FileSpecBuilderOpt allows to configure [FileSpecBuilder].
type FileSpecBuilderOpt func(*FileSpecBuilder)

// WithReleasesManifest configures the URL to read the Releases manifest.
func WithReleasesManifest(url string) FileSpecBuilderOpt {
	return func(b *FileSpecBuilder) {
		b.releasesManifestURL = url
	}
}

// WithOverrideBundlesManifest configures the URL to read the Bundles manifest.
// This overrides the Bundles declared in the Releases so reading the Releases
// manifest is skipped.
func WithOverrideBundlesManifest(url string) FileSpecBuilderOpt {
	return func(b *FileSpecBuilder) {
		b.bundlesManifestURL = url
	}
}

// NewFileSpecBuilder builds a new [FileSpecBuilder].
// cliVersion is used to chose the right Bundles from the the Release manifest.
func NewFileSpecBuilder(reader manifests.FileReader, cliVersion version.Info, opts ...FileSpecBuilderOpt) FileSpecBuilder {
	f := &FileSpecBuilder{
		cliVersion: cliVersion,
		reader:     reader,
	}

	for _, opt := range opts {
		opt(f)
	}

	return *f
}

// Build constructs a new [Spec] by reading the cluster config in yaml from a file and
// Releases, Bundles and EKS-D manifests from the configured URLs.
func (b FileSpecBuilder) Build(clusterConfigURL string) (*Spec, error) {
	config, err := b.getConfig(clusterConfigURL)
	if err != nil {
		return nil, err
	}

	mReader := b.createManifestReader()
	bundlesManifest, err := b.getBundles(mReader)
	if err != nil {
		return nil, errors.Wrapf(err, "getting Bundles file")
	}
	bundle := bundlesManifest.Spec.VersionsBundles[0]
	var infraProviderName, infraProviderVersion string
	switch config.Cluster.Spec.DatacenterRef.Kind {
	case v1alpha1.CloudStackDatacenterKind:
		infraProviderName = "Cluster API Provider CloudStack"
		infraProviderVersion = bundle.CloudStack.Version
	case v1alpha1.DockerDatacenterKind:
		infraProviderName = "Cluster API Provider Docker"
		infraProviderVersion = bundle.Docker.Version
	case v1alpha1.NutanixDatacenterKind:
		infraProviderName = "Cluster API Provider Nutanix"
		infraProviderVersion = bundle.Nutanix.Version
	case v1alpha1.SnowDatacenterKind:
		infraProviderName = "Cluster API Provider AWS Snow"
		infraProviderVersion = bundle.Snow.Version
	case v1alpha1.TinkerbellDatacenterKind:
		infraProviderName = "Cluster API Provider Tinkerbell"
		infraProviderVersion = bundle.Tinkerbell.Version
	case v1alpha1.VSphereDatacenterKind:
		infraProviderName = "Cluster API Provider VSphere"
		infraProviderVersion = bundle.VSphere.Version
	}
	logger.V(4).Info(
		"Using CAPI provider versions",
		"Core Cluster API", bundle.ClusterAPI.Version,
		"Kubeadm Bootstrap", bundle.Bootstrap.Version,
		"Kubeadm Control Plane", bundle.ControlPlane.Version,
		"External etcd Bootstrap", bundle.ExternalEtcdBootstrap.Version,
		"External etcd Controller", bundle.ExternalEtcdController.Version,
		infraProviderName, infraProviderVersion,
	)
	bundlesManifest.Namespace = constants.EksaSystemNamespace

	configManager, err := NewDefaultConfigManager()
	if err != nil {
		return nil, err
	}

	config.Cluster.Spec.BundlesRef = nil

	if err = configManager.SetDefaults(config); err != nil {
		return nil, err
	}

	eksdReleases, err := getAllEksdReleases(config.Cluster, bundlesManifest, b.reader)
	if err != nil {
		return nil, err
	}

	release, err := b.getEksaRelease(mReader)
	if err != nil {
		return nil, err
	}

	releaseVersion := v1alpha1.EksaVersion(release.Version)
	config.Cluster.Spec.EksaVersion = &releaseVersion
	eksaRelease := buildEKSARelease(release, bundlesManifest)

	return NewSpec(config, bundlesManifest, eksdReleases, eksaRelease)
}

func getAllEksdReleases(cluster *v1alpha1.Cluster, bundlesManifest *releasev1.Bundles, reader bundles.Reader) ([]eksdv1.Release, error) {
	versions := cluster.KubernetesVersions()
	m := make([]eksdv1.Release, 0, len(versions))
	for _, version := range versions {
		eksd, err := getEksdReleases(version, bundlesManifest, reader)
		if err != nil {
			return nil, err
		}
		m = append(m, *eksd)
	}
	return m, nil
}

func getEksdReleases(version v1alpha1.KubernetesVersion, bundlesManifest *releasev1.Bundles, reader bundles.Reader) (*eksdv1.Release, error) {
	versionsBundle, err := GetVersionsBundle(version, bundlesManifest)
	if err != nil {
		return nil, err
	}

	eksd, err := bundles.ReadEKSD(reader, *versionsBundle)
	if err != nil {
		return nil, err
	}

	return eksd, nil
}

func (b FileSpecBuilder) getConfig(clusterConfigURL string) (*Config, error) {
	yaml, err := b.reader.ReadFile(clusterConfigURL)
	if err != nil {
		return nil, errors.Wrapf(err, "reading cluster config file")
	}

	return ParseConfig(yaml)
}

func (b FileSpecBuilder) createManifestReader() *manifests.Reader {
	var opts []manifests.ReaderOpt
	if b.releasesManifestURL != "" {
		opts = append(opts, manifests.WithReleasesManifest(b.releasesManifestURL))
	}
	return manifests.NewReader(b.reader, opts...)
}

func (b FileSpecBuilder) getBundles(manifestReader *manifests.Reader) (*releasev1.Bundles, error) {
	bundlesURL := b.bundlesManifestURL
	if bundlesURL == "" {
		return manifestReader.ReadBundlesForVersion(b.cliVersion.GitVersion)
	}

	return bundles.Read(b.reader, bundlesURL)
}

func (b FileSpecBuilder) getEksaRelease(mReader *manifests.Reader) (*releasev1.EksARelease, error) {
	if b.bundlesManifestURL == "" {
		// this shouldn't return an error at this point due to getBundles performing similar operations prior to this call
		release, err := mReader.ReadReleaseForVersion(b.cliVersion.GitVersion)
		if err != nil {
			return nil, err
		}
		return release, nil
	}

	// When using bundles-override or a custom bundle, a fake EksaRelease can be used since using a custom bundle
	// is like creating a new EKS-A version.
	return &releasev1.EksARelease{
		Version: b.cliVersion.GitVersion,
	}, nil
}

func buildEKSARelease(release *releasev1.EksARelease, bundle *releasev1.Bundles) *releasev1.EKSARelease {
	eksaRelease := &releasev1.EKSARelease{
		TypeMeta: v1.TypeMeta{
			Kind:       releasev1.EKSAReleaseKind,
			APIVersion: releasev1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      releasev1.GenerateEKSAReleaseName(release.Version),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: releasev1.EKSAReleaseSpec{
			ReleaseDate:       release.Date,
			Version:           release.Version,
			GitCommit:         release.GitCommit,
			BundleManifestURL: release.BundleManifestUrl,
			BundlesRef: releasev1.BundlesRef{
				APIVersion: releasev1.GroupVersion.String(),
				Name:       bundle.Name,
				Namespace:  bundle.Namespace,
			},
		},
	}
	return eksaRelease
}

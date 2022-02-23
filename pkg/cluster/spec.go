package cluster

import (
	"embed"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	FluxDefaultNamespace = "flux-system"
	FluxDefaultBranch    = "main"
)

var releasesManifestURL string

type Spec struct {
	*eksav1alpha1.Cluster
	OIDCConfig          *eksav1alpha1.OIDCConfig
	AWSIamConfig        *eksav1alpha1.AWSIamConfig
	GitOpsConfig        *eksav1alpha1.GitOpsConfig
	DatacenterConfig    *metav1.ObjectMeta
	releasesManifestURL string
	bundlesManifestURL  string
	configFS            embed.FS
	userAgent           string
	reader              *ManifestReader
	VersionsBundle      *VersionsBundle
	eksdRelease         *eksdv1alpha1.Release
	Bundles             *v1alpha1.Bundles
	ManagementCluster   *types.Cluster
}

func (s *Spec) DeepCopy() *Spec {
	return &Spec{
		Cluster:             s.Cluster.DeepCopy(),
		OIDCConfig:          s.OIDCConfig.DeepCopy(),
		GitOpsConfig:        s.GitOpsConfig.DeepCopy(),
		releasesManifestURL: s.releasesManifestURL,
		bundlesManifestURL:  s.bundlesManifestURL,
		configFS:            s.configFS,
		reader:              s.reader,
		userAgent:           s.userAgent,
		VersionsBundle: &VersionsBundle{
			VersionsBundle: s.VersionsBundle.VersionsBundle.DeepCopy(),
			KubeDistro:     s.VersionsBundle.KubeDistro.deepCopy(),
		},
		eksdRelease: s.eksdRelease.DeepCopy(),
		Bundles:     s.Bundles.DeepCopy(),
	}
}

func (cs *Spec) SetDefaultGitOps() {
	if cs != nil && cs.GitOpsConfig != nil {
		c := &cs.GitOpsConfig.Spec.Flux
		if len(c.Github.ClusterConfigPath) == 0 {
			if cs.Cluster.IsSelfManaged() {
				c.Github.ClusterConfigPath = path.Join("clusters", cs.Name)
			} else {
				c.Github.ClusterConfigPath = path.Join("clusters", cs.Cluster.ManagedBy())
			}
		}
		if len(c.Github.FluxSystemNamespace) == 0 {
			c.Github.FluxSystemNamespace = FluxDefaultNamespace
		}

		if len(c.Github.Branch) == 0 {
			c.Github.Branch = FluxDefaultBranch
		}
	}
}

type VersionsBundle struct {
	*v1alpha1.VersionsBundle
	KubeDistro *KubeDistro
}

type KubeDistro struct {
	Kubernetes          VersionedRepository
	CoreDNS             VersionedRepository
	Etcd                VersionedRepository
	NodeDriverRegistrar v1alpha1.Image
	LivenessProbe       v1alpha1.Image
	ExternalAttacher    v1alpha1.Image
	ExternalProvisioner v1alpha1.Image
	Pause               v1alpha1.Image
	EtcdImage           v1alpha1.Image
	EtcdVersion         string
	AwsIamAuthIamge     v1alpha1.Image
}

func (k *KubeDistro) deepCopy() *KubeDistro {
	k2 := *k
	return &k2
}

type VersionedRepository struct {
	Repository, Tag string
}

type SpecOpt func(*Spec)

func WithReleasesManifest(manifestURL string) SpecOpt {
	return func(s *Spec) {
		s.releasesManifestURL = manifestURL
	}
}

func WithEmbedFS(embedFS embed.FS) SpecOpt {
	return func(s *Spec) {
		s.configFS = embedFS
	}
}

func WithOverrideBundlesManifest(fileURL string) SpecOpt {
	return func(s *Spec) {
		s.bundlesManifestURL = fileURL
	}
}

func WithManagementCluster(cluster *types.Cluster) SpecOpt {
	return func(s *Spec) {
		s.ManagementCluster = cluster
	}
}

func WithUserAgent(userAgent string) SpecOpt {
	return func(s *Spec) {
		s.userAgent = userAgent
	}
}

func WithGitOpsConfig(gitOpsConfig *eksav1alpha1.GitOpsConfig) SpecOpt {
	return func(s *Spec) {
		s.GitOpsConfig = gitOpsConfig
	}
}

func NewSpec(opts ...SpecOpt) *Spec {
	s := &Spec{
		releasesManifestURL: releasesManifestURL,
		configFS:            configFS,
		userAgent:           userAgent("unknown", "unknown"),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.reader = s.newManifestReader()

	return s
}

func newWithCliVersion(cliVersion version.Info, opts ...SpecOpt) *Spec {
	opts = append(opts, WithUserAgent(userAgent("cli", cliVersion.GitVersion)))
	return NewSpec(opts...)
}

func NewSpecFromClusterConfig(clusterConfigPath string, cliVersion version.Info, opts ...SpecOpt) (*Spec, error) {
	s := newWithCliVersion(cliVersion, opts...)

	clusterConfig, err := eksav1alpha1.GetClusterConfig(clusterConfigPath)
	if err != nil {
		return nil, err
	}

	bundles, err := s.GetBundles(cliVersion)
	if err != nil {
		return nil, err
	}

	versionsBundle, err := s.getVersionsBundle(clusterConfig, bundles)
	if err != nil {
		return nil, err
	}

	eksd, err := s.reader.GetEksdRelease(versionsBundle)
	if err != nil {
		return nil, err
	}

	kubeDistro, err := buildKubeDistro(eksd)
	if err != nil {
		return nil, err
	}

	s.Bundles = bundles
	s.Cluster = clusterConfig
	s.VersionsBundle = &VersionsBundle{
		VersionsBundle: versionsBundle,
		KubeDistro:     kubeDistro,
	}
	s.eksdRelease = eksd
	for _, identityProvider := range s.Cluster.Spec.IdentityProviderRefs {
		switch identityProvider.Kind {
		case eksav1alpha1.OIDCConfigKind:
			oidcConfig, err := eksav1alpha1.GetAndValidateOIDCConfig(clusterConfigPath, identityProvider.Name, clusterConfig)
			if err != nil {
				return nil, err
			}
			s.OIDCConfig = oidcConfig
		case eksav1alpha1.AWSIamConfigKind:
			awsIamConfig, err := eksav1alpha1.GetAndValidateAWSIamConfig(clusterConfigPath, identityProvider.Name, clusterConfig)
			if err != nil {
				return nil, err
			}
			s.AWSIamConfig = awsIamConfig
		}
	}

	if s.Cluster.Spec.GitOpsRef != nil {
		gitOpsConfig, err := eksav1alpha1.GetAndValidateGitOpsConfig(clusterConfigPath, s.Cluster.Spec.GitOpsRef.Name, clusterConfig)
		if err != nil {
			return nil, err
		}
		s.GitOpsConfig = gitOpsConfig
	}

	switch s.Cluster.Spec.DatacenterRef.Kind {
	case eksav1alpha1.VSphereDatacenterKind:
		datacenterConfig, err := eksav1alpha1.GetVSphereDatacenterConfig(clusterConfigPath)
		if err != nil {
			return nil, err
		}
		s.DatacenterConfig = &datacenterConfig.ObjectMeta
	case eksav1alpha1.DockerDatacenterKind:
		datacenterConfig, err := eksav1alpha1.GetDockerDatacenterConfig(clusterConfigPath)
		if err != nil {
			return nil, err
		}
		s.DatacenterConfig = &datacenterConfig.ObjectMeta
	}

	if s.ManagementCluster != nil {
		s.SetManagedBy(s.ManagementCluster.Name)
	} else {
		s.SetSelfManaged()
	}

	return s, nil
}

func BuildSpecFromBundles(cluster *eksav1alpha1.Cluster, bundles *v1alpha1.Bundles, opts ...SpecOpt) (*Spec, error) {
	s := NewSpec(opts...)

	versionsBundle, err := s.getVersionsBundle(cluster, bundles)
	if err != nil {
		return nil, err
	}

	eksd, err := s.reader.GetEksdRelease(versionsBundle)
	if err != nil {
		return nil, err
	}

	kubeDistro, err := buildKubeDistro(eksd)
	if err != nil {
		return nil, err
	}

	s.Bundles = bundles
	s.Cluster = cluster
	s.VersionsBundle = &VersionsBundle{
		VersionsBundle: versionsBundle,
		KubeDistro:     kubeDistro,
	}
	s.eksdRelease = eksd
	return s, nil
}

func (s *Spec) newManifestReader() *ManifestReader {
	return NewManifestReader(files.WithEmbedFS(s.configFS), files.WithUserAgent(s.userAgent))
}

func (s *Spec) getVersionsBundle(clusterConfig *eksav1alpha1.Cluster, bundles *v1alpha1.Bundles) (*v1alpha1.VersionsBundle, error) {
	for _, versionsBundle := range bundles.Spec.VersionsBundles {
		if versionsBundle.KubeVersion == string(clusterConfig.Spec.KubernetesVersion) {
			return &versionsBundle, nil
		}
	}
	return nil, fmt.Errorf("kubernetes version %s is not supported by bundles manifest %d", clusterConfig.Spec.KubernetesVersion, bundles.Spec.Number)
}

func (s *Spec) GetBundles(cliVersion version.Info) (*v1alpha1.Bundles, error) {
	bundlesURL := s.bundlesManifestURL
	if bundlesURL == "" {
		release, err := s.GetRelease(cliVersion)
		if err != nil {
			return nil, err
		}

		bundlesURL = release.BundleManifestUrl
	}

	return s.reader.GetBundles(bundlesURL)
}

func (s *Spec) GetRelease(cliVersion version.Info) (*v1alpha1.EksARelease, error) {
	cliSemVersion, err := semver.New(cliVersion.GitVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid cli version: %v", err)
	}

	releases, err := s.reader.GetReleases(s.releasesManifestURL)
	if err != nil {
		return nil, err
	}

	for _, release := range releases.Spec.Releases {
		releaseVersion, err := semver.New(release.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid version for release %d: %v", release.Number, err)
		}

		if cliSemVersion.SamePrerelease(releaseVersion) {
			return &release, nil
		}
	}

	return nil, fmt.Errorf("eksa release %s does not exist in manifest %s", cliVersion, s.releasesManifestURL)
}

func (s *Spec) KubeDistroImages() []v1alpha1.Image {
	images := []v1alpha1.Image{}
	for _, component := range s.eksdRelease.Status.Components {
		for _, asset := range component.Assets {
			if asset.Image != nil {
				images = append(images, v1alpha1.Image{URI: asset.Image.URI})
			}
		}
	}
	return images
}

func buildKubeDistro(eksd *eksdv1alpha1.Release) (*KubeDistro, error) {
	kubeDistro := &KubeDistro{}
	assets := make(map[string]*eksdv1alpha1.AssetImage)
	for _, component := range eksd.Status.Components {
		for _, asset := range component.Assets {
			if asset.Image != nil {
				assets[asset.Name] = asset.Image
			}
		}
		if component.Name == "etcd" {
			kubeDistro.EtcdVersion = strings.TrimPrefix(component.GitTag, "v")
		}
	}

	kubeDistroComponents := map[string]*v1alpha1.Image{
		"node-driver-registrar-image": &kubeDistro.NodeDriverRegistrar,
		"livenessprobe-image":         &kubeDistro.LivenessProbe,
		"external-attacher-image":     &kubeDistro.ExternalAttacher,
		"external-provisioner-image":  &kubeDistro.ExternalProvisioner,
		"pause-image":                 &kubeDistro.Pause,
		"etcd-image":                  &kubeDistro.EtcdImage,
		"aws-iam-authenticator-image": &kubeDistro.AwsIamAuthIamge,
	}

	for assetName, image := range kubeDistroComponents {
		i := assets[assetName]
		if i == nil {
			return nil, fmt.Errorf("asset %s is no present in eksd release %s", assetName, eksd.Spec.Channel)
		}

		image.URI = i.URI
	}

	kubeDistroRepositories := map[string]*VersionedRepository{
		"coredns-image":        &kubeDistro.CoreDNS,
		"etcd-image":           &kubeDistro.Etcd,
		"kube-apiserver-image": &kubeDistro.Kubernetes,
	}

	for assetName, image := range kubeDistroRepositories {
		i := assets[assetName]
		if i == nil {
			return nil, fmt.Errorf("asset %s is not present in eksd release %s", assetName, eksd.Spec.Channel)
		}

		image.Repository, image.Tag = kubeDistroRepository(i)
	}

	return kubeDistro, nil
}

func kubeDistroRepository(image *eksdv1alpha1.AssetImage) (repo, tag string) {
	i := v1alpha1.Image{
		URI: image.URI,
	}

	lastInd := strings.LastIndex(i.Image(), "/")
	if lastInd == -1 {
		return i.Image(), i.Tag()
	}

	return i.Image()[:lastInd], i.Tag()
}

func GetEksdRelease(cliVersion version.Info, clusterConfig *eksav1alpha1.Cluster) (*v1alpha1.EksDRelease, *eksdv1alpha1.Release, error) {
	s := newWithCliVersion(cliVersion)

	bundles, err := s.GetBundles(cliVersion)
	if err != nil {
		return nil, nil, err
	}

	versionsBundle, err := s.getVersionsBundle(clusterConfig, bundles)
	if err != nil {
		return nil, nil, err
	}

	eksdRelease, err := s.reader.GetEksdRelease(versionsBundle)
	if err != nil {
		return nil, nil, err
	}

	return &versionsBundle.EksD, eksdRelease, nil
}

type Manifest struct {
	Filename string
	Content  []byte
}

func (s *Spec) LoadManifest(manifest v1alpha1.Manifest) (*Manifest, error) {
	url, err := url.Parse(manifest.URI)
	if err != nil {
		return nil, fmt.Errorf("invalid manifest URI: %v", err)
	}

	content, err := s.reader.ReadFile(manifest.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %v", err)
	}

	return &Manifest{
		Filename: filepath.Base(url.Path),
		Content:  content,
	}, nil
}

func userAgent(eksAComponent, version string) string {
	return fmt.Sprintf("eks-a-%s/%s", eksAComponent, version)
}

func (vb *VersionsBundle) KubeDistroImages() []v1alpha1.Image {
	var images []v1alpha1.Image
	images = append(images, vb.KubeDistro.EtcdImage)
	images = append(images, vb.KubeDistro.ExternalAttacher)
	images = append(images, vb.KubeDistro.ExternalProvisioner)
	images = append(images, vb.KubeDistro.LivenessProbe)
	images = append(images, vb.KubeDistro.NodeDriverRegistrar)
	images = append(images, vb.KubeDistro.Pause)

	return images
}

func (vb *VersionsBundle) Images() []v1alpha1.Image {
	var images []v1alpha1.Image
	images = append(images, vb.SharedImages()...)
	images = append(images, vb.KubeDistroImages()...)
	images = append(images, vb.DockerImages()...)
	images = append(images, vb.VsphereImages()...)

	return images
}

func (vb *VersionsBundle) Ovas() []v1alpha1.Archive {
	return vb.VersionsBundle.Ovas()
}

func (s *Spec) GetReleaseManifestUrl() string {
	return s.releasesManifestURL
}

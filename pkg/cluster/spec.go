package cluster

import (
	"embed"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
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
	*Config
	OIDCConfig                *eksav1alpha1.OIDCConfig
	AWSIamConfig              *eksav1alpha1.AWSIamConfig
	releasesManifestURL       string
	bundlesManifestURL        string
	configFS                  embed.FS
	userAgent                 string
	reader                    *ManifestReader
	VersionsBundle            *VersionsBundle
	eksdRelease               *eksdv1alpha1.Release
	Bundles                   *v1alpha1.Bundles
	ManagementCluster         *types.Cluster
	TinkerbellTemplateConfigs map[string]*eksav1alpha1.TinkerbellTemplateConfig
}

func (s *Spec) DeepCopy() *Spec {
	return &Spec{
		Config:              s.Config.DeepCopy(),
		OIDCConfig:          s.OIDCConfig.DeepCopy(),
		AWSIamConfig:        s.AWSIamConfig.DeepCopy(),
		releasesManifestURL: s.releasesManifestURL,
		bundlesManifestURL:  s.bundlesManifestURL,
		configFS:            s.configFS,
		reader:              s.reader,
		userAgent:           s.userAgent,
		VersionsBundle: &VersionsBundle{
			VersionsBundle: s.VersionsBundle.VersionsBundle.DeepCopy(),
			KubeDistro:     s.VersionsBundle.KubeDistro.deepCopy(),
		},
		eksdRelease:               s.eksdRelease.DeepCopy(),
		Bundles:                   s.Bundles.DeepCopy(),
		TinkerbellTemplateConfigs: s.TinkerbellTemplateConfigs,
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
	AwsIamAuthImage     v1alpha1.Image
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

func WithEksdRelease(release *eksdv1alpha1.Release) SpecOpt {
	return func(s *Spec) {
		s.eksdRelease = release
	}
}

func WithGitOpsConfig(gitOpsConfig *eksav1alpha1.GitOpsConfig) SpecOpt {
	return func(s *Spec) {
		s.GitOpsConfig = gitOpsConfig
	}
}

func WithOIDCConfig(oidcConfig *eksav1alpha1.OIDCConfig) SpecOpt {
	return func(s *Spec) {
		s.OIDCConfig = oidcConfig
	}
}

func NewSpec(opts ...SpecOpt) *Spec {
	s := &Spec{
		Config:              &Config{},
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

	clusterConfig, err := ParseConfigFromFile(clusterConfigPath)
	if err != nil {
		return nil, err
	}
	if err = SetConfigDefaults(clusterConfig); err != nil {
		return nil, err
	}

	bundles, err := s.GetBundles(cliVersion)
	if err != nil {
		return nil, err
	}

	versionsBundle, err := s.getVersionsBundle(clusterConfig.Cluster.Spec.KubernetesVersion, bundles)
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
	s.Config = clusterConfig
	s.VersionsBundle = &VersionsBundle{
		VersionsBundle: versionsBundle,
		KubeDistro:     kubeDistro,
	}
	s.eksdRelease = eksd

	// Get first aws iam config if it exists
	// Config supports multiple configs because Cluster references a slice
	// But we validate that only one of each type is referenced
	for _, ac := range s.Config.AWSIAMConfigs {
		s.AWSIamConfig = ac
		break
	}

	// Get first oidc config if it exists
	for _, oc := range s.Config.OIDCConfigs {
		s.OIDCConfig = oc
		break
	}

	switch s.Cluster.Spec.DatacenterRef.Kind {
	case eksav1alpha1.TinkerbellDatacenterKind:
		if features.IsActive(features.TinkerbellProvider()) {
			templateConfigs, err := eksav1alpha1.GetTinkerbellTemplateConfig(clusterConfigPath)
			if err != nil {
				return nil, err
			}
			s.TinkerbellTemplateConfigs = templateConfigs
		} else {
			return nil, fmt.Errorf("unsupported DatacenterRef.Kind: %s", eksav1alpha1.TinkerbellDatacenterKind)
		}
	}

	if s.ManagementCluster != nil {
		s.Cluster.SetManagedBy(s.ManagementCluster.Name)
	} else {
		s.Cluster.SetSelfManaged()
	}

	return s, nil
}

func BuildSpecFromBundles(cluster *eksav1alpha1.Cluster, bundles *v1alpha1.Bundles, opts ...SpecOpt) (*Spec, error) {
	s := NewSpec(opts...)

	versionsBundle, err := s.getVersionsBundle(cluster.Spec.KubernetesVersion, bundles)
	if err != nil {
		return nil, err
	}

	if s.eksdRelease == nil {
		eksd, err := s.reader.GetEksdRelease(versionsBundle)
		if err != nil {
			return nil, err
		}
		s.eksdRelease = eksd
	}
	kubeDistro, err := buildKubeDistro(s.eksdRelease)
	if err != nil {
		return nil, err
	}

	s.Bundles = bundles
	s.Config.Cluster = cluster
	s.VersionsBundle = &VersionsBundle{
		VersionsBundle: versionsBundle,
		KubeDistro:     kubeDistro,
	}

	return s, nil
}

func (s *Spec) newManifestReader() *ManifestReader {
	return NewManifestReader(files.WithEmbedFS(s.configFS), files.WithUserAgent(s.userAgent))
}

func (s *Spec) getVersionsBundle(kubeVersion eksav1alpha1.KubernetesVersion, bundles *v1alpha1.Bundles) (*v1alpha1.VersionsBundle, error) {
	for _, versionsBundle := range bundles.Spec.VersionsBundles {
		if versionsBundle.KubeVersion == string(kubeVersion) {
			return &versionsBundle, nil
		}
	}
	return nil, fmt.Errorf("kubernetes version %s is not supported by bundles manifest %d", kubeVersion, bundles.Spec.Number)
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
		"aws-iam-authenticator-image": &kubeDistro.AwsIamAuthImage,
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

	versionsBundle, err := s.getVersionsBundle(clusterConfig.Spec.KubernetesVersion, bundles)
	if err != nil {
		return nil, nil, err
	}

	eksdRelease, err := s.reader.GetEksdRelease(versionsBundle)
	if err != nil {
		return nil, nil, err
	}

	return &versionsBundle.EksD, eksdRelease, nil
}

// GetVersionsBundleForVersion returns the  versionBundle for gitVersion and kubernetes version
func GetVersionsBundleForVersion(cliVersion version.Info, kubernetesVersion eksav1alpha1.KubernetesVersion) (*v1alpha1.VersionsBundle, error) {
	s := newWithCliVersion(cliVersion)
	bundles, err := s.GetBundles(cliVersion)
	if err != nil {
		return nil, err
	}

	return s.getVersionsBundle(kubernetesVersion, bundles)
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

type EksdManifests struct {
	ReleaseManifestContent []byte
	ReleaseCrdContent      []byte
}

func (s *Spec) ReadEksdManifests(release v1alpha1.EksDRelease) (*EksdManifests, error) {
	releaseCrdContent, err := s.reader.ReadFile(release.Components)
	if err != nil {
		return nil, err
	}

	releaseManifestContent, err := s.reader.ReadFile(release.EksDReleaseUrl)
	if err != nil {
		return nil, err
	}

	return &EksdManifests{
		ReleaseManifestContent: releaseManifestContent,
		ReleaseCrdContent:      releaseCrdContent,
	}, nil
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
	images = append(images, vb.CloudStackImages()...)

	return images
}

func (vb *VersionsBundle) Ovas() []v1alpha1.Archive {
	return vb.VersionsBundle.Ovas()
}

func (s *Spec) GetReleaseManifestUrl() string {
	return s.releasesManifestURL
}

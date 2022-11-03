package cluster

import (
	"embed"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
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
	reader                    *files.Reader
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

func WithFluxConfig(fluxConfig *eksav1alpha1.FluxConfig) SpecOpt {
	return func(s *Spec) {
		s.FluxConfig = fluxConfig
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

func WithAWSIamConfig(awsIamConfig *eksav1alpha1.AWSIamConfig) SpecOpt {
	return func(s *Spec) {
		s.AWSIamConfig = awsIamConfig
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

	s.reader = s.newReader()

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
	bundlesManifest, err := s.GetBundles(cliVersion)
	if err != nil {
		return nil, err
	}
	bundlesManifest.Namespace = constants.EksaSystemNamespace

	configManager, err := NewDefaultConfigManager()
	if err != nil {
		return nil, err
	}
	configManager.RegisterDefaulters(BundlesRefDefaulter())

	if err = configManager.SetDefaults(clusterConfig); err != nil {
		return nil, err
	}

	// We are pulling the latest available Bundles, so making sure we update the ref to make the spec consistent
	clusterConfig.Cluster.Spec.BundlesRef.Name = bundlesManifest.Name
	clusterConfig.Cluster.Spec.BundlesRef.Namespace = bundlesManifest.Namespace
	clusterConfig.Cluster.Spec.BundlesRef.APIVersion = v1alpha1.GroupVersion.String()

	versionsBundle, err := GetVersionsBundle(clusterConfig.Cluster, bundlesManifest)
	if err != nil {
		return nil, err
	}

	eksd, err := bundles.ReadEKSD(s.reader, *versionsBundle)
	if err != nil {
		return nil, err
	}

	if err = s.init(clusterConfig, bundlesManifest, versionsBundle, eksd); err != nil {
		return nil, err
	}

	switch s.Cluster.Spec.DatacenterRef.Kind {
	case eksav1alpha1.TinkerbellDatacenterKind:
		templateConfigs, err := eksav1alpha1.GetTinkerbellTemplateConfig(clusterConfigPath)
		if err != nil {
			return nil, err
		}
		s.TinkerbellTemplateConfigs = templateConfigs
	}

	return s, nil
}

// init does the basic initialization with the provided necessary api objects.
func (s *Spec) init(config *Config, bundles *v1alpha1.Bundles, versionsBundle *v1alpha1.VersionsBundle, eksdRelease *eksdv1alpha1.Release) error {
	kubeDistro, err := buildKubeDistro(eksdRelease)
	if err != nil {
		return err
	}

	s.Bundles = bundles
	s.Config = config
	s.VersionsBundle = &VersionsBundle{
		VersionsBundle: versionsBundle,
		KubeDistro:     kubeDistro,
	}
	s.eksdRelease = eksdRelease

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

	return nil
}

func BuildSpecFromBundles(cluster *eksav1alpha1.Cluster, bundlesManifest *v1alpha1.Bundles, opts ...SpecOpt) (*Spec, error) {
	s := NewSpec(opts...)

	versionsBundle, err := GetVersionsBundle(cluster, bundlesManifest)
	if err != nil {
		return nil, err
	}

	if s.eksdRelease == nil {
		eksd, err := bundles.ReadEKSD(s.reader, *versionsBundle)
		if err != nil {
			return nil, err
		}
		s.eksdRelease = eksd
	}
	kubeDistro, err := buildKubeDistro(s.eksdRelease)
	if err != nil {
		return nil, err
	}

	s.Bundles = bundlesManifest
	s.Config.Cluster = cluster
	s.VersionsBundle = &VersionsBundle{
		VersionsBundle: versionsBundle,
		KubeDistro:     kubeDistro,
	}

	return s, nil
}

func (s *Spec) newReader() *files.Reader {
	return files.NewReader(files.WithEmbedFS(s.configFS), files.WithUserAgent(s.userAgent))
}

func (s *Spec) GetBundles(cliVersion version.Info) (*v1alpha1.Bundles, error) {
	bundlesURL := s.bundlesManifestURL
	if bundlesURL == "" {
		manifestReader := manifests.NewReader(s.reader, manifests.WithReleasesManifest(s.releasesManifestURL))
		return manifestReader.ReadBundlesForVersion(cliVersion.GitVersion)
	}

	return bundles.Read(s.reader, bundlesURL)
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
	images = append(images, vb.VersionsBundle.Images()...)
	images = append(images, vb.KubeDistroImages()...)

	return images
}

func (vb *VersionsBundle) Ovas() []v1alpha1.Archive {
	return vb.VersionsBundle.Ovas()
}

func BundlesRefDefaulter() Defaulter {
	return func(c *Config) error {
		if c.Cluster.Spec.BundlesRef == nil {
			c.Cluster.Spec.BundlesRef = &eksav1alpha1.BundlesRef{}
		}
		return nil
	}
}

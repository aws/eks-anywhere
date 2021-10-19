package cluster

import (
	"embed"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"sigs.k8s.io/yaml"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	httpsScheme          = "https"
	embedScheme          = "embed"
	FluxDefaultNamespace = "flux-system"
	FluxDefaultBranch    = "main"
)

var releasesManifestURL string

type Spec struct {
	*eksav1alpha1.Cluster
	OIDCConfig          *eksav1alpha1.OIDCConfig
	GitOpsConfig        *eksav1alpha1.GitOpsConfig
	releasesManifestURL string
	bundlesManifestURL  string
	configFS            embed.FS
	httpClient          *http.Client
	userAgent           string
	VersionsBundle      *VersionsBundle
	eksdRelease         *eksdv1alpha1.Release
	Bundles             *v1alpha1.Bundles
}

func (s *Spec) DeepCopy() *Spec {
	return &Spec{
		Cluster:             s.Cluster.DeepCopy(),
		OIDCConfig:          s.OIDCConfig.DeepCopy(),
		GitOpsConfig:        s.GitOpsConfig.DeepCopy(),
		releasesManifestURL: s.releasesManifestURL,
		bundlesManifestURL:  s.bundlesManifestURL,
		configFS:            s.configFS,
		httpClient:          s.httpClient,
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
			c.Github.ClusterConfigPath = path.Join("clusters", cs.Name)
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

func NewSpec(clusterConfigPath string, cliVersion version.Info, opts ...SpecOpt) (*Spec, error) {
	clusterConfig, err := eksav1alpha1.GetClusterConfig(clusterConfigPath)
	if err != nil {
		return nil, err
	}

	s := &Spec{
		releasesManifestURL: releasesManifestURL,
		configFS:            configFS,
		httpClient:          &http.Client{},
		userAgent:           userAgent("cli", cliVersion.GitVersion),
	}
	for _, opt := range opts {
		opt(s)
	}

	bundles, err := s.GetBundles(cliVersion)
	if err != nil {
		return nil, err
	}

	versionsBundle, err := s.getVersionsBundle(clusterConfig, bundles)
	if err != nil {
		return nil, err
	}

	eksd, err := s.getEksdRelease(versionsBundle)
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
	if len(s.Cluster.Spec.IdentityProviderRefs) != 0 {
		// Since we only support one configuration, and only OIDCConfig, for identityProviderRefs, it is safe to assume that
		// it is the only element that exists in the array
		oidcConfig, err := eksav1alpha1.GetAndValidateOIDCConfig(clusterConfigPath, s.Cluster.Spec.IdentityProviderRefs[0].Name, clusterConfig)
		if err != nil {
			return nil, err
		}
		s.OIDCConfig = oidcConfig
	}

	if s.Cluster.Spec.GitOpsRef != nil {
		gitOpsConfig, err := eksav1alpha1.GetAndValidateGitOpsConfig(clusterConfigPath, s.Cluster.Spec.GitOpsRef.Name, clusterConfig)
		if err != nil {
			return nil, err
		}
		s.GitOpsConfig = gitOpsConfig
	}
	s.SetDefaultGitOps()

	return s, nil
}

func BuildSpecFromBundles(cluster *eksav1alpha1.Cluster, bundles *v1alpha1.Bundles, opts ...SpecOpt) (*Spec, error) {
	s := &Spec{
		releasesManifestURL: releasesManifestURL,
		configFS:            configFS,
		httpClient:          &http.Client{},
		userAgent:           userAgent("unknown", "unknown"),
	}
	for _, opt := range opts {
		opt(s)
	}

	versionsBundle, err := s.getVersionsBundle(cluster, bundles)
	if err != nil {
		return nil, err
	}

	eksd, err := s.getEksdRelease(versionsBundle)
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
		release, err := s.getRelease(cliVersion)
		if err != nil {
			return nil, err
		}

		bundlesURL = release.BundleManifestUrl
	}

	logger.V(4).Info("Reading bundles manifest", "url", bundlesURL)
	content, err := s.readFile(bundlesURL)
	if err != nil {
		return nil, err
	}

	bundles := &v1alpha1.Bundles{}
	if err = yaml.Unmarshal(content, bundles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundles manifest from [%s] to build cluster spec: %v", bundlesURL, err)
	}

	return bundles, nil
}

func (s *Spec) getRelease(cliVersion version.Info) (*v1alpha1.EksARelease, error) {
	cliSemVersion, err := semver.New(cliVersion.GitVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid cli version: %v", err)
	}

	releases, err := s.getReleases(s.releasesManifestURL)
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

func (s *Spec) getReleases(releasesManifest string) (*v1alpha1.Release, error) {
	logger.V(4).Info("Reading releases manifest", "url", releasesManifestURL)
	content, err := s.readFile(releasesManifest)
	if err != nil {
		return nil, err
	}

	releases := &v1alpha1.Release{}
	if err = yaml.Unmarshal(content, releases); err != nil {
		return nil, fmt.Errorf("failed to unmarshal release manifest to build cluster spec: %v", err)
	}

	return releases, nil
}

func (s *Spec) readFile(uri string) ([]byte, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("can't build cluster spec, invalid release manifest url: %v", err)
	}

	switch url.Scheme {
	case httpsScheme:
		return s.readHttpFile(uri)
	case embedScheme:
		return s.readEmbedFile(url)
	default:
		return readLocalFile(uri)
	}
}

func (s *Spec) readHttpFile(uri string) ([]byte, error) {
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating http GET request for downloading file: %v", err)
	}

	request.Header.Set("User-Agent", s.userAgent)
	resp, err := s.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed reading file from url [%s] for cluster spec: %v", uri, err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading file from url [%s] for cluster spec: %v", uri, err)
	}

	return data, nil
}

func (s *Spec) readEmbedFile(url *url.URL) ([]byte, error) {
	data, err := s.configFS.ReadFile(strings.TrimPrefix(url.Path, "/"))
	if err != nil {
		return nil, fmt.Errorf("failed reading embed file [%s] for cluster spec: %v", url.Path, err)
	}

	return data, nil
}

func readLocalFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed reading local file [%s] for cluster spec: %v", filename, err)
	}

	return data, nil
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

func (s *Spec) getEksdRelease(versionsBundle *v1alpha1.VersionsBundle) (*eksdv1alpha1.Release, error) {
	content, err := s.readFile(versionsBundle.EksD.EksDReleaseUrl)
	if err != nil {
		return nil, err
	}

	eksd := &eksdv1alpha1.Release{}
	if err = yaml.Unmarshal(content, eksd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal eksd manifest to build cluster spec: %v", err)
	}

	return eksd, nil
}

func GetEksdRelease(cliVersion version.Info, clusterConfig *eksav1alpha1.Cluster) (*v1alpha1.EksDRelease, error) {
	s := &Spec{
		releasesManifestURL: releasesManifestURL,
		configFS:            configFS,
		httpClient:          &http.Client{},
		userAgent:           userAgent("cli", cliVersion.GitVersion),
	}

	bundles, err := s.GetBundles(cliVersion)
	if err != nil {
		return nil, err
	}

	versionsBundle, err := s.getVersionsBundle(clusterConfig, bundles)
	if err != nil {
		return nil, err
	}

	return &versionsBundle.EksD, nil
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

	content, err := s.readFile(manifest.URI)
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

func (vb *VersionsBundle) SharedImages() []v1alpha1.Image {
	var images []v1alpha1.Image
	images = append(images, vb.Bootstrap.Controller)
	images = append(images, vb.Bootstrap.KubeProxy)

	images = append(images, vb.BottleRocketBootstrap.Bootstrap)

	images = append(images, vb.CertManager.Acmesolver)
	images = append(images, vb.CertManager.Cainjector)
	images = append(images, vb.CertManager.Controller)
	images = append(images, vb.CertManager.Webhook)

	images = append(images, vb.Cilium.Cilium)
	images = append(images, vb.Cilium.Operator)

	images = append(images, vb.ClusterAPI.Controller)
	images = append(images, vb.ClusterAPI.KubeProxy)

	images = append(images, vb.ControlPlane.Controller)
	images = append(images, vb.ControlPlane.KubeProxy)

	images = append(images, vb.EksD.KindNode)
	images = append(images, vb.Eksa.CliTools)
	images = append(images, vb.Eksa.ClusterController)

	images = append(images, vb.Flux.HelmController)
	images = append(images, vb.Flux.KustomizeController)
	images = append(images, vb.Flux.NotificationController)
	images = append(images, vb.Flux.SourceController)

	images = append(images, vb.ExternalEtcdBootstrap.Controller)
	images = append(images, vb.ExternalEtcdBootstrap.KubeProxy)

	images = append(images, vb.ExternalEtcdController.Controller)
	images = append(images, vb.ExternalEtcdController.KubeProxy)

	images = append(images, vb.KubeDistro.EtcdImage)
	images = append(images, vb.KubeDistro.ExternalAttacher)
	images = append(images, vb.KubeDistro.ExternalProvisioner)
	images = append(images, vb.KubeDistro.LivenessProbe)
	images = append(images, vb.KubeDistro.NodeDriverRegistrar)
	images = append(images, vb.KubeDistro.Pause)

	return images
}

func (vb *VersionsBundle) VsphereImages() []v1alpha1.Image {
	var images []v1alpha1.Image
	images = append(images, vb.VSphere.ClusterAPIController)
	images = append(images, vb.VSphere.Driver)
	images = append(images, vb.VSphere.KubeProxy)
	images = append(images, vb.VSphere.KubeVip)
	images = append(images, vb.VSphere.Manager)
	images = append(images, vb.VSphere.Syncer)

	return images
}

func (vb *VersionsBundle) DockerImages() []v1alpha1.Image {
	var images []v1alpha1.Image
	images = append(images, vb.Docker.KubeProxy)
	images = append(images, vb.Docker.Manager)

	return images
}

func (vb *VersionsBundle) Images() []v1alpha1.Image {
	var images []v1alpha1.Image
	images = append(images, vb.SharedImages()...)
	images = append(images, vb.DockerImages()...)
	images = append(images, vb.VsphereImages()...)

	return images
}

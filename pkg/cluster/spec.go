package cluster

import (
	"fmt"
	"strings"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type Spec struct {
	*Config
	Bundles           *v1alpha1.Bundles
	OIDCConfig        *eksav1alpha1.OIDCConfig
	AWSIamConfig      *eksav1alpha1.AWSIamConfig
	ManagementCluster *types.Cluster // TODO(g-gaston): cleanup, this doesn't belong here
	EKSARelease       *v1alpha1.EKSARelease
	VersionsBundles   map[eksav1alpha1.KubernetesVersion]*VersionsBundle
}

func (s *Spec) DeepCopy() *Spec {
	ns := &Spec{
		Config:          s.Config.DeepCopy(),
		OIDCConfig:      s.OIDCConfig.DeepCopy(),
		AWSIamConfig:    s.AWSIamConfig.DeepCopy(),
		Bundles:         s.Bundles.DeepCopy(),
		VersionsBundles: deepCopyVersionsBundles(s.VersionsBundles),
		EKSARelease:     s.EKSARelease.DeepCopy(),
	}

	if s.ManagementCluster != nil {
		ns.ManagementCluster = s.ManagementCluster.DeepCopy()
	}

	return ns
}

type VersionsBundle struct {
	*v1alpha1.VersionsBundle
	KubeDistro *KubeDistro
}

func deepCopyVersionsBundles(v map[eksav1alpha1.KubernetesVersion]*VersionsBundle) map[eksav1alpha1.KubernetesVersion]*VersionsBundle {
	m := make(map[eksav1alpha1.KubernetesVersion]*VersionsBundle, len(v))
	for key, val := range v {
		m[key] = &VersionsBundle{
			VersionsBundle: val.VersionsBundle.DeepCopy(),
			KubeDistro:     val.KubeDistro.deepCopy(),
		}
	}
	return m
}

// EKSD represents an eks-d release.
type EKSD struct {
	// Channel is the minor Kubernetes version for the eks-d release (eg. "1.23", "1.24", etc.)
	Channel string
	// Number is the monotonically increasing number that distinguishes the different eks-d releases
	// for the same Kubernetes minor version (channel).
	Number int
}

func (k *KubeDistro) deepCopy() *KubeDistro {
	k2 := *k
	return &k2
}

type KubeDistro struct {
	EKSD                EKSD
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
	EtcdURL             string
	AwsIamAuthImage     v1alpha1.Image
	KubeProxy           v1alpha1.Image
}

type VersionedRepository struct {
	Repository, Tag string
}

// NewSpec builds a new [Spec].
func NewSpec(config *Config, bundles *v1alpha1.Bundles, eksdReleases []eksdv1alpha1.Release, eksaRelease *v1alpha1.EKSARelease) (*Spec, error) {
	s := &Spec{}

	s.Bundles = bundles
	s.Config = config

	vb, err := getAllVersionsBundles(s.Cluster, bundles, eksdReleases)
	if err != nil {
		return nil, err
	}
	s.VersionsBundles = vb
	s.EKSARelease = eksaRelease

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

	return s, nil
}

func getAllVersionsBundles(cluster *eksav1alpha1.Cluster, bundles *v1alpha1.Bundles, eksdReleases []eksdv1alpha1.Release) (map[eksav1alpha1.KubernetesVersion]*VersionsBundle, error) {
	if len(eksdReleases) < 1 {
		return nil, fmt.Errorf("no eksd releases were found")
	}

	m := make(map[eksav1alpha1.KubernetesVersion]*VersionsBundle, len(eksdReleases))

	for _, eksd := range eksdReleases {
		channel := strings.Replace(eksd.Spec.Channel, "-", ".", 1)
		version := eksav1alpha1.KubernetesVersion(channel)

		if _, ok := m[version]; ok {
			continue
		}

		versionBundle, err := getVersionBundles(version, bundles, &eksd)
		if err != nil {
			return nil, err
		}

		m[version] = versionBundle
	}

	return m, nil
}

func getVersionBundles(version eksav1alpha1.KubernetesVersion, b *v1alpha1.Bundles, eksdRelease *eksdv1alpha1.Release) (*VersionsBundle, error) {
	v, err := GetVersionsBundle(version, b)
	if err != nil {
		return nil, err
	}

	kd, err := buildKubeDistro(eksdRelease)
	if err != nil {
		return nil, err
	}

	vb := &VersionsBundle{
		VersionsBundle: v,
		KubeDistro:     kd,
	}

	return vb, nil
}

// VersionsBundle returns a VersionsBundle if one exists for the provided kubernetes version and nil otherwise.
func (s *Spec) VersionsBundle(version eksav1alpha1.KubernetesVersion) *VersionsBundle {
	vb, ok := s.VersionsBundles[version]
	if !ok {
		return nil
	}

	return vb
}

// RootVersionsBundle returns a VersionsBundle for the Cluster objects root Kubernetes versions.
func (s *Spec) RootVersionsBundle() *VersionsBundle {
	return s.VersionsBundle(s.Cluster.Spec.KubernetesVersion)
}

// WorkerNodeGroupVersionsBundle returns a VersionsBundle for the Worker Node's kubernetes version.
func (s *Spec) WorkerNodeGroupVersionsBundle(w eksav1alpha1.WorkerNodeGroupConfiguration) *VersionsBundle {
	if w.KubernetesVersion != nil {
		return s.VersionsBundle(*w.KubernetesVersion)
	}
	return s.RootVersionsBundle()
}

func buildKubeDistro(eksd *eksdv1alpha1.Release) (*KubeDistro, error) {
	kubeDistro := &KubeDistro{
		EKSD: EKSD{
			Channel: eksd.Spec.Channel,
			Number:  eksd.Spec.Number,
		},
	}
	assets := make(map[string]*eksdv1alpha1.AssetImage)
	for _, component := range eksd.Status.Components {
		for _, asset := range component.Assets {
			if asset.Image != nil {
				assets[asset.Name] = asset.Image
			}

			if component.Name == "etcd" {
				kubeDistro.EtcdVersion = strings.TrimPrefix(component.GitTag, "v")

				// Get archive uri for amd64
				if asset.Archive != nil && len(asset.Arch) > 0 {
					if asset.Arch[0] == "amd64" {
						kubeDistro.EtcdURL = asset.Archive.URI
					}
				}
			}
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
		"kube-proxy-image":            &kubeDistro.KubeProxy,
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

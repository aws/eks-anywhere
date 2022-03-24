package cluster_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestConfigManagerParseSuccess(t *testing.T) {
	g := NewWithT(t)
	c := cluster.NewConfigManager()
	g.Expect(c.RegisterMapping(anywherev1.DockerDatacenterKind, func() cluster.APIObject {
		return &anywherev1.DockerDatacenterConfig{}
	})).To(Succeed())
	c.RegisterProcessors(func(c *cluster.Config, ol cluster.ObjectLookup) {
		d := ol.GetFromRef(c.Cluster.APIVersion, c.Cluster.Spec.DatacenterRef)
		c.DockerDatacenter = d.(*anywherev1.DockerDatacenterConfig)
	})

	yamlManifest := []byte(test.ReadFile(t, "testdata/docker_cluster_oidc_awsiam_flux.yaml"))
	wantCluster := &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "m-docker",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.21",
			ManagementCluster: anywherev1.ManagementCluster{
				Name: "m-docker",
			},
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Count: 1,
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Name:  "workers-1",
					Count: 1,
				},
			},
			DatacenterRef: anywherev1.Ref{
				Kind: anywherev1.DockerDatacenterKind,
				Name: "m-docker",
			},
			ClusterNetwork: anywherev1.ClusterNetwork{
				Pods: anywherev1.Pods{
					CidrBlocks: []string{"192.168.0.0/16"},
				},
				Services: anywherev1.Services{
					CidrBlocks: []string{"10.96.0.0/12"},
				},
				CNI: "cilium",
			},
			IdentityProviderRefs: []anywherev1.Ref{
				{
					Kind: anywherev1.OIDCConfigKind,
					Name: "eksa-unit-test",
				},
				{
					Kind: anywherev1.AWSIamConfigKind,
					Name: "eksa-unit-test",
				},
			},
			GitOpsRef: &anywherev1.Ref{
				Kind: anywherev1.FluxConfigKind,
				Name: "eksa-unit-test",
			},
		},
	}
	wantDockerDatacenter := &anywherev1.DockerDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.DockerDatacenterKind,
			APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "m-docker",
		},
	}
	wantConfig := &cluster.Config{
		Cluster:          wantCluster,
		DockerDatacenter: wantDockerDatacenter,
	}

	config, err := c.Parse(yamlManifest)
	g.Expect(err).To(BeNil())
	g.Expect(config).To(Equal(wantConfig))
}

func TestConfigManagerParseMissingCluster(t *testing.T) {
	g := NewWithT(t)
	c := cluster.NewConfigManager()
	_, err := c.Parse([]byte{})
	g.Expect(err).To(MatchError(ContainSubstring("no Cluster found in manifest")))
}

func TestConfigManagerParseTwoClusters(t *testing.T) {
	g := NewWithT(t)
	manifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: eksa-unit-test
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: eksa-unit-test-2
`
	c := cluster.NewConfigManager()
	_, err := c.Parse([]byte(manifest))
	g.Expect(err).To(MatchError(ContainSubstring("only one Cluster per yaml manifest is allowed")))
}

func TestConfigManagerParseUnknownKind(t *testing.T) {
	g := NewWithT(t)
	manifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: eksa-unit-test
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: MysteryCRD
`

	wantCluster := &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "eksa-unit-test",
		},
	}
	wantConfig := &cluster.Config{
		Cluster: wantCluster,
	}

	c := cluster.NewConfigManager()
	g.Expect(c.Parse([]byte(manifest))).To(Equal(wantConfig))
}

func TestConfigManagerSetDefaultsSuccess(t *testing.T) {
	g := NewWithT(t)
	defaultNamespace := "default"
	c := cluster.NewConfigManager()
	c.RegisterDefaulters(func(c *cluster.Config) error {
		if c.Cluster.Namespace == "" {
			c.Cluster.Namespace = defaultNamespace
		}
		return nil
	})

	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       anywherev1.ClusterKind,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "eksa-unit-test",
			},
		},
	}

	g.Expect(c.SetDefaults(config)).To(Succeed())
	g.Expect(config.Cluster.Namespace).To(Equal(defaultNamespace))
}

func TestConfigManagerValidateMultipleErrors(t *testing.T) {
	g := NewWithT(t)
	c := cluster.NewConfigManager()
	c.RegisterValidations(
		func(c *cluster.Config) error {
			if c.Cluster.Namespace == "" {
				return errors.New("cluster namespace can't be empty")
			}
			return nil
		},
		func(c *cluster.Config) error {
			if c.Cluster.Name == "" {
				return errors.New("cluster name can't be empty")
			}
			return nil
		},
		func(c *cluster.Config) error {
			if c.Cluster.Spec.KubernetesVersion == "" {
				return errors.New("cluster kubernetes version can't be empty")
			}
			return nil
		},
	)

	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       anywherev1.ClusterKind,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "eksa-unit-test",
			},
		},
	}

	g.Expect(c.Validate(config)).To(MatchError(ContainSubstring(
		"invalid cluster config: [cluster namespace can't be empty, cluster kubernetes version can't be empty]",
	)))
}

func TestConfigManagerValidateMultipleSuccess(t *testing.T) {
	g := NewWithT(t)
	c := cluster.NewConfigManager()
	c.RegisterValidations(
		func(c *cluster.Config) error {
			if c.Cluster.Name == "" {
				return errors.New("cluster name can't be empty")
			}
			return nil
		},
	)

	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       anywherev1.ClusterKind,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "eksa-unit-test",
			},
		},
	}

	g.Expect(c.Validate(config)).To(Succeed())
}

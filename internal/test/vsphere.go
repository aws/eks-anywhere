package test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

// VSphereClusterSpec builds a complete and valid cluster spec for a vSphere cluster.
func VSphereClusterSpec(tb testing.TB, namespace string, opts ...ClusterSpecOpt) *cluster.Spec {
	bundle := Bundle()
	version := DevEksaVersion()
	machineConfigCP := VSphereMachineConfig(func(m *anywherev1.VSphereMachineConfig) {
		m.Name = "cp-machine-config"
		m.Namespace = namespace
	})
	machineConfigWN := VSphereMachineConfig(func(m *anywherev1.VSphereMachineConfig) {
		m.Name = "worker-machine-config"
		m.Namespace = namespace
	})

	workloadClusterDatacenter := VSphereDatacenter(func(d *anywherev1.VSphereDatacenterConfig) {
		d.Status.SpecValid = true
		d.Namespace = namespace
	})

	c := Cluster(func(c *anywherev1.Cluster) {
		c.Name = "my-c"
		c.Namespace = namespace
		c.Spec.BundlesRef = &anywherev1.BundlesRef{
			Name:       bundle.Name,
			Namespace:  bundle.Namespace,
			APIVersion: bundle.APIVersion,
		}
		c.Spec.ControlPlaneConfiguration = anywherev1.ControlPlaneConfiguration{
			Count: 1,
			Endpoint: &anywherev1.Endpoint{
				Host: "1.1.1.1",
			},
			MachineGroupRef: &anywherev1.Ref{
				Kind: anywherev1.VSphereMachineConfigKind,
				Name: machineConfigCP.Name,
			},
		}
		c.Spec.DatacenterRef = anywherev1.Ref{
			Kind: anywherev1.VSphereDatacenterKind,
			Name: workloadClusterDatacenter.Name,
		}

		c.Spec.WorkerNodeGroupConfigurations = append(c.Spec.WorkerNodeGroupConfigurations,
			anywherev1.WorkerNodeGroupConfiguration{
				Count: ptr.Int(1),
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.VSphereMachineConfigKind,
					Name: machineConfigWN.Name,
				},
				Name:   "md-0",
				Labels: nil,
			},
		)

		c.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{
			Cilium: &anywherev1.CiliumConfig{},
		}
		c.Spec.EksaVersion = &version
	})

	config := &cluster.Config{
		Cluster: c,
		VSphereMachineConfigs: map[string]*anywherev1.VSphereMachineConfig{
			machineConfigCP.Name: machineConfigCP,
			machineConfigWN.Name: machineConfigWN,
		},
		VSphereDatacenter: workloadClusterDatacenter,
	}

	spec, err := cluster.NewSpec(
		config,
		bundle,
		EksdReleases(),
		EKSARelease(),
	)
	if err != nil {
		tb.Fatalf("Failed to build cluster spec: %s", err)
	}

	for _, opt := range opts {
		opt(spec)
	}

	return spec
}

// VSphereMachineOpt allows to customize a VSphereMachineConfig.
type VSphereMachineOpt func(config *anywherev1.VSphereMachineConfig)

// VSphereMachineConfig builds a VSphereMachineConfig with some basic defaults.
func VSphereMachineConfig(opts ...VSphereMachineOpt) *anywherev1.VSphereMachineConfig {
	m := &anywherev1.VSphereMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.VSphereMachineConfigKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		Spec: anywherev1.VSphereMachineConfigSpec{
			DiskGiB:           40,
			Datastore:         "test",
			Folder:            "test",
			NumCPUs:           2,
			MemoryMiB:         16,
			OSFamily:          "ubuntu",
			ResourcePool:      "test",
			StoragePolicyName: "test",
			Template:          "test",
			Users: []anywherev1.UserConfiguration{
				{
					Name:              "user",
					SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8ZEibIrz1AUBKDvmDiWLs9f5DnOerC4qPITiDtSOuPAsxgZbRMavBfVTxodMdAkYRYlXxK6PqNo0ve0qcOV2yvpxH1OogasMMetck6BlM/dIoo3vEY4ZoG9DuVRIf9Iry5gJKbpMDYWpx1IGZrDMOFcIM20ii2qLQQk5hfq9OqdqhToEJFixdgJt/y/zt6Koy3kix+XsnrVdAHgWAq4CZuwt1G6JUAqrpob3H8vPmL7aS+35ktf0pHBm6nYoxRhslnWMUb/7vpzWiq+fUBIm2LYqvrnm7t3fRqFx7p2sZqAm2jDNivyYXwRXkoQPR96zvGeMtuQ5BVGPpsDfVudSW21+pEXHI0GINtTbua7Ogz7wtpVywSvHraRgdFOeY9mkXPzvm2IhoqNrteck2GErwqSqb19mPz6LnHueK0u7i6WuQWJn0CUoCtyMGIrowXSviK8qgHXKrmfTWATmCkbtosnLskNdYuOw8bKxq5S4WgdQVhPps2TiMSZndjX5NTr8= ubuntu@ip-10-2-0-19"},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// VSphereDatacenterConfigOpt allows to customize a VSphereDatacenterConfig.
type VSphereDatacenterConfigOpt func(config *anywherev1.VSphereDatacenterConfig)

// VSphereDatacenter builds a VSphereDatacenterConfig for tests with some basix defaults.
func VSphereDatacenter(opts ...VSphereDatacenterConfigOpt) *anywherev1.VSphereDatacenterConfig {
	d := &anywherev1.VSphereDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.VSphereDatacenterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "datacenter",
		},
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// VSphereCredentialsSecret builds a new Secret follwoing the format expected for vSphere credentials.
func VSphereCredentialsSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.EksaSystemNamespace,
			Name:      "vsphere-credentials",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Data: map[string][]byte{
			"username":   []byte("user"),
			"password":   []byte("pass"),
			"usernameCP": []byte("userCP"),
			"passwordCP": []byte("passCP"),
		},
	}
}

// ClusterOpt allows to customize a Cluster.
type ClusterOpt func(*anywherev1.Cluster)

// Cluster builds a Cluster for tests with some basic defaults.
func Cluster(opts ...ClusterOpt) *anywherev1.Cluster {
	c := &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.22",
			ClusterNetwork: anywherev1.ClusterNetwork{
				Pods: anywherev1.Pods{
					CidrBlocks: []string{"0.0.0.0"},
				},
				Services: anywherev1.Services{
					CidrBlocks: []string{"0.0.0.0"},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

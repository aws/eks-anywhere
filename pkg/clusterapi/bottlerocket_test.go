package clusterapi_test

import (
	"testing"

	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	. "github.com/onsi/gomega"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

var pause = bootstrapv1.Pause{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-distro/kubernetes/pause",
		ImageTag:        "0.0.1",
	},
}

var bootstrap = bootstrapv1.BottlerocketBootstrap{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap",
		ImageTag:        "0.0.1",
	},
}

var adminContainer = bootstrapv1.BottlerocketAdmin{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-anywhere/bottlerocket-admin",
		ImageTag:        "0.0.1",
	},
}

var controlContainer = bootstrapv1.BottlerocketControl{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-anywhere/bottlerocket-control",
		ImageTag:        "0.0.1",
	},
}

var kernel = &bootstrapv1.BottlerocketSettings{
	Kernel: &bootstrapv1.BottlerocketKernelSettings{
		SysctlSettings: map[string]string{
			"foo": "bar",
			"abc": "def",
		},
	},
}

func TestSetBottlerocketInKubeadmControlPlane(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmControlPlane()
	want := got.DeepCopy()
	want.Spec.KubeadmConfigSpec.Format = "bottlerocket"
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketBootstrap = bootstrap
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.Pause = pause
	want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketBootstrap = bootstrap
	want.Spec.KubeadmConfigSpec.JoinConfiguration.Pause = pause
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.ControllerManager.ExtraVolumes = append(want.Spec.KubeadmConfigSpec.ClusterConfiguration.ControllerManager.ExtraVolumes,
		bootstrapv1.HostPathMount{
			HostPath:  "/var/lib/kubeadm/controller-manager.conf",
			MountPath: "/etc/kubernetes/controller-manager.conf",
			Name:      "kubeconfig",
			PathType:  "File",
			ReadOnly:  true,
		},
	)
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.Scheduler.ExtraVolumes = append(want.Spec.KubeadmConfigSpec.ClusterConfiguration.Scheduler.ExtraVolumes,
		bootstrapv1.HostPathMount{
			HostPath:  "/var/lib/kubeadm/scheduler.conf",
			MountPath: "/etc/kubernetes/scheduler.conf",
			Name:      "kubeconfig",
			PathType:  "File",
			ReadOnly:  true,
		},
	)
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.CertificatesDir = "/var/lib/kubeadm/pki"

	bundle := g.clusterSpec.RootVersionsBundle()
	g.Expect(bundle).ToNot(BeNil())

	clusterapi.SetBottlerocketInKubeadmControlPlane(got, bundle)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketAdminContainerImageInKubeadmControlPlane(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmControlPlane()
	want := got.DeepCopy()
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketAdmin = adminContainer
	want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketAdmin = adminContainer

	bundle := g.clusterSpec.RootVersionsBundle()
	g.Expect(bundle).ToNot(BeNil())

	clusterapi.SetBottlerocketAdminContainerImageInKubeadmControlPlane(got, bundle)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketControlContainerImageInKubeadmControlPlane(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmControlPlane()
	want := got.DeepCopy()
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketControl = controlContainer
	want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketControl = controlContainer

	bundle := g.clusterSpec.RootVersionsBundle()
	g.Expect(bundle).ToNot(BeNil())

	clusterapi.SetBottlerocketControlContainerImageInKubeadmControlPlane(got, bundle)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketInKubeadmConfigTemplate(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmConfigTemplate()
	want := got.DeepCopy()
	want.Spec.Template.Spec.Format = "bottlerocket"
	want.Spec.Template.Spec.JoinConfiguration.BottlerocketBootstrap = bootstrap
	want.Spec.Template.Spec.JoinConfiguration.Pause = pause

	bundle := g.clusterSpec.RootVersionsBundle()
	g.Expect(bundle).ToNot(BeNil())

	clusterapi.SetBottlerocketInKubeadmConfigTemplate(got, bundle)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketAdminContainerImageInKubeadmConfigTemplate(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmConfigTemplate()
	want := got.DeepCopy()
	want.Spec.Template.Spec.JoinConfiguration.BottlerocketAdmin = adminContainer

	bundle := g.clusterSpec.RootVersionsBundle()
	g.Expect(bundle).ToNot(BeNil())

	clusterapi.SetBottlerocketAdminContainerImageInKubeadmConfigTemplate(got, bundle)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketControlContainerImageInKubeadmConfigTemplate(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmConfigTemplate()
	want := got.DeepCopy()
	want.Spec.Template.Spec.JoinConfiguration.BottlerocketControl = controlContainer

	bundle := g.clusterSpec.RootVersionsBundle()
	g.Expect(bundle).ToNot(BeNil())

	clusterapi.SetBottlerocketControlContainerImageInKubeadmConfigTemplate(got, bundle)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketInEtcdCluster(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantEtcdCluster()
	want := got.DeepCopy()
	want.Spec.EtcdadmConfigSpec.Format = etcdbootstrapv1.Format("bottlerocket")
	want.Spec.EtcdadmConfigSpec.BottlerocketConfig = &etcdbootstrapv1.BottlerocketConfig{
		EtcdImage:      "public.ecr.aws/eks-distro/etcd-io/etcd:0.0.1",
		BootstrapImage: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap:0.0.1",
		PauseImage:     "public.ecr.aws/eks-distro/kubernetes/pause:0.0.1",
	}
	bundle := g.clusterSpec.RootVersionsBundle()
	g.Expect(bundle).ToNot(BeNil())
	clusterapi.SetBottlerocketInEtcdCluster(got, bundle)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketAdminContainerImageInEtcdCluster(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantEtcdCluster()
	got.Spec.EtcdadmConfigSpec.BottlerocketConfig = &etcdbootstrapv1.BottlerocketConfig{
		EtcdImage:      "public.ecr.aws/eks-distro/etcd-io/etcd:0.0.1",
		BootstrapImage: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap:0.0.1",
		PauseImage:     "public.ecr.aws/eks-distro/kubernetes/pause:0.0.1",
	}
	want := got.DeepCopy()
	want.Spec.EtcdadmConfigSpec.BottlerocketConfig.AdminImage = "public.ecr.aws/eks-anywhere/bottlerocket-admin:0.0.1"
	bundle := g.clusterSpec.RootVersionsBundle()
	g.Expect(bundle).ToNot(BeNil())
	clusterapi.SetBottlerocketAdminContainerImageInEtcdCluster(got, bundle.BottleRocketHostContainers.Admin)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketControlContainerImageInEtcdCluster(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantEtcdCluster()
	got.Spec.EtcdadmConfigSpec.BottlerocketConfig = &etcdbootstrapv1.BottlerocketConfig{
		EtcdImage:      "public.ecr.aws/eks-distro/etcd-io/etcd:0.0.1",
		BootstrapImage: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap:0.0.1",
		PauseImage:     "public.ecr.aws/eks-distro/kubernetes/pause:0.0.1",
	}
	want := got.DeepCopy()
	want.Spec.EtcdadmConfigSpec.BottlerocketConfig.ControlImage = "public.ecr.aws/eks-anywhere/bottlerocket-control:0.0.1"
	bundle := g.clusterSpec.RootVersionsBundle()
	g.Expect(bundle).ToNot(BeNil())
	clusterapi.SetBottlerocketControlContainerImageInEtcdCluster(got, bundle.BottleRocketHostContainers.Control)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketHostConfigInKubeadmControlPlane(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmControlPlane()
	want := got.DeepCopy()
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.Bottlerocket = kernel
	want.Spec.KubeadmConfigSpec.JoinConfiguration.Bottlerocket = kernel

	clusterapi.SetBottlerocketHostConfigInKubeadmControlPlane(got, &anywherev1.HostOSConfiguration{
		BottlerocketConfiguration: &anywherev1.BottlerocketConfiguration{
			Kernel: &bootstrapv1.BottlerocketKernelSettings{
				SysctlSettings: map[string]string{
					"foo": "bar",
					"abc": "def",
				},
			},
		},
	})
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketHostConfigInKubeadmConfigTemplate(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmConfigTemplate()
	want := got.DeepCopy()
	want.Spec.Template.Spec.JoinConfiguration.Bottlerocket = kernel

	clusterapi.SetBottlerocketHostConfigInKubeadmConfigTemplate(got, &anywherev1.HostOSConfiguration{
		BottlerocketConfiguration: &anywherev1.BottlerocketConfiguration{
			Kernel: &bootstrapv1.BottlerocketKernelSettings{
				SysctlSettings: map[string]string{
					"foo": "bar",
					"abc": "def",
				},
			},
		},
	})
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketKernelSettingsInEtcdCluster(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantEtcdCluster()
	got.Spec.EtcdadmConfigSpec.BottlerocketConfig = &etcdbootstrapv1.BottlerocketConfig{
		EtcdImage:      "public.ecr.aws/eks-distro/etcd-io/etcd:0.0.1",
		BootstrapImage: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap:0.0.1",
		PauseImage:     "public.ecr.aws/eks-distro/kubernetes/pause:0.0.1",
	}
	want := got.DeepCopy()
	want.Spec.EtcdadmConfigSpec.BottlerocketConfig.Kernel = &bootstrapv1.BottlerocketKernelSettings{
		SysctlSettings: map[string]string{
			"foo": "bar",
			"abc": "def",
		},
	}

	clusterapi.SetBottlerocketHostConfigInEtcdCluster(got, &anywherev1.HostOSConfiguration{
		BottlerocketConfiguration: &anywherev1.BottlerocketConfiguration{
			Kernel: &bootstrapv1.BottlerocketKernelSettings{
				SysctlSettings: map[string]string{
					"foo": "bar",
					"abc": "def",
				},
			},
		},
	})
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketBootSettingsInKubeadmControlPlane(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmControlPlane()
	want := got.DeepCopy()
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.Bottlerocket = &bootstrapv1.BottlerocketSettings{
		Boot: &bootstrapv1.BottlerocketBootSettings{
			BootKernelParameters: map[string][]string{
				"foo": {
					"abc",
					"def",
				},
			},
		},
	}
	want.Spec.KubeadmConfigSpec.JoinConfiguration.Bottlerocket = &bootstrapv1.BottlerocketSettings{
		Boot: &bootstrapv1.BottlerocketBootSettings{
			BootKernelParameters: map[string][]string{
				"foo": {
					"abc",
					"def",
				},
			},
		},
	}

	clusterapi.SetBottlerocketHostConfigInKubeadmControlPlane(got, &anywherev1.HostOSConfiguration{
		BottlerocketConfiguration: &anywherev1.BottlerocketConfiguration{
			Boot: &bootstrapv1.BottlerocketBootSettings{
				BootKernelParameters: map[string][]string{
					"foo": {
						"abc",
						"def",
					},
				},
			},
		},
	})
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketBootSettingsInKubeadmConfigTemplate(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmConfigTemplate()
	want := got.DeepCopy()
	want.Spec.Template.Spec.JoinConfiguration.Bottlerocket = &bootstrapv1.BottlerocketSettings{
		Kernel: &bootstrapv1.BottlerocketKernelSettings{
			SysctlSettings: map[string]string{
				"foo": "bar",
				"abc": "def",
			},
		},
		Boot: &bootstrapv1.BottlerocketBootSettings{
			BootKernelParameters: map[string][]string{
				"foo": {
					"abc",
					"def",
				},
			},
		},
	}

	clusterapi.SetBottlerocketHostConfigInKubeadmConfigTemplate(got, &anywherev1.HostOSConfiguration{
		BottlerocketConfiguration: &anywherev1.BottlerocketConfiguration{
			Boot: &bootstrapv1.BottlerocketBootSettings{
				BootKernelParameters: map[string][]string{
					"foo": {
						"abc",
						"def",
					},
				},
			},
			Kernel: &bootstrapv1.BottlerocketKernelSettings{
				SysctlSettings: map[string]string{
					"foo": "bar",
					"abc": "def",
				},
			},
		},
	})
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketBootSettingsInEtcdCluster(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantEtcdCluster()
	got.Spec.EtcdadmConfigSpec.BottlerocketConfig = &etcdbootstrapv1.BottlerocketConfig{
		EtcdImage:      "public.ecr.aws/eks-distro/etcd-io/etcd:0.0.1",
		BootstrapImage: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap:0.0.1",
		PauseImage:     "public.ecr.aws/eks-distro/kubernetes/pause:0.0.1",
	}
	want := got.DeepCopy()
	want.Spec.EtcdadmConfigSpec.BottlerocketConfig.Kernel = &bootstrapv1.BottlerocketKernelSettings{
		SysctlSettings: map[string]string{
			"foo": "bar",
			"abc": "def",
		},
	}
	want.Spec.EtcdadmConfigSpec.BottlerocketConfig.Boot = &bootstrapv1.BottlerocketBootSettings{
		BootKernelParameters: map[string][]string{
			"foo": {
				"abc",
				"def",
			},
		},
	}

	clusterapi.SetBottlerocketHostConfigInEtcdCluster(got, &anywherev1.HostOSConfiguration{
		BottlerocketConfiguration: &anywherev1.BottlerocketConfiguration{
			Boot: &bootstrapv1.BottlerocketBootSettings{
				BootKernelParameters: map[string][]string{
					"foo": {
						"abc",
						"def",
					},
				},
			},
			Kernel: &bootstrapv1.BottlerocketKernelSettings{
				SysctlSettings: map[string]string{
					"foo": "bar",
					"abc": "def",
				},
			},
		},
	})

	g.Expect(got).To(Equal(want))
}

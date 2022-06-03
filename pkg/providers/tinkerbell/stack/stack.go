package stack

import (
	"context"
	"fmt"
	"path/filepath"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	args              = "args"
	createNamespace   = "createNamespace"
	deploy            = "deploy"
	env               = "env"
	image             = "image"
	namespace         = "namespace"
	overridesFileName = "tinkerbell-chart-overrides.yaml"

	boots          = "boots"
	hegel          = "hegel"
	tinkController = "tinkController"
	tinkServer     = "tinkServer"
	rufio          = "rufio"
	grpcPort       = "42113"

	// TODO: remove this once the chart is added to bundle
	helmChartOci     = "oci://public.ecr.aws/i7k6m1j7/tinkerbell/tinkerbell-chart"
	helmChartName    = "tinkerbell-chart"
	helmChartVersion = "0.1.0"
)

type Docker interface {
	ForceRemove(ctx context.Context, name string) error
	Run(ctx context.Context, image string, name string, cmd []string, flags ...string) error
}

type Helm interface {
	InstallChartWithValuesFile(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, valuesFilePath string) error
}

type Installer struct {
	docker          Docker
	filewriter      filewriter.FileWriter
	helm            Helm
	namespace       string
	createNamespace bool
	bootsOnDocker   bool
}

type InstallerOption func(s *Installer)

type StackInstaller interface {
	// WithNamespace(ns string, create bool) *Installer
	// WithBootsOnDocker() *Installer
	// WithBootsOnKubernetes() *Installer
	Install(ctx context.Context, bundle releasev1alpha1.TinkerbellStackBundle, tinkServerIP, kubeconfig string, opts ...InstallerOption) error
	UninstallLocal(ctx context.Context) error
}

func WithNamespace(ns string, create bool) InstallerOption {
	return func(s *Installer) {
		s.namespace = ns
		s.createNamespace = create
	}
}

func WithBootsOnDocker() InstallerOption {
	return func(s *Installer) {
		s.bootsOnDocker = true
	}
}

func WithBootsOnKubernetes() InstallerOption {
	return func(s *Installer) {
		s.bootsOnDocker = false
	}
}

func NewInstaller(docker Docker, filewriter filewriter.FileWriter, helm Helm) StackInstaller {
	return &Installer{
		docker:     docker,
		filewriter: filewriter,
		helm:       helm,
	}
}

// func (s *Installer) WithNamespace(ns string, create bool) *Installer {
// 	s.namespace = ns
// 	s.createNamespace = create
// 	return s
// }

// func (s *Installer) WithBootsOnDocker() *Installer {
// 	s.bootsOnDocker = true
// 	return s
// }

// func (s *Installer) WithBootsOnKubernetes() *Installer {
// 	s.bootsOnDocker = false
// 	return s
// }

func (s *Installer) Install(ctx context.Context, bundle releasev1alpha1.TinkerbellStackBundle, tinkServerIP, kubeconfig string, opts ...InstallerOption) error {
	logger.V(6).Info("Installing Tinkerbell helm chart")

	for _, option := range opts {
		option(s)
	}

	bootEnv := []map[string]string{}
	for k, v := range s.getBootsEnv(bundle, tinkServerIP) {
		bootEnv = append(bootEnv, map[string]string{
			"name":  k,
			"value": v,
		})
	}

	valuesMap := map[string]interface{}{
		namespace:       s.namespace,
		createNamespace: s.createNamespace,
		tinkController: map[string]interface{}{
			deploy: true,
			image:  bundle.Tink.TinkController.URI,
		},
		tinkServer: map[string]interface{}{
			deploy: true,
			image:  bundle.Tink.TinkServer.URI,
			args:   []string{"--tls=false"},
		},
		hegel: map[string]interface{}{
			deploy: true,
			image:  bundle.Hegel.Image.URI,
			args:   []string{"--grpc-use-tls=false"},
		},
		boots: map[string]interface{}{
			deploy: !s.bootsOnDocker,
			image:  bundle.Boots.Image.URI,
			env:    bootEnv,
		},
		rufio: map[string]interface{}{
			deploy: true,
			image:  bundle.Rufio.Image.URI,
		},
	}

	values, err := yaml.Marshal(valuesMap)
	if err != nil {
		return fmt.Errorf("marshalling values override for Tinkerbell Installer helm chart: %s", err)
	}

	valuesPath, err := s.filewriter.Write(overridesFileName, values)
	if err != nil {
		return fmt.Errorf("writing values override for Tinkerbell Installer helm chart: %s", err)
	}

	if err := s.helm.InstallChartWithValuesFile(ctx, helmChartName, helmChartOci, helmChartVersion, kubeconfig, valuesPath); err != nil {
		return fmt.Errorf("installing Tinkerbell helm chart: %v", err)
	}

	return s.installBootsOnDocker(ctx, bundle, tinkServerIP, kubeconfig)
}

func (s *Installer) installBootsOnDocker(ctx context.Context, bundle releasev1alpha1.TinkerbellStackBundle, tinkServerIP, kubeconfig string) error {
	if !s.bootsOnDocker {
		return nil
	}

	kubeconfig, err := filepath.Abs(kubeconfig)
	if err != nil {
		return fmt.Errorf("getting absolute path for kubeconfig: %v", err)
	}

	flags := []string{
		"-v", fmt.Sprintf("%s:/kubeconfig", kubeconfig),
		"--network", "host",
	}

	for name, value := range s.getBootsEnv(bundle, tinkServerIP) {
		flags = append(flags, "-e", fmt.Sprintf("%s=%s", name, value))
	}

	cmd := []string{"-kubeconfig", "/kubeconfig", "-dhcp-addr", "0.0.0.0:67"}
	if err := s.docker.Run(ctx, bundle.Boots.Image.URI, boots, cmd, flags...); err != nil {
		return fmt.Errorf("running boots with docker: %v", err)
	}

	return nil
}

func (s *Installer) getBootsEnv(bundle releasev1alpha1.TinkerbellStackBundle, tinkServerIP string) map[string]string {
	return map[string]string{
		"DATA_MODEL_VERSION":        "kubernetes",
		"TINKERBELL_TLS":            "false",
		"TINKERBELL_GRPC_AUTHORITY": fmt.Sprintf("%s:%s", tinkServerIP, grpcPort),
		"BOOTS_EXTRA_KERNEL_ARGS":   fmt.Sprintf("tink_worker_image=%s", bundle.Tink.TinkWorker.URI),
		// TODO: Pull this from bundle instead
		"MIRROR_BASE_URL": "https://tinkerbell-storage-for-eksa.s3.us-west-2.amazonaws.com",
	}
}

func (s *Installer) UninstallLocal(ctx context.Context) error {
	return s.uninstallBootsFromDocker(ctx)
}

func (s *Installer) uninstallBootsFromDocker(ctx context.Context) error {
	logger.V(4).Info("Removing local boots container")
	if err := s.docker.ForceRemove(ctx, boots); err != nil {
		return fmt.Errorf("removing local boots container: %v", err)
	}

	return nil
}

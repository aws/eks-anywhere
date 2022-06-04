package stack

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

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

	osiePath, err := getURIDir(bundle.Hook.Initramfs.Amd.URI)
	if err != nil {
		return fmt.Errorf("getting directory path from hook uri: %v", err)
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
			args: []string{
				"-dhcp-addr=0.0.0.0:67",
				fmt.Sprintf("-osie-path-override=%s", osiePath),
			},
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

	err = s.helm.InstallChartWithValuesFile(
		ctx,
		bundle.TinkebellChart.Name,
		fmt.Sprintf("oci://%s", bundle.TinkebellChart.Image()),
		bundle.TinkebellChart.Tag(),
		kubeconfig,
		valuesPath,
	)
	if err != nil {
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

	osiePath, err := getURIDir(bundle.Hook.Initramfs.Amd.URI)
	if err != nil {
		return fmt.Errorf("getting directory path from hook uri: %v", err)
	}

	cmd := []string{
		"-kubeconfig", "/kubeconfig",
		"-dhcp-addr", "0.0.0.0:67",
		"-osie-path-override", osiePath,
	}
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

func getURIDir(uri string) (string, error) {
	index := strings.LastIndex(uri, "/")
	if index == -1 {
		return "", fmt.Errorf("uri is invalid: %s", uri)
	}
	return uri[:index], nil
}

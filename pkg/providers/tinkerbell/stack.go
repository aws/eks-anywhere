package tinkerbell

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"sigs.k8s.io/yaml"
)

const (
	args            = "args"
	createNamespace = "createNamespace"
	deploy          = "deploy"
	env             = "env"
	image           = "image"
	namespace       = "namespace"

	boots          = "boots"
	hegel          = "hegel"
	tinkController = "tinkController"
	tinkServer     = "tinkServer"
	grpcPort       = "42113"

	// TODO: remove this once the chart is added to bundle
	helmChartOci     = "oci://public.ecr.aws/h6q6q4n4/tinkerbell"
	helmChartName    = "tinkerbell"
	helmChartVersion = "0.1.0"
)

type stack struct {
	bundle     releasev1alpha1.TinkerbellStackBundle
	docker     Docker
	filewriter filewriter.FileWriter
	helm       Helm
	ip         string
	values     map[string]interface{}
}

func newStack(bundle releasev1alpha1.TinkerbellStackBundle, docker Docker, filewriter filewriter.FileWriter, helm Helm, ip string) *stack {
	return &stack{
		bundle:     bundle,
		docker:     docker,
		filewriter: filewriter,
		helm:       helm,
		ip:         ip,
		values:     make(map[string]interface{}),
	}
}

func (s *stack) withBoots() *stack {
	environment := []map[string]string{}
	for k, v := range s.getBootsEnv() {
		environment = append(environment, map[string]string{
			"name":  k,
			"value": v,
		})
	}

	s.values[boots] = map[string]interface{}{
		deploy: true,
		image:  s.bundle.Boots.Image.URI,
		env:    environment,
	}
	return s
}

func (s *stack) withoutBoots() *stack {
	s.values[boots] = map[string]interface{}{
		deploy: false,
	}
	return s
}

func (s *stack) withHegel() *stack {
	s.values[hegel] = map[string]interface{}{
		deploy: true,
		image:  s.bundle.Hegel.Image.URI,
		args:   []string{"--grpc-use-tls=false"},
	}
	return s
}

func (s *stack) withTinkController() *stack {
	s.values[tinkController] = map[string]interface{}{
		deploy: true,
		image:  s.bundle.Tink.TinkController.URI,
	}
	return s
}

func (s *stack) withTinkServer() *stack {
	s.values[tinkServer] = map[string]interface{}{
		deploy: true,
		image:  s.bundle.Tink.TinkServer.URI,
		args:   []string{"--tls=false"},
	}
	return s
}

func (s *stack) withNamespace(ns string, create bool) *stack {
	s.values[namespace] = ns
	s.values[createNamespace] = create
	return s
}

func (s *stack) getValues() []string {
	values := []string{}
	for k, v := range s.values {
		values = append(values, fmt.Sprintf("%s=%s", k, v))
	}
	return values
}

func (s *stack) install(ctx context.Context, kubeconfig string) error {
	logger.V(6).Info("Installing Tinkerbell helm chart")

	values, err := yaml.Marshal(s.values)
	if err != nil {
		return fmt.Errorf("marshalling values override for Tinkerbell stack helm chart: %s", err)
	}

	valuesPath, err := s.filewriter.Write("tinkerbell-chart-values.yaml", values)
	if err != nil {
		return fmt.Errorf("writing values override for Tinkerbell stack helm chart: %s", err)
	}

	if err := s.helm.InstallChartWithValuesFile(ctx, helmChartName, helmChartOci, helmChartVersion, kubeconfig, valuesPath); err != nil {
		return fmt.Errorf("installing Tinkerbell helm chart: %v", err)
	}
	return nil
}

func (s *stack) installBootsOnDocker(ctx context.Context, kubeconfig string) error {
	kubeconfig, err := filepath.Abs(kubeconfig)
	if err != nil {
		return fmt.Errorf("getting absolute path for kubeconfig: %v", err)
	}

	flags := []string{
		"-v", fmt.Sprintf("%s:/kubeconfig", kubeconfig),
		"--network", "host",
	}

	for name, value := range s.getBootsEnv() {
		flags = append(flags, "-e", fmt.Sprintf("%s=%s", name, value))
	}

	cmd := []string{"-kubeconfig", "/kubeconfig", "-dhcp-addr", "0.0.0.0:67"}
	if err := s.docker.Run(ctx, s.bundle.Boots.Image.URI, boots, cmd, flags...); err != nil {
		return fmt.Errorf("run boots with docker: %v", err)
	}

	return nil
}

func (s *stack) getBootsEnv() map[string]string {
	return map[string]string{
		"DATA_MODEL_VERSION":        "kubernetes",
		"TINKERBELL_TLS":            "false",
		"TINKERBELL_GRPC_AUTHORITY": fmt.Sprintf("%s:%s", s.ip, grpcPort),
		"BOOTS_EXTRA_KERNEL_ARGS":   fmt.Sprintf("tink_worker_image=%s", s.bundle.Tink.TinkWorker.URI),
		// TODO: Pull this from bundle instead
		"MIRROR_BASE_URL": "https://tinkerbell-storage-for-eksa.s3.us-west-2.amazonaws.com",
	}
}

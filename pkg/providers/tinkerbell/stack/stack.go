package stack

import (
	_ "embed"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed manifests/database.yaml
var databaseManifest string

//go:embed manifests/tink.yaml
var tinkManifest string

//go:embed manifests/hegel.yaml
var hegelManifest string

//go:embed manifests/pbnj.yaml
var pbnjManifest string

const defaultEksaNamespace = "eksa-system"

type TinkerbellStack struct {
	values map[string]interface{}
}

func NewTinkerbellStack(bundle cluster.VersionsBundle, tinkerbellIp string) *TinkerbellStack {
	return &TinkerbellStack{
		values: map[string]interface{}{
			"tinkServerImage":  bundle.Tinkerbell.TinkServer.URI,
			"pbnjImage":        bundle.Tinkerbell.Pbnj.URI,
			"hegelImage":       bundle.Tinkerbell.Hegel.URI,
			"namespace":        defaultEksaNamespace,
			"tinkerbellHostIp": tinkerbellIp,
			"tinkGrpcPort":     42113,
			"tinkCertPort":     42114,
			"pbnjGrpcPort":     50051,
			"hegelPort":        50061,
		},
	}
}

func (s *TinkerbellStack) GenerateDatabaseManifest() ([]byte, error) {
	return templater.Execute(databaseManifest, s.values)
}

func (s *TinkerbellStack) GenerateTinkManifest() ([]byte, error) {
	return templater.Execute(tinkManifest, s.values)
}

func (s *TinkerbellStack) GenerateHegelManifest() ([]byte, error) {
	return templater.Execute(hegelManifest, s.values)
}

func (s *TinkerbellStack) GeneratePbnjManifest() ([]byte, error) {
	return templater.Execute(pbnjManifest, s.values)
}

package framework

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed testdata/emissary_route.yaml
var emissaryRouteTemplate string

// ApplyEmissaryRoute creates an Emissary Listener, Host and Mapping in namespace that route
// http://<Emissary service>/<name>/ to port 80 of service, which must be in the same namespace.
// The Emissary package must be installed; see EmissaryRouteURL for the URL.
func (e *ClusterE2ETest) ApplyEmissaryRoute(name, namespace, service string) {
	manifest, err := templater.Execute(emissaryRouteTemplate, map[string]interface{}{
		"name":      name,
		"namespace": namespace,
		"service":   service,
	})
	if err != nil {
		e.T.Fatalf("Rendering Emissary route %s: %v", name, err)
	}

	// Right after the install, the getambassador.io CRDs, or Emissary's conversion webhook when it's
	// enabled, may not serve requests yet, so retry.
	e.T.Logf("Creating Emissary route /%s/ to service %s/%s", name, namespace, service)
	err = retrier.New(5*time.Minute, retrier.WithRetryPolicy(retrier.BackOffPolicy(10*time.Second))).Retry(func() error {
		return e.KubectlClient.ApplyKubeSpecFromBytes(context.Background(), e.Cluster(), manifest)
	})
	if err != nil {
		e.T.Fatalf("Creating Emissary route %s: %v", name, err)
	}
}

// EmissaryRouteURL returns the in-cluster URL of a route created with ApplyEmissaryRoute. emissaryPackage is
// the Package name of Emissary, which is also the name of its Service.
func EmissaryRouteURL(emissaryPackage, namespace, route string) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local/%s/", emissaryPackage, namespace, route)
}

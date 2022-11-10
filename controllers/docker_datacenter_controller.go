package controllers

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DockerDatacenterReconciler reconciles a DockerDatacenterConfig object.
type DockerDatacenterReconciler struct {
	log    logr.Logger
	client client.Client
}

// NewDockerDatacenterReconciler creates a new instance of the DockerDatacenterReconciler struct.
func NewDockerDatacenterReconciler(client client.Client, log logr.Logger) *DockerDatacenterReconciler {
	return &DockerDatacenterReconciler{
		client: client,
		log:    log,
	}
}

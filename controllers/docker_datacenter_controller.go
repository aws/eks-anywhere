package controllers

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DockerDatacenterReconciler reconciles a DockerDatacenterConfig object.
type DockerDatacenterReconciler struct {
	client client.Client
}

// NewDockerDatacenterReconciler creates a new instance of the DockerDatacenterReconciler struct.
func NewDockerDatacenterReconciler(client client.Client) *DockerDatacenterReconciler {
	return &DockerDatacenterReconciler{
		client: client,
	}
}

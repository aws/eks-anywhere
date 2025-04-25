package kubernetes

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Object is a Kubernetes object.
type Object client.Object

// ObjectList is a Kubernetes object list.
type ObjectList client.ObjectList

// Patch is controller runtime client patch type.
type Patch client.Patch

// Client is Kubernetes API client.
type Client interface {
	Reader
	Writer
}

// Reader knows how to read and list Kubernetes objects.
type Reader interface {
	// Get retrieves an obj for the given name and namespace from the Kubernetes Cluster.
	Get(ctx context.Context, name, namespace string, obj Object) error

	// List retrieves list of objects. On a successful call, Items field
	// in the list will be populated with the result returned from the server.
	List(ctx context.Context, list ObjectList, opts ...ListOption) error
}

// ListOption is some configuration that modifies options for an apply request.
type ListOption interface {
	ApplyToList(*ListOptions)
}

// ListOptions contains options for listing object.
type ListOptions struct {
	// Namespace represents the namespace to list for, or empty for
	// non-namespaced objects, or to list across all namespaces.
	Namespace string
}

// ApplyToList implements ApplyToList.
func (o ListOptions) ApplyToList(do *ListOptions) {
	if o.Namespace != "" {
		do.Namespace = o.Namespace
	}
}

// Writer knows how to create, delete, and update Kubernetes objects.
type Writer interface {
	// Create saves the object obj in the Kubernetes cluster.
	Create(ctx context.Context, obj Object) error

	// Update updates the given obj in the Kubernetes cluster.
	Update(ctx context.Context, obj Object) error

	// ApplyServerSide creates or patches and object using server side logic.
	ApplyServerSide(ctx context.Context, fieldManager string, obj Object, opts ...ApplyServerSideOption) error

	// Delete deletes the given obj from Kubernetes cluster.
	Delete(ctx context.Context, obj Object) error

	// DeleteAllOf deletes all objects of the given type matching the given options.
	DeleteAllOf(ctx context.Context, obj Object, opts ...DeleteAllOfOption) error

	// Patch patches the given obj in the Kubernetes cluster. obj must be a
	// struct pointer so that obj can be updated with the content returned by the Server.
	Patch(ctx context.Context, obj Object, patch Patch, opts ...PatchOption) error
}

// PatchOption is some configuration that modifies options for a patch request.
type PatchOption interface {
	// ApplyToPatch applies this configuration to the given patch options.
	ApplyToPatch(*PatchOptions)
}

// PatchOptions contains options for patch requests.
type PatchOptions struct {
	// FieldManager is the name of the manager used to track field ownership.
	FieldManager string

	// Force forces the patch to be applied even if it would change the resource's
	// current spec to a state that is in conflict with other managers.
	Force bool

	// DryRun specifies that the patch request is dry run. The request is processed
	// normally, but changes are not persisted.
	DryRun []string
}

var _ PatchOption = &PatchOptions{}

// ApplyToPatch implements PatchOption.
func (o *PatchOptions) ApplyToPatch(po *PatchOptions) {
	if o.FieldManager != "" {
		po.FieldManager = o.FieldManager
	}
	if o.Force {
		po.Force = true
	}
	if o.DryRun != nil {
		po.DryRun = o.DryRun
	}
}

// DeleteAllOfOption is some configuration that modifies options for a delete request.
type DeleteAllOfOption interface {
	// ApplyToDeleteAllOf applies this configuration to the given deletecollection options.
	ApplyToDeleteAllOf(*DeleteAllOfOptions)
}

// DeleteAllOfOptions contains options for deletecollection (deleteallof) requests.
type DeleteAllOfOptions struct {
	// HasLabels filters results by label and value. The requirement is an AND match
	// for all labels.
	HasLabels map[string]string

	// Namespace represents the namespace to list for, or empty for
	// non-namespaced objects, or to list across all namespaces.
	Namespace string
}

var _ DeleteAllOfOption = &DeleteAllOfOptions{}

// ApplyToDeleteAllOf implements DeleteAllOfOption.
func (o *DeleteAllOfOptions) ApplyToDeleteAllOf(do *DeleteAllOfOptions) {
	if o.HasLabels != nil {
		do.HasLabels = o.HasLabels
	}
	if o.Namespace != "" {
		do.Namespace = o.Namespace
	}
}

// ApplyServerSideOption is some configuration that modifies options for an apply request.
type ApplyServerSideOption interface {
	ApplyToApplyServerSide(*ApplyServerSideOptions)
}

// ApplyServerSideOptions contains options for server side apply requests.
type ApplyServerSideOptions struct {
	// ForceOwnership indicates that in case of conflicts with server-side apply,
	// the client should acquire ownership of the conflicting field.
	ForceOwnership bool
}

var _ ApplyServerSideOption = ApplyServerSideOptions{}

// ApplyToApplyServerSide implements ApplyServerSideOption.
func (o ApplyServerSideOptions) ApplyToApplyServerSide(do *ApplyServerSideOptions) {
	if o.ForceOwnership {
		do.ForceOwnership = true
	}
}

package kubernetes

import (
	"context"
	"io/fs"
	"os"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/logger"
)

var _ kubernetes.Reader = &BundleReader{}

// BundleReader is a Kubernetes Client reader that reads objects from a support bundle.
// It is not thread-safe.
type BundleReader struct {
	cache  fsObjectsStorage
	scheme *runtime.Scheme
}

// NewBundleReaderForFolder returns a new BundleReader that reads objects from files in disk from the given folder.
func NewBundleReaderForFolder(folder string) *BundleReader {
	return NewBundleReader(os.DirFS(folder))
}

// NewBundleReader returns a new BundleReader.
func NewBundleReader(fs fs.FS) *BundleReader {
	scheme := runtime.NewScheme()
	return &BundleReader{
		scheme: scheme,
		cache:  newStorage(scheme, fs),
	}
}

// Init initializes the client. It has always be invoked at least once before making any API call.
// It is not thread safe.
func (b *BundleReader) Init() error {
	logger.V(3).Info("Reading kubernetes resources from bundle")
	start := time.Now()
	if err := kubernetes.InitScheme(b.scheme); err != nil {
		return err
	}

	if err := b.cache.init(); err != nil {
		return errors.Wrap(err, "initializing storage")
	}

	logger.V(3).Info("Initialized", "took", time.Since(start))

	return nil
}

// Get retrieves an obj for the given name and namespace from the support bundle.
func (b *BundleReader) Get(_ context.Context, name, namespace string, obj kubernetes.Object) error {
	gvk, err := apiutil.GVKForObject(obj, b.scheme)
	if err != nil {
		return err
	}

	collection := b.cache.collection(gvk)
	if collection == nil {
		return newNotFoundErr(gvk, name, namespace)
	}

	if collection.namespaced && namespace == "" {
		return errors.Errorf("kubernetes object %s/%s/%s is namespaced, but namespace is not provided", gvk.String(), namespace, name)
	} else if !collection.namespaced && namespace != "" {
		return errors.Errorf("kubernetes object %s/%s/%s is cluster scoped, but namespace is provided", gvk.String(), namespace, name)
	}

	var objs interface {
		get(name string) kubernetes.Object
	}

	if namespace == "" {
		objs = collection.clusterScopedObjects()
	} else {
		nsObjs := collection.namespace(namespace)

		if nsObjs == nil {
			return newNotFoundErr(gvk, name, namespace)
		}

		objs = nsObjs
	}

	o := objs.get(name)
	if o == nil {
		return newNotFoundErr(gvk, name, namespace)
	}

	copyObj := o.DeepCopyObject()
	reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(copyObj).Elem())

	return nil
}

func newNotFoundErr(gvk schema.GroupVersionKind, name, namespace string) error {
	// TODO: make this compliant with client errors api from apimachinery
	return errors.Errorf("kubernetes object %s/%s/%s not found", gvk.String(), namespace, name)
}

// List retrieves list of objects from the support bundle.
func (b *BundleReader) List(_ context.Context, obj kubernetes.ObjectList, opts ...kubernetes.ListOption) error {
	listGVK, err := apiutil.GVKForObject(obj, b.scheme)
	if err != nil {
		return err
	}

	singleGVK := listGVK

	singleGVK.Kind = kubernetes.TrimListFromKind(listGVK.Kind)

	t, ok := b.scheme.AllKnownTypes()[singleGVK]
	if !ok {
		return errors.Errorf("kubernetes object %s part of list %s is not registered in the scheme", singleGVK, listGVK)
	}

	collection := b.cache.collection(singleGVK)
	if collection == nil {
		return nil
	}

	options := &kubernetes.ListOptions{}
	for _, opt := range opts {
		opt.ApplyToList(options)
	}

	objs := collection.all()
	items := reflect.MakeSlice(reflect.SliceOf(t), 0, len(objs))
	for _, o := range objs {
		if options.LabelSelector != nil && !options.LabelSelector.Matches(labels.Set(o.GetLabels())) {
			continue
		}
		copyObj := o.DeepCopyObject()
		items = reflect.Append(items, reflect.ValueOf(copyObj).Elem())
	}

	reflect.Indirect(reflect.ValueOf(obj)).FieldByName("Items").Set(items)

	return nil
}

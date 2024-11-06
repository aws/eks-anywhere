package kubernetes

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/logger"
)

func newStorage(scheme *runtime.Scheme, fs fs.FS) fsObjectsStorage {
	return fsObjectsStorage{
		byGVK:  make(map[schema.GroupVersionKind]objectsCollection),
		scheme: scheme,
		fs:     fs,
	}
}

type fsObjectsStorage struct {
	byGVK  map[schema.GroupVersionKind]objectsCollection
	scheme *runtime.Scheme
	fs     fs.FS
}

type objectsCollection struct {
	namespaced  bool
	gvk         schema.GroupVersionKind
	byNamespace map[string]namespaceObjects
}

type namespaceObjects struct {
	objects
	namespace string
}

type objects struct {
	byName map[string]kubernetes.Object
	gvk    schema.GroupVersionKind
}

// init initializes the storage by reading all the files in the fs and creating the objects.
func (s *fsObjectsStorage) init() error {
	entries, err := fs.ReadDir(s.fs, ".")
	if err != nil {
		return err
	}

	var coreClusterScoped, coreNamespaced, customClusterScoped, customNamespace []fsItem

	for _, entry := range entries {
		if entry.IsDir() {
			if !validResource(entry.Name()) {
				continue
			}

			if entry.Name() == "custom-resources" {
				customNamespace, customClusterScoped, err = customResourceEntries(s.fs, entry)
				if err != nil {
					return err
				}
			} else {
				coreNamespaced = append(coreNamespaced, fsItem{DirEntry: entry})
			}
		} else if validResourceFile(entry) {
			coreClusterScoped = append(coreClusterScoped, fsItem{DirEntry: entry})
		}
	}

	// We are reading all the files and loading them into memory. This increases the command "start up" time
	// and uses more memory. However, it simplifies a lot the Get and List code, since it's easy to know what
	// exists and what doesn't and it avoid dealing with partially initialized namespaced collections. In addition,
	// once initialized, the collection is thread safe.
	// However, this could present a challenge if the number of objects is too big. In that case, we could consider
	// loading the objects lazily and paying the price in complexity.

	for _, entry := range coreClusterScoped {
		collection, err := s.processClusterScoped(s.fs, entry, s.processListObjectResourceFile)
		if err != nil {
			return err
		}

		if collection == nil {
			continue
		}

		s.byGVK[collection.gvk] = *collection
	}

	for _, entry := range coreNamespaced {
		collection, err := s.processNamespaced(s.fs, entry, s.processListObjectResourceFile)
		if err != nil {
			return err
		}

		s.byGVK[collection.gvk] = *collection
	}

	for _, entry := range customClusterScoped {
		collection, err := s.processClusterScoped(s.fs, entry, s.processListOfObjectsResourceFile)
		if err != nil {
			return err
		}

		if collection == nil {
			continue
		}

		s.byGVK[collection.gvk] = *collection
	}

	for _, entry := range customNamespace {
		collection, err := s.processNamespaced(s.fs, entry, s.processListOfObjectsResourceFile)
		if err != nil {
			return err
		}

		s.byGVK[collection.gvk] = *collection
	}
	return nil
}

func customResourceEntries(f fs.FS, folder fs.DirEntry) (namespaced, clusterScoped []fsItem, err error) {
	customResourceEntries, err := fs.ReadDir(f, folder.Name())
	if err != nil {
		return nil, nil, err
	}

	for _, customResourceEntry := range customResourceEntries {
		i := fsItem{
			DirEntry: customResourceEntry,
			path:     folder.Name(),
		}
		if customResourceEntry.IsDir() {
			namespaced = append(namespaced, i)
		} else {
			clusterScoped = append(clusterScoped, i)
		}
	}

	return namespaced, clusterScoped, nil
}

type fsItem struct {
	fs.DirEntry
	path string
}

func (f fsItem) fullPath() string {
	return filepath.Join(f.path, f.Name())
}

type fileProcessor func(f fs.FS, file fsItem) (*objects, error)

// processClusterScoped processes a cluster scoped resource file and returns non-namespaced object collection.
func (s *fsObjectsStorage) processClusterScoped(f fs.FS, file fsItem, process fileProcessor) (*objectsCollection, error) {
	objs, err := process(f, file)
	if err != nil {
		return nil, err
	}

	if objs == nil {
		return nil, nil
	}

	return &objectsCollection{
		namespaced: false,
		gvk:        objs.gvk,
		byNamespace: map[string]namespaceObjects{
			"": {
				objects: *objs,
			},
		},
	}, nil
}

// processNamespaced processes a folder containing namespaced resource files and returns a namespaced object collection.
// It is expected that name of each file is the namespace of the objects in the file.
// It will only process .json files and will ignore any subdirectories.
func (s *fsObjectsStorage) processNamespaced(f fs.FS, folder fsItem, processor fileProcessor) (*objectsCollection, error) {
	collection := &objectsCollection{
		namespaced:  true,
		byNamespace: make(map[string]namespaceObjects),
	}

	files, err := fs.ReadDir(f, folder.fullPath())
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			logger.V(4).Info("Ignoring directory", "path", filepath.Join(folder.fullPath(), file.Name()))
			continue
		}

		objs, err := processor(f, fsItem{DirEntry: file, path: folder.fullPath()})
		if err != nil {
			return nil, err
		}

		if objs == nil {
			// TODO: if this returns nil, probably all files will
			// We need a way to signal the caller that this collection is empty
			continue
		}

		namespace := strings.TrimSuffix(file.Name(), ".json")

		collection.byNamespace[namespace] = namespaceObjects{
			namespace: namespace,
			objects:   *objs,
		}

		if collection.gvk.Empty() {
			collection.gvk = objs.gvk
		} else if collection.gvk != objs.gvk {
			return nil, errors.Errorf("kubernetes object kind %s for file %s is not the same as the other objects in the folder %s", objs.gvk, file.Name(), folder.Name())
		}
	}

	return collection, nil
}

// objectList is generic list Object to help deserialize an unknown type.
type objectList struct {
	Kind       string                 `json:"kind"`
	APIVersion string                 `json:"apiVersion"`
	Items      []runtime.RawExtension `json:"items"`
}

// processListObjectResourceFile processes a file containing kubernetes resources and returns the parsed objects.
// It expects the file to contain a List object (with a .items field).
func (s *fsObjectsStorage) processListObjectResourceFile(f fs.FS, file fsItem) (*objects, error) {
	content, err := fs.ReadFile(f, file.fullPath())
	if err != nil {
		return nil, err
	}

	var bList objectList
	if err := json.Unmarshal(content, &bList); err != nil {
		return nil, errors.Wrapf(err, "can't unmarshal basic list for file %s", file.fullPath())
	}

	singleResourceKind := kubernetes.TrimListFromKind(bList.Kind)
	if singleResourceKind == bList.Kind {
		return nil, errors.Errorf("kubernetes object type for file %s is not a list, it can't be processed", file.fullPath())
	}

	gvk := schema.FromAPIVersionAndKind(bList.APIVersion, singleResourceKind)

	t, ok := s.scheme.AllKnownTypes()[gvk]
	if !ok {
		return nil, errors.Errorf("kubernetes object kind %s for file %s is not known, it can't be processed", gvk.Kind, file.fullPath())
	}

	objs := make(map[string]kubernetes.Object, len(bList.Items))

	for _, item := range bList.Items {
		o := reflect.New(t).Interface().(kubernetes.Object)
		if err := json.Unmarshal(item.Raw, o); err != nil {
			return nil, errors.Wrapf(err, "can't unmarshal object for file %s and kind %s", file.fullPath(), gvk.Kind)
		}

		objs[o.GetName()] = o
	}

	return &objects{
		byName: objs,
		gvk:    gvk,
	}, nil
}

// basicObject is a barebones k8s object to help unmarshal an object of unknown type.
type basicObject struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
}

// processListOfObjectsResourceFile processes a file containing a list of kubernetes resources and returns the parsed objects.
// It expects the file to contain a list of k8s objects as the top level json object.
// If the resource kind has not been registered in the scheme, it will return nil.
func (s *fsObjectsStorage) processListOfObjectsResourceFile(f fs.FS, file fsItem) (*objects, error) {
	if !validResourceFile(file) {
		return nil, nil
	}

	content, err := fs.ReadFile(f, file.fullPath())
	if err != nil {
		return nil, err
	}

	var bList []basicObject
	if err := json.Unmarshal(content, &bList); err != nil {
		return nil, errors.Wrapf(err, "can't unmarshal custom resource file %s into basic list", file.fullPath())
	}

	if len(bList) == 0 {
		return nil, errors.Errorf("file %s is not a list or it's empty, it can't be processed", file.fullPath())
	}

	gvk := schema.FromAPIVersionAndKind(bList[0].APIVersion, bList[0].Kind)

	t, ok := s.scheme.AllKnownTypes()[gvk]
	if !ok {
		logger.V(4).Info("Ignoring unknown kind", "kind", gvk.Kind, "file", file.fullPath())
		return nil, nil
	}

	list := make([]runtime.RawExtension, len(bList))
	if err := json.Unmarshal(content, &list); err != nil {
		return nil, errors.Wrapf(err, "can't unmarshal custom resource file %s into raw object list", file.fullPath())
	}

	objs := make(map[string]kubernetes.Object, len(list))

	for _, obj := range list {
		o := reflect.New(t).Interface().(kubernetes.Object)
		if err := json.Unmarshal(obj.Raw, o); err != nil {
			return nil, errors.Wrapf(err, "can't unmarshal object for file %s and kind %s", file.fullPath(), gvk.Kind)
		}
		objs[o.GetName()] = o
	}

	return &objects{
		byName: objs,
		gvk:    gvk,
	}, nil
}

// validResourceFile returns true if the file should be processed.
func validResourceFile(file fs.DirEntry) bool {
	name := file.Name()
	// We ignore CRD definitions for now. If we ever need the, we just first need to add them to the scheme.
	// We ignore groups.json since we don't need them and would require special parsing.
	// We ignore resources.json since we don't need them and would require special parsing.
	// We ignore non json files since right now we make the assumption that all resource files are in json.
	// We ignore error files since they don't contain k8s objects, there are just artifacts from troubleshoot.sh
	return name != "custom-resource-definitions.json" &&
		name != "groups.json" &&
		name != "resources.json" &&
		strings.HasSuffix(name, ".json") &&
		!strings.HasSuffix(name, "-errors.json")
}

// validResource returns true if the folder should be processed.
func validResource(folder string) bool {
	// We ignore auth-cani-lis since we don't need them and would require special handling.
	// We ignore image-pull-secrets since it's not a k8s resource.
	return folder != "auth-cani-list" &&
		folder != "image-pull-secrets"
}

func (s *fsObjectsStorage) collection(gvk schema.GroupVersionKind) *objectsCollection {
	if collection, ok := s.byGVK[gvk]; ok {
		return &collection
	}
	return nil
}

func (c *objectsCollection) namespace(namespace string) *namespaceObjects {
	if !c.namespaced {
		panic("namespace called on non-namespaced collection")
	}
	if objs, ok := c.byNamespace[namespace]; ok {
		return &objs
	}
	return nil
}

func (c *objectsCollection) clusterScopedObjects() *objects {
	if c.namespaced {
		panic("clusterScopedObjects called on namespaced collection")
	}
	if objs, ok := c.byNamespace[""]; ok {
		return &objs.objects
	}
	return nil
}

func (c *objectsCollection) all() []kubernetes.Object {
	var all []kubernetes.Object
	for _, nsObjs := range c.byNamespace {
		all = append(all, nsObjs.objects.all()...)
	}
	return all
}

func (o *objects) get(name string) kubernetes.Object {
	if obj, ok := o.byName[name]; ok {
		return obj
	}
	return nil
}

func (o *objects) all() []kubernetes.Object {
	var all []kubernetes.Object
	for _, obj := range o.byName {
		all = append(all, obj)
	}
	return all
}

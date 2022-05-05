package hardware

import (
	"fmt"
	"reflect"
)

// FieldIndexer indexes collection of objects for a single type against one of its fields.
// FieldIndexer is not thread safe.
type FieldIndexer struct {
	expectedType reflect.Type
	indexes      map[string]*fieldIndex
}

// NewFieldIndexer creates a new FieldIndexer instance. object is the object to be indexed and will
// be checked during Insert() calls. NewFieldIndexer will panic if object is nil.
func NewFieldIndexer(object interface{}) *FieldIndexer {
	objectType := reflect.TypeOf(object)
	if objectType == nil {
		panic("object cannot be nil")
	}

	return &FieldIndexer{
		expectedType: objectType,
		indexes:      make(map[string]*fieldIndex),
	}
}

// KeyExtractorFunc returns a key from object that can be used to look up the object.
type KeyExtractorFunc func(object interface{}) string

// IndexField registers a new index with i. field is the index name and should represent a path
// to the field such as `.Spec.ID`. fn is used to extract the lookup key on Insert() from the object
// to be inserted.
func (i *FieldIndexer) IndexField(field string, fn KeyExtractorFunc) {
	i.indexes[field] = &fieldIndex{
		index:            make(map[string][]interface{}),
		keyExtractorFunc: fn,
	}
}

// Insert inserts v into i on all indexed fields registered with IndexField. If v is not of the
// expected type defined by NewFieldIndexer() ErrIncorrectType is returned. Multiple objects
// with the same index value may be inserted.
func (i *FieldIndexer) Insert(v interface{}) error {
	objectType := reflect.TypeOf(v)
	if objectType != i.expectedType {
		return ErrIncorrectType{Expected: i.expectedType, Received: objectType}
	}

	for _, idx := range i.indexes {
		idx.Insert(v)
	}

	return nil
}

// Lookup uses the index associated with field to find and return all objects associated with key.
// If field has no associated index created by IndexField, Lookup panics.
func (i *FieldIndexer) Lookup(field string, key string) []interface{} {
	idx, ok := i.indexes[field]
	if !ok {
		// Indexes shouldn't be missing. If they are, the developer has messed something up.
		panic("trying to access unknown index")
	}
	return idx.Lookup(key)
}

// Remove removes v from all indexes if present. If v is not present Remove is a no-op. If v is of
// an incorrect type Remove is a no-op for convinience.
func (i *FieldIndexer) Remove(v interface{}) {
	objectType := reflect.TypeOf(v)
	if objectType != i.expectedType {
		return
	}

	for _, idx := range i.indexes {
		idx.Remove(v)
	}
}

// fieldIndex represents a single index on a particular object. When inserting into the fieldIndex
// the key is extracted from the object using the KeyExtractorFunc.
type fieldIndex struct {
	index            map[string][]interface{}
	keyExtractorFunc KeyExtractorFunc
}

func (i *fieldIndex) Insert(v interface{}) {
	key := i.keyExtractorFunc(v)
	i.index[key] = append(i.index[key], v)
}

func (i *fieldIndex) Lookup(key string) []interface{} {
	return i.index[key]
}

func (i *fieldIndex) Remove(v interface{}) {
	key := i.keyExtractorFunc(v)
	delete(i.index, key)
}

// ErrIncorrectType indicates an incorrect type was used with a FieldIndexer.
type ErrIncorrectType struct {
	Expected reflect.Type
	Received reflect.Type
}

func (e ErrIncorrectType) Error() string {
	return fmt.Sprintf("expected type '%s', received object of type '%v'", e.Expected, e.Received)
}

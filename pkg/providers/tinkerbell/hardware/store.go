package hardware

type IndexLookup interface {
	// Lookup retrieves objects associated with the index => value pair.
	Lookup(index, value string) ([]interface{}, error)
}

// Indexer provides indexing behavior for objects.
type Indexer interface {
	IndexLookup
	// IndexField associated index with fn such that Lookup may be used to retrieve objects.
	IndexField(index string, fn KeyExtractorFunc)
	// Insert inserts v int the index.
	Insert(v interface{}) error
}

// IndexedStore is an object store providing efficient lookup capabilities through an Indexer.
type IndexedStore interface {
	Indexer
	// All retrieves all objects in the store as a slice. Modifying the returned slice has
	// undefined behavior.
	All() []interface{}
	// Size returns the total number of objects in the store.
	Size() int
}

// indexedStore is a generic object store that can be used for efficient lookups based on consumer
// defined indexes.
type indexedStore struct {
	Indexer
	objects []interface{}
}

// NewIndexedStore returns a IndexedStore for cataloguing objects of type o.
func NewIndexedStore(o interface{}) IndexedStore {
	return &indexedStore{
		Indexer: NewFieldIndexer(o),
		objects: []interface{}{},
	}
}

// Insert object o in the store indexing it on all registered indexes.
func (s *indexedStore) Insert(o interface{}) error {
	if err := s.Indexer.Insert(o); err != nil {
		return err
	}
	s.objects = append(s.objects, o)
	return nil
}

// All retrieves all the catalogued instances as a slice. The returned slice should not be modified
// as it shares a backing array with s.
func (s *indexedStore) All() []interface{} {
	return s.objects
}

// Total returns the total number of objects in the store.
func (s *indexedStore) Size() int {
	return len(s.objects)
}

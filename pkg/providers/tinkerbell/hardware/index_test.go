package hardware_test

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestFieldIndexer_InsertAndLookup(t *testing.T) {
	g := gomega.NewWithT(t)
	type Object struct{ Name string }
	const Index = ".Name"

	indexer := hardware.NewFieldIndexer(&Object{})
	indexer.IndexField(Index, func(o interface{}) string {
		object := o.(*Object)
		return object.Name
	})

	objects := indexer.Lookup(Index, "hello")
	g.Expect(objects).To(gomega.BeEmpty())

	const name = "hello world"
	expect := &Object{Name: name}
	indexer.Insert(expect)

	objects = indexer.Lookup(Index, name)
	g.Expect(objects).To(gomega.HaveLen(1))
	g.Expect(objects[0]).To(gomega.Equal(expect))
}

func TestFieldIndexer_InsertIncorrectType(t *testing.T) {
	g := gomega.NewWithT(t)
	type Object struct{ Name string }
	const Index = ".Name"

	indexer := hardware.NewFieldIndexer(&Object{})
	indexer.IndexField(Index, func(o interface{}) string {
		object := o.(*Object)
		return object.Name
	})

	type IncorrectObject struct{}
	err := indexer.Insert(IncorrectObject{})
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err).To(gomega.BeAssignableToTypeOf(hardware.ErrIncorrectType{}))
}

func TestFieldIndexer_NilObjectTypePanics(t *testing.T) {
	g := gomega.NewWithT(t)
	g.Expect(func() {
		hardware.NewFieldIndexer(nil)
	}).To(gomega.Panic())
}

func TestFieldIndexer_NilInterfacePanics(t *testing.T) {
	g := gomega.NewWithT(t)
	g.Expect(func() {
		var i interface{}
		hardware.NewFieldIndexer(i)
	}).To(gomega.Panic())
}

func TestFieldIndexer_LookupUnknownIndexPanics(t *testing.T) {
	g := gomega.NewWithT(t)

	type Object struct{ Name string }
	indexer := hardware.NewFieldIndexer(&Object{})

	g.Expect(func() {
		indexer.Lookup("unknown index", "key")
	}).To(gomega.Panic())
}

func TestFieldIndexer_RemoveValue(t *testing.T) {
	g := gomega.NewWithT(t)
	type Object struct{ Name string }
	const Index = ".Name"

	indexer := hardware.NewFieldIndexer(&Object{})
	indexer.IndexField(Index, func(o interface{}) string {
		object := o.(*Object)
		return object.Name
	})

	const name = "hello world"
	o := &Object{Name: name}
	indexer.Insert(o)

	objects := indexer.Lookup(Index, name)
	g.Expect(objects).To(gomega.HaveLen(1))

	indexer.Remove(o)

	objects = indexer.Lookup(Index, name)
	g.Expect(objects).To(gomega.BeEmpty())
}

func TestFieldIndexer_RemoveIncorrectTypeIsNoop(t *testing.T) {
	g := gomega.NewWithT(t)
	type Object struct{ Name string }
	const Index = ".Name"

	indexer := hardware.NewFieldIndexer(&Object{})
	indexer.IndexField(Index, func(o interface{}) string {
		object := o.(*Object)
		return object.Name
	})

	objects := indexer.Lookup(Index, "hello")
	g.Expect(objects).To(gomega.BeEmpty())
	indexer.Remove("hello")

	objects = indexer.Lookup(Index, "hello")
	g.Expect(objects).To(gomega.BeEmpty())
}

func TestFieldIndexer_RemoveUnknownValueIsNoop(t *testing.T) {
	g := gomega.NewWithT(t)
	type Object struct{ Name string }
	const Index = ".Name"

	indexer := hardware.NewFieldIndexer(&Object{})
	indexer.IndexField(Index, func(o interface{}) string {
		object := o.(*Object)
		return object.Name
	})

	objects := indexer.Lookup(Index, "hello")
	g.Expect(objects).To(gomega.BeEmpty())

	o := &Object{Name: "i am unknown"}
	indexer.Remove(o)

	objects = indexer.Lookup(Index, "hello")
	g.Expect(objects).To(gomega.BeEmpty())
}

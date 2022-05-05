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

	objects, err := indexer.Lookup(Index, "hello")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(objects).To(gomega.BeEmpty())

	const name = "hello world"
	expect := &Object{Name: name}
	err = indexer.Insert(expect)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	objects, err = indexer.Lookup(Index, name)
	g.Expect(err).ToNot(gomega.HaveOccurred())
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

	_, err := indexer.Lookup("unknown index", "key")
	g.Expect(err).To(gomega.HaveOccurred())
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
	err := indexer.Insert(o)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	objects, err := indexer.Lookup(Index, name)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(objects).To(gomega.HaveLen(1))

	err = indexer.Remove(o)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	objects, err = indexer.Lookup(Index, name)
	g.Expect(err).ToNot(gomega.HaveOccurred())
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

	objects, err := indexer.Lookup(Index, "hello")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(objects).To(gomega.BeEmpty())

	err = indexer.Remove("hello")
	g.Expect(err).To(gomega.HaveOccurred())

	objects, err = indexer.Lookup(Index, "hello")
	g.Expect(err).ToNot(gomega.HaveOccurred())
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

	objects, err := indexer.Lookup(Index, "hello")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(objects).To(gomega.BeEmpty())

	o := &Object{Name: "i am unknown"}
	err = indexer.Remove(o)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	objects, err = indexer.Lookup(Index, "hello")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(objects).To(gomega.BeEmpty())
}

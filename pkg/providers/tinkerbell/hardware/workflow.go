package hardware

import (
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
)

const WorkflowHardwareMACIndex = ".Spec.HardwareMap.MAC"

// WorkflowCatalogue stores v1alpha1.Workflow objects. It provides efficient lookup of objects
// based on registered indexes.
type WorkflowCatalogue struct {
	store IndexedStore
}

// NewWorkflowCatalogue constructs a WorkflowCatalogue object for the v1alpha1.Workflow type.
func NewWorkflowCatalogue() *WorkflowCatalogue {
	catalogue := &WorkflowCatalogue{
		store: NewIndexedStore(&v1alpha1.Workflow{}),
	}

	catalogue.IndexWorkflow(WorkflowHardwareMACIndex, func(o interface{}) []string {
		var macs []string
		for _, mac := range o.(*v1alpha1.Workflow).Spec.HardwareMap {
			macs = append(macs, mac)
		}
		return macs
	})

	return catalogue
}

// IndexWorkflow indexes Workflow instances on index by extracfting the key using fn. Objects passed
// to fn will be the same type passed to NewWorkflowCatalogue().
func (c *WorkflowCatalogue) IndexWorkflow(index string, fn KeyExtractorFunc) {
	c.store.IndexField(index, fn)
}

// InsertWorkflow inserts Workflow into the catalogue. If any indexes exist, the workflow is
// indexed.
func (c *WorkflowCatalogue) InsertWorkflow(workflow *v1alpha1.Workflow) error {
	return c.store.Insert(workflow)
}

// // AllWorkflow retrieves a copy of the catalogued Workflow instances.
// func (c *WorkflowCatalogue) AllWorkflow() []*v1alpha1.Workflow {
// 	workflow := make([]*v1alpha1.Workflow, c.store.Size())
// 	for i, v := range c.store.All() {
// 		workflow[i] = v.(*v1alpha1.Workflow)
// 	}
// 	return workflow
// }

// LookupWorkflow retrieves Workflow instances on index with a key of key. Multiple workflow _may_
// have the same key hence it can return multiple Workflow.
func (c *WorkflowCatalogue) LookupWorkflow(index, key string) ([]*v1alpha1.Workflow, error) {
	untyped, err := c.store.Lookup(index, key)
	if err != nil {
		return nil, err
	}

	workflow := make([]*v1alpha1.Workflow, len(untyped))
	for i, v := range untyped {
		workflow[i] = v.(*v1alpha1.Workflow)
	}

	return workflow, nil
}

// // TotalWorkflow returns the total workflow registered in the catalogue.
// func (c *WorkflowCatalogue) TotalWorkflow() int {
// 	return c.store.Size()
// }

package hardware_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestWorkflowCatalogue_Insert(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewWorkflowCatalogue()

	err := catalogue.InsertWorkflow(&v1alpha1.Workflow{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
}

func TestWorkflowCatalogue_WorkflowHardwareMACIndex(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewWorkflowCatalogue()

	workflows, err := catalogue.LookupWorkflow(hardware.WorkflowHardwareMACIndex, "ID")
	g.Expect(err).To(gomega.Succeed())
	g.Expect(workflows).To(gomega.HaveLen(0))

	const mac = "00:00:00:00:00"
	wf := &v1alpha1.Workflow{
		Spec: v1alpha1.WorkflowSpec{
			HardwareMap: map[string]string{"device": mac},
		},
	}
	g.Expect(catalogue.InsertWorkflow(wf)).To(gomega.Succeed())
	workflows, err = catalogue.LookupWorkflow(hardware.WorkflowHardwareMACIndex, mac)
	g.Expect(err).To(gomega.Succeed())
	g.Expect(workflows).To(gomega.HaveLen(1))
	g.Expect(workflows[0]).To(gomega.Equal(wf))
}

// func TestWorkflowCatalogue_AllWorkflowReceivesCopy(t *testing.T) {
// 	g := gomega.NewWithT(t)

// 	catalogue := hardware.NewWorkflowCatalogue()

// 	const totalWorkflow = 1
// 	err := catalogue.InsertWorkflow(&v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "foo"}})
// 	g.Expect(err).ToNot(gomega.HaveOccurred())

// 	changedWorkflow := catalogue.AllWorkflow()
// 	g.Expect(changedWorkflow).To(gomega.HaveLen(totalWorkflow))

// 	changedWorkflow[0] = &v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "bar"}}

// 	unchangedWorkflow := catalogue.AllWorkflow()
// 	g.Expect(unchangedWorkflow).ToNot(gomega.Equal(changedWorkflow))
// }

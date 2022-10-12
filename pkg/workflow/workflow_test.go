package workflow_test

import (
	"context"
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/workflow"
)

func TestWorkflowExecute(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	task1 := NewMockTask(ctrl)
	runTask1 := task1.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	task2 := NewMockTask(ctrl)
	runTask2 := task2.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	gomock.InOrder(runTask1, runTask2)

	wflw := workflow.New(workflow.Config{})
	g.Expect(wflw).ToNot(gomega.BeNil())

	err := wflw.AppendTask("task1", task1)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = wflw.AppendTask("task2", task2)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = wflw.Execute(context.Background())
	g.Expect(err).ToNot(gomega.HaveOccurred())
}

func TestWorkflowHooks(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	const taskName workflow.TaskName = "MockTask"

	preWorkflowHook := NewMockTask(ctrl)
	runPreWorkflowHook := preWorkflowHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	preTaskHook := NewMockTask(ctrl)
	runPreTaskHook := preTaskHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	task := NewMockTask(ctrl)
	runTask := task.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	postTaskHook := NewMockTask(ctrl)
	runPostTaskHook := postTaskHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	postWorkflowHook := NewMockTask(ctrl)
	runPostWorkflowHook := postWorkflowHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	gomock.InOrder(
		runPreWorkflowHook,
		runPreTaskHook,
		runTask,
		runPostTaskHook,
		runPostWorkflowHook,
	)

	wflw := workflow.New(workflow.Config{})
	g.Expect(wflw).ToNot(gomega.BeNil())

	err := wflw.AppendTask(taskName, task)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	wflw.BindPreWorkflowHook(preWorkflowHook)
	wflw.BindPostWorkflowHook(postWorkflowHook)
	wflw.BindPreTaskHook(taskName, preTaskHook)
	wflw.BindPostTaskHook(taskName, postTaskHook)

	err = wflw.Execute(context.Background())
	g.Expect(err).ToNot(gomega.HaveOccurred())
}

func TestErroneousTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	const taskName workflow.TaskName = "MockTask"

	expect := errors.New("expected error")

	preWorkflowHook := NewMockTask(ctrl)
	runPreWorkflowHook := preWorkflowHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	preTaskHook := NewMockTask(ctrl)
	runPreTaskHook := preTaskHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	task1 := NewMockTask(ctrl)
	runTask1 := task1.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), expect)

	postTaskHook := NewMockTask(ctrl)

	// Subsequent tasks after error shouldn't run.
	task2 := NewMockTask(ctrl)

	// These shouldn't run
	postWorkflowHook := NewMockTask(ctrl)

	gomock.InOrder(runPreWorkflowHook, runPreTaskHook, runTask1)

	wflw := workflow.New(workflow.Config{})
	g.Expect(wflw).ToNot(gomega.BeNil())

	err := wflw.AppendTask(taskName, task1)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = wflw.AppendTask("task2", task2)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	wflw.BindPreWorkflowHook(preWorkflowHook)
	wflw.BindPostWorkflowHook(postWorkflowHook)
	wflw.BindPreTaskHook(taskName, preTaskHook)
	wflw.BindPostTaskHook(taskName, postTaskHook)

	err = wflw.Execute(context.Background())
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestErroneousPreWorkflowHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	expect := errors.New("expected error")

	const taskName workflow.TaskName = "MockTask"

	preWorkflowHook := NewMockTask(ctrl)
	preWorkflowHook.EXPECT().
		RunTask(gomock.Any()).
		Return(nil, expect)

	preTaskHook := NewMockTask(ctrl)

	task := NewMockTask(ctrl)

	postTaskHook := NewMockTask(ctrl)

	postWorkflowHook := NewMockTask(ctrl)

	wflw := workflow.New(workflow.Config{})
	g.Expect(wflw).ToNot(gomega.BeNil())

	err := wflw.AppendTask(taskName, task)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	wflw.BindPreWorkflowHook(preWorkflowHook)
	wflw.BindPostWorkflowHook(postWorkflowHook)
	wflw.BindPreTaskHook(taskName, preTaskHook)
	wflw.BindPostTaskHook(taskName, postTaskHook)

	err = wflw.Execute(context.Background())
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestErroneousPostWorkflowHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	expect := errors.New("expected error")

	const taskName workflow.TaskName = "MockTask"

	preWorkflowHook := NewMockTask(ctrl)
	runPreWorkflowHook := preWorkflowHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	preTaskHook := NewMockTask(ctrl)
	runPreTaskHook := preTaskHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	task := NewMockTask(ctrl)
	runTask := task.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	postTaskHook := NewMockTask(ctrl)
	runPostTaskHook := postTaskHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	postWorkflowHook := NewMockTask(ctrl)
	runPostWorkflowHook := postWorkflowHook.EXPECT().
		RunTask(gomock.Any()).
		Return(nil, expect)

	gomock.InOrder(
		runPreWorkflowHook,
		runPreTaskHook,
		runTask,
		runPostTaskHook,
		runPostWorkflowHook,
	)

	wflw := workflow.New(workflow.Config{})
	g.Expect(wflw).ToNot(gomega.BeNil())

	err := wflw.AppendTask(taskName, task)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	wflw.BindPreWorkflowHook(preWorkflowHook)
	wflw.BindPostWorkflowHook(postWorkflowHook)
	wflw.BindPreTaskHook(taskName, preTaskHook)
	wflw.BindPostTaskHook(taskName, postTaskHook)

	err = wflw.Execute(context.Background())
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestErroneousPreTaskHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	expect := errors.New("expected error")

	const taskName workflow.TaskName = "MockTask"

	// These shouldn't run.
	postTaskHook := NewMockTask(ctrl)
	postWorkflowHook := NewMockTask(ctrl)

	task := NewMockTask(ctrl)

	preTaskHook := NewMockTask(ctrl)
	runPreTaskHook := preTaskHook.EXPECT().
		RunTask(gomock.Any()).
		Return(nil, expect)

	preWorkflowHook := NewMockTask(ctrl)
	runPreWorkflowHook := preWorkflowHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	gomock.InOrder(runPreWorkflowHook, runPreTaskHook)

	wflw := workflow.New(workflow.Config{})
	g.Expect(wflw).ToNot(gomega.BeNil())

	err := wflw.AppendTask(taskName, task)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	wflw.BindPreWorkflowHook(preWorkflowHook)
	wflw.BindPostWorkflowHook(postWorkflowHook)
	wflw.BindPreTaskHook(taskName, preTaskHook)
	wflw.BindPostTaskHook(taskName, postTaskHook)

	err = wflw.Execute(context.Background())
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestErroneousPostTaskHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	expect := errors.New("expected error")

	const taskName workflow.TaskName = "MockTask"

	preWorkflowHook := NewMockTask(ctrl)
	runPreWorkflowHook := preWorkflowHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	preTaskHook := NewMockTask(ctrl)
	runPreTaskHook := preTaskHook.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	task := NewMockTask(ctrl)
	runTask := task.EXPECT().
		RunTask(gomock.Any()).
		Return(context.Background(), nil)

	postTaskHook := NewMockTask(ctrl)
	runPostTaskHook := postTaskHook.EXPECT().
		RunTask(gomock.Any()).
		Return(nil, expect)

	postWorkflowHook := NewMockTask(ctrl)

	gomock.InOrder(
		runPreWorkflowHook,
		runPreTaskHook,
		runTask,
		runPostTaskHook,
	)

	wflw := workflow.New(workflow.Config{})
	g.Expect(wflw).ToNot(gomega.BeNil())

	err := wflw.AppendTask(taskName, task)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	wflw.BindPreWorkflowHook(preWorkflowHook)
	wflw.BindPostWorkflowHook(postWorkflowHook)
	wflw.BindPreTaskHook(taskName, preTaskHook)
	wflw.BindPostTaskHook(taskName, postTaskHook)

	err = wflw.Execute(context.Background())
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestDuplicateTaskNames(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	const taskName workflow.TaskName = "MockTask"

	task1 := NewMockTask(ctrl)
	task2 := NewMockTask(ctrl)

	wflw := workflow.New(workflow.Config{})
	g.Expect(wflw).ToNot(gomega.BeNil())

	err := wflw.AppendTask(taskName, task1)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = wflw.AppendTask(taskName, task2)
	g.Expect(err).To(gomega.HaveOccurred())
}

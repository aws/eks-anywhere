package task_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/task"
	mocktasks "github.com/aws/eks-anywhere/pkg/task/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

func TestTaskRunnerRunTask(t *testing.T) {
	tr := newTaskRunnerTest(t)

	tr.taskA.EXPECT().Run(tr.ctx, tr.cmdContext).Return(tr.taskB).Times(1)
	tr.taskA.EXPECT().Name().Return("taskA").Times(7)
	tr.taskA.EXPECT().Checkpoint()
	tr.taskB.EXPECT().Run(tr.ctx, tr.cmdContext).Return(tr.taskC).Times(1)
	tr.taskB.EXPECT().Name().Return("taskB").Times(7)
	tr.taskB.EXPECT().Checkpoint()
	tr.taskC.EXPECT().Run(tr.ctx, tr.cmdContext).Return(nil).Times(1)
	tr.taskC.EXPECT().Name().Return("taskC").Times(7)
	tr.taskC.EXPECT().Checkpoint()

	type fields struct {
		tasks []task.Task
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Task runs and next Task is triggered and profiles are captured",
			fields: fields{
				tasks: []task.Task{tr.taskA, tr.taskB, tr.taskC},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := task.NewTaskRunner(tt.fields.tasks[0], tr.writer)
			if err := runner.RunTask(tr.ctx, tr.cmdContext); err != nil {
				t.Fatal(err)
			}
		})
		for _, task := range tt.fields.tasks {
			if _, ok := tr.cmdContext.Profiler.Metrics()[task.Name()]; !ok {
				t.Fatal("Error Profiler doesn't have metrics")
			}
		}
	}
}

func TestTaskRunnerRunTaskWithCheckpointSecondRunSuccess(t *testing.T) {
	tt := newTaskRunnerTest(t)

	tt.taskA.EXPECT().Restore(tt.ctx, tt.cmdContext, gomock.Any()).Return(tt.taskB, nil)
	tt.taskA.EXPECT().Name().Return("taskA").Times(2)
	tt.taskB.EXPECT().Run(tt.ctx, tt.cmdContext).Return(tt.taskC).Times(1)
	tt.taskB.EXPECT().Name().Return("taskB").Times(6)
	tt.taskB.EXPECT().Checkpoint()
	tt.taskC.EXPECT().Run(tt.ctx, tt.cmdContext).Return(nil).Times(1)
	tt.taskC.EXPECT().Name().Return("taskC").Times(6)
	tt.taskC.EXPECT().Checkpoint()
	tt.writer.EXPECT().TempDir().Return("testdata")

	tasks := []task.Task{tt.taskA, tt.taskB, tt.taskC}

	t.Setenv(features.CheckpointEnabledEnvVar, "true")
	runner := task.NewTaskRunner(tasks[0], tt.cmdContext.Writer, task.WithCheckpointFile())
	if err := runner.RunTask(tt.ctx, tt.cmdContext); err != nil {
		t.Fatal(err)
	}

	if err := os.Unsetenv(features.CheckpointEnabledEnvVar); err != nil {
		t.Fatal(err)
	}
}

func TestTaskRunnerRunTaskWithCheckpointFirstRunFailed(t *testing.T) {
	tt := newTaskRunnerTest(t)
	tt.cmdContext.OriginalError = fmt.Errorf("error")

	tt.taskA.EXPECT().Run(tt.ctx, tt.cmdContext).Return(nil)
	tt.taskA.EXPECT().Name().Return("taskA").Times(5)
	tt.writer.EXPECT().TempDir()
	tt.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", tt.cmdContext.ClusterSpec.Cluster.Name), gomock.Any())

	tasks := []task.Task{tt.taskA, tt.taskB}
	t.Setenv(features.CheckpointEnabledEnvVar, "true")
	runner := task.NewTaskRunner(tasks[0], tt.cmdContext.Writer, task.WithCheckpointFile())
	if err := runner.RunTask(tt.ctx, tt.cmdContext); err == nil {
		t.Fatalf("Task.RunTask want err, got nil")
	}

	if err := os.Unsetenv(features.CheckpointEnabledEnvVar); err != nil {
		t.Fatal(err)
	}
}

func TestTaskRunnerRunTaskWithCheckpointSecondRunRestoreFailure(t *testing.T) {
	tt := newTaskRunnerTest(t)

	tt.taskA.EXPECT().Restore(tt.ctx, tt.cmdContext, gomock.Any()).Return(nil, fmt.Errorf("error"))
	tt.taskA.EXPECT().Name().Return("taskA").Times(2)
	tt.writer.EXPECT().TempDir().Return("testdata")

	tasks := []task.Task{tt.taskA, tt.taskB, tt.taskC}

	t.Setenv(features.CheckpointEnabledEnvVar, "true")
	runner := task.NewTaskRunner(tasks[0], tt.cmdContext.Writer, task.WithCheckpointFile())
	if err := runner.RunTask(tt.ctx, tt.cmdContext); err == nil {
		t.Fatalf("Task.Restore want err, got nil")
	}

	if err := os.Unsetenv(features.CheckpointEnabledEnvVar); err != nil {
		t.Fatal(err)
	}
}

func TestTaskRunnerRunTaskWithCheckpointSaveFailed(t *testing.T) {
	tt := newTaskRunnerTest(t)
	tt.cmdContext.OriginalError = fmt.Errorf("error")

	tt.taskA.EXPECT().Run(tt.ctx, tt.cmdContext).Return(nil)
	tt.taskA.EXPECT().Name().Return("taskA").Times(5)
	tt.writer.EXPECT().TempDir()
	tt.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", tt.cmdContext.ClusterSpec.Cluster.Name), gomock.Any()).Return("", fmt.Errorf("error"))

	tasks := []task.Task{tt.taskA, tt.taskB}

	t.Setenv(features.CheckpointEnabledEnvVar, "true")
	runner := task.NewTaskRunner(tasks[0], tt.cmdContext.Writer, task.WithCheckpointFile())
	if err := runner.RunTask(tt.ctx, tt.cmdContext); err == nil {
		t.Fatalf("Task.RunTask want err, got nil")
	}

	if err := os.Unsetenv(features.CheckpointEnabledEnvVar); err != nil {
		t.Fatal(err)
	}
}

func TestTaskRunnerRunTaskWithCheckpointReadFailure(t *testing.T) {
	tt := newTaskRunnerTest(t)
	tt.cmdContext.ClusterSpec.Cluster.Name = "invalid"

	tt.writer.EXPECT().TempDir().Return("testdata")

	tasks := []task.Task{tt.taskA, tt.taskB, tt.taskC}

	t.Setenv(features.CheckpointEnabledEnvVar, "true")
	runner := task.NewTaskRunner(tasks[0], tt.cmdContext.Writer, task.WithCheckpointFile())
	if err := runner.RunTask(tt.ctx, tt.cmdContext); err == nil {
		t.Fatalf("Task.ReadCheckpointFile want err, got nil")
	}

	if err := os.Unsetenv(features.CheckpointEnabledEnvVar); err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalTaskCheckpointSuccess(t *testing.T) {
	testConfigType := types.Cluster{}
	testTaskCheckpoint := types.Cluster{
		Name:               "test-cluster",
		KubeconfigFile:     "test.kubeconfig",
	}

	if err := task.UnmarshalTaskCheckpoint(testTaskCheckpoint, testConfigType); err != nil {
		t.Fatalf("task.UnmarshalTaskCheckpoint err = %v, want nil", err)
	}
}

type taskRunnerTest struct {
	ctx        context.Context
	cmdContext *task.CommandContext
	taskA      *mocktasks.MockTask
	taskB      *mocktasks.MockTask
	taskC      *mocktasks.MockTask
	writer     *writermocks.MockFileWriter
}

func newTaskRunnerTest(t *testing.T) *taskRunnerTest {
	ctrl := gomock.NewController(t)

	cmdContext := &task.CommandContext{
		ClusterSpec: &cluster.Spec{
			Config: &cluster.Config{
				Cluster: &v1alpha1.Cluster{},
			},
		},
	}
	cmdContext.ClusterSpec.Cluster.Name = "test-cluster"
	writer := writermocks.NewMockFileWriter(ctrl)
	cmdContext.Writer = writer
	cleanTaskA := mocktasks.NewMockTask(ctrl)
	cleanTaskB := mocktasks.NewMockTask(ctrl)
	cleanTaskC := mocktasks.NewMockTask(ctrl)

	return &taskRunnerTest{
		ctx:        context.Background(),
		cmdContext: cmdContext,
		taskA:      cleanTaskA,
		taskB:      cleanTaskB,
		taskC:      cleanTaskC,
		writer:     writer,
	}
}

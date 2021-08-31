package task_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/task"
	mocktasks "github.com/aws/eks-anywhere/pkg/task/mocks"
)

func TestTaskRunnerRunTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	cmdContext := &task.CommandContext{}
	cleanTaskA := mocktasks.NewMockTask(ctrl)
	cleanTaskB := mocktasks.NewMockTask(ctrl)
	cleanTaskC := mocktasks.NewMockTask(ctrl)

	cleanTaskA.EXPECT().Run(ctx, cmdContext).Return(cleanTaskB).Times(1)
	cleanTaskA.EXPECT().Name().Return("taskA").Times(5)
	cleanTaskB.EXPECT().Run(ctx, cmdContext).Return(cleanTaskC).Times(1)
	cleanTaskB.EXPECT().Name().Return("taskB").Times(5)
	cleanTaskC.EXPECT().Run(ctx, cmdContext).Return(nil).Times(1)
	cleanTaskC.EXPECT().Name().Return("taskC").Times(5)

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
				tasks: []task.Task{cleanTaskA, cleanTaskB, cleanTaskC},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := task.NewTaskRunner(tt.fields.tasks[0])
			if err := runner.RunTask(ctx, cmdContext); err != nil {
				t.Fatal(err)
			}
		})
		for _, task := range tt.fields.tasks {
			if _, ok := cmdContext.Profiler.Metrics()[task.Name()]; !ok {
				t.Fatal("Error Profiler doesn't have metrics")
			}
		}
	}
}

func TestTaskRunnerRunTaskRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	cmdContext := &task.CommandContext{}
	expectedError := fmt.Errorf("error")
	cmdContext.OriginalError = expectedError
	cleanTaskA := mocktasks.NewMockTask(ctrl)
	cleanTaskB := mocktasks.NewMockTask(ctrl)

	cleanTaskA.EXPECT().Run(ctx, cmdContext).Return(cleanTaskB).Times(1)
	cleanTaskA.EXPECT().Name().Return("taskA").Times(4)
	cleanTaskB.EXPECT().Run(ctx, cmdContext).Return(nil).Times(0)
	cleanTaskB.EXPECT().Name().Return("taskB").Times(0)

	runner := task.NewTaskRunner(cleanTaskA)
	err := runner.RunTask(ctx, cmdContext)
	if err != expectedError {
		t.Fatal("Expected error not returned")
	}
}

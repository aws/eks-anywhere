package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// WorkflowIDAnnotation is used by the controller to store the
	// ID assigned to the workflow by Tinkerbell for migrated workflows.
	WorkflowIDAnnotation = "workflow.tinkerbell.org/id"
)

// TinkID returns the Tinkerbell ID associated with this Workflow.
func (w *Workflow) TinkID() string {
	return w.Annotations[WorkflowIDAnnotation]
}

// SetTinkID sets the Tinkerbell ID associated with this Workflow.
func (w *Workflow) SetTinkID(id string) {
	if w.Annotations == nil {
		w.Annotations = make(map[string]string)
	}
	w.Annotations[WorkflowIDAnnotation] = id
}

// GetStartTime returns the start time, for the first action of the first task.
func (w *Workflow) GetStartTime() *metav1.Time {
	if len(w.Status.Tasks) > 0 {
		if len(w.Status.Tasks[0].Actions) > 0 {
			return w.Status.Tasks[0].Actions[0].StartedAt
		}
	}
	return nil
}

type taskInfo struct {
	CurrentWorker        string
	CurrentTask          string
	CurrentTaskIndex     int
	CurrentAction        string
	CurrentActionIndex   int
	CurrentActionState   WorkflowState
	TotalNumberOfActions int
}

// helper function for task info.
func (w *Workflow) getTaskActionInfo() taskInfo {
	var (
		found           bool
		taskIndex       = -1
		actionIndex     int
		actionTaskIndex int
		actionCount     int
	)
	for ti, task := range w.Status.Tasks {
		actionCount += len(task.Actions)
		if found {
			continue
		}
	INNER:
		for ai, action := range task.Actions {
			// Find the first non-successful action
			switch action.Status { //nolint:exhaustive // WorkflowStateWaiting is only used in Workflows not Actions.
			case WorkflowStateSuccess:
				actionIndex++
				continue
			case WorkflowStatePending, WorkflowStateRunning, WorkflowStateFailed, WorkflowStateTimeout:
				taskIndex = ti
				actionTaskIndex = ai
				found = true
				break INNER
			}
		}
	}

	ti := taskInfo{
		TotalNumberOfActions: actionCount,
		CurrentActionIndex:   actionIndex,
	}
	if taskIndex >= 0 {
		ti.CurrentWorker = w.Status.Tasks[taskIndex].WorkerAddr
		ti.CurrentTask = w.Status.Tasks[taskIndex].Name
		ti.CurrentTaskIndex = taskIndex
	}
	if taskIndex >= 0 && actionIndex >= 0 {
		ti.CurrentAction = w.Status.Tasks[taskIndex].Actions[actionTaskIndex].Name
		ti.CurrentActionState = w.Status.Tasks[taskIndex].Actions[actionTaskIndex].Status
	}

	return ti
}

func (w *Workflow) GetCurrentWorker() string {
	return w.getTaskActionInfo().CurrentWorker
}

func (w *Workflow) GetCurrentTask() string {
	return w.getTaskActionInfo().CurrentTask
}

func (w *Workflow) GetCurrentTaskIndex() int {
	return w.getTaskActionInfo().CurrentTaskIndex
}

func (w *Workflow) GetCurrentAction() string {
	return w.getTaskActionInfo().CurrentAction
}

func (w *Workflow) GetCurrentActionIndex() int {
	return w.getTaskActionInfo().CurrentActionIndex
}

func (w *Workflow) GetCurrentActionState() WorkflowState {
	return w.getTaskActionInfo().CurrentActionState
}

func (w *Workflow) GetTotalNumberOfActions() int {
	return w.getTaskActionInfo().TotalNumberOfActions
}

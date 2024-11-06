package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}

type (
	WorkflowState         string
	WorkflowConditionType string
	TemplateRendering     string
	BootMode              string
)

const (
	WorkflowStatePreparing = WorkflowState("STATE_PREPARING")
	WorkflowStatePending   = WorkflowState("STATE_PENDING")
	WorkflowStateRunning   = WorkflowState("STATE_RUNNING")
	WorkflowStatePost      = WorkflowState("STATE_POST")
	WorkflowStateSuccess   = WorkflowState("STATE_SUCCESS")
	WorkflowStateFailed    = WorkflowState("STATE_FAILED")
	WorkflowStateTimeout   = WorkflowState("STATE_TIMEOUT")

	NetbootJobFailed        WorkflowConditionType = "NetbootJobFailed"
	NetbootJobComplete      WorkflowConditionType = "NetbootJobComplete"
	NetbootJobRunning       WorkflowConditionType = "NetbootJobRunning"
	NetbootJobSetupFailed   WorkflowConditionType = "NetbootJobSetupFailed"
	NetbootJobSetupComplete WorkflowConditionType = "NetbootJobSetupComplete"
	ToggleAllowNetbootTrue  WorkflowConditionType = "AllowNetbootTrue"
	ToggleAllowNetbootFalse WorkflowConditionType = "AllowNetbootFalse"
	TemplateRenderedSuccess WorkflowConditionType = "TemplateRenderedSuccess"

	TemplateRenderingSuccessful TemplateRendering = "successful"
	TemplateRenderingFailed     TemplateRendering = "failed"

	BootModeNetboot BootMode = "netboot"
	BootModeISO     BootMode = "iso"
)

// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=workflows,scope=Namespaced,categories=tinkerbell,shortName=wf,singular=workflow
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:JSONPath=".spec.templateRef",name=Template,type=string
// +kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string
// +kubebuilder:printcolumn:JSONPath=".status.currentAction",name=Current-Action,type=string
// +kubebuilder:printcolumn:JSONPath=".status.templateRending",name=Template-Rendering,type=string

// Workflow is the Schema for the Workflows API.
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkflowList contains a list of Workflows.
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

// WorkflowSpec defines the desired state of Workflow.
type WorkflowSpec struct {
	// Name of the Template associated with this workflow.
	TemplateRef string `json:"templateRef,omitempty"`

	// Name of the Hardware associated with this workflow.
	// +optional
	HardwareRef string `json:"hardwareRef,omitempty"`

	// A mapping of template devices to hadware mac addresses.
	HardwareMap map[string]string `json:"hardwareMap,omitempty"`

	// BootOptions are options that control the booting of Hardware.
	BootOptions BootOptions `json:"bootOptions,omitempty"`
}

// BootOptions are options that control the booting of Hardware.
type BootOptions struct {
	// ToggleAllowNetboot indicates whether the controller should toggle the field in the associated hardware for allowing PXE booting.
	// This will be enabled before a Workflow is executed and disabled after the Workflow has completed successfully.
	// A HardwareRef must be provided.
	// +optional
	ToggleAllowNetboot bool `json:"toggleAllowNetboot,omitempty"`

	// ISOURL is the URL of the ISO that will be one-time booted. When this field is set, the controller will create a job.bmc.tinkerbell.org object
	// for getting the associated hardware into a CDROM booting state.
	// A HardwareRef that contains a spec.BmcRef must be provided.
	// +optional
	// +kubebuilder:validation:Format=url
	ISOURL string `json:"isoURL,omitempty"`

	// BootMode is the type of booting that will be done.
	// +optional
	// +kubebuilder:validation:Enum=netboot;iso
	BootMode BootMode `json:"bootMode,omitempty"`
}

// BootOptionsStatus holds the state of any boot options.
type BootOptionsStatus struct {
	// AllowNetboot holds the state of the the controller's interactions with the allowPXE field in a Hardware object.
	AllowNetboot AllowNetbootStatus `json:"allowNetboot,omitempty"`
	// Jobs holds the state of any job.bmc.tinkerbell.org objects created.
	Jobs map[string]JobStatus `json:"jobs,omitempty"`
}

type AllowNetbootStatus struct {
	ToggledTrue  bool `json:"toggledTrue,omitempty"`
	ToggledFalse bool `json:"toggledFalse,omitempty"`
}

// WorkflowStatus defines the observed state of a Workflow.
type WorkflowStatus struct {
	// State is the current overall state of the Workflow.
	State WorkflowState `json:"state,omitempty"`

	// CurrentAction is the action that is currently in the running state.
	CurrentAction string `json:"currentAction,omitempty"`

	// BootOptions holds the state of any boot options.
	BootOptions BootOptionsStatus `json:"bootOptions,omitempty"`

	// TemplateRendering indicates whether the template was rendered successfully.
	// Possible values are "successful" or "failed" or "unknown".
	TemplateRendering TemplateRendering `json:"templateRending,omitempty"`

	// GlobalTimeout represents the max execution time.
	GlobalTimeout int64 `json:"globalTimeout,omitempty"`

	// Tasks are the tasks to be run by the worker(s).
	Tasks []Task `json:"tasks,omitempty"`

	// Conditions are the latest available observations of an object's current state.
	//
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=atomic
	Conditions []WorkflowCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// JobStatus holds the state of a specific job.bmc.tinkerbell.org object created.
type JobStatus struct {
	// UID is the UID of the job.bmc.tinkerbell.org object associated with this workflow.
	// This is used to uniquely identify the job.bmc.tinkerbell.org object, as
	// all objects for a specific Hardware/Machine.bmc.tinkerbell.org are created with the same name.
	UID types.UID `json:"uid,omitempty"`

	// Complete indicates whether the created job.bmc.tinkerbell.org has reported its conditions as complete.
	Complete bool `json:"complete,omitempty"`

	// ExistingJobDeleted indicates whether any existing job.bmc.tinkerbell.org was deleted.
	// The name of each job.bmc.tinkerbell.org object created by the controller is the same, so only one can exist at a time.
	// Using the same name was chosen so that there is only ever 1 job.bmc.tinkerbell.org per Hardware/Machine.bmc.tinkerbell.org.
	// This makes clean up easier and we dont just orphan jobs every time.
	ExistingJobDeleted bool `json:"existingJobDeleted,omitempty"`
}

// JobCondition describes current state of a job.
type WorkflowCondition struct {
	// Type of job condition, Complete or Failed.
	Type WorkflowConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=WorkflowConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status metav1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Reason is a (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Message is a human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
	// Time when the condition was created.
	// +optional
	Time *metav1.Time `json:"time,omitempty" protobuf:"bytes,7,opt,name=time"`
}

// Task represents a series of actions to be completed by a worker.
type Task struct {
	Name        string            `json:"name"`
	WorkerAddr  string            `json:"worker"`
	Actions     []Action          `json:"actions"`
	Volumes     []string          `json:"volumes,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// Action represents a workflow action.
type Action struct {
	Name        string            `json:"name,omitempty"`
	Image       string            `json:"image,omitempty"`
	Timeout     int64             `json:"timeout,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Volumes     []string          `json:"volumes,omitempty"`
	Pid         string            `json:"pid,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Status      WorkflowState     `json:"status,omitempty"`
	StartedAt   *metav1.Time      `json:"startedAt,omitempty"`
	Seconds     int64             `json:"seconds,omitempty"`
	Message     string            `json:"message,omitempty"`
}

// HasCondition checks if the cType condition is present with status cStatus on a bmj.
func (w *WorkflowStatus) HasCondition(wct WorkflowConditionType, cs metav1.ConditionStatus) bool {
	for _, c := range w.Conditions {
		if c.Type == wct {
			return c.Status == cs
		}
	}

	return false
}

// SetCondition updates conditions. If the condition already exists, it updates it.
// If the condition doesn't exist then it appends the new one (wc).
func (w *WorkflowStatus) SetCondition(wc WorkflowCondition) {
	index := -1
	for i, c := range w.Conditions {
		if c.Type == wc.Type {
			index = i
			break
		}
	}
	if index != -1 {
		w.Conditions[index] = wc
		return
	}

	w.Conditions = append(w.Conditions, wc)
}

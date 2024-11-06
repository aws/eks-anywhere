package analyzer

import (
	"context"

	tinkv1 "github.com/aws/eks-anywhere/internal/thirdparty/tink/api/v1alpha1"
)

type workflow struct {
	name      string
	namespace string
}

func newWorkflow(name, namespace string) workflow {
	return workflow{
		name:      name,
		namespace: namespace,
	}
}

func (w workflow) analyze(ctx context.Context, readers readers) (*analyzeResults, error) {
	workflow := &tinkv1.Workflow{}
	if err := readers.client.Get(ctx, w.name, w.namespace, workflow); err != nil {
		return nil, err
	}

	r := &analyzeResults{}

	switch workflow.Status.State {
	case tinkv1.WorkflowStateSuccess:
		return nil, nil
	case tinkv1.WorkflowStateFailed:
		r.Finding = Finding{
			Severity: SeverityError,
			Message:  resourceStatusMessage(SeverityError, "Workflow", w.name, w.namespace, "failed", ""),
		}
	case tinkv1.WorkflowStatePending:
		r.Finding = Finding{
			Severity: SeverityWarning,
			Message:  resourceStatusMessage(SeverityWarning, "Workflow", w.name, w.namespace, "still pending", ""),
		}
	case tinkv1.WorkflowStateRunning:
		r.Finding = Finding{
			Severity: SeverityWarning,
			Message:  resourceStatusMessage(SeverityWarning, "Workflow", w.name, w.namespace, "still running", ""),
		}
	case tinkv1.WorkflowStatePost:
		r.Finding = Finding{
			Severity: SeverityWarning,
			Message:  resourceStatusMessage(SeverityWarning, "Workflow", w.name, w.namespace, "still in post state", ""),
		}
	case tinkv1.WorkflowStatePreparing:
		r.Finding = Finding{
			Severity: SeverityWarning,
			Message:  resourceStatusMessage(SeverityWarning, "Workflow", w.name, w.namespace, "still preparing", ""),
		}
	case tinkv1.WorkflowStateTimeout:
		r.Finding = Finding{
			Severity: SeverityError,
			Message:  resourceStatusMessage(SeverityError, "Workflow", w.name, w.namespace, "timed out", ""),
		}
	default:
		return nil, nil
	}

	for _, task := range workflow.Status.Tasks {
		for _, a := range task.Actions {
			r.nextAnalyzers = append(r.nextAnalyzers, action{Action: a})
		}
	}

	return r, nil
}

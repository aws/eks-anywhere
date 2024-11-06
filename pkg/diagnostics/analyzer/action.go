package analyzer

import (
	"context"
	"fmt"

	tinkv1 "github.com/aws/eks-anywhere/internal/thirdparty/tink/api/v1alpha1"
)

type action struct {
	tinkv1.Action
}

func (a action) analyze(ctx context.Context, readers readers) (*analyzeResults, error) {
	r := &analyzeResults{}

	switch a.Status {
	case tinkv1.WorkflowStateSuccess:
		return nil, nil
	case tinkv1.WorkflowStateFailed:
		r.Finding = Finding{
			Severity:       SeverityError,
			Message:        actionStatusMessage(SeverityError, "Action", a.Name, "failed", a.Image),
			Recommendation: bootsRecommendation(),
		}
	case tinkv1.WorkflowStatePending:
		r.Finding = Finding{
			Severity: SeverityWarning,
			Message:  actionStatusMessage(SeverityWarning, "Action", a.Name, "still pending", a.Image),
		}
	case tinkv1.WorkflowStateRunning:
		r.Finding = Finding{
			Severity: SeverityWarning,
			Message:  actionStatusMessage(SeverityWarning, "Action", a.Name, "still running", a.Image),
		}
	case tinkv1.WorkflowStatePost:
		r.Finding = Finding{
			Severity: SeverityWarning,
			Message:  actionStatusMessage(SeverityWarning, "Action", a.Name, "still in post state", a.Image),
		}
	case tinkv1.WorkflowStatePreparing:
		r.Finding = Finding{
			Severity: SeverityWarning,
			Message:  actionStatusMessage(SeverityWarning, "Action", a.Name, "still preparing", a.Image),
		}
	case tinkv1.WorkflowStateTimeout:
		r.Finding = Finding{
			Severity: SeverityError,
			Message:  actionStatusMessage(SeverityError, "Action", a.Name, "timed out", a.Image),
		}
	default:
		return nil, nil
	}

	return r, nil
}

func bootsRecommendation() string {
	return fmt.Sprintf("Look for relevant logs in the %s controller. It's possible that if there was a problem loading the task image, you might need to go the machine's console to retrieve more information.", bold("eksa-system/boots"))
}

package analyzer

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	v1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

func newInfraMachine(client kubernetes.Reader, ref v1.ObjectReference) (analyzer, error) {
	switch ref.Kind {
	case "TinkerbellMachine":
		return newTinkerbellMachine(ref.Name, ref.Namespace), nil
	default:
		return nil, errors.Errorf("infra machine kind %s not supported", ref.Kind)
	}
}

type tinkerbellMachine struct {
	name      string
	namespace string
}

func newTinkerbellMachine(name, namespace string) tinkerbellMachine {
	return tinkerbellMachine{
		name:      name,
		namespace: namespace,
	}
}

func (m tinkerbellMachine) analyze(ctx context.Context, readers readers) (*analyzeResults, error) {
	machine := &tinkerbellv1.TinkerbellMachine{}
	if err := readers.client.Get(ctx, m.name, m.namespace, machine); err != nil {
		return nil, err
	}

	if machine.Status.Ready {
		return nil, nil
	}

	var message string
	if machine.Spec.HardwareName != "" {
		message = fmt.Sprintf("has hardware %s assigned", bold(machine.Spec.HardwareName))
	} else {
		message = "has no hardware assigned"
	}

	r := &analyzeResults{
		Finding: Finding{
			Severity: SeverityWarning,
			Message:  resourceStatusMessage(SeverityWarning, "TinkerbellMachine", m.name, m.namespace, "not ready", message),
		},
	}

	logs, err := readers.podLogs.LogsFromDeployment("capt-controller-manager", "capt-system", func(l string) bool {
		return strings.Contains(l, `"controllerKind"="TinkerbellMachine"`) &&
			strings.Contains(l, fmt.Sprintf(`"name"="%s" "namespace"="%s"`, m.name, m.namespace))
	})
	if err != nil {
		return nil, errors.Wrapf(err, "getting logs for TinkerbellMachine %s/%s", m.namespace, m.name)
	}

	if len(logs) == 0 {
		return r, nil
	}

	lastLog := logs[len(logs)-1]

	r.Finding.Logs = append(r.Finding.Logs,
		Log{
			Source: "capt-system/capt-controller-manager",
			Lines:  []string{lastLog},
		},
	)

	if strings.Contains(lastLog, `"error"="workflow failed"`) {
		r.nextAnalyzers = append(r.nextAnalyzers, newWorkflow(m.name, m.namespace))
	}

	if machine.Spec.HardwareName == "" {
		return r, nil
	}

	logs, err = readers.podLogs.LogsFromDeployment("boots", "eksa-system", func(l string) bool {
		return strings.Contains(l, fmt.Sprintf(` app-name=%s `, machine.Spec.HardwareName))
	})
	if err != nil {
		return nil, errors.Wrapf(err, "getting logs from boots for hardware ", machine.Spec.HardwareName)
	}

	bootsSource := "eksa-system/boots"
	if len(logs) > 30 {
		logs = logs[len(logs)-30:]
		bootsSource = fmt.Sprintf("%s (last 30 lines)", bootsSource)
	}

	r.Finding.Logs = append(r.Finding.Logs,
		Log{
			Source: bootsSource,
			Lines:  logs,
		},
	)

	return r, nil
}

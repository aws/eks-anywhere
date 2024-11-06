package analyzer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
)

type machine struct {
	name      string
	namespace string
}

func newMachine(name, namespace string) machine {
	return machine{
		name:      name,
		namespace: namespace,
	}
}

func (m machine) analyze(ctx context.Context, readers readers) (*analyzeResults, error) {
	machine := &clusterv1.Machine{}
	if err := readers.client.Get(ctx, m.name, m.namespace, machine); err != nil {
		return nil, err
	}

	readyCondition := conditions.Get(machine, clusterv1.ReadyCondition)

	if isTrue(readyCondition) {
		return nil, nil
	}

	r := &analyzeResults{
		Finding: Finding{
			Severity: SeverityWarning,
			Message:  resourceStatusMessage(SeverityWarning, "Machine", m.name, m.namespace, "not ready", fmt.Sprintf("[%s] %s", readyCondition.Reason, readyCondition.Message)),
		},
	}

	if conditions.IsFalse(machine, clusterv1.InfrastructureReadyCondition) {
		infraMachine, err := newInfraMachine(readers.client, machine.Spec.InfrastructureRef)
		if err != nil {
			return nil, errors.Wrapf(err, "analyzing infra machine for machine %s/%s", m.namespace, m.name)
		}
		r.nextAnalyzers = append(r.nextAnalyzers, infraMachine)
	}

	return r, nil
}

package conditions

import (
	"sort"

	corev1 "k8s.io/api/core/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type conditionPolarity string

const (
	// PositivePolarity describe a condition with positive polarity (Status=True good).
	PositivePolarity conditionPolarity = "Positive"

	// NegativePolarity describe a condition with negative polarity (Status=False good).
	NegativePolarity conditionPolarity = "Negative"
)

// localizedCondition defines a condition with the information of the object the conditions
// was originated from.
type localizedCondition struct {
	*anywherev1.Condition
	Polarity conditionPolarity
	Getter
}

// merge a list of condition into a single one.
func merge(conditions []localizedCondition, targetCondition anywherev1.ConditionType, options *mergeOptions) *anywherev1.Condition {
	g := getConditionGroups(conditions)
	if len(g) == 0 {
		return nil
	}

	if g.TopGroup().status == corev1.ConditionTrue {
		return TrueCondition(targetCondition)
	}

	targetReason := getReason(g, options)
	targetMessage := getMessage(g, options)
	if g.TopGroup().status == corev1.ConditionFalse {
		return FalseCondition(targetCondition, targetReason, g.TopGroup().severity, "%s", targetMessage)
	}
	return UnknownCondition(targetCondition, targetReason, "%s", targetMessage)
}

// getConditionGroups groups a list of conditions according to status, severity values.
//
//nolint:gocyclo
func getConditionGroups(conditions []localizedCondition) conditionGroups {
	groups := conditionGroups{}

	for _, condition := range conditions {
		if condition.Condition == nil {
			continue
		}

		added := false

		groupStatus := condition.Status
		if condition.Polarity == NegativePolarity {
			switch groupStatus {
			case corev1.ConditionFalse:
				groupStatus = corev1.ConditionTrue
			case corev1.ConditionTrue:
				groupStatus = corev1.ConditionFalse
			case corev1.ConditionUnknown:
				groupStatus = corev1.ConditionUnknown
			}
		}
		for i := range groups {
			if groups[i].status == groupStatus && groups[i].severity == condition.Severity {
				groups[i].conditions = append(groups[i].conditions, condition)
				added = true
				break
			}
		}
		if !added {
			groups = append(groups, conditionGroup{
				conditions: []localizedCondition{condition},
				status:     groupStatus,
				severity:   condition.Severity,
			})
		}
	}

	sort.Sort(groups)

	if len(groups) > 0 {
		sort.Slice(groups[0].conditions, func(i, j int) bool {
			a := groups[0].conditions[i]
			b := groups[0].conditions[j]
			if a.Type != b.Type {
				return lexicographicLess(a.Condition, b.Condition)
			}
			return a.GetName() < b.GetName()
		})
	}

	return groups
}

type conditionGroups []conditionGroup

func (g conditionGroups) Len() int {
	return len(g)
}

func (g conditionGroups) Less(i, j int) bool {
	return g[i].mergePriority() < g[j].mergePriority()
}

func (g conditionGroups) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

// TopGroup returns the condition group with the highest mergePriority.
func (g conditionGroups) TopGroup() *conditionGroup {
	if len(g) == 0 {
		return nil
	}
	return &g[0]
}

// TrueGroup returns the condition group with status True, if any.
func (g conditionGroups) TrueGroup() *conditionGroup {
	return g.getByStatusAndSeverity(corev1.ConditionTrue, anywherev1.ConditionSeverityNone)
}

func (g conditionGroups) getByStatusAndSeverity(status corev1.ConditionStatus, severity anywherev1.ConditionSeverity) *conditionGroup {
	if len(g) == 0 {
		return nil
	}
	for _, group := range g {
		if group.status == status && group.severity == severity {
			return &group
		}
	}
	return nil
}

type conditionGroup struct {
	status     corev1.ConditionStatus
	severity   anywherev1.ConditionSeverity
	conditions []localizedCondition
}

func (g conditionGroup) mergePriority() int {
	switch g.status {
	case corev1.ConditionFalse:
		switch g.severity {
		case anywherev1.ConditionSeverityError:
			return 0
		case anywherev1.ConditionSeverityWarning:
			return 1
		case anywherev1.ConditionSeverityInfo:
			return 2
		}
	case corev1.ConditionTrue:
		return 3
	case corev1.ConditionUnknown:
		return 4
	}
	return 99
}

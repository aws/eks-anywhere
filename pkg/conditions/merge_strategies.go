package conditions

import (
	"fmt"
	"strings"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// mergeOptions allows to set strategies for merging a set of conditions into a single condition.
type mergeOptions struct {
	conditionTypes                     []anywherev1.ConditionType
	negativeConditionTypes             []anywherev1.ConditionType
	addSourceRef                       bool
	addStepCounter                     bool
	addStepCounterIfOnlyConditionTypes []anywherev1.ConditionType
	stepCounter                        int
}

// MergeOption defines an option for computing a summary of conditions.
type MergeOption func(*mergeOptions)

// WithConditions instructs merge about the condition types to consider when doing a merge operation.
func WithConditions(t ...anywherev1.ConditionType) MergeOption {
	return func(c *mergeOptions) {
		c.conditionTypes = t
	}
}

// WithNegativePolarityConditions instruct merge about which conditions should be considered having negative polarity.
func WithNegativePolarityConditions(t ...anywherev1.ConditionType) MergeOption {
	return func(c *mergeOptions) {
		c.negativeConditionTypes = t
	}
}

// WithStepCounter instructs merge to add a "x of y completed" string to the message.
func WithStepCounter() MergeOption {
	return func(c *mergeOptions) {
		c.addStepCounter = true
	}
}

// WithStepCounterIf adds a step counter if the value is true.
func WithStepCounterIf(value bool) MergeOption {
	return func(c *mergeOptions) {
		c.addStepCounter = value
	}
}

// WithStepCounterIfOnly ensure a step counter is show only if a subset of condition exists.
func WithStepCounterIfOnly(t ...anywherev1.ConditionType) MergeOption {
	return func(c *mergeOptions) {
		c.addStepCounterIfOnlyConditionTypes = t
	}
}

// AddSourceRef instructs merge to add info about the originating object to the target Reason.
func AddSourceRef() MergeOption {
	return func(c *mergeOptions) {
		c.addSourceRef = true
	}
}

// getReason returns the reason to be applied to the condition resulting by merging a set of condition groups.
func getReason(groups conditionGroups, options *mergeOptions) string {
	return getFirstReason(groups, options.conditionTypes, options.addSourceRef)
}

func getFirstReason(g conditionGroups, order []anywherev1.ConditionType, addSourceRef bool) string {
	if condition := getFirstCondition(g, order); condition != nil {
		reason := condition.Reason
		if addSourceRef {
			return localizeReason(reason, condition.Getter)
		}
		return reason
	}
	return ""
}

func localizeReason(reason string, from Getter) string {
	if strings.Contains(reason, "@") {
		return reason
	}
	return fmt.Sprintf("%s @ %s/%s", reason, from.GetObjectKind().GroupVersionKind().Kind, from.GetName())
}

// getMessage returns the message to be applied to the condition resulting by merging a set of condition groups.
func getMessage(groups conditionGroups, options *mergeOptions) string {
	if options.addStepCounter {
		return getStepCounterMessage(groups, options.stepCounter)
	}

	return getFirstMessage(groups, options.conditionTypes)
}

func getStepCounterMessage(groups conditionGroups, to int) string {
	ct := 0
	if trueGroup := groups.TrueGroup(); trueGroup != nil {
		ct = len(trueGroup.conditions)
	}
	return fmt.Sprintf("%d of %d completed", ct, to)
}

func getFirstMessage(groups conditionGroups, order []anywherev1.ConditionType) string {
	if condition := getFirstCondition(groups, order); condition != nil {
		return condition.Message
	}
	return ""
}

func getFirstCondition(g conditionGroups, priority []anywherev1.ConditionType) *localizedCondition {
	topGroup := g.TopGroup()
	if topGroup == nil {
		return nil
	}

	switch len(topGroup.conditions) {
	case 0:
		return nil
	case 1:
		return &topGroup.conditions[0]
	default:
		for _, p := range priority {
			for _, c := range topGroup.conditions {
				if c.Type == p {
					return &c
				}
			}
		}
		return &topGroup.conditions[0]
	}
}

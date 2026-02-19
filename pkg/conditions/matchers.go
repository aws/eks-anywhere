package conditions

import (
	"errors"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// HaveSameStateOf matches a condition to have the same state of another.
func HaveSameStateOf(expected *anywherev1.Condition) types.GomegaMatcher {
	return &conditionMatcher{
		Expected: expected,
	}
}

type conditionMatcher struct {
	Expected *anywherev1.Condition
}

func (matcher *conditionMatcher) Match(actual interface{}) (success bool, err error) {
	actualCondition, ok := actual.(*anywherev1.Condition)
	if !ok {
		return false, errors.New("value should be a condition")
	}

	return HasSameState(actualCondition, matcher.Expected), nil
}

func (matcher *conditionMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to have the same state of", matcher.Expected)
}

func (matcher *conditionMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have the same state of", matcher.Expected)
}

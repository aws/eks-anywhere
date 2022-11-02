package controller

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// Result represents the result of a reconciliation
// It allows to express intent for a reconciliation interruption
// without necessarily requeueing the request.
type Result struct {
	Result *ctrl.Result
}

// ToCtrlResult converts Result to a controller-runtime result.
func (r Result) ToCtrlResult() ctrl.Result {
	if r.Result == nil {
		return ctrl.Result{}
	}

	return *r.Result
}

// Return evaluates the intent of a Result to interrupt the reconciliation
// process or not.
func (r *Result) Return() bool {
	return r.Result != nil
}

// ResultWithReturn creates a new Result that interrupts the reconciliation
// without requeueing.
func ResultWithReturn() Result {
	return Result{Result: &ctrl.Result{}}
}

// ResultWithReturn creates a new Result that requeues the request after
// the provided duration.
func ResultWithRequeue(after time.Duration) Result {
	return Result{Result: &ctrl.Result{RequeueAfter: after}}
}

package controller

import ctrl "sigs.k8s.io/controller-runtime"

type Result struct {
	Result *ctrl.Result
}

func (r *Result) ToCtrlResult() ctrl.Result {
	if r.Result == nil {
		return ctrl.Result{}
	}

	return *r.Result
}

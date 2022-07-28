package validations

import "errors"

var errRunnerValidation = errors.New("validations failed")

type Validation func() *ValidationResult

type Runner struct {
	validations []Validation
}

func NewRunner() *Runner {
	return &Runner{validations: make([]Validation, 0)}
}

func (r *Runner) Register(validations ...Validation) {
	r.validations = append(r.validations, validations...)
}

func (r *Runner) Run() error {
	failed := false
	for _, v := range r.validations {
		result := v()
		result.Report()
		if result.Err != nil {
			failed = true
		}
	}
	r.validations = make([]Validation, 0)

	if failed {
		return errRunnerValidation
	}

	return nil
}

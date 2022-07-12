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
	for _, v := range r.validations {
		result := v()
		result.Report()
		if result.Err != nil {
			return errRunnerValidation
		}
	}

	return nil
}

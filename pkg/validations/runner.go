package validations

import "errors"

var errRunnerValidation = errors.New("validations failed")

type Validation func() *ValidationResult

type ValidationFunc func() error

type Runner struct {
	validations []Validation
	results     []*ValidationResult
}

func NewRunner() *Runner {
	return &Runner{
		validations: make([]Validation, 0),
		results:     make([]*ValidationResult, 0),
	}
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

	if failed {
		return errRunnerValidation
	}

	return nil
}

func (r *Runner) StoreValidationResults() error {
	failed := false
	for _, v := range r.validations {
		result := v()
		r.results = append(r.results, result)

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

func (r *Runner) ReportResults() {
	for _, result := range r.results {
		result.Report()
	}
}

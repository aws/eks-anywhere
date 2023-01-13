package validations

// ProcessValidationResults is currently used for unit test processing.
func ProcessValidationResults(validations []Validation) error {
	var errs []string
	results := make([]ValidationResult, 0, len(validations))
	for _, validation := range validations {
		results = append(results, *validation())
	}
	for _, result := range results {
		if result.Err != nil {
			errs = append(errs, result.Err.Error())
		} else if !result.Silent {
			result.LogPass()
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errs: errs}
	}
	return nil
}

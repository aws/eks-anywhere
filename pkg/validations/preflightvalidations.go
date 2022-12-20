package validations

func ProcessValidationResults(validations []ValidationResult) error {
	var errs []string
	for _, validation := range validations {
		if validation.Err != nil {
			errs = append(errs, validation.Err.Error())
		} else if !validation.Silent {
			validation.LogPass()
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errs: errs}
	}
	return nil
}

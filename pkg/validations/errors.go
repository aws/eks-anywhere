package validations

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Errs []string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation failed with %d errors: %s", len(v.Errs), strings.Join(v.Errs[:], ","))
}

func (v *ValidationError) String() string {
	return v.Error()
}

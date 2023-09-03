package validations

import (
	"unicode"

	"github.com/aws/eks-anywhere/pkg/logger"
)

type ValidationResult struct {
	Name        string
	Err         error
	Remediation string
	Silent      bool
}

func (v *ValidationResult) Report() {
	if v.Err != nil {
		logger.MarkFail("Validation failed", "validation", v.Name, "error", v.Err.Error(), "remediation", v.Remediation)
		return
	}
	if !v.Silent {
		v.LogPass()
	}
}

func (v *ValidationResult) LogPass() {
	logger.MarkPass(capitalize(v.Name))
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

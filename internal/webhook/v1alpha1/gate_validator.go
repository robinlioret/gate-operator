package v1alpha1

import (
	"fmt"
	"regexp"

	"github.com/robinlioret/gate-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var PascalCaseRegex = regexp.MustCompile("^[A-Z][A-Za-z0-9]*$")

func ValidateGateSpec(spec *v1alpha1.GateSpec) (admission.Warnings, error) {
	for _, target := range spec.Targets {
		if !PascalCaseRegex.MatchString(target.Name) {
			return nil, fmt.Errorf("target name must be PascalCase: %s", target.Name)
		}
		for _, validator := range target.Validators {
			if validator.MatchCondition.Type != "" {
				if !PascalCaseRegex.MatchString(validator.MatchCondition.Type) {
					return nil, fmt.Errorf("target name must be PascalCase: %s", validator.MatchCondition.Type)
				}
			}
		}
	}
	return nil, nil
}

/*
Copyright 2025 Robin LIORET.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package gate

import (
	"slices"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
)

func Evaluate(spec gateshv1alpha1.GateSpec) (bool, error) {
	return evaluateExpression(spec.Expression)
}

func evaluateExpression(expression gateshv1alpha1.GateExpression) (bool, error) {
	var result bool
	var err error

	if expression.Or != nil {
		subResults, err := evaluateChildrenExpressions(expression.Or)
		if err != nil {
			return false, err
		}
		result = slices.Contains(subResults, true)
	}

	if expression.And != nil {
		subResults, err := evaluateChildrenExpressions(expression.And)
		if err != nil {
			return false, err
		}
		for _, r := range subResults {
			if !r {
				result = false
				break
			}
		}
	}

	if isTargetValid(expression.Target) {
		result, err = evaluateTarget(expression.Target)
		if err != nil {
			return false, err
		}
	}

	if expression.Invert {
		result = !result
	}
	return result, nil
}

func evaluateChildrenExpressions(expressions []*gateshv1alpha1.GateExpressionWrap) ([]bool, error) {
	var results = make([]bool, len(expressions))
	for i, sub := range expressions {
		r, err := evaluateExpression(sub.GateExpression)
		if err != nil {
			return nil, err
		}
		results[i] = r
	}
	return results, nil
}

func isTargetValid(target gateshv1alpha1.GateTarget) bool {
	return target.ObjectRef.Kind != ""
}

func evaluateTarget(target gateshv1alpha1.GateTarget) (bool, error) {
	return true, nil
}

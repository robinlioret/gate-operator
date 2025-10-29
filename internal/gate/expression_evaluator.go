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
	"context"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
	"github.com/robinlioret/gate-operator/internal"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExpressionEvaluator struct {
	client.Client
	context.Context
	DefaultNamespace string
}

func NewExpressionEvaluator(ctx context.Context, c client.Client, defaultNamespace string) *ExpressionEvaluator {
	return &ExpressionEvaluator{
		Client:           c,
		Context:          ctx,
		DefaultNamespace: defaultNamespace,
	}
}

func (e *ExpressionEvaluator) Evaluate(gateSpec gateshv1alpha1.GateSpec) (bool, error) {
	result, err := e.evaluateExpression(gateSpec.Expression)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (e *ExpressionEvaluator) evaluateExpression(expression gateshv1alpha1.GateExpression) (bool, error) {
	var result bool
	var err error

	if expression.Or != nil {
		result = false
		for _, subExpression := range expression.Or {
			subResult, err := e.evaluateExpression(subExpression.GateExpression)
			if err != nil {
				return false, err
			}
			if subResult {
				result = true
				break
			}
		}
	}

	if expression.And != nil {
		result = true
		for _, subExpression := range expression.And {
			subResult, err := e.evaluateExpression(subExpression.GateExpression)
			if err != nil {
				return false, err
			}
			if !subResult {
				result = false
				break
			}
		}
	}

	if isValidTarget(expression.TargetOne) {
		result, err = e.evaluateTarget(expression.TargetOne)
		if err != nil {
			return false, err
		}
	}

	if expression.Invert {
		result = !result
	}

	return result, nil
}

func isValidTarget(target gateshv1alpha1.GateTargetOne) bool {
	return target.ObjectRef.Kind != ""
}

func (e *ExpressionEvaluator) evaluateTarget(target gateshv1alpha1.GateTargetOne) (bool, error) {
	obj, err := internal.GetReferencedObject(e.Context, e.Client, &target.ObjectRef, e.DefaultNamespace)
	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if target.Condition.Type != "" {
		conditions, err := internal.GetObjectStatusConditions(obj)
		if err != nil {
			return false, err
		}
		if conditions == nil {
			return false, nil
		}
		for _, condition := range conditions {
			if condition.Type == target.Condition.Type {
				return condition.Status == target.Condition.Status, nil
			}
		}
		return false, nil
	}

	// When only testing the presence of the object
	return true, nil
}

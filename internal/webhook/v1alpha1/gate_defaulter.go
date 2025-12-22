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

package v1alpha1

import (
	"strconv"
	"time"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DefaultEvaluationPeriod = &metav1.Duration{Duration: 60 * time.Second}
var DefaultConsolidationDelay = &metav1.Duration{Duration: 5 * time.Second}
var DefaultConsolidationCount = 1
var DefaultTargetValidators = []gateshv1alpha1.GateTargetValidator{{AtLeast: gateshv1alpha1.GateTargetValidatorAtLeast{Count: 1, Percent: 0}}}
var DefaultOperationOperator = gateshv1alpha1.GateOperatorAnd
var DefaultMatchConditionStatus = metav1.ConditionTrue

func ApplyDefaultSpec(spec *gateshv1alpha1.GateSpec) {
	if spec.EvaluationPeriod == nil {
		spec.EvaluationPeriod = DefaultEvaluationPeriod
	}
	if spec.Consolidation.Count == 0 {
		spec.Consolidation.Count = DefaultConsolidationCount
	}
	if spec.Consolidation.Delay == nil {
		spec.Consolidation.Delay = DefaultConsolidationDelay
	}
	if spec.Operation.Operator == "" {
		spec.Operation.Operator = DefaultOperationOperator
	}
	for idx := range spec.Targets {
		if spec.Targets[idx].Name == "" {
			spec.Targets[idx].Name = "Target" + strconv.Itoa(idx+1)
		}
		if spec.Targets[idx].Validators == nil {
			spec.Targets[idx].Validators = DefaultTargetValidators
		} else {
			for idx2 := range spec.Targets[idx].Validators {
				if spec.Targets[idx].Validators[idx2].MatchCondition.Type != "" && spec.Targets[idx].Validators[idx2].MatchCondition.Status == "" {
					spec.Targets[idx].Validators[idx2].MatchCondition.Status = DefaultMatchConditionStatus
				}
			}
		}
	}
}

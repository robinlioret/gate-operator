package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
	"github.com/robinlioret/gate-operator/internal"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const MessageSeparator = "\n"
const MessageObjectResult = "%s -> %s"

type TargetConditionReason string

const ReasonErrorWhileFetching = "ErrorWhileFetching"
const MessageErrorWhileFetching = "not able to fetch target objects: %s"
const ReasonOkObjectsFound = "ObjectsFound"
const MessageOkObjectsFound = "%d object(s) found"
const ReasonKoNoObjectsFound = "NoObjectsFound"
const MessageKoNoObjectFound = "no object found"
const ReasonOk = "ConditionMet"
const ReasonKo = "ConditionNotMet"

type GateCommonReconciler struct {
	context.Context
	client.Client
	*gateshv1alpha1.Gate
}

type TargetObjectResult struct {
	Result  bool
	Message string
}

func (g *GateCommonReconciler) Reconcile() error {
	result, targetConditions := g.EvaluateSpec()
	g.UpdateGateStatusFromResult(result, targetConditions)
	g.Gate.Status.NextEvaluation = metav1.Time{Time: time.Now().Add(g.Gate.Spec.RequeueAfter.Duration)}
	return nil
}

func (g *GateCommonReconciler) GetObjectName(object unstructured.Unstructured) string {
	return fmt.Sprintf("%s/%s", object.GetNamespace(), object.GetName())
}

func (g *GateCommonReconciler) UpdateGateStatusFromResult(
	result bool,
	targetConditions []metav1.Condition,
) {
	var message string
	var reason string
	var openedCondition metav1.ConditionStatus
	var closedCondition metav1.ConditionStatus

	if result {
		g.Gate.Status.State = gateshv1alpha1.GateStateOpened
		message = "Gate was evaluated to true"
		reason = "GateConditionMet"
		openedCondition = metav1.ConditionTrue
		closedCondition = metav1.ConditionFalse
	} else {
		g.Gate.Status.State = gateshv1alpha1.GateStateClosed
		message = "Gate was evaluated to false"
		reason = "GateConditionNotMet"
		openedCondition = metav1.ConditionFalse
		closedCondition = metav1.ConditionTrue
	}

	// Opened conditions
	meta.SetStatusCondition(&g.Gate.Status.Conditions, metav1.Condition{Type: gateshv1alpha1.GateStateOpened, Status: openedCondition, Reason: reason, Message: message})
	//meta.SetStatusCondition(&g.Gate.Status.Conditions, metav1.Condition{Type: "Available", Status: openedCondition, Reason: reason, Message: message})

	// Closed conditions
	meta.SetStatusCondition(&g.Gate.Status.Conditions, metav1.Condition{Type: gateshv1alpha1.GateStateClosed, Status: closedCondition, Reason: reason, Message: message})
	//meta.SetStatusCondition(&g.Gate.Status.Conditions, metav1.Condition{Type: "Progressing", Status: closedCondition, Reason: reason, Message: message})

	g.Gate.Status.TargetConditions = targetConditions
}

func (g *GateCommonReconciler) EvaluateSpec() (bool, []metav1.Condition) {
	targetConditions := make([]metav1.Condition, 0)
	for _, target := range g.Gate.Spec.Targets {
		meta.SetStatusCondition(&targetConditions, g.EvaluateTarget(&target))
	}
	result := g.ComputeOperation(targetConditions)
	return result, targetConditions
}

func (g *GateCommonReconciler) EvaluateTarget(target *gateshv1alpha1.GateTarget) metav1.Condition {
	log := logf.FromContext(g.Context)

	objects, err := internal.FetchGateTargetObjects(g.Context, g.Client, target, g.Gate.Namespace)
	if err != nil {
		log.Error(err, "unable to fetch target objects")
		return metav1.Condition{Type: target.TargetName, Status: metav1.ConditionFalse, Reason: ReasonErrorWhileFetching, Message: fmt.Sprintf(MessageErrorWhileFetching, err.Error())}
	}

	if len(objects) == 0 {
		return metav1.Condition{Type: target.TargetName, Status: metav1.ConditionFalse, Reason: ReasonKoNoObjectsFound, Message: MessageKoNoObjectFound}
	}

	if target.ExistsOnly {
		return metav1.Condition{Type: target.TargetName, Status: metav1.ConditionTrue, Reason: ReasonOkObjectsFound, Message: fmt.Sprintf(MessageOkObjectsFound, len(objects))}
	}

	objectResults := make(map[string]TargetObjectResult)
	for _, object := range objects {
		objectResults[g.GetObjectName(object)] = g.EvaluateTargetObjectCondition(object, target)
	}

	finalResult, message := g.ComputeTargetResult(objects, objectResults)
	if finalResult {
		return metav1.Condition{Type: target.TargetName, Status: metav1.ConditionTrue, Reason: ReasonOk, Message: message}
	} else {
		return metav1.Condition{Type: target.TargetName, Status: metav1.ConditionFalse, Reason: ReasonKo, Message: message}
	}
}

func (g *GateCommonReconciler) ComputeTargetResult(
	objects []unstructured.Unstructured,
	objectResults map[string]TargetObjectResult,
) (bool, string) {
	finalResult := true
	message := []string{fmt.Sprintf(MessageOkObjectsFound, len(objects))}
	for objName, result := range objectResults {
		if !result.Result {
			message = append(message, fmt.Sprintf(MessageObjectResult, objName, result.Message))
			finalResult = false
		}
	}
	return finalResult, strings.Join(message, MessageSeparator)
}

func (g *GateCommonReconciler) EvaluateTargetObjectCondition(
	object unstructured.Unstructured,
	target *gateshv1alpha1.GateTarget,
) TargetObjectResult {
	conditions, err := internal.GetObjectStatusConditions(&object)
	if err != nil {
		return TargetObjectResult{Result: false, Message: fmt.Sprintf("error while fetching object conditions: %s", err.Error())}
	}

	if len(conditions) == 0 {
		return TargetObjectResult{Result: false, Message: "no conditions found"}
	}

	condition := meta.FindStatusCondition(conditions, target.DesiredCondition.Type)
	if condition == nil {
		return TargetObjectResult{Result: false, Message: "desired condition not found"}
	}

	if condition.Status != target.DesiredCondition.Status {
		return TargetObjectResult{Result: false, Message: "desired condition status not equal"}
	}

	return TargetObjectResult{Result: true, Message: "condition matches"}
}

func (g *GateCommonReconciler) ComputeOperation(targetConditions []metav1.Condition) bool {
	switch g.Gate.Spec.Operation.Operator {
	case gateshv1alpha1.GateOperatorOr:
		for _, targetCondition := range targetConditions {
			if targetCondition.Status == metav1.ConditionTrue {
				// logf.FromContext(ctx).Info("Or operation to true")
				return true
			}
		}
		// logf.FromContext(ctx).Info("Or operation to false")
		return false

	default: // And
		for _, targetCondition := range targetConditions {
			if targetCondition.Status != metav1.ConditionTrue {
				// logf.FromContext(ctx).Info(fmt.Sprintf("And operation to false (%s)", targetCondition.Type))
				return false
			}
		}
		// logf.FromContext(ctx).Info("And operation to true")
		return true
	}
}

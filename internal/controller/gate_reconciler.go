package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	Context context.Context
	Client  client.Client
	Gate    *gateshv1alpha1.Gate
}

type TargetObjectResult struct {
	Result  bool
	Message string
}

func (g *GateCommonReconciler) Reconcile() error {
	log := logf.FromContext(g.Context)
	log.Info(fmt.Sprintf("Start reconciling %s %s", g.Gate.Kind, g.Gate.Name))
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

	meta.SetStatusCondition(&g.Gate.Status.Conditions, metav1.Condition{Type: gateshv1alpha1.GateStateOpened, Status: openedCondition, Reason: reason, Message: message})
	meta.SetStatusCondition(&g.Gate.Status.Conditions, metav1.Condition{Type: gateshv1alpha1.GateStateClosed, Status: closedCondition, Reason: reason, Message: message})
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

	objects, err := g.FetchGateTargetObjects(target)
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
	conditions, err := g.GetObjectStatusConditions(&object)
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

// FetchGateTargetObjects retrieves Kubernetes objects based on the GateTarget specification.
// It handles both Name-based and LabelSelector-based lookups and returns a slice of unstructured objects.
func (g *GateCommonReconciler) FetchGateTargetObjects(gateTarget *gateshv1alpha1.GateTarget) ([]unstructured.Unstructured, error) {
	// Determine the namespace to use
	namespace := gateTarget.Namespace
	if namespace == "" {
		namespace = g.Gate.Namespace
	}

	// Create GroupVersionKind for the target resource
	gvk := schema.GroupVersionKind{
		Group:   "", // Will be set based on ApiVersion
		Version: gateTarget.ApiVersion,
		Kind:    gateTarget.Kind,
	}

	// Parse Group and Version from ApiVersion
	gv, err := schema.ParseGroupVersion(gateTarget.ApiVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid ApiVersion %s: %w", gateTarget.ApiVersion, err)
	}
	gvk.Group = gv.Group
	gvk.Version = gv.Version

	// Validate that either Name or LabelSelector is set, but not both
	if gateTarget.Name != "" && !g.IsLabelSelectorEmpty(gateTarget.LabelSelector) {
		return nil, fmt.Errorf("name and labelSelector are mutually exclusive in GateTarget %s", gateTarget.TargetName)
	}
	if gateTarget.Name == "" && g.IsLabelSelectorEmpty(gateTarget.LabelSelector) {
		return nil, fmt.Errorf("either name or labelSelector must be specified in GateTarget %s", gateTarget.TargetName)
	}

	var objects []unstructured.Unstructured

	if gateTarget.Name != "" {
		// Case 1: Fetch by Name
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gvk)
		err := g.Client.Get(g.Context, client.ObjectKey{
			Namespace: namespace,
			Name:      gateTarget.Name,
		}, obj)
		if err != nil {
			if errors.IsNotFound(err) {
				return objects, nil // Return empty slice if not found
			}
			return nil, fmt.Errorf("failed to get object %s/%s: %w", namespace, gateTarget.Name, err)
		}
		objects = append(objects, *obj)
	} else {
		// Case 2: Fetch by LabelSelector
		selector, err := metav1.LabelSelectorAsSelector(&gateTarget.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("invalid label selector in GateTarget %s: %w", gateTarget.TargetName, err)
		}

		// Create a list to hold the results
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gvk.Group,
			Version: gvk.Version,
			Kind:    gvk.Kind + "List", // Append "List" for the list kind
		})

		// List options with the label selector
		listOptions := &client.ListOptions{
			Namespace:     namespace,
			LabelSelector: selector,
		}

		// Fetch the list of objects
		err = g.Client.List(g.Context, list, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects for GateTarget %s: %w", gateTarget.TargetName, err)
		}

		objects = append(objects, list.Items...)
	}

	return objects, nil
}

// IsLabelSelectorEmpty checks if the LabelSelector is empty
func (g *GateCommonReconciler) IsLabelSelectorEmpty(selector metav1.LabelSelector) bool {
	return len(selector.MatchLabels) == 0 && len(selector.MatchExpressions) == 0
}

func (g *GateCommonReconciler) GetObjectStatusConditions(obj client.Object) ([]metav1.Condition, error) {
	unstrObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expected unstructured.Unstructured, got %T", obj)
	}

	status, found, err := unstructured.NestedMap(unstrObj.Object, "status")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil // No status field; return empty slice
	}

	// Extract the conditions field from status
	conditions, found, err := unstructured.NestedSlice(status, "conditions")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil // No conditions field; return empty slice
	}

	conditionList := []metav1.Condition{}
	for _, cond := range conditions {
		condition, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		// Marshal the condition map to JSON
		condBytes, err := json.Marshal(condition)
		if err != nil {
			continue
		}

		// Unmarshal into metav1.Condition
		var metaCond metav1.Condition
		if err := json.Unmarshal(condBytes, &metaCond); err != nil {
			continue
		}

		conditionList = append(conditionList, metaCond)
	}

	return conditionList, nil
}

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/jsonpointer"
	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type TargetConditionReason string

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
		return metav1.Condition{Type: target.Name, Status: metav1.ConditionFalse, Reason: "ErrorWhileFetching", Message: fmt.Sprintf("not able to fetch target objects: %s", err.Error())}
	}

	var message []string
	atLeast := -1
	results := make([]bool, len(objects))
	for i := range results {
		results[i] = true
	}

	for _, validator := range target.Validators {
		atLeastCount := atLeast
		if validator.AtLeast.Count > 0 {
			atLeastCount = validator.AtLeast.Count
		}
		atLeastPercent := atLeast
		if validator.AtLeast.Percent > 0 {
			atLeastPercent = validator.AtLeast.Percent * len(objects) / 100
		}
		atLeast = max(atLeastCount, atLeastPercent, atLeast)

		if validator.MatchCondition.Type != "" {
			message = g.EvaluateTargetMatchCondition(objects, results, message, validator)
		}
		if validator.JsonPointer.Pointer != "" {
			message = g.EvaluateTargetJsonPointer(objects, results, message, validator)
		}
	}

	result, message := g.ComputeTargetEvaluationResult(atLeast, results, message)
	status := metav1.ConditionFalse
	reason := "ConditionNotMet"
	if result {
		status = metav1.ConditionTrue
		reason = "ConditionMet"
	}
	return metav1.Condition{Type: target.Name, Status: status, Reason: reason, Message: strings.Join(message, "\n")}
}

func (g *GateCommonReconciler) EvaluateTargetMatchCondition(objects []unstructured.Unstructured, results []bool, message []string, validator gateshv1alpha1.GateTargetValidator) []string {
	for idx, object := range objects {
		objectConditions, err := g.GetObjectStatusConditions(&object)
		if err != nil {
			results[idx] = false
			message = append(message, fmt.Sprintf("[%s] error while fetching condition %s: %s", g.GetObjectName(object), validator.MatchCondition.Type, err.Error()))
			continue
		}
		if len(objectConditions) == 0 {
			results[idx] = false
			message = append(message, fmt.Sprintf("[%s] object doesn't have conditions", g.GetObjectName(object)))
			continue
		}

		condition := meta.FindStatusCondition(objectConditions, validator.MatchCondition.Type)
		if condition == nil {
			results[idx] = false
			message = append(message, fmt.Sprintf("[%s] condition %s is missing", g.GetObjectName(object), validator.MatchCondition.Type))
			continue
		}
		if condition.Status != validator.MatchCondition.Status {
			results[idx] = false
			message = append(message, fmt.Sprintf("[%s] condition %s is wrong (expected %s, got %s)", g.GetObjectName(object), validator.MatchCondition.Type, string(validator.MatchCondition.Status), string(condition.Status)))
			continue
		}
	}
	return message
}

func (g *GateCommonReconciler) EvaluateTargetJsonPointer(objects []unstructured.Unstructured, results []bool, message []string, validator gateshv1alpha1.GateTargetValidator) []string {
	for idx, object := range objects {
		fieldValue, err := g.GetObjectFieldByJsonPointer(&object, validator.JsonPointer.Pointer)
		if err != nil {
			results[idx] = false
			message = append(message, fmt.Sprintf("[%s] error while fetching field value for the JSON Pointer %s: %s", g.GetObjectName(object), validator.JsonPointer.Pointer, err.Error()))
			continue
		}

		if fieldValue != validator.JsonPointer.Value {
			results[idx] = false
			message = append(message, fmt.Sprintf("[%s] field value not matching expected for the JSON Pointer %s '%s', got '%s'", g.GetObjectName(object), validator.JsonPointer.Pointer, validator.JsonPointer.Value, fieldValue))
			continue
		}
	}
	return message
}

func (g *GateCommonReconciler) ComputeTargetEvaluationResult(atLeast int, results []bool, message []string) (bool, []string) {
	objectsCount := len(results)
	if atLeast <= 0 {
		// If not specified, need at least one object or all the found objects to match.
		atLeast = max(1, objectsCount)
	}
	count := 0
	for _, result := range results {
		if result {
			count++
		}
	}

	message = append(message,
		fmt.Sprintf("%d objects found", objectsCount),
		fmt.Sprintf("%d objects match target validators", count),
		fmt.Sprintf("%d/%d valid objects", count, atLeast),
	)
	return count >= atLeast, message
}

func (g *GateCommonReconciler) ComputeOperation(targetConditions []metav1.Condition) bool {
	var result bool
	switch g.Gate.Spec.Operation.Operator {
	case gateshv1alpha1.GateOperatorOr:
		result = false
		for _, targetCondition := range targetConditions {
			if targetCondition.Status == metav1.ConditionTrue {
				result = true
				break
			}
		}

	default: // And
		result = true
		for _, targetCondition := range targetConditions {
			if targetCondition.Status != metav1.ConditionTrue {
				result = false
				break
			}
		}
	}
	if g.Gate.Spec.Operation.Invert {
		result = !result
	}
	return result
}

// FetchGateTargetObjects retrieves Kubernetes objects based on the GateTarget specification.
// It handles both Name-based and LabelSelector-based lookups and returns a slice of unstructured objects.
func (g *GateCommonReconciler) FetchGateTargetObjects(gateTarget *gateshv1alpha1.GateTarget) ([]unstructured.Unstructured, error) {
	// Determine the namespace to use
	namespace := gateTarget.Selector.Namespace
	if namespace == "" {
		namespace = g.Gate.Namespace
	}

	// Create GroupVersionKind for the target resource
	gvk := schema.GroupVersionKind{
		Group:   "", // Will be set based on ApiVersion
		Version: gateTarget.Selector.ApiVersion,
		Kind:    gateTarget.Selector.Kind,
	}

	// Parse Group and Version from ApiVersion
	gv, err := schema.ParseGroupVersion(gateTarget.Selector.ApiVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid ApiVersion %s: %w", gateTarget.Selector.ApiVersion, err)
	}
	gvk.Group = gv.Group
	gvk.Version = gv.Version

	// Validate that either Name or LabelSelector is set, but not both
	if gateTarget.Selector.Name != "" && !g.IsLabelSelectorEmpty(gateTarget.Selector.LabelSelector) {
		return nil, fmt.Errorf("name and labelSelector are mutually exclusive in GateTarget %s", gateTarget.Selector.Name)
	}
	if gateTarget.Selector.Name == "" && g.IsLabelSelectorEmpty(gateTarget.Selector.LabelSelector) {
		return nil, fmt.Errorf("either name or labelSelector must be specified in GateTarget %s", gateTarget.Selector.Name)
	}

	var objects []unstructured.Unstructured

	if gateTarget.Selector.Name != "" {
		// Case 1: Fetch by Name
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gvk)
		err := g.Client.Get(g.Context, client.ObjectKey{
			Namespace: namespace,
			Name:      gateTarget.Selector.Name,
		}, obj)
		if err != nil {
			if errors.IsNotFound(err) {
				return objects, nil // Return empty slice if not found
			}
			return nil, fmt.Errorf("failed to get object %s/%s: %w", namespace, gateTarget.Selector.Name, err)
		}
		objects = append(objects, *obj)
	} else {
		// Case 2: Fetch by LabelSelector
		selector, err := metav1.LabelSelectorAsSelector(&gateTarget.Selector.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("invalid label selector in GateTarget %s: %w", gateTarget.Selector.Name, err)
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
			return nil, fmt.Errorf("failed to list objects for GateTarget %s: %w", gateTarget.Selector.Name, err)
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

func (g *GateCommonReconciler) GetObjectFieldByJsonPointer(obj client.Object, jsonPointer string) (interface{}, error) {
	unstrObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expected unstructured.Unstructured, got %T", obj)
	}

	pointer, err := jsonpointer.New(jsonPointer)
	if err != nil {
		return nil, fmt.Errorf("invalid pointer: %s", err)
	}

	data := unstrObj.UnstructuredContent()
	value, _, err := pointer.Get(data)
	if err != nil {
		return nil, fmt.Errorf("failed to get value: %s", err)
	}

	return value, nil
}

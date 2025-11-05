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
package internal

import (
	"context"
	"encoding/json"
	"fmt"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchGateTargetObjects retrieves Kubernetes objects based on the GateTarget specification.
// It handles both Name-based and LabelSelector-based lookups and returns a slice of unstructured objects.
func FetchGateTargetObjects(
	ctx context.Context,
	cl client.Client,
	gateTarget *gateshv1alpha1.GateTarget,
	defaultNamespace string,
) ([]unstructured.Unstructured, error) {
	// Determine the namespace to use
	namespace := gateTarget.Namespace
	if namespace == "" {
		namespace = defaultNamespace
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
	if gateTarget.Name != "" && !isLabelSelectorEmpty(gateTarget.LabelSelector) {
		return nil, fmt.Errorf("name and labelSelector are mutually exclusive in GateTarget %s", gateTarget.TargetName)
	}
	if gateTarget.Name == "" && isLabelSelectorEmpty(gateTarget.LabelSelector) {
		return nil, fmt.Errorf("either name or labelSelector must be specified in GateTarget %s", gateTarget.TargetName)
	}

	var objects []unstructured.Unstructured

	if gateTarget.Name != "" {
		// Case 1: Fetch by Name
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gvk)
		err := cl.Get(ctx, client.ObjectKey{
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
		err = cl.List(ctx, list, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects for GateTarget %s: %w", gateTarget.TargetName, err)
		}

		objects = append(objects, list.Items...)
	}

	return objects, nil
}

// isLabelSelectorEmpty checks if the LabelSelector is empty
func isLabelSelectorEmpty(selector metav1.LabelSelector) bool {
	return len(selector.MatchLabels) == 0 && len(selector.MatchExpressions) == 0
}

func GetObjectStatusConditions(obj client.Object) ([]metav1.Condition, error) {
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

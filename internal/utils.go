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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetReferencedObject(
	ctx context.Context,
	r client.Reader,
	ref *corev1.ObjectReference,
	namespace string,
) (client.Object, error) {
	if ref == nil {
		return nil, errors.NewBadRequest("ObjectReference is nil")
	}

	gv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return nil, err
	}

	gvk := schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    ref.Kind,
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)

	ns := ref.Namespace
	if ns == "" {
		ns = namespace
	}

	key := client.ObjectKey{Name: ref.Name, Namespace: ns}
	if err := r.Get(ctx, key, obj); err != nil {
		return nil, err
	}

	return obj, nil
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

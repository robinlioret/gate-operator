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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GateTarget defines how objects should be evaluated to validate the expression
type GateTarget struct {
	// Base reference to the object(s) to evaluate. It can match multiple objects in the cluster
	ObjectRef v1.ObjectReference `json:"objectRef"`

	// Select objects among the one matching the objectRef field using the labels
	// ++kubebuilder:validation:Optional
	// +kubebuilder:default:value=nil
	Selector metav1.LabelSelector `json:"selector,omitempty"`
}

// GateExpression defines the conditions for the gate to be available
// +kubebuilder:validation:XValidation:rule="has(self.target) || has(self.and) || has(self.or)",message="At least one of 'target', 'and', 'or' must be specified"
type GateExpression struct {
	// Target to evaluate
	// ++kubebuilder:validation:Optional
	// +kubebuilder:default:value=nil
	Target GateTarget `json:"target,omitempty"`

	// If true, inverts the result of the target
	// ++kubebuilder:validation:Optional
	// +kubebuilder:default:value=false
	// +kubebuilder:example=true
	Invert bool `json:"invert,omitempty"`

	// Apply AND logical operator to the expressions
	// ++kubebuilder:validation:Optional
	// +kubebuilder:default:value=nil
	And []GateExpression `json:"and,omitempty"`

	// Apply AND logical operator to the expressions
	// ++kubebuilder:validation:Optional
	// +kubebuilder:default:value=nil
	Or []GateExpression `json:"or,omitempty"`
}

// GateSpec defines the desired state of Gate
type GateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// The set of conditions to make the Gate available
	Expression GateExpression `json:"expression,omitempty"`
}

// GateStatus defines the observed state of Gate.
type GateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the Gate resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Gate is the Schema for the gates API
type Gate struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of Gate
	// +required
	Spec GateSpec `json:"spec"`

	// status defines the observed state of Gate
	// +optional
	Status GateStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// GateList contains a list of Gate
type GateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gate{}, &GateList{})
}

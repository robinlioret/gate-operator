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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type GateOperator = string

const (
	GateOperatorAnd GateOperator = "And"
	GateOperatorOr  GateOperator = "Or"
)

type GateState = string

const (
	GateStateOpened GateState = "Opened"
	GateStateClosed GateState = "Closed"
)

type GateOperation struct {
	// Operation to perform. By default, it is "And".
	// +kubebuilder:validation:Enum=And;Or
	Operator GateOperator `json:"operator,omitempty"`

	// If true, will invert the result of the Gate. By default, false.
	Invert bool `json:"invert,omitempty"`
}

// GateTargetSelector defines how to select resources to evaluate for the target.
// +kubebuilder:validation:XValidation:rule="(has(self.name) ? 1 : 0) + (has(self.labelSelector) ? 1 : 0) == 1",message="Invalid target specification: you must provide exactly one of 'name' (for a single resource) or 'labelSelector' (for multiple resources). Providing both or neither is not allowed."
type GateTargetSelector struct {
	// Kind of the resource(s) to target
	// +required
	Kind string `json:"kind"`

	// ApiVersion of the resource(s) to target
	// +required
	ApiVersion string `json:"apiVersion"`

	// Namespace of the resource(s) to target. By default, the namespace of the gate if relevant.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Name of the resource to target. Incompatible with label selection.
	// +optional
	Name string `json:"name,omitempty"`

	// Select the resources using labels. Incompatible with name selection.
	// +optional
	LabelSelector metav1.LabelSelector `json:"labelSelector,omitempty"`
}

// GateTargetValidatorMatchCondition defines what condition is desired on the target objects.
type GateTargetValidatorMatchCondition struct {
	// Type of the kubernetes conditions.
	// +required
	Type string `json:"type"`

	// Type of the kubernetes conditions. By default, "True"
	// +optional
	Status metav1.ConditionStatus `json:"status"`
}

// GateTargetValidator defines a part of the logic to evaluate the target.
// +kubebuilder:validation:XValidation:rule="(has(self.atLeast) ? 1 : 0) + (has(self.matchCondition) ? 1 : 0) == 1",message="The validator must have exactly one key."
type GateTargetValidator struct {
	// Validate the target if at least N objects matches other validator. If no validator are provided, will check if
	// at least N object was found. If not specified alongside other validator, all found objects must match them.
	// +optional
	AtLeast uint `json:"atLeast,omitempty"`

	// Desired condition of the resources.
	// +optional
	MatchCondition GateTargetValidatorMatchCondition `json:"matchCondition,omitempty"`
}

// GateTarget defines the conditions for the gate to be available
type GateTarget struct {
	// Name of the target. Must be PascalCase. This will be used to make the matching target condition humanly
	// identifiable.
	// +kubebuilder:validation:Pattern=`^[A-Z][a-zA-Z0-9]*$`
	// +required
	Name string `json:"name"`

	// Selector
	// +required
	Selector GateTargetSelector `json:"selector"`

	// Validators defines how the target should be validated. By default, the target will be validated if at least one
	// object was found by the selector regardless of its state.
	// +optional
	Validators []GateTargetValidator `json:"validators,omitempty"`
}

// GateSpec defines the desired state of Gate
type GateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// The set of conditions to make the Gate ready.
	// +kubebuilder:validation:MinItems:=1
	// +required
	Targets []GateTarget `json:"targets"`

	// Indicates how to combine the targets results. By default, they will simply be anded.
	// +optional
	Operation GateOperation `json:"operation"`

	// Defines the duration between evaluations of a Gate. By default, 60 seconds
	// +optional
	RequeueAfter *metav1.Duration `json:"requeueAfter,omitempty"`
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

	// List of the targets conditions
	// +optional
	TargetConditions []metav1.Condition `json:"targetConditions,omitempty"`

	// lastSuccessfulEvaluation defines when was the last time the gate was successfully evaluated.
	// +optional
	NextEvaluation metav1.Time `json:"nextEvaluation,omitempty"`

	// Easy access field representing the gate's condition
	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Gate is the Schema for the gates API
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=`.status.state`
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

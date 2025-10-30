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

type GateOperation struct {
	// Operation to perform. By default, it is "And".
	// +kubebuilder:validation:Enum=And;Or
	// +default:value="And"
	Operator GateOperator `json:"operator,omitempty"`

	// TODO: add threshold, and other operations possibilties
}

type GateTargetCondition struct {
	// Type of the kubernetes conditions.
	// +default:value="Available"
	Type string `json:"type"`

	// Type of the kubernetes conditions.
	// +default:value="True"
	Status metav1.ConditionStatus `json:"status"`
}

// GateExpression defines the conditions for the gate to be available
type GateTarget struct {
	// TargetName is the name of the target. Must be PascalCase.
	// +kubebuilder:validation:Pattern=`^[A-Z][a-zA-Z0-9]*$`
	// +required
	TargetName string `json:"targetName"`

	// Kind of the resource(s) to target
	// +required
	Kind string `json:"kind"`

	// ApiVersion of the resource(s) to target
	// +required
	ApiVersion string `json:"apiVersion"`

	// Namespace of the resource(s) to target. By default, the namespace of the gate.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Name of the resource to target. Incompatible with label selection.
	// +optional
	Name string `json:"name,omitempty"`

	// Select the resources using labels. Incompatible with name selection.
	// +optional
	LabelSelector metav1.LabelSelector `json:"labelSelector,omitempty"`

	// If true, the target will be validated if the resources is found regardless of its condition.
	// If true and using labels selection, this will validate the target if at least one resource is found.
	// By default, false.
	// +optional
	// +default:value=false
	ExistsOnly bool `json:"existsOnly,omitempty"`

	// DesiredCondition of the resources. By default, will look for Available to be "True".
	// +optional
	DesiredCondition GateTargetCondition `json:"desiredCondition,omitempty"`

	// TODO: add threshold, and other operations possibilties
}

// GateSpec defines the desired state of Gate
type GateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// The set of conditions to make the Gate ready.
	// +required
	Targets []GateTarget `json:"targets"`

	// Indicates how to combine the targets results. By default, they will simply be anded.
	// +optional
	Operation GateOperation `json:"operation"`

	// Defines the duration between evaluations of a Gate. By default, 60 seconds
	// +optional
	// +default:value="60s"
	RequeueAfter *metav1.Duration `json:"requeueAfter,omitempty"`
}

type GateState = string

const (
	GateStateOpened GateState = "Opened"
	GateStateClosed GateState = "Closed"
)

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

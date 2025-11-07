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
	"context"
	"fmt"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var clustergatelog = logf.Log.WithName("clustergate-resource")

// SetupClusterGateWebhookWithManager registers the webhook for ClusterGate in the manager.
func SetupClusterGateWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&gateshv1alpha1.ClusterGate{}).
		WithValidator(&ClusterGateCustomValidator{}).
		WithDefaulter(&ClusterGateCustomDefaulter{
			DefaultRequeueAfter:     &metav1.Duration{Duration: 60 * time.Second},
			DefaultTargetValidators: []gateshv1alpha1.GateTargetValidator{{AtLeast: 1}},
			DefaultOperation:        gateshv1alpha1.GateOperation{Operator: gateshv1alpha1.GateOperatorAnd},
		}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-gate-sh-v1alpha1-clustergate,mutating=true,failurePolicy=fail,sideEffects=None,groups=gate.sh,resources=clustergates,verbs=create;update,versions=v1alpha1,name=mclustergate-v1alpha1.kb.io,admissionReviewVersions=v1

// ClusterGateCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind ClusterGate when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ClusterGateCustomDefaulter struct {
	DefaultRequeueAfter     *metav1.Duration
	DefaultTargetValidators []gateshv1alpha1.GateTargetValidator
	DefaultOperation        gateshv1alpha1.GateOperation
}

var _ webhook.CustomDefaulter = &ClusterGateCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind ClusterGate.
func (d *ClusterGateCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	clustergate, ok := obj.(*gateshv1alpha1.ClusterGate)

	if !ok {
		return fmt.Errorf("expected an ClusterGate object but got %T", obj)
	}
	clustergatelog.Info("Defaulting for ClusterGate", "name", clustergate.GetName())
	d.ApplyDefault(&clustergate.Spec)
	return nil
}

func (d *ClusterGateCustomDefaulter) ApplyDefault(spec *gateshv1alpha1.GateSpec) {
	if spec.RequeueAfter == nil {
		spec.RequeueAfter = d.DefaultRequeueAfter
	}
	for idx := range spec.Targets {
		if spec.Targets[idx].Name == "" {
			spec.Targets[idx].Name = "Target" + strconv.Itoa(idx+1)
		}
		if len(spec.Targets[idx].Validators) == 0 {
			spec.Targets[idx].Validators = d.DefaultTargetValidators
		}
	}
	if spec.Operation.Operator == "" {
		spec.Operation = d.DefaultOperation
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-gate-sh-v1alpha1-clustergate,mutating=false,failurePolicy=fail,sideEffects=None,groups=gate.sh,resources=clustergates,verbs=create;update,versions=v1alpha1,name=vclustergate-v1alpha1.kb.io,admissionReviewVersions=v1

// ClusterGateCustomValidator struct is responsible for validating the ClusterGate resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ClusterGateCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &ClusterGateCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ClusterGate.
func (v *ClusterGateCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	clustergate, ok := obj.(*gateshv1alpha1.ClusterGate)
	if !ok {
		return nil, fmt.Errorf("expected a ClusterGate object but got %T", obj)
	}
	clustergatelog.Info("Validation for ClusterGate upon creation", "name", clustergate.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ClusterGate.
func (v *ClusterGateCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	clustergate, ok := newObj.(*gateshv1alpha1.ClusterGate)
	if !ok {
		return nil, fmt.Errorf("expected a ClusterGate object for the newObj but got %T", newObj)
	}
	clustergatelog.Info("Validation for ClusterGate upon update", "name", clustergate.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ClusterGate.
func (v *ClusterGateCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	clustergate, ok := obj.(*gateshv1alpha1.ClusterGate)
	if !ok {
		return nil, fmt.Errorf("expected a ClusterGate object but got %T", obj)
	}
	clustergatelog.Info("Validation for ClusterGate upon deletion", "name", clustergate.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}

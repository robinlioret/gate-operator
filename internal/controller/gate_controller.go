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

package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
)

// GateReconciler reconciles a Gate object
type GateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=gate.sh,resources=gates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gate.sh,resources=gates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gate.sh,resources=gates/finalizers,verbs=update
// +kubebuilder:rbac:groups="*",resources="*",verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Gate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.3/pkg/reconcile
func (r *GateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Get the object by reference
	var gate gateshv1alpha1.Gate
	err := r.Get(ctx, req.NamespacedName, &gate)
	if errors.IsNotFound(err) {
		log.Info("Gate not found")
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "unable to fetch Gate")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Do not evaluate if it was updated too recently
	if gate.Status.NextEvaluation.After(time.Now()) {
		log.V(1).Info("Gate was already processed recently")
		return ctrl.Result{}, nil
	}

	result, err := r.EvaluateGateSpec(ctx, &gate)
	if err != nil {
		log.Error(err, "unable to evaluate gate")
		return ctrl.Result{}, err
	}
	r.UpdateGateStatusFromResult(result, &gate.Status)

	gate.Status.NextEvaluation = metav1.Time{Time: time.Now().Add(gate.Spec.RequeueAfter.Duration)}
	if err := r.Status().Update(ctx, &gate); err != nil {
		log.Error(err, "unable to update Gate")
		return ctrl.Result{}, err
	}
	log.V(1).Info("Gate processed successfully")
	return ctrl.Result{RequeueAfter: gate.Spec.RequeueAfter.Duration}, nil
}

func (r *GateReconciler) UpdateGateStatusFromResult(result bool, status *gateshv1alpha1.GateStatus) {
	var message string
	var reason string
	var openedCondition metav1.ConditionStatus
	var closedCondition metav1.ConditionStatus

	if result {
		status.State = gateshv1alpha1.GateStateOpened
		message = "Gate was evaluated to true"
		reason = "GateConditionMet"
		openedCondition = metav1.ConditionTrue
		closedCondition = metav1.ConditionFalse
	} else {
		status.State = gateshv1alpha1.GateStateClosed
		message = "Gate was evaluated to false"
		reason = "GateConditionNotMet"
		openedCondition = metav1.ConditionFalse
		closedCondition = metav1.ConditionTrue
	}

	// Opened conditions
	meta.SetStatusCondition(&status.Conditions, metav1.Condition{Type: gateshv1alpha1.GateStateOpened, Status: openedCondition, Reason: reason, Message: message})
	meta.SetStatusCondition(&status.Conditions, metav1.Condition{Type: "Available", Status: openedCondition, Reason: reason, Message: message})

	// Closed conditions
	meta.SetStatusCondition(&status.Conditions, metav1.Condition{Type: gateshv1alpha1.GateStateClosed, Status: closedCondition, Reason: reason, Message: message})
	meta.SetStatusCondition(&status.Conditions, metav1.Condition{Type: "Progressing", Status: closedCondition, Reason: reason, Message: message})
}

// SetupWithManager sets up the controller with the Manager.
func (r *GateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gateshv1alpha1.Gate{}).
		Named("gate").
		Complete(r)
}

func (r *GateReconciler) EvaluateGateSpec(ctx context.Context, gate *gateshv1alpha1.Gate) (bool, error) {
	return false, nil
}

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

	expreval "github.com/robinlioret/gate-operator/internal/gate"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
)

const GateRequeueCooldownSeconds = 1

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
	if err := r.Get(ctx, req.NamespacedName, &gate); err != nil {
		log.Error(err, "unable to fetch Gate")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Do not evaluate if it was updated too recently
	if gate.Status.LastEvaluation != nil && gate.Status.LastEvaluation.Add(GateRequeueCooldownSeconds*time.Second).After(time.Now()) {
		log.V(1).Info("Gate was already processed recently")
		return ctrl.Result{}, nil
	}

	// Evaluate the expression
	evaluator := expreval.NewExpressionEvaluator(ctx, r.Client, gate.Namespace)
	result, err := evaluator.Evaluate(gate.Spec)
	if err != nil {
		log.Error(err, "failed to evaluate Gate")
		return ctrl.Result{}, err
	}

	// Process the result and update the gate's status
	if result {
		log.Info("Gate is evaluated to true")
		meta.SetStatusCondition(&gate.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  "True",
			Reason:  "GateEvaluatedTrue",
			Message: "Gate was evaluated to true",
		})
		meta.RemoveStatusCondition(&gate.Status.Conditions, "Progressing")
		gate.Status.State = "Ready"
	} else {
		log.Info("Gate is evaluated to false")
		meta.SetStatusCondition(&gate.Status.Conditions, metav1.Condition{
			Type:    "Progressing",
			Status:  "True",
			Reason:  "GateEvaluatedTrue",
			Message: "Gate was evaluated to true",
		})
		meta.RemoveStatusCondition(&gate.Status.Conditions, "Ready")
		gate.Status.State = "Progressing"
	}

	gate.Status.LastEvaluation = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, &gate); err != nil {
		log.Error(err, "failed to update Gate")
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: r.GetRequeueAfter(gate)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gateshv1alpha1.Gate{}).
		Named("gate").
		Complete(r)
}

func (r *GateReconciler) GetRequeueAfter(gate gateshv1alpha1.Gate) time.Duration {
	if gate.Spec.RequeueAfter != nil {
		return gate.Spec.RequeueAfter.Duration
	} else {
		return time.Second * 10
	}
}

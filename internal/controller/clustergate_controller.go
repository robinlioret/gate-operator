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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
)

// ClusterGateReconciler reconciles a ClusterGate object
type ClusterGateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=gate.sh,resources=clustergates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gate.sh,resources=clustergates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gate.sh,resources=clustergates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterGate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/reconcile
func (r *ClusterGateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Get the object by reference
	var gate gateshv1alpha1.ClusterGate
	err := r.Get(ctx, req.NamespacedName, &gate)
	if errors.IsNotFound(err) {
		log.Info("ClusterGate not found")
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "unable to fetch ClusterGate")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Do not evaluate if it was updated too recently
	// if gate.Status.NextEvaluation.After(time.Now()) {
	//	log.V(1).Info("Gate was already processed recently")
	//	return ctrl.Result{RequeueAfter: gate.Status.NextEvaluation.Sub(time.Now())}, nil
	// }

	gateReconciler := GateReconciler{Client: r.Client, Scheme: r.Scheme}
	result, targetConditions := gateReconciler.EvaluateGateSpec(ctx, gate.Namespace, gate.Spec)
	gateReconciler.UpdateGateStatusFromResult(result, targetConditions, &gate.Status)

	gate.Status.NextEvaluation = metav1.Time{Time: time.Now().Add(gate.Spec.RequeueAfter.Duration)}
	if err := r.Status().Update(ctx, &gate); err != nil {
		log.Error(err, "unable to update ClusterGate")
		return ctrl.Result{RequeueAfter: gate.Spec.RequeueAfter.Duration}, err
	}
	log.V(1).Info("ClusterGate processed successfully")
	return ctrl.Result{RequeueAfter: gate.Spec.RequeueAfter.Duration}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterGateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gateshv1alpha1.ClusterGate{}).
		Named("clustergate").
		Complete(r)
}

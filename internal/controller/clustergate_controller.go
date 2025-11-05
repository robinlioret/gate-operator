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

	gateObject := gateshv1alpha1.Gate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterGate",
			APIVersion: gateshv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: gate.ObjectMeta,
		Spec:       gate.Spec,
		Status:     gate.Status,
	}
	gcr := GateCommonReconciler{
		Context: ctx,
		Client:  r.Client,
		Gate:    &gateObject,
	}
	err = gcr.Reconcile()
	gate.Status = gateObject.Status
	if err != nil {
		log.Error(err, "unable to reconcile ClusterGate")
		return ctrl.Result{RequeueAfter: gate.Spec.RequeueAfter.Duration}, err
	}
	if err := r.Status().Update(ctx, &gate); err != nil {
		log.Error(err, "unable to update ClusterGate")
	}
	return ctrl.Result{RequeueAfter: gate.Spec.RequeueAfter.Duration}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterGateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gateshv1alpha1.ClusterGate{}).
		Named("clustergate").
		Complete(r)
}

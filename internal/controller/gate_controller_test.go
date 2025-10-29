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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
)

var testResources = []gateshv1alpha1.Gate{
	// Simpliest gate
	{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gateshv1alpha1.GroupVersion.String(),
			Kind:       "Gate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gate-simpliest",
			Namespace: "default",
		},
		Spec: gateshv1alpha1.GateSpec{
			Targets: []gateshv1alpha1.GateTarget{
				{
					Kind:       "Deployment",
					ApiVersion: "apps/v1",
					Name:       "coredns",
					Namespace:  "kube-system",
					ExistsOnly: true,
				},
			},
		},
	},
}

var _ = Describe("Gate Controller", func() {
	for _, resource := range testResources {
		Context(fmt.Sprintf("When reconciling a resource: %s", resource.Name), func() {
			ctx := context.Background()

			typeNamespacedName := types.NamespacedName{
				Name:      resource.Name,
				Namespace: resource.Namespace,
			}
			gate := &gateshv1alpha1.Gate{}

			BeforeEach(func() {
				By("Creating the custom resource for the Kind Gate")
				err := k8sClient.Get(ctx, typeNamespacedName, gate)
				if err != nil && errors.IsNotFound(err) {
					Expect(k8sClient.Create(ctx, &resource)).To(Succeed())
				}
			})

			AfterEach(func() {
				err := k8sClient.Get(ctx, typeNamespacedName, &resource)
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance Gate")
				Expect(k8sClient.Delete(ctx, &resource)).To(Succeed())
			})

			It("should successfully reconcile the resource", func() {
				By("Reconciling the created resource")
				controllerReconciler := &GateReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
				// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
				// Example: If you expect a certain status condition after reconciliation, verify it here.

				By("Getting the reconciled resource")
				err = k8sClient.Get(ctx, typeNamespacedName, gate)
				Expect(err).NotTo(HaveOccurred())
				AddReportEntry("gate", gate)

				By("Having set the status next evaluation field")
				Expect(gate.Status.NextEvaluation.Time.After(time.Now())).To(BeTrue())

				By("Having set the conditions field")
				Expect(len(gate.Status.Conditions)).NotTo(Equal(0))

				By("Having set the status state field")
				Expect(gate.Status.State).NotTo(Equal(""))

				By("Opening the gate")
				Expect(gate.Status.State).NotTo(Equal(gateshv1alpha1.GateStateOpened))

				oldNextEvaluation := gate.Status.NextEvaluation
				By("Reconciling the created resource again (too soon)")
				_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("Getting the reconciled resource")
				err = k8sClient.Get(ctx, typeNamespacedName, gate)
				Expect(err).NotTo(HaveOccurred())

				By("Having kept the status next evaluation field")
				Expect(gate.Status.NextEvaluation).To(Equal(oldNextEvaluation))
			})
		})
	}
})

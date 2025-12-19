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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	schemeBuilder "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("ClusterGate Controller", func() {
	scheme := runtime.NewScheme()
	_ = schemeBuilder.AddToScheme(scheme)
	_ = gateshv1alpha1.AddToScheme(scheme) // Add CRDs

	clustergate1 := &gateshv1alpha1.ClusterGate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "clustergate1",
			Namespace: "default",
		},
		Spec: gateshv1alpha1.GateSpec{
			Targets: []gateshv1alpha1.GateTarget{
				{
					Name: "Target1",
					Selector: gateshv1alpha1.GateTargetSelector{
						ApiVersion: "v1",
						Kind:       "ConfigMap",
						Namespace:  "default",
						Name:       "cm1",
					},
					Validators: []gateshv1alpha1.GateTargetValidator{
						{
							AtLeast: gateshv1alpha1.GateTargetValidatorAtLeast{Count: 1},
						},
					},
				},
				{
					Name: "Target2",
					Selector: gateshv1alpha1.GateTargetSelector{
						ApiVersion: "apps/v1",
						Kind:       "Deployment",
						Namespace:  "default",
						LabelSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test1",
							},
						},
					},
					Validators: []gateshv1alpha1.GateTargetValidator{
						{
							MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
								Type:   "Available",
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			Operation:    gateshv1alpha1.GateOperation{Operator: gateshv1alpha1.GateOperatorAnd},
			RequeueAfter: &metav1.Duration{Duration: 10 * time.Second},
		},
	}
	cm1 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm1",
			Namespace: "default",
		},
	}
	deploy1 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy1",
			Namespace: "default",
			Labels:    map[string]string{"app": "test1"},
		},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 1,
			Conditions:         []appsv1.DeploymentCondition{},
		},
	}
	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(clustergate1, cm1, deploy1).
		Build()

	Context("When reconciling a non-existing resource", func() {
		typeNamespacedName := types.NamespacedName{
			Name:      "non-existing-resource",
			Namespace: "default",
		}
		It("Should not fail while reconcile the non-existing resource", func() {
			By("Reconciling the non-existing resource")
			controllerReconciler := &ClusterGateReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When reconciling a valid resource", func() {
		typeNamespacedName := types.NamespacedName{
			Name:      "clustergate1",
			Namespace: "default",
		}
		It("Should not fail while reconcile the non-existing resource", func() {
			By("Reconciling the non-existing resource")
			reconciler := &ClusterGateReconciler{
				Client: client,
				Scheme: client.Scheme(),
			}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

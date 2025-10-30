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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
)

type TestGate struct {
	Gate           gateshv1alpha1.Gate
	ExpectedStatus gateshv1alpha1.GateStatus
}

// var gateOpenedExpectedConditions = []metav1.Condition{
//	{
//		Type:    "Opened",
//		Status:  "True",
//		Reason:  "GateConditionMet",
//		Message: "Gate was evaluated to true",
//	},
//	{
//		Type:    "Closed",
//		Status:  "False",
//		Reason:  "GateConditionMet",
//		Message: "Gate was evaluated to true",
//	},
//	{
//		Type:    "Available",
//		Status:  "True",
//		Reason:  "GateConditionMet",
//		Message: "Gate was evaluated to true",
//	},
//	{
//		Type:    "Progressing",
//		Status:  "False",
//		Reason:  "GateConditionMet",
//		Message: "Gate was evaluated to true",
//	},
// }

var gateClosedExpectedConditions = []metav1.Condition{
	{
		Type:    "Opened",
		Status:  "False",
		Reason:  "GateConditionNotMet",
		Message: "Gate was evaluated to false",
	},
	{
		Type:    "Closed",
		Status:  "True",
		Reason:  "GateConditionNotMet",
		Message: "Gate was evaluated to false",
	},
	{
		Type:    "Available",
		Status:  "False",
		Reason:  "GateConditionNotMet",
		Message: "Gate was evaluated to false",
	},
	{
		Type:    "Progressing",
		Status:  "True",
		Reason:  "GateConditionNotMet",
		Message: "Gate was evaluated to false",
	},
}

var testResources = []TestGate{
	// Simplest opened gate
	// {
	//	Gate: gateshv1alpha1.Gate{
	//		TypeMeta: metav1.TypeMeta{
	//			APIVersion: gateshv1alpha1.GroupVersion.String(),
	//			Kind:       "Gate",
	//		},
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name:      "test-simplest-opened",
	//			Namespace: "default",
	//		},
	//		Spec: gateshv1alpha1.GateSpec{
	//			Targets: []gateshv1alpha1.GateTarget{
	//				{
	//					TargetName: "CoreDns",
	//					Kind:       "Deployment",
	//					ApiVersion: "apps/v1",
	//					Name:       "coredns",
	//					Namespace:  "kube-system",
	//					ExistsOnly: true,
	//				},
	//			},
	//		},
	//	},
	//	ExpectedStatus: gateshv1alpha1.GateStatus{
	//		Conditions: gateOpenedExpectedConditions,
	//		State:      gateshv1alpha1.GateStateOpened,
	//		TargetConditions: []metav1.Condition{
	//			{
	//				Type:    "CoreDns",
	//				Status:  "True",
	//				Reason:  "TargetConditionMet",
	//				Message: "object found",
	//			},
	//		},
	//	},
	// },

	// Simplest closed gate
	{
		Gate: gateshv1alpha1.Gate{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gateshv1alpha1.GroupVersion.String(),
				Kind:       "Gate",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-simplest-closed",
				Namespace: "default",
			},
			Spec: gateshv1alpha1.GateSpec{
				Targets: []gateshv1alpha1.GateTarget{
					{
						TargetName: "NotFound",
						Kind:       "Deployment",
						ApiVersion: "apps/v1",
						Name:       "not-found",
						Namespace:  "default",
						ExistsOnly: true,
					},
				},
			},
		},
		ExpectedStatus: gateshv1alpha1.GateStatus{
			Conditions: gateClosedExpectedConditions,
			State:      gateshv1alpha1.GateStateClosed,
			TargetConditions: []metav1.Condition{
				{
					Type:    "NotFound",
					Status:  "False",
					Reason:  "TargetConditionNotMet",
					Message: "object not found",
				},
			},
		},
	},

	// Mono target opened gate
	// {
	//	Gate: gateshv1alpha1.Gate{
	//		TypeMeta: metav1.TypeMeta{
	//			APIVersion: gateshv1alpha1.GroupVersion.String(),
	//			Kind:       "Gate",
	//		},
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name:      "test-mono-target-opened",
	//			Namespace: "default",
	//		},
	//		Spec: gateshv1alpha1.GateSpec{
	//			Targets: []gateshv1alpha1.GateTarget{
	//				{
	//					TargetName: "CoreDns",
	//					Kind:       "Deployment",
	//					ApiVersion: "apps/v1",
	//					Name:       "coredns",
	//					Namespace:  "kube-system",
	//					ExistsOnly: false,
	//					DesiredCondition: gateshv1alpha1.GateTargetCondition{
	//						Type:   "Available",
	//						Status: "True",
	//					},
	//				},
	//			},
	//		},
	//	},
	//	ExpectedStatus: gateshv1alpha1.GateStatus{
	//		Conditions: gateOpenedExpectedConditions,
	//		State:      gateshv1alpha1.GateStateOpened,
	//		TargetConditions: []metav1.Condition{
	//			{
	//				Type:    "CoreDns",
	//				Status:  "True",
	//				Reason:  "TargetConditionMet",
	//				Message: "object found,desired condition met",
	//			},
	//		},
	//	},
	// },
	//
	// Mono target closed gate
	// {
	//	Gate: gateshv1alpha1.Gate{
	//		TypeMeta: metav1.TypeMeta{
	//			APIVersion: gateshv1alpha1.GroupVersion.String(),
	//			Kind:       "Gate",
	//		},
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name:      "test-mono-target-closed",
	//			Namespace: "default",
	//		},
	//		Spec: gateshv1alpha1.GateSpec{
	//			Targets: []gateshv1alpha1.GateTarget{
	//				{
	//					TargetName: "NotFound",
	//					Kind:       "Deployment",
	//					ApiVersion: "apps/v1",
	//					Name:       "not-found",
	//					Namespace:  "default",
	//					ExistsOnly: false,
	//					DesiredCondition: gateshv1alpha1.GateTargetCondition{
	//						Type:   "Available",
	//						Status: "True",
	//					},
	//				},
	//			},
	//		},
	//	},
	//	ExpectedStatus: gateshv1alpha1.GateStatus{
	//		Conditions: gateClosedExpectedConditions,
	//		State:      gateshv1alpha1.GateStateClosed,
	//		TargetConditions: []metav1.Condition{
	//			{
	//				Type:    "NotFound",
	//				Status:  "False",
	//				Reason:  "TargetConditionNotMet",
	//				Message: "object not found",
	//			},
	//		},
	//	},
	// },
	//
	// Multi targets and opened gate
	// {
	//	Gate: gateshv1alpha1.Gate{
	//		TypeMeta: metav1.TypeMeta{
	//			APIVersion: gateshv1alpha1.GroupVersion.String(),
	//			Kind:       "Gate",
	//		},
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name:      "test-mono-target-opened",
	//			Namespace: "default",
	//		},
	//		Spec: gateshv1alpha1.GateSpec{
	//			Targets: []gateshv1alpha1.GateTarget{
	//				{
	//					TargetName: "CoreDns1",
	//					Kind:       "Deployment",
	//					ApiVersion: "apps/v1",
	//					Name:       "coredns",
	//					Namespace:  "kube-system",
	//					ExistsOnly: false,
	//					DesiredCondition: gateshv1alpha1.GateTargetCondition{
	//						Type:   "Available",
	//						Status: "True",
	//					},
	//				},
	//				{
	//					TargetName: "CoreDns2",
	//					Kind:       "Deployment",
	//					ApiVersion: "apps/v1",
	//					Name:       "coredns",
	//					Namespace:  "kube-system",
	//					ExistsOnly: false,
	//					DesiredCondition: gateshv1alpha1.GateTargetCondition{
	//						Type:   "Available",
	//						Status: "True",
	//					},
	//				},
	//			},
	//			Operation: gateshv1alpha1.GateOperation{
	//				Operator: gateshv1alpha1.GateOperatorAnd,
	//			},
	//		},
	//	},
	//	ExpectedStatus: gateshv1alpha1.GateStatus{
	//		Conditions: gateOpenedExpectedConditions,
	//		State:      gateshv1alpha1.GateStateOpened,
	//		TargetConditions: []metav1.Condition{
	//			{
	//				Type:    "CoreDns1",
	//				Status:  "True",
	//				Reason:  "TargetConditionMet",
	//				Message: "object found,desired condition met",
	//			},
	//			{
	//				Type:    "CoreDns2",
	//				Status:  "True",
	//				Reason:  "TargetConditionMet",
	//				Message: "object found,desired condition met",
	//			},
	//		},
	//	},
	// },
	//
	// Multi targets and closed gate
	// {
	//	Gate: gateshv1alpha1.Gate{
	//		TypeMeta: metav1.TypeMeta{
	//			APIVersion: gateshv1alpha1.GroupVersion.String(),
	//			Kind:       "Gate",
	//		},
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name:      "test-mono-target-closed",
	//			Namespace: "default",
	//		},
	//		Spec: gateshv1alpha1.GateSpec{
	//			Targets: []gateshv1alpha1.GateTarget{
	//				{
	//					TargetName: "CoreDns",
	//					Kind:       "Deployment",
	//					ApiVersion: "apps/v1",
	//					Name:       "coredns",
	//					Namespace:  "kube-system",
	//					ExistsOnly: false,
	//					DesiredCondition: gateshv1alpha1.GateTargetCondition{
	//						Type:   "Available",
	//						Status: "True",
	//					},
	//				},
	//				{
	//					TargetName: "NotFound",
	//					Kind:       "Deployment",
	//					ApiVersion: "apps/v1",
	//					Name:       "not-found",
	//					Namespace:  "default",
	//					ExistsOnly: false,
	//					DesiredCondition: gateshv1alpha1.GateTargetCondition{
	//						Type:   "Available",
	//						Status: "True",
	//					},
	//				},
	//			},
	//			Operation: gateshv1alpha1.GateOperation{
	//				Operator: gateshv1alpha1.GateOperatorAnd,
	//			},
	//		},
	//	},
	//	ExpectedStatus: gateshv1alpha1.GateStatus{
	//		Conditions: gateOpenedExpectedConditions,
	//		State:      gateshv1alpha1.GateStateOpened,
	//		TargetConditions: []metav1.Condition{
	//			{
	//				Type:    "CoreDns",
	//				Status:  "True",
	//				Reason:  "TargetConditionMet",
	//				Message: "object found,desired condition met",
	//			},
	//			{
	//				Type:    "NotFound",
	//				Status:  "False",
	//				Reason:  "TargetConditionNotMet",
	//				Message: "object not found",
	//			},
	//		},
	//	},
	// },
}

var _ = Describe("Gate Controller", func() {
	for _, resource := range testResources {
		Context(fmt.Sprintf("When reconciling a resource: %s", resource.Gate.Name), func() {
			ctx := context.Background()

			typeNamespacedName := types.NamespacedName{
				Name:      resource.Gate.Name,
				Namespace: resource.Gate.Namespace,
			}
			gate := &gateshv1alpha1.Gate{}

			BeforeEach(func() {
				By("Creating the custom resource for the Kind Gate")
				err := k8sClient.Get(ctx, typeNamespacedName, gate)
				if err != nil && errors.IsNotFound(err) {
					Expect(k8sClient.Create(ctx, &resource.Gate)).To(Succeed())
				}
			})

			AfterEach(func() {
				err := k8sClient.Get(ctx, typeNamespacedName, &resource.Gate)
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance Gate")
				Expect(k8sClient.Delete(ctx, &resource.Gate)).To(Succeed())
			})

			It("Should successfully reconcile the resource", func() {
				AddReportEntry("test gate name", resource.Gate.Name)

				By("Reconciling the created resource")
				controllerReconciler := &GateReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
				// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
				// Example: If you expect a certain status condition after reconciliation, verify it here.

				By("Getting the reconciled resource")
				err = k8sClient.Get(ctx, typeNamespacedName, gate)
				Expect(err).NotTo(HaveOccurred())

				By("Having set the status next evaluation field")
				Expect(gate.Status.NextEvaluation.Time.After(time.Now())).To(BeTrue())

				By("Having set the target conditions field")
				Expect(gate.Status.TargetConditions).NotTo(BeEmpty())
				Expect(gate.Status.TargetConditions).To(HaveLen(len(resource.ExpectedStatus.TargetConditions)))
				for _, desiredCondition := range resource.ExpectedStatus.TargetConditions {
					condition := meta.FindStatusCondition(gate.Status.TargetConditions, desiredCondition.Type)
					By(fmt.Sprintf("Having updated the condition '%s' correctly", condition.Type))
					AddReportEntry(fmt.Sprintf("TargetCondition: %s", condition.Type), condition)
					Expect(condition).ToNot(BeNil())
					Expect(condition.Status).To(Equal(desiredCondition.Status))
					Expect(condition.Reason).To(Equal(desiredCondition.Reason))
					Expect(condition.Message).To(Equal(desiredCondition.Message))
				}

				By("Having set the conditions field")
				Expect(gate.Status.Conditions).NotTo(BeEmpty())
				Expect(gate.Status.Conditions).To(HaveLen(len(resource.ExpectedStatus.Conditions)))
				for _, desiredCondition := range resource.ExpectedStatus.Conditions {
					condition := meta.FindStatusCondition(gate.Status.Conditions, desiredCondition.Type)
					By(fmt.Sprintf("Having updated the condition '%s' correctly", condition.Type))
					AddReportEntry(fmt.Sprintf("GateCondition: %s", condition.Type), condition)
					Expect(condition).ToNot(BeNil())
					Expect(condition.Status).To(Equal(desiredCondition.Status))
					Expect(condition.Reason).To(Equal(desiredCondition.Reason))
					Expect(condition.Message).To(Equal(desiredCondition.Message))
				}

				By("Having set the status state field")
				Expect(gate.Status.State).NotTo(Equal(""))
				Expect(gate.Status.State).To(Equal(resource.ExpectedStatus.State))

				By("Having the right status")
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

	Context("When reconciling a non-existing resource", func() {
		typeNamespacedName := types.NamespacedName{
			Name:      "non-existing-resource",
			Namespace: "default",
		}
		It("Should not fail while reconcile the non-existing resource", func() {
			By("Reconciling the non-existing resource")
			controllerReconciler := &GateReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

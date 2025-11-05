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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	schemeBuilder "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Gate Common Reconciler", func() {
	// ================================================
	// INITIALIZE FAKE CLIENT
	// ------------------------------------------------
	scheme := runtime.NewScheme()
	_ = schemeBuilder.AddToScheme(scheme)
	_ = gateshv1alpha1.AddToScheme(scheme) // Add CRDs

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
		},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 1,
			Conditions:         []appsv1.DeploymentCondition{},
		},
	}
	deploy2 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy2",
			Namespace: "default",
		},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 1,
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
	deploy3 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy3",
			Namespace: "default",
		},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 1,
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}
	deploy4 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy4",
			Namespace: "default",
		},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 1,
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentProgressing,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(cm1, deploy1, deploy2, deploy3, deploy4).
		Build()

	// ================================================
	// UNIT TESTS
	// ------------------------------------------------
	Context("Test GetObjectName", func() {
		gcr := &GateCommonReconciler{
			Client:  client,
			Context: context.Background(),
			Gate:    &gateshv1alpha1.Gate{},
		}
		obj := unstructured.Unstructured{
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Dummy",
				"metadata": map[string]interface{}{
					"namespace": "some-namespace",
					"name":      "a-name",
				},
				"spec": map[string]interface{}{},
			},
		}

		It("Should return the object name formatted", func() {
			By("Calling the function GetObjectName")
			result := gcr.GetObjectName(obj)
			Expect(result).To(Equal("some-namespace/a-name"))
		})
	})

	Context("Test ComputeOperation with the AND operator", func() {
		gcr := &GateCommonReconciler{
			Client:  client,
			Context: context.Background(),
			Gate: &gateshv1alpha1.Gate{
				Spec: gateshv1alpha1.GateSpec{
					Operation: gateshv1alpha1.GateOperation{Operator: gateshv1alpha1.GateOperatorAnd},
				},
			},
		}

		It("Should evaluate a valid AND condition true", func() {
			By("Creating a fake Gate Common reconciler")
			By("Creating a set of target conditions")
			targetConditions := []metav1.Condition{
				metav1.Condition{Type: "target1", Status: metav1.ConditionTrue, Reason: "", Message: ""},
				metav1.Condition{Type: "target2", Status: metav1.ConditionTrue, Reason: "", Message: ""},
				metav1.Condition{Type: "target3", Status: metav1.ConditionTrue, Reason: "", Message: ""},
			}
			By("Calling the function ComputeOperation")
			result := gcr.ComputeOperation(targetConditions)
			Expect(result).To(BeTrue())
		})

		It("Should evaluate an invalid AND condition to false", func() {
			By("Creating a set of target conditions")
			targetConditions := []metav1.Condition{
				metav1.Condition{Type: "target1", Status: metav1.ConditionTrue, Reason: "", Message: ""},
				metav1.Condition{Type: "target2", Status: metav1.ConditionFalse, Reason: "", Message: ""},
				metav1.Condition{Type: "target3", Status: metav1.ConditionTrue, Reason: "", Message: ""},
			}
			By("Calling the function ComputeOperation")
			result := gcr.ComputeOperation(targetConditions)
			Expect(result).To(BeFalse())
		})
	})

	Context("Test ComputeOperation with the OR operator", func() {
		gcr := &GateCommonReconciler{
			Client:  client,
			Context: context.Background(),
			Gate: &gateshv1alpha1.Gate{
				Spec: gateshv1alpha1.GateSpec{
					Operation: gateshv1alpha1.GateOperation{Operator: gateshv1alpha1.GateOperatorOr},
				},
			},
		}

		It("Should evaluate a valid OR condition to true", func() {
			By("Creating a set of target conditions")
			targetConditions := []metav1.Condition{
				metav1.Condition{Type: "target1", Status: metav1.ConditionFalse, Reason: "", Message: ""},
				metav1.Condition{Type: "target2", Status: metav1.ConditionFalse, Reason: "", Message: ""},
				metav1.Condition{Type: "target3", Status: metav1.ConditionTrue, Reason: "", Message: ""},
			}
			By("Calling the function ComputeOperation")
			result := gcr.ComputeOperation(targetConditions)
			Expect(result).To(BeTrue())
		})

		It("Should evaluate an invalid OR condition to false", func() {
			By("Creating a set of target conditions")
			targetConditions := []metav1.Condition{
				metav1.Condition{Type: "target1", Status: metav1.ConditionFalse, Reason: "", Message: ""},
				metav1.Condition{Type: "target2", Status: metav1.ConditionFalse, Reason: "", Message: ""},
				metav1.Condition{Type: "target3", Status: metav1.ConditionFalse, Reason: "", Message: ""},
			}
			By("Calling the function ComputeOperation")
			result := gcr.ComputeOperation(targetConditions)
			Expect(result).To(BeFalse())
		})
	})

	Context("Test UpdateGateStatusFromResult", func() {
		gcr := &GateCommonReconciler{
			Client:  client,
			Context: context.Background(),
			Gate: &gateshv1alpha1.Gate{
				Spec: gateshv1alpha1.GateSpec{
					Operation: gateshv1alpha1.GateOperation{Operator: gateshv1alpha1.GateOperatorOr},
				},
			},
		}

		It("Should successfully update the gate status", func() {
			targetConditions := []metav1.Condition{
				metav1.Condition{Type: "target1", Status: metav1.ConditionTrue, Reason: "", Message: ""},
				metav1.Condition{Type: "target2", Status: metav1.ConditionTrue, Reason: "", Message: ""},
			}

			By("Updating the gate status when result is true")
			gcr.UpdateGateStatusFromResult(true, targetConditions)
			Expect(gcr.Gate.Status.State).To(Equal(gateshv1alpha1.GateStateOpened))
			Expect(gcr.Gate.Status.Conditions).To(HaveLen(2))
			Expect(gcr.Gate.Status.TargetConditions).To(HaveLen(2))
			Expect(meta.FindStatusCondition(gcr.Gate.Status.Conditions, "Opened").Status).To(Equal(metav1.ConditionTrue))
			Expect(meta.FindStatusCondition(gcr.Gate.Status.Conditions, "Closed").Status).To(Equal(metav1.ConditionFalse))

			By("Updating the gate status when result is false")
			gcr.UpdateGateStatusFromResult(false, targetConditions)
			Expect(gcr.Gate.Status.State).To(Equal(gateshv1alpha1.GateStateClosed))
			Expect(gcr.Gate.Status.Conditions).To(HaveLen(2))
			Expect(gcr.Gate.Status.TargetConditions).To(HaveLen(2))
			Expect(meta.FindStatusCondition(gcr.Gate.Status.Conditions, "Opened").Status).To(Equal(metav1.ConditionFalse))
			Expect(meta.FindStatusCondition(gcr.Gate.Status.Conditions, "Closed").Status).To(Equal(metav1.ConditionTrue))
		})
	})

	Context("Test EvaluateTarget", func() {
		gcr := &GateCommonReconciler{
			Client:  client,
			Context: context.Background(),
			Gate:    &gateshv1alpha1.Gate{},
		}

		It("Should evaluate a valid target object with exist only set to true", func() {
			By("Calling the function EvaluateTarget")
			target := gateshv1alpha1.GateTarget{
				TargetName: "Target1",
				Kind:       "ConfigMap",
				ApiVersion: "v1",
				Namespace:  "default",
				Name:       "cm1",
				ExistsOnly: true,
			}
			condition := gcr.EvaluateTarget(&target)
			AddReportEntry("Condition", condition)
			Expect(condition.Type).To(Equal("Target1"))
			Expect(condition.Status).To(Equal(metav1.ConditionTrue))
			Expect(condition.Reason).To(Equal(ReasonOkObjectsFound))
		})

		It("Should evaluate a invalid target object with exist only set to true", func() {
			By("Calling the function EvaluateTarget")
			target := gateshv1alpha1.GateTarget{
				TargetName: "Target1",
				Kind:       "ConfigMap",
				ApiVersion: "v1",
				Namespace:  "default",
				Name:       "not-found",
				ExistsOnly: true,
			}
			condition := gcr.EvaluateTarget(&target)
			AddReportEntry("Condition", condition)
			Expect(condition.Type).To(Equal("Target1"))
			Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			Expect(condition.Reason).To(Equal(ReasonKoNoObjectsFound))
		})

		It("Should evaluate to false a valid target with no conditions", func() {
			By("Calling the function EvaluateTarget")
			target := gateshv1alpha1.GateTarget{
				TargetName: "Target1",
				Kind:       "Deployment",
				ApiVersion: "apps/v1",
				Namespace:  "default",
				Name:       "deploy1",
				ExistsOnly: false,
				DesiredCondition: gateshv1alpha1.GateTargetCondition{
					Type:   "Available",
					Status: "True",
				},
			}
			condition := gcr.EvaluateTarget(&target)
			AddReportEntry("Condition", condition)
			Expect(condition.Type).To(Equal("Target1"))
			Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			Expect(condition.Reason).To(Equal(ReasonKo))
		})

		It("Should evaluate to true a valid target with matching condition", func() {
			By("Calling the function EvaluateTarget")
			target := gateshv1alpha1.GateTarget{
				TargetName: "Target1",
				Kind:       "Deployment",
				ApiVersion: "apps/v1",
				Namespace:  "default",
				Name:       "deploy2",
				ExistsOnly: false,
				DesiredCondition: gateshv1alpha1.GateTargetCondition{
					Type:   "Available",
					Status: "True",
				},
			}
			condition := gcr.EvaluateTarget(&target)
			AddReportEntry("Condition", condition)
			Expect(condition.Type).To(Equal("Target1"))
			Expect(condition.Status).To(Equal(metav1.ConditionTrue))
			Expect(condition.Reason).To(Equal(ReasonOk))
		})

		It("Should evaluate to true a valid target with condition no matching", func() {
			By("Calling the function EvaluateTarget")
			target := gateshv1alpha1.GateTarget{
				TargetName: "Target1",
				Kind:       "Deployment",
				ApiVersion: "apps/v1",
				Namespace:  "default",
				Name:       "deploy3",
				ExistsOnly: false,
				DesiredCondition: gateshv1alpha1.GateTargetCondition{
					Type:   "Available",
					Status: "True",
				},
			}
			condition := gcr.EvaluateTarget(&target)
			AddReportEntry("Condition", condition)
			Expect(condition.Type).To(Equal("Target1"))
			Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			Expect(condition.Reason).To(Equal(ReasonKo))
		})

		It("Should evaluate to true a valid target with no matching condition", func() {
			By("Calling the function EvaluateTarget")
			target := gateshv1alpha1.GateTarget{
				TargetName: "Target1",
				Kind:       "Deployment",
				ApiVersion: "apps/v1",
				Namespace:  "default",
				Name:       "deploy4",
				ExistsOnly: false,
				DesiredCondition: gateshv1alpha1.GateTargetCondition{
					Type:   "Available",
					Status: "True",
				},
			}
			condition := gcr.EvaluateTarget(&target)
			AddReportEntry("Condition", condition)
			Expect(condition.Type).To(Equal("Target1"))
			Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			Expect(condition.Reason).To(Equal(ReasonKo))
		})
	})

	Context("Test EvaluateSpec", func() {
		gcr := &GateCommonReconciler{
			Client:  client,
			Context: context.Background(),
			Gate: &gateshv1alpha1.Gate{
				Spec: gateshv1alpha1.GateSpec{
					Targets: []gateshv1alpha1.GateTarget{
						{
							TargetName: "Target1",
							Kind:       "Deployment",
							ApiVersion: "apps/v1",
							Namespace:  "default",
							Name:       "deploy2",
							ExistsOnly: false,
							DesiredCondition: gateshv1alpha1.GateTargetCondition{
								Type:   "Available",
								Status: "True",
							},
						},
						{
							TargetName: "Target2",
							Kind:       "ConfigMap",
							ApiVersion: "v1",
							Namespace:  "default",
							Name:       "cm1",
							ExistsOnly: true,
						},
					},
				},
			},
		}

		It("Should evaluate a valid target", func() {
			result, conditions := gcr.EvaluateSpec()
			AddReportEntry("Condition", conditions)
			Expect(result).To(Equal(true))
			Expect(conditions).To(HaveLen(2))
		})
	})
})

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	schemeBuilder "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("GateCommonReconciler", func() {
	var ctx context.Context
	var scheme *runtime.Scheme

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(schemeBuilder.AddToScheme(scheme)).To(Succeed())
		Expect(gateshv1alpha1.AddToScheme(scheme)).To(Succeed())
	})

	Describe("Reconcile", func() {
		It("should open the gate for a single target by name with matching condition", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "target-pod",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								Name:       "target-pod",
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			}

			pod := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "target-pod",
						"namespace": "default",
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate, pod).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateOpened))
			Expect(meta.FindStatusCondition(gate.Status.Conditions, gateshv1alpha1.GateStateOpened).Status).To(Equal(metav1.ConditionTrue))
			Expect(meta.FindStatusCondition(gate.Status.Conditions, gateshv1alpha1.GateStateClosed).Status).To(Equal(metav1.ConditionFalse))
			Expect(gate.Status.TargetConditions).To(HaveLen(1))
			Expect(gate.Status.TargetConditions[0].Status).To(Equal(metav1.ConditionTrue))
			Expect(gate.Status.TargetConditions[0].Type).To(Equal("target-pod"))
		})

		It("should close the gate for a single target by name with non-matching condition", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "target-pod",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								Name:       "target-pod",
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			}

			pod := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "target-pod",
						"namespace": "default",
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
							},
						},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate, pod).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateClosed))
			Expect(meta.FindStatusCondition(gate.Status.Conditions, gateshv1alpha1.GateStateOpened).Status).To(Equal(metav1.ConditionFalse))
			Expect(meta.FindStatusCondition(gate.Status.Conditions, gateshv1alpha1.GateStateClosed).Status).To(Equal(metav1.ConditionTrue))
			Expect(gate.Status.TargetConditions).To(HaveLen(1))
			Expect(gate.Status.TargetConditions[0].Status).To(Equal(metav1.ConditionFalse))
			Expect(gate.Status.TargetConditions[0].Message).To(ContainSubstring("condition Ready is wrong (expected True, got False)"))
		})

		It("should open the gate for a single target by name with matching jsonpointer", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "target-pod",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								Name:       "target-pod",
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									JsonPointer: gateshv1alpha1.GateTargetValidatorJsonPointer{
										JsonPointer: "/metadata/name",
										Value:       "target-pod",
									},
								},
							},
						},
					},
				},
			}

			pod := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "target-pod",
						"namespace": "default",
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate, pod).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateOpened))
			Expect(meta.FindStatusCondition(gate.Status.Conditions, gateshv1alpha1.GateStateOpened).Status).To(Equal(metav1.ConditionTrue))
			Expect(meta.FindStatusCondition(gate.Status.Conditions, gateshv1alpha1.GateStateClosed).Status).To(Equal(metav1.ConditionFalse))
		})

		It("should close the gate for a single target by name with non-matching jsonpointer", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "target-pod",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								Name:       "target-pod",
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									JsonPointer: gateshv1alpha1.GateTargetValidatorJsonPointer{
										JsonPointer: "/metadata/name",
										Value:       "no-name",
									},
								},
							},
						},
					},
				},
			}

			pod := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "target-pod",
						"namespace": "default",
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate, pod).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateClosed))
			Expect(meta.FindStatusCondition(gate.Status.Conditions, gateshv1alpha1.GateStateOpened).Status).To(Equal(metav1.ConditionFalse))
			Expect(meta.FindStatusCondition(gate.Status.Conditions, gateshv1alpha1.GateStateClosed).Status).To(Equal(metav1.ConditionTrue))
		})

		It("should close the gate for a single target by name with a missing jsonpointer", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "target-pod",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								Name:       "target-pod",
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									JsonPointer: gateshv1alpha1.GateTargetValidatorJsonPointer{
										JsonPointer: "/metadata/not-found",
										Value:       "target-pod",
									},
								},
							},
						},
					},
				},
			}

			pod := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "target-pod",
						"namespace": "default",
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate, pod).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateClosed))
			Expect(meta.FindStatusCondition(gate.Status.Conditions, gateshv1alpha1.GateStateOpened).Status).To(Equal(metav1.ConditionFalse))
			Expect(meta.FindStatusCondition(gate.Status.Conditions, gateshv1alpha1.GateStateClosed).Status).To(Equal(metav1.ConditionTrue))
		})

		It("should close the gate if no objects are found for a target", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "non-existent-pod",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								Name:       "target-pod",
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateClosed))
			Expect(gate.Status.TargetConditions).To(HaveLen(1))
			Expect(gate.Status.TargetConditions[0].Status).To(Equal(metav1.ConditionFalse))
			Expect(gate.Status.TargetConditions[0].Message).To(ContainSubstring("0 objects found\n0 objects match target validators\n0/1 valid objects"))
		})

		It("should close the gate for multiple targets with AND operation where one fails", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Operation: gateshv1alpha1.GateOperation{
						Operator: gateshv1alpha1.GateOperatorAnd,
					},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "pod1",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								Name:       "pod1",
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
						{
							Name: "pod2",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								Name:       "pod2",
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			}

			pod1 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod1",
						"namespace": "default",
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			}

			pod2 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod2",
						"namespace": "default",
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
							},
						},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate, pod1, pod2).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateClosed))
			Expect(gate.Status.TargetConditions).To(HaveLen(2))
			Expect(gate.Status.TargetConditions[0].Status).To(Equal(metav1.ConditionTrue))
			Expect(gate.Status.TargetConditions[1].Status).To(Equal(metav1.ConditionFalse))
		})

		It("should open the gate for multiple targets with OR operation where one succeeds", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Operation: gateshv1alpha1.GateOperation{
						Operator: gateshv1alpha1.GateOperatorOr,
					},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "pod1",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								Name:       "pod1",
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
						{
							Name: "pod2",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								Name:       "pod2",
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			}

			pod1 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod1",
						"namespace": "default",
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			}

			pod2 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod2",
						"namespace": "default",
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
							},
						},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate, pod1, pod2).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateOpened))
			Expect(gate.Status.TargetConditions).To(HaveLen(2))
			Expect(gate.Status.TargetConditions[0].Status).To(Equal(metav1.ConditionTrue))
			Expect(gate.Status.TargetConditions[1].Status).To(Equal(metav1.ConditionFalse))
		})

		It("should invert the result when Operation.Invert is true", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Operation: gateshv1alpha1.GateOperation{
						Invert: true,
					},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "target-pod",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								Name:       "target-pod",
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			}

			pod := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "target-pod",
						"namespace": "default",
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate, pod).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateClosed))
		})

		It("should handle label selector with multiple matching objects, all matching validators", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "target-pods",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								LabelSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "test"},
								},
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			}

			pod1 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod1",
						"namespace": "default",
						"labels": map[string]interface{}{
							"app": "test",
						},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			}

			pod2 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod2",
						"namespace": "default",
						"labels": map[string]interface{}{
							"app": "test",
						},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate, pod1, pod2).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateOpened))
			Expect(gate.Status.TargetConditions[0].Status).To(Equal(metav1.ConditionTrue))
			Expect(gate.Status.TargetConditions[0].Message).To(ContainSubstring("2 objects found\n2 objects match target validators\n2/2 valid objects"))
		})

		It("should close the gate if not all objects match validators in label selector", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "target-pods",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								LabelSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "test"},
								},
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			}

			pod1 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod1",
						"namespace": "default",
						"labels": map[string]interface{}{
							"app": "test",
						},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			}

			pod2 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod2",
						"namespace": "default",
						"labels": map[string]interface{}{
							"app": "test",
						},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
							},
						},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate, pod1, pod2).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateClosed))
			Expect(gate.Status.TargetConditions[0].Status).To(Equal(metav1.ConditionFalse))
			Expect(gate.Status.TargetConditions[0].Message).To(ContainSubstring("2 objects found\n1 objects match target validators\n1/2 valid objects"))
		})

		It("should set error condition if invalid label selector", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "target-pods",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								LabelSelector: metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "app",
											Operator: "InvalidOperator",
											Values:   []string{"test"},
										},
									},
								},
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateClosed))
			Expect(gate.Status.TargetConditions[0].Status).To(Equal(metav1.ConditionFalse))
			Expect(gate.Status.TargetConditions[0].Reason).To(Equal("ErrorWhileFetching"))
			Expect(gate.Status.TargetConditions[0].Message).To(ContainSubstring("invalid label selector"))
		})

		It("should handle target with atLeast", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gate", Namespace: "default"},
				Spec: gateshv1alpha1.GateSpec{
					RequeueAfter: &metav1.Duration{Duration: 5 * time.Minute},
					Targets: []gateshv1alpha1.GateTarget{
						{
							Name: "target-pods",
							Selector: gateshv1alpha1.GateTargetSelector{
								ApiVersion: "v1",
								Kind:       "Pod",
								LabelSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "test"},
								},
							},
							Validators: []gateshv1alpha1.GateTargetValidator{
								{
									AtLeast: 1,
									MatchCondition: gateshv1alpha1.GateTargetValidatorMatchCondition{
										Type:   "Ready",
										Status: metav1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			}

			pod1 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod1",
						"namespace": "default",
						"labels": map[string]interface{}{
							"app": "test",
						},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			}

			pod2 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod2",
						"namespace": "default",
						"labels": map[string]interface{}{
							"app": "test",
						},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
							},
						},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gate, pod1, pod2).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			err := reconciler.Reconcile()
			Expect(err).NotTo(HaveOccurred())

			Expect(gate.Status.State).To(Equal(gateshv1alpha1.GateStateOpened))
			Expect(gate.Status.TargetConditions[0].Status).To(Equal(metav1.ConditionTrue))
			Expect(gate.Status.TargetConditions[0].Message).To(ContainSubstring("1/1 valid objects")) // atLeast=1 in message, but result ignores it
		})
	})

	Describe("FetchGateTargetObjects", func() {
		It("should fetch object by name", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
			}
			target := &gateshv1alpha1.GateTarget{
				Name: "target-pod",
				Selector: gateshv1alpha1.GateTargetSelector{
					ApiVersion: "v1",
					Kind:       "Pod",
					Name:       "target-pod",
				},
			}

			pod := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "target-pod",
						"namespace": "default",
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			objects, err := reconciler.FetchGateTargetObjects(target)
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(HaveLen(1))
			Expect(objects[0].GetName()).To(Equal("target-pod"))
		})

		It("should return empty if object by name not found", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
			}
			target := &gateshv1alpha1.GateTarget{
				Name: "non-existent",
				Selector: gateshv1alpha1.GateTargetSelector{
					ApiVersion: "v1",
					Kind:       "Pod",
					Name:       "target-pod",
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			objects, err := reconciler.FetchGateTargetObjects(target)
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(BeEmpty())
		})

		It("should fetch objects by label selector", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
			}
			target := &gateshv1alpha1.GateTarget{
				Name: "target-pods",
				Selector: gateshv1alpha1.GateTargetSelector{
					ApiVersion: "v1",
					Kind:       "Pod",
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
				},
			}

			pod1 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod1",
						"namespace": "default",
						"labels": map[string]interface{}{
							"app": "test",
						},
					},
				},
			}

			pod2 := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod2",
						"namespace": "default",
						"labels": map[string]interface{}{
							"app": "other",
						},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod1, pod2).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			objects, err := reconciler.FetchGateTargetObjects(target)
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(HaveLen(1))
			Expect(objects[0].GetName()).To(Equal("pod1"))
		})

		It("should return error if both name and label selector are set", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
			}
			target := &gateshv1alpha1.GateTarget{
				Name: "target-pod",
				Selector: gateshv1alpha1.GateTargetSelector{
					ApiVersion: "v1",
					Kind:       "Pod",
					Name:       "target-pod",
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			objects, err := reconciler.FetchGateTargetObjects(target)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name and labelSelector are mutually exclusive"))
			Expect(objects).To(BeNil())
		})

		It("should return error if neither name nor label selector is set", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
			}
			target := &gateshv1alpha1.GateTarget{
				Name: "target",
				Selector: gateshv1alpha1.GateTargetSelector{
					ApiVersion: "v1",
					Kind:       "Pod",
				},
			}
			target.Name = "" // Clear name

			cl := fake.NewClientBuilder().WithScheme(scheme).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			objects, err := reconciler.FetchGateTargetObjects(target)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("either name or labelSelector must be specified"))
			Expect(objects).To(BeNil())
		})

		It("should use target namespace if specified", func() {
			gate := &gateshv1alpha1.Gate{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
			}
			target := &gateshv1alpha1.GateTarget{
				Name: "target-pod",
				Selector: gateshv1alpha1.GateTargetSelector{
					ApiVersion: "v1",
					Kind:       "Pod",
					Namespace:  "custom-ns",
					Name:       "target-pod",
				},
			}

			pod := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "target-pod",
						"namespace": "custom-ns",
					},
				},
			}

			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()

			reconciler := GateCommonReconciler{
				Context: ctx,
				Client:  cl,
				Gate:    gate,
			}

			objects, err := reconciler.FetchGateTargetObjects(target)
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(HaveLen(1))
			Expect(objects[0].GetNamespace()).To(Equal("custom-ns"))
		})
	})

	Describe("GetObjectStatusConditions", func() {
		It("should extract conditions from unstructured object", func() {
			obj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
							map[string]interface{}{
								"type":   "Initialized",
								"status": "True",
							},
						},
					},
				},
			}

			reconciler := GateCommonReconciler{} // No need for full setup

			conditions, err := reconciler.GetObjectStatusConditions(obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(conditions).To(HaveLen(2))
			Expect(conditions[0].Type).To(Equal("Ready"))
			Expect(conditions[0].Status).To(Equal(metav1.ConditionTrue))
		})

		It("should return nil if no status field", func() {
			obj := &unstructured.Unstructured{
				Object: map[string]interface{}{},
			}

			reconciler := GateCommonReconciler{}

			conditions, err := reconciler.GetObjectStatusConditions(obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(conditions).To(BeNil())
		})

		It("should return nil if no conditions field", func() {
			obj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{},
				},
			}

			reconciler := GateCommonReconciler{}

			conditions, err := reconciler.GetObjectStatusConditions(obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(conditions).To(BeNil())
		})

		It("should skip invalid condition entries", func() {
			obj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
							"invalid-string", // Bad type
						},
					},
				},
			}

			reconciler := GateCommonReconciler{}

			conditions, err := reconciler.GetObjectStatusConditions(obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0].Type).To(Equal("Ready"))
		})
	})

	Describe("GetObjectFieldByJsonPointer", func() {
		It("should extract field from json pointer", func() {
			obj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"health": map[string]any{
							"state": "progressing",
						},
					},
				},
			}

			reconciler := GateCommonReconciler{}
			value, err := reconciler.GetObjectFieldByJsonPointer(obj, "/status/health/state")
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("progressing"))
		})

		It("should return an error if the pointer points to nothing", func() {
			obj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{},
				},
			}

			reconciler := GateCommonReconciler{}
			value, err := reconciler.GetObjectFieldByJsonPointer(obj, "/status/health/state")
			Expect(err).To(HaveOccurred())
			Expect(value).To(BeNil())
		})

		It("should return an error if the pointer is invalid", func() {
			obj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{},
				},
			}

			reconciler := GateCommonReconciler{}
			value, err := reconciler.GetObjectFieldByJsonPointer(obj, "invalid.pointer")
			Expect(err).To(HaveOccurred())
			Expect(value).To(BeNil())
		})
	})

	Describe("ComputeOperation", func() {
		var reconciler GateCommonReconciler
		BeforeEach(func() {
			reconciler = GateCommonReconciler{
				Gate: &gateshv1alpha1.Gate{
					Spec: gateshv1alpha1.GateSpec{},
				},
			}
		})

		It("should return true for AND with all true conditions", func() {
			conditions := []metav1.Condition{
				{Status: metav1.ConditionTrue},
				{Status: metav1.ConditionTrue},
			}
			result := reconciler.ComputeOperation(conditions)
			Expect(result).To(BeTrue())
		})

		It("should return false for AND with one false condition", func() {
			conditions := []metav1.Condition{
				{Status: metav1.ConditionTrue},
				{Status: metav1.ConditionFalse},
			}
			result := reconciler.ComputeOperation(conditions)
			Expect(result).To(BeFalse())
		})

		It("should return true for OR with one true condition", func() {
			reconciler.Gate.Spec.Operation.Operator = gateshv1alpha1.GateOperatorOr
			conditions := []metav1.Condition{
				{Status: metav1.ConditionFalse},
				{Status: metav1.ConditionTrue},
			}
			result := reconciler.ComputeOperation(conditions)
			Expect(result).To(BeTrue())
		})

		It("should return false for OR with all false conditions", func() {
			reconciler.Gate.Spec.Operation.Operator = gateshv1alpha1.GateOperatorOr
			conditions := []metav1.Condition{
				{Status: metav1.ConditionFalse},
				{Status: metav1.ConditionFalse},
			}
			result := reconciler.ComputeOperation(conditions)
			Expect(result).To(BeFalse())
		})

		It("should invert the result if Invert is true", func() {
			reconciler.Gate.Spec.Operation.Invert = true
			conditions := []metav1.Condition{
				{Status: metav1.ConditionTrue},
			}
			result := reconciler.ComputeOperation(conditions)
			Expect(result).To(BeFalse())
		})
	})
})

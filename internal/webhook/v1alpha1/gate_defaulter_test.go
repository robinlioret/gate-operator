package v1alpha1

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gateshv1alpha1 "github.com/robinlioret/gate-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	schemeBuilder "k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("GateDefaulter webhook", func() {
	scheme := runtime.NewScheme()
	_ = schemeBuilder.AddToScheme(scheme)
	_ = gateshv1alpha1.AddToScheme(scheme) // Add CRDs

	It("should apply default value to an empty gate spec", func() {
		By("Creating the empty gate spec")
		spec := gateshv1alpha1.GateSpec{}
		ApplyDefaultSpec(&spec)
		Expect(spec.EvaluationPeriod).To(Equal(&metav1.Duration{Duration: 60 * time.Second}))
		Expect(spec.Consolidation.Delay).To(Equal(&metav1.Duration{Duration: 5 * time.Second}))
		Expect(spec.Consolidation.Count).To(Equal(1))
		Expect(spec.Operation.Operator).To(Equal(gateshv1alpha1.GateOperatorAnd))
		Expect(spec.Targets).To(HaveLen(0))
	})
})

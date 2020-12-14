package componentry

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	adregv1 "k8s.io/api/admissionregistration/v1"
)

var _ = Describe("WebhookData", func() {
	Context("With a basic WebhookData instance", func() {
		var failPolicy adregv1.FailurePolicyType = "Fail"
		var noneSideEffect adregv1.SideEffectClass = "None"
		mutate := "/mutate"
		tenSeconds := int32(10)

		name := "foo"
		nameVal := "bar"
		annotations := map[string]string{name: nameVal}
		mwh := []adregv1.MutatingWebhook{
			{
				Name: name,
				AdmissionReviewVersions: []string{
					"v1",
					"v1beta1",
				},
				ClientConfig: adregv1.WebhookClientConfig{
					Service: &adregv1.ServiceReference{
						Name: "some-service",
						Path: &mutate,
					},
				},
				FailurePolicy:  &failPolicy,
				SideEffects:    &noneSideEffect,
				TimeoutSeconds: &tenSeconds,
				Rules: []adregv1.RuleWithOperations{
					{
						Operations: []adregv1.OperationType{adregv1.Create, adregv1.Update},
						Rule: adregv1.Rule{
							Resources:   []string{"*/*"},
							APIGroups:   []string{"example.com"},
							APIVersions: []string{"*"},
						},
					},
				},
			},
		}

		vwh := []adregv1.ValidatingWebhook{}

		wh := WebhookData{
			name:               name,
			annotations:        annotations,
			mutatingWebhooks:   mwh,
			validatingWebhooks: vwh,
		}

		It("Should provide access (via getter) to instance key: name", func() {
			Expect(wh.GetName()).To(Equal(name))
		})

		It("Should provide access (via getter) to instance key: annotations", func() {
			Expect(wh.GetAnnotations()[name]).To(Equal(nameVal))
		})

		It("Should provide access (via getter) to instance key mutatingWebhooks", func() {
			h := wh.GetMutatingWebhooks()[0]
			Expect(h.Name).To(Equal(name))
		})
		It("Should provide access (via getter) to instance key validatingWebhooks", func() {
			h := wh.GetValidatingWebhooks()
			Expect(len(h)).To(BeZero())
		})
	})

	Context("With no name or webhook data", func() {
		wh := WebhookData{}
		It("Should indicate that it is an empty instance", func() {
			Expect(wh.IsEmpty()).To(BeTrue())
		})
	})
})

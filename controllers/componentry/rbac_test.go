package componentry

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
)

var _ = Describe("RoleData", func() {
	nameVal := "foo"
	nameLabelVal := "bar"
	aggregateVal := false
	labelsVal := map[string]string{nameVal: nameLabelVal}
	verb := "create"
	policyRulesVal := []rbacv1.PolicyRule{
		{
			Verbs:     []string{verb},
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
		},
	}

	rd := RoleData{
		name:        nameVal,
		isAggregate: aggregateVal,
		labels:      labelsVal,
		policyRules: policyRulesVal,
	}

	Context("Given a basic RoleData instance", func() {
		It("Should provide access (via getter) to instance key: name", func() {
			Expect(rd.GetName()).To(Equal(nameVal))
		})
		It("Should provide access (via getter) to instance key: isAggregate", func() {
			Expect(rd.IsAggregate()).To(Equal(false))
		})
		It("Should provide access (via getter) to instance key: labels", func() {
			Expect(rd.labels[nameVal]).To(Equal(nameLabelVal))
		})
		It("Should provide access (via getter) to instance key: policyRules", func() {
			Expect(rd.policyRules[0].Verbs[0]).To(Equal(verb))
		})
	})
})

// Package matchers provides some custom matchers for use in testing this operator.
package matchers

import (
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// BeReady marries observed generation checks and pod readiness checks on an appsv1.Deployment
// to determine if a deployment has successfully rolled out.
func BeReady() types.GomegaMatcher {
	return gomega.SatisfyAll(
		HaveObservedGenerationMatchExpectedGeneration(),
		HaveReadyPodsMatchDesiredPods(),
	)
}

// HaveSuccessfullyRolledOut is an alias for BeReady. Should only be used with
// appsv1.Deployment resources.
func HaveSuccessfullyRolledOut() types.GomegaMatcher {
	return BeReady()
}

// BeReadyAndOnTheLatestRevision must satisfy all parameters of pod readiness that BeReady()
// implements as well as ensure the StatefulSet.Status.currentRevision == StatefulSet.Status.updateRevision.
// Passing something other than a StatefulSet will fail the assertion.
func BeReadyAndOnTheLatestRevision() types.GomegaMatcher {
	return gomega.SatisfyAll(
		BeReady(),
		BeOnTheLatestRevision(),
	)
}

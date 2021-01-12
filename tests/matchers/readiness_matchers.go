package matchers

import (
	"fmt"

	"github.com/onsi/gomega/types"
	appsv1 "k8s.io/api/apps/v1"
)

// HaveObservedGenerationMatchExpectedGeneration checks to see if a
// deployment/statefulset/daemonset's observed generation matches its
// metadata generation field indicating that the state resource is as
// expected. Generations or observedGeneration with a value as 0 is
// assumed to be an uninstantiated resource and returns an error.
func HaveObservedGenerationMatchExpectedGeneration() types.GomegaMatcher {
	return &haveObservedGenerationMatchExpectedGenerationMatcher{}
}

type haveObservedGenerationMatchExpectedGenerationMatcher struct {
	typeMismatchError  bool
	observedGeneration int64
	generation         int64
}

func (matcher *haveObservedGenerationMatchExpectedGenerationMatcher) Match(actual interface{}) (bool, error) {
	// This is repetitive, but I'm unsure of how to support all three resource Status blocks
	// as I haven't found a shared interface that I can assert to that gives access to observedGeneration
	switch v := actual.(type) {
	case appsv1.Deployment:
		matcher.observedGeneration = v.Status.ObservedGeneration
		matcher.generation = v.Generation
	case appsv1.StatefulSet:
		matcher.observedGeneration = v.Status.ObservedGeneration
		matcher.generation = v.Generation
	case appsv1.DaemonSet:
		matcher.observedGeneration = v.Status.ObservedGeneration
		matcher.generation = v.Generation
	default:
		matcher.typeMismatchError = true
		return false, nil
	}

	if (matcher.observedGeneration == 0) || (matcher.generation == 0) {
		return false, fmt.Errorf(
			"expected non-zero generation/observedGeneration\nGeneration: %d\nObserved Generation: %d",
			matcher.generation,
			matcher.observedGeneration)
	}

	return matcher.observedGeneration == matcher.generation, nil
}

func (matcher *haveObservedGenerationMatchExpectedGenerationMatcher) FailureMessage(actual interface{}) string {
	if matcher.typeMismatchError {
		return ("Expected input resource to be an appsv1.Deployment, appsv1.StatefulSet, or appsv1.Daemonset")
	}

	return fmt.Sprintf("Expected observedGeneration %d to match generation %d", matcher.observedGeneration, matcher.generation)
}

func (matcher *haveObservedGenerationMatchExpectedGenerationMatcher) NegatedFailureMessage(actual interface{}) string {
	if matcher.typeMismatchError {
		return ("Expected input resource to not be an appsv1.Deployment, appsv1.StatefulSet, or appsv1.Daemonset")
	}

	return fmt.Sprintf("Expected observedGeneration %d to not match generation %d", matcher.observedGeneration, matcher.generation)
}

// HaveReadyPodsMatchDesiredPods checks to see if the ready replica
// count matches the desired replica count to determine if the
// deployment/statefulset/daemonset is ready.
func HaveReadyPodsMatchDesiredPods() types.GomegaMatcher {
	return &haveReadyPodsMatchDesiredPodsMatcher{}
}

type haveReadyPodsMatchDesiredPodsMatcher struct {
	typeMismatchError bool
	desiredReplicas   int32
	readyReplicas     int32
}

func (matcher *haveReadyPodsMatchDesiredPodsMatcher) Match(actual interface{}) (bool, error) {
	switch v := actual.(type) {
	case appsv1.Deployment:
		matcher.desiredReplicas = *v.Spec.Replicas
		matcher.readyReplicas = v.Status.ReadyReplicas
	case appsv1.StatefulSet:
		matcher.desiredReplicas = *v.Spec.Replicas
		matcher.readyReplicas = v.Status.ReadyReplicas
	case appsv1.DaemonSet:
		matcher.desiredReplicas = v.Status.DesiredNumberScheduled
		matcher.readyReplicas = v.Status.NumberReady
	default:
		matcher.typeMismatchError = true
		return false, nil
	}

	return matcher.desiredReplicas == matcher.readyReplicas, nil
}

func (matcher *haveReadyPodsMatchDesiredPodsMatcher) FailureMessage(actual interface{}) string {
	if matcher.typeMismatchError {
		return ("Expected input resource to not be an appsv1.Deployment")
	}

	return fmt.Sprintf("Expected ReadyReplicas %d to match DesiredReplicas %d", matcher.readyReplicas, matcher.desiredReplicas)
}

func (matcher *haveReadyPodsMatchDesiredPodsMatcher) NegatedFailureMessage(actual interface{}) string {
	if matcher.typeMismatchError {
		return ("Expected input resource to not be an appsv1.Deployment")
	}

	return fmt.Sprintf("Expected ReadyReplicas %d to not match DesiredReplicas %d", matcher.readyReplicas, matcher.desiredReplicas)
}

// BeALaterGenerationThan accepts a before and after generation to evaluate
// if the after generation is later than the before generation. Used to check
// if rollouts have completed on the same resource.
func BeALaterGenerationThan(before int64) types.GomegaMatcher {
	return &beALaterGenerationThanMatcher{
		generationBefore: before,
	}
}

type beALaterGenerationThanMatcher struct {
	generationBefore int64
}

func (matcher *beALaterGenerationThanMatcher) Match(actual interface{}) (bool, error) {
	in, ok := actual.(int64)
	if !ok {
		return false, nil
	}

	return in > matcher.generationBefore, nil
}

func (matcher *beALaterGenerationThanMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected generation %d to be newer than original generation %d", actual.(int64), matcher.generationBefore)
}

func (matcher *beALaterGenerationThanMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected generation %d to not be newer than original generation %d", actual.(int64), matcher.generationBefore)
}

// BeOnTheLatestRevision is a check for Statefulsets to ensure the updateRevision matches
// the currentRevision to help determine if a rollout has completed.
func BeOnTheLatestRevision() types.GomegaMatcher {
	return &beOnTheLatestRevisionMatcher{}
}

type beOnTheLatestRevisionMatcher struct {
	typeMismatchError bool
	valueEmptyError   bool
	currentRevision   string
	targetRevision    string
}

func (matcher *beOnTheLatestRevisionMatcher) Match(actual interface{}) (bool, error) {
	in, ok := actual.(appsv1.StatefulSet)
	if !ok {
		matcher.typeMismatchError = ok
		return false, nil
	}

	if in.Status.CurrentRevision == "" || in.Status.UpdateRevision == "" {
		matcher.valueEmptyError = true
		return false, nil
	}

	matcher.currentRevision = in.Status.CurrentRevision
	matcher.targetRevision = in.Status.UpdateRevision
	return matcher.currentRevision == matcher.targetRevision, nil
}

func (matcher *beOnTheLatestRevisionMatcher) FailureMessage(actual interface{}) string {
	if matcher.typeMismatchError {
		return ("Expected input resource to be an appsv1.StatefulSet")
	}

	if matcher.valueEmptyError {
		return fmt.Sprintf("Expected both currentRevision and updateRevision to be non-empty values\ncurrentRevision: %s\nupdateRevision: %s", matcher.currentRevision, matcher.targetRevision)
	}

	return fmt.Sprintf("Expected currentRevision %s to match updateRevision %s", matcher.currentRevision, matcher.targetRevision)
}

func (matcher *beOnTheLatestRevisionMatcher) NegatedFailureMessage(actual interface{}) string {
	if matcher.typeMismatchError {
		return ("Expected input resource to not be an appsv1.StatefulSet")
	}

	return fmt.Sprintf("Expected currentRevision %s not to match updateRevision %s", matcher.currentRevision, matcher.targetRevision)
}

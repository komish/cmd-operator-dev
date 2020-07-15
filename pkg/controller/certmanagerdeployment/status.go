package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

// deploymentState is a type to help facilitate reading the current state of existing deployments
// in the cluster.
type deploymentState struct {
	// Count is the number of deployments
	count                 int
	availableMatchesReady []bool
	readyMatchesDesired   []bool
}

// allAvailableAreReady return true if all deployments add to the deploymentState struct have matching
// available replicas (as determined by their status) and matching ready replicas (as determined by their status).
func (ds *deploymentState) allAvailableAreReady() bool {
	res := true
	for _, v := range ds.availableMatchesReady {
		if !v {
			res = false
		}
	}

	return res
}

// readyCountMatchesDesiredCount return true if all the deployments added to the deploymentState struct have matching
// ready replicas (as determined by their status) and matching desired replicas (as determined by their spec).
func (ds *deploymentState) readyCountMatchesDesiredCount() bool {
	res := true
	for _, v := range ds.readyMatchesDesired {
		if !v {
			res = false
		}
	}

	return res
}

// deploymentCountMatchesCountOfStoredStates will return true if the number of stored states match the expected number of deployments
// to be evaluated as a part of this struct (stored in count).
func (ds *deploymentState) deploymentCountMatchesCountOfStoredStates() bool {
	return len(ds.availableMatchesReady) == len(ds.readyMatchesDesired) && ds.count == len(ds.availableMatchesReady)
}

// crdState is a type to help facilitate reading the current state of existing CRDs
// in the cluster.
type crdState struct {
	// Count is the number of CRDs
	count             int
	crdIsEstablished  []bool
	crdNameIsAccepted []bool
}

// allAreEstablished return true if all CRD added to the crdState struct have a status.condition
// Established and the status of that condition is true.
func (cs *crdState) allAreEstablished() bool {
	res := true
	for _, v := range cs.crdIsEstablished {
		if !v {
			res = false
		}
	}

	return res
}

// allNamesAreAccepted returns true if all the CRDs added to the crdState struct have a status.condition
// NameAccepted and the status of that condition is true.
func (cs *crdState) allNamesAreAccepted() bool {
	res := true
	for _, v := range cs.crdNameIsAccepted {
		if !v {
			res = false
		}
	}

	return res
}

// crdCountMatchesCountOfStoredStates will return true if the number of stored states match the expected number of CRDs
// to be evaluated as a part of this struct (stored in count).
func (cs *crdState) crdCountMatchesCountOfStoredStates() bool {
	return len(cs.crdIsEstablished) == len(cs.crdNameIsAccepted) && cs.count == len(cs.crdIsEstablished)
}

// getUninitializedCertManagerDeploymentStatus returns a CertManagerDeploymentStatus with unknown values
// to be modified and added to the API.
func getUninitializedCertManagerDeploymentStatus() redhatv1alpha1.CertManagerDeploymentStatus {
	return redhatv1alpha1.CertManagerDeploymentStatus{
		Version: "unknown",
		Phase:   "unknown",
	}
}

// getStateOfDeployments compares the list
func getStateOfDeployments(existingDeploys []*appsv1.Deployment) deploymentState {
	state := deploymentState{count: len(existingDeploys)}
	for _, deploy := range existingDeploys {
		state.availableMatchesReady = append(state.availableMatchesReady, (deploy.Status.AvailableReplicas == deploy.Status.ReadyReplicas))
		state.readyMatchesDesired = append(state.readyMatchesDesired, (*deploy.Spec.Replicas == deploy.Status.ReadyReplicas))
	}
	return state
}

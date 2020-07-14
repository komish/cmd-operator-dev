package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
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

// getNewCertManagerDeploymentStatus returns a CertManagerDeploymentStatus with unknown values
// to be modified and added to the API.
func getNewCertManagerDeploymentStatus() redhatv1alpha1.CertManagerDeploymentStatus {
	return redhatv1alpha1.CertManagerDeploymentStatus{
		Version: "unknown",
		Phase:   "unknown",
	}
}

// func setVersionStatus will update a status object in place with the version that's provided.
func setVersionStatus(status *redhatv1alpha1.CertManagerDeploymentStatus, version string) {
	status.Version = version
}

// func setPhaseStatus will update a status object in place with the phase that's provided.
func setPhaseStatus(status *redhatv1alpha1.CertManagerDeploymentStatus, phase string) {
	status.Phase = phase
}

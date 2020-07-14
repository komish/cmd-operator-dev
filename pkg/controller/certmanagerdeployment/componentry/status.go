package componentry

const (
	// StatusPhasePending indicates that the object has persisted to the API but downstream objects have
	// not been persisted to the API.
	StatusPhasePending string = "pending"

	// StatusPhaseProgressing indicates that the object and some downstream objects have persisted to the API
	// but are not all in a ready state.
	StatusPhaseProgressing string = "progressing"

	// StatusPhaseRunning indicates that the object as well as all downstream objects have been successfully
	// persisted to the API and are running/functional.
	StatusPhaseRunning string = "running"
)

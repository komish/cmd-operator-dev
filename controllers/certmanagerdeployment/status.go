package certmanagerdeployment

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// reconcileStatus reconciles the status block of a CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileStatus(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: status")
	defer reqLogger.Info("Ending reconciliation: status")

	// get an empty status, copy the object we're working with,
	// and get a getter to query for expected states of resources.
	status := getUninitializedCertManagerDeploymentStatus()
	obj := instance.DeepCopy()
	getter := ResourceGetter{CustomResource: *instance}

	r.reconcileStatusVersion(status, cmdoputils.CRVersionOrDefaultVersion(instance.Spec.Version, componentry.CertManagerDefaultVersion))
	r.reconcileStatusDeploymentsHealthy(status, getter, reqLogger)
	r.reconcileStatusCRDsHealthy(status, getter, reqLogger)
	r.reconcileStatusPhase(status)

	// Update the object with new status
	obj.Status = *status
	reqLogger.V(2).Info("Updating Status for object", "CertManagerDeployment.Name", instance.GetName())
	if err := r.Status().Update(context.TODO(), obj); err != nil {
		reqLogger.Info("Error updating CertManagerDeployment's Status", "name", instance.GetName())
		return err
	}

	return nil
}

// deploymentState facilitates identifying the current
// state of managed deployment resources.
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

// crdState facilitates determining the current state of managed CRDs.
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
func getUninitializedCertManagerDeploymentStatus() *operatorsv1alpha1.CertManagerDeploymentStatus {
	return &operatorsv1alpha1.CertManagerDeploymentStatus{
		Version:    "unknown",
		Phase:      "unknown",
		Conditions: []operatorsv1alpha1.CertManagerDeploymentCondition{},
	}
}

// deploymentsAreReady evaluates the status fields in existingDeploys, and return true
// if all existingDeploys are in an acceptable state.
func deploymentsAreReady(existingDeploys []*appsv1.Deployment) corev1.ConditionStatus {
	state, ok := deploymentState{count: len(existingDeploys)}, false
	for _, deploy := range existingDeploys {
		state.availableMatchesReady = append(state.availableMatchesReady, (deploy.Status.AvailableReplicas == deploy.Status.ReadyReplicas))
		state.readyMatchesDesired = append(state.readyMatchesDesired, (*deploy.Spec.Replicas == deploy.Status.ReadyReplicas))
	}

	if state.allAvailableAreReady() &&
		state.readyCountMatchesDesiredCount() &&
		state.deploymentCountMatchesCountOfStoredStates() {
		ok = true
	}

	// so much formatting to get this to work...
	return corev1.ConditionStatus(strings.Title(strconv.FormatBool(ok)))
}

// crdsAreReady evaluates the status fields in existingCRDs and returns true
// if all existingCRDs are in an acceptable state.
func crdsAreReady(existingCRDs []*apiextv1.CustomResourceDefinition) corev1.ConditionStatus {
	state, ok := crdState{count: len(existingCRDs)}, false
	for _, crd := range existingCRDs {
		// search specifically for conditions we care about - NamesAccepted and Established
		// it is possible that NamesAccepted or Established is not available in the condition list
		// in which case we want to return false.
		accepted, established := false, false
		conditions := crd.Status.Conditions
		for _, condition := range conditions {
			switch condition.Type {
			case "Established":
				if condition.Status == apiextv1.ConditionTrue {
					established = true
				}
			case "NamesAccepted":
				if condition.Status == apiextv1.ConditionTrue {
					accepted = true
				}
			}
		}
		state.crdIsEstablished = append(state.crdIsEstablished, established)
		state.crdNameIsAccepted = append(state.crdNameIsAccepted, accepted)
	}
	if state.allAreEstablished() &&
		state.allNamesAreAccepted() &&
		state.crdCountMatchesCountOfStoredStates() {
		ok = true
	}

	return corev1.ConditionStatus(strings.Title(strconv.FormatBool(ok)))
}

// reconcileStatusVersion is a subreconciliation function called by ReconcileStatus that injects the version
// of cert-manager desired into the status field.
func (r *CertManagerDeploymentReconciler) reconcileStatusVersion(inStatus *operatorsv1alpha1.CertManagerDeploymentStatus, vers string) *operatorsv1alpha1.CertManagerDeploymentStatus {
	inStatus.Version = vers
	// TODO() check deployed versions of things to make sure they reflect the right version.
	// requires updating all types that we create with some form of identifier stating which
	// version is deployed for that CR.
	return inStatus
}

// reconcileStatusDeploymentsHealthy updated the DeploymentsHealthy status field. This
// checks if the deployments expected for the CR instance exists per the API and are
// ready and available with the expected replica count.
func (r *CertManagerDeploymentReconciler) reconcileStatusDeploymentsHealthy(
	inStatus *operatorsv1alpha1.CertManagerDeploymentStatus,
	rg ResourceGetter,
	reqLogger logr.Logger) *operatorsv1alpha1.CertManagerDeploymentStatus {

	condition := operatorsv1alpha1.CertManagerDeploymentCondition{
		Type:    operatorsv1alpha1.ConditionDeploymentsAreReady,
		Status:  corev1.ConditionUnknown,
		Reason:  "AllDeploymentsHealthy",
		Message: "Deployment available and ready pods matches desired.",
	}

	existingDeploys, ok := queryAPIForExpectedDeployments(r, rg, reqLogger)
	if existingDeploys == nil {
		reqLogger.Info("unable to determine status of expected deployments for this instance")
	}

	if ok {
		condition.Status = deploymentsAreReady(existingDeploys)
	}

	condition.LastUpdateTime = metav1.Now()
	inStatus.Conditions = append(inStatus.Conditions, condition)

	// bubble up deployment conditions to the CR.
	updateStatusDeploymentConditions(inStatus, existingDeploys)

	return inStatus
}

// reconcileStatusCRDsHealthy updated the DeploymentsHealthy status field. This
// checks if the deployments expected for the CR instance exists per the API and are
// ready and available with the expected replica count.
func (r *CertManagerDeploymentReconciler) reconcileStatusCRDsHealthy(
	inStatus *operatorsv1alpha1.CertManagerDeploymentStatus,
	rg ResourceGetter,
	reqLogger logr.Logger) *operatorsv1alpha1.CertManagerDeploymentStatus {

	condition := operatorsv1alpha1.CertManagerDeploymentCondition{
		Type:    operatorsv1alpha1.ConditionCRDsAreReady,
		Status:  corev1.ConditionUnknown,
		Reason:  "AllCRDsHealthy",
		Message: "CRDs NamesAccepted and Established Conditions are true.",
	}
	existingCRDs, ok := queryAPIForExpectedCRDs(r, rg, reqLogger)
	if existingCRDs == nil {
		reqLogger.Info("unable to determine status of expected CRDs for this instance")
	}

	if ok {
		condition.Status = crdsAreReady(existingCRDs)
	}

	condition.LastUpdateTime = metav1.Now()
	inStatus.Conditions = append(inStatus.Conditions, condition)

	// bubble up crd conditions to the CR
	updateStatusCRDConditions(inStatus, existingCRDs)

	return inStatus
}

// queryAPIForExpectedDeployments will check that the deployments expected for a given instance actually
// exist in the API. will return ok as true when all were found, and false if not. Will return the
// deployment slice as nil if an error other than IsNotfound was encountered trying to obtain the data, as well
// as return ok as false.
func queryAPIForExpectedDeployments(client client.Client, rg ResourceGetter, reqLogger logr.Logger) ([]*appsv1.Deployment, bool) {
	foundDeployments := make([]*appsv1.Deployment, 0)
	var ok bool

	expectedDeployments := rg.GetDeployments()
	for _, deploy := range expectedDeployments {
		receiver := appsv1.Deployment{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: deploy.GetName(), Namespace: deploy.GetNamespace()}, &receiver)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// if we got an IsNotFound error, we later make sure that ok is false.
				continue
			} else {
				reqLogger.Error(err, "unable to query for existing deployment")
				return nil, false
			}
		}

		foundDeployments = append(foundDeployments, &receiver)
	}

	// evaluate if we found what we expected
	ok = len(foundDeployments) == len(expectedDeployments)

	return foundDeployments, ok
}

// queryAPIForExpectedCRDs will check that the CRDs expected for a given instance actually
// exist in the API. will return ok as true when all were found, and false if not. Will return the
// CRD slice as nil if an error other than IsNotfound was encountered trying to obtain the data, as well as
// set return ok as false.
func queryAPIForExpectedCRDs(client client.Client, rg ResourceGetter, reqLogger logr.Logger) ([]*apiextv1.CustomResourceDefinition, bool) {
	foundCRDs := make([]*apiextv1.CustomResourceDefinition, 0)
	var ok bool

	expectedCRDs, err := rg.GetCRDs()
	if err != nil {
		// GetCRDs returns an error in case it can't read the CRD from the filesystem.
		// we have to handle it.
		return nil, false
	}

	for _, crd := range expectedCRDs {
		receiver := apiextv1.CustomResourceDefinition{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: crd.GetName()}, &receiver)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// if we got an IsNotFound error, we later make sure that ok is false.
				continue
			} else {
				reqLogger.Error(err, "unable to query for existing custom resource definitions")
				return nil, false
			}
		}

		foundCRDs = append(foundCRDs, &receiver)
	}

	// evaluate if we found what we expected
	ok = len(foundCRDs) == len(expectedCRDs)

	return foundCRDs, ok
}

// reconcileStatusPhase will update the status phase indicator based on the discovered status of deployments
// and CRDs. This must run after DeploymentsHealthy and CRDsHealthy have been updated by the status reconciler.
func (r *CertManagerDeploymentReconciler) reconcileStatusPhase(inStatus *operatorsv1alpha1.CertManagerDeploymentStatus) *operatorsv1alpha1.CertManagerDeploymentStatus {
	var crdsHealthy bool
	var deploymentsHealthy bool
	cmap := conditionsAsMap(inStatus.Conditions)

	// query for the conditions and parse the evaluative state
	if condition, ok := cmap[operatorsv1alpha1.ConditionCRDsAreReady]; ok {
		if condition.Status == corev1.ConditionUnknown || condition.Status == corev1.ConditionFalse {
			// for this check, we'll evaluate unknown as false so that we force the CR to reflect
			// a pending status
			crdsHealthy = false
		} else if condition.Status == corev1.ConditionTrue {
			// only other options.
			crdsHealthy = true
		}
	} else {
		// the condition was not found so we're going to report unhealthy
		crdsHealthy = false
	}

	if condition, ok := cmap[operatorsv1alpha1.ConditionDeploymentsAreReady]; ok {
		if condition.Status == corev1.ConditionUnknown || condition.Status == corev1.ConditionFalse {
			// for this check, we'll evaluate unknown as false so that we force the CR to reflect
			// a pending status
			deploymentsHealthy = false
		} else if condition.Status == corev1.ConditionTrue {
			// only other options.
			deploymentsHealthy = true
		}
	} else {
		// the condition was not found so we're going to report unhealthy
		deploymentsHealthy = false
	}

	if crdsHealthy && deploymentsHealthy {
		inStatus.Phase = componentry.StatusPhaseRunning
	} else if !crdsHealthy || !deploymentsHealthy {
		inStatus.Phase = componentry.StatusPhasePending
	}

	return inStatus
}

// updateStatusDeploymentConditions adds the conditions associated with found managed deployments to the status block for the CertManagerDeployment.
// This is an overwrite action and only stores conditions that are found in the API server. Deployments that are not in the APIServer when this is
// executed will not be stored, regardless of whether they should exist. This can indicate an issue reaching a running phase for the CertManagerDeployment.
func updateStatusDeploymentConditions(inStatus *operatorsv1alpha1.CertManagerDeploymentStatus, existingDeploys []*appsv1.Deployment) *operatorsv1alpha1.CertManagerDeploymentStatus {
	conditions := make([]operatorsv1alpha1.ManagedDeploymentWithConditions, 0)
	for _, deploy := range existingDeploys {
		// we could use a NamespacedName struct here but there's no need at the moment, we just want the format.
		// so the user knows what exactly is being stored here.
		deployCondition := operatorsv1alpha1.ManagedDeploymentWithConditions{
			NamespacedName: fmt.Sprintf("%s%c%s", deploy.GetNamespace(), '/', deploy.GetName()),
			Conditions:     deploy.Status.Conditions,
		}

		conditions = append(conditions, deployCondition)
	}

	inStatus.DeploymentConditions = conditions
	return inStatus
}

// updateStatusCRDConditions adds the conditions associated with found managed CRDs to the status block for the CertManagerDeployment.
// This is an overwrite action and only stores conditions that are found in the API server. CRDs that are not in the APIServer when this is
// executed will not be stored, regardless of whether they should exist. This can indicate an issue reaching a running phase for the CertManagerDeployment.
func updateStatusCRDConditions(inStatus *operatorsv1alpha1.CertManagerDeploymentStatus, existingCRDs []*apiextv1.CustomResourceDefinition) *operatorsv1alpha1.CertManagerDeploymentStatus {
	conditions := make([]operatorsv1alpha1.ManagedCRDWithConditions, 0)
	for _, crd := range existingCRDs {
		crdCondition := operatorsv1alpha1.ManagedCRDWithConditions{
			Name:       crd.GetName(),
			Conditions: crd.Status.Conditions,
		}

		conditions = append(conditions, crdCondition)
	}

	inStatus.CRDConditions = conditions
	return inStatus
}

// conditionsAsMap takes in a slice of conditions for the CertManagerDeployment resource and returns a map where the key is
// the condition.Type and the value is the condition itself.
func conditionsAsMap(conditions []operatorsv1alpha1.CertManagerDeploymentCondition) map[operatorsv1alpha1.CertManagerDeploymentConditionType]operatorsv1alpha1.CertManagerDeploymentCondition {
	c := make(map[operatorsv1alpha1.CertManagerDeploymentConditionType]operatorsv1alpha1.CertManagerDeploymentCondition, 0)

	for _, cond := range conditions {
		c[cond.Type] = cond
	}

	return c
}

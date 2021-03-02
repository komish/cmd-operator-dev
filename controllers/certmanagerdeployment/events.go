package certmanagerdeployment

import (
	"context"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileRoles will reconcile the Role resources for a given CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileRoles(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: roles")
	defer reqLogger.Info("Ending reconciliation: roles")
	var err error

	getter := ResourceGetter{CustomResource: *instance}
	roles := getter.GetRoles()

	for _, role := range roles {
		if err := controllerutil.SetControllerReference(instance, role, r.Scheme); err != nil {
			return err
		}
		found := &rbacv1.Role{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: role.GetName(), Namespace: role.GetNamespace()}, found)

		if err != nil && apierrors.IsNotFound(err) {
			reqLogger.Info("Creating role", "Role.Namespace", role.Namespace, "Role.Name", role.Name)
			r.Eventf(instance,
				createManagedRole.etype,
				createManagedRole.reason,
				"%s: %s/%s",
				createManagedRole.message,
				role.GetNamespace(),
				role.GetName())
			if err = r.Create(context.TODO(), role); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			return err
		}

		// A role exists. Determine if it needs updating and do so.
		genRulesInterface, err := cmdoputils.Interfacer{Data: role.Rules}.ToJSONInterface()
		if err != nil { // errors indicate a marshaling problem.
			return err
		}
		foundRulesInterface, err := cmdoputils.Interfacer{Data: found.Rules}.ToJSONInterface()
		if err != nil {
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: role.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		rulesMatch := cmdoputils.ObjectsMatch(genRulesInterface, foundRulesInterface)
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)

		if !(rulesMatch && labelsMatch) {
			reqLogger.Info("Role already exists, but needs an update.",
				"Role.Name", role.GetName(),
				"Role.Namespace", role.GetNamespace(),
				"HasExpectedRules", rulesMatch,
				"HasExpectedLabels", labelsMatch)
			r.Eventf(instance,
				updatingManagedRole.etype,
				updatingManagedRole.reason,
				"%s: %s/%s",
				updatingManagedRole.message,
				role.GetNamespace(),
				role.GetName())

			updated := found.DeepCopy()

			if !rulesMatch {
				updated.Rules = role.Rules
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = role.GetLabels()
			}

			reqLogger.Info("Updating Role.", "Role.Name", role.GetName(), "Role.Namespace", role.GetNamespace())
			if err := r.Update(context.TODO(), updated); err != nil {
				return err
			}

			r.Eventf(instance, updatedManagedRole.etype, updatedManagedRole.reason, "%s: %s/%s", updatedManagedRole.message, role.GetNamespace(), role.GetName())
		}
	}

	return err
}

// Event is a helper type for controller event logging.
type Event struct {
	etype   string
	reason  string
	message string
}

const (
	// EventTypeNormal is an informational event.
	EventTypeNormal = corev1.EventTypeNormal
	// EventTypeWarning is a warning event indicating something may go wrong.
	EventTypeWarning = corev1.EventTypeWarning
	// EventTypeError indicates that something has gone wrong.
	// TODO(): if this doesn't get used, remove and revert to corev1.EventTypes only
	EventTypeError = "Error"
)

var (
	// createManagedDeployment is an event indicating that a deployment is being created.
	createManagedDeployment = Event{
		etype:   EventTypeNormal,
		reason:  "CreatingDeployment",
		message: "Deployment does not exist and needs to be created",
	}

	// updateManagedDeployment is an event indicating that a deployment is being updated.
	updatingManagedDeployment = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatingDeployment",
		message: "Deployment exists but does not match desired state and needs updating",
	}

	// updateManagedDeployment is an event indicating that a deployment has been updated.
	updatedManagedDeployment = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatedDeployment",
		message: "Deployment has been successfully updated",
	}

	// createManagedCRD is an event indicating that a CRD is being created.
	createManagedCRD = Event{
		etype:   EventTypeNormal,
		reason:  "CreatingCRD",
		message: "CRD does not exist and needs to be created",
	}

	// updateManagedCRD is an event indicating that a CRD is being updated.
	updatingManagedCRD = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatingCRD",
		message: "CRD exists but does not match desired state and needs updating",
	}

	// updateManagedCRD is an event indicating that a CRD has been updated.
	updatedManagedCRD = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatedCRD",
		message: "CRD has been successfully updated",
	}

	// createManagedNamespace is an event indicating that a namespace is being created.
	createManagedNamespace = Event{
		etype:   EventTypeNormal,
		reason:  "CreatingNamespace",
		message: "Namespace does not exist and needs to be created",
	}

	// createManagedRole is an event indicating that a roleis being created
	createManagedRole = Event{
		etype:   EventTypeNormal,
		reason:  "CreatingRole",
		message: "Role does not exist and needs to be created",
	}

	// updatingManagedRole is an event indicating that a role needs updating
	updatingManagedRole = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatingRole",
		message: "Role exists but does not match desired state and needs updating",
	}

	// updatedManagedrole is an event indicating that a role has been updated
	updatedManagedRole = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatedRole",
		message: "Role has been successfully updated",
	}

	// createManagedRoleBinding is an event indicating that a rolebinding is being created
	createManagedRoleBinding = Event{
		etype:   EventTypeNormal,
		reason:  "CreatingRoleBinding",
		message: "RoleBinding does not exist and needs to be created",
	}

	// updatingManagedRoleBinding is an event indicating that a rolebinding needs updating
	updatingManagedRoleBinding = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatingRoleBinding",
		message: "RoleBinding exists but does not match desired state and needs updating",
	}

	// updatedManagedRole Binding is an event indicating that a rolebinding has been updated
	updatedManagedRoleBinding = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatedRoleBinding",
		message: "RoleBinding has been successfully updated",
	}

	// createManagedClusterRnole is an event indicating that a cluster role is being created
	createManagedClusterRole = Event{
		etype:   EventTypeNormal,
		reason:  "CreatingClusterRole",
		message: "Cluster role does not exist and needs to be created",
	}

	// updatingManagedClusterRole is an event indicating that a role needs updating
	updatingManagedClusterRole = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatingClusterRole",
		message: "Cluster role exists but does not match desired state and needs updating",
	}

	// updatedManagedClusterRole is an event indicating that a role has been updated
	updatedManagedClusterRole = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatedClusterRole",
		message: "Cluster role has been successfully updated",
	}

	// createManagedClusterRoleBinding is an event indicating that a cluster rolebinding is being created
	createManagedClusterRoleBinding = Event{
		etype:   EventTypeNormal,
		reason:  "CreatingClusterRoleBinding",
		message: "Cluster rolebinding does not exist and needs to be created",
	}

	// updatedManagedClusterRoleBinding is an event indicating that a cluster rolebinding needs updating
	updatingManagedClusterRoleBinding = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatingClusterRoleBinding",
		message: "ClusterRoleBinding exists but does not match desired state and needs updating",
	}

	// updatedManagedClusterRoleBinding Binding is an event indicating that a cluster rolebinding has been updated
	updatedManagedClusterRoleBinding = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatedClusterRoleBinding",
		message: "ClusterRoleBinding has been successfully updated",
	}

	// createManagedServiceAccount is an event indicating that a service account is being created
	createManagedServiceAccount = Event{
		etype:   EventTypeNormal,
		reason:  "CreatingServiceAccount",
		message: "Service account does not exist and needs to be created",
	}

	// createManagedService is an event indicating that a service is being created
	createManagedService = Event{
		etype:   EventTypeNormal,
		reason:  "CreatingService",
		message: "Service does not exist and needs to be created",
	}

	// updateManagedService is an event indicating that a service is being updated
	updatingManagedService = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatingService",
		message: "Service exists but does not match desired state and needs updating",
	}

	// updateManagedService is an event indicating that a service has been updated
	updatedManagedService = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatedService",
		message: "Service has been successfully updated",
	}

	// createManagedWebhook is an event indicating that a webhook is being created
	createManagedWebhook = Event{
		etype:   EventTypeNormal,
		reason:  "CreatingWebhook",
		message: "Webhook does not exist and needs to be created",
	}
	// updatingManagedWebhook is an event indicating that a webhook is being updated
	updatingManagedWebhook = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatingWebhook",
		message: "Webhook exists but does not match desired state and needs updating",
	}
	// updatedManagedWebhook is an event indicating that a webhook has been updated.
	updatedManagedWebhook = Event{
		etype:   EventTypeNormal,
		reason:  "UpdatedWebhook",
		message: "Webhook has been successfully updated",
	}
)

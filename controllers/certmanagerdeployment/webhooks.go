package certmanagerdeployment

import (
	"context"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	adregv1 "k8s.io/api/admissionregistration/v1"
)

// reconcileWebhooks reconciles MutatingWebhookConfiguration and ValidatingWebhookConfiguration resource(s) for a given CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileWebhooks(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: webhooks")
	defer reqLogger.Info("Ending reconciliation: webhooks")

	getter := ResourceGetter{CustomResource: *instance}

	mwhs := getter.GetMutatingWebhooks()

	for _, mwh := range mwhs {
		if err := controllerutil.SetControllerReference(instance, mwh, r.Scheme); err != nil {
			return err
		}
		found := &adregv1.MutatingWebhookConfiguration{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: mwh.GetName()}, found)
		if err != nil && apierrors.IsNotFound(err) {
			reqLogger.Info("Creating MutatingWebhookConfiguration", "MutatingWebhookConfiguration.Name", mwh.GetName())
			r.Eventf(instance,
				createManagedWebhook.etype,
				createManagedWebhook.reason,
				"%s: %s",
				createManagedWebhook.message,
				mwh.GetName())
			if err := r.Create(context.TODO(), mwh); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			return err
		}

		// Update the mutating webhook if needed
		genLabelsInterface, err := cmdoputils.Interfacer{Data: mwh.Labels}.ToJSONInterface()
		if err != nil { // err indicates a marshaling problem
			return err
		}

		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		genAnnotsInterface, err := cmdoputils.Interfacer{Data: mwh.Annotations}.ToJSONInterface()
		if err != nil {
			return err
		}
		foundAnnotsInterface, err := cmdoputils.Interfacer{Data: found.Annotations}.ToJSONInterface()
		if err != nil {
			return err
		}

		genWebhooksInterface, err := cmdoputils.Interfacer{Data: mwh.Webhooks}.ToJSONInterface()
		if err != nil {
			return err
		}

		foundWebhooksInterface, err := cmdoputils.Interfacer{Data: found.Webhooks}.ToJSONInterface()
		if err != nil {
			return err
		}

		// Check for equality
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)
		annotsMatch := cmdoputils.ObjectsMatch(genAnnotsInterface, foundAnnotsInterface)
		webhooksMatch := cmdoputils.ObjectsMatch(genWebhooksInterface, foundWebhooksInterface)

		// If not equal, update.
		if !(webhooksMatch && labelsMatch && annotsMatch) {
			reqLogger.Info("MutatingWebhookConfiguration already exists, but needs an update. Updating.",
				"MutatingWebhookConfiguration.Name", mwh.GetName(),
				"HasExpectedWebhooks", webhooksMatch,
				"HasExpectedLabels", labelsMatch,
				"HasExpectedAnnotations", annotsMatch)
			r.Eventf(instance, updatingManagedWebhook.etype, updatingManagedWebhook.reason, "%s: %s", updatingManagedWebhook.message, mwh.GetName())

			updated := found.DeepCopy()

			if !webhooksMatch {
				// TODO(): should we avoid clobbering and instead just add our annotations?
				updated.Webhooks = mwh.Webhooks
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = mwh.GetLabels()
			}

			if !annotsMatch {
				// TODO(): should we avoid clobbering and instead just add our annotations?
				updated.ObjectMeta.Annotations = mwh.GetAnnotations()
			}

			reqLogger.Info("Updating MutatingWebhookConfiguration", "MutatingWebhookConfiguration.Name", mwh.GetName())
			if err := r.Update(context.TODO(), updated); err != nil {
				return err
			}

			r.Eventf(instance, updatedManagedWebhook.etype, updatedManagedWebhook.reason, "%s: %s", updatedManagedWebhook.message, mwh.GetName())
		}
	}

	// validating webhooks
	vwhs := getter.GetValidatingWebhooks()
	for _, vwh := range vwhs {
		if err := controllerutil.SetControllerReference(instance, vwh, r.Scheme); err != nil {
			return err
		}

		found := &adregv1.ValidatingWebhookConfiguration{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: vwh.GetName()}, found)
		if err != nil && apierrors.IsNotFound(err) {
			reqLogger.Info("Creating ValidatingWebhookConfiguration", "ValidatingWebhookConfiguration.Name", vwh.GetName())
			r.Eventf(instance,
				createManagedWebhook.etype,
				createManagedWebhook.reason,
				"%s: %s",
				createManagedWebhook.message,
				vwh.GetName())
			if err := r.Create(context.TODO(), vwh); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			return err
		}

		// Update the validating webhook if needed
		genLabelsInterface, err := cmdoputils.Interfacer{Data: vwh.Labels}.ToJSONInterface()
		if err != nil { // error indicates a marshaling problem
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		genAnnotsInterface, err := cmdoputils.Interfacer{Data: vwh.Annotations}.ToJSONInterface()
		if err != nil {
			return err
		}
		foundAnnotsInterface, err := cmdoputils.Interfacer{Data: found.Annotations}.ToJSONInterface()
		if err != nil {
			return err
		}

		genWebhooksInterface, err := cmdoputils.Interfacer{Data: vwh.Webhooks}.ToJSONInterface()
		if err != nil {
			return err
		}

		foundWebhooksInterface, err := cmdoputils.Interfacer{Data: found.Webhooks}.ToJSONInterface()
		if err != nil {
			return err
		}

		// Check for equality
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)
		annotsMatch := cmdoputils.ObjectsMatch(genAnnotsInterface, foundAnnotsInterface)
		webhooksMatch := cmdoputils.ObjectsMatch(genWebhooksInterface, foundWebhooksInterface)

		// If not equal, update.
		if !(webhooksMatch && labelsMatch && annotsMatch) {
			reqLogger.Info("ValidatingWebhookConfiguration already exists, but needs an update. Updating.",
				"ValidatingWebhookConfiguration.Name", vwh.GetName(),
				"HasExpectedWebhooks", webhooksMatch,
				"HasExpectedLabels", labelsMatch,
				"HasExpectedAnnotations", annotsMatch)
			r.Eventf(instance, updatingManagedWebhook.etype, updatingManagedWebhook.reason, "%s: %s", updatingManagedWebhook.message, vwh.GetName())

			updated := found.DeepCopy()

			// TODO(): should we avoid clobbering and
			// instead just merge these meta items
			if !webhooksMatch {
				updated.Webhooks = vwh.Webhooks
			}

			if !labelsMatch {
				updated.ObjectMeta.Labels = vwh.GetLabels()
			}

			if !annotsMatch {
				updated.ObjectMeta.Annotations = vwh.GetAnnotations()
			}

			reqLogger.Info("Updating ValidatingWebhookConfiguration", "ValidatingWebhookConfiguration.Name", vwh.GetName())
			if err := r.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.Eventf(instance, updatedManagedWebhook.etype, updatedManagedWebhook.reason, "%s: %s", updatedManagedWebhook.message, vwh.GetName())
		}
	}

	return nil
}

// GetMutatingWebhooks returns MutatingWebhookConfiguration objects for a given CertManagerDeployment
// custom resource.
func (r *ResourceGetter) GetMutatingWebhooks() []*adregv1.MutatingWebhookConfiguration {
	var hooks []*adregv1.MutatingWebhookConfiguration
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(
			cmdoputils.CRVersionOrDefaultVersion(
				r.CustomResource.Spec.Version,
				componentry.CertManagerDefaultVersion),
		)
		for _, webhookData := range component.GetWebhooks() {
			if webhookData.IsEmpty() {
				// The component doesn't have any webhooks
				return nil
			}

			mutateHooks := webhookData.GetMutatingWebhooks()
			if len(mutateHooks) == 0 {
				// The component doesn't have any mutating webhooks so it must
				// have validating webhooks only
				return nil
			}

			for _, validateHook := range mutateHooks {
				hooks = append(hooks, newMutatingWebhook(component, r.CustomResource, webhookData.GetName(), validateHook, webhookData.GetAnnotations()))
			}
		}

	}

	return hooks
}

// GetValidatingWebhooks returns ValidatingWebhookConfiguration objects for a given CertManagerDeployment
// custom resource.
func (r *ResourceGetter) GetValidatingWebhooks() []*adregv1.ValidatingWebhookConfiguration {
	var hooks []*adregv1.ValidatingWebhookConfiguration
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(
			cmdoputils.CRVersionOrDefaultVersion(
				r.CustomResource.Spec.Version,
				componentry.CertManagerDefaultVersion),
		)
		for _, webhookData := range component.GetWebhooks() {
			if webhookData.IsEmpty() {
				// The component doesn't have any webhooks
				return nil
			}

			validateHooks := webhookData.GetValidatingWebhooks()
			if len(validateHooks) == 0 {
				// The component doesn't have any validating webhooks so it must
				// have mutating webhooks only
				return nil
			}

			for _, validateHook := range validateHooks {
				hooks = append(hooks, newValidatingWebhook(component, r.CustomResource, webhookData.GetName(), validateHook, webhookData.GetAnnotations()))
			}
		}

	}

	return hooks
}

// newMutatingWebhook returns a Webhook object for a given CertManagerComponent
// and CertManagerDeployment CustomResource
func newMutatingWebhook(comp componentry.CertManagerComponent,
	cr operatorsv1alpha1.CertManagerDeployment,
	webhookName string,
	webhookConfig adregv1.MutatingWebhook,
	annotations map[string]string) *adregv1.MutatingWebhookConfiguration {

	hook := adregv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:        webhookName,
			Labels:      cr.GetLabels(),
			Annotations: annotations,
		},
		Webhooks: []adregv1.MutatingWebhook{},
	}

	webhookConfig.ClientConfig.Service.Namespace = componentry.CertManagerDeploymentNamespace
	hook.Webhooks = append(hook.Webhooks, webhookConfig)

	// if CR customizations, add here (skip currently)

	return &hook
}

// newValidatingWebhook returns a Webhook object for a given CertManagerComponent
// and CertManagerDeployment CustomResource
func newValidatingWebhook(comp componentry.CertManagerComponent, cr operatorsv1alpha1.CertManagerDeployment,
	webhookName string, webhookConfig adregv1.ValidatingWebhook, annotations map[string]string) *adregv1.ValidatingWebhookConfiguration {
	// get initial structure
	hook := adregv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:        webhookName,
			Labels:      cr.GetLabels(),
			Annotations: annotations,
		},
		Webhooks: []adregv1.ValidatingWebhook{},
	}

	webhookConfig.ClientConfig.Service.Namespace = componentry.CertManagerDeploymentNamespace
	hook.Webhooks = append(hook.Webhooks, webhookConfig)

	// if CR customizations, add here (skip currently)

	return &hook
}

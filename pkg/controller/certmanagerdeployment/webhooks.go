package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/cmdoputils"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/componentry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	adregv1beta1 "k8s.io/api/admissionregistration/v1beta1"
)

// GetMutatingWebhooks returns MutatingWebhookConfiguration objects for a given CertManagerDeployment
// custom resource.
func (r *ResourceGetter) GetMutatingWebhooks() []*adregv1beta1.MutatingWebhookConfiguration {
	var hooks []*adregv1beta1.MutatingWebhookConfiguration
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
				hooks = append(hooks, newMutatingWebhook(component, r.CustomResource, webhookData.GetName(), validateHook))
			}
		}

	}

	return hooks
}

// GetValidatingWebhooks returns ValidatingWebhookConfiguration objects for a given CertManagerDeployment
// custom resource.
func (r *ResourceGetter) GetValidatingWebhooks() []*adregv1beta1.ValidatingWebhookConfiguration {
	var hooks []*adregv1beta1.ValidatingWebhookConfiguration
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
				// The componet doesn't have any validating webhooks so it must
				// have mutating webhooks only
				return nil
			}

			for _, validateHook := range validateHooks {
				hooks = append(hooks, newValidatingWebhook(component, r.CustomResource, webhookData.GetName(), validateHook))
			}
		}

	}

	return hooks
}

// newMutatingWebhook returns a Webhook object for a given CertManagerComponent
// and CertManagerDeployment CustomResource
func newMutatingWebhook(comp componentry.CertManagerComponent, cr redhatv1alpha1.CertManagerDeployment,
	webhookName string, webhookConfig adregv1beta1.MutatingWebhook) *adregv1beta1.MutatingWebhookConfiguration {
	// get initial structure
	hook := adregv1beta1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webhookName,
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    cr.GetLabels(),
		},
		Webhooks: []adregv1beta1.MutatingWebhook{},
	}

	// set clientConfig namespace value that is placeheld
	// TODO(): Is there a reliable way to extract the service reference here
	// based on the component associated with this webhook? Currently this matches
	// GetResourceName() but do we want to use that in all cases?
	webhookConfig.ClientConfig.Service.Namespace = componentry.CertManagerDeploymentNamespace
	hook.Webhooks = append(hook.Webhooks, webhookConfig)

	// if CR customizations, add here (skip currently)

	return &hook
}

// newValidatingWebhook returns a Webhook object for a given CertManagerComponent
// and CertManagerDeployment CustomResource
func newValidatingWebhook(comp componentry.CertManagerComponent, cr redhatv1alpha1.CertManagerDeployment,
	webhookName string, webhookConfig adregv1beta1.ValidatingWebhook) *adregv1beta1.ValidatingWebhookConfiguration {
	// get initial structure
	hook := adregv1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webhookName,
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    cr.GetLabels(),
		},
		Webhooks: []adregv1beta1.ValidatingWebhook{},
	}

	// set clientConfig namespace value that is placeheld
	// TODO(): Is there a reliable way to extract the service reference here
	// based on the component associated with this webhook? Currently this matches
	// GetResourceName() but do we want to use that in all cases?
	webhookConfig.ClientConfig.Service.Namespace = componentry.CertManagerDeploymentNamespace

	// if CR customizations, add here (skip currently)

	return &hook
}

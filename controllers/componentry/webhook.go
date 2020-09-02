package componentry

import (
	adregv1 "k8s.io/api/admissionregistration/v1"
)

// WebhookData contains the webhook mappings for a given CertManagerComponent.
type WebhookData struct {
	name               string
	annotations        map[string]string
	mutatingWebhooks   []adregv1.MutatingWebhook
	validatingWebhooks []adregv1.ValidatingWebhook
}

// GetName returns the name of the webhook being described by the WebhookData object.
func (wd *WebhookData) GetName() string {
	return wd.name
}

// GetAnnotations returns a string map of annotations that describe objects in the WebhookData
func (wd *WebhookData) GetAnnotations() map[string]string {
	copy := make(map[string]string)
	for key, value := range wd.annotations {
		copy[key] = value
	}
	return copy
}

// GetMutatingWebhooks returns the mutating webhooks for the WebhookData object.
func (wd *WebhookData) GetMutatingWebhooks() []adregv1.MutatingWebhook {
	return wd.mutatingWebhooks
}

// GetValidatingWebhooks returns validating webhooks for the WebhookData object.
func (wd *WebhookData) GetValidatingWebhooks() []adregv1.ValidatingWebhook {
	return wd.validatingWebhooks
}

// IsEmpty returns true if all of struct keys in WebhookData contain respective zero values.
func (wd *WebhookData) IsEmpty() bool {
	return wd.nameIsEmpty() && wd.mutatingbWebhookIsEmpty() && wd.validatingWebhookIsEmpty()
}

// nameIsEmpty returns true if the name of the webhook is an empty string.
func (wd *WebhookData) nameIsEmpty() bool {
	return wd.name == ""
}

// mutatingWebhookIsEmpty returns true if there are no mutating webhooks in the WebhookData struct.
func (wd *WebhookData) mutatingbWebhookIsEmpty() bool {
	return len(wd.mutatingWebhooks) == 0
}

// validatingWebhookIsEmpty returns true if there are no validating webhooks in the WebhookData struct.
func (wd *WebhookData) validatingWebhookIsEmpty() bool {
	return len(wd.validatingWebhooks) == 0
}

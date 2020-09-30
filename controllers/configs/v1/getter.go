package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	controller = "controller"
	webhook    = "webhook"
	cainjector = "cainjector"
)

// GetEmptyConfigFor gives you the empty configuration object of the component specified.
func GetEmptyConfigFor(componentName string) runtime.Object {
	switch componentName {
	case controller:
		return &CertManagerControllerConfig{}
	case webhook:
		return &CertManagerWebhookConfig{}
	case cainjector:
		return &CertManagerCAInjectorConfig{}
	}

	return nil
}

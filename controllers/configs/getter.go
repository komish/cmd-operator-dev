package configs

import (
	"fmt"

	v1_0_0defaults "github.com/komish/cmd-operator-dev/controllers/configs/v1_0_0/defaults"
	v1_0_0types "github.com/komish/cmd-operator-dev/controllers/configs/v1_0_0/types"

	v1_1_0defaults "github.com/komish/cmd-operator-dev/controllers/configs/v1_1_0/defaults"
	v1_1_0types "github.com/komish/cmd-operator-dev/controllers/configs/v1_1_0/types"

	"k8s.io/apimachinery/pkg/runtime"
)

var (
	controller = "controller"
	webhook    = "webhook"
	cainjector = "cainjector"
)

// GetDefaultConfigFor will return a default config in a byte slice of yaml for the component
// at the specified version. This function will return a panic if an incorrect component name is provided.
func GetDefaultConfigFor(componentName, version string) []byte {
	switch componentName {
	case controller:
		return getDefaultControllerConfigForVersion(version)
	case webhook:
		return getDefaultWebhookConfigForVersion(version)
	case cainjector:
		return getDefaultCAInjectorConfigForVersion(version)
	default:
		panic(fmt.Sprintf("should have received a valid component string of options: controller, webhook, cainjector but received: %s\n", componentName))
	}
}

func getDefaultControllerConfigForVersion(version string) []byte {
	switch version {
	case "v1.1.0":
		return v1_1_0defaults.ConfigForController()
	case "v1.0.0", "v1.0.1", "v1.0.2", "v1.0.3", "v1.0.4":
		return v1_0_0defaults.ConfigForController()
	default:
		panic(fmt.Sprintf("should not have received version string that was not a supported version but received: %s\n", version))
	}
}

func getDefaultWebhookConfigForVersion(version string) []byte {
	switch version {
	case "v1.1.0":
		return v1_1_0defaults.ConfigForWebhook()
	case "v1.0.0", "v1.0.1", "v1.0.2", "v1.0.3", "v1.0.4":
		return v1_0_0defaults.ConfigForWebhook()
	default:
		panic(fmt.Sprintf("should not have received version string that was not a supported version but received: %s\n", version))
	}
}

func getDefaultCAInjectorConfigForVersion(version string) []byte {
	switch version {
	case "v1.1.0":
		return v1_1_0defaults.ConfigForCAInjector()
	case "v1.0.0", "v1.0.1", "v1.0.2", "v1.0.3", "v1.0.4":
		return v1_0_0defaults.ConfigForCAInjector()
	default:
		panic(fmt.Sprintf("should not have received version string that was not a supported version but received: %s\n", version))
	}
}

// GetEmptyConfigFor gives you the empty configuration object of the component at the
// specified version. This function will return a panic if an incorrect component name is provided.
func GetEmptyConfigFor(componentName, version string) runtime.Object {
	switch componentName {
	case controller:
		return getEmptyControllerConfigForVersion(version)
	case webhook:
		return getEmptyWebhookConfigForVersion(version)
	case cainjector:
		return getEmptyCAInjectorConfigForVersion(version)
	default:
		panic(fmt.Sprintf("should have received a valid component string of options: controller, webhook, cainjector but received: %s\n", componentName))
	}
}

func getEmptyControllerConfigForVersion(version string) runtime.Object {
	switch version {
	case "v1.1.0":
		return &v1_1_0types.CertManagerControllerConfig{}
	case "v1.0.0", "v1.0.1", "v1.0.2", "v1.0.3", "v1.0.4":
		return &v1_0_0types.CertManagerControllerConfig{}
	default:
		panic(fmt.Sprintf("should not have received version string that was not a supported version but received: %s\n", version))
	}

}

func getEmptyWebhookConfigForVersion(version string) runtime.Object {
	switch version {
	case "v1.1.0":
		return &v1_1_0types.CertManagerWebhookConfig{}
	case "v1.0.0", "v1.0.1", "v1.0.2", "v1.0.3", "v1.0.4":
		return &v1_0_0types.CertManagerWebhookConfig{}
	default:
		panic(fmt.Sprintf("should not have received version string that was not a supported version but received: %s\n", version))
	}
}

func getEmptyCAInjectorConfigForVersion(version string) runtime.Object {
	switch version {
	case "v1.1.0":
		return &v1_1_0types.CertManagerCAInjectorConfig{}
	case "v1.0.0", "v1.0.1", "v1.0.2", "v1.0.3", "v1.0.4":
		return &v1_0_0types.CertManagerCAInjectorConfig{}
	default:
		panic(fmt.Sprintf("should not have received version string that was not a supported version but received: %s\n", version))
	}
}

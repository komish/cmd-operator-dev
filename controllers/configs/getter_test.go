package configs

import (
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("GetDefaultConfigFor", func() {
	Context("When getting default configurations for valid components", func() {
		for _, component := range []string{controller, webhook, cainjector} {
			It("Should not panic", func() {
				Expect(func() { GetDefaultConfigFor(component, componentry.CertManagerDefaultVersion) }).ToNot(Panic())
			})

			It("Should not be empty", func() {
				config := GetDefaultConfigFor(component, componentry.CertManagerDefaultVersion)
				Expect(config).ToNot(BeEmpty())
			})
		}
	})

	Context("When getting default configurations with an invalid component", func() {
		invalid := "foo"
		It("Should Panic", func() {
			Expect(func() { GetDefaultConfigFor(invalid, componentry.CertManagerDefaultVersion) }).Should(Panic())
		})
	})
})

var _ = Describe("getDefaultControllerConfigForVersion", func() {
	Context("When getting default controller configurations for a given version of cert-manager", func() {
		It("Should not panic when passed a valid version", func() {
			Expect(func() {
				getDefaultControllerConfigForVersion(componentry.CertManagerDefaultVersion)
			}).ShouldNot(Panic())
		})
		It("Should panic when passed an invalid version", func() {
			invalid := "v0.0.0"
			Expect(func() {
				getDefaultControllerConfigForVersion(invalid)
			}).Should(Panic())
		})
	})
})

var _ = Describe("getDefaultWebhookConfigForVersion", func() {
	Context("When getting default webhook configurations for a given version of cert-manager", func() {
		It("Should not panic when passed a valid version", func() {
			Expect(func() {
				getDefaultWebhookConfigForVersion(componentry.CertManagerDefaultVersion)
			}).ShouldNot(Panic())
		})
		It("Should panic when passed an invalid version", func() {
			invalid := "v0.0.0"
			Expect(func() {
				getDefaultWebhookConfigForVersion(invalid)
			}).Should(Panic())
		})
	})
})

var _ = Describe("getDefaultCAInjectorConfigForVersion", func() {
	Context("When getting default cainjector configurations for a given version of cert-manager", func() {
		It("Should not panic when passed a valid version", func() {
			Expect(func() {
				getDefaultCAInjectorConfigForVersion(componentry.CertManagerDefaultVersion)
			}).ShouldNot(Panic())
		})
		It("Should panic when passed an invalid version", func() {
			invalid := "v0.0.0"
			Expect(func() {
				getDefaultCAInjectorConfigForVersion(invalid)
			}).Should(Panic())
		})
	})
})

var _ = Describe("GetEmptyConfigFor", func() {
	Context("When getting an empty configurations for valid components", func() {
		for _, component := range []string{controller, webhook, cainjector} {
			It("Should not panic", func() {
				Expect(func() { GetEmptyConfigFor(component, componentry.CertManagerDefaultVersion) }).ToNot(Panic())
			})

			It("Should fulfill the runtime.Object interface", func() {
				config := GetEmptyConfigFor(component, componentry.CertManagerDefaultVersion)
				_, ok := config.(runtime.Object)
				Expect(ok).To(BeTrue())
			})
		}
	})

	Context("When getting empty configurations with an invalid component", func() {
		invalid := "foo"
		It("Should Panic", func() {
			Expect(func() { GetDefaultConfigFor(invalid, componentry.CertManagerDefaultVersion) }).Should(Panic())
		})
	})
})

var _ = Describe("getEmptyControllerConfigForVersion", func() {
	Context("When getting empty controller configurations for a given version of cert-manager", func() {
		It("Should not panic when passed a valid version", func() {
			Expect(func() {
				getEmptyControllerConfigForVersion(componentry.CertManagerDefaultVersion)
			}).ShouldNot(Panic())
		})
		It("Should panic when passed an invalid version", func() {
			invalid := "v0.0.0"
			Expect(func() {
				getEmptyControllerConfigForVersion(invalid)
			}).Should(Panic())
		})
	})
})

var _ = Describe("getEmptyWebhookConfigForVersion", func() {
	Context("When getting empty webhook configurations for a given version of cert-manager", func() {
		It("Should not panic when passed a valid version", func() {
			Expect(func() {
				getEmptyWebhookConfigForVersion(componentry.CertManagerDefaultVersion)
			}).ShouldNot(Panic())
		})
		It("Should panic when passed an invalid version", func() {
			invalid := "v0.0.0"
			Expect(func() {
				getEmptyWebhookConfigForVersion(invalid)
			}).Should(Panic())
		})
	})
})

var _ = Describe("getEmptyCAInjectorConfigForVersion", func() {
	Context("When getting empty cainjector configurations for a given version of cert-manager", func() {
		It("Should not panic when passed a valid version", func() {
			Expect(func() {
				getEmptyCAInjectorConfigForVersion(componentry.CertManagerDefaultVersion)
			}).ShouldNot(Panic())
		})
		It("Should panic when passed an invalid version", func() {
			invalid := "v0.0.0"
			Expect(func() {
				getEmptyCAInjectorConfigForVersion(invalid)
			}).Should(Panic())
		})
	})
})

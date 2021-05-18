package defaults

// ConfigForController returns a default config for the controller component as a byte slice of YAML.
func ConfigForController() []byte {
	return []byte(`apiVersion: certmanagerconfigs.operators.opdev.io/v1
kind: CertManagerControllerConfig
flags:
  v: 2
  cluster-resource-namespace: $(POD_NAMESPACE)
  leader-election-namespace: $(POD_NAMESPACE)`)
}

// ConfigForWebhook returns a default config for the webhook component as a byte slice of YAML.
func ConfigForWebhook() []byte {
	return []byte(`apiVersion: certmanagerconfigs.operators.opdev.io/v1
kind: CertManagerWebhookConfig
flags:
  v: 2
  secure-port: 10250
  dynamic-serving-ca-secret-namespace: $(POD_NAMESPACE)
  dynamic-serving-ca-secret-name: cert-manager-webhook-ca
  dynamic-serving-dns-names:
  - cert-manager-webhook
  - cert-manager-webhook.cert-manager
  - cert-manager-webhook.cert-manager.svc`)
}

// ConfigForCAInjector returns a default config for the webhook component as a byte slice of YAML.
func ConfigForCAInjector() []byte {
	return []byte(`apiVersion: certmanagerconfigs.operators.opdev.io/v1
kind: CertManagerCAInjectorConfig
flags:
  v: 2
  leader-election-namespace: $(POD_NAMESPACE)`)
}

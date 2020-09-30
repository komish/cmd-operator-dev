package v1

var (
	// DefaultConfigsFor is a map of default configs for the various components. This is a map
	// because the existing workflows all query for information based on a component name string
	// (e.g. comp.GetName()). Default Configs are byte that look like runtime.Object and are defined
	// in this package.
	DefaultConfigsFor = map[string][]byte{
		controller: []byte(`apiVersion: certmanagerconfigs.operators.redhat.io
		kind: CertManagerControllerConfig
		flags:
			v: 2
			cluster-resource-namespace: $(POD_NAMESPACE)
			leader-election-namespace: $(POD_NAMESPACE)
		`),

		webhook: []byte(`apiVersion: certmanagerconfigs.operators.redhat.io
		kind: CertManagerWebhookConfig
		flags:
		  v: 2
		  secure-port: 10250
		  dynamic-serving-ca-secret-namespace: $(POD_NAMESPACE)
		  dynamic-serving-ca-secret-name: cert-manager-webhook-ca
		  dynamic-serving-dns-names: 
			- cert-manager-webhook
			- cert-manager-webhook.cert-manager
			- cert-manager-webhook.cert-manager.svc
		`),

		cainjector: []byte(`apiVersion: certmanagerconfigs.operators.redhat.io
		kind: CertManagerCAInjectorConfig
		flags:
		  v: 2
		  leader-election-namespace: $(POD_NAMESPACE)
		`),
	}
)

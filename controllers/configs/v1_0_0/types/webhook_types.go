package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CertManagerWebhookConfig struct {
	metav1.TypeMeta `json:",inline"`

	Flags CertManagerWebhookFlags `json:"flags"`
}

type CertManagerWebhookFlags struct {
	loggingFlags
	ListenPort                      int      `json:"secure-port"`
	HealthzPort                     int      `json:"healthz-port"`
	TLSCertFile                     string   `json:"tls-cert-file"`
	TLSKeyFile                      string   `json:"tls-private-key-file"`
	DynamicServingCASecretNamespace string   `json:"dynamic-serving-ca-secret-namespace"`
	DynamicServingCASecretName      string   `json:"dynamic-serving-ca-secret-name"`
	DynamicServingDNSNames          []string `json:"dynamic-serving-dns-names"`
	Kubeconfig                      string   `json:"kubeconfig"`
	TLSCipherSuites                 []string `json:"tls-cipher-suites"` // https://github.com/kubernetes/component-base/blob/master/cli/flag/ciphersuites_flag.go
	MinTLSVersion                   string   `json:"tls-min-version"`
}

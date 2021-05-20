package types

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CertManagerControllerConfig struct {
	metav1.TypeMeta `json:",inline"`

	Flags CertManagerControllerFlags `json:"flags"`
}

type CertManagerControllerFlags struct {
	loggingFlags                                         // most of these come from klog
	Kubeconfig                             string        `json:"kubeconfig"`
	Master                                 string        `json:"master"`
	ACMEHTTP01SolverImage                  string        `json:"acme-http01-solver-image"`
	ACMEHTTP01SolveCPUResourceLimits       string        `json:"acme-http01-solver-resource-limits-cpu"`
	ACMEHTTP01SolverMemoryResourceLimits   string        `json:"acme-http01-solver-resource-limits-memory"`
	ACMEHTTP01SolverCPURequestRequests     string        `json:"acme-http01-solver-resource-request-cpu"`
	ACMEHTTP01SolverMemoryResourceRequests string        `json:"acme-http01-solver-resource-request-memory"`
	AutoCertificateAnnotations             []string      `json:"auto-certificate-annotations"`
	ClusterIssuerAmbientCredentials        bool          `json:"cluster-issuer-ambient-credentials"`
	ClusterResourceNamespace               string        `json:"cluster-resource-namespace"`
	Controllers                            []string      `json:"controllers"`
	DefaultIssuerGroup                     string        `json:"default-issuer-group"`
	DefaultIssuerKind                      string        `json:"default-issuer-kind"`
	DefaultIssuerName                      string        `json:"default-issuer-name"`
	DNS01CheckRetryPeriod                  time.Duration `json:"dns01-check-retry-period"`
	DNS01RecursiveNameservers              []string      `json:"dns01-recursive-nameservers"`
	DNS01RecursiveNameserversOnly          bool          `json:"dns01-recursive-nameservers-only"`
	EnableCertificateOwnerRefs             bool          `json:"enable-certificate-owner-ref"`
	EnableProfiling                        bool          `json:"enable-profiling"`
	FeatureGates                           []string      `json:"feature-gates"` // supposed to be map[string]bool
	IssuerAmbientCredentials               bool          `json:"issuer-ambient-credentials"`
	KubeAPIBurst                           float64       `json:"kube-api-burst"`
	KubeAPIQPS                             float64       `json:"kube-api-qps"`
	LeaderElect                            bool          `json:"leader-elect"`
	LeaderElectLeaseDuration               time.Duration `json:"leader-election-lease-duration"`
	LeaderElectionNamespace                string        `json:"leader-election-namespace"`
	LeaderElectRenewDeadline               time.Duration `json:"leader-election-renew-deadline"`
	LeaderElectionRetryPeriod              time.Duration `json:"leader-election-retry-period"`
	MaxConcurrentChallenges                float64       `json:"max-concurrent-challenges"`
	MetricsListenAddress                   string        `json:"metrics-listen-address"`
	Namespace                              string        `json:"namespace"`
}

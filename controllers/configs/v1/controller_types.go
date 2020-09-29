package v1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CertManagerControllerConfig struct {
	metav1.TypeMeta `json:",inline"`

	Flags CertManagerControllerFlags `json:"flags"`
}

type CertManagerControllerFlags struct {
	// TODO can these strings be changed to actual resource request types?
	ACMEHTTP01SolverImage                  string          `json:"acme-http01-solver-image"`
	ACMEHTTP01SolveCPUResourceLimits       string          `json:"acme-http01-solver-resource-limits-cpu"`
	ACMEHTTP01SolverMemoryResourceLimits   string          `json:"acme-http01-solver-resource-limits-memory"`
	ACMEHTTP01SolverCPURequestRequests     string          `json:"acme-http01-solver-resource-request-cpu"`
	ACMEHTTP01SolverMemoryResourceRequests string          `json:"acme-http01-solver-resource-request-memory"`
	AddDirectoryHeaders                    bool            `json:"add_dir_header"`
	AlsoLogToSTDERR                        bool            `json:"alsologtostderr"`
	AutoCertificateAnnotations             []string        `json:"auto-certificate-annotations"`
	ClusterIssuerAmbientCredentials        bool            `json:"cluster-issuer-ambient-credentials"`
	ClusterResourceNamespace               string          `json:"cluster-resource-namespace"`
	Controllers                            []string        `json:"controllers"`
	DefaultIssuerGroup                     string          `json:"default-issuer-group"`
	DefaultIssuerKind                      string          `json:"default-issuer-kind string"`
	DefaultIssuerName                      string          `json:"default-issuer-name"`
	DNS01RecursiveNameservers              []string        `json:"dns01-recursive-nameservers"`
	DNS01RecursiveNameserversOnly          bool            `json:"dns01-recursive-nameservers-only"`
	EnableCertificateOwnerRefs             bool            `json:"enable-certificate-owner-ref"`
	FeatureGates                           map[string]bool `json:"feature-gates mapStringBool"`
	IssuerAmbientCredentials               bool            `json:"issuer-ambient-credentials"`
	Kubeconfig                             string          `json:"kubeconfig"`
	LeaderElect                            bool            `json:"leader-elect"`
	LeaderElectLeaseDuration               time.Duration   `json:"leader-election-lease-duration"`
	LeaderElectionNamespace                string          `json:"leader-election-namespace"`
	LeaderElectRenewDeadline               time.Duration   `json:"leader-election-renew-deadline"`
	LeaderElectionRetryPeriod              time.Duration   `json:"leader-election-retry-period"`
	LogFlushFrequency                      time.Duration   `json:"log-flush-frequency"`
	LogBacktraceAt                         traceLocation   `json:"log_backtrace_at"`
	LogDir                                 string          `json:"log_dir"`
	LogFile                                string          `json:"log_file"`
	LogFileMaxSize                         uint            `json:"log_file_max_size"`
	Master                                 string          `json:"master"`
	MaxConcurrentChallenges                int             `json:"max-concurrent-challenges"`
	MetricsListenAddress                   string          `json:"metrics-listen-address"`
	Namespace                              string          `json:"namespace"`
	RenewBeforeExpiryDuration              time.Duration   `json:"renew-before-expiry-duration duration"`
	SkipHeaders                            bool            `json:"skip_headers"`
	SkipLogHeaders                         bool            `json:"skip_log_headers"`
	STDERRThreshold                        severity        `json:"stderrthreshold"`
	VerbosityLevel                         klog.Level      `json:"v"`       // exported from klog
	VModule                                string          `json:"vmodule"` // accepting string because original type ModuleSpec is unexported from klog
}

// traceLocation is an emulation of the unexported struct traceLocation in klog https://github.com/kubernetes/klog/blob/master/klog.go#L325
type traceLocation struct {
	file string
	line int
}

// severity is an emulation of the unexported struct severity in klog https://github.com/kubernetes/klog/blob/master/klog.go#L98
type severity int32

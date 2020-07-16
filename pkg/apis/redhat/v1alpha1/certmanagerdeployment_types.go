package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CertManagerDeploymentSpec defines the desired state of CertManagerDeployment
type CertManagerDeploymentSpec struct {
	// Version indicates the version of CertManager to deploy. The operator only
	// supports a subset of versions.
	// +optional
	// +kubebuilder:validation:Enum=v0.14.3;v0.15.0;v0.15.1
	Version *string `json:"version"`
	// DangerZone contains a series of options that aren't necessarily accounted
	// for by the operator, but can be configured in edge cases if needed.
	// +optional
	DangerZone DangerZone `json:"dangerZone,omitempty"`
	// ImagePullPolicy is the policy to apply to all CertManagerComponent deployments.
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

// DangerZone is a set of configuration options that may cause the stability
// or reliability of the controller to break, but are exposed in case they
// need to be tweaked.
type DangerZone struct {
	// ImageOverrides is a map of CertManagerComponent names to image strings
	// in format /registry/image-name:tag. Valid keys are controller, webhook,
	// and cainjector.
	// +optional
	ImageOverrides map[string]string `json:"imageOverrides,omitempty"`
	// ContainerArgOverrides allows the full overriding of container arguments for
	// each component. These arguments must holistically cover what's needed for
	// the CertManagerComponent to run as it replaces the containers[].args key
	// in its entirety.
	// Omitting this results in the default container arguments the operator has
	// configured for each component.
	// +optional
	ContainerArgOverrides map[string][]string `json:"containerArgOverrides,omitempty"`
}

// CertManagerDeploymentStatus defines the observed state of CertManagerDeployment
type CertManagerDeploymentStatus struct {
	// Version is a status indicator showing the requested version of cert-manager deployed
	// by this CertManagerDeployment custom resource.
	Version string `json:"version,omitempty"`
	// Phase is a status indicator showing the state of the object and all downstream resources
	// it manages.
	Phase string `json:"phase,omitempty"`
	// ManagedDeploymentPhase is a status indicator showing the state of the deployments managed by
	// this custom resource.
	DeploymentsHealthy bool `json:"deploymentsHealthy,omitEmpty"`
	// ManagedCRDPhase is a status indicator showing the state of CRDs managed by this
	// custom resource.
	CRDsHealthy bool `json:"crdsHealthy,omitEmpty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertManagerDeployment is the Schema for the certmanagerdeployments API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=certmanagerdeployments,scope=Cluster
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.version`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="DeploymentsAreHealthy",type=string,JSONPath=`.status.deploymentsHealthy`
// +kubebuilder:printcolumn:name="CRDsAreHealthy",type=string,JSONPath=`.status.crdsHealthy`
type CertManagerDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertManagerDeploymentSpec   `json:"spec,omitempty"`
	Status CertManagerDeploymentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertManagerDeploymentList contains a list of CertManagerDeployment
type CertManagerDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertManagerDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertManagerDeployment{}, &CertManagerDeploymentList{})
}

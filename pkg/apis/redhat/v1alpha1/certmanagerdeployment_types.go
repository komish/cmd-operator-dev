package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CertManagerDeploymentSpec defines the desired state of CertManagerDeployment
type CertManagerDeploymentSpec struct {
	// Identifier is a string identifying a given CertManagerDeployment.
	Identifier string `json:"identifier"`
}

// CertManagerDeploymentStatus defines the observed state of CertManagerDeployment
type CertManagerDeploymentStatus struct {
	// DiscoveredIdentifier is the identifier used to initialize this resource
	// at initial creation.
	DiscoveredIdentifier string `json:"discoveredIdentifier"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertManagerDeployment is the Schema for the certmanagerdeployments API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=certmanagerdeployments,scope=Cluster
// +kubebuilder:storageversion
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

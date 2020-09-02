/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CertManagerDeploymentSpec defines the desired state of CertManagerDeployment
type CertManagerDeploymentSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Version indicates the version of CertManager to deploy. The operator only
	// supports a subset of versions.
	// +optional
	// +kubebuilder:validation:Enum=v0.15.0;v0.15.1;v0.15.2;v0.16.0;v1.0.0
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

// CertManagerDeploymentStatus defines the observed state of CertManagerDeployment
type CertManagerDeploymentStatus struct {
	// Conditions Represents the latest available observations of a CertManagerDeployment's current state.
	Conditions []CertManagerDeploymentCondition `json:"conditions,omitempty"`
	// Version is a status indicator showing the requested version of cert-manager deployed
	// by this CertManagerDeployment custom resource.
	Version string `json:"version,omitempty"`
	// Phase is a status indicator showing the state of the object and all downstream resources
	// it manages.
	Phase string `json:"phase,omitempty"`
	// DeploymentConditions is a report of conditions on owned deployments by this CertManagerDeployment.
	DeploymentConditions []ManagedDeploymentWithConditions `json:"deploymentConditions,omitEmpty"`
	// CRDConditions is a report of conditions on owned CRDs by this CertManagerDeployment.
	CRDConditions []ManagedCRDWithConditions `json:"crdConditions,omitEmpty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.version`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// CertManagerDeployment is the Schema for the certmanagerdeployments API
type CertManagerDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertManagerDeploymentSpec   `json:"spec,omitempty"`
	Status CertManagerDeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertManagerDeploymentList contains a list of CertManagerDeployment
type CertManagerDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertManagerDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertManagerDeployment{}, &CertManagerDeploymentList{})
}

// ManagedCRDWithConditions defines a deployment name and conditions associated with that CRD.
type ManagedCRDWithConditions struct {
	// Name is the name given to a specific CRD.
	Name string `json:"name"`
	// Conditions is the DeploymentConditions associated with that deployment.
	Conditions []apiextv1beta1.CustomResourceDefinitionCondition `json:"conditions"`
}

// CertManagerDeploymentCondition represents conditions that can be applied to a CertManagerDeployment object.
type CertManagerDeploymentCondition struct {
	// Type of certmanagerdeployment condition.
	Type CertManagerDeploymentConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,7,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// CertManagerDeploymentConditionType is the type of condition that is being reported on a CertManagerDeployment object.
type CertManagerDeploymentConditionType string

const (
	// ConditionCRDsAreReady indicates that the API contains CRDs that match the expected CRD object
	// that the operator would typically deploy, and that NamesAccepted and Established conditions
	// on each CRD are both true.
	ConditionCRDsAreReady CertManagerDeploymentConditionType = "CRDsAreReady"
	// ConditionDeploymentsAreReady indicates that the API contains deployments that match the expected
	// deployment names that the operator would typically deploy, and that the ready and available pods
	// match the desired count of pods.
	ConditionDeploymentsAreReady CertManagerDeploymentConditionType = "DeploymentsAreReady"
)

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

// ManagedDeploymentWithConditions defines a deployment namespaced name and conditions associated with that deployment.
type ManagedDeploymentWithConditions struct {
	// NamespacedName is the NamespacedName of the given deployment.
	NamespacedName string `json:"namespacedName"`
	// Conditions is the DeploymentConditions associated with that deployment.
	Conditions []appsv1.DeploymentCondition `json:"conditions"`
}

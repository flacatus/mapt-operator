/*
Copyright 2025.

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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KindPhase represents the lifecycle phase of a Kind resource.
type KindPhase string

const (
	KindPhasePending      KindPhase = "Pending"
	KindPhaseProvisioning KindPhase = "Provisioning"
	KindPhaseRunning      KindPhase = "Running"
	KindPhaseFailed       KindPhase = "Failed"
	KindPhaseDeleting     KindPhase = "Deleting"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KindSpec defines the desired state of Kind.
type KindSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// CloudConfig holds cloud provider and credential configurations.
	// +kubebuilder:validation:Required
	CloudConfig CloudConfig `json:"cloudConfig"`

	// MachineConfig defines the configuration for the EC2 spot machine.
	// +kubebuilder:validation:Required
	MachineConfig MachineConfig `json:"machineConfig"`

	// KindClusterConfig defines the configuration for the Kind cluster itself.
	// +kubebuilder:validation:Required
	KindClusterConfig KindClusterConfig `json:"kindClusterConfig"`

	// OutputKubeconfigSecretName defines the prefix for the name of the Kubernetes Secret
	// that will store the kubeconfig for the provisioned Kind cluster.
	// If not provided, a default prefix will be used (e.g., "kindspot-<random>-kubeconfig").
	// The final name is generated using this prefix via Kubernetes' `generateName`.
	// This also corresponds to the Tekton 'cluster-access-secret-name' param.
	// +optional
	OutputKubeconfigSecretName string `json:"outputKubeconfigSecretName,omitempty"`

	// TerminationPolicy defines when and how the cluster should be terminated.
	// +optional
	TerminationPolicy *TerminationPolicy `json:"terminationPolicy,omitempty"`
}

// KindClusterConfig contains parameters for the Kind cluster itself.
type KindClusterConfig struct {
	// KubernetesVersion specifies the Kubernetes version for the Kind cluster (e.g., "v1.29.2").
	// This field is required.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	KubernetesVersion string `json:"kubernetesVersion"`
}

// CloudConfig contains parameters to specify the cloud provider and access credentials.
// TODO: Add Azure machines support. Ussually Azure spot instances are more cheaper.
type CloudConfig struct {
	// Provider specifies the cloud provider name.
	// Currently, only "AWS" is supported. This field is designed for future extension.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=AWS
	// +kubebuilder:default=AWS
	Provider string `json:"provider"`

	// CredentialsSecretRef is a reference to a Kubernetes Secret in the same namespace
	// as the KindSpot resource. This Secret must contain all necessary cloud provider
	// credentials and configurations, including the region.
	// The required keys within the Secret depend on the specified 'Provider'.
	// For 'AWS', this Secret is expected to contain:
	//   - "access-key": Your AWS access key ID.
	//   - "secret-key": Your AWS secret access key.
	//   - "region": The AWS region (e.g., "us-east-1").
	//   - "bucket": The S3 bucket name (for the provisioning tool's backend state, if applicable).
	// +kubebuilder:validation:Required
	CredentialsSecretRef corev1.LocalObjectReference `json:"credentialsSecretRef"`
}

// TerminationPolicy defines automatic deletion parameters.
type TerminationPolicy struct {
	// DeleteAfterSeconds specifies a Time-To-Live (TTL) for the provisioned KindSpot.
	// After this duration (in seconds, starting from when the cluster becomes Ready or from creation),
	// the KindSpot and its underlying resources will be automatically destroyed.
	// This corresponds to the provisioning tool's '--timeout' parameter, which often expects a Go duration string.
	// The operator will convert these seconds into the required Go duration format for the tool.
	// +optional
	// +kubebuilder:validation:Minimum=60
	DeleteAfterSeconds *int64 `json:"deleteAfterSeconds,omitempty"`
}

// KindStatus defines the observed state of Kind.
type KindStatus struct {
	// Phase indicates the current lifecycle phase of the Kind cluster.
	// E.g., Pending, Provisioning, Ready, Deleting, Error.
	// +optional
	Phase KindPhase `json:"phase,omitempty"`

	// Message provides a human-readable status message.
	// +optional
	Message string `json:"message,omitempty"`

	// Conditions represent the latest available observations of the Kind cluster state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// AWSInstanceID is the ID of the EC2 instance provisioned with a Kind Cluster.
	// +optional
	AWSInstanceID *string `json:"awsInstanceID,omitempty"`

	// KubeconfigSecretName is the name of the Kubernetes Secret where the cluster's
	// kubeconfig has been stored. This will match `spec.outputKubeconfigSecretName` if provided,
	// or be a generated name.
	// +optional
	KubeconfigSecretName *string `json:"kubeconfigSecretName,omitempty"`

	// KindVersion is the actual Kubernetes version of the provisioned Kind cluster.
	// +optional
	KindVersion *string `json:"kindVersion,omitempty"`

	// ClusterReady indicates if the Kind cluster is fully provisioned and accessible.
	// +optional
	ClusterReady bool `json:"clusterReady,omitempty"`

	// AveragePrice reports the average acquisition price of the spot instance(s).
	// This field is a string to allow for currency.
	// +optional
	AveragePrice string `json:"averagePrice,omitempty"`

	// ExpirationTimestamp indicates when the cluster is scheduled to be terminated, based on TerminationPolicy.
	// +optional
	ExpirationTimestamp *metav1.Time `json:"expirationTimestamp,omitempty"`

	// ProvisionId is the id of the backend used by the Kind provisioning tool.
	ProvisionId *string `json:"provisionId,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Kind is the Schema for the kinds API.
type Kind struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KindSpec   `json:"spec,omitempty"`
	Status KindStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KindList contains a list of Kind.
type KindList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kind `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kind{}, &KindList{})
}

func (a *Kind) GetKindSecretName() string {
	if a.Spec.OutputKubeconfigSecretName != "" {
		return a.Spec.OutputKubeconfigSecretName
	}
	return fmt.Sprintf("kindspot-%s-kubeconfig", a.Name)
}

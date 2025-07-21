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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OpenshiftSncPhase represents the lifecycle phase of a OpenshiftSnc resource.
// +kubebuilder:validation:Enum=Pending;Provisioning;Running;Failed;Deleting
type OpenshiftSncPhase string

const (
	// OpenshiftSnc lifecycle phases
	OpenshiftSncPhasePending OpenshiftSncPhase = "Pending"
	// OpenshiftSncPhaseProvisioning indicates that the OpenshiftSnc cluster is being provisioned.
	// This phase is used when the cluster is in the process of being set up, including
	// provisioning the underlying infrastructure, installing the cluster components, etc.
	// It is a transient state that occurs after the initial request to create the cluster
	// and before it is fully operational.
	// This phase is particularly useful for tracking the progress of cluster creation
	OpenshiftSncPhaseProvisioning OpenshiftSncPhase = "Provisioning"
	// OpenshiftSncPhaseRunning indicates that the OpenshiftSnc cluster is fully operational and ready for use.
	// This phase is used when the cluster has been successfully provisioned, all components are running,
	// and it is ready to accept workloads. It signifies that the cluster is in a healthy state and can be interacted with.
	// This phase is typically reached after the Provisioning phase has completed successfully.
	// It is important for users to know when the cluster is ready for deployment of applications and services.
	OpenshiftSncPhaseRunning OpenshiftSncPhase = "Running"
	// OpenshiftSncPhaseFailed indicates that the OpenshiftSnc cluster failed to provision or encountered an error.
	// This phase is used when there was an issue during the provisioning process, such as infrastructure
	// failures, configuration errors, or other problems that prevent the cluster from being created successfully.
	OpenshiftSncPhaseFailed OpenshiftSncPhase = "Failed"
	// OpenshiftSncPhaseDeleting indicates that the OpenshiftSnc cluster is in the process of being deleted.
	// This phase is used when a request has been made to delete the cluster, and the
	// controller is actively working to clean up the resources associated with the cluster.
	// It signifies that the cluster is no longer available for use and that the deletion process is ongoing.
	OpenshiftSncPhaseDeleting OpenshiftSncPhase = "Deleting"
)

// OpenshiftSpec defines the desired state of Openshift.
type OpenshiftSpec struct {
	// OpenshiftClusterConfig defines the configuration for the Openshift cluster itself.
	// This includes the version of Openshift to install, networking settings, and other cluster-level configurations.
	OpenshiftClusterConfig OpenshiftClusterConfig `json:"openshiftClusterConfig"`

	// MachineConfig defines the configuration for the EC2 spot machine.
	// This includes the instance type, AMI, and other machine-level configurations.
	// This configuration is used to provision the underlying infrastructure for the Openshift cluster.
	// +kubebuilder:validation:Required
	MachineConfig MachineConfig `json:"machineConfig"`

	// TerminationPolicy defines the policy for terminating the Openshift cluster.
	TerminationPolicy TerminationPolicy `json:"terminationPolicy"`
}

type OpenshiftClusterConfig struct {
	// OpenshiftVersion specifies the version of Openshift to install.
	// This field is required to ensure that the correct version of Openshift is deployed.
	// It allows users to specify the desired version of Openshift for their cluster.
	OpenshiftVersion string `json:"openshiftVersion"`
}

// OpenshiftStatus defines the observed state of Openshift.
// OpenshiftStatus provides information about the current state of the Openshift cluster.
// It includes details such as the current phase of the cluster, conditions, and other relevant status information.
// This status is updated by the controller to reflect the latest state of the Openshift cluster.
// It is used to communicate the lifecycle status of the cluster to users and other components in the system.
// The status includes fields for phase, message, conditions, observed generation, and other relevant information
type OpenshiftStatus struct {
	// Phase indicates the current lifecycle phase of the Kind cluster.
	// This field is used to track the progress of the cluster through its lifecycle stages.
	// Possible values include:
	// - Pending: The cluster is being created.
	// - Provisioning: The cluster is being provisioned, including setting up the infrastructure and
	//   installing the necessary components.
	// - Running: The cluster is fully operational and ready for use
	// E.g., Pending, Provisioning, Ready, Deleting, Error.
	// +optional
	Phase OpenshiftSncPhase `json:"phase,omitempty"`

	// Message provides a human-readable status message.
	// This field is used to convey additional information about the current state of the cluster,
	// such as errors, warnings, or important updates.
	// It can be used to provide context about the cluster's status, such as whether it
	// is currently being provisioned, if there are any issues, or if it is ready for use.
	// +optional
	Message string `json:"message,omitempty"`

	// Conditions are used to provide detailed information about the state of the cluster,
	// including whether it is ready, if there are any issues, and other relevant status information
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// ObservedGeneration is the .metadata.generation that was last processed by the controller.
	// This field is used to track the version of the resource that the controller has observed.
	// It helps ensure that the controller is working with the most recent version of the resource
	// and can be used to detect changes in the resource that may require action.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// AWSInstanceID is the ID of the EC2 instance provisioned with a Kind Cluster.
	// This field is used to track the specific instance that is running the Kind cluster.
	// It is particularly useful for managing the lifecycle of the cluster and for debugging purposes.
	// +optional
	AWSInstanceID *string `json:"awsInstanceID,omitempty"` // NOTE: Consider replacing "Kind Cluster" with "Openshift cluster" for clarity.

	// KubeconfigSecretName is the name of the Kubernetes Secret where the cluster's
	// kubeconfig has been stored. This field is used to reference the secret that contains
	// the kubeconfig file for accessing the Openshift cluster.
	// +optional
	KubeconfigSecretName *string `json:"kubeconfigSecretName,omitempty"`

	// ClusterReady indicates if the Kind cluster is fully provisioned and accessible.
	// This field is used to determine if the cluster is ready for use, meaning that all
	// +optional
	ClusterReady bool `json:"clusterReady,omitempty"`

	// This field is used to provide information about the cost of the spot instances used
	// for the Openshift cluster. It can help users understand the financial implications of
	// using spot instances for their cluster.
	// +optional
	AveragePrice string `json:"averagePrice,omitempty"`

	// ProvisionStartTime records when the provisioning process began.
	// This field is used to track the start time of the provisioning process for the Openshift cluster.
	// +optional
	ProvisionStartTime *metav1.Time `json:"provisionStartTime,omitempty"`

	// LastUpdateTime records the last time the status was updated.
	// This field is used to track when the status of the Openshift cluster was last modified.
	// It helps ensure that users and other components can see the most recent status of the cluster.
	// This is particularly useful for monitoring and debugging purposes.
	// +optional
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	// ExpirationTimestamp indicates when the cluster is scheduled to be terminated, based on TerminationPolicy.
	// This field is used to specify when the Openshift cluster is expected to be terminated.
	// +optional
	ExpirationTimestamp *metav1.Time `json:"expirationTimestamp,omitempty"`

	// This field is used to track the specific provisioning session for the Openshift cluster.
	ProvisionId *string `json:"provisionId,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="Cluster phase"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.openshiftVersion`,description="Openshift version"
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.clusterReady`,description="Is the cluster ready?"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Openshift is the Schema for the openshifts API.
type Openshift struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenshiftSpec   `json:"spec,omitempty"`
	Status OpenshiftStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OpenshiftList contains a list of Openshift.
type OpenshiftList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Openshift `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Openshift{}, &OpenshiftList{})
}

func (a *Openshift) GetOpenshiftSncSecretName() string {
	return fmt.Sprintf("openshift-%s-kubeconfig", a.Name)
}

# API Reference

## Packages
- [mapt.redhat.com/v1alpha1](#maptredhatcomv1alpha1)


## mapt.redhat.com/v1alpha1

Package v1alpha1 contains API Schema definitions for the mapt v1alpha1 API group.

### Resource Types
- [Openshift](#openshift)
- [OpenshiftList](#openshiftlist)







#### MachineConfig



MachineConfig contains parameters for configuring the EC2 spot machine.



_Appears in:_
- [OpenshiftSpec](#openshiftspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `architecture` _string_ | Architecture for the EC2 instance. | x86_64 | Enum: [x86_64 arm64] <br /> |
| `cpus` _integer_ | CPUs is the number of vCPUs for the EC2 instance. | 8 |  |
| `gpu` _boolean_ | Indicates if the EC2 instance should have GPU support.<br />In case GPU is true, the instance type will be selected from the list of supported GPU instances. | false |  |
| `memoryGiB` _integer_ | MemoryGiB is the amount of RAM for the EC2 instance in GiB. | 16 |  |
| `nestedVirtualizationEnabled` _boolean_ | NestedVirtualizationEnabled specifies if the EC2 instance should have nested virtualization support. | false |  |
| `useSpotInstances` _boolean_ | UseSpotInstances specifies whether to use EC2 spot instances.<br />Corresponds to the Tekton 'spot' param. | true |  |
| `spotPriceIncreasePercentage` _integer_ | SpotPriceIncreasePercentage is the percentage to add on top of the current calculated spot price<br />to increase the chances of acquiring the machine. Only applies if UseSpotInstances is true.<br />A nil value means the underlying provisioning tool's default (if any) will be used, or the relevant command-line flag will be omitted. '0' is a valid percentage.<br />Corresponds to the Tekton 'spot-increase-rate' param (default '20'). |  |  |
| `tags` _object (keys:string, values:string)_ | Tags to apply to the AWS resources created by the provisioning tool.<br />The operator will convert this map into the string format the tool expects (e.g., "key1=value1,key2=value2").<br />Corresponds to the Tekton 'tags' param. |  |  |


#### Openshift



Openshift is the Schema for the openshifts API.



_Appears in:_
- [OpenshiftList](#openshiftlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `mapt.redhat.com/v1alpha1` | | |
| `kind` _string_ | `Openshift` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  |  |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  |  |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[OpenshiftSpec](#openshiftspec)_ |  |  |  |
| `status` _[OpenshiftStatus](#openshiftstatus)_ |  |  |  |


#### OpenshiftClusterConfig







_Appears in:_
- [OpenshiftSpec](#openshiftspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `openshiftVersion` _string_ | OpenshiftVersion specifies the version of Openshift to install.<br />This field is required to ensure that the correct version of Openshift is deployed.<br />It allows users to specify the desired version of Openshift for their cluster. |  |  |


#### OpenshiftList



OpenshiftList contains a list of Openshift.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `mapt.redhat.com/v1alpha1` | | |
| `kind` _string_ | `OpenshiftList` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  |  |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  |  |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[Openshift](#openshift) array_ |  |  |  |


#### OpenshiftSncPhase

_Underlying type:_ _string_

OpenshiftSncPhase represents the lifecycle phase of a OpenshiftSnc resource.

_Validation:_
- Enum: [Pending Provisioning Running Failed Deleting]

_Appears in:_
- [OpenshiftStatus](#openshiftstatus)

| Field | Description |
| --- | --- |
| `Pending` | OpenshiftSnc lifecycle phases<br /> |
| `Provisioning` | OpenshiftSncPhaseProvisioning indicates that the OpenshiftSnc cluster is being provisioned.<br />This phase is used when the cluster is in the process of being set up, including<br />provisioning the underlying infrastructure, installing the cluster components, etc.<br />It is a transient state that occurs after the initial request to create the cluster<br />and before it is fully operational.<br />This phase is particularly useful for tracking the progress of cluster creation<br /> |
| `Running` | OpenshiftSncPhaseRunning indicates that the OpenshiftSnc cluster is fully operational and ready for use.<br />This phase is used when the cluster has been successfully provisioned, all components are running,<br />and it is ready to accept workloads. It signifies that the cluster is in a healthy state and can be interacted with.<br />This phase is typically reached after the Provisioning phase has completed successfully.<br />It is important for users to know when the cluster is ready for deployment of applications and services.<br /> |
| `Failed` | OpenshiftSncPhaseFailed indicates that the OpenshiftSnc cluster failed to provision or encountered an error.<br />This phase is used when there was an issue during the provisioning process, such as infrastructure<br />failures, configuration errors, or other problems that prevent the cluster from being created successfully.<br /> |
| `Deleting` | OpenshiftSncPhaseDeleting indicates that the OpenshiftSnc cluster is in the process of being deleted.<br />This phase is used when a request has been made to delete the cluster, and the<br />controller is actively working to clean up the resources associated with the cluster.<br />It signifies that the cluster is no longer available for use and that the deletion process is ongoing.<br /> |


#### OpenshiftSpec



OpenshiftSpec defines the desired state of Openshift.



_Appears in:_
- [Openshift](#openshift)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `openshiftClusterConfig` _[OpenshiftClusterConfig](#openshiftclusterconfig)_ | OpenshiftClusterConfig defines the configuration for the Openshift cluster itself.<br />This includes the version of Openshift to install, networking settings, and other cluster-level configurations. |  |  |
| `machineConfig` _[MachineConfig](#machineconfig)_ | MachineConfig defines the configuration for the EC2 spot machine.<br />This includes the instance type, AMI, and other machine-level configurations.<br />This configuration is used to provision the underlying infrastructure for the Openshift cluster. |  | Required: \{\} <br /> |
| `terminationPolicy` _[TerminationPolicy](#terminationpolicy)_ | TerminationPolicy defines the policy for terminating the Openshift cluster. |  |  |


#### OpenshiftStatus



OpenshiftStatus defines the observed state of Openshift.
OpenshiftStatus provides information about the current state of the Openshift cluster.
It includes details such as the current phase of the cluster, conditions, and other relevant status information.
This status is updated by the controller to reflect the latest state of the Openshift cluster.
It is used to communicate the lifecycle status of the cluster to users and other components in the system.
The status includes fields for phase, message, conditions, observed generation, and other relevant information



_Appears in:_
- [Openshift](#openshift)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `phase` _[OpenshiftSncPhase](#openshiftsncphase)_ | Phase indicates the current lifecycle phase of the Kind cluster.<br />This field is used to track the progress of the cluster through its lifecycle stages.<br />Possible values include:<br />- Pending: The cluster is being created.<br />- Provisioning: The cluster is being provisioned, including setting up the infrastructure and<br />  installing the necessary components.<br />- Running: The cluster is fully operational and ready for use<br />E.g., Pending, Provisioning, Ready, Deleting, Error. |  | Enum: [Pending Provisioning Running Failed Deleting] <br /> |
| `message` _string_ | Message provides a human-readable status message.<br />This field is used to convey additional information about the current state of the cluster,<br />such as errors, warnings, or important updates.<br />It can be used to provide context about the cluster's status, such as whether it<br />is currently being provisioned, if there are any issues, or if it is ready for use. |  |  |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#condition-v1-meta) array_ | Conditions are used to provide detailed information about the state of the cluster,<br />including whether it is ready, if there are any issues, and other relevant status information |  |  |
| `observedGeneration` _integer_ | ObservedGeneration is the .metadata.generation that was last processed by the controller.<br />This field is used to track the version of the resource that the controller has observed.<br />It helps ensure that the controller is working with the most recent version of the resource<br />and can be used to detect changes in the resource that may require action. |  |  |
| `awsInstanceID` _string_ | AWSInstanceID is the ID of the EC2 instance provisioned with a Kind Cluster.<br />This field is used to track the specific instance that is running the Kind cluster.<br />It is particularly useful for managing the lifecycle of the cluster and for debugging purposes. |  |  |
| `kubeconfigSecretName` _string_ | KubeconfigSecretName is the name of the Kubernetes Secret where the cluster's<br />kubeconfig has been stored. This field is used to reference the secret that contains<br />the kubeconfig file for accessing the Openshift cluster. |  |  |
| `clusterReady` _boolean_ | ClusterReady indicates if the Kind cluster is fully provisioned and accessible.<br />This field is used to determine if the cluster is ready for use, meaning that all |  |  |
| `averagePrice` _string_ | This field is used to provide information about the cost of the spot instances used<br />for the Openshift cluster. It can help users understand the financial implications of<br />using spot instances for their cluster. |  |  |
| `provisionStartTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#time-v1-meta)_ | ProvisionStartTime records when the provisioning process began.<br />This field is used to track the start time of the provisioning process for the Openshift cluster. |  |  |
| `lastUpdateTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#time-v1-meta)_ | LastUpdateTime records the last time the status was updated.<br />This field is used to track when the status of the Openshift cluster was last modified.<br />It helps ensure that users and other components can see the most recent status of the cluster.<br />This is particularly useful for monitoring and debugging purposes. |  |  |
| `expirationTimestamp` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#time-v1-meta)_ | ExpirationTimestamp indicates when the cluster is scheduled to be terminated, based on TerminationPolicy.<br />This field is used to specify when the Openshift cluster is expected to be terminated. |  |  |
| `provisionId` _string_ | This field is used to track the specific provisioning session for the Openshift cluster. |  |  |


#### TerminationPolicy



TerminationPolicy defines automatic deletion parameters.



_Appears in:_
- [OpenshiftSpec](#openshiftspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `deleteAfterSeconds` _integer_ | DeleteAfterSeconds specifies a Time-To-Live (TTL) for the provisioned KindSpot.<br />After this duration (in seconds, starting from when the cluster becomes Ready or from creation),<br />the KindSpot and its underlying resources will be automatically destroyed.<br />This corresponds to the provisioning tool's '--timeout' parameter, which often expects a Go duration string.<br />The operator will convert these seconds into the required Go duration format for the tool. |  | Minimum: 60 <br /> |



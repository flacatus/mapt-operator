# API Reference

## Packages
- [mapt.redhat.com/v1alpha1](#maptredhatcomv1alpha1)


## mapt.redhat.com/v1alpha1

Package v1alpha1 contains API Schema definitions for the mapt v1alpha1 API group.

### Resource Types
- [Kind](#kind)
- [KindList](#kindlist)



#### CloudConfig



CloudConfig contains parameters to specify the cloud provider and access credentials.
TODO: Add Azure machines support. Ussually Azure spot instances are more cheaper.



_Appears in:_
- [KindSpec](#kindspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `provider` _string_ | Provider specifies the cloud provider name.<br />Currently, only "AWS" is supported. This field is designed for future extension. | AWS | Enum: [AWS] <br />Required: \{\} <br /> |
| `credentialsSecretRef` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#localobjectreference-v1-core)_ | CredentialsSecretRef is a reference to a Kubernetes Secret in the same namespace<br />as the KindSpot resource. This Secret must contain all necessary cloud provider<br />credentials and configurations, including the region.<br />The required keys within the Secret depend on the specified 'Provider'.<br />For 'AWS', this Secret is expected to contain:<br />  - "access-key": Your AWS access key ID.<br />  - "secret-key": Your AWS secret access key.<br />  - "region": The AWS region (e.g., "us-east-1").<br />  - "bucket": The S3 bucket name (for the provisioning tool's backend state, if applicable). |  | Required: \{\} <br /> |


#### Kind



Kind is the Schema for the kinds API.



_Appears in:_
- [KindList](#kindlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `mapt.redhat.com/v1alpha1` | | |
| `kind` _string_ | `Kind` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  |  |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  |  |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[KindSpec](#kindspec)_ |  |  |  |
| `status` _[KindStatus](#kindstatus)_ |  |  |  |


#### KindClusterConfig



KindClusterConfig contains parameters for the Kind cluster itself.



_Appears in:_
- [KindSpec](#kindspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `kubernetesVersion` _string_ | KubernetesVersion specifies the Kubernetes version for the Kind cluster (e.g., "v1.29.2").<br />This field is required. |  | MinLength: 1 <br />Required: \{\} <br /> |


#### KindList



KindList contains a list of Kind.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `mapt.redhat.com/v1alpha1` | | |
| `kind` _string_ | `KindList` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  |  |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  |  |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[Kind](#kind) array_ |  |  |  |


#### KindPhase

_Underlying type:_ _string_

KindPhase represents the lifecycle phase of a Kind resource.



_Appears in:_
- [KindStatus](#kindstatus)

| Field | Description |
| --- | --- |
| `Pending` |  |
| `Provisioning` |  |
| `Running` |  |
| `Failed` |  |
| `Deleting` |  |


#### KindSpec



KindSpec defines the desired state of Kind.



_Appears in:_
- [Kind](#kind)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `cloudConfig` _[CloudConfig](#cloudconfig)_ | CloudConfig holds cloud provider and credential configurations. |  | Required: \{\} <br /> |
| `machineConfig` _[MachineConfig](#machineconfig)_ | MachineConfig defines the configuration for the EC2 spot machine. |  | Required: \{\} <br /> |
| `kindClusterConfig` _[KindClusterConfig](#kindclusterconfig)_ | KindClusterConfig defines the configuration for the Kind cluster itself. |  | Required: \{\} <br /> |
| `outputKubeconfigSecretName` _string_ | OutputKubeconfigSecretName defines the prefix for the name of the Kubernetes Secret<br />that will store the kubeconfig for the provisioned Kind cluster.<br />If not provided, a default prefix will be used (e.g., "kindspot-<random>-kubeconfig").<br />The final name is generated using this prefix via Kubernetes' `generateName`.<br />This also corresponds to the Tekton 'cluster-access-secret-name' param. |  |  |
| `terminationPolicy` _[TerminationPolicy](#terminationpolicy)_ | TerminationPolicy defines when and how the cluster should be terminated. |  |  |


#### KindStatus



KindStatus defines the observed state of Kind.



_Appears in:_
- [Kind](#kind)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `phase` _[KindPhase](#kindphase)_ | Phase indicates the current lifecycle phase of the Kind cluster.<br />E.g., Pending, Provisioning, Ready, Deleting, Error. |  |  |
| `message` _string_ | Message provides a human-readable status message. |  |  |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#condition-v1-meta) array_ | Conditions represent the latest available observations of the Kind cluster state. |  |  |
| `awsInstanceID` _string_ | AWSInstanceID is the ID of the EC2 instance provisioned with a Kind Cluster. |  |  |
| `kubeconfigSecretName` _string_ | KubeconfigSecretName is the name of the Kubernetes Secret where the cluster's<br />kubeconfig has been stored. This will match `spec.outputKubeconfigSecretName` if provided,<br />or be a generated name. |  |  |
| `kindVersion` _string_ | KindVersion is the actual Kubernetes version of the provisioned Kind cluster. |  |  |
| `clusterReady` _boolean_ | ClusterReady indicates if the Kind cluster is fully provisioned and accessible. |  |  |
| `averagePrice` _string_ | AveragePrice reports the average acquisition price of the spot instance(s).<br />This field is a string to allow for currency. |  |  |
| `expirationTimestamp` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v/#time-v1-meta)_ | ExpirationTimestamp indicates when the cluster is scheduled to be terminated, based on TerminationPolicy. |  |  |
| `provisionId` _string_ | ProvisionId is the id of the backend used by the Kind provisioning tool. |  |  |


#### MachineConfig



MachineConfig contains parameters for configuring the EC2 spot machine.



_Appears in:_
- [KindSpec](#kindspec)

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


#### TerminationPolicy



TerminationPolicy defines automatic deletion parameters.



_Appears in:_
- [KindSpec](#kindspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `deleteAfterSeconds` _integer_ | DeleteAfterSeconds specifies a Time-To-Live (TTL) for the provisioned KindSpot.<br />After this duration (in seconds, starting from when the cluster becomes Ready or from creation),<br />the KindSpot and its underlying resources will be automatically destroyed.<br />This corresponds to the provisioning tool's '--timeout' parameter, which often expects a Go duration string.<br />The operator will convert these seconds into the required Go duration format for the tool. |  | Minimum: 60 <br /> |



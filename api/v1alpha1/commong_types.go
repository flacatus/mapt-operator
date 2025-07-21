package v1alpha1

// MachineConfig contains parameters for configuring the EC2 spot machine.
type MachineConfig struct {
	// Architecture for the EC2 instance.
	// +kubebuilder:validation:Enum=x86_64;arm64
	// +kubebuilder:default="x86_64"
	Architecture string `json:"architecture,omitempty"`

	// CPUs is the number of vCPUs for the EC2 instance.
	CPUs int32 `json:"cpus,omitempty"`

	// Indicates if the EC2 instance should have GPU support.
	// In case GPU is true, the instance type will be selected from the list of supported GPU instances.
	// +optional
	// +kubebuilder:default=false
	GPU bool `json:"gpu,omitempty"`

	// MemoryGiB is the amount of RAM for the EC2 instance in GiB.
	MemoryGiB int32 `json:"memoryGiB,omitempty"`

	// NestedVirtualizationEnabled specifies if the EC2 instance should have nested virtualization support.
	// +optional
	// +kubebuilder:default=false
	NestedVirtualizationEnabled bool `json:"nestedVirtualizationEnabled,omitempty"`

	// UseSpotInstances specifies whether to use EC2 spot instances.
	// Corresponds to the Tekton 'spot' param.
	// +optional
	// +kubebuilder:default=true
	UseSpotInstances bool `json:"useSpotInstances,omitempty"`

	// SpotPriceIncreasePercentage is the percentage to add on top of the current calculated spot price
	// to increase the chances of acquiring the machine. Only applies if UseSpotInstances is true.
	// A nil value means the underlying provisioning tool's default (if any) will be used, or the relevant command-line flag will be omitted. '0' is a valid percentage.
	// Corresponds to the Tekton 'spot-increase-rate' param (default '20').
	// +optional
	SpotPriceIncreasePercentage *int `json:"spotPriceIncreasePercentage,omitempty"`

	// Tags to apply to the AWS resources created by the provisioning tool.
	// The operator will convert this map into the string format the tool expects (e.g., "key1=value1,key2=value2").
	// Corresponds to the Tekton 'tags' param.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

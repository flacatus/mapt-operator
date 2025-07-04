package controller

import (
	"fmt"

	"github.com/flacatus/mapt-operator/api/v1alpha1"
	"github.com/google/uuid"
	"github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/action/kind"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
)

// ClusterProvisioner abstracts the logic required to provision a Kind or other cluster type.
// It returns the kubeconfig path and a unique provision ID to be tracked in status.
type ClusterProvisioner interface {
	Provision(cluster *v1alpha1.Kind) (kubeconfigPath string, provisionID string, err error)
	Deprovision(cluster *v1alpha1.Kind) error
}

// kindClusterProvisioner is a concrete implementation of ClusterProvisioner using Kind and Terraform.
type kindClusterProvisioner struct{}

// Compile-time check: ensure kindClusterProvisioner implements ClusterProvisioner.
var _ ClusterProvisioner = (*kindClusterProvisioner)(nil)

// NewKindProvisioner creates a new instance of a Kind-based cluster provisioner.
func NewKindProvisioner() ClusterProvisioner {
	return &kindClusterProvisioner{}
}

// Provision provisions a Kind cluster and returns the path to the generated kubeconfig
// and a unique provision ID that can be tracked in the resource's status.
func (p *kindClusterProvisioner) Provision(cluster *v1alpha1.Kind) (kubeconfig string, provisionId string, err error) {
	// Generate a UUID to uniquely identify this provisioning session
	provisionID := uuid.New().String()

	// Define the remote backend URL for Terraform
	backedURL := fmt.Sprintf("s3://mapt-kind-bucket/mapt/kind/%s", provisionID)

	// Construct the Terraform context configuration
	ctxArgs := &context.ContextArgs{
		ProjectName:           cluster.Name,
		BackedURL:             backedURL,
		SpotPriceIncreaseRate: 20,
		Tags:                  cluster.Spec.MachineConfig.Tags,
		ForceDestroy:          true,
		ResultsOutput:         "/tmp/results",
	}

	// Define instance configuration based on the Kind resource
	instanceRequest := &instancetypes.AwsInstanceRequest{
		CPUs:      cluster.Spec.MachineConfig.CPUs,
		MemoryGib: cluster.Spec.MachineConfig.MemoryGiB,
		Arch:      instancetypes.Amd64,
	}

	// Define Kind-specific provisioning arguments
	kindArgs := &kind.KindArgs{
		Prefix:          cluster.Name,
		Arch:            cluster.Spec.MachineConfig.Architecture,
		InstanceRequest: instanceRequest,
		Version:         cluster.Spec.KindClusterConfig.KubernetesVersion,
		Spot:            true,
	}

	// Perform the provisioning process
	if err := kind.Create(ctxArgs, kindArgs); err != nil {
		return "", "", err
	}

	// Return the generated kubeconfig path and provision ID
	return "/tmp/results/kubeconfig", provisionID, nil
}

func (p *kindClusterProvisioner) Deprovision(cluster *v1alpha1.Kind) error {
	return kind.Destroy(&context.ContextArgs{
		ProjectName:           cluster.Name,
		BackedURL:             fmt.Sprintf("s3://mapt-kind-bucket/mapt/kind/%s", *cluster.Status.ProvisionId),
		SpotPriceIncreaseRate: *cluster.Spec.MachineConfig.SpotPriceIncreasePercentage,
		ForceDestroy:          true,
	})
}

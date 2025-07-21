package clusters

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mapt-oss/mapt-operator/api/v1alpha1"
	"github.com/redhat-developer/mapt/pkg/manager/context"
	instancetypes "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"github.com/redhat-developer/mapt/pkg/provider/aws/action/kind"
)

type KindProvisioner interface {
	Provision(cluster *v1alpha1.Kind) (*KindMetadata, error)
	Deprovision(cluster *v1alpha1.Kind) error
}

type kindClusterProvisioner struct {
	CloudCredentials *ProvisionCloudCredentials
}

func (p *kindClusterProvisioner) Provision(cluster *v1alpha1.Kind) (*KindMetadata, error) {
	if cluster.Status.ProvisionId == nil || *cluster.Status.ProvisionId == "" {
		return nil, fmt.Errorf("missing or empty Status.ProvisionId")
	}

	if _, ok := kind.KindK8sVersions[cluster.Spec.KindClusterConfig.KubernetesVersion]; !ok {
		var supported []string
		for version := range kind.KindK8sVersions {
			supported = append(supported, version)
		}
		return nil, fmt.Errorf(
			"mapt does not support Kubernetes version: %s (supported versions: %v)",
			cluster.Spec.KindClusterConfig.KubernetesVersion,
			supported,
		)
	}

	provisionID := *cluster.Status.ProvisionId
	if err := os.Mkdir(filepath.Join(".", provisionID), 0755); err != nil {
		return nil, fmt.Errorf("failed to create provision directory: %w", err)
	}

	ctxArgs := &context.ContextArgs{
		ProjectName:           cluster.Name,
		BackedURL:             p.buildBackendURL(provisionID),
		SpotPriceIncreaseRate: *cluster.Spec.MachineConfig.SpotPriceIncreasePercentage,
		Tags:                  cluster.Spec.MachineConfig.Tags,
		ForceDestroy:          true,
	}

	kindArgs := &kind.KindArgs{
		Prefix:         cluster.Name,
		Arch:           cluster.Spec.MachineConfig.Architecture,
		ComputeRequest: p.buildComputeRequest(cluster),
		Version:        cluster.Spec.KindClusterConfig.KubernetesVersion,
		Spot:           true,
	}

	kindMetadataResults, err := kind.Create(ctxArgs, kindArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to create kind cluster: %w", err)
	}

	return &KindMetadata{
		Username:   kindMetadataResults.Username,
		PrivateKey: kindMetadataResults.PrivateKey,
		Host:       kindMetadataResults.Host,
		Kubeconfig: kindMetadataResults.Kubeconfig,
		SpotPrice:  *kindMetadataResults.SpotPrice,
	}, nil
}

func (p *kindClusterProvisioner) Deprovision(cluster *v1alpha1.Kind) error {
	return kind.Destroy(&context.ContextArgs{
		ProjectName:           cluster.Name,
		BackedURL:             p.buildBackendURL(*cluster.Status.ProvisionId),
		SpotPriceIncreaseRate: *cluster.Spec.MachineConfig.SpotPriceIncreasePercentage,
		ForceDestroy:          true,
	})
}

func (p *kindClusterProvisioner) buildBackendURL(provisionID string) string {
	return fmt.Sprintf("s3://%s/mapt/kind/%s", p.CloudCredentials.S3BucketName, provisionID)
}

func (p *kindClusterProvisioner) buildComputeRequest(cluster *v1alpha1.Kind) *instancetypes.ComputeRequestArgs {
	if cluster.Spec.MachineConfig.GPU {
		return &instancetypes.ComputeRequestArgs{
			ComputeSizes: SupportedAwsGPUsInstances,
			Arch:         instancetypes.Amd64,
		}
	}

	return &instancetypes.ComputeRequestArgs{
		CPUs:      cluster.Spec.MachineConfig.CPUs,
		MemoryGib: cluster.Spec.MachineConfig.MemoryGiB,
		Arch:      instancetypes.Amd64,
	}
}

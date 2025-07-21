package clusters

import (
	"fmt"
	"os"

	"github.com/mapt-oss/mapt-operator/api/v1alpha1"
	"github.com/redhat-developer/mapt/pkg/manager/context"
	instancetypes "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	openshiftsnc "github.com/redhat-developer/mapt/pkg/provider/aws/action/openshift-snc"
)

var OpenshiftSNCSupportedVersions = []string{
	"4.19.0",
}

type OpenshiftProvisioner interface {
	Provision(cluster *v1alpha1.Openshift) (*OpenshiftMetadata, error)
	Deprovision(cluster *v1alpha1.Openshift) error
}

type openshiftSncProvisioner struct {
	CloudCredentials *ProvisionCloudCredentials
}

func (p *openshiftSncProvisioner) Provision(cluster *v1alpha1.Openshift) (*OpenshiftMetadata, error) {
	if !isSupportedOpenshiftVersion(cluster.Spec.OpenshiftClusterConfig.OpenshiftVersion) {
		return nil, fmt.Errorf("unsupported OpenShift version: %s (supported: %v)", cluster.Spec.OpenshiftClusterConfig.OpenshiftVersion, OpenshiftSNCSupportedVersions)
	}

	if err := validateProvisionInput(cluster); err != nil {
		return nil, err
	}

	pullSecretFile, err := getValidatedPullSecretFile()
	if err != nil {
		return nil, err
	}

	ctxArgs := p.buildContextArgs(cluster)
	sncArgs := p.buildSNCArgs(cluster, pullSecretFile)

	metadata, err := openshiftsnc.Create(ctxArgs, sncArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to create openshift snc cluster: %w", err)
	}
	if metadata == nil {
		return nil, fmt.Errorf("received nil metadata from OpenShift SNC creation")
	}

	return &OpenshiftMetadata{
		Username:          metadata.Username,
		PrivateKey:        metadata.PrivateKey,
		Host:              metadata.Host,
		Kubeconfig:        metadata.Kubeconfig,
		KubeadminPassword: metadata.KubeadminPass,
		SpotPrice:         *metadata.SpotPrice,
		ConsoleURL:        metadata.ConsoleUrl,
	}, nil
}

func (p *openshiftSncProvisioner) Deprovision(cluster *v1alpha1.Openshift) error {
	return openshiftsnc.Destroy(&context.ContextArgs{
		ProjectName:           cluster.Name,
		BackedURL:             fmt.Sprintf("s3://%s/mapt/openshift-snc/%s", p.CloudCredentials.S3BucketName, *cluster.Status.ProvisionId),
		SpotPriceIncreaseRate: *cluster.Spec.MachineConfig.SpotPriceIncreasePercentage,
		ForceDestroy:          true,
	})
}

func (p *openshiftSncProvisioner) buildContextArgs(cluster *v1alpha1.Openshift) *context.ContextArgs {
	return &context.ContextArgs{
		ProjectName:           cluster.Name,
		BackedURL:             fmt.Sprintf("s3://%s/mapt/openshift-snc/%s", p.CloudCredentials.S3BucketName, *cluster.Status.ProvisionId),
		SpotPriceIncreaseRate: *cluster.Spec.MachineConfig.SpotPriceIncreasePercentage,
		Tags:                  cluster.Spec.MachineConfig.Tags,
		ForceDestroy:          true,
	}
}

func (p *openshiftSncProvisioner) buildSNCArgs(cluster *v1alpha1.Openshift, pullSecretFile string) *openshiftsnc.OpenshiftSNCArgs {
	return &openshiftsnc.OpenshiftSNCArgs{
		Prefix:         cluster.Name,
		Version:        "4.19.0",
		ComputeRequest: p.buildComputeRequest(cluster),
		Arch:           cluster.Spec.MachineConfig.Architecture,
		PullSecretFile: pullSecretFile,
		Spot:           true,
	}
}

func (p *openshiftSncProvisioner) buildComputeRequest(cluster *v1alpha1.Openshift) *instancetypes.ComputeRequestArgs {
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

func getFromEnvOrError(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s is not set or empty", key)
	}
	return value, nil
}

func getValidatedPullSecretFile() (string, error) {
	path, err := getFromEnvOrError("OPENSHIFT_PULL_SECRET_FILE")
	if err != nil {
		return "", err
	}
	info, statErr := os.Stat(path)
	if statErr != nil {
		return "", fmt.Errorf("cannot access pull secret file at %s: %w", path, statErr)
	}
	if info.Size() == 0 {
		return "", fmt.Errorf("pull secret file at %s is empty", path)
	}
	return path, nil
}

func validateProvisionInput(cluster *v1alpha1.Openshift) error {
	if cluster.Status.ProvisionId == nil || *cluster.Status.ProvisionId == "" {
		return fmt.Errorf("missing ProvisionID")
	}
	return nil
}

func isSupportedOpenshiftVersion(version string) bool {
	for _, supported := range OpenshiftSNCSupportedVersions {
		if version == supported {
			return true
		}
	}
	return false
}

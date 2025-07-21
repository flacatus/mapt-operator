package clusters

import (
	"context"
	"errors"
	"fmt"

	"github.com/mapt-oss/mapt-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProvisionerMetadata struct {
	ClusterType string
	Provisioner string
}

type GenericMaptProvisioner interface {
	Provision(cluster *MaptCluster) (*ClusterProvisionerMetadata, error)
	Deprovision(cluster *MaptCluster) error
}

type maptProvisioner struct {
	openshiftProv OpenshiftProvisioner
	kindProv      KindProvisioner
}

func NewGenericMaptProvisioner(ctx context.Context, c client.Client) (GenericMaptProvisioner, error) {
	creds, err := loadCloudCredentials(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to load cloud credentials: %w", err)
	}

	return &maptProvisioner{
		openshiftProv: &openshiftSncProvisioner{
			CloudCredentials: creds,
		},
		kindProv: &kindClusterProvisioner{
			CloudCredentials: creds,
		},
	}, nil
}

func (p *maptProvisioner) Provision(cluster *MaptCluster) (*ClusterProvisionerMetadata, error) {
	switch cluster.Type {
	case OpenshiftClusterType:
		ocp, err := getOpenshift(cluster.Object)
		if err != nil {
			return nil, err
		}
		metadata, err := p.openshiftProv.Provision(ocp)
		if err != nil {
			return nil, fmt.Errorf("failed to provision OpenShift cluster: %w", err)
		}
		return &ClusterProvisionerMetadata{
			Type:              OpenshiftClusterType,
			OpenshiftMetadata: metadata,
		}, nil

	case KindClusterType:
		kind, err := getKind(cluster.Object)
		if err != nil {
			return nil, err
		}
		kindMetadata, err := p.kindProv.Provision(kind)
		if err != nil {
			return nil, fmt.Errorf("failed to provision Kind cluster: %w", err)
		}
		return &ClusterProvisionerMetadata{
			Type:         KindClusterType,
			KindMetadata: kindMetadata,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported cluster type: %s", cluster.Type)
	}
}

func (p *maptProvisioner) Deprovision(cluster *MaptCluster) error {
	switch cluster.Type {
	case OpenshiftClusterType:
		ocp, err := getOpenshift(cluster.Object)
		if err != nil {
			return err
		}
		return p.openshiftProv.Deprovision(ocp)

	case KindClusterType:
		kind, err := getKind(cluster.Object)
		if err != nil {
			return err
		}
		return p.kindProv.Deprovision(kind)

	default:
		return fmt.Errorf("unsupported cluster type: %s", cluster.Type)
	}
}

func loadCloudCredentials(ctx context.Context, c client.Client) (*ProvisionCloudCredentials, error) {
	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{
		Name:      CloudCredentialsSecretName,
		Namespace: CloudCredentialsSecretNamespace,
	}
	if err := c.Get(ctx, secretKey, secret); err != nil {
		return nil, fmt.Errorf("failed to get secret '%s' in namespace '%s': %w", secretKey.Name, secretKey.Namespace, err)
	}

	creds := &ProvisionCloudCredentials{
		AccessKeyID:     string(secret.Data["access-key"]),
		SecretAccessKey: string(secret.Data["secret-key"]),
		Region:          string(secret.Data["region"]),
		S3BucketName:    string(secret.Data["bucket"]),
	}

	if err := creds.Validate(); err != nil {
		return nil, err
	}
	return creds, nil
}

func (c *ProvisionCloudCredentials) Validate() error {
	if c.AccessKeyID == "" {
		return errors.New("missing cloud credential: access-key")
	}
	if c.SecretAccessKey == "" {
		return errors.New("missing cloud credential: secret-key")
	}
	if c.Region == "" {
		return errors.New("missing cloud credential: region")
	}
	if c.S3BucketName == "" {
		return errors.New("missing cloud credential: bucket")
	}
	return nil
}

func getOpenshift(obj client.Object) (*v1alpha1.Openshift, error) {
	ocp, ok := obj.(*v1alpha1.Openshift)
	if !ok {
		return nil, errors.New("object is not of type *v1alpha1.Openshift")
	}
	return ocp, nil
}

func getKind(obj client.Object) (*v1alpha1.Kind, error) {
	kind, ok := obj.(*v1alpha1.Kind)
	if !ok {
		return nil, errors.New("object is not of type *v1alpha1.Kind")
	}
	return kind, nil
}

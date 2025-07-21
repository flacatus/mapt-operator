package kind

import (
	"errors"

	"github.com/mapt-oss/mapt-operator/pkg/clusters"
)

// MockProvisioner is a mock implementation of GenericMaptProvisioner for testing.
type MockProvisioner struct {
	MockProvision   func(cluster *clusters.MaptCluster) (*clusters.ClusterProvisionerMetadata, error)
	MockDeprovision func(cluster *clusters.MaptCluster) error
}

func (m *MockProvisioner) Provision(cluster *clusters.MaptCluster) (*clusters.ClusterProvisionerMetadata, error) {
	if m.MockProvision != nil {
		return m.MockProvision(cluster)
	}
	return nil, errors.New("MockProvision function was not implemented for this test")
}

func (m *MockProvisioner) Deprovision(cluster *clusters.MaptCluster) error {
	if m.MockDeprovision != nil {
		return m.MockDeprovision(cluster)
	}
	return errors.New("MockDeprovision function was not implemented for this test")
}

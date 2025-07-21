package clusters

import "sigs.k8s.io/controller-runtime/pkg/client"

type ClusterType string

const (
	OpenshiftClusterType            ClusterType = "openshift"
	KindClusterType                 ClusterType = "kind"
	CloudCredentialsSecretName      string      = "mapt-operator-mapt-kind-secret"
	CloudCredentialsSecretNamespace string      = "mapt-operator-system"
)

var (
	SupportedAwsGPUsInstances = []string{
		// G6e
		"g6e.12xlarge", "g6e.16xlarge", "g6e.24xlarge", "g6e.48xlarge",

		// G6
		"g6.12xlarge", "g6.16xlarge", "g6.24xlarge", "g6.48xlarge",

		// G5
		"g5.12xlarge", "g5.16xlarge", "g5.48xlarge",

		// P4 (A100)
		"p4d.24xlarge", "p4de.24xlarge",

		// P5 (H100)
		"p5.48xlarge", "p5e.48xlarge", "p5en.48xlarge",
	}
)

type MaptCluster struct {
	Type   ClusterType
	Object client.Object
}

type ClusterProvisionerMetadata struct {
	Type              ClusterType
	OpenshiftMetadata *OpenshiftMetadata `json:"openshiftMetadata,omitempty"`
	KindMetadata      *KindMetadata      `json:"kindMetadata,omitempty"`
}

type OpenshiftMetadata struct {
	Username          string  `json:"username"`
	PrivateKey        string  `json:"privateKey"`
	Host              string  `json:"host"`
	Kubeconfig        string  `json:"kubeconfig"`
	KubeadminPassword string  `json:"kubeadminPassword"`
	SpotPrice         float64 `json:"spotPrice"`
	ConsoleURL        string  `json:"consoleURL"`
}

type KindMetadata struct {
	Username   string  `json:"username"`
	PrivateKey string  `json:"privateKey"`
	Host       string  `json:"host"`
	Kubeconfig string  `json:"kubeconfig"`
	SpotPrice  float64 `json:"spotPrice"`
}

type ProvisionCloudCredentials struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	S3BucketName    string `json:"s3BucketName"`
	Region          string `json:"region"`
}

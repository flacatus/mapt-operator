# MAPT Operator

A Kubernetes operator for managing cloud-based Kubernetes, OpenShift SNO clusters, or empty cloud instances with GPUs support using spot instances for cost optimization.

## Description

The MAPT Operator is a powerful Kubernetes operator that revolutionizes cloud infrastructure management by automating the provisioning and lifecycle management of cloud-based Kubernetes clusters, OpenShift Single Node OpenShift (SNO) clusters, and standalone instances on AWS.

**Perfect for cost-effective local testing and development** - The operator leverages AWS spot instances to provide cheap, on-demand resources for testing, development, and experimentation without the overhead of maintaining permanent infrastructure.

Built for modern AI/ML workloads, all cluster types support GPU-enabled instances optimized for model training and development. The operator leverages AWS spot instances to deliver significant cost savings while providing enterprise-grade reliability through declarative configuration and automated infrastructure management.

Key capabilities include intelligent resource provisioning, automatic cluster termination based on configurable policies, seamless kubeconfig management, and real-time status tracking - all designed to streamline your cloud-native development and AI/ML workflows.

## Usage and API Reference

For detailed information about using the operator, API references, and comprehensive documentation, please refer to the [`./docs`](./docs) directory.

The documentation includes:

- **API Reference**: Complete CRD specifications and field descriptions
- **Usage Examples**: Sample configurations for different use cases
- **Troubleshooting**: Common issues and solutions
- **Advanced Configuration**: Detailed setup and customization guides

## Features

- **Kubernetes Vanilla Clusters**: Provision Kubernetes clusters using Kind on AWS spot instances
- **OpenShift SNC**: Provision OpenShift Single Node Clusters on AWS spot instances
- **GPU Support**: Deploy both Kubernetes and OpenShift clusters with GPU-enabled spot instances for AI model training and development
- **Cost Optimization**: Uses AWS spot instances to reduce infrastructure costs
- **Automatic Termination**: Configurable TTL policies for automatic cluster cleanup
- **Kubeconfig Management**: Automatic generation and storage of cluster access credentials
- **Status Tracking**: Real-time status updates and lifecycle management

## Getting Started

### Prerequisites

- Go version v1.24.0+
- Docker version 17.03+ or Podman
- kubectl version v1.21.3+
- Access to a Kubernetes v1.21.3+ cluster
- AWS credentials with appropriate permissions for EC2 and S3

### Cloud Credentials Setup

Before using the operator, you need to configure your AWS credentials and OpenShift pull secret in the [`config/manager/secret.yaml`](config/manager/secret.yaml) file:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mapt-kind-secret
  namespace: mapt-operator-system
type: Opaque
stringData:
  access-key: "YOUR_AWS_ACCESS_KEY"
  bucket: "YOUR_S3_BUCKET_NAME"
  region: "us-east-1"
  secret-key: "YOUR_AWS_SECRET_KEY"
  pull-secret.json: |
    {"auths":{"cloud.openshift.com":{"auth":"YOUR_PULL_SECRET_AUTH","email":"your-email@example.com"},"quay.io":{"auth":"YOUR_PULL_SECRET_AUTH","email":"your-email@example.com"},"registry.connect.redhat.com":{"auth":"YOUR_PULL_SECRET_AUTH","email":"your-email@example.com"},"registry.redhat.io":{"auth":"YOUR_PULL_SECRET_AUTH","email":"your-email@example.com"}}}
```

**Note**: The actual secret file uses placeholder values. Replace them with your actual AWS credentials and OpenShift pull secret.

**OpenShift Pull Secret:**
For OpenShift cluster provisioning, you need a valid OpenShift pull secret. You can obtain this from:

- [Red Hat OpenShift Cluster Manager](https://console.redhat.com/openshift/install/pull-secret)
- Your Red Hat account

The pull secret should be saved as a JSON file and referenced in the secret configuration above.

**Secret Structure:**
The secret must contain the following keys:

- `access-key`: Your AWS access key ID
- `secret-key`: Your AWS secret access key
- `region`: AWS region (e.g., us-east-1)
- `bucket`: S3 bucket name for state management
- `pull-secret.json`: OpenShift pull secret JSON content

**Important Notes:**

- The secret will be created in the `mapt-operator-system` namespace during deployment
- The secret name must be `mapt-kind-secret` for the operator to work correctly. The operator deployment uses a prefix for all resources. The final secret name in the cluster will be `mapt-operator-mapt-kind-secret`.
- Ensure your pull secret has access to the required OpenShift registries (quay.io, registry.redhat.io, etc.)

### Installation

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/mapt-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don't work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/mapt-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

### To Uninstall

**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/mapt-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/mapt-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Development

### Running Tests

```bash
# Run all tests
make test

# Run specific test suites
go test ./internal/controller/kind/... -v
go test ./internal/controller/openshift-snc/... -v
```

### Building Locally

```bash
# Build the operator
make build

# Run locally
make run
```

## Contributing

We welcome contributions to the MAPT Operator! Here's how you can get involved:

### Development Setup

1. **Fork and clone the repository**
2. **Set up your development environment**:

   ```bash
   make setup-envtest
   make manifests generate
   ```

3. **Run tests**:

   ```bash
   make test
   ```

### Contribution Guidelines

- **Code Style**: Follow Go conventions and use `gofmt` for formatting
- **Testing**: Add tests for new features and ensure all tests pass
- **Documentation**: Update documentation for any API changes
- **Commits**: Use conventional commit messages
- **Pull Requests**: Create descriptive PRs with clear titles

### Areas for Contribution

- **Bug fixes**: Report and fix issues
- **Feature development**: Add new capabilities
- **Documentation**: Improve docs and examples
- **Testing**: Add test coverage
- **Performance**: Optimize resource usage

### Getting Help

- **Issues**: Use GitHub issues for bug reports and feature requests
- **Discussions**: Use GitHub Discussions for questions and ideas
- **Code of Conduct**: Be respectful and inclusive

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

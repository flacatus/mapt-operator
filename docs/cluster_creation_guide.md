# Cluster Creation Guide

This guide explains how to use Custom Resources (CRs) to create Kubernetes and OpenShift clusters with the MAPT Operator, including how to configure GPU support for AI/ML workloads development.

**Ideal for cost-effective local testing and development** - The operator provides cheap, on-demand resources using AWS spot instances, perfect for testing, development, and experimentation without the overhead of maintaining permanent infrastructure.

## Overview

The MAPT Operator supports two main cluster types:

- **Kubernetes Clusters** (using Kind)
- **OpenShift Single Node OpenShift (SNO) Clusters**

Both cluster types can be configured with or without GPU support, making them suitable for various workloads including AI/ML model training and development.

## Prerequisites

Before creating clusters, ensure you have:

1. The MAPT Operator deployed in your cluster
2. AWS credentials configured in the [`config/manager/secret.yaml`](../../config/manager/secret.yaml) file
3. OpenShift pull secret (for OpenShift clusters) - configured in the same secret file

## Cluster Types

### 1. Kubernetes Clusters (Kind)

Kubernetes clusters are provisioned using Kind (Kubernetes in Docker) on AWS spot instances.

#### Basic Kubernetes Cluster (No GPU)

```yaml
apiVersion: mapt.redhat.com/v1alpha1
kind: Kind
metadata:
  name: my-k8s-cluster
  namespace: mapt-operator-system
spec:
  cloudConfig:
    provider: AWS
    credentialsSecretRef:
      name: mapt-kind-secret

  machineConfig:
    architecture: x86_64
    cpus: 8
    memoryGiB: 16
    useSpotInstances: true
    spotPriceIncreasePercentage: 20
    tags:
      env: development
      owner: team-name

  kindClusterConfig:
    kubernetesVersion: v1.32.0

  outputKubeconfigSecretName: my-k8s-kubeconfig
  terminationPolicy:
    deleteAfterSeconds: 86400  # 24 hours
```

#### Kubernetes Cluster with GPU Support

```yaml
apiVersion: mapt.redhat.com/v1alpha1
kind: Kind
metadata:
  name: my-gpu-k8s-cluster
  namespace: mapt-operator-system
spec:
  cloudConfig:
    provider: AWS
    credentialsSecretRef:
      name: mapt-kind-secret

  machineConfig:
    architecture: x86_64
    gpu: true  # Enable GPU support
    useSpotInstances: true
    spotPriceIncreasePercentage: 30
    tags:
      env: ai-training
      workload: gpu
      owner: ml-team

  kindClusterConfig:
    kubernetesVersion: v1.29.2

  outputKubeconfigSecretName: my-gpu-k8s-kubeconfig
  terminationPolicy:
    deleteAfterSeconds: 172800  # 48 hours for longer training jobs
```

### 2. OpenShift SNO Clusters

OpenShift Single Node OpenShift (SNO) clusters provide a full OpenShift experience on a single node.

#### Basic OpenShift SNO Cluster (No GPU)

```yaml
apiVersion: mapt.redhat.com/v1alpha1
kind: Openshift
metadata:
  name: my-openshift-cluster
  namespace: mapt-operator-system
spec:
  machineConfig:
    architecture: x86_64
    cpus: 32
    memoryGiB: 128
    useSpotInstances: true
    spotPriceIncreasePercentage: 30
    tags:
      environment: development
      team: platform

  openshiftClusterConfig:
    openshiftVersion: "4.19.0"  # Required field - specify OpenShift version

  terminationPolicy:
    deleteAfterSeconds: 86400  # 24 hours
```

#### OpenShift SNO Cluster with GPU Support

```yaml
apiVersion: mapt.redhat.com/v1alpha1
kind: Openshift
metadata:
  name: my-gpu-openshift-cluster
  namespace: mapt-operator-system
spec:
  machineConfig:
    architecture: x86_64
    gpu: true  # Enable GPU support
    useSpotInstances: true
    spotPriceIncreasePercentage: 40
    tags:
      environment: ai-production
      workload: gpu-training
      team: ml-platform

  openshiftClusterConfig:
    openshiftVersion: "4.19.0"

  terminationPolicy:
    deleteAfterSeconds: 259200  # 72 hours for extended training
```

## Machine Configuration Options

### GPU Configuration

To enable GPU support, set `gpu: true` in the `machineConfig` section:

```yaml
machineConfig:
  gpu: true  # This enables GPU instance selection
```

**Important Notes:**

- When `gpu: true`, the operator automatically selects GPU-enabled instance types
- GPU instances typically have higher costs but provide significant acceleration for AI/ML workloads
- Available GPU instance types depend on the AWS region and current availability

### Architecture Options

```yaml
machineConfig:
  architecture: x86_64  # Intel/AMD processors
```

### Resource Configuration

```yaml
machineConfig:
  cpus: 8              # Number of vCPUs (default: 8)
  memoryGiB: 16        # Memory in GiB (default: 16)
  nestedVirtualizationEnabled: false  # Enable nested virtualization
```

### Spot Instance Configuration

```yaml
machineConfig:
  useSpotInstances: true              # Use spot instances (default: true)
  spotPriceIncreasePercentage: 20     # Increase bid by 20% for better availability
```

### Resource Tagging

```yaml
machineConfig:
  tags:
    env: production
    owner: ml-team
    project: ai-training
    cost-center: research
```

## Cluster Lifecycle Management

### Termination Policy

All clusters support automatic termination based on configurable TTL:

```yaml
terminationPolicy:
  deleteAfterSeconds: 86400  # Cluster will be deleted after 24 hours
```

**Common TTL Values:**

- **Development**: 86400 seconds (24 hours)
- **Testing**: 43200 seconds (12 hours)
- **AI Training**: 259200 seconds (72 hours)
- **Production**: Set to 0 or omit for no automatic termination

### Kubeconfig Management

The operator automatically creates Kubernetes secrets containing cluster access credentials:

```yaml
spec:
  outputKubeconfigSecretName: my-cluster-kubeconfig
```

## Monitoring Cluster Status

### Check Cluster Status

```bash
# List all clusters
kubectl get kinds -n mapt-operator-system
kubectl get openshifts -n mapt-operator-system

# Get detailed status
kubectl describe kind my-k8s-cluster -n mapt-operator-system
kubectl describe openshift my-openshift-cluster -n mapt-operator-system
```

### Cluster Phases

Clusters go through the following phases:

- **Pending**: Initial creation request
- **Provisioning**: Infrastructure and cluster setup
- **Running**: Cluster is ready for use
- **Failed**: Provisioning encountered an error
- **Deleting**: Cluster is being terminated

### Access Your Clusters

```bash
# Get the kubeconfig secret name
kubectl get kind my-k8s-cluster -n mapt-operator-system -o jsonpath='{.status.kubeconfigSecretName}'

# Extract and use the kubeconfig
kubectl get secret <secret-name> -n mapt-operator-system -o jsonpath='{.data.kubeconfig}' | base64 -d > cluster-kubeconfig.yaml
kubectl --kubeconfig=cluster-kubeconfig.yaml get nodes
```

## Best Practices

### For AI/ML Workloads

1. **GPU Configuration**:

   ```yaml
   machineConfig:
     gpu: true
   ```

2. **Extended TTL for Training**:

   ```yaml
   terminationPolicy:
     deleteAfterSeconds: 259200  # 72 hours for long training jobs
   ```

3. **Higher Spot Price Increase**:

   ```yaml
   machineConfig:
     spotPriceIncreasePercentage: 40  # Better availability for GPU instances
   ```

### For Development/Testing

1. **Standard Configuration**:

   ```yaml
   machineConfig:
     gpu: false
     cpus: 8
     memoryGiB: 16
   ```

2. **Shorter TTL**:

   ```yaml
   terminationPolicy:
     deleteAfterSeconds: 43200  # 12 hours
   ```

### Cost Optimization

1. **Use Spot Instances** (default):

   ```yaml
   machineConfig:
     useSpotInstances: true
   ```

2. **Appropriate Spot Price Increase**:

   ```yaml
   machineConfig:
     spotPriceIncreasePercentage: 20  # Balance between cost and availability
   ```

3. **Resource Tagging**:

   ```yaml
   machineConfig:
     tags:
       cost-center: development
       project: ai-research
   ```

## Troubleshooting

### Common Issues

1. **Cluster Stuck in Provisioning**:
   - Check AWS credentials and permissions
   - Verify spot instance availability in the region
   - Review operator logs: `kubectl logs -n mapt-operator-system deployment/controller-manager`

2. **GPU Instance Not Available**:
   - Try different AWS regions
   - Increase `spotPriceIncreasePercentage`
   - Check current GPU instance availability

3. **Cluster Creation Fails**:
   - Verify OpenShift pull secret (for OpenShift clusters)
   - Check AWS quota limits
   - Review the cluster status: `kubectl describe kind <cluster-name>`

### Getting Help

- Check the operator logs: `kubectl logs -n mapt-operator-system deployment/controller-manager`
- Review cluster events: `kubectl get events -n mapt-operator-system`
- Verify AWS credentials and permissions
- Check spot instance availability in your region

## Examples

### Complete Examples

See the sample configurations in the `config/samples/` directory:

- [`kind_spot.yaml`](../../config/samples/kind_spot.yaml) - Basic Kubernetes cluster
- [`kind_gpu_spot.yaml`](../../config/samples/kind_gpu_spot.yaml) - Kubernetes cluster with GPU
- [`openshift_spot.yaml`](../../config/samples/openshift_spot.yaml) - Basic OpenShift SNO cluster

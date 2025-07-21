package kind

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/konflux-ci/operator-toolkit/controller"
	"github.com/mapt-oss/mapt-operator/api/v1alpha1"
	"github.com/mapt-oss/mapt-operator/internal/metadata"
	"github.com/mapt-oss/mapt-operator/pkg/clusters"
	"github.com/mapt-oss/mapt-operator/pkg/controllerutils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// adapter wraps the reconciliation logic for the Kind custom resource.
// It handles provisioning, deprovisioning, status updates, and secret management.
type adapter struct {
	// client is the Kubernetes client used to interact with the API server.
	client client.Client

	// ctx is the context for the reconciliation process.
	ctx context.Context

	// kind is the Kind custom resource being reconciled.
	kind *v1alpha1.Kind

	// validations holds any validation functions to be executed during reconciliation.
	validations []controller.ValidationFunction

	// provisioner is the cluster provisioner used to manage Kind clusters.
	// It must not be nil.
	provisioner clusters.GenericMaptProvisioner

	// cloudCrentials holds metadata about the cloud provider used for provisioning.
	cloudCrentials *clusters.ClusterProvisionerMetadata

	// log is the logger used for logging messages during reconciliation.
	log logr.Logger
}

// newAdapter initializes the Kind adapter with necessary dependencies and context.
// Returns an error if the provisioner is nil.
func newAdapter(ctx context.Context, c client.Client, kind *v1alpha1.Kind, prv clusters.GenericMaptProvisioner, l logr.Logger) (*adapter, error) {
	if prv == nil {
		return nil, fmt.Errorf("no provisioner provided")
	}
	return &adapter{
		client:      c,
		ctx:         ctx,
		kind:        kind,
		log:         l.WithValues("name", kind.Name, "namespace", kind.Namespace),
		provisioner: prv,
		validations: []controller.ValidationFunction{},
	}, nil
}

// EnsureFinalizerIsAdded ensures the finalizer is present on the Kind resource.
// If it's missing, it adds and patches the resource.
func (a *adapter) EnsureFinalizerIsAdded() (controller.OperationResult, error) {
	if controllerutil.ContainsFinalizer(a.kind, metadata.KindFinalizer) {
		return controller.ContinueProcessing()
	}

	kindCopy := a.kind.DeepCopy()
	patch := client.MergeFrom(kindCopy)
	controllerutil.AddFinalizer(a.kind, metadata.KindFinalizer)

	if err := a.client.Patch(a.ctx, a.kind, patch); err != nil {
		a.log.Error(err, "Failed to add finalizer.")
		return controller.RequeueWithError(err)
	}
	return controller.ContinueProcessing()
}

// EnsureFinalizersAreCalled triggers cleanup logic if the Kind resource is being deleted
// and the finalizer is present. Removes the finalizer after successful cleanup.
func (a *adapter) EnsureFinalizersAreCalled() (controller.OperationResult, error) {
	if a.kind.GetDeletionTimestamp() == nil || !controllerutil.ContainsFinalizer(a.kind, metadata.KindFinalizer) {
		a.log.Info("Skipping finalizer logic; no deletion or no finalizer present")
		return controller.ContinueProcessing()
	}

	if err := a.finalizeKind(); err != nil {
		a.log.Error(err, "Finalization failed during deprovisioning.")
		return controller.RequeueWithError(err)
	}

	kindCopy := a.kind.DeepCopy()
	patch := client.MergeFrom(kindCopy)
	controllerutil.RemoveFinalizer(a.kind, metadata.KindFinalizer)

	if err := a.client.Patch(a.ctx, a.kind, patch); err != nil {
		a.log.Error(err, "Failed to remove finalizer from resource.")
		return controller.RequeueWithError(err)
	}
	return controller.Requeue()
}

// EnsureKindClusterIsProvisioned checks if provisioning should proceed,
// skips if already provisioned or marked for deletion.
func (a *adapter) EnsureKindClusterIsProvisioned() (controller.OperationResult, error) {
	if a.kind.GetDeletionTimestamp() != nil {
		a.log.Info("Resource is marked for deletion, skipping provisioning.")
		return controller.ContinueProcessing()
	}

	switch a.kind.Status.Phase {
	case v1alpha1.KindPhaseProvisioning:
		a.log.Info("Cluster is currently being provisioned.", "phase", a.kind.Status.Phase)
		return controller.StopProcessing()
	case v1alpha1.KindPhaseRunning:
		a.log.Info("Cluster is already provisioned and running.", "phase", a.kind.Status.Phase)
		return controller.StopProcessing()
	case v1alpha1.KindPhaseFailed:
		a.log.Info("Cluster provisioning previously failed. Stopping further retries.", "phase", a.kind.Status.Phase)
		return controller.StopProcessing()
	}

	return a.provisionClusterResources()
}

// provisionClusterResources provisions the Kind cluster using the provisioner,
// then creates the kubeconfig secret and updates status.
func (a *adapter) provisionClusterResources() (controller.OperationResult, error) {
	if a.provisioner == nil {
		err := fmt.Errorf("provisioner is nil")
		a.log.Error(err, "Cannot proceed with provisioning")
		return a.markProvisioningFailed(err)
	}

	if err := a.markClusterProvisioningStarted(); err != nil {
		return controller.RequeueWithError(err)
	}

	if a.kind.Status.ProvisionId == nil || *a.kind.Status.ProvisionId == "" {
		if err := a.client.Get(a.ctx, client.ObjectKeyFromObject(a.kind), a.kind); err != nil {
			a.log.Error(err, "Failed to re-fetch Kind resource after marking provisioning started")
			return controller.RequeueWithError(err)
		}
	}

	provisionMetadata, provisionErr := a.provisioner.Provision(&clusters.MaptCluster{
		Type:   clusters.KindClusterType,
		Object: a.kind,
	})

	if err := validateKindMetadata(provisionMetadata); err != nil {
		return a.markProvisioningFailed(err)
	}
	if provisionErr != nil {
		return a.markProvisioningFailed(provisionErr)
	}

	secretData := map[string][]byte{
		"kubeconfig": []byte(provisionMetadata.KindMetadata.Kubeconfig),
	}

	secretName := fmt.Sprintf("kubeconfig-%s", a.kind.Name)
	existingSecret := &corev1.Secret{}
	err := a.client.Get(a.ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: a.kind.Namespace,
	}, existingSecret)

	if err == nil {
		a.log.Info("Kubeconfig secret already exists; skipping secret creation.", "secret", secretName)
		return a.finalizeSuccessfulProvisioning(secretName, provisionMetadata.KindMetadata.SpotPrice)
	}

	generatedSecretName, err := controllerutils.CreateGeneratedSecret(
		a.ctx, a.client, a.client.Scheme(), secretData, a.kind,
	)
	if err != nil {
		a.log.Error(err, "Failed to create kubeconfig secret after successful provisioning.")
		return a.markSecretCreationFailed(err)
	}

	return a.finalizeSuccessfulProvisioning(generatedSecretName, provisionMetadata.KindMetadata.SpotPrice)
}

// validateKindMetadata ensures the provisioner's response contains valid data.
func validateKindMetadata(meta *clusters.ClusterProvisionerMetadata) error {
	if meta == nil || meta.KindMetadata == nil {
		return fmt.Errorf("provisioner returned nil metadata")
	}
	if meta.KindMetadata.Kubeconfig == "" {
		return fmt.Errorf("provisioner returned empty kubeconfig")
	}
	return nil
}

// markClusterProvisioningStarted sets the status to "Provisioning" and assigns a new provision ID.
func (a *adapter) markClusterProvisioningStarted() error {
	if a.kind.Status.ProvisionId != nil && *a.kind.Status.ProvisionId != "" {
		a.log.Info("Provisioning already started; skipping ProvisionId generation")
		return nil
	}

	provisionId := uuid.New().String()
	err := a.updateStatus(func(s *v1alpha1.KindStatus) {
		*s = *newStatusBuilder(a.kind).
			phase(v1alpha1.KindPhaseProvisioning).
			message("Provisioning of Kind cluster has started.").
			condition("Ready", metav1.ConditionFalse, "ProvisioningStarted", "Cluster provisioning has been initiated and is in progress.").
			backendID(provisionId).
			status
	})
	if err != nil {
		return err
	}

	a.kind.Status.ProvisionId = &provisionId
	return nil
}

// finalizeKind handles cleanup logic during deletion of the Kind resource.
func (a *adapter) finalizeKind() error {
	if a.kind.Status.ProvisionId == nil || *a.kind.Status.ProvisionId == "" {
		a.log.Info("No provision ID found; skipping deprovisioning.")
		return a.updateStatus(func(s *v1alpha1.KindStatus) {
			*s = *newStatusBuilder(a.kind).
				phase(v1alpha1.KindPhaseDeleting).
				message("Skipping deprovisioning: no external resources found for this Kind cluster.").
				condition("Ready", metav1.ConditionFalse, "DeprovisionSkipped", "Cluster marked for deletion, but no provision ID exists; assuming no external resources.").
				status
		})
	}

	if err := a.updateStatus(func(s *v1alpha1.KindStatus) {
		*s = *newStatusBuilder(a.kind).
			phase(v1alpha1.KindPhaseDeleting).
			message("Deprovisioning in progress: external resources are being deleted.").
			condition("Ready", metav1.ConditionFalse, string(v1alpha1.KindPhaseDeleting), "Cluster deletion requested; associated infrastructure cleanup in progress.").
			status
	}); err != nil {
		return err
	}

	if err := a.provisioner.Deprovision(&clusters.MaptCluster{
		Type:   clusters.KindClusterType,
		Object: a.kind,
	}); err != nil {
		a.log.Error(err, "Deprovisioning failed.", "provisionId", *a.kind.Status.ProvisionId)
		_ = a.updateStatus(func(s *v1alpha1.KindStatus) {
			*s = *newStatusBuilder(a.kind).
				phase(v1alpha1.KindPhaseFailed).
				message(fmt.Sprintf("Failed to deprovision cluster: %s", err.Error())).
				condition("Ready", metav1.ConditionFalse, "DeprovisioningFailed", fmt.Sprintf("Error while deprovisioning Kind cluster: %s", err.Error())).
				status
		})
		return err
	}

	return a.updateStatus(func(s *v1alpha1.KindStatus) {
		*s = *newStatusBuilder(a.kind).
			phase(v1alpha1.KindPhaseDeleting).
			message("Kind resources successfully deprovisioned.").
			condition("Ready", metav1.ConditionFalse, "Deprovisioned", "Cluster marked as deleted.").
			status
	})
}

// markProvisioningFailed updates the Kind status when provisioning fails.
func (a *adapter) markProvisioningFailed(err error) (controller.OperationResult, error) {
	a.log.Error(err, "Cluster provisioning failed.")
	_ = a.updateStatus(func(s *v1alpha1.KindStatus) {
		*s = *newStatusBuilder(a.kind).
			phase(v1alpha1.KindPhaseFailed).
			message(fmt.Sprintf("Failed to provision Kind cluster: %s", err.Error())).
			condition("Ready", metav1.ConditionFalse, "ProvisioningFailed", fmt.Sprintf("Provisioning error: %s", err.Error())).
			status
	})
	return controller.RequeueWithError(err)
}

// markSecretCreationFailed sets the Kind status when kubeconfig secret creation fails.
func (a *adapter) markSecretCreationFailed(err error) (controller.OperationResult, error) {
	_ = a.updateStatus(func(s *v1alpha1.KindStatus) {
		*s = *newStatusBuilder(a.kind).
			phase(v1alpha1.KindPhaseFailed).
			message(fmt.Sprintf("Error creating kubeconfig secret: %s", err.Error())).
			condition("Ready", metav1.ConditionFalse, "SecretCreationFailed", fmt.Sprintf("Could not create kubeconfig secret: %s", err.Error())).
			status
	})
	return controller.RequeueWithError(err)
}

// finalizeSuccessfulProvisioning updates the status after a successful cluster provision.
func (a *adapter) finalizeSuccessfulProvisioning(secretName string, avgPrice float64) (controller.OperationResult, error) {
	a.log.Info("Cluster successfully provisioned and kubeconfig secret created.")
	err := a.updateStatus(func(s *v1alpha1.KindStatus) {
		builder := newStatusBuilder(a.kind).
			phase(v1alpha1.KindPhaseRunning).
			message("Kind cluster successfully provisioned and ready.").
			condition("Ready", metav1.ConditionTrue, "Provisioned", "The Kind cluster has been successfully created and is ready for use.").
			avgPrice(avgPrice)
		*s = *builder.status
		s.ClusterReady = true
		s.KubeconfigSecretName = &secretName
	})
	if err != nil {
		a.log.Error(err, "Failed to update status to Running after successful provisioning.")
		return controller.RequeueWithError(err)
	}
	return controller.ContinueProcessing()
}

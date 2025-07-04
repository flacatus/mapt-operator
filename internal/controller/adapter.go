package controller

import (
	"context"

	"github.com/flacatus/mapt-operator/api/v1alpha1"
	"github.com/flacatus/mapt-operator/internal/metadata"
	"github.com/go-logr/logr"
	"github.com/konflux-ci/operator-toolkit/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type adapter struct {
	client      client.Client
	ctx         context.Context
	kind        *v1alpha1.Kind
	validations []controller.ValidationFunction
	provisioner ClusterProvisioner
	log         logr.Logger
}

func newAdapter(ctx context.Context, client client.Client, kind *v1alpha1.Kind, logger logr.Logger) *adapter {
	return &adapter{
		client:      client,
		ctx:         ctx,
		kind:        kind,
		log:         logger.WithValues("name", kind.Name, "namespace", kind.Namespace),
		provisioner: NewKindProvisioner(),
		validations: []controller.ValidationFunction{},
	}
}

func (a *adapter) EnsureFinalizerIsAdded() (controller.OperationResult, error) {
	if controllerutil.ContainsFinalizer(a.kind, metadata.KindFinalizer) {
		return controller.ContinueProcessing()
	}

	a.log.Info("Adding finalizer")
	patch := client.MergeFrom(a.kind.DeepCopy())
	controllerutil.AddFinalizer(a.kind, metadata.KindFinalizer)
	err := a.client.Patch(a.ctx, a.kind, patch)

	return controller.RequeueOnErrorOrContinue(err)
}

func (a *adapter) EnsureFinalizersAreCalled() (controller.OperationResult, error) {
	if a.kind.GetDeletionTimestamp() == nil {
		return controller.ContinueProcessing()
	}

	if controllerutil.ContainsFinalizer(a.kind, metadata.KindFinalizer) {
		a.log.Info("Executing finalizer")

		if err := a.finalizeKind(); err != nil {
			a.log.Error(err, "Finalization failed")
			return controller.RequeueWithError(err)
		}

		a.log.Info("Finalization successful, removing finalizer")
		patch := client.MergeFrom(a.kind.DeepCopy())
		controllerutil.RemoveFinalizer(a.kind, metadata.KindFinalizer)
		if err := a.client.Patch(a.ctx, a.kind, patch); err != nil {
			a.log.Error(err, "Failed to remove finalizer")
			return controller.RequeueWithError(err)
		}
	}

	return controller.Requeue()
}

func (a *adapter) finalizeKind() error {
	if a.kind.Status.ProvisionId == nil || *a.kind.Status.ProvisionId == "" {
		a.log.Info("No provision ID found; skipping deprovisioning")
		return a.patchStatus(func(s *v1alpha1.KindStatus) {
			*s = *newStatusBuilder(a.kind).
				phase(v1alpha1.KindPhaseDeleting).
				message("Skipping deprovisioning: no provision ID found.").
				condition("Ready", metav1.ConditionFalse, "DeprovisionSkipped", "Cluster marked as deleted, but no external resources were found.").
				status
		})
	}

	a.log.Info("Finalizing Kind resource", "provisionId", *a.kind.Status.ProvisionId)
	if err := a.provisioner.Deprovision(a.kind); err != nil {
		a.log.Error(err, "Deprovisioning failed")
		_ = a.patchStatus(func(s *v1alpha1.KindStatus) {
			*s = *newStatusBuilder(a.kind).
				phase(v1alpha1.KindPhaseFailed).
				message("Deprovisioning failed: "+err.Error()).
				condition("Ready", metav1.ConditionFalse, "DeprovisioningFailed", err.Error()).
				backendID(*a.kind.Status.ProvisionId).
				status
		})
		return err
	}

	return a.patchStatus(func(s *v1alpha1.KindStatus) {
		*s = *newStatusBuilder(a.kind).
			phase(v1alpha1.KindPhaseDeleting).
			message("Kind resources successfully deprovisioned.").
			condition("Ready", metav1.ConditionFalse, "Deprovisioned", "Cluster marked as deleted.").
			backendID(*a.kind.Status.ProvisionId).
			status
	})
}

func (a *adapter) EnsureKindClusterIsProvisioned() (controller.OperationResult, error) {
	if a.kind.GetDeletionTimestamp() != nil {
		a.log.Info("Resource is being deleted. Skipping provisioning.")
		return controller.ContinueProcessing()
	}

	switch a.kind.Status.Phase {
	case v1alpha1.KindPhaseProvisioning, v1alpha1.KindPhaseRunning:
		a.log.Info("Cluster is already in progress or running", "phase", a.kind.Status.Phase)
		return controller.ContinueProcessing()
	}

	a.log.Info("Provisioning new Kind cluster")

	if err := a.patchStatus(func(s *v1alpha1.KindStatus) {
		*s = *newStatusBuilder(a.kind).
			phase(v1alpha1.KindPhaseProvisioning).
			message("The Kind cluster is being provisioned.").
			condition("Ready", metav1.ConditionFalse, "ProvisioningStarted", "Cluster provisioning has started.").
			status
	}); err != nil {
		return controller.RequeueWithError(err)
	}

	kubeconfigPath, provisionId, err := a.provisioner.Provision(a.kind)
	if err != nil {
		a.log.Error(err, "Cluster provisioning failed")
		_ = a.patchStatus(func(s *v1alpha1.KindStatus) {
			*s = *newStatusBuilder(a.kind).
				phase(v1alpha1.KindPhaseFailed).
				message("Provisioning failed: "+err.Error()).
				condition("Ready", metav1.ConditionFalse, "ProvisioningFailed", err.Error()).
				backendID(provisionId).
				status
		})
		return controller.RequeueWithError(err)
	}

	if err := createKubeconfigSecret(kubeconfigPath, "kind-kubeconfig-secret", a.kind, a.client); err != nil {
		a.log.Error(err, "Failed to create kubeconfig secret")
		return controller.RequeueWithError(err)
	}

	if err := a.patchStatus(func(s *v1alpha1.KindStatus) {
		*s = *newStatusBuilder(a.kind).
			phase(v1alpha1.KindPhaseRunning).
			message("Cluster successfully provisioned.").
			condition("Ready", metav1.ConditionTrue, "Provisioned", "Cluster is ready.").
			backendID(provisionId).
			status
	}); err != nil {
		return controller.RequeueWithError(err)
	}

	return controller.ContinueProcessing()
}

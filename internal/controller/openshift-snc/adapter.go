package openshiftsnc

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type adapter struct {
	client      client.Client
	ctx         context.Context
	openshift   *v1alpha1.Openshift
	provisioner clusters.GenericMaptProvisioner
	log         logr.Logger
}

func newAdapter(ctx context.Context, c client.Client, p clusters.GenericMaptProvisioner, o *v1alpha1.Openshift, l logr.Logger) *adapter {
	return &adapter{
		client: c, ctx: ctx, openshift: o, provisioner: p,
		log: l.WithValues("name", o.Name, "namespace", o.Namespace),
	}
}

func (a *adapter) EnsureFinalizerIsAdded() (controller.OperationResult, error) {
	if controllerutil.ContainsFinalizer(a.openshift, metadata.OpenshiftSncFinalizer) {
		return controller.ContinueProcessing()
	}
	patch := client.MergeFrom(a.openshift.DeepCopy())
	controllerutil.AddFinalizer(a.openshift, metadata.OpenshiftSncFinalizer)
	if err := a.client.Patch(a.ctx, a.openshift, patch); err != nil {
		return controller.RequeueWithError(controllerutils.LogError(a.log, err, "Failed to add finalizer"))
	}
	return controller.ContinueProcessing()
}

func (a *adapter) EnsureFinalizersAreCalled() (controller.OperationResult, error) {
	if a.openshift.GetDeletionTimestamp() == nil || !controllerutil.ContainsFinalizer(a.openshift, metadata.OpenshiftSncFinalizer) {
		a.log.Info("Skipping finalizer execution")
		return controller.ContinueProcessing()
	}
	if err := a.finalizeOpenshift(); err != nil {
		return controller.RequeueWithError(controllerutils.LogError(a.log, err, "Finalization failed"))
	}
	patch := client.MergeFrom(a.openshift.DeepCopy())
	controllerutil.RemoveFinalizer(a.openshift, metadata.OpenshiftSncFinalizer)
	if err := a.client.Patch(a.ctx, a.openshift, patch); err != nil {
		return controller.RequeueWithError(controllerutils.LogError(a.log, err, "Failed to remove finalizer"))
	}
	return controller.Requeue()
}

func (a *adapter) EnsureOpenshiftClusterIsProvisioned() (controller.OperationResult, error) {
	if a.openshift.GetDeletionTimestamp() != nil {
		a.log.Info("Skipping provisioning: resource is being deleted")
		return controller.ContinueProcessing()
	}
	switch a.openshift.Status.Phase {
	case v1alpha1.OpenshiftSncPhaseProvisioning:
		a.log.Info("Cluster is currently being provisioned.", "phase", a.openshift.Status.Phase)
		return controller.StopProcessing()
	case v1alpha1.OpenshiftSncPhaseRunning:
		a.log.Info("Cluster is already provisioned and running.", "phase", a.openshift.Status.Phase)
		return controller.StopProcessing()
	case v1alpha1.OpenshiftSncPhaseFailed:
		a.log.Info("Cluster provisioning previously failed. Stopping further retries.", "phase", a.openshift.Status.Phase)
		return controller.StopProcessing()
	default:
		return a.provisionClusterResources()
	}
}

func (a *adapter) provisionClusterResources() (controller.OperationResult, error) {
	if a.provisioner == nil {
		return a.fail("provisioner is nil")
	}
	if err := a.markClusterProvisioningStarted(); err != nil {
		return controller.RequeueWithError(err)
	}
	if a.openshift.Status.ProvisionId == nil {
		if err := a.client.Get(a.ctx, client.ObjectKeyFromObject(a.openshift), a.openshift); err != nil {
			return a.fail("failed to refetch after provisioning start", err)
		}
	}
	return a.runProvisioning()
}

func (a *adapter) runProvisioning() (controller.OperationResult, error) {
	meta, err := a.provisioner.Provision(&clusters.MaptCluster{
		Type: clusters.OpenshiftClusterType, Object: a.openshift,
	})
	if err != nil || meta == nil || meta.OpenshiftMetadata == nil {
		return a.fail("provisioning failed", err)
	}
	return a.createAndFinalizeSecret(meta)
}

func (a *adapter) createAndFinalizeSecret(meta *clusters.ClusterProvisionerMetadata) (controller.OperationResult, error) {
	data := map[string][]byte{
		"kubeconfig":        []byte(meta.OpenshiftMetadata.Kubeconfig),
		"kubeadminPassword": []byte(meta.OpenshiftMetadata.KubeadminPassword),
		"consoleURL":        []byte(meta.OpenshiftMetadata.ConsoleURL),
		"privateKey":        []byte(meta.OpenshiftMetadata.PrivateKey),
		"host":              []byte(meta.OpenshiftMetadata.Host),
		"username":          []byte(meta.OpenshiftMetadata.Username),
	}
	name, err := controllerutils.CreateGeneratedSecret(a.ctx, a.client, a.client.Scheme(), data, a.openshift)
	if err != nil {
		return a.fail("failed to create kubeconfig secret", err)
	}
	return a.success(name)
}

func (a *adapter) success(secret string) (controller.OperationResult, error) {
	a.log.Info("Cluster provisioned", "secret", secret)
	err := a.updateStatus(func(s *v1alpha1.OpenshiftStatus) {
		*s = *newStatusBuilder(a.openshift).
			phase(v1alpha1.OpenshiftSncPhaseRunning).
			message("Cluster provisioning completed successfully.").
			condition("Ready", metav1.ConditionTrue, "Provisioned", "The OpenShift cluster is fully provisioned and operational.").
			kubeconfigSecret(secret).status
		s.ClusterReady = true
	})
	if err != nil {
		return controller.RequeueWithError(err)
	}
	return controller.ContinueProcessing()
}

func (a *adapter) fail(msg string, err ...error) (controller.OperationResult, error) {
	var e error
	if len(err) > 0 {
		e = err[0]
	} else {
		e = fmt.Errorf("%s", msg)
	}
	fullMessage := fmt.Sprintf("%s: %v", msg, e)
	a.log.Error(e, msg)
	_ = a.updateStatus(func(s *v1alpha1.OpenshiftStatus) {
		*s = *newStatusBuilder(a.openshift).
			phase(v1alpha1.OpenshiftSncPhaseFailed).
			message(fullMessage).
			condition("Ready", metav1.ConditionFalse, "Failed", "Provisioning failed: "+e.Error()).status
	})
	return controller.RequeueWithError(e)
}

func (a *adapter) markClusterProvisioningStarted() error {
	if a.openshift.Status.ProvisionId != nil {
		return nil
	}
	id := uuid.New().String()

	err := a.updateStatus(func(s *v1alpha1.OpenshiftStatus) {
		*s = *newStatusBuilder(a.openshift).
			phase(v1alpha1.OpenshiftSncPhaseProvisioning).
			message("Cluster provisioning has started.").
			condition("Ready", metav1.ConditionFalse, "ProvisioningStarted", "The provisioning process has been initiated.").
			backendID(id).status
	})
	if err == nil {
		a.openshift.Status.ProvisionId = &id
	}
	return err
}

func (a *adapter) finalizeOpenshift() error {
	if a.openshift.Status.ProvisionId == nil {
		a.log.Info("No provision ID; skipping deprovisioning")
		return a.updateStatus(func(s *v1alpha1.OpenshiftStatus) {
			*s = *newStatusBuilder(a.openshift).
				phase(v1alpha1.OpenshiftSncPhaseDeleting).
				message("No provision ID found; skipping deprovisioning.").
				condition("Ready", metav1.ConditionFalse, "DeprovisionSkipped", "Cluster deletion completed without deprovisioning.").status
		})
	}
	if err := a.provisioner.Deprovision(&clusters.MaptCluster{
		Type: clusters.OpenshiftClusterType, Object: a.openshift,
	}); err != nil {
		return controllerutils.LogError(a.log, err, "Deprovisioning failed")
	}
	a.log.Info("Resources deprovisioned")
	return a.updateStatus(func(s *v1alpha1.OpenshiftStatus) {
		*s = *newStatusBuilder(a.openshift).
			phase(v1alpha1.OpenshiftSncPhaseDeleting).
			message("Cluster resources have been deprovisioned.").
			condition("Ready", metav1.ConditionFalse, "Deprovisioned", "Cluster was deprovisioned and marked for deletion.").status
	})
}

/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	"github.com/flacatus/mapt-operator/api/v1alpha1"
	"github.com/flacatus/mapt-operator/internal/metadata"
	"github.com/go-logr/logr"
	"github.com/konflux-ci/operator-toolkit/controller"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/action/kind"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// adapter holds the objects needed to reconcile a Kind resource.
type adapter struct {
	client      client.Client
	ctx         context.Context
	kind        *v1alpha1.Kind
	validations []controller.ValidationFunction
	log         logr.Logger // Add this line
}

// newAdapter creates and returns an adapter instance for a given Kind resource.
func newAdapter(ctx context.Context, client client.Client, kind *v1alpha1.Kind, logger logr.Logger) *adapter { // <--- MODIFIED SIGNATURE
	kindAdapter := &adapter{
		client: client,
		ctx:    ctx,
		kind:   kind,
		log:    logger, // <--- ASSIGN THE PASSED LOGGER
	}

	kindAdapter.validations = []controller.ValidationFunction{}

	return kindAdapter
}

// finalizeKind defines the cleanup logic that is executed when a Kind resource is marked for deletion.
// This function is responsible for ensuring that all external resources, such as the AWS EC2 instance,
// are properly destroyed before the Kubernetes object is removed.
func (a *adapter) finalizeKind() error {
	// TODO: Implement the logic to delete the external AWS resources.
	// This would involve calling the AWS helper function to terminate the instance
	// associated with this Kind resource. For example:
	//
	// if a.kind.Status.AWSInstanceID != nil {
	//     creds, err := r.getCloudCredentials(a.ctx, a.kind)
	//     if err != nil {
	//         return err
	//     }
	//     return r.deleteAWSInstance(a.ctx, *a.kind.Status.AWSInstanceID, creds)
	// }
	return nil
}

// EnsureFinalizersAreCalled handles the cleanup logic when a Kind resource is marked for deletion.
// It checks for a deletion timestamp. If found, it executes the `finalizeKind` function to delete any
// external resources. Upon successful cleanup, it removes the finalizer, allowing Kubernetes to fully
// delete the resource. If the cleanup fails, it returns an error to retry the operation.
func (a *adapter) EnsureFinalizersAreCalled() (controller.OperationResult, error) {
	// Check if the Kind resource is marked for deletion and continue processing other operations otherwise
	if a.kind.GetDeletionTimestamp() == nil {
		return controller.ContinueProcessing()
	}

	if controllerutil.ContainsFinalizer(a.kind, metadata.KindFinalizer) {
		// Call finalizeKind to perform the actual cleanup of external resources.
		if err := a.finalizeKind(); err != nil {
			return controller.RequeueWithError(err)
		}

		// Once cleanup is successful, remove the finalizer from the Kind resource.
		patch := client.MergeFrom(a.kind.DeepCopy())
		controllerutil.RemoveFinalizer(a.kind, metadata.KindFinalizer)
		err := a.client.Patch(a.ctx, a.kind, patch)
		if err != nil {
			return controller.RequeueWithError(err)
		}
	}

	// Requeue the Kind resource again so it gets deleted and other operations are not executed.
	return controller.Requeue()
}

// EnsureFinalizerIsAdded checks if the Kind resource has the necessary finalizer. If the finalizer
// is missing, it adds it to the resource. This is a crucial step to ensure that the cleanup logic in
// `EnsureFinalizersAreCalled` can be executed when the resource is eventually deleted. It prevents the
// resource from being deleted prematurely before its external dependencies are cleaned up.
func (a *adapter) EnsureFinalizerIsAdded() (controller.OperationResult, error) {
	var finalizerFound bool
	for _, finalizer := range a.kind.GetFinalizers() {
		if finalizer == metadata.KindFinalizer {
			finalizerFound = true
		}
	}

	if !finalizerFound {
		patch := client.MergeFrom(a.kind.DeepCopy())
		controllerutil.AddFinalizer(a.kind, metadata.KindFinalizer)
		err := a.client.Patch(a.ctx, a.kind, patch)

		return controller.RequeueOnErrorOrContinue(err)
	}

	return controller.ContinueProcessing()
}

// This approach ensures that the status is always patched with the latest state from the cluster.
// It's a common pattern to avoid race conditions.
func (a *adapter) patchStatus(mutator func(status *v1alpha1.KindStatus)) (controller.OperationResult, error) {
	// Create a patch from the original Kind resource
	patch := client.MergeFrom(a.kind.DeepCopy())

	// Mutate the status using the provided function
	mutator(&a.kind.Status)

	// Automatically update the LastUpdateTime field
	a.kind.Status.LastUpdateTime = &metav1.Time{Time: time.Now()}

	// Patch the status subresource
	err := a.client.Status().Patch(a.ctx, a.kind, patch)
	if err != nil {
		return controller.RequeueWithError(err)
	}

	return controller.ContinueProcessing()
}

// setStatusCondition is a helper to set a condition on the Kind resource's status.
// It uses the meta.SetStatusCondition from apimachinery to correctly add or update a condition in the slice.
func (a *adapter) setStatusCondition(conditionType string, status metav1.ConditionStatus, reason, message string) {
	newCondition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: a.kind.Generation,
		LastTransitionTime: metav1.Now(),
	}
	meta.SetStatusCondition(&a.kind.Status.Conditions, newCondition)
}

func (a *adapter) EnsureKindClusterIsProvisioned() (controller.OperationResult, error) {
	if a.kind.GetDeletionTimestamp() != nil || (a.kind.Status.Phase == "Provisioning" || a.kind.Status.Phase == "Running") {
		a.log.V(1).Info("Skipping Kind cluster provisioning as it's already in progress or completed, or marked for deletion.")
		return controller.ContinueProcessing() // Or return ctrl.Result{}
	}

	a.log.Info("Starting Kind cluster provisioning", "name", a.kind.Name, "namespace", a.kind.Namespace)
	_, err := a.patchStatus(func(status *v1alpha1.KindStatus) {
		status.Phase = "Provisioning"
		status.Message = "The Kind cluster is being provisioned."
		a.setStatusCondition("Ready", metav1.ConditionFalse, "ProvisioningStarted", "Cluster provisioning has started.")
	})
	if err != nil {
		a.log.Error(err, "Failed to patch status to Provisioning")
		return controller.RequeueWithError(err)
	}

	a.log.Info("Calling external Kind create function", "cpus", a.kind.Spec.MachineConfig.CPUs, "memoryGiB", a.kind.Spec.MachineConfig.MemoryGiB)
	err = kind.Create(
		&maptContext.ContextArgs{
			ProjectName:           a.kind.Name,
			BackedURL:             "s3://mapt-kind-bucket/mapt/kind/91782",
			SpotPriceIncreaseRate: 10,
			Tags:                  a.kind.Spec.MachineConfig.Tags,
			ForceDestroy:          true,
		},
		&kind.KindArgs{
			Prefix: a.kind.Name,
			Arch:   a.kind.Spec.MachineConfig.Architecture,
			InstanceRequest: &instancetypes.AwsInstanceRequest{
				CPUs:       a.kind.Spec.MachineConfig.CPUs,
				MemoryGib:  a.kind.Spec.MachineConfig.MemoryGiB,
				Arch:       instancetypes.Amd64,
				NestedVirt: false,
			},
			Version: a.kind.Spec.KindClusterConfig.KubernetesVersion,
			Spot:    true,
		})

	if err != nil {
		a.log.Error(err, "Failed to create Kind cluster via external call")
		_, patchErr := a.patchStatus(func(status *v1alpha1.KindStatus) {
			status.Phase = "Failed"
			status.Message = "Provisioning failed: " + err.Error()
			a.setStatusCondition("Ready", metav1.ConditionFalse, "ProvisioningFailed", err.Error())
		})
		if patchErr != nil {
			a.log.Error(patchErr, "Failed to patch status to Failed after Kind creation error")
		}
		return controller.RequeueWithError(err)
	}

	a.log.Info("Kind cluster successfully provisioned")
	_, err = a.patchStatus(func(status *v1alpha1.KindStatus) {
		status.Phase = "Running"
		status.Message = "Cluster successfully provisioned."
		a.setStatusCondition("Ready", metav1.ConditionTrue, "Provisioned", "Cluster is ready for use.")
	})
	if err != nil {
		a.log.Error(err, "Failed to patch status to Running after successful provisioning")
		return controller.RequeueWithError(err)
	}

	return controller.ContinueProcessing()
}

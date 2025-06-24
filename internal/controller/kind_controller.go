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

	maptv1alpha1 "github.com/flacatus/mapt-operator/api/v1alpha1"
	"github.com/konflux-ci/operator-toolkit/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// KindReconciler reconciles a Kind object
type KindReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mapt.redhat.com,resources=kinds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mapt.redhat.com,resources=kinds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mapt.redhat.com,resources=kinds/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
func (r *KindReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("controller", "KindReconciler", "resource", req.NamespacedName)

	var kind maptv1alpha1.Kind
	if err := r.Get(ctx, req.NamespacedName, &kind); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Kind resource not found. Assuming it was deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to fetch Kind resource")
		return ctrl.Result{}, err
	}

	adapter := newAdapter(ctx, r.Client, &kind, logger)

	result, err := controller.ReconcileHandler([]controller.Operation{
		adapter.EnsureFinalizersAreCalled,
		adapter.EnsureFinalizerIsAdded,
		adapter.EnsureKindClusterIsProvisioned,
	})

	if err != nil {
		logger.Error(err, "Reconciliation step failed")
		return result, err
	}

	logger.Info("Reconciliation successful. Requeueing after 10 seconds")
	result.RequeueAfter = 10 * time.Hour
	return result, nil
}

func (r *KindReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&maptv1alpha1.Kind{}).
		Owns(&corev1.Secret{}).
		Named("kind").
		Complete(r)
}
